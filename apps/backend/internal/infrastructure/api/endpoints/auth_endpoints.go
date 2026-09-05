package endpoints

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/dto"
)

// refreshTokenHeaderName carries the raw refresh token for native clients
// (which have no cookie jar) on POST /auth/refresh and /auth/logout.
const refreshTokenHeaderName = "X-Refresh-Token"

// refreshTransportHeaderName, when sent as "header", tells loginWithGoogle/
// refresh to also include the refresh token in the JSON body. Native
// clients send this; web clients don't (the cookie already carries it, and
// a web response body should never carry a token a script on the page
// could exfiltrate).
const refreshTransportHeaderName = "X-Refresh-Token-Transport"

// AuthEndpoints wires the Google login/registration HTTP surface, plus
// refresh-token rotation and logout, to the application service.
type AuthEndpoints struct {
	svc       *services.AuthService
	cookieCfg CookieConfig
}

func NewAuthEndpoints(svc *services.AuthService, cookieCfg CookieConfig) *AuthEndpoints {
	return &AuthEndpoints{svc: svc, cookieCfg: cookieCfg}
}

// Routes returns the /auth sub-router. Unlike /stands, these routes are
// public - they are how a caller obtains a token in the first place, so
// /google gets its own (tighter, IP-keyed) rate-limit tier rather than
// relying on the global one alone. /refresh and /logout are cookie-
// authenticated (or header-authenticated for native clients) rather than
// bearer-authenticated, so they sit outside RequireAuth too, behind their
// own rate-limit tier and the requireCSRFHeader defense (see csrf.go).
func (e *AuthEndpoints) Routes(rateCfg RateLimitConfig) chi.Router {
	r := chi.NewRouter()
	r.With(loginRateLimit(rateCfg)).Post("/google", Wrap(e.loginWithGoogle))
	r.With(refreshRateLimit(rateCfg), requireCSRFHeader).Post("/refresh", Wrap(e.refresh))
	r.With(refreshRateLimit(rateCfg), requireCSRFHeader).Post("/logout", Wrap(e.logout))
	return r
}

// wantsRefreshTokenInBody reports whether the caller opted into header
// transport for the refresh token (native clients) rather than the default
// cookie-only transport (web clients).
func wantsRefreshTokenInBody(r *http.Request) bool {
	return r.Header.Get(refreshTransportHeaderName) == "header"
}

// writeLoginResponse sets the refresh-token cookie and writes status/body
// for a LoginResult, shared by loginWithGoogle and refresh so both stay
// consistent about the header-vs-cookie transport rule.
func (e *AuthEndpoints) writeLoginResponse(w http.ResponseWriter, r *http.Request, status int, result *services.LoginResult) error {
	setRefreshCookie(w, e.cookieCfg, result.RefreshToken, result.RefreshExpiresAt)

	userResp, err := dto.NewUserResponse(r.Context(), result.User, e.svc.PictureURL)
	if err != nil {
		return err
	}

	resp := dto.LoginResponse{
		AccessToken: result.AccessToken,
		TokenType:   "Bearer",
		ExpiresAt:   result.ExpiresAt,
		User:        userResp,
	}
	if wantsRefreshTokenInBody(r) {
		resp.RefreshToken = result.RefreshToken
	}

	writeJSON(w, status, resp)
	return nil
}

// refreshTokenFromRequest reads the refresh token from, in order: the
// X-Refresh-Token header (native clients), then the AUTH_COOKIE_NAME cookie
// (web clients). Returns ports.ErrUnauthenticated if neither is present.
func (e *AuthEndpoints) refreshTokenFromRequest(r *http.Request) (string, error) {
	if raw := r.Header.Get(refreshTokenHeaderName); raw != "" {
		return raw, nil
	}
	if cookie, err := r.Cookie(e.cookieCfg.Name); err == nil && cookie.Value != "" {
		return cookie.Value, nil
	}
	return "", ports.ErrUnauthenticated
}

// loginWithGoogle godoc
//
//	@Summary		Log in or register via Google
//	@Description	Verifies a Google ID token and returns an access token, creating the User the first time this Google account is seen. Also mints a refresh token, set as an HttpOnly cookie (or, with X-Refresh-Token-Transport: header, returned in the body instead).
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.GoogleLoginRequest	true	"Google ID token"
//	@Param			X-Refresh-Token-Transport	header	string	false	"Set to \"header\" to receive the refresh token in the response body instead of a cookie (native clients)"
//	@Success		200		{object}	dto.LoginResponse	"existing user logged in"
//	@Success		201		{object}	dto.LoginResponse	"new user registered"
//	@Header			200		{string}	Set-Cookie	"jops_rt=<refresh token>; Path=/api/v1/auth; HttpOnly; Secure; SameSite=Strict"
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

	return e.writeLoginResponse(w, r, status, result)
}

// refresh godoc
//
//	@Summary		Rotate a refresh token for a new access token
//	@Description	Redeems the caller's refresh token (from X-Refresh-Token or the refresh cookie), returning a new access token and a rotated refresh token. A replayed (already-used) refresh token is rejected and revokes its whole token family server-side.
//	@Tags			auth
//	@Produce		json
//	@Param			X-Refresh-Token	header	string	false	"Refresh token (native clients); omit to use the refresh cookie instead"
//	@Param			X-Refresh-Token-Transport	header	string	false	"Set to \"header\" to receive the refresh token in the response body instead of a cookie (native clients)"
//	@Success		200		{object}	dto.LoginResponse
//	@Header			200		{string}	Set-Cookie	"jops_rt=<rotated refresh token>; Path=/api/v1/auth; HttpOnly; Secure; SameSite=Strict"
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		403		{object}	dto.ErrorResponse	"missing X-JOPS-Refresh CSRF header, or a form-encoded body"
//	@Failure		429		{object}	dto.ErrorResponse
//	@Router			/auth/refresh [post]
func (e *AuthEndpoints) refresh(w http.ResponseWriter, r *http.Request) error {
	token, err := e.refreshTokenFromRequest(r)
	if err != nil {
		return err
	}

	result, err := e.svc.Refresh(r.Context(), token)
	if err != nil {
		return err
	}

	return e.writeLoginResponse(w, r, http.StatusOK, result)
}

// logout godoc
//
//	@Summary		Log out
//	@Description	Revokes the caller's whole refresh-token family and clears the refresh cookie. Always succeeds, even if the token was already invalid.
//	@Tags			auth
//	@Success		204	"logged out"
//	@Header			204	{string}	Set-Cookie	"jops_rt=; Path=/api/v1/auth; HttpOnly; Secure; SameSite=Strict; Max-Age=-1"
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		403	{object}	dto.ErrorResponse	"missing X-JOPS-Refresh CSRF header, or a form-encoded body"
//	@Failure		429	{object}	dto.ErrorResponse
//	@Router			/auth/logout [post]
func (e *AuthEndpoints) logout(w http.ResponseWriter, r *http.Request) error {
	token, err := e.refreshTokenFromRequest(r)
	if err != nil {
		return err
	}

	_ = e.svc.Logout(r.Context(), token)

	clearRefreshCookie(w, e.cookieCfg)
	writeJSON(w, http.StatusNoContent, nil)
	return nil
}
