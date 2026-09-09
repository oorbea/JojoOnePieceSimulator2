package endpoints

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/dto"
)

// groupIDPattern is what makes the media route safe against path traversal:
// a 32-hex-char group id can never contain "/" or "..". Combined with the
// fixed variant allowlist below, there is no user-controlled path segment
// that reaches the filesystem or object storage unvalidated.
var groupIDPattern = regexp.MustCompile(`^[0-9a-f]{32}$`)

// mediaVariants is the fixed allowlist of servable variants, in
// most-specific-first fallback order per requested variant - see
// variantLadder. Falling back doesn't mean "next smallest": card (~512px)
// falls to main (~1024px) rather than thumb (~256px), because the card
// well is large enough that a 256px fallback would look as blurry as the
// bug this ladder exists to avoid - an oversized-but-sharp main beats an
// undersized thumb here.
var mediaVariants = map[string][]string{
	"card":  {"card", "main", "thumb"},
	"thumb": {"thumb", "main"},
	"main":  {"main"},
}

// MediaConfig bounds MediaEndpoints' behavior - see config.Config's
// Media* fields for where each value comes from.
type MediaConfig struct {
	Mode          string // "proxy" or "redirect"
	CacheDir      string
	CacheMaxBytes int64
}

// MediaEndpoints serves the immutable content-addressed media proxy:
// GET /api/v1/media/{group}/{variant}.webp (public) and
// GET /api/v1/media/p/{exp}/{sig}/{group}/{variant}.webp (private, signed).
// Mounted outside RequireAuth (see router.go) - an <img src> can't carry a
// bearer token, so authority lives in the URL itself: an unguessable group
// id for public scope, a time-limited HMAC signature for private scope.
type MediaEndpoints struct {
	pictures ports.IPictureStorage
	media    ports.IMediaRepository
	urls     dto.MediaURLBuilder
	cfg      MediaConfig
}

func NewMediaEndpoints(pictures ports.IPictureStorage, media ports.IMediaRepository, urls dto.MediaURLBuilder, cfg MediaConfig) *MediaEndpoints {
	return &MediaEndpoints{pictures: pictures, media: media, urls: urls, cfg: cfg}
}

// Routes returns the /media sub-router. rateCfg's media tier is applied by
// the caller (router.go), not here - same layering as every other
// Routes(rateCfg, ...) method in this package.
func (e *MediaEndpoints) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/{group}/{variantFile}", Wrap(e.public))
	r.Head("/{group}/{variantFile}", Wrap(e.public))
	r.Get("/p/{exp}/{sig}/{group}/{variantFile}", Wrap(e.private))
	r.Head("/p/{exp}/{sig}/{group}/{variantFile}", Wrap(e.private))
	return r
}

// parseVariantFile splits "{variant}.webp" into variant, validating both the
// ".webp" suffix and that variant is in the fixed allowlist.
func parseVariantFile(variantFile string) (string, bool) {
	variant, ok := strings.CutSuffix(variantFile, ".webp")
	if !ok {
		return "", false
	}
	_, known := mediaVariants[variant]
	return variant, known
}

// public serves a public-scope group's rendition - no signature required.
func (e *MediaEndpoints) public(w http.ResponseWriter, r *http.Request) error {
	group := chi.URLParam(r, "group")
	if !groupIDPattern.MatchString(group) {
		return &dto.ValidationError{Errors: []string{"invalid group id"}}
	}
	variant, ok := parseVariantFile(chi.URLParam(r, "variantFile"))
	if !ok {
		return &dto.ValidationError{Errors: []string{"invalid variant"}}
	}
	return e.serve(w, r, group, variant, "public", nil)
}

// private serves a private-scope group's rendition, requiring a valid,
// unexpired signature over exp|group|variant - see dto.MediaURLBuilder.
func (e *MediaEndpoints) private(w http.ResponseWriter, r *http.Request) error {
	group := chi.URLParam(r, "group")
	if !groupIDPattern.MatchString(group) {
		return &dto.ValidationError{Errors: []string{"invalid group id"}}
	}
	variant, ok := parseVariantFile(chi.URLParam(r, "variantFile"))
	if !ok {
		return &dto.ValidationError{Errors: []string{"invalid variant"}}
	}
	exp, err := strconv.ParseInt(chi.URLParam(r, "exp"), 10, 64)
	if err != nil {
		return &dto.ValidationError{Errors: []string{"invalid exp"}}
	}
	sig := chi.URLParam(r, "sig")
	now := time.Now()
	if !e.urls.VerifyPrivate(exp, group, variant, sig, now) {
		return errMediaSignatureInvalid
	}
	maxAge := time.Unix(exp, 0).Sub(now)
	return e.serve(w, r, group, variant, "private", &maxAge)
}

// serve resolves group/requestedVariant against media_objects (falling down
// the card->thumb->main ladder so a partially-backfilled group never 404s),
// then either redirects to a presigned URL (MediaMode=redirect) or proxies
// the bytes itself through the on-disk cache (the default). privateMaxAge is
// nil for public scope (immutable forever); non-nil for private scope
// (bounded by how much of the signature's window remains).
func (e *MediaEndpoints) serve(w http.ResponseWriter, r *http.Request, group, requestedVariant, scope string, privateMaxAge *time.Duration) error {
	objects, err := e.media.GetMediaObjectsByGroup(r.Context(), group)
	if err != nil {
		return err
	}
	byVariant := make(map[string]ports.MediaObject, len(objects))
	for _, o := range objects {
		if o.Scope != scope {
			continue
		}
		byVariant[o.Variant] = o
	}

	var selected ports.MediaObject
	found := false
	for _, v := range mediaVariants[requestedVariant] {
		if o, ok := byVariant[v]; ok {
			selected = o
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("%w: group %s has no %s-scope renditions", ports.ErrObjectNotFound, group, scope)
	}

	if e.cfg.Mode == "redirect" {
		url, err := e.pictures.PresignGetURL(r.Context(), selected.StorageKey)
		if err != nil {
			return err
		}
		http.Redirect(w, r, url, http.StatusFound)
		return nil
	}

	etag := fmt.Sprintf(`"%s-%s"`, group, selected.Variant)
	if r.Header.Get("If-None-Match") == etag {
		w.Header().Set("ETag", etag)
		w.WriteHeader(http.StatusNotModified)
		return nil
	}

	data, err := e.readThroughCache(r.Context(), selected)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "image/webp")
	w.Header().Set("ETag", etag)
	w.Header().Set("Accept-Ranges", "none")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if privateMaxAge != nil {
		secs := int64(*privateMaxAge / time.Second)
		if secs < 0 {
			secs = 0
		}
		w.Header().Set("Cache-Control", fmt.Sprintf("private, max-age=%d, immutable", secs))
	} else {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	}

	if r.Method == http.MethodHead {
		w.Header().Set("Content-Length", strconv.Itoa(len(data)))
		w.WriteHeader(http.StatusOK)
		return nil
	}
	_, _ = w.Write(data)
	return nil
}

// cacheFileName builds the on-disk cache path for a media object - keyed by
// group+variant (content-addressed), never by storage key, so a re-upload
// of identical bytes naturally reuses the same cache entry.
func (e *MediaEndpoints) cacheFileName(o ports.MediaObject) string {
	return filepath.Join(e.cfg.CacheDir, fmt.Sprintf("%s_%s.webp", o.GroupID, o.Variant))
}

// readThroughCache serves o's bytes from the on-disk cache, downloading
// from object storage on a miss and best-effort writing the cache entry
// (a write failure never fails the request - the bytes are still served
// from what was already downloaded). Triggers a background eviction sweep
// after every miss, so the cache directory never grows unbounded.
func (e *MediaEndpoints) readThroughCache(ctx context.Context, o ports.MediaObject) ([]byte, error) {
	path := e.cacheFileName(o)
	if data, err := os.ReadFile(path); err == nil {
		now := time.Now()
		_ = os.Chtimes(path, now, now) // bump recency for the LRU sweep
		return data, nil
	}

	rc, _, err := e.pictures.Download(ctx, o.StorageKey)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rc.Close() }()
	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}

	if e.cfg.CacheDir != "" {
		if err := os.MkdirAll(e.cfg.CacheDir, 0o755); err != nil {
			log.Printf("media cache: creating %s: %v", e.cfg.CacheDir, err)
		} else if err := os.WriteFile(path, data, 0o644); err != nil {
			log.Printf("media cache: writing %s: %v", path, err)
		} else {
			go e.evictIfOversize()
		}
	}
	return data, nil
}

// evictIfOversize deletes the least-recently-used cache files until the
// directory is back under CacheMaxBytes. Runs in its own goroutine off the
// request path - a slow directory scan must never add latency to serving
// an image. Best-effort throughout: any error here is logged, never fatal.
func (e *MediaEndpoints) evictIfOversize() {
	if e.cfg.CacheMaxBytes <= 0 {
		return
	}
	entries, err := os.ReadDir(e.cfg.CacheDir)
	if err != nil {
		return
	}
	type fileInfo struct {
		path    string
		size    int64
		modTime time.Time
	}
	files := make([]fileInfo, 0, len(entries))
	var total int64
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil || entry.IsDir() {
			continue
		}
		files = append(files, fileInfo{path: filepath.Join(e.cfg.CacheDir, entry.Name()), size: info.Size(), modTime: info.ModTime()})
		total += info.Size()
	}
	if total <= e.cfg.CacheMaxBytes {
		return
	}
	sort.Slice(files, func(i, j int) bool { return files[i].modTime.Before(files[j].modTime) })
	for _, f := range files {
		if total <= e.cfg.CacheMaxBytes {
			break
		}
		if err := os.Remove(f.path); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				log.Printf("media cache: evicting %s: %v", f.path, err)
			}
			continue
		}
		total -= f.size
	}
}
