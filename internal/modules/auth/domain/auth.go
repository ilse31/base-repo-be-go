package domain

import (
	"context"
	"errors"

	userdomain "github.com/ilse31/base-repo-be-go/internal/modules/user/domain"
	"github.com/ilse31/base-repo-be-go/internal/shared/domain/events"
	"github.com/ilse31/base-repo-be-go/internal/shared/domain/valueobjects"
)

// PasswordResetToken represents a password reset token
type PasswordResetToken struct {
	Token  string
	UserID string
}

// AuthRepository defines the interface for auth-related data operations
type AuthRepository interface {
	StorePasswordResetToken(ctx context.Context, token string, userID string) error
	GetPasswordResetToken(ctx context.Context, token string) (string, error)
	DeletePasswordResetToken(ctx context.Context, token string) error
	StoreRefreshToken(ctx context.Context, token string, userID string) error
	GetRefreshToken(ctx context.Context, token string) (string, error)
	DeleteRefreshToken(ctx context.Context, token string) error
}

// AuthContext represents the Authentication bounded context
// This context handles authentication, authorization, and session management

// AuthAggregate is the aggregate root for the Auth context
type AuthAggregate struct {
	userAggregate *userdomain.UserAggregate
}

// NewAuthAggregate creates a new auth aggregate from a user aggregate
func NewAuthAggregate(userAggregate *userdomain.UserAggregate) *AuthAggregate {
	return &AuthAggregate{
		userAggregate: userAggregate,
	}
}

// Authenticate verifies user credentials
func (a *AuthAggregate) Authenticate(password string) error {
	return a.userAggregate.VerifyPassword(password)
}

// Login authenticates the user and records login event
func (a *AuthAggregate) Login(ipAddress, userAgent string) {
	a.userAggregate.Login(ipAddress, userAgent)
}

// Logout records logout event
func (a *AuthAggregate) Logout() {
	a.userAggregate.Logout()
}

// ResetPassword resets the user's password (for forgot password flow)
func (a *AuthAggregate) ResetPassword(newPassword string) error {
	return a.userAggregate.ResetPassword(newPassword)
}

// RequestPasswordReset initiates password reset process
func (a *AuthAggregate) RequestPasswordReset(resetToken string) {
	a.raiseEvent(events.NewUserPasswordResetRequested(
		a.userAggregate.ID().String(),
		a.userAggregate.Email(),
		resetToken,
	))
}

// GetUserID returns the user ID
func (a *AuthAggregate) GetUserID() *valueobjects.UserID {
	return a.userAggregate.ID()
}

// GetUserEmail returns the user email
func (a *AuthAggregate) GetUserEmail() *valueobjects.Email {
	return a.userAggregate.Email()
}

// PullDomainEvents returns domain events from both auth and user aggregate
func (a *AuthAggregate) PullDomainEvents() []events.DomainEvent {
	return a.userAggregate.PullDomainEvents()
}

func (a *AuthAggregate) raiseEvent(event events.DomainEvent) {
	// Events are raised on the user aggregate
	// This is a placeholder for auth-specific events if needed
}

// GetUserAggregate returns the underlying user aggregate
func (a *AuthAggregate) GetUserAggregate() *userdomain.UserAggregate {
	return a.userAggregate
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)
