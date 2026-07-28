package events

import (
	"time"

	"github.com/ilse31/base-repo-be-go/internal/shared/domain/valueobjects"
)

// UserRegistered is raised when a new user is registered
type UserRegistered struct {
	*BaseEvent
	UserID     string
	Email      string
	Name       string
	OccurredAt string
}

func NewUserRegistered(userID string, email *valueobjects.Email, name string) *UserRegistered {
	return &UserRegistered{
		BaseEvent:  NewBaseEvent("UserRegistered", userID),
		UserID:     userID,
		Email:      email.String(),
		Name:       name,
		OccurredAt: time.Now().Format(time.RFC3339),
	}
}

// UserLoggedIn is raised when a user logs in
type UserLoggedIn struct {
	*BaseEvent
	UserID    string
	Email     string
	IPAddress string
	UserAgent string
}

func NewUserLoggedIn(userID string, email *valueobjects.Email, ipAddress, userAgent string) *UserLoggedIn {
	return &UserLoggedIn{
		BaseEvent: NewBaseEvent("UserLoggedIn", userID),
		UserID:    userID,
		Email:     email.String(),
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}
}

// UserPasswordChanged is raised when a user changes their password
type UserPasswordChanged struct {
	*BaseEvent
	UserID    string
	ChangedAt string
}

func NewUserPasswordChanged(userID string) *UserPasswordChanged {
	return &UserPasswordChanged{
		BaseEvent: NewBaseEvent("UserPasswordChanged", userID),
		UserID:    userID,
		ChangedAt: time.Now().Format(time.RFC3339),
	}
}

// UserPasswordResetRequested is raised when a user requests a password reset
type UserPasswordResetRequested struct {
	*BaseEvent
	UserID      string
	Email       string
	ResetToken  string
	RequestedAt string
}

func NewUserPasswordResetRequested(userID string, email *valueobjects.Email, resetToken string) *UserPasswordResetRequested {
	return &UserPasswordResetRequested{
		BaseEvent:   NewBaseEvent("UserPasswordResetRequested", userID),
		UserID:      userID,
		Email:       email.String(),
		ResetToken:  resetToken,
		RequestedAt: time.Now().Format(time.RFC3339),
	}
}

// UserLoggedOut is raised when a user logs out
type UserLoggedOut struct {
	*BaseEvent
	UserID      string
	LoggedOutAt string
}

func NewUserLoggedOut(userID string) *UserLoggedOut {
	return &UserLoggedOut{
		BaseEvent:   NewBaseEvent("UserLoggedOut", userID),
		UserID:      userID,
		LoggedOutAt: time.Now().Format(time.RFC3339),
	}
}
