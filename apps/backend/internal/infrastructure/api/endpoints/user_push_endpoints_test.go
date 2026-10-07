package endpoints_test

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/dto"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/endpoints"
)

type fakePushRepo struct {
	mu   sync.Mutex
	subs map[string]ports.PushSubscription // by endpoint, like the real UNIQUE column
}

func newFakePushRepo() *fakePushRepo {
	return &fakePushRepo{subs: make(map[string]ports.PushSubscription)}
}

func (f *fakePushRepo) Upsert(_ context.Context, sub ports.PushSubscription) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.subs[sub.Endpoint] = sub
	return nil
}

func (f *fakePushRepo) DeleteForUser(_ context.Context, id user.UserID, endpoint string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if sub, ok := f.subs[endpoint]; ok && sub.UserID == id {
		delete(f.subs, endpoint)
	}
	return nil
}

func (f *fakePushRepo) DeleteByEndpoint(_ context.Context, endpoint string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.subs, endpoint)
	return nil
}

func (f *fakePushRepo) ListByUser(_ context.Context, id user.UserID) ([]ports.PushSubscription, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []ports.PushSubscription
	for _, sub := range f.subs {
		if sub.UserID == id {
			out = append(out, sub)
		}
	}
	return out, nil
}

// newPushTestServer mounts only the /users routes behind RequireAuth, which
// is all the push routes need (they never touch the UserService).
func newPushTestServer(repo ports.IPushSubscriptionRepository, publicKey string) http.Handler {
	userEndpoints := endpoints.NewUserEndpoints(nil)
	if repo != nil {
		userEndpoints.SetPushSubscriptions(repo, publicKey)
	}
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(endpoints.RequireAuth(fakeTokenIssuer{}))
		r.Mount("/users", userEndpoints.Routes(endpoints.RateLimitConfig{}))
	})
	return r
}

const goodEndpoint = "https://fcm.googleapis.com/fcm/send/device-1"

func goodSubscribeBody() dto.PushSubscribeRequest {
	return dto.PushSubscribeRequest{
		Endpoint: goodEndpoint,
		Keys: dto.PushKeys{
			P256dh: "BNcRdreALRFXTkOOUHK1EtK2wtaz5Ry4YfYCA_0QTpQtUbVlUls0VJXg7A8u-Ts1XbjhazAkj7I99e8QcYP7DkM",
			Auth:   "tBHItJI5svbpez7KI4CCXg",
		},
	}
}

func TestGetPushConfig_ReportsKeyWhenEnabled(t *testing.T) {
	h := newPushTestServer(newFakePushRepo(), "vapid-public-key")

	rec := doRequestAs(t, h, http.MethodGet, "/users/me/push", "user-token", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
	}
	var got dto.PushConfigResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decoding: %v", err)
	}
	if !got.Enabled || got.PublicKey != "vapid-public-key" {
		t.Errorf("config = %+v, want enabled with the key", got)
	}
}

func TestGetPushConfig_DisabledWithoutVAPIDKeys(t *testing.T) {
	h := newPushTestServer(nil, "")

	rec := doRequestAs(t, h, http.MethodGet, "/users/me/push", "user-token", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var got dto.PushConfigResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decoding: %v", err)
	}
	if got.Enabled || got.PublicKey != "" {
		t.Errorf("config = %+v, want disabled and no key", got)
	}
}

func TestPushRoutes_RequireAuthentication(t *testing.T) {
	h := newPushTestServer(newFakePushRepo(), "k")
	for _, route := range []struct{ method, path string }{
		{http.MethodGet, "/users/me/push"},
		{http.MethodPost, "/users/me/push/subscriptions"},
		{http.MethodDelete, "/users/me/push/subscriptions"},
	} {
		rec := doRequestAs(t, h, route.method, route.path, "", goodSubscribeBody())
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s %s without a token = %d, want 401", route.method, route.path, rec.Code)
		}
	}
}

func TestSubscribePush_StoresTheSubscriptionForTheCaller(t *testing.T) {
	repo := newFakePushRepo()
	h := newPushTestServer(repo, "k")

	rec := doRequestAs(t, h, http.MethodPost, "/users/me/push/subscriptions", "user-token", goodSubscribeBody())

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
	}
	subs, _ := repo.ListByUser(context.Background(), userIDForToken["user-token"])
	if len(subs) != 1 || subs[0].Endpoint != goodEndpoint {
		t.Fatalf("subscriptions = %+v, want the one registered", subs)
	}
	if other, _ := repo.ListByUser(context.Background(), userIDForToken["user2-token"]); len(other) != 0 {
		t.Errorf("another user got %d subscriptions", len(other))
	}
}

func TestSubscribePush_RejectsAnEndpointOutsideTheBrowserPushServices(t *testing.T) {
	repo := newFakePushRepo()
	h := newPushTestServer(repo, "k")
	body := goodSubscribeBody()
	body.Endpoint = "https://169.254.169.254/latest/meta-data"

	rec := doRequestAs(t, h, http.MethodPost, "/users/me/push/subscriptions", "user-token", body)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (SSRF guard), body = %s", rec.Code, rec.Body)
	}
	if len(repo.subs) != 0 {
		t.Errorf("a rejected subscription was stored: %+v", repo.subs)
	}
}

func TestSubscribePush_UnknownFieldIs400(t *testing.T) {
	h := newPushTestServer(newFakePushRepo(), "k")

	rec := doRequestAs(t, h, http.MethodPost, "/users/me/push/subscriptions", "user-token", map[string]any{
		"endpoint": goodEndpoint,
		"keys":     map[string]string{"p256dh": "abc", "auth": "def"},
		"userId":   "someone-else",
	})

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestSubscribePush_NotFoundWhenDisabled(t *testing.T) {
	h := newPushTestServer(nil, "")

	rec := doRequestAs(t, h, http.MethodPost, "/users/me/push/subscriptions", "user-token", goodSubscribeBody())

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestUnsubscribePush_OnlyRemovesTheCallersOwn(t *testing.T) {
	repo := newFakePushRepo()
	h := newPushTestServer(repo, "k")
	doRequestAs(t, h, http.MethodPost, "/users/me/push/subscriptions", "user-token", goodSubscribeBody())

	// Someone else naming the endpoint changes nothing.
	rec := doRequestAs(t, h, http.MethodDelete, "/users/me/push/subscriptions", "user2-token", dto.PushUnsubscribeRequest{Endpoint: goodEndpoint})
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
	if subs, _ := repo.ListByUser(context.Background(), userIDForToken["user-token"]); len(subs) != 1 {
		t.Fatalf("subscription removed by a different user")
	}

	rec = doRequestAs(t, h, http.MethodDelete, "/users/me/push/subscriptions", "user-token", dto.PushUnsubscribeRequest{Endpoint: goodEndpoint})
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
	if subs, _ := repo.ListByUser(context.Background(), userIDForToken["user-token"]); len(subs) != 0 {
		t.Errorf("subscription survived its owner's unsubscribe")
	}
}

func TestUnsubscribePush_RequiresAnEndpoint(t *testing.T) {
	h := newPushTestServer(newFakePushRepo(), "k")

	rec := doRequestAs(t, h, http.MethodDelete, "/users/me/push/subscriptions", "user-token", dto.PushUnsubscribeRequest{})

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}
