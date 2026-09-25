package endpoints

import (
	"net"
	"net/http"
)

// proxyHeaders lists every header a reverse proxy in front of this backend
// is known to add (see ratelimit.go's keyByClientIP doc: prod's NPM always
// appends to X-Forwarded-For). requireLocalRequest treats the presence of
// any of these as proof the request did not arrive directly, regardless of
// what the header actually says - a caller could forge them, but a direct
// local caller (curl, the dev frontend) never sends them at all.
var proxyHeaders = []string{"X-Forwarded-For", "Forwarded", "X-Real-IP", "CF-Connecting-IP"}

// requireLocalRequest gates POST /auth/dev-login (only mounted at all when
// config.Config.DevAuthBypass is set - see AuthEndpoints.Routes) behind two
// checks: the TCP peer (r.RemoteAddr, deliberately not the
// XFF-aware middleware.GetClientIP used elsewhere) must be loopback or a
// private-range address, and the request must carry none of proxyHeaders.
// In prod every request arrives through Nginx Proxy Manager, which always
// appends to X-Forwarded-For (see ratelimit.go), so this combination is
// unreachable there even if DEV_AUTH_BYPASS were somehow set (config.Load's
// boot guard is the primary defense; this is defense in depth on the route
// itself).
//
// A failed check 404s rather than 403s - the route's existence is not
// revealed to a caller that shouldn't be talking to it at all.
func requireLocalRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isLocalRequest(r) {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func isLocalRequest(r *http.Request) bool {
	for _, h := range proxyHeaders {
		if r.Header.Get(h) != "" {
			return false
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// RemoteAddr with no port (unusual, but not proof of anything) -
		// fall back to treating the whole value as the host.
		host = r.RemoteAddr
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	return ip.IsLoopback() || ip.IsPrivate()
}
