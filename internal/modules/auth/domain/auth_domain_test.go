package domain

import (
	"testing"

	userdomain "github.com/ilse31/base-repo-be-go/internal/modules/user/domain"
	"github.com/ilse31/base-repo-be-go/internal/shared/domain/valueobjects"
)

func TestAuthAggregate(t *testing.T) {
	email, _ := valueobjects.NewEmail("auth@example.com")
	pwd, _ := valueobjects.NewPassword("Password123!")
	userAgg, err := userdomain.NewUserAggregate(email, pwd, "Auth User")
	if err != nil {
		t.Fatalf("failed to create user aggregate: %v", err)
	}

	authAgg := NewAuthAggregate(userAgg)
	if authAgg == nil {
		t.Fatal("expected non-nil auth aggregate")
	}

	if !authAgg.GetUserID().Equals(userAgg.ID()) {
		t.Errorf("expected user ID %v, got %v", userAgg.ID(), authAgg.GetUserID())
	}
	if !authAgg.GetUserEmail().Equals(userAgg.Email()) {
		t.Errorf("expected email %v, got %v", userAgg.Email(), authAgg.GetUserEmail())
	}
	if authAgg.GetUserAggregate() != userAgg {
		t.Error("expected underlying user aggregate match")
	}

	// Authenticate
	if err := authAgg.Authenticate("Password123!"); err != nil {
		t.Errorf("expected authenticate success, got %v", err)
	}

	// Login
	_ = authAgg.PullDomainEvents() // clear
	authAgg.Login("127.0.0.1", "test-agent")
	evs := authAgg.PullDomainEvents()
	if len(evs) != 1 || evs[0].EventType() != "UserLoggedIn" {
		t.Errorf("expected UserLoggedIn event, got %v", evs)
	}

	// Logout
	authAgg.Logout()
	evs = authAgg.PullDomainEvents()
	if len(evs) != 1 || evs[0].EventType() != "UserLoggedOut" {
		t.Errorf("expected UserLoggedOut event, got %v", evs)
	}

	// RequestPasswordReset
	authAgg.RequestPasswordReset("token-xyz")
	// raiseEvent on AuthAggregate is a placeholder, but PullDomainEvents pulls from userAggregate
	_ = authAgg.PullDomainEvents()

	// ResetPassword
	if err := authAgg.ResetPassword("NewPassword123!"); err != nil {
		t.Fatalf("ResetPassword failed: %v", err)
	}
	if err := authAgg.Authenticate("NewPassword123!"); err != nil {
		t.Errorf("expected authenticate with new password to succeed, got %v", err)
	}
}
