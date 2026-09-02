package endpoints_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/endpoints"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/auth"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/idgen"
)

// RequireAuth/RequireAdmin are exercised indirectly all over this package
// (every /users and /stands test goes through them), but only ever with the
// happy path or a token fakeTokenIssuer already rejects. These tests drive
// the two middlewares on their own, so the header parsing itself - and the
// distinction between "no claims at all" and "claims without ADMIN" - is
// pinned rather than inferred.

// probeHandler records whether the middleware chain let the request through,
// and what claims it saw.
type probeHandler struct {
	called bool
	role   enums.UserRole
	userID user.UserID
}

func (p *probeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.called = true
	if claims, ok := endpoints.ClaimsFromRequest(r); ok {
		p.role = claims.Role
		p.userID = claims.UserID
	}
	w.WriteHeader(http.StatusNoContent)
}

// The three chains under test: RequireAuth alone, the real production
// pairing (RequireAuth then RequireAdmin), and RequireAdmin mounted on its
// own by mistake.
func requireAuthChainWith(issuer ports.ITokenIssuer, next http.Handler) http.Handler {
	return endpoints.RequireAuth(issuer)(next)
}

func requireAuthChain(next http.Handler) http.Handler {
	return requireAuthChainWith(fakeTokenIssuer{}, next)
}

func requireAdminChain(next http.Handler) http.Handler {
	return endpoints.RequireAuth(fakeTokenIssuer{})(endpoints.RequireAdmin(next))
}

func adminOnlyChain(next http.Handler) http.Handler {
	return endpoints.RequireAdmin(next)
}

func serveWithHeader(h http.Handler, header string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if header != "" {
		req.Header.Set("Authorization", header)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestRequireAuth_RejectsUnusableAuthorizationHeaders(t *testing.T) {
	cases := map[string]string{
		"no header":            "",
		"empty header":         " ",
		"bare token":           "user-token",
		"wrong scheme":         "Token user-token",
		"basic auth":           "Basic dXNlcjpwYXNz",
		"lowercase bearer":     "bearer user-token",
		"prefix without space": "Beareruser-token",
		"bearer with no token": "Bearer ",
		"unknown token":        "Bearer nonsense",
		"empty after trim":     "Bearer    ",
	}
	for name, header := range cases {
		probe := &probeHandler{}
		h := requireAuthChain(probe)

		rec := serveWithHeader(h, header)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s: status = %d, want %d, body = %s", name, rec.Code, http.StatusUnauthorized, rec.Body.String())
		}
		if probe.called {
			t.Errorf("%s: the protected handler ran anyway", name)
		}
	}
}

func TestRequireAuth_RejectsExpiredToken(t *testing.T) {
	// A real JWTIssuer with a negative TTL, so this is a genuinely expired
	// signature rather than a fake issuer's hardcoded rejection.
	expiredIssuer := auth.NewJWTIssuer([]byte("middleware-test-secret-at-least-32-bytes"), "test-issuer", -time.Minute)
	u := middlewareTestUser(t, enums.Regular)
	expired, _, err := expiredIssuer.Issue(u)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	// Parsed by a live issuer (positive TTL) - only the token's own exp is
	// stale, which is exactly the production case.
	liveIssuer := auth.NewJWTIssuer([]byte("middleware-test-secret-at-least-32-bytes"), "test-issuer", time.Hour)
	probe := &probeHandler{}
	h := requireAuthChainWith(liveIssuer, probe)

	rec := serveWithHeader(h, "Bearer "+expired)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
	if probe.called {
		t.Error("the protected handler ran on an expired token")
	}
}

func TestRequireAuth_PassesValidTokenAndPopulatesClaims(t *testing.T) {
	probe := &probeHandler{}
	h := requireAuthChain(probe)

	rec := serveWithHeader(h, "Bearer user-token")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusNoContent, rec.Body.String())
	}
	if !probe.called {
		t.Fatal("the protected handler never ran")
	}
	if probe.role != enums.Regular {
		t.Errorf("claims.Role = %v, want %v", probe.role, enums.Regular)
	}
	if probe.userID != userIDForToken["user-token"] {
		t.Errorf("claims.UserID = %v, want %v", probe.userID, userIDForToken["user-token"])
	}
}

func TestRequireAdmin_RegularUserIsForbidden(t *testing.T) {
	probe := &probeHandler{}
	h := requireAdminChain(probe)

	rec := serveWithHeader(h, "Bearer user-token")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
	if probe.called {
		t.Error("an admin-only handler ran for a REGULAR caller")
	}
}

func TestRequireAdmin_AdminPasses(t *testing.T) {
	probe := &probeHandler{}
	h := requireAdminChain(probe)

	rec := serveWithHeader(h, "Bearer admin-token")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusNoContent, rec.Body.String())
	}
	if probe.role != enums.Admin {
		t.Errorf("claims.Role = %v, want %v", probe.role, enums.Admin)
	}
}

// RequireAdmin reads claims RequireAuth put on the context. Mounted on its
// own - a wiring mistake that skipped RequireAuth - it must fail closed, not
// wave the request through for want of anything to check.
func TestRequireAdmin_WithoutRequireAuth_FailsClosed(t *testing.T) {
	probe := &probeHandler{}
	h := adminOnlyChain(probe)

	rec := serveWithHeader(h, "Bearer admin-token")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
	if probe.called {
		t.Error("an admin-only handler ran with no claims on the context")
	}
}

func middlewareTestUser(t *testing.T, role enums.UserRole) *user.User {
	t.Helper()
	u, err := user.NewUser(idgen.UUIDGenerator[user.UserID]{}.NewID(), "google-sub", "middleware@example.com", "middleware_user", "Middleware User", "", role)
	if err != nil {
		t.Fatalf("building test user: %v", err)
	}
	return u
}
