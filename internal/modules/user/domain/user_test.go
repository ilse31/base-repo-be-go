package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/ilse31/base-repo-be-go/internal/shared/domain/valueobjects"
)

func TestUserAggregate(t *testing.T) {
	email, _ := valueobjects.NewEmail("user@example.com")
	password, _ := valueobjects.NewPassword("Password123!")

	t.Run("NewUserAggregate valid", func(t *testing.T) {
		agg, err := NewUserAggregate(email, password, "Alice")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if agg.Name() != "Alice" {
			t.Errorf("expected name Alice, got %s", agg.Name())
		}
		if !agg.Email().Equals(email) {
			t.Errorf("expected email %v, got %v", email, agg.Email())
		}
		if agg.ID() == nil {
			t.Error("expected non-nil ID")
		}
		if agg.CreatedAt().IsZero() || agg.UpdatedAt().IsZero() {
			t.Error("expected non-zero timestamps")
		}

		// Events
		events := agg.PullDomainEvents()
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}
		if events[0].EventType() != "UserRegistered" {
			t.Errorf("expected UserRegistered event, got %s", events[0].EventType())
		}
		// Pulling again clears events
		if len(agg.PullDomainEvents()) != 0 {
			t.Error("expected empty events after pull")
		}
	})

	t.Run("NewUserAggregate empty name", func(t *testing.T) {
		_, err := NewUserAggregate(email, password, "")
		if err == nil {
			t.Error("expected error for empty name, got nil")
		}
	})

	t.Run("ReconstructUserAggregate", func(t *testing.T) {
		userID := valueobjects.GenerateUserID()
		now := time.Now()
		agg := ReconstructUserAggregate(userID, email, password, "Bob", now, now)

		if !agg.ID().Equals(userID) || agg.Name() != "Bob" {
			t.Errorf("unexpected reconstructed aggregate: %+v", agg)
		}
	})

	t.Run("VerifyPassword, ChangePassword, ResetPassword", func(t *testing.T) {
		agg, _ := NewUserAggregate(email, password, "Alice")
		_ = agg.PullDomainEvents() // clear initial event

		if err := agg.VerifyPassword("Password123!"); err != nil {
			t.Errorf("expected verify success, got %v", err)
		}

		// ChangePassword invalid old password
		if err := agg.ChangePassword("WrongPassword", "NewPassword123!"); !errors.Is(err, ErrInvalidCredentials) {
			t.Errorf("expected ErrInvalidCredentials, got %v", err)
		}

		// ChangePassword invalid new password (too short)
		if err := agg.ChangePassword("Password123!", "short"); err == nil {
			t.Error("expected error for short new password, got nil")
		}

		// ChangePassword success
		if err := agg.ChangePassword("Password123!", "NewPassword123!"); err != nil {
			t.Fatalf("ChangePassword failed: %v", err)
		}
		if err := agg.VerifyPassword("NewPassword123!"); err != nil {
			t.Errorf("expected new password to be verified, got %v", err)
		}
		evs := agg.PullDomainEvents()
		if len(evs) != 1 || evs[0].EventType() != "UserPasswordChanged" {
			t.Errorf("expected UserPasswordChanged event, got %v", evs)
		}

		// ResetPassword invalid new password
		if err := agg.ResetPassword("short"); err == nil {
			t.Error("expected error for short reset password, got nil")
		}

		// ResetPassword success
		if err := agg.ResetPassword("ResetPassword123!"); err != nil {
			t.Fatalf("ResetPassword failed: %v", err)
		}
		if err := agg.VerifyPassword("ResetPassword123!"); err != nil {
			t.Errorf("expected reset password to verify, got %v", err)
		}
	})

	t.Run("UpdateName", func(t *testing.T) {
		agg, _ := NewUserAggregate(email, password, "Alice")
		if err := agg.UpdateName(""); err == nil {
			t.Error("expected error for empty name update, got nil")
		}
		if err := agg.UpdateName("Charlie"); err != nil {
			t.Errorf("UpdateName failed: %v", err)
		}
		if agg.Name() != "Charlie" {
			t.Errorf("expected Charlie, got %s", agg.Name())
		}
	})

	t.Run("Login and Logout", func(t *testing.T) {
		agg, _ := NewUserAggregate(email, password, "Alice")
		_ = agg.PullDomainEvents()

		agg.Login("127.0.0.1", "test-agent")
		evs := agg.PullDomainEvents()
		if len(evs) != 1 || evs[0].EventType() != "UserLoggedIn" {
			t.Errorf("expected UserLoggedIn event, got %v", evs)
		}

		agg.Logout()
		evs = agg.PullDomainEvents()
		if len(evs) != 1 || evs[0].EventType() != "UserLoggedOut" {
			t.Errorf("expected UserLoggedOut event, got %v", evs)
		}
	})

	t.Run("ToDTO and ToEntity", func(t *testing.T) {
		agg, _ := NewUserAggregate(email, password, "Alice")
		dto := agg.ToDTO()
		if dto["name"] != "Alice" || dto["email"] != "user@example.com" {
			t.Errorf("unexpected DTO: %+v", dto)
		}

		entity := agg.ToEntity()
		if entity.Name != "Alice" || entity.Email != "user@example.com" || entity.ID != agg.ID().String() {
			t.Errorf("unexpected entity: %+v", entity)
		}
	})
}
