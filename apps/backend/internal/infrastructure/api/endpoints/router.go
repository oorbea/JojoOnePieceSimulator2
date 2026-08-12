package endpoints

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	_ "github.com/oorbea/JojoOnePieceSimulator2/docs" // registers the generated swagger spec
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// CORSConfig configures the CORS middleware. AllowedOrigins is deny-all
// (the middleware isn't even added, so no Access-Control-* headers are ever
// sent) when empty - that's the safe default for an unconfigured deployment.
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// NewRouter builds the full HTTP handler for the backend: common middleware,
// a health check, Swagger UI, and the versioned API routes. /auth and
// /swagger are public; everything else requires a valid access token. Every
// route, including /health and /swagger, sits behind the global rate limit
// tier; /auth/google, /stands, and /devil-fruits additionally get their own
// tighter tiers (see rateCfg and ratelimit.go). cacheCfg configures the
// ETag/Cache-Control layer applied to the Stand and DevilFruit read routes
// (see cache_headers.go).
//
// middleware.Timeout(60s) is applied per-group rather than globally: every
// normal route keeps that bound, but /api/v1/events (events_endpoints.go)
// and /api/v1/games/{id}/ws (game_ws_endpoints.go) are deliberately
// long-lived streams and must not inherit it. /events relies on the
// client's own reconnect logic, a server-side heartbeat, and the app's
// shutdown context; the game WebSocket does the same, plus its own
// protocol-level ping. GameEndpoints.Routes applies Timeout+RequireAuth to
// its own REST sub-group internally (chi can't mount two handlers on the
// same pattern with different middleware), so /games is mounted here
// alongside /events, outside this file's own Timeout group.
func NewRouter(authEndpoints *AuthEndpoints, standEndpoints *StandEndpoints, devilFruitEndpoints *DevilFruitEndpoints, userEndpoints *UserEndpoints, eventsEndpoints *EventsEndpoints, gameEndpoints *GameEndpoints, stageEndpoints *StageEndpoints, issuer ports.ITokenIssuer, corsCfg CORSConfig, rateCfg RateLimitConfig, cacheCfg CacheConfig) http.Handler {
	r := chi.NewRouter()

	if len(corsCfg.AllowedOrigins) > 0 {
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   corsCfg.AllowedOrigins,
			AllowedMethods:   corsCfg.AllowedMethods,
			AllowedHeaders:   corsCfg.AllowedHeaders,
			AllowCredentials: corsCfg.AllowCredentials,
			MaxAge:           corsCfg.MaxAge,
		}))
	}

	r.Use(middleware.RequestID)
	r.Use(middleware.ClientIPFromRemoteAddr)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(globalRateLimit(rateCfg))
	r.Use(resolveLocale)

	r.Group(func(r chi.Router) {
		r.Use(middleware.Timeout(60 * time.Second))

		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		})

		r.Get("/swagger/*", httpSwagger.WrapHandler)
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(middleware.Timeout(60 * time.Second))

			r.Mount("/auth", authEndpoints.Routes(rateCfg))

			r.Group(func(r chi.Router) {
				r.Use(RequireAuth(issuer))
				r.Mount("/stands", standEndpoints.Routes(rateCfg, cacheCfg))
				r.Mount("/devil-fruits", devilFruitEndpoints.Routes(rateCfg, cacheCfg))
				r.Mount("/users", userEndpoints.Routes(rateCfg))
				r.Mount("/stages", stageEndpoints.Routes(rateCfg, cacheCfg))
			})
		})

		r.Mount("/events", eventsEndpoints.Routes())
		// Outside the Timeout group on purpose - see this function's doc.
		r.Mount("/games", gameEndpoints.Routes(rateCfg))
	})

	return r
}
