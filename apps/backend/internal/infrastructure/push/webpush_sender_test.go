package push_test

import (
	"context"
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	webpush "github.com/SherClockHolmes/webpush-go"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/push"
)

// newSubscription builds a syntactically valid browser subscription (a real
// P-256 key and 16-byte auth secret) pointing at endpoint, since webpush-go
// really encrypts the payload for it.
func newSubscription(t *testing.T, endpoint string) ports.PushSubscription {
	t.Helper()
	key, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generating subscriber key: %v", err)
	}
	auth := make([]byte, 16)
	if _, err := rand.Read(auth); err != nil {
		t.Fatalf("generating auth secret: %v", err)
	}
	return ports.PushSubscription{
		Endpoint: endpoint,
		P256dh:   base64.RawURLEncoding.EncodeToString(key.PublicKey().Bytes()),
		Auth:     base64.RawURLEncoding.EncodeToString(auth),
	}
}

func newSender(t *testing.T) *push.WebPushSender {
	t.Helper()
	private, public, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		t.Fatalf("generating VAPID keys: %v", err)
	}
	return push.NewWebPushSender(push.Config{PublicKey: public, PrivateKey: private, Subject: "mailto:ops@example.com"}, nil)
}

func TestWebPushSender_DeliversAnEncryptedVAPIDSignedRequest(t *testing.T) {
	var got *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	err := newSender(t).Send(context.Background(), newSubscription(t, srv.URL+"/push/abc"), ports.PushMessage{
		Title: "Votación", Body: "Te toca votar", URL: "/play/1", Tag: "game-1",
	})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if got == nil {
		t.Fatal("push service was never called")
	}
	if got.Method != http.MethodPost || got.URL.Path != "/push/abc" {
		t.Errorf("request = %s %s, want POST /push/abc", got.Method, got.URL.Path)
	}
	if !strings.HasPrefix(got.Header.Get("Authorization"), "vapid ") {
		t.Errorf("Authorization = %q, want a vapid credential", got.Header.Get("Authorization"))
	}
	if got.Header.Get("Content-Encoding") != "aes128gcm" {
		t.Errorf("Content-Encoding = %q, want aes128gcm", got.Header.Get("Content-Encoding"))
	}
	if got.Header.Get("Urgency") != "high" {
		t.Errorf("Urgency = %q, want high", got.Header.Get("Urgency"))
	}
	if got.Header.Get("TTL") != "60" {
		t.Errorf("TTL = %q, want 60", got.Header.Get("TTL"))
	}
}

func TestWebPushSender_ReportsAGoneSubscription(t *testing.T) {
	for _, status := range []int{http.StatusNotFound, http.StatusGone} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(status)
		}))

		err := newSender(t).Send(context.Background(), newSubscription(t, srv.URL), ports.PushMessage{Title: "t"})
		srv.Close()

		if !errors.Is(err, ports.ErrPushSubscriptionGone) {
			t.Errorf("status %d: err = %v, want ErrPushSubscriptionGone", status, err)
		}
	}
}

func TestWebPushSender_OtherFailuresAreNotGone(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	err := newSender(t).Send(context.Background(), newSubscription(t, srv.URL), ports.PushMessage{Title: "t"})
	if err == nil || errors.Is(err, ports.ErrPushSubscriptionGone) {
		t.Errorf("err = %v, want a plain delivery failure", err)
	}
}

func TestWebPushSender_RejectsMalformedKeys(t *testing.T) {
	sub := ports.PushSubscription{Endpoint: "https://fcm.googleapis.com/x", P256dh: "!!not-base64!!", Auth: "!!"}
	if err := newSender(t).Send(context.Background(), sub, ports.PushMessage{Title: "t"}); err == nil {
		t.Error("Send with malformed keys succeeded, want an error")
	}
}
