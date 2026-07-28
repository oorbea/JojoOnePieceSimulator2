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
