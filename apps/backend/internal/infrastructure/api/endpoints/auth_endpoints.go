package endpoints

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/dto"
)

// AuthEndpoints wires the Google login/registration HTTP surface to the
// application service.
type AuthEndpoints struct {
	svc *services.AuthService
}

func NewAuthEndpoints(svc *services.AuthService) *AuthEndpoints {
	return &AuthEndpoints{svc: svc}
}

// Routes returns the /auth sub-router. Unlike /stands, these routes are
// public - they are how a caller obtains a token in the first place, so
// /google gets its own (tighter, IP-keyed) rate-limit tier rather than
// relying on the global one alone.
func (e *AuthEndpoints) Routes(rateCfg RateLimitConfig) chi.Router {
	r := chi.NewRouter()
	r.With(loginRateLimit(rateCfg)).Post("/google", Wrap(e.loginWithGoogle))
	return r
}

// loginWithGoogle godoc
//
//	@Summary		Log in or register via Google
//	@Description	Verifies a Google ID token and returns an access token, creating the User the first time this Google account is seen.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.GoogleLoginRequest	true	"Google ID token"
//	@Success		200		{object}	dto.LoginResponse	"existing user logged in"
//	@Success		201		{object}	dto.LoginResponse	"new user registered"
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		429		{object}	dto.ErrorResponse
//	@Router			/auth/google [post]
func (e *AuthEndpoints) loginWithGoogle(w http.ResponseWriter, r *http.Request) error {
	var req dto.GoogleLoginRequest
	if err := decode(w, r, &req); err != nil {
		return err
	}
	if err := req.Validate(); err != nil {
		return err
	}

	result, err := e.svc.LoginWithGoogle(r.Context(), req.IDToken)
	if err != nil {
		return err
	}

	status := http.StatusOK
	if result.Registered {
		status = http.StatusCreated
	}

	writeJSON(w, status, dto.LoginResponse{
		AccessToken: result.AccessToken,
		TokenType:   "Bearer",
		ExpiresAt:   result.ExpiresAt,
		User:        dto.NewUserResponse(result.User),
	})
	return nil
}
