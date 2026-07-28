package endpoints

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
)

// CacheConfig configures the ETag/Cache-Control layer applied to the read
// endpoints. HTTPMaxAge <= 0 disables Cache-Control entirely (ETag/304
// handling still applies, since it costs nothing extra to keep on).
type CacheConfig struct {
	HTTPMaxAge int // seconds
}

// bufferingResponseWriter buffers a handler's body instead of writing it
// straight through, so cacheHeaders can compute an ETag and possibly
// short-circuit to a 304 before anything is sent to the client. writeJSON
// (respond.go) always calls WriteHeader itself (even for a nil body), so
// status is always observed before Flush.
type bufferingResponseWriter struct {
	http.ResponseWriter
	status int
	body   bytes.Buffer
}

func (w *bufferingResponseWriter) WriteHeader(status int) {
	w.status = status
}

func (w *bufferingResponseWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

// flush sends whatever was buffered through to the real ResponseWriter
// unchanged - used for every response cacheHeaders doesn't itself rewrite.
func (w *bufferingResponseWriter) flush() {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	w.ResponseWriter.WriteHeader(w.status)
	_, _ = w.ResponseWriter.Write(w.body.Bytes())
}

// cacheHeaders adds an ETag (a hash of the response body) and, when
// cfg.HTTPMaxAge > 0, a Cache-Control header to successful GET responses,
// and answers a matching If-None-Match with an empty 304 instead of
// resending the body. Only GET + status 200 are touched; every other
// response (errors, 202/204/etc.) passes through untouched.
//
// Cache-Control is deliberately "private" and paired with Vary: Authorization
// - these are Bearer-authenticated, per-caller responses (RequireAuth sits in
// front of every route this is mounted on) that must never be stored by a
// shared proxy or CDN.
func cacheHeaders(cfg CacheConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				next.ServeHTTP(w, r)
				return
			}

			bw := &bufferingResponseWriter{ResponseWriter: w}
			next.ServeHTTP(bw, r)

			if bw.status != http.StatusOK {
				bw.flush()
				return
			}

			sum := sha256.Sum256(bw.body.Bytes())
			etag := `"` + hex.EncodeToString(sum[:]) + `"`

			w.Header().Set("ETag", etag)
			w.Header().Set("Vary", "Authorization")
			if cfg.HTTPMaxAge > 0 {
				w.Header().Set("Cache-Control", "private, max-age="+strconv.Itoa(cfg.HTTPMaxAge))
			} else {
				w.Header().Set("Cache-Control", "private, no-cache")
			}

			if match := r.Header.Get("If-None-Match"); match != "" && etagMatches(match, etag) {
				w.WriteHeader(http.StatusNotModified)
				return
			}

			w.Header().Set("Content-Length", strconv.Itoa(bw.body.Len()))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(bw.body.Bytes())
		})
	}
}

// etagMatches reports whether etag appears among the comma-separated list
// of entity tags in the If-None-Match header value ifNoneMatch, honoring
// the wildcard "*" (matches anything).
func etagMatches(ifNoneMatch, etag string) bool {
	if ifNoneMatch == "*" {
		return true
	}
	for _, candidate := range strings.Split(ifNoneMatch, ",") {
		if strings.TrimSpace(candidate) == etag {
			return true
		}
	}
	return false
}
