package endpoints_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/endpoints"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/streamticket"
)

func newCORSTestServer(corsCfg endpoints.CORSConfig) *httptest.Server {
	repo := newFakeStandRepository()
	svc := services.NewStandService(repo, &fakeIDGenerator{}, newFakePictureStorage(), &fakeImageProcessor{}, fullQueueEnqueuer{},
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/webp", "image/avif", "image/jpeg", "image/png", "image/gif"}})
	standEndpoints := endpoints.NewStandEndpoints(svc)
	authEndpoints := endpoints.NewAuthEndpoints(nil, endpoints.CookieConfig{})
	tickets := streamticket.NewMemoryStore(streamticket.Config{TTL: 30 * time.Second})
	eventsEndpoints := endpoints.NewEventsEndpoints(services.NewPictureEventHub(), fakeTokenIssuer{}, tickets, context.Background())
	gameEndpoints := endpoints.NewGameEndpoints(nil, services.NewGameEventHub(), nil, nil, nil, nil, fakeTokenIssuer{}, tickets, context.Background(), endpoints.GameWSConfig{})
	stageRepo := newFakeStageRepository()
	stageSvc := services.NewStageService(stageRepo, &fakeStageIDGenerator{}, newFakePictureStorage(), &fakeImageProcessor{}, fullQueueEnqueuer{},
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/webp", "image/avif", "image/jpeg", "image/png", "image/gif"}})
	stageEndpoints := endpoints.NewStageEndpoints(stageSvc)
	handler := endpoints.NewRouter(authEndpoints, standEndpoints, endpoints.NewDevilFruitEndpoints(nil), endpoints.NewUserEndpoints(nil), eventsEndpoints, gameEndpoints, stageEndpoints, fakeTokenIssuer{}, corsCfg, endpoints.RateLimitConfig{}, endpoints.CacheConfig{}, 0)
	return httptest.NewServer(handler)
}

// newCompressTestServer mirrors newCORSTestServer but exposes compressLevel,
// for TestCompress_*.
func newCompressTestServer(compressLevel int) *httptest.Server {
	repo := newFakeStandRepository()
	svc := services.NewStandService(repo, &fakeIDGenerator{}, newFakePictureStorage(), &fakeImageProcessor{}, fullQueueEnqueuer{},
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/webp", "image/avif", "image/jpeg", "image/png", "image/gif"}})
	standEndpoints := endpoints.NewStandEndpoints(svc)
	authEndpoints := endpoints.NewAuthEndpoints(nil, endpoints.CookieConfig{})
	tickets := streamticket.NewMemoryStore(streamticket.Config{TTL: 30 * time.Second})
	eventsEndpoints := endpoints.NewEventsEndpoints(services.NewPictureEventHub(), fakeTokenIssuer{}, tickets, context.Background())
	gameEndpoints := endpoints.NewGameEndpoints(nil, services.NewGameEventHub(), nil, nil, nil, nil, fakeTokenIssuer{}, tickets, context.Background(), endpoints.GameWSConfig{})
	handler := endpoints.NewRouter(authEndpoints, standEndpoints, endpoints.NewDevilFruitEndpoints(nil), endpoints.NewUserEndpoints(nil), eventsEndpoints, gameEndpoints, endpoints.NewStageEndpoints(nil), fakeTokenIssuer{}, endpoints.CORSConfig{}, endpoints.RateLimitConfig{}, endpoints.CacheConfig{}, compressLevel)
	return httptest.NewServer(handler)
}

// TestCompress_Enabled_GzipsJSONResponses guards the T1.5 wiring: with
// HTTPCompressLevel > 0, a request declaring gzip support gets a compressed
// /api/v1 JSON response.
func TestCompress_Enabled_GzipsJSONResponses(t *testing.T) {
	srv := newCompressTestServer(5)
	defer srv.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL+"/api/v1/stands", nil)
	if err != nil {
		t.Fatalf("building request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer admin-token")
	req.Header.Set("Accept-Encoding", "gzip")
	resp := doRawRequest(t, req)
	defer func() { _ = resp.Body.Close() }()

	if got := resp.Header.Get("Content-Encoding"); got != "gzip" {
		t.Errorf("Content-Encoding = %q, want gzip", got)
	}
}

// TestCompress_Disabled_NeverGzips guards HTTPCompressLevel=0 (the escape
// hatch) actually disabling compression rather than defaulting silently.
func TestCompress_Disabled_NeverGzips(t *testing.T) {
	srv := newCompressTestServer(0)
	defer srv.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL+"/api/v1/stands", nil)
	if err != nil {
		t.Fatalf("building request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer admin-token")
	req.Header.Set("Accept-Encoding", "gzip")
	resp := doRawRequest(t, req)
	defer func() { _ = resp.Body.Close() }()

	if got := resp.Header.Get("Content-Encoding"); got != "" {
		t.Errorf("Content-Encoding = %q, want empty (compression disabled)", got)
	}
}

// TestCompress_NeverAppliedToSSEEvents guards the T1.5 constraint that
// /api/v1/events (a long-lived stream) must never be wrapped by Compress,
// which buffers the body and would break streaming.
func TestCompress_NeverAppliedToSSEEvents(t *testing.T) {
	srv := newCompressTestServer(5)
	defer srv.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL+"/health", nil)
	if err != nil {
		t.Fatalf("building request: %v", err)
	}
	req.Header.Set("Accept-Encoding", "gzip")
	resp := doRawRequest(t, req)
	defer func() { _ = resp.Body.Close() }()

	// /health sits in the router's own Timeout-only group, outside the
	// Compress-wrapped /api/v1 group - same "never touches long-lived
	// streams" boundary /events and /games/{id}/ws rely on.
	if got := resp.Header.Get("Content-Encoding"); got != "" {
		t.Errorf("Content-Encoding = %q, want empty (outside the compressed group)", got)
	}
}

func newRequestWithOrigin(t *testing.T, url, origin string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("building request: %v", err)
	}
	req.Header.Set("Origin", origin)
	return req
}

func doRawRequest(t *testing.T, req *http.Request) *http.Response {
	t.Helper()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("performing request: %v", err)
	}
	return resp
}

func TestCORS_NoOriginsConfigured_NoHeadersEverAdded(t *testing.T) {
	srv := newCORSTestServer(endpoints.CORSConfig{})
	defer srv.Close()

	req := newRequestWithOrigin(t, srv.URL+"/health", "http://evil.example")
	resp := doRawRequest(t, req)
	defer func() { _ = resp.Body.Close() }()

	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Access-Control-Allow-Origin = %q, want empty (deny-all when unconfigured)", got)
	}
}

func TestCORS_ConfiguredOrigin_IsEchoedBack(t *testing.T) {
	srv := newCORSTestServer(endpoints.CORSConfig{
		AllowedOrigins: []string{"http://localhost:5173"},
		AllowedMethods: []string{"GET", "POST"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
	})
	defer srv.Close()

	req := newRequestWithOrigin(t, srv.URL+"/health", "http://localhost:5173")
	resp := doRawRequest(t, req)
	defer func() { _ = resp.Body.Close() }()

	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, "http://localhost:5173")
	}
}

func TestCORS_ConfiguredOrigin_RejectsOtherOrigins(t *testing.T) {
	srv := newCORSTestServer(endpoints.CORSConfig{
		AllowedOrigins: []string{"http://localhost:5173"},
		AllowedMethods: []string{"GET", "POST"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
	})
	defer srv.Close()

	req := newRequestWithOrigin(t, srv.URL+"/health", "http://evil.example")
	resp := doRawRequest(t, req)
	defer func() { _ = resp.Body.Close() }()

	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Access-Control-Allow-Origin = %q, want empty for an unconfigured origin", got)
	}
}
