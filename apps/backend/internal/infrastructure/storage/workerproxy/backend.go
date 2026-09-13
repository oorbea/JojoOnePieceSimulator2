// Package workerproxy adapts ports.IStorageBackend onto a small Cloudflare
// Worker (apps/r2-worker-proxy) that fronts one R2 bucket via a native R2
// binding, instead of talking to R2's S3 API endpoint
// (<account>.r2.cloudflarestorage.com) directly.
//
// Why this exists: that S3 API endpoint sits on a specific Cloudflare
// anycast prefix, and an ISP can have a broken route to it while every
// other Cloudflare-fronted domain works fine (see the 2026-09-13 incident
// in ObsidianVault/storage-fallback-chain.md - the ISP's own network
// dropped every packet to that one prefix). The Worker lives on the
// generic *.workers.dev pool instead, and the Worker-to-R2 hop happens
// entirely inside Cloudflare's network, never touching that prefix. This
// package is opt-in: infrastructure/storage/s3store.Backend remains the
// direct-S3-API implementation and is what's used when a Worker isn't
// configured (see cmd/app/main.go's buildStorageTiers).
package workerproxy

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// Config holds everything needed to reach one Worker-fronted bucket.
type Config struct {
	// Name identifies this backend for the ledger and config, e.g. "r2".
	Name string
	// BaseURL is the Worker's own URL, e.g.
	// "https://r2-proxy.<subdomain>.workers.dev" (no trailing slash
	// required - New trims one if present).
	BaseURL string
	// Secret is the shared bearer token this backend sends as
	// "Authorization: Bearer <secret>", and PresignGet signs URLs with.
	// Must match the Worker's own R2_PROXY_SECRET.
	Secret string
	// PresignTTL bounds how long a URL returned by PresignGet stays valid.
	PresignTTL time.Duration
}

// Backend is the Config-configured storage.Backend implementation talking
// to apps/r2-worker-proxy over plain HTTP.
type Backend struct {
	name    string
	baseURL string
	secret  string
	ttl     time.Duration
	client  *http.Client
}

var _ ports.IStorageBackend = (*Backend)(nil)

// New builds a Backend talking to the Worker described by cfg.
func New(cfg Config) (*Backend, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("workerproxy: BaseURL must not be empty")
	}
	if cfg.Secret == "" {
		return nil, fmt.Errorf("workerproxy: Secret must not be empty")
	}
	return &Backend{
		name:    cfg.Name,
		baseURL: strings.TrimSuffix(cfg.BaseURL, "/"),
		secret:  cfg.Secret,
		ttl:     cfg.PresignTTL,
		client:  http.DefaultClient,
	}, nil
}

// Name implements ports.IStorageBackend.
func (b *Backend) Name() string { return b.name }

// objectURL builds {baseURL}/objects/{key}, percent-encoding each path
// segment individually so a key containing "/" (the normal case - keys are
// always "<prefix>/<subjectID>/<file>") round-trips as separate path
// segments rather than one opaque escaped blob.
func (b *Backend) objectURL(key string) string {
	segments := strings.Split(key, "/")
	for i, s := range segments {
		segments[i] = url.PathEscape(s)
	}
	return b.baseURL + "/objects/" + strings.Join(segments, "/")
}

func (b *Backend) newRequest(ctx context.Context, method, rawURL string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, rawURL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+b.secret)
	return req, nil
}

// Put implements ports.IStorageBackend.
func (b *Backend) Put(ctx context.Context, key string, content io.Reader, contentType string, size int64) error {
	req, err := b.newRequest(ctx, http.MethodPut, b.objectURL(key), content)
	if err != nil {
		return fmt.Errorf("building request to upload %q via %s worker: %w", key, b.name, err)
	}
	req.Header.Set("Content-Type", contentType)
	req.ContentLength = size

	resp, err := b.client.Do(req)
	if err != nil {
		return fmt.Errorf("uploading %q via %s worker: %w", key, b.name, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("uploading %q via %s worker: unexpected status %d", key, b.name, resp.StatusCode)
	}
	return nil
}

// Get implements ports.IStorageBackend, mapping a 404 from the Worker to
// ports.ErrObjectNotFound instead of a generic error.
func (b *Backend) Get(ctx context.Context, key string) (io.ReadCloser, ports.ObjectInfo, error) {
	req, err := b.newRequest(ctx, http.MethodGet, b.objectURL(key), nil)
	if err != nil {
		return nil, ports.ObjectInfo{}, fmt.Errorf("building request to download %q via %s worker: %w", key, b.name, err)
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, ports.ObjectInfo{}, fmt.Errorf("downloading %q via %s worker: %w", key, b.name, err)
	}
	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return nil, ports.ObjectInfo{}, fmt.Errorf("%w: %q on %s", ports.ErrObjectNotFound, key, b.name)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		resp.Body.Close()
		return nil, ports.ObjectInfo{}, fmt.Errorf("downloading %q via %s worker: unexpected status %d", key, b.name, resp.StatusCode)
	}
	return resp.Body, ports.ObjectInfo{
		ContentType: resp.Header.Get("Content-Type"),
		Size:        resp.ContentLength,
	}, nil
}

// PresignGet implements ports.IStorageBackend. Instead of a real S3
// presigned URL, it hands out a time-limited URL to the Worker itself,
// signed with an HMAC of key|exp under the shared secret - the Worker
// validates that signature as an alternative to the Authorization header
// on GET (see apps/r2-worker-proxy/src/index.ts).
func (b *Backend) PresignGet(_ context.Context, key string) (string, error) {
	exp := time.Now().Add(b.ttl).Unix()
	mac := hmac.New(sha256.New, []byte(b.secret))
	mac.Write([]byte(fmt.Sprintf("%s|%d", key, exp)))
	sig := hex.EncodeToString(mac.Sum(nil))

	q := url.Values{}
	q.Set("exp", strconv.FormatInt(exp, 10))
	q.Set("sig", sig)
	return b.objectURL(key) + "?" + q.Encode(), nil
}

// Del implements ports.IStorageBackend. Deleting a key the Worker reports
// as missing (404) is not an error, matching every other backend.
func (b *Backend) Del(ctx context.Context, key string) error {
	req, err := b.newRequest(ctx, http.MethodDelete, b.objectURL(key), nil)
	if err != nil {
		return fmt.Errorf("building request to delete %q via %s worker: %w", key, b.name, err)
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return fmt.Errorf("deleting %q via %s worker: %w", key, b.name, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("deleting %q via %s worker: unexpected status %d", key, b.name, resp.StatusCode)
	}
	return nil
}

// listPage is the JSON shape the Worker's GET /objects list endpoint
// returns - see apps/r2-worker-proxy/src/index.ts.
type listPage struct {
	Keys []struct {
		Key  string `json:"key"`
		Size int64  `json:"size"`
	} `json:"keys"`
	Cursor string `json:"cursor"`
}

// Walk implements ports.IStorageBackend, paging through the Worker's list
// endpoint via its opaque cursor until it comes back empty.
func (b *Backend) Walk(ctx context.Context, fn func(key string, bytes int64) error) error {
	cursor := ""
	for {
		listURL := b.baseURL + "/objects"
		if cursor != "" {
			q := url.Values{}
			q.Set("cursor", cursor)
			listURL += "?" + q.Encode()
		}

		req, err := b.newRequest(ctx, http.MethodGet, listURL, nil)
		if err != nil {
			return fmt.Errorf("building request to list objects via %s worker: %w", b.name, err)
		}
		resp, err := b.client.Do(req)
		if err != nil {
			return fmt.Errorf("listing objects via %s worker: %w", b.name, err)
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			resp.Body.Close()
			return fmt.Errorf("listing objects via %s worker: unexpected status %d", b.name, resp.StatusCode)
		}

		var page listPage
		decodeErr := json.NewDecoder(resp.Body).Decode(&page)
		resp.Body.Close()
		if decodeErr != nil {
			return fmt.Errorf("decoding object list from %s worker: %w", b.name, decodeErr)
		}

		for _, obj := range page.Keys {
			if err := fn(obj.Key, obj.Size); err != nil {
				return err
			}
		}

		if page.Cursor == "" {
			return nil
		}
		cursor = page.Cursor
	}
}
