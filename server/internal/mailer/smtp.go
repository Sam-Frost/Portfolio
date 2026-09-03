package mailer

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

// SMTPConfig is the connection + identity config for SMTPMailer. Host/Port
// default to Gmail's submission endpoint in NewSMTPMailer when left blank.
type SMTPConfig struct {
	Host     string // e.g. "smtp.gmail.com"
	Port     string // e.g. "587"
	Username string // full email address used for AUTH
	Password string // app password (Gmail) or account password
	From     string // From header; defaults to Username
}

// SMTPMailer sends mail over an authenticated STARTTLS submission
// connection. Built for Gmail (smtp.gmail.com:587) but works with any
// provider that speaks SMTP+STARTTLS+AUTH LOGIN/PLAIN.
type SMTPMailer struct {
	cfg SMTPConfig
}

func NewSMTPMailer(cfg SMTPConfig) *SMTPMailer {
	if cfg.Host == "" {
		cfg.Host = "smtp.gmail.com"
	}
	if cfg.Port == "" {
		cfg.Port = "587"
	}
	if cfg.From == "" {
		cfg.From = cfg.Username
	}
	return &SMTPMailer{cfg: cfg}
}

func (m *SMTPMailer) Send(ctx context.Context, to []string, subject, htmlBody, textBody string) error {
	if len(to) == 0 {
		return fmt.Errorf("mailer: no recipients")
	}

	msg := buildMessage(m.cfg.From, to, subject, htmlBody, textBody)

	addr := net.JoinHostPort(m.cfg.Host, m.cfg.Port)
	d := net.Dialer{Timeout: 15 * time.Second}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("mailer: dial %s: %w", addr, err)
	}

	c, err := smtp.NewClient(conn, m.cfg.Host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("mailer: smtp client: %w", err)
	}
	defer c.Close()

	if err := c.StartTLS(&tls.Config{ServerName: m.cfg.Host}); err != nil {
		return fmt.Errorf("mailer: starttls: %w", err)
	}

	auth := smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)
	if err := c.Auth(auth); err != nil {
		return fmt.Errorf("mailer: auth: %w", err)
	}

	if err := c.Mail(m.cfg.fromAddress()); err != nil {
		return fmt.Errorf("mailer: MAIL FROM: %w", err)
	}
	for _, rcpt := range to {
		if err := c.Rcpt(rcpt); err != nil {
			return fmt.Errorf("mailer: RCPT TO %s: %w", rcpt, err)
		}
	}

	wc, err := c.Data()
	if err != nil {
		return fmt.Errorf("mailer: DATA: %w", err)
	}
	if _, err := wc.Write(msg); err != nil {
		wc.Close()
		return fmt.Errorf("mailer: write body: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("mailer: close body: %w", err)
	}

	return c.Quit()
}

// fromAddress strips a display name ("Name <a@b.com>" -> "a@b.com") for the
// SMTP envelope; the full value stays in the From header.
func (c SMTPConfig) fromAddress() string {
	if i := strings.LastIndex(c.From, "<"); i != -1 {
		return strings.TrimSuffix(c.From[i+1:], ">")
	}
	return c.From
}

func buildMessage(from string, to []string, subject, htmlBody, textBody string) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "From: %s\r\n", from)
	fmt.Fprintf(&b, "To: %s\r\n", strings.Join(to, ", "))
	fmt.Fprintf(&b, "Subject: %s\r\n", subject)
	b.WriteString("MIME-Version: 1.0\r\n")

	if htmlBody == "" {
		b.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
		b.WriteString(textBody)
		return []byte(b.String())
	}

	const boundary = "domain-expansion-boundary-6f1e"
	fmt.Fprintf(&b, "Content-Type: multipart/alternative; boundary=%q\r\n\r\n", boundary)

	fmt.Fprintf(&b, "--%s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s\r\n\r\n", boundary, textBody)
	fmt.Fprintf(&b, "--%s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s\r\n\r\n", boundary, htmlBody)
	fmt.Fprintf(&b, "--%s--\r\n", boundary)
	return []byte(b.String())
}
