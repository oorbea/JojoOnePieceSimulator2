package endpoints_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/endpoints"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/idgen"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/refreshtoken"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/streamticket"
)

// newDevAuthTestServer mirrors newAuthTestServer but calls
// SetDevAuthBypass(devAuth), exercising the actual toggle a real deployment
// flips via cfg.DevAuthBypass (see cmd/app/main.go), instead of relying on
// Routes' internal default.
func newDevAuthTestServer(repo *fakeUserRepo, devAuth bool) http.Handler {
	verifier := &fakeGoogleVerifier{identity: verifiedIdentity()}
	refreshStore := refreshtoken.NewMemoryStore(refreshtoken.Config{TTL: time.Hour})
	svc := services.NewAuthService(repo, idgen.UUIDGenerator[user.UserID]{}, verifier, loginTokenIssuer{}, refreshStore, nil, newFakePictureStorage())

	authEndpoints := endpoints.NewAuthEndpoints(svc, testCookieCfg)
	authEndpoints.SetDevAuthBypass(devAuth)

	return endpoints.NewRouter(
		authEndpoints,
		endpoints.NewStandEndpoints(nil),
		endpoints.NewDevilFruitEndpoints(nil),
		endpoints.NewUserEndpoints(nil),
		endpoints.NewEventsEndpoints(services.NewPictureEventHub(), fakeTokenIssuer{}, streamticket.NewMemoryStore(streamticket.Config{TTL: 30 * time.Second}), context.Background()),
		endpoints.NewGameEndpoints(nil, services.NewGameEventHub(), nil, nil, nil, nil, fakeTokenIssuer{}, streamticket.NewMemoryStore(streamticket.Config{TTL: 30 * time.Second}), context.Background(), endpoints.GameWSConfig{}),
		endpoints.NewStageEndpoints(nil),
		nil,
		endpoints.NewJojoCharacterEndpoints(nil),
		endpoints.NewOnePieceCharacterEndpoints(nil),
		fakeTokenIssuer{}, endpoints.CORSConfig{}, endpoints.RateLimitConfig{}, endpoints.CacheConfig{}, 0)
}

// devLoginRequest issues POST /api/v1/auth/dev-login with the given body and
// remoteAddr (mimicking r.RemoteAddr - httptest.NewRequest lets us set it
// directly), plus any extra headers (used to simulate a reverse proxy).
func devLoginRequest(h http.Handler, body, remoteAddr string, extraHeaders map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/dev-login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = remoteAddr
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestPostAuthDevLogin_FlagOff_Returns404(t *testing.T) {
	repo := newFakeUserRepo()
	h := newDevAuthTestServer(repo, false)

	rec := devLoginRequest(h, `{"name":"alice","admin":false}`, "172.17.0.1:54321", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404, body = %s", rec.Code, rec.Body.String())
	}
}

func TestPostAuthDevLogin_LocalRequest_Returns201AndCreatesUser(t *testing.T) {
	repo := newFakeUserRepo()
	h := newDevAuthTestServer(repo, true)

	rec := devLoginRequest(h, `{"name":"alice","admin":false}`, "172.17.0.1:54321", nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201, body = %s", rec.Code, rec.Body.String())
	}

	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, present := got["refreshToken"]; !present {
		t.Error("dev-login body must always carry refreshToken")
	}
	if findCookie(rec, testCookieCfg.Name) != nil {
		t.Error("dev-login must not set the shared refresh cookie")
	}
	userResp, _ := got["user"].(map[string]any)
	if userResp["email"] != "alice"+services.DevEmailDomain {
		t.Errorf("user.email = %v, want alice%s", userResp["email"], services.DevEmailDomain)
	}
}

func TestPostAuthDevLogin_LoopbackAddr_Returns201(t *testing.T) {
	repo := newFakeUserRepo()
	h := newDevAuthTestServer(repo, true)

	rec := devLoginRequest(h, `{"name":"alice","admin":false}`, "127.0.0.1:54321", nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201, body = %s", rec.Code, rec.Body.String())
	}
}

func TestPostAuthDevLogin_PublicRemoteAddr_Returns404(t *testing.T) {
	repo := newFakeUserRepo()
	h := newDevAuthTestServer(repo, true)

	rec := devLoginRequest(h, `{"name":"alice","admin":false}`, "203.0.113.5:54321", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404, body = %s", rec.Code, rec.Body.String())
	}
}

func TestPostAuthDevLogin_XForwardedForPresent_Returns404EvenFromLocalAddr(t *testing.T) {
	repo := newFakeUserRepo()
	h := newDevAuthTestServer(repo, true)

	rec := devLoginRequest(h, `{"name":"alice","admin":false}`, "172.17.0.1:54321", map[string]string{"X-Forwarded-For": "1.2.3.4"})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404, body = %s", rec.Code, rec.Body.String())
	}
}

func TestPostAuthDevLogin_AdminFlag_SetsAdminRole(t *testing.T) {
	repo := newFakeUserRepo()
	h := newDevAuthTestServer(repo, true)

	rec := devLoginRequest(h, `{"name":"alice","admin":true}`, "172.17.0.1:54321", nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var got map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &got)
	userResp, _ := got["user"].(map[string]any)
	if userResp["role"] != "ADMIN" {
		t.Errorf("user.role = %v, want ADMIN", userResp["role"])
	}
}

func TestPostAuthDevLogin_InvalidName_Returns400(t *testing.T) {
	repo := newFakeUserRepo()
	h := newDevAuthTestServer(repo, true)

	rec := devLoginRequest(h, `{"name":"Not Valid!","admin":false}`, "172.17.0.1:54321", nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400, body = %s", rec.Code, rec.Body.String())
	}
}
