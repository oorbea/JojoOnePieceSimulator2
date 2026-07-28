package endpoints

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// NewRouter builds the full HTTP handler for the backend: common middleware,
// a health check, and the versioned API routes. /auth is public (it's how a
// caller gets a token); everything else requires a valid access token.
func NewRouter(authEndpoints *AuthEndpoints, standEndpoints *StandEndpoints, issuer ports.ITokenIssuer) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.ClientIPFromRemoteAddr)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Mount("/auth", authEndpoints.Routes())

		r.Group(func(r chi.Router) {
			r.Use(RequireAuth(issuer))
			r.Mount("/stands", standEndpoints.Routes())
		})
	})

	return r
}
