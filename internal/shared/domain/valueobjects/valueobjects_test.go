package valueobjects

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestEmail(t *testing.T) {
	// Valid email
	e1, err := NewEmail("  Test.User@Domain.COM  ")
	if err != nil {
		t.Fatalf("NewEmail failed: %v", err)
	}
	if e1.String() != "test.user@domain.com" {
		t.Errorf("expected test.user@domain.com, got %s", e1.String())
	}

	// Equals
	e2, _ := NewEmail("test.user@domain.com")
	e3, _ := NewEmail("other@domain.com")
	if !e1.Equals(e2) {
		t.Error("expected e1 to equal e2")
	}
	if e1.Equals(e3) {
		t.Error("expected e1 not to equal e3")
	}
	if e1.Equals(nil) {
		t.Error("expected e1.Equals(nil) to be false")
	}

	// Invalid emails
	invalidEmails := []string{"", "invalid", "user@", "@domain.com", "user@domain"}
	for _, inv := range invalidEmails {
		if _, err := NewEmail(inv); !errors.Is(err, ErrInvalidEmail) {
			t.Errorf("expected ErrInvalidEmail for %q, got %v", inv, err)
		}
	}
}

func TestPassword(t *testing.T) {
	// Too short
	_, err := NewPassword("short")
	if !errors.Is(err, ErrPasswordTooShort) {
		t.Errorf("expected ErrPasswordTooShort, got %v", err)
	}

	// Valid password
	p1, err := NewPassword("Password123!")
	if err != nil {
		t.Fatalf("NewPassword failed: %v", err)
	}
	if p1.Hash() == "" {
		t.Error("expected non-empty hash")
	}

	// Verify
	if err := p1.Verify("Password123!"); err != nil {
		t.Errorf("expected verify success, got %v", err)
	}
	if err := p1.Verify("WrongPassword"); !errors.Is(err, ErrPasswordMismatch) {
		t.Errorf("expected ErrPasswordMismatch, got %v", err)
	}

	// NewPasswordFromHash
	_, err = NewPasswordFromHash("")
	if !errors.Is(err, ErrInvalidPassword) {
		t.Errorf("expected ErrInvalidPassword, got %v", err)
	}
	p2, err := NewPasswordFromHash(p1.Hash())
	if err != nil {
		t.Fatalf("NewPasswordFromHash failed: %v", err)
	}
	if !p1.Equals(p2) {
		t.Error("expected p1 to equal p2")
	}
	if p1.Equals(nil) {
		t.Error("expected p1.Equals(nil) to be false")
	}

	// Random password generation
	randPwd, err := GenerateRandomPassword(16)
	if err != nil {
		t.Fatalf("GenerateRandomPassword failed: %v", err)
	}
	if len(randPwd) < 16 {
		t.Errorf("expected random password length at least 16, got %d", len(randPwd))
	}

	// Small length falls back to 16
	randPwd2, _ := GenerateRandomPassword(4)
	if len(randPwd2) < 16 {
		t.Errorf("expected fallback length 16, got %d", len(randPwd2))
	}

	// simple password fallback test
	simplePwd, err := generateSimplePassword(12)
	if err != nil || len(simplePwd) != 12 {
		t.Errorf("generateSimplePassword failed, len: %d, err: %v", len(simplePwd), err)
	}
}

func TestUserID(t *testing.T) {
	// Empty or invalid UUID
	_, err := NewUserID("")
	if !errors.Is(err, ErrInvalidUserID) {
		t.Errorf("expected ErrInvalidUserID for empty string, got %v", err)
	}
	_, err = NewUserID("invalid-uuid")
	if !errors.Is(err, ErrInvalidUserID) {
		t.Errorf("expected ErrInvalidUserID for invalid-uuid, got %v", err)
	}

	// Valid UUID
	rawUUID := uuid.New()
	u1, err := NewUserID(rawUUID.String())
	if err != nil {
		t.Fatalf("NewUserID failed: %v", err)
	}
	if u1.String() != rawUUID.String() {
		t.Errorf("expected %s, got %s", rawUUID.String(), u1.String())
	}

	// GenerateUserID
	u2 := GenerateUserID()
	if u2.String() == "" {
		t.Error("expected non-empty generated user id")
	}

	u3 := NewUserIDFromUUID(rawUUID)
	if !u1.Equals(u3) {
		t.Error("expected u1 to equal u3")
	}
	if u1.Equals(u2) {
		t.Error("expected u1 not to equal u2")
	}
	if u1.Equals(nil) {
		t.Error("expected u1.Equals(nil) to be false")
	}
}
