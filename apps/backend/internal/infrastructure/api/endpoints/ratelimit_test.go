package endpoints_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/endpoints"
)

func TestRateLimit_Disabled_NoLimiting(t *testing.T) {
	h := newTestServerWithRateLimit(endpoints.RateLimitConfig{Enabled: false})

	for i := 0; i < 10; i++ {
		rec := noAuthRequest(t, h, http.MethodGet, "/health")
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: status = %d, want %d", i, rec.Code, http.StatusOK)
		}
	}
}

func TestRateLimit_GlobalPerIP_Returns429(t *testing.T) {
	h := newTestServerWithRateLimit(endpoints.RateLimitConfig{
		Enabled:     true,
		Window:      time.Minute,
		GlobalPerIP: 2,
	})

	for i := 0; i < 2; i++ {
		rec := noAuthRequest(t, h, http.MethodGet, "/health")
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: status = %d, want %d, body = %s", i, rec.Code, http.StatusOK, rec.Body.String())
		}
	}

	rec := noAuthRequest(t, h, http.MethodGet, "/health")
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusTooManyRequests, rec.Body.String())
	}
	if got := rec.Body.String(); got != `{"error":"too many requests","code":"RATE_LIMITED"}`+"\n" {
		t.Errorf("body = %q, want %q", got, `{"error":"too many requests","code":"RATE_LIMITED"}`+"\n")
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Error("Retry-After header not set")
	}
	if rec.Header().Get("X-RateLimit-Limit") == "" {
		t.Error("X-RateLimit-Limit header not set")
	}
}

func TestRateLimit_LoginTier_IsStricterThanGlobal(t *testing.T) {
	h := newTestServerWithRateLimit(endpoints.RateLimitConfig{
		Enabled:     true,
		Window:      time.Minute,
		GlobalPerIP: 100,
		LoginPerIP:  1,
	})

	rec := doRawJSONRequest(t, h, http.MethodPost, "/api/v1/auth/google", `{"idToken":"bogus"}`)
	if rec.Code == http.StatusTooManyRequests {
		t.Fatalf("first login request already rate-limited, body = %s", rec.Body.String())
	}

	rec = doRawJSONRequest(t, h, http.MethodPost, "/api/v1/auth/google", `{"idToken":"bogus"}`)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("second login request: status = %d, want %d, body = %s", rec.Code, http.StatusTooManyRequests, rec.Body.String())
	}

	// /health is unaffected: it only draws from the (much larger) global tier.
	healthRec := noAuthRequest(t, h, http.MethodGet, "/health")
	if healthRec.Code != http.StatusOK {
		t.Fatalf("/health status = %d, want %d", healthRec.Code, http.StatusOK)
	}
}

func TestRateLimit_WriteTier_SeparateFromReadTier(t *testing.T) {
	h := newTestServerWithRateLimit(endpoints.RateLimitConfig{
		Enabled:      true,
		Window:       time.Minute,
		GlobalPerIP:  1000,
		ReadPerUser:  1000,
		WritePerUser: 1,
	})

	rec := doRequest(t, h, http.MethodPost, "/api/v1/stands", validStandBody("Crazy Diamond"))
	if rec.Code != http.StatusCreated {
		t.Fatalf("first create: status = %d, want %d, body = %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	rec = doRequest(t, h, http.MethodPost, "/api/v1/stands", validStandBody("Star Platinum"))
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("second create: status = %d, want %d, body = %s", rec.Code, http.StatusTooManyRequests, rec.Body.String())
	}

	// Reads still work: the write tier being exhausted must not affect reads.
	rec = doRequest(t, h, http.MethodGet, "/api/v1/stands", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("read after write exhausted: status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestRateLimit_PerUser_KeysAreIndependent(t *testing.T) {
	h := newTestServerWithRateLimit(endpoints.RateLimitConfig{
		Enabled:     true,
		Window:      time.Minute,
		GlobalPerIP: 1000,
		ReadPerUser: 1,
	})

	rec := userAuthRequest(t, h, http.MethodGet, "/api/v1/stands", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("user 1 first read: status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	rec = userAuthRequest(t, h, http.MethodGet, "/api/v1/stands", nil)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("user 1 second read: status = %d, want %d, body = %s", rec.Code, http.StatusTooManyRequests, rec.Body.String())
	}

	// A different authenticated user has an independent budget.
	rec = user2AuthRequest(t, h, http.MethodGet, "/api/v1/stands")
	if rec.Code != http.StatusOK {
		t.Fatalf("user 2 first read: status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestRateLimit_HealthAndSwaggerAreLimited(t *testing.T) {
	h := newTestServerWithRateLimit(endpoints.RateLimitConfig{
		Enabled:     true,
		Window:      time.Minute,
		GlobalPerIP: 1,
	})

	rec := noAuthRequest(t, h, http.MethodGet, "/health")
	if rec.Code != http.StatusOK {
		t.Fatalf("first /health: status = %d, want %d", rec.Code, http.StatusOK)
	}

	rec = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("/swagger after global budget exhausted by /health: status = %d, want %d, body = %s", rec.Code, http.StatusTooManyRequests, rec.Body.String())
	}
}

func TestRateLimit_KeysOnRightmostXFFEntry_NotRemoteAddr(t *testing.T) {
	h := newTestServerWithRateLimit(endpoints.RateLimitConfig{
		Enabled:     true,
		Window:      time.Minute,
		GlobalPerIP: 1,
	})

	// Same RemoteAddr (httptest's default), different rightmost XFF entries:
	// each must get its own bucket, or the fix didn't take.
	rec := xffRequest(t, h, "203.0.113.10")
	if rec.Code != http.StatusOK {
		t.Fatalf("client A first request: status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	rec = xffRequest(t, h, "203.0.113.20")
	if rec.Code != http.StatusOK {
		t.Fatalf("client B first request: status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	// Each is now at its own budget of 1/1 - a second request from either
	// must 429, proving the two really are independent buckets rather than
	// both having silently fallen back to the shared RemoteAddr bucket.
	rec = xffRequest(t, h, "203.0.113.10")
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("client A second request: status = %d, want %d, body = %s", rec.Code, http.StatusTooManyRequests, rec.Body.String())
	}
	rec = xffRequest(t, h, "203.0.113.20")
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("client B second request: status = %d, want %d, body = %s", rec.Code, http.StatusTooManyRequests, rec.Body.String())
	}
}

func TestRateLimit_XFF_RightmostEntryWinsOverSpoofedLeftEntry(t *testing.T) {
	h := newTestServerWithRateLimit(endpoints.RateLimitConfig{
		Enabled:     true,
		Window:      time.Minute,
		GlobalPerIP: 1,
	})

	// "spoofed" is attacker-supplied and sits to the left of the real,
	// proxy-appended entry - it must be ignored. Both requests carry the
	// same rightmost entry, so the second must 429.
	rec := xffRequest(t, h, "198.51.100.1, 203.0.113.30")
	if rec.Code != http.StatusOK {
		t.Fatalf("first request: status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	rec = xffRequest(t, h, "198.51.100.99, 203.0.113.30")
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("second request (different spoofed prefix, same real IP): status = %d, want %d, body = %s", rec.Code, http.StatusTooManyRequests, rec.Body.String())
	}
}

func TestRateLimit_NoXFFHeader_FallsBackToRemoteAddr(t *testing.T) {
	h := newTestServerWithRateLimit(endpoints.RateLimitConfig{
		Enabled:     true,
		Window:      time.Minute,
		GlobalPerIP: 2,
	})

	// No X-Forwarded-For at all (plain dev/direct traffic) - unchanged
	// behavior, keyed on httptest's default RemoteAddr.
	for i := 0; i < 2; i++ {
		rec := noAuthRequest(t, h, http.MethodGet, "/health")
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: status = %d, want %d", i, rec.Code, http.StatusOK)
		}
	}
	rec := noAuthRequest(t, h, http.MethodGet, "/health")
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("third request: status = %d, want %d, body = %s", rec.Code, http.StatusTooManyRequests, rec.Body.String())
	}
}

func TestRateLimit_XFF_SameRightmostEntryFromDifferentRemoteAddrsSharesBucket(t *testing.T) {
	h := newTestServerWithRateLimit(endpoints.RateLimitConfig{
		Enabled:     true,
		Window:      time.Minute,
		GlobalPerIP: 1,
	})

	rec := xffRequestFrom(t, h, "203.0.113.40", "10.0.0.5:1111")
	if rec.Code != http.StatusOK {
		t.Fatalf("first request: status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	// Different RemoteAddr (a different NPM upstream connection), same
	// resolved client IP via XFF - must share the exhausted bucket.
	rec = xffRequestFrom(t, h, "203.0.113.40", "10.0.0.6:2222")
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("second request from different RemoteAddr: status = %d, want %d, body = %s", rec.Code, http.StatusTooManyRequests, rec.Body.String())
	}
}

// xffRequest issues a GET /health carrying the given X-Forwarded-For value.
func xffRequest(t *testing.T, h http.Handler, xff string) *httptest.ResponseRecorder {
	t.Helper()
	return xffRequestFrom(t, h, xff, "")
}

// xffRequestFrom is like xffRequest but also lets the test set RemoteAddr,
// for asserting XFF wins over it regardless of which TCP peer connected.
func xffRequestFrom(t *testing.T, h http.Handler, xff, remoteAddr string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("X-Forwarded-For", xff)
	if remoteAddr != "" {
		req.RemoteAddr = remoteAddr
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// doRawJSONRequest posts a raw JSON body without an Authorization header, for
// exercising the public /auth/google route.
func doRawJSONRequest(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// user2AuthRequest is like userAuthRequest but authenticates as a second,
// distinct regular user (see userIDForToken), for exercising per-user
// rate-limit key independence.
func user2AuthRequest(t *testing.T, h http.Handler, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Authorization", "Bearer user2-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}
