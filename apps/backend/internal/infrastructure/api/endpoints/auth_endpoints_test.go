package endpoints_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/endpoints"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/idgen"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/refreshtoken"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/streamticket"
)

// testCookieCfg mirrors the config.Config defaults so cookie assertions
// below pin real production attributes, not zero values.
var testCookieCfg = endpoints.CookieConfig{
	Name:     "jops_rt",
	Path:     "/api/v1/auth",
	Secure:   true,
	SameSite: http.SameSiteStrictMode,
}

// POST /api/v1/auth/google is the only public write route in the API - it is
// how a caller gets a token in the first place - so the two things worth
// pinning here are the 200-vs-201 split (an existing login must not be
// reported as a registration, and vice versa: the frontend branches on it)
// and that a malformed body never reaches the service at all.

// fakeGoogleVerifier stands in for the real GoogleVerifier so these tests
// never touch google.golang.org/api/idtoken (which would need Google's live
// signing keys - see google_verifier_test.go's header for why that can't be
// faked at the library level).
type fakeGoogleVerifier struct {
	identity ports.GoogleIdentity
	err      error
	calls    int
}

func (f *fakeGoogleVerifier) Verify(_ context.Context, _ string) (ports.GoogleIdentity, error) {
	f.calls++
	if f.err != nil {
		return ports.GoogleIdentity{}, f.err
	}
	return f.identity, nil
}

var _ ports.IGoogleTokenVerifier = (*fakeGoogleVerifier)(nil)

// loginTokenIssuer, unlike the package's fakeTokenIssuer (stand_endpoints_test.go),
// can actually Issue - AuthService needs a working issuer to build a login
// response. Parse is delegated so a token minted here is still accepted by
// RequireAuth on the rest of the routes.
type loginTokenIssuer struct{}

func (loginTokenIssuer) Issue(u *user.User) (string, time.Time, error) {
	return "issued-for-" + u.ID().String(), time.Now().Add(time.Hour), nil
}

func (loginTokenIssuer) Parse(token string) (ports.Claims, error) {
	return fakeTokenIssuer{}.Parse(token)
}

var _ ports.ITokenIssuer = loginTokenIssuer{}

func newAuthTestServer(repo *fakeUserRepo, verifier ports.IGoogleTokenVerifier, adminEmails []string) http.Handler {
	refreshStore := refreshtoken.NewMemoryStore(refreshtoken.Config{TTL: time.Hour})
	svc := services.NewAuthService(repo, idgen.UUIDGenerator[user.UserID]{}, verifier, loginTokenIssuer{}, refreshStore, adminEmails, newFakePictureStorage())

	// Everything but /auth is wired bare: these tests only exercise the
	// public login/refresh/logout routes, but going through NewRouter keeps
	// the assertion that /auth/google really is reachable without a bearer
	// token honest.
	return endpoints.NewRouter(
		endpoints.NewAuthEndpoints(svc, testCookieCfg),
		endpoints.NewStandEndpoints(nil),
		endpoints.NewDevilFruitEndpoints(nil),
		endpoints.NewUserEndpoints(nil),
		endpoints.NewEventsEndpoints(services.NewPictureEventHub(), fakeTokenIssuer{}, streamticket.NewMemoryStore(streamticket.Config{TTL: 30 * time.Second}), context.Background()),
		endpoints.NewGameEndpoints(nil, services.NewGameEventHub(), nil, nil, nil, nil, fakeTokenIssuer{}, streamticket.NewMemoryStore(streamticket.Config{TTL: 30 * time.Second}), context.Background(), endpoints.GameWSConfig{}),
		endpoints.NewStageEndpoints(nil),
		fakeTokenIssuer{}, endpoints.CORSConfig{}, endpoints.RateLimitConfig{}, endpoints.CacheConfig{})
}

func verifiedIdentity() ports.GoogleIdentity {
	return ports.GoogleIdentity{
		Subject:       "google-sub-jotaro",
		Email:         "jotaro@example.com",
		EmailVerified: true,
		Name:          "Jotaro Kujo",
		Picture:       "https://google.example/pic.png",
	}
}

// postRaw sends body verbatim, without marshalling it first, so tests can
// send bodies that aren't valid JSON at all.
func postRaw(t *testing.T, h http.Handler, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestPostAuthGoogle_FirstLogin_Returns201AndRegisters(t *testing.T) {
	repo := newFakeUserRepo()
	verifier := &fakeGoogleVerifier{identity: verifiedIdentity()}
	h := newAuthTestServer(repo, verifier, nil)

	rec := postRaw(t, h, "/api/v1/auth/google", `{"idToken":"a-google-id-token"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got["tokenType"] != "Bearer" {
		t.Errorf("tokenType = %v, want Bearer", got["tokenType"])
	}
	if token, _ := got["accessToken"].(string); token == "" {
		t.Error("accessToken must not be empty")
	}
	userResp, _ := got["user"].(map[string]any)
	if userResp["email"] != "jotaro@example.com" {
		t.Errorf("user.email = %v, want jotaro@example.com", userResp["email"])
	}
	// Username is derived from the email's local part on registration.
	if userResp["username"] != "jotaro" {
		t.Errorf("user.username = %v, want jotaro", userResp["username"])
	}
	if userResp["role"] != enums.Regular.String() {
		t.Errorf("user.role = %v, want %v", userResp["role"], enums.Regular)
	}

	if _, err := repo.FindByGoogleSub(context.Background(), "google-sub-jotaro"); err != nil {
		t.Errorf("registered user not persisted: %v", err)
	}
}

func TestPostAuthGoogle_SecondLogin_Returns200NotCreated(t *testing.T) {
	repo := newFakeUserRepo()
	verifier := &fakeGoogleVerifier{identity: verifiedIdentity()}
	h := newAuthTestServer(repo, verifier, nil)

	if rec := postRaw(t, h, "/api/v1/auth/google", `{"idToken":"a-google-id-token"}`); rec.Code != http.StatusCreated {
		t.Fatalf("first login status = %d, want %d, body = %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	rec := postRaw(t, h, "/api/v1/auth/google", `{"idToken":"a-google-id-token"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("second login status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	// Same Google account twice must never produce a second User.
	all, err := repo.List(context.Background(), 100, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(all) != 1 {
		t.Errorf("user count = %d, want 1", len(all))
	}
}

// An email in ADMIN_EMAILS is what promotes a caller to ADMIN, so the 201
// branch has to honour it on the very first login too, not only on a later
// re-sync.
func TestPostAuthGoogle_AdminEmail_RegistersAsAdmin(t *testing.T) {
	repo := newFakeUserRepo()
	verifier := &fakeGoogleVerifier{identity: verifiedIdentity()}
	// Matched case-insensitively, hence the deliberately mixed case here.
	h := newAuthTestServer(repo, verifier, []string{"  JoTaRo@Example.com "})

	rec := postRaw(t, h, "/api/v1/auth/google", `{"idToken":"a-google-id-token"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	var got map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &got)
	userResp, _ := got["user"].(map[string]any)
	if userResp["role"] != enums.Admin.String() {
		t.Errorf("user.role = %v, want %v", userResp["role"], enums.Admin)
	}
}

func TestPostAuthGoogle_InvalidBody_Returns400WithoutVerifying(t *testing.T) {
	cases := map[string]string{
		"missing idToken":  `{}`,
		"empty idToken":    `{"idToken":""}`,
		"null idToken":     `{"idToken":null}`,
		"not json":         `not json at all`,
		"empty body":       ``,
		"wrong type":       `{"idToken":123}`,
		"unknown field":    `{"idToken":"tok","role":"ADMIN"}`,
		"json array":       `[{"idToken":"tok"}]`,
		"truncated object": `{"idToken":"tok"`,
	}
	for name, body := range cases {
		repo := newFakeUserRepo()
		verifier := &fakeGoogleVerifier{identity: verifiedIdentity()}
		h := newAuthTestServer(repo, verifier, nil)

		rec := postRaw(t, h, "/api/v1/auth/google", body)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d, want %d, body = %s", name, rec.Code, http.StatusBadRequest, rec.Body.String())
		}
		if verifier.calls != 0 {
			t.Errorf("%s: verifier was called %d times, want 0 - a rejected body must never reach the service", name, verifier.calls)
		}
	}
}

func TestPostAuthGoogle_RejectedToken_Returns401(t *testing.T) {
	repo := newFakeUserRepo()
	verifier := &fakeGoogleVerifier{err: ports.ErrInvalidGoogleToken}
	h := newAuthTestServer(repo, verifier, nil)

	rec := postRaw(t, h, "/api/v1/auth/google", `{"idToken":"forged"}`)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
	// The 401 body must not leak why the token was rejected.
	if body := rec.Body.String(); !strings.Contains(body, "unauthenticated") {
		t.Errorf("body = %s, want the generic \"unauthenticated\" message", body)
	}
	all, _ := repo.List(context.Background(), 100, 0)
	if len(all) != 0 {
		t.Errorf("user count = %d, want 0 - a rejected token must never register anyone", len(all))
	}
}

// EmailVerified is reported by the verifier but enforced by AuthService: a
// Google account whose email is unverified must not get an account here.
func TestPostAuthGoogle_UnverifiedEmail_Returns400(t *testing.T) {
	repo := newFakeUserRepo()
	identity := verifiedIdentity()
	identity.EmailVerified = false
	h := newAuthTestServer(repo, &fakeGoogleVerifier{identity: identity}, nil)

	rec := postRaw(t, h, "/api/v1/auth/google", `{"idToken":"a-google-id-token"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	all, _ := repo.List(context.Background(), 100, 0)
	if len(all) != 0 {
		t.Errorf("user count = %d, want 0", len(all))
	}
}

// The login route is the one place in the API that must work without a
// token; a regression that put it behind RequireAuth would lock everyone out.
func TestPostAuthGoogle_NeedsNoBearerToken(t *testing.T) {
	repo := newFakeUserRepo()
	h := newAuthTestServer(repo, &fakeGoogleVerifier{identity: verifiedIdentity()}, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/google", bytes.NewBufferString(`{"idToken":"a-google-id-token"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code == http.StatusUnauthorized {
		t.Fatalf("POST /auth/google answered 401 without a bearer token, body = %s", rec.Body.String())
	}
}

// findCookie returns the cookie named name from rec's Set-Cookie headers,
// or nil if not present.
func findCookie(rec *httptest.ResponseRecorder, name string) *http.Cookie {
	for _, c := range (&http.Response{Header: rec.Header()}).Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}

// doAuthPost sends a POST to path with the CSRF header set (and, if cookie
// is non-nil, that cookie attached) plus a JSON content type - the shape
// every /auth/refresh and /auth/logout test needs.
func doAuthPost(h http.Handler, path string, cookie *http.Cookie, extraHeaders map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, nil)
	req.Header.Set("X-JOPS-Refresh", "1")
	if cookie != nil {
		req.AddCookie(cookie)
	}
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestPostAuthGoogle_DefaultTransport_SetsCookieAndOmitsBodyToken(t *testing.T) {
	repo := newFakeUserRepo()
	h := newAuthTestServer(repo, &fakeGoogleVerifier{identity: verifiedIdentity()}, nil)

	rec := postRaw(t, h, "/api/v1/auth/google", `{"idToken":"a-google-id-token"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	cookie := findCookie(rec, testCookieCfg.Name)
	if cookie == nil {
		t.Fatal("no refresh cookie set")
	}
	if cookie.Path != testCookieCfg.Path {
		t.Errorf("cookie Path = %q, want %q", cookie.Path, testCookieCfg.Path)
	}
	if !cookie.HttpOnly {
		t.Error("cookie must be HttpOnly")
	}
	if !cookie.Secure {
		t.Error("cookie must be Secure")
	}
	if cookie.SameSite != http.SameSiteStrictMode {
		t.Errorf("cookie SameSite = %v, want Strict", cookie.SameSite)
	}
	if cookie.Value == "" {
		t.Error("cookie value must not be empty")
	}

	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, present := got["refreshToken"]; present {
		t.Errorf("body must not carry refreshToken by default, got %v", got["refreshToken"])
	}
}

func TestPostAuthGoogle_HeaderTransport_IncludesRefreshTokenInBody(t *testing.T) {
	repo := newFakeUserRepo()
	h := newAuthTestServer(repo, &fakeGoogleVerifier{identity: verifiedIdentity()}, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/google", bytes.NewBufferString(`{"idToken":"a-google-id-token"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Refresh-Token-Transport", "header")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	refreshToken, _ := got["refreshToken"].(string)
	if refreshToken == "" {
		t.Error("body must carry refreshToken with X-Refresh-Token-Transport: header")
	}
}

func TestPostAuthRefresh_ViaCookie_Rotates(t *testing.T) {
	repo := newFakeUserRepo()
	h := newAuthTestServer(repo, &fakeGoogleVerifier{identity: verifiedIdentity()}, nil)

	loginRec := postRaw(t, h, "/api/v1/auth/google", `{"idToken":"a-google-id-token"}`)
	cookie := findCookie(loginRec, testCookieCfg.Name)
	if cookie == nil {
		t.Fatal("no refresh cookie set on login")
	}

	rec := doAuthPost(h, "/api/v1/auth/refresh", cookie, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("refresh status = %d, body = %s", rec.Code, rec.Body.String())
	}

	rotated := findCookie(rec, testCookieCfg.Name)
	if rotated == nil {
		t.Fatal("refresh did not set a rotated cookie")
	}
	if rotated.Value == cookie.Value {
		t.Error("refresh must rotate the token, not reuse it")
	}
}

func TestPostAuthRefresh_ViaHeader_Works(t *testing.T) {
	repo := newFakeUserRepo()
	h := newAuthTestServer(repo, &fakeGoogleVerifier{identity: verifiedIdentity()}, nil)

	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/google", bytes.NewBufferString(`{"idToken":"a-google-id-token"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginReq.Header.Set("X-Refresh-Token-Transport", "header")
	loginRec := httptest.NewRecorder()
	h.ServeHTTP(loginRec, loginReq)

	var loginBody map[string]any
	if err := json.Unmarshal(loginRec.Body.Bytes(), &loginBody); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	rawToken, _ := loginBody["refreshToken"].(string)
	if rawToken == "" {
		t.Fatal("no refreshToken in login body")
	}

	rec := doAuthPost(h, "/api/v1/auth/refresh", nil, map[string]string{"X-Refresh-Token": rawToken})
	if rec.Code != http.StatusOK {
		t.Fatalf("refresh via header status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestPostAuthRefresh_ReplayedToken_Returns401AndKillsFamily(t *testing.T) {
	repo := newFakeUserRepo()
	h := newAuthTestServer(repo, &fakeGoogleVerifier{identity: verifiedIdentity()}, nil)

	loginRec := postRaw(t, h, "/api/v1/auth/google", `{"idToken":"a-google-id-token"}`)
	cookie := findCookie(loginRec, testCookieCfg.Name)

	first := doAuthPost(h, "/api/v1/auth/refresh", cookie, nil)
	if first.Code != http.StatusOK {
		t.Fatalf("first refresh status = %d, body = %s", first.Code, first.Body.String())
	}
	rotated := findCookie(first, testCookieCfg.Name)

	// Replaying the already-burned original token must fail...
	replay := doAuthPost(h, "/api/v1/auth/refresh", cookie, nil)
	if replay.Code != http.StatusUnauthorized {
		t.Fatalf("replayed refresh status = %d, want 401, body = %s", replay.Code, replay.Body.String())
	}

	// ...and kill the whole family, so even the legitimately rotated token
	// from the same family now fails too.
	legit := doAuthPost(h, "/api/v1/auth/refresh", rotated, nil)
	if legit.Code != http.StatusUnauthorized {
		t.Fatalf("legit refresh after replay status = %d, want 401, body = %s", legit.Code, legit.Body.String())
	}
}

func TestPostAuthLogout_ClearsCookieAndRevokesToken(t *testing.T) {
	repo := newFakeUserRepo()
	h := newAuthTestServer(repo, &fakeGoogleVerifier{identity: verifiedIdentity()}, nil)

	loginRec := postRaw(t, h, "/api/v1/auth/google", `{"idToken":"a-google-id-token"}`)
	cookie := findCookie(loginRec, testCookieCfg.Name)

	rec := doAuthPost(h, "/api/v1/auth/logout", cookie, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("logout status = %d, body = %s", rec.Code, rec.Body.String())
	}

	cleared := findCookie(rec, testCookieCfg.Name)
	if cleared == nil {
		t.Fatal("logout did not clear the cookie")
	}
	if cleared.Path != testCookieCfg.Path || cleared.HttpOnly != true || cleared.Secure != true || cleared.SameSite != http.SameSiteStrictMode {
		t.Errorf("cleared cookie attributes mismatch setter: %+v", cleared)
	}
	if cleared.MaxAge >= 0 {
		t.Errorf("cleared cookie MaxAge = %d, want negative", cleared.MaxAge)
	}

	// A replay of the now-logged-out token must fail.
	replay := doAuthPost(h, "/api/v1/auth/refresh", cookie, nil)
	if replay.Code != http.StatusUnauthorized {
		t.Fatalf("refresh after logout status = %d, want 401, body = %s", replay.Code, replay.Body.String())
	}
}

func TestPostAuthRefresh_MissingCSRFHeader_Returns403(t *testing.T) {
	repo := newFakeUserRepo()
	h := newAuthTestServer(repo, &fakeGoogleVerifier{identity: verifiedIdentity()}, nil)

	loginRec := postRaw(t, h, "/api/v1/auth/google", `{"idToken":"a-google-id-token"}`)
	cookie := findCookie(loginRec, testCookieCfg.Name)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403, body = %s", rec.Code, rec.Body.String())
	}
}

func TestPostAuthLogout_MissingCSRFHeader_Returns403(t *testing.T) {
	repo := newFakeUserRepo()
	h := newAuthTestServer(repo, &fakeGoogleVerifier{identity: verifiedIdentity()}, nil)

	loginRec := postRaw(t, h, "/api/v1/auth/google", `{"idToken":"a-google-id-token"}`)
	cookie := findCookie(loginRec, testCookieCfg.Name)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403, body = %s", rec.Code, rec.Body.String())
	}
}

func TestPostAuthRefresh_FormEncodedBody_Returns403EvenWithCSRFHeader(t *testing.T) {
	repo := newFakeUserRepo()
	h := newAuthTestServer(repo, &fakeGoogleVerifier{identity: verifiedIdentity()}, nil)

	loginRec := postRaw(t, h, "/api/v1/auth/google", `{"idToken":"a-google-id-token"}`)
	cookie := findCookie(loginRec, testCookieCfg.Name)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewBufferString("a=b"))
	req.Header.Set("X-JOPS-Refresh", "1")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403, body = %s", rec.Code, rec.Body.String())
	}
}

func TestPostAuthLogout_FormEncodedBody_Returns403EvenWithCSRFHeader(t *testing.T) {
	repo := newFakeUserRepo()
	h := newAuthTestServer(repo, &fakeGoogleVerifier{identity: verifiedIdentity()}, nil)

	loginRec := postRaw(t, h, "/api/v1/auth/google", `{"idToken":"a-google-id-token"}`)
	cookie := findCookie(loginRec, testCookieCfg.Name)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", bytes.NewBufferString("a=b"))
	req.Header.Set("X-JOPS-Refresh", "1")
	req.Header.Set("Content-Type", "multipart/form-data; boundary=x")
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403, body = %s", rec.Code, rec.Body.String())
	}
}
