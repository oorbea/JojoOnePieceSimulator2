package endpoints

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// RateLimitConfig configures the tiered request limiter. Enabled=false turns
// limiting off entirely (useful for tests and local development); a zero or
// negative limit on an individual tier disables just that tier.
type RateLimitConfig struct {
	Enabled       bool
	Window        time.Duration
	GlobalPerIP   int
	LoginPerIP    int
	ReadPerUser   int
	WritePerUser  int
	TicketPerUser int
	// RefreshPerIP bounds POST /auth/refresh and /auth/logout, keyed by
	// client IP - both sit outside RequireAuth, so there is no user id to
	// key on before the token is redeemed.
	RefreshPerIP int
}

// keyByClientIP keys the limiter on chi's resolved client IP, populated by
// two middlewares chained in NewRouter: ClientIPFromRemoteAddr first (the
// TCP peer), then ClientIPFromXFF() with no trusted-prefix arguments, which
// overwrites it with the rightmost X-Forwarded-For entry when one is
// present. That "no arguments" form trusts exactly one hop in front of this
// server - correct here because prod's backend container publishes no port
// and is reachable only through Nginx Proxy Manager on the shared
// public-net network (docker-compose.prod.yml), which appends the real
// connecting IP to XFF on every request. A client-forged XFF value lands to
// the left of NPM's own append and is ignored. Not httprate.KeyByRealIP,
// which trusts X-Real-IP unconditionally with no such single-hop guarantee.
//
// This key func doesn't care which of the two middlewares won - it just
// reads whatever middleware.GetClientIP resolved. If the backend is ever
// exposed to more than one untrusted hop (a CDN in front of NPM, say),
// switch to middleware.ClientIPFromXFF(trustedPrefixes...) with NPM's actual
// address/CIDR instead of the argument-less form.
func keyByClientIP(r *http.Request) (string, error) {
	return httprate.CanonicalizeIP(middleware.GetClientIP(r.Context())), nil
}

// keyByUserID keys the limiter on the authenticated caller's user id, set by
// RequireAuth. Falls back to the client IP so this middleware stays safe if
// it is ever mounted outside a RequireAuth group.
func keyByUserID(r *http.Request) (string, error) {
	if claims, ok := claimsFromContext(r.Context()); ok {
		return claims.UserID.String(), nil
	}
	return keyByClientIP(r)
}

// limit builds a rate-limit middleware for one tier. requests <= 0 disables
// the tier (returns a no-op passthrough) so callers/config never need an
// extra branch. Rejected requests are routed through handleError with
// ports.ErrRateLimited, keeping the 429 response body identical to every
// other error in the API.
func limit(requests int, window time.Duration, keyFn httprate.KeyFunc) func(http.Handler) http.Handler {
	if requests <= 0 {
		return func(next http.Handler) http.Handler { return next }
	}
	return httprate.LimitBy(requests, window, keyFn,
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			handleError(w, ports.ErrRateLimited)
		}),
	)
}

// globalRateLimit applies to every route (see NewRouter), keyed by client IP.
func globalRateLimit(cfg RateLimitConfig) func(http.Handler) http.Handler {
	if !cfg.Enabled {
		return func(next http.Handler) http.Handler { return next }
	}
	return limit(cfg.GlobalPerIP, cfg.Window, keyByClientIP)
}

// loginRateLimit applies to POST /auth/google, keyed by client IP since the
// caller isn't authenticated yet.
func loginRateLimit(cfg RateLimitConfig) func(http.Handler) http.Handler {
	if !cfg.Enabled {
		return func(next http.Handler) http.Handler { return next }
	}
	return limit(cfg.LoginPerIP, cfg.Window, keyByClientIP)
}

// readRateLimit applies to the authenticated GET /stands routes, keyed by
// user id.
func readRateLimit(cfg RateLimitConfig) func(http.Handler) http.Handler {
	if !cfg.Enabled {
		return func(next http.Handler) http.Handler { return next }
	}
	return limit(cfg.ReadPerUser, cfg.Window, keyByUserID)
}

// writeRateLimit applies to the admin-only write routes on /stands, keyed by
// user id.
func writeRateLimit(cfg RateLimitConfig) func(http.Handler) http.Handler {
	if !cfg.Enabled {
		return func(next http.Handler) http.Handler { return next }
	}
	return limit(cfg.WritePerUser, cfg.Window, keyByUserID)
}

// ticketRateLimit applies to the two stream-ticket mint routes (POST
// /events/ticket, POST /games/{id}/ws-ticket), keyed by user id since both
// sit behind RequireAuth. Sized for a reconnect storm rather than steady
// state: every SSE/WS (re)connect now costs one mint, and the frontend's own
// backoff (2s -> 30s) bounds one tab's reconnect rate well under one per
// second even at its worst.
func ticketRateLimit(cfg RateLimitConfig) func(http.Handler) http.Handler {
	if !cfg.Enabled {
		return func(next http.Handler) http.Handler { return next }
	}
	return limit(cfg.TicketPerUser, cfg.Window, keyByUserID)
}

// refreshRateLimit applies to POST /auth/refresh and /auth/logout, keyed by
// client IP since both sit outside RequireAuth (the caller's identity comes
// from the refresh token/cookie itself, redeemed inside the handler, not
// from a bearer claim available beforehand).
func refreshRateLimit(cfg RateLimitConfig) func(http.Handler) http.Handler {
	if !cfg.Enabled {
		return func(next http.Handler) http.Handler { return next }
	}
	return limit(cfg.RefreshPerIP, cfg.Window, keyByClientIP)
}
