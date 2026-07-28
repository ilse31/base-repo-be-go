// Package maileroo is an adapter implementing the mailer.Mailer port using the
// official Maileroo Go SDK. It hides SDK-specific types behind the generic
// mailer.Email so application code stays provider-agnostic.
package maileroo

import (
	"context"

	maileroosdk "github.com/maileroo/maileroo-go-sdk/maileroo"

	"github.com/ilse31/base-repo-be-go/pkg/mailer"
)

// Sender wraps the Maileroo SDK client together with a default "from"
// address, fulfilling the mailer.Mailer interface.
type Sender struct {
	client *maileroosdk.Client
	from   maileroosdk.EmailAddress
}

// New creates a Maileroo-backed Mailer. apiKey is the Maileroo API key;
// fromAddress/fromName is the verified sender identity used for all emails.
func New(apiKey, fromAddress, fromName string) (mailer.Mailer, error) {
	client, err := maileroosdk.NewClient(apiKey, 30)
	if err != nil {
		return nil, err
	}
	return &Sender{
		client: client,
		from:   maileroosdk.NewEmail(fromAddress, fromName),
	}, nil
}

// Send delivers a single email via Maileroo. It returns the provider's error
// (if any); the caller decides whether to retry or swallow it.
func (s *Sender) Send(ctx context.Context, msg mailer.Email) error {
	html := maileroosdk.StrPtr(msg.HTML)
	plain := maileroosdk.StrPtr(msg.Plain)

	_, err := s.client.SendBasicEmail(ctx, maileroosdk.BasicEmailData{
		From:    s.from,
		To:      []maileroosdk.EmailAddress{maileroosdk.NewEmail(msg.To.Email, msg.To.Name)},
		Subject: msg.Subject,
		HTML:    html,
		Plain:   plain,
	})
	return err
}
