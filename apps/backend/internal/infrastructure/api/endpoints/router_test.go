package endpoints_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/endpoints"
)

func newCORSTestServer(corsCfg endpoints.CORSConfig) *httptest.Server {
	repo := newFakeStandRepository()
	svc := services.NewStandService(repo, &fakeIDGenerator{}, newFakePictureStorage(), &fakeImageProcessor{}, fullQueueEnqueuer{},
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/webp", "image/avif", "image/jpeg", "image/png", "image/gif"}})
	standEndpoints := endpoints.NewStandEndpoints(svc)
	authEndpoints := endpoints.NewAuthEndpoints(nil)
	eventsEndpoints := endpoints.NewEventsEndpoints(services.NewPictureEventHub(), fakeTokenIssuer{}, context.Background())
	gameEndpoints := endpoints.NewGameEndpoints(nil, services.NewGameEventHub(), fakeTokenIssuer{}, context.Background(), endpoints.GameWSConfig{})
	stageEndpoints := endpoints.NewStageEndpoints(nil)
	handler := endpoints.NewRouter(authEndpoints, standEndpoints, endpoints.NewDevilFruitEndpoints(nil), endpoints.NewUserEndpoints(nil), eventsEndpoints, gameEndpoints, stageEndpoints, fakeTokenIssuer{}, corsCfg, endpoints.RateLimitConfig{}, endpoints.CacheConfig{})
	return httptest.NewServer(handler)
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
