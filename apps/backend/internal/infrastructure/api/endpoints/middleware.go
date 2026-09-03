package endpoints

import (
	"net/http"
	"strings"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

const bearerPrefix = "Bearer "

// RequireAuth parses the Authorization: Bearer <token> header with issuer
// and stores the resulting claims on the request context. Missing or
// invalid tokens fail as ports.ErrUnauthenticated, without revealing why
// the token was rejected.
func RequireAuth(issuer ports.ITokenIssuer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, bearerPrefix) {
				handleError(w, ports.ErrUnauthenticated)
				return
			}

			rawToken := strings.TrimSpace(header[len(bearerPrefix):])
			claims, err := issuer.Parse(rawToken)
			if err != nil {
				handleError(w, ports.ErrUnauthenticated)
				return
			}

			next.ServeHTTP(w, r.WithContext(withClaims(r.Context(), claims)))
		})
	}
}

// RequireAdmin rejects any request whose claims (set by RequireAuth, which
// must run first) do not carry the ADMIN role.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := claimsFromContext(r.Context())
		if !ok || claims.Role != enums.Admin {
			handleError(w, ports.ErrForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// authenticateStream authenticates a long-lived stream connection
// (EventSource or WebSocket). The Authorization: Bearer header is tried
// first and is full authority - native clients, curl, and tests can always
// use it. A browser cannot set headers on either connection type, so it
// instead presents ?ticket=<opaque>, minted seconds earlier through a
// normal authenticated POST (see EventsEndpoints.mintTicket/
// GameEndpoints.mintWSTicket): single-use, and bound to purpose/resource so
// a ticket minted for one stream (or one game) can never open another.
// Redeem burns the ticket regardless of the outcome below, so a
// wrong-purpose or wrong-resource ticket cannot be retried.
//
// There is deliberately no ?token=<jwt> fallback: a full access token must
// never appear in a URL (and therefore in access logs, which is exactly the
// weakness this replaces). A stale client still sending ?token= simply
// finds no ?ticket= and falls through to ErrUnauthenticated.
func authenticateStream(
	r *http.Request,
	issuer ports.ITokenIssuer,
	tickets ports.IStreamTicketStore,
	purpose ports.TicketPurpose,
	resource string,
) (ports.Claims, error) {
	if header := r.Header.Get("Authorization"); strings.HasPrefix(header, bearerPrefix) {
		return issuer.Parse(strings.TrimSpace(header[len(bearerPrefix):]))
	}

	raw := r.URL.Query().Get("ticket")
	if raw == "" {
		return ports.Claims{}, ports.ErrUnauthenticated
	}
	t, err := tickets.Redeem(r.Context(), raw)
	if err != nil {
		return ports.Claims{}, err
	}
	if t.Purpose != purpose || t.Resource != resource {
		return ports.Claims{}, ports.ErrTicketInvalid
	}
	return ports.Claims{UserID: t.UserID, Role: t.Role}, nil
}
