// Package mailer is the application's outbound email gateway.
//
// One SMTP implementation works for both environments: dev points at
// Mailpit (plain text, no auth, port 1025) and prod points at AWS SES
// SMTP (TLS + STARTTLS auth on port 587). All toggles are env-driven via
// platform/config — no separate build tag, no separate impl.
//
// The Mailer interface deliberately accepts only the small payload the
// app actually needs (to, subject, html, text). Templating lives in the
// domain modules that own the message — auth owns the verify-email body,
// not this package.
package mailer

import (
	"context"
	"fmt"
	"log/slog"

	gomail "github.com/wneessen/go-mail"

	"github.com/apudiu/quranprism/api/internal/platform/config"
)

// Mail is the payload accepted by Mailer.Send.
type Mail struct {
	To      string
	Subject string
	HTML    string // optional — at least one of HTML or Text must be set
	Text    string // optional — falls back to a stripped version of HTML if blank
}

// Mailer is the abstraction every module talks to.
//
// Test doubles implement this interface directly; production wires the
// SMTP-backed impl provided by NewSMTP.
type Mailer interface {
	Send(ctx context.Context, m Mail) error
}

type smtpMailer struct {
	client *gomail.Client
	from   string
	log    *slog.Logger
}

// NewSMTP builds the SMTP-backed Mailer from configured connection
// details. Returned via the Mailer interface so callers can be doubled.
func NewSMTP(cfg *config.Config, log *slog.Logger) (Mailer, error) {
	opts := []gomail.Option{
		gomail.WithPort(cfg.Mailer.SMTPPort),
		gomail.WithTLSPolicy(tlsPolicy(cfg.Mailer.TLSPolicy)),
	}
	if cfg.Mailer.User != "" {
		opts = append(opts,
			gomail.WithSMTPAuth(gomail.SMTPAuthLogin),
			gomail.WithUsername(cfg.Mailer.User),
			gomail.WithPassword(cfg.Mailer.Password),
		)
	}
	c, err := gomail.NewClient(cfg.Mailer.SMTPHost, opts...)
	if err != nil {
		return nil, fmt.Errorf("mailer: client: %w", err)
	}
	return &smtpMailer{
		client: c,
		from:   cfg.Mailer.From,
		log:    log.With("subsystem", "mailer"),
	}, nil
}

func (m *smtpMailer) Send(ctx context.Context, mail Mail) error {
	if mail.HTML == "" && mail.Text == "" {
		return fmt.Errorf("mailer: empty body")
	}

	msg := gomail.NewMsg()
	if err := msg.From(m.from); err != nil {
		return fmt.Errorf("mailer: from: %w", err)
	}
	if err := msg.To(mail.To); err != nil {
		return fmt.Errorf("mailer: to: %w", err)
	}
	msg.Subject(mail.Subject)

	if mail.HTML != "" {
		msg.SetBodyString(gomail.TypeTextHTML, mail.HTML)
	}
	if mail.Text != "" {
		if mail.HTML != "" {
			msg.AddAlternativeString(gomail.TypeTextPlain, mail.Text)
		} else {
			msg.SetBodyString(gomail.TypeTextPlain, mail.Text)
		}
	}

	if err := m.client.DialAndSendWithContext(ctx, msg); err != nil {
		return fmt.Errorf("mailer: send: %w", err)
	}
	m.log.Debug("mail sent", "to", mail.To, "subject", mail.Subject)
	return nil
}

func tlsPolicy(s string) gomail.TLSPolicy {
	switch s {
	case "mandatory":
		return gomail.TLSMandatory
	case "opportunistic":
		return gomail.TLSOpportunistic
	default:
		return gomail.NoTLS
	}
}
