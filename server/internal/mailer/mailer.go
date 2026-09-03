// Package mailer is the shared outbound-email boundary, used by any feature
// that needs to send mail (today: internal/notification). Like
// internal/httpx it's infrastructure, not a feature slice — it knows
// nothing about what's being sent, only how to send it.
//
// The concrete SMTP sender (smtp.go) is built from env in cmd/main.go;
// with SMTP unconfigured a NoopMailer is used instead so local dev and the
// no-credentials deployment keep working — the same pattern as
// cms.NewNoopPublisher.
package mailer

import (
	"context"
	"log/slog"
)

// Mailer sends one email to one or more recipients. htmlBody and textBody
// are the two alternatives of a multipart/alternative message; a caller
// that only has plain text may pass "" for htmlBody.
type Mailer interface {
	Send(ctx context.Context, to []string, subject, htmlBody, textBody string) error
}

// NoopMailer logs and drops the message. Selected when SMTP is not
// configured.
type NoopMailer struct{}

func NewNoopMailer() NoopMailer { return NoopMailer{} }

func (NoopMailer) Send(_ context.Context, to []string, subject, _, _ string) error {
	slog.Info("email disabled (SMTP not configured) — dropping message", "to", to, "subject", subject)
	return nil
}
