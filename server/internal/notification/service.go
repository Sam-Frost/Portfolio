package notification

import (
	"context"
	"fmt"
	"html"
	"log/slog"
	"strings"

	"github.com/Sam-Frost/portfolio/internal/apperr"
	"github.com/Sam-Frost/portfolio/internal/mailer"
	"github.com/Sam-Frost/portfolio/internal/settings"
)

// SettingsReader is the slice of settings.Service the notification service
// needs — kept narrow so there's no package cycle (settings doesn't import
// notification, and this doesn't need the write side).
type SettingsReader interface {
	Get(ctx context.Context) (settings.Settings, error)
}

type Service struct {
	repo     Repository
	push     PushSender
	mail     mailer.Mailer
	settings SettingsReader
}

func NewService(repo Repository, push PushSender, mail mailer.Mailer, settings SettingsReader) *Service {
	return &Service{repo: repo, push: push, mail: mail, settings: settings}
}

func (s *Service) Subscribe(ctx context.Context, input SubscribeInput) error {
	if input.Endpoint == "" || input.P256dh == "" || input.Auth == "" {
		return apperr.InvalidInput("endpoint, p256dh and auth are required")
	}
	return s.repo.Subscribe(ctx, PushSubscription{
		Endpoint:  input.Endpoint,
		P256dh:    input.P256dh,
		Auth:      input.Auth,
		UserAgent: input.UserAgent,
	})
}

// Resync re-registers a push subscription the browser rotated, moving the
// row from oldEndpoint if we still have it. Unauthenticated (see the
// handler) — the service worker has no bearer token when
// pushsubscriptionchange fires.
func (s *Service) Resync(ctx context.Context, oldEndpoint string, input SubscribeInput) error {
	if input.Endpoint == "" || input.P256dh == "" || input.Auth == "" {
		return apperr.InvalidInput("endpoint, p256dh and auth are required")
	}
	return s.repo.Resync(ctx, oldEndpoint, PushSubscription{
		Endpoint:  input.Endpoint,
		P256dh:    input.P256dh,
		Auth:      input.Auth,
		UserAgent: input.UserAgent,
	})
}

func (s *Service) Unsubscribe(ctx context.Context, endpoint string) error {
	if endpoint == "" {
		return apperr.InvalidInput("endpoint is required")
	}
	return s.repo.Unsubscribe(ctx, endpoint)
}

// VAPIDPublicKey is what the client passes to PushManager.subscribe(). Empty
// when push is not configured — the client treats that as "push unavailable".
func (s *Service) VAPIDPublicKey() string { return s.push.PublicKey() }

// SendTest delivers a fixed message so the user can confirm a device (and
// their email) is wired up from the Settings screen.
func (s *Service) SendTest(ctx context.Context) error {
	return s.Notify(ctx, Message{
		Title: "Domain Expansion",
		Body:  "Test notification — you're all set.",
		URL:   "/settings",
		Tag:   "test",
	})
}

// AlreadySent reports whether a once-per-day notification of kind has
// already gone out for istDate ("YYYY-MM-DD"). Used by the scheduler to
// keep the morning digest to one send per day.
func (s *Service) AlreadySent(ctx context.Context, kind, istDate string) (bool, error) {
	return s.repo.LogExists(ctx, kind, istDate)
}

// RecordSent marks a once-per-day notification of kind as delivered for
// istDate.
func (s *Service) RecordSent(ctx context.Context, kind, istDate string) error {
	return s.repo.InsertLog(ctx, kind, istDate)
}

// Notify fans a message out to every enabled channel. It's best-effort:
// a failure on one channel is logged and doesn't block the other, so a
// flaky SMTP connection never costs you the push (and vice versa).
func (s *Service) Notify(ctx context.Context, m Message) error {
	cfg, err := s.settings.Get(ctx)
	if err != nil {
		return err
	}

	emailed := false
	if cfg.Notifications.EmailEnabled && cfg.Notifications.RecipientEmail != nil && *cfg.Notifications.RecipientEmail != "" {
		if err := s.mail.Send(ctx, []string{*cfg.Notifications.RecipientEmail}, m.Title, emailHTML(m), emailText(m)); err != nil {
			slog.Error("notification: email send failed", "err", err)
		} else {
			emailed = true
		}
	}

	pushed := 0
	if cfg.Notifications.PushEnabled {
		pushed = s.pushAll(ctx, m)
	}

	// One line per notification so it's obvious in prod logs whether a
	// channel actually did anything (vs. "fired" but nothing configured).
	slog.Info("notification dispatched",
		"title", m.Title,
		"emailed", emailed,
		"pushed", pushed,
		"emailEnabled", cfg.Notifications.EmailEnabled,
		"hasRecipient", cfg.Notifications.RecipientEmail != nil && *cfg.Notifications.RecipientEmail != "",
		"pushEnabled", cfg.Notifications.PushEnabled,
	)

	return nil
}

// pushAll delivers m to every subscription and returns how many were sent
// successfully.
func (s *Service) pushAll(ctx context.Context, m Message) int {
	subs, err := s.repo.ListSubscriptions(ctx)
	if err != nil {
		slog.Error("notification: list subscriptions failed", "err", err)
		return 0
	}

	sent := 0
	payload := encodePushPayload(m)
	for _, sub := range subs {
		res := s.push.Send(ctx, sub, payload)
		switch {
		case res.Gone:
			if err := s.repo.Unsubscribe(ctx, res.Endpoint); err != nil {
				slog.Error("notification: pruning dead subscription failed", "err", err)
			}
		case res.Err != nil:
			slog.Error("notification: web push failed", "endpoint", res.Endpoint, "err", res.Err)
		default:
			sent++
		}
	}
	return sent
}

func emailText(m Message) string {
	if m.Body == "" {
		return m.Title
	}
	return m.Title + "\n\n" + m.Body
}

func emailHTML(m Message) string {
	var b strings.Builder
	fmt.Fprintf(&b, `<div style="font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;font-size:15px;line-height:1.5;color:#111">`)
	fmt.Fprintf(&b, `<p style="font-weight:600;margin:0 0 8px">%s</p>`, html.EscapeString(m.Title))
	if m.Body != "" {
		fmt.Fprintf(&b, `<div>%s</div>`, strings.ReplaceAll(html.EscapeString(m.Body), "\n", "<br>"))
	}
	b.WriteString(`</div>`)
	return b.String()
}
