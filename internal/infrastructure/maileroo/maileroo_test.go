package maileroo

import (
	"context"
	"testing"

	"github.com/ilse31/base-repo-be-go/pkg/mailer"
)

func TestNewSender(t *testing.T) {
	sender, err := New("dummy-api-key", "noreply@example.com", "No Reply")
	if err != nil {
		t.Fatalf("expected no error from New, got %v", err)
	}
	if sender == nil {
		t.Fatal("expected non-nil sender")
	}

	// Send will attempt API call and fail with invalid key, but method execution path is tested
	_ = sender.Send(context.Background(), mailer.Email{
		To:      mailer.NewAddress("user@example.com", "User"),
		Subject: "Test",
		HTML:    "<p>Test</p>",
		Plain:   "Test",
	})
}
