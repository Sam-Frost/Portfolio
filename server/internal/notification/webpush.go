package notification

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	webpush "github.com/SherClockHolmes/webpush-go"
)

// pushResult reports what happened for one subscription so the service can
// prune the ones the push service has permanently rejected.
type pushResult struct {
	Endpoint string
	Gone     bool // 404/410 — subscription is dead, delete it
	Err      error
}

// PushSender delivers an already-built payload to a single Web Push
// endpoint. The noop implementation is used when VAPID keys aren't
// configured, mirroring mailer.NoopMailer.
type PushSender interface {
	Send(ctx context.Context, sub PushSubscription, payload []byte) pushResult
	// PublicKey is the VAPID public key the client needs for
	// PushManager.subscribe(); "" when push is disabled.
	PublicKey() string
}

type VAPIDConfig struct {
	PublicKey  string
	PrivateKey string
	Subject    string // "mailto:you@example.com"
}

type webPushSender struct {
	cfg VAPIDConfig
}

func NewWebPushSender(cfg VAPIDConfig) PushSender { return &webPushSender{cfg: cfg} }

func (s *webPushSender) PublicKey() string { return s.cfg.PublicKey }

func (s *webPushSender) Send(ctx context.Context, sub PushSubscription, payload []byte) pushResult {
	resp, err := webpush.SendNotificationWithContext(ctx, payload, &webpush.Subscription{
		Endpoint: sub.Endpoint,
		Keys:     webpush.Keys{P256dh: sub.P256dh, Auth: sub.Auth},
	}, &webpush.Options{
		Subscriber:      s.cfg.Subject,
		VAPIDPublicKey:  s.cfg.PublicKey,
		VAPIDPrivateKey: s.cfg.PrivateKey,
		TTL:             60 * 60 * 24, // hold up to a day if the device is offline
		Urgency:         webpush.UrgencyNormal,
	})
	if err != nil {
		return pushResult{Endpoint: sub.Endpoint, Err: err}
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone:
		return pushResult{Endpoint: sub.Endpoint, Gone: true}
	case resp.StatusCode >= 300:
		return pushResult{Endpoint: sub.Endpoint, Err: &pushStatusError{resp.StatusCode}}
	default:
		return pushResult{Endpoint: sub.Endpoint}
	}
}

type pushStatusError struct{ code int }

func (e *pushStatusError) Error() string { return "web push rejected: HTTP " + http.StatusText(e.code) }

// noopPushSender logs and drops. Selected when VAPID isn't configured.
type noopPushSender struct{}

func NewNoopPushSender() PushSender { return noopPushSender{} }

func (noopPushSender) PublicKey() string { return "" }

func (noopPushSender) Send(_ context.Context, sub PushSubscription, _ []byte) pushResult {
	slog.Info("web push disabled (VAPID not configured) — dropping message", "endpoint", sub.Endpoint)
	return pushResult{Endpoint: sub.Endpoint}
}

// pushPayload is the JSON the service worker's 'push' handler reads.
type pushPayload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	URL   string `json:"url"`
	Tag   string `json:"tag"`
}

func encodePushPayload(m Message) []byte {
	b, _ := json.Marshal(pushPayload{Title: m.Title, Body: m.Body, URL: m.URL, Tag: m.Tag})
	return b
}
