package mailer

import (
	"context"
	"testing"
)

func TestAddressAndNoopMailer(t *testing.T) {
	addr := NewAddress("test@example.com", "Test User")
	if addr.Email != "test@example.com" || addr.Name != "Test User" {
		t.Errorf("unexpected Address: %+v", addr)
	}

	m := NewNoop()
	err := m.Send(context.Background(), Email{
		To:      addr,
		Subject: "Subject",
		HTML:    "<p>Hello</p>",
		Plain:   "Hello",
	})

	if err != nil {
		t.Errorf("expected Noop.Send to return nil error, got %v", err)
	}
}
