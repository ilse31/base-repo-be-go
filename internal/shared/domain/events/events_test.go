package events

import (
	"testing"

	"github.com/ilse31/base-repo-be-go/internal/shared/domain/valueobjects"
)

func TestDomainEvents(t *testing.T) {
	email, err := valueobjects.NewEmail("user@example.com")
	if err != nil {
		t.Fatalf("failed to create email: %v", err)
	}
	userID := "user-123"

	// BaseEvent
	base := NewBaseEvent("CustomEvent", userID)
	if base.EventType() != "CustomEvent" {
		t.Errorf("expected EventType CustomEvent, got %s", base.EventType())
	}
	if base.AggregateID() != userID {
		t.Errorf("expected AggregateID %s, got %s", userID, base.AggregateID())
	}
	if base.OccurredOn().IsZero() {
		t.Error("expected non-zero OccurredOn time")
	}

	// UserRegistered
	reg := NewUserRegistered(userID, email, "Alice")
	if reg.EventType() != "UserRegistered" || reg.UserID != userID || reg.Email != "user@example.com" || reg.Name != "Alice" || reg.OccurredAt == "" {
		t.Errorf("unexpected UserRegistered event: %+v", reg)
	}

	// UserLoggedIn
	login := NewUserLoggedIn(userID, email, "127.0.0.1", "Mozilla/5.0")
	if login.EventType() != "UserLoggedIn" || login.IPAddress != "127.0.0.1" || login.UserAgent != "Mozilla/5.0" {
		t.Errorf("unexpected UserLoggedIn event: %+v", login)
	}

	// UserPasswordChanged
	pwdChanged := NewUserPasswordChanged(userID)
	if pwdChanged.EventType() != "UserPasswordChanged" || pwdChanged.UserID != userID || pwdChanged.ChangedAt == "" {
		t.Errorf("unexpected UserPasswordChanged event: %+v", pwdChanged)
	}

	// UserPasswordResetRequested
	resetReq := NewUserPasswordResetRequested(userID, email, "token-123")
	if resetReq.EventType() != "UserPasswordResetRequested" || resetReq.ResetToken != "token-123" || resetReq.RequestedAt == "" {
		t.Errorf("unexpected UserPasswordResetRequested event: %+v", resetReq)
	}

	// UserLoggedOut
	logout := NewUserLoggedOut(userID)
	if logout.EventType() != "UserLoggedOut" || logout.LoggedOutAt == "" {
		t.Errorf("unexpected UserLoggedOut event: %+v", logout)
	}
}
