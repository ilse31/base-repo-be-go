package service

import (
	"context"
	"errors"
	"testing"

	authdomain "github.com/ilse31/base-repo-be-go/internal/modules/auth/domain"
	userdomain "github.com/ilse31/base-repo-be-go/internal/modules/user/domain"
)

type mockUserAggregateRepo struct {
	users map[string]*userdomain.UserAggregate
}

func newMockUserAggregateRepo() *mockUserAggregateRepo {
	return &mockUserAggregateRepo{users: make(map[string]*userdomain.UserAggregate)}
}

func (r *mockUserAggregateRepo) Save(ctx context.Context, aggregate *userdomain.UserAggregate) error {
	r.users[aggregate.ID().String()] = aggregate
	return nil
}

func (r *mockUserAggregateRepo) GetByID(ctx context.Context, id string) (*userdomain.UserAggregate, error) {
	u, ok := r.users[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return u, nil
}

func (r *mockUserAggregateRepo) GetByEmail(ctx context.Context, email string) (*userdomain.UserAggregate, error) {
	for _, u := range r.users {
		if u.Email().String() == email {
			return u, nil
		}
	}
	return nil, errors.New("not found")
}

func (r *mockUserAggregateRepo) Delete(ctx context.Context, id string) error {
	delete(r.users, id)
	return nil
}

func TestAuthenticationDomainService(t *testing.T) {
	repo := newMockUserAggregateRepo()
	svc := NewAuthenticationDomainService(repo)
	ctx := context.Background()

	// RegisterUser invalid email
	_, err := svc.RegisterUser(ctx, "invalid-email", "Password123!", "Alice")
	if err == nil {
		t.Error("expected error for invalid email")
	}

	// RegisterUser invalid password
	_, err = svc.RegisterUser(ctx, "alice@example.com", "short", "Alice")
	if err == nil {
		t.Error("expected error for short password")
	}

	// RegisterUser success
	authAgg, err := svc.RegisterUser(ctx, "alice@example.com", "Password123!", "Alice")
	if err != nil {
		t.Fatalf("RegisterUser failed: %v", err)
	}
	if authAgg.GetUserEmail().String() != "alice@example.com" {
		t.Errorf("unexpected user email: %s", authAgg.GetUserEmail().String())
	}

	// RegisterUser existing user
	_, err = svc.RegisterUser(ctx, "alice@example.com", "Password123!", "Alice")
	if err == nil {
		t.Error("expected error for duplicate user registration, got nil")
	}

	// AuthenticateUser invalid email
	_, err = svc.AuthenticateUser(ctx, "invalid-email", "Password123!")
	if err == nil {
		t.Error("expected error for invalid email format, got nil")
	}

	// AuthenticateUser user not found
	_, err = svc.AuthenticateUser(ctx, "unknown@example.com", "Password123!")
	if !errors.Is(err, authdomain.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials for unknown email, got %v", err)
	}

	// AuthenticateUser wrong password
	_, err = svc.AuthenticateUser(ctx, "alice@example.com", "WrongPassword")
	if !errors.Is(err, authdomain.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials for wrong password, got %v", err)
	}

	// AuthenticateUser success
	authAgg2, err := svc.AuthenticateUser(ctx, "alice@example.com", "Password123!")
	if err != nil {
		t.Fatalf("AuthenticateUser failed: %v", err)
	}
	if !authAgg2.GetUserID().Equals(authAgg.GetUserID()) {
		t.Errorf("expected user ID %v, got %v", authAgg.GetUserID(), authAgg2.GetUserID())
	}

	// ResetUserPassword user not found
	err = svc.ResetUserPassword(ctx, "non-existent-id", "NewPassword123!")
	if err == nil {
		t.Error("expected error for non-existent user id")
	}

	// ResetUserPassword short new password
	userID := authAgg.GetUserID().String()
	err = svc.ResetUserPassword(ctx, userID, "short")
	if err == nil {
		t.Error("expected error for short new password")
	}

	// ResetUserPassword success
	err = svc.ResetUserPassword(ctx, userID, "NewPassword123!")
	if err != nil {
		t.Fatalf("ResetUserPassword failed: %v", err)
	}
	_, err = svc.AuthenticateUser(ctx, "alice@example.com", "NewPassword123!")
	if err != nil {
		t.Errorf("expected authenticate with new password to succeed, got %v", err)
	}
}
