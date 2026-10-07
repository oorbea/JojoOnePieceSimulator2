// Package push delivers Web Push notifications (RFC 8030 + VAPID, RFC 8292)
// through the browser vendors' push services.
package push

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	webpush "github.com/SherClockHolmes/webpush-go"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

const (
	// messageTTL is how long the push service keeps an undelivered message
	// for an offline device. A game moment ("voting opened") is worthless a
	// minute later, so a stale buzz after the phone comes back is worse than
	// none.
	messageTTL = 60

	// requestTimeout bounds one delivery attempt end to end.
	requestTimeout = 10 * time.Second

	// maxResponseBytes caps how much of a push service reply is read - only
	// the status matters.
	maxResponseBytes = 4096
)

// Config holds the VAPID identity every message is signed with.
type Config struct {
	PublicKey  string
	PrivateKey string
	// Subject is a mailto: or https: contact the push services can reach if
	// the sender misbehaves (VAPID `sub` claim).
	Subject string
}

// WebPushSender is the ports.IPushSender adapter over webpush-go.
type WebPushSender struct {
	cfg    Config
	client webpush.HTTPClient
}

var _ ports.IPushSender = (*WebPushSender)(nil)

// NewWebPushSender builds a sender. A nil client gets a default one with a
// request timeout (tests inject an httptest-backed client).
func NewWebPushSender(cfg Config, client webpush.HTTPClient) *WebPushSender {
	if client == nil {
		client = &http.Client{Timeout: requestTimeout}
	}
	return &WebPushSender{cfg: cfg, client: client}
}

// payload is the JSON the service worker's push handler reads (public/sw.js).
type payload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	URL   string `json:"url"`
	Tag   string `json:"tag"`
}

// Send implements ports.IPushSender.
func (s *WebPushSender) Send(ctx context.Context, sub ports.PushSubscription, msg ports.PushMessage) error {
	body, err := json.Marshal(payload{Title: msg.Title, Body: msg.Body, URL: msg.URL, Tag: msg.Tag})
	if err != nil {
		return fmt.Errorf("encoding push payload: %w", err)
	}

	resp, err := webpush.SendNotificationWithContext(ctx, body, &webpush.Subscription{
		Endpoint: sub.Endpoint,
		Keys:     webpush.Keys{P256dh: sub.P256dh, Auth: sub.Auth},
	}, &webpush.Options{
		HTTPClient:      s.client,
		Subscriber:      s.cfg.Subject,
		VAPIDPublicKey:  s.cfg.PublicKey,
		VAPIDPrivateKey: s.cfg.PrivateKey,
		TTL:             messageTTL,
		Urgency:         webpush.UrgencyHigh,
	})
	if err != nil {
		return fmt.Errorf("sending web push: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, maxResponseBytes))

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		return nil
	case resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone:
		return ports.ErrPushSubscriptionGone
	default:
		return fmt.Errorf("push service answered HTTP %d", resp.StatusCode)
	}
}
