package s3store_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/storage/s3store"
)

// newTestBackend points a Backend at a local httptest.Server standing in for
// the provider's S3-compatible endpoint, so Put/Del/Walk can be exercised
// against scripted HTTP responses with no real bucket or credentials.
func newTestBackend(t *testing.T, srv *httptest.Server) *s3store.Backend {
	t.Helper()
	backend, err := s3store.New(context.Background(), s3store.Config{
		Name: "r2", Endpoint: srv.URL, Region: "auto",
		AccessKeyID: "id", SecretAccessKey: "secret", Bucket: "bucket", PresignTTL: 15 * time.Minute,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return backend
}

func TestR2Endpoint(t *testing.T) {
	got := s3store.R2Endpoint("abc123")
	want := "https://abc123.r2.cloudflarestorage.com"
	if got != want {
		t.Errorf("R2Endpoint(%q) = %q, want %q", "abc123", got, want)
	}
}

// TestNew_BuildsBackendForEachProviderShape exercises New with the endpoint
// shape each configured provider actually uses (R2's own helper, and the
// literal endpoints B2/Supabase display in their dashboards). No network
// call happens here - aws-sdk-go-v2's LoadDefaultConfig with static
// credentials doesn't touch the network, so this only checks that Config is
// wired into a working *Backend, not that the bucket is reachable.
func TestNew_BuildsBackendForEachProviderShape(t *testing.T) {
	cases := []struct {
		name string
		cfg  s3store.Config
	}{
		{"r2", s3store.Config{
			Name: "r2", Endpoint: s3store.R2Endpoint("acct"), Region: "auto",
			AccessKeyID: "id", SecretAccessKey: "secret", Bucket: "bucket", PresignTTL: 15 * time.Minute,
		}},
		{"b2", s3store.Config{
			Name: "b2", Endpoint: "https://s3.us-west-004.backblazeb2.com", Region: "us-west-004",
			AccessKeyID: "id", SecretAccessKey: "secret", Bucket: "bucket", PresignTTL: 15 * time.Minute,
		}},
		{"supabase", s3store.Config{
			Name: "supabase", Endpoint: "https://xyzcompany.supabase.co/storage/v1/s3", Region: "eu-west-1",
			AccessKeyID: "id", SecretAccessKey: "secret", Bucket: "bucket", PresignTTL: 15 * time.Minute,
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			backend, err := s3store.New(context.Background(), tc.cfg)
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			if backend.Name() != tc.cfg.Name {
				t.Errorf("Name() = %q, want %q", backend.Name(), tc.cfg.Name)
			}
			var _ ports.IStorageBackend = backend
		})
	}
}

// accessDeniedXML is a minimal S3 error body with a non-retryable code, so
// these tests don't eat the SDK's default retry/backoff delay.
const accessDeniedXML = `<?xml version="1.0" encoding="UTF-8"?>
<Error><Code>AccessDenied</Code><Message>denied</Message><RequestId>1</RequestId></Error>`

func TestPut_WrapsUnderlyingError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = io.WriteString(w, accessDeniedXML)
	}))
	defer srv.Close()
	backend := newTestBackend(t, srv)

	err := backend.Put(context.Background(), "stands/1/main.webp", strings.NewReader("hi"), "image/webp", 2)
	if err == nil {
		t.Fatal("Put: want error when the backend rejects the request, got nil")
	}
	if !strings.Contains(err.Error(), "stands/1/main.webp") || !strings.Contains(err.Error(), "r2") {
		t.Errorf("err = %q, want it to mention the key and provider name", err.Error())
	}
}

func TestDel_WrapsUnderlyingError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = io.WriteString(w, accessDeniedXML)
	}))
	defer srv.Close()
	backend := newTestBackend(t, srv)

	err := backend.Del(context.Background(), "stands/1/main.webp")
	if err == nil {
		t.Fatal("Del: want error when the backend rejects the request, got nil")
	}
	if !strings.Contains(err.Error(), "stands/1/main.webp") || !strings.Contains(err.Error(), "r2") {
		t.Errorf("err = %q, want it to mention the key and provider name", err.Error())
	}
}

func TestWalk_WrapsUnderlyingError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = io.WriteString(w, accessDeniedXML)
	}))
	defer srv.Close()
	backend := newTestBackend(t, srv)

	err := backend.Walk(context.Background(), func(string, int64) error { return nil })
	if err == nil {
		t.Fatal("Walk: want error when listing fails, got nil")
	}
	if !strings.Contains(err.Error(), "r2") {
		t.Errorf("err = %q, want it to mention the provider name", err.Error())
	}
}

// listObjectsPage renders a minimal ListObjectsV2 XML response for one page
// of keys.
func listObjectsPage(keys []string, sizes []int64, truncated bool, nextToken string) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?><ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/">`)
	fmt.Fprintf(&b, "<IsTruncated>%t</IsTruncated>", truncated)
	if nextToken != "" {
		fmt.Fprintf(&b, "<NextContinuationToken>%s</NextContinuationToken>", nextToken)
	}
	for i, k := range keys {
		fmt.Fprintf(&b, "<Contents><Key>%s</Key><Size>%d</Size></Contents>", k, sizes[i])
	}
	b.WriteString(`</ListBucketResult>`)
	return b.String()
}

func TestWalk_PaginatesAllPages(t *testing.T) {
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/xml")
		token := r.URL.Query().Get("continuation-token")
		if token == "" {
			_, _ = io.WriteString(w, listObjectsPage([]string{"a", "b"}, []int64{10, 20}, true, "page2"))
			return
		}
		if token != "page2" {
			t.Errorf("unexpected continuation-token %q", token)
		}
		_, _ = io.WriteString(w, listObjectsPage([]string{"c"}, []int64{30}, false, ""))
	}))
	defer srv.Close()
	backend := newTestBackend(t, srv)

	got := map[string]int64{}
	if err := backend.Walk(context.Background(), func(key string, bytes int64) error {
		got[key] = bytes
		return nil
	}); err != nil {
		t.Fatalf("Walk: %v", err)
	}

	want := map[string]int64{"a": 10, "b": 20, "c": 30}
	if len(got) != len(want) {
		t.Fatalf("Walk visited %v, want %v", got, want)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("got[%q] = %d, want %d", k, got[k], v)
		}
	}
	if requests != 2 {
		t.Errorf("requests = %d, want 2 (one per page)", requests)
	}
}

func TestWalk_StopsWhenCallbackErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		_, _ = io.WriteString(w, listObjectsPage([]string{"a", "b"}, []int64{10, 20}, true, "page2"))
	}))
	defer srv.Close()
	backend := newTestBackend(t, srv)

	stopErr := fmt.Errorf("stop here")
	calls := 0
	err := backend.Walk(context.Background(), func(string, int64) error {
		calls++
		return stopErr
	})
	if err != stopErr {
		t.Fatalf("err = %v, want the callback's own error propagated as-is", err)
	}
	if calls != 1 {
		t.Errorf("callback calls = %d, want 1 (Walk must stop at the first error)", calls)
	}
}

func TestPresignGet_URLReflectsBucketKeyAndTTL(t *testing.T) {
	backend, err := s3store.New(context.Background(), s3store.Config{
		Name: "r2", Endpoint: "https://example.invalid", Region: "auto",
		AccessKeyID: "id", SecretAccessKey: "secret", Bucket: "my-bucket", PresignTTL: 7 * time.Minute,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	raw, err := backend.PresignGet(context.Background(), "stands/1/thumb.webp")
	if err != nil {
		t.Fatalf("PresignGet: %v", err)
	}
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parsing presigned URL %q: %v", raw, err)
	}
	if !strings.Contains(u.Path, "my-bucket") || !strings.Contains(u.Path, "stands/1/thumb.webp") {
		t.Errorf("path = %q, want it to contain the bucket and key", u.Path)
	}
	expires, err := strconv.Atoi(u.Query().Get("X-Amz-Expires"))
	if err != nil {
		t.Fatalf("X-Amz-Expires missing or not an int: %v", err)
	}
	if expires != 7*60 {
		t.Errorf("X-Amz-Expires = %d, want %d (PresignTTL in seconds)", expires, 7*60)
	}
}
