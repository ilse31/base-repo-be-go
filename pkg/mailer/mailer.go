// Package mailer defines the application's email-sending port (interface).
// It is intentionally transport-agnostic: domain and application code depend
// only on this interface, while concrete providers (e.g. Maileroo) live in
// internal/infrastructure and are injected at the composition root.
package mailer

import "context"

// Address is a single email recipient or sender.
type Address struct {
	Email string
	Name  string
}

// NewAddress is a convenience constructor for Address.
func NewAddress(email, name string) Address {
	return Address{Email: email, Name: name}
}

// Email is a self-contained message to be delivered by a Mailer provider.
type Email struct {
	To      Address
	Subject string
	HTML    string
	Plain   string
}

// Mailer sends transactional emails. Implementations must be safe for
// concurrent use.
type Mailer interface {
	Send(ctx context.Context, msg Email) error
}

// Noop is a Mailer that discards every message. It is useful for local
// development and tests where real delivery is undesirable.
type Noop struct{}

// NewNoop returns a Mailer that performs no work.
func NewNoop() Mailer { return Noop{} }

// Send satisfies the Mailer interface without sending anything.
func (Noop) Send(_ context.Context, _ Email) error { return nil }
