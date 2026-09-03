package endpoints_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/endpoints"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/streamticket"
)

// newEventsTestServer wires a full router (every other endpoint nil-backed,
// same convention as router_test.go/user_endpoints_test.go) around a real
// streamticket.MemoryStore with ttl, so mint and stream can be exercised
// end-to-end against the same store a real deployment would use.
func newEventsTestServer(t *testing.T, ttl time.Duration) (http.Handler, *streamticket.MemoryStore) {
	t.Helper()
	tickets := streamticket.NewMemoryStore(streamticket.Config{TTL: ttl})
	eventsEndpoints := endpoints.NewEventsEndpoints(services.NewPictureEventHub(), fakeTokenIssuer{}, tickets, context.Background())
	gameEndpoints := endpoints.NewGameEndpoints(nil, services.NewGameEventHub(), nil, nil, nil, nil, fakeTokenIssuer{}, tickets, context.Background(), endpoints.GameWSConfig{})
	h := endpoints.NewRouter(
		endpoints.NewAuthEndpoints(nil), endpoints.NewStandEndpoints(nil), endpoints.NewDevilFruitEndpoints(nil),
		endpoints.NewUserEndpoints(nil), eventsEndpoints, gameEndpoints, endpoints.NewStageEndpoints(nil),
		fakeTokenIssuer{}, endpoints.CORSConfig{}, endpoints.RateLimitConfig{}, endpoints.CacheConfig{},
	)
	return h, tickets
}

// mintEventsTicket does NOT assert expiresAt is in the future - callers
// exercising a near-zero TTL (TestStream_ExpiredTicket_Fails) expect exactly
// the opposite. TestMintEventsTicket_Admin_ReturnsTicket checks that
// separately, against a normal TTL.
func mintEventsTicket(t *testing.T, h http.Handler, token string) (ticket string, expiresAt time.Time, status int) {
	t.Helper()
	rec := doRequestAs(t, h, http.MethodPost, "/api/v1/events/ticket", token, nil)
	if rec.Code != http.StatusOK {
		return "", time.Time{}, rec.Code
	}
	var body struct {
		Ticket    string    `json:"ticket"`
		ExpiresAt time.Time `json:"expiresAt"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decoding mint response: %v, body = %s", err, rec.Body.String())
	}
	if body.Ticket == "" {
		t.Fatalf("mint response has an empty ticket: %s", rec.Body.String())
	}
	return body.Ticket, body.ExpiresAt, rec.Code
}

func TestMintEventsTicket_Unauthenticated(t *testing.T) {
	h, _ := newEventsTestServer(t, 30*time.Second)
	_, _, status := mintEventsTicket(t, h, "")
	if status != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", status)
	}
}

func TestMintEventsTicket_RegularUser_Forbidden(t *testing.T) {
	h, _ := newEventsTestServer(t, 30*time.Second)
	_, _, status := mintEventsTicket(t, h, "user-token")
	if status != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", status)
	}
}

func TestMintEventsTicket_Admin_ReturnsTicket(t *testing.T) {
	h, _ := newEventsTestServer(t, 30*time.Second)
	ticket, expiresAt, status := mintEventsTicket(t, h, "admin-token")
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200", status)
	}
	if ticket == "" {
		t.Fatal("expected a non-empty ticket")
	}
	if !expiresAt.After(time.Now()) {
		t.Fatalf("expiresAt %v is not in the future", expiresAt)
	}
}

// streamOnce drives GET /api/v1/events with a short-lived context so the
// handler's long-lived loop returns quickly instead of hanging the test -
// stream() itself only exits on ctx.Done()/e.ctx.Done()/an event/a
// heartbeat, none of which otherwise fire in a unit test.
func streamOnce(h http.Handler, path string, header http.Header) *httptest.ResponseRecorder {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	req := httptest.NewRequest(http.MethodGet, path, nil).WithContext(ctx)
	if header != nil {
		req.Header = header
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestStream_ValidTicket_Connects(t *testing.T) {
	h, _ := newEventsTestServer(t, 30*time.Second)
	ticket, _, status := mintEventsTicket(t, h, "admin-token")
	if status != http.StatusOK {
		t.Fatalf("mint status = %d, want 200", status)
	}

	rec := streamOnce(h, "/api/v1/events?ticket="+ticket, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body = %s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("Content-Type = %q, want text/event-stream", ct)
	}
	if !strings.Contains(rec.Body.String(), ": connected") {
		t.Fatalf("body = %q, want it to contain \": connected\"", rec.Body.String())
	}
}

func TestStream_ReusedTicket_Fails(t *testing.T) {
	h, _ := newEventsTestServer(t, 30*time.Second)
	ticket, _, _ := mintEventsTicket(t, h, "admin-token")

	first := streamOnce(h, "/api/v1/events?ticket="+ticket, nil)
	if first.Code != http.StatusOK {
		t.Fatalf("first connect status = %d, want 200", first.Code)
	}

	second := streamOnce(h, "/api/v1/events?ticket="+ticket, nil)
	if second.Code != http.StatusUnauthorized {
		t.Fatalf("reused ticket status = %d, want 401", second.Code)
	}
}

func TestStream_WrongPurposeTicket_Fails(t *testing.T) {
	h, tickets := newEventsTestServer(t, 30*time.Second)
	// Minted directly against the store (not through the HTTP mint route)
	// so it carries the game-ws purpose instead of events.
	token, _, err := tickets.Issue(context.Background(), ports.StreamTicket{
		Purpose: ports.TicketPurposeGameWS, Resource: "some-game",
	})
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	rec := streamOnce(h, "/api/v1/events?ticket="+token, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestStream_ExpiredTicket_Fails(t *testing.T) {
	h, _ := newEventsTestServer(t, time.Nanosecond)
	ticket, _, status := mintEventsTicket(t, h, "admin-token")
	if status != http.StatusOK {
		t.Fatalf("mint status = %d, want 200", status)
	}
	time.Sleep(time.Millisecond)

	rec := streamOnce(h, "/api/v1/events?ticket="+ticket, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

// TestStream_OldTokenQueryParam_Fails is the regression guard for the
// removed ?token=<jwt> fallback: even a token the fake issuer would happily
// Parse as a valid admin credential must be rejected once it's presented as
// ?token= instead of ?ticket= or a real Authorization header.
func TestStream_OldTokenQueryParam_Fails(t *testing.T) {
	h, _ := newEventsTestServer(t, 30*time.Second)
	rec := streamOnce(h, "/api/v1/events?token=admin-token", nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestStream_NoTicketOrToken_Fails(t *testing.T) {
	h, _ := newEventsTestServer(t, 30*time.Second)
	rec := streamOnce(h, "/api/v1/events", nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestStream_BearerAdmin_Connects(t *testing.T) {
	h, _ := newEventsTestServer(t, 30*time.Second)
	header := http.Header{"Authorization": []string{"Bearer admin-token"}}
	rec := streamOnce(h, "/api/v1/events", header)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body = %s", rec.Code, rec.Body.String())
	}
}

func TestStream_BearerRegular_Forbidden(t *testing.T) {
	h, _ := newEventsTestServer(t, 30*time.Second)
	header := http.Header{"Authorization": []string{"Bearer user-token"}}
	rec := streamOnce(h, "/api/v1/events", header)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}
