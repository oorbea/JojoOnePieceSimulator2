package endpoints

import (
	"net/http"
	"strings"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// csrfHeaderName is a custom request header that a cross-site form/fetch
// POST cannot set without triggering a CORS preflight (which then fails
// against our origin allowlist), unlike a "simple request".
const csrfHeaderName = "X-JOPS-Refresh"

// requireCSRFHeader defends the two cookie-authenticated POST routes
// (/auth/refresh, /auth/logout) against CSRF. SameSite=Strict is the primary
// defense; this is defense in depth for the case a cookie attribute ever
// regresses. A cross-site attacker cannot set a custom header without
// triggering a CORS preflight that fails against our origin allowlist, and
// cannot use application/x-www-form-urlencoded or multipart/form-data
// (the only content-types a preflight-free cross-site POST can use) because
// both are rejected here too.
func requireCSRFHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(csrfHeaderName) == "" {
			handleError(w, ports.ErrForbidden)
			return
		}
		ct := r.Header.Get("Content-Type")
		if strings.HasPrefix(ct, "application/x-www-form-urlencoded") || strings.HasPrefix(ct, "multipart/form-data") {
			handleError(w, ports.ErrForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
