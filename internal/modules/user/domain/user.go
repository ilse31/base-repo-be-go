package domain

import (
	"context"
	"errors"
	"time"

	"github.com/ilse31/base-repo-be-go/internal/shared/domain"
	"github.com/ilse31/base-repo-be-go/internal/shared/domain/events"
	"github.com/ilse31/base-repo-be-go/internal/shared/domain/valueobjects"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// UserAggregate represents a user aggregate with rich domain model
type UserAggregate struct {
	id           *valueobjects.UserID
	email        *valueobjects.Email
	password     *valueobjects.Password
	name         string
	createdAt    time.Time
	updatedAt    time.Time
	domainEvents []events.DomainEvent
}

// NewUserAggregate creates a new user aggregate with the given parameters
func NewUserAggregate(email *valueobjects.Email, password *valueobjects.Password, name string) (*UserAggregate, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	user := &UserAggregate{
		id:           valueobjects.GenerateUserID(),
		email:        email,
		password:     password,
		name:         name,
		createdAt:    time.Now(),
		updatedAt:    time.Now(),
		domainEvents: []events.DomainEvent{},
	}

	// Raise UserRegistered event
	user.raiseEvent(events.NewUserRegistered(user.id.String(), email, name))

	return user, nil
}

// ReconstructUserAggregate reconstructs a user aggregate from persistence (for repository)
func ReconstructUserAggregate(id *valueobjects.UserID, email *valueobjects.Email, password *valueobjects.Password, name string, createdAt, updatedAt time.Time) *UserAggregate {
	return &UserAggregate{
		id:           id,
		email:        email,
		password:     password,
		name:         name,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
		domainEvents: []events.DomainEvent{},
	}
}

// Getters
func (u *UserAggregate) ID() *valueobjects.UserID {
	return u.id
}

func (u *UserAggregate) Email() *valueobjects.Email {
	return u.email
}

func (u *UserAggregate) Name() string {
	return u.name
}

func (u *UserAggregate) CreatedAt() time.Time {
	return u.createdAt
}

func (u *UserAggregate) UpdatedAt() time.Time {
	return u.updatedAt
}

// Business Logic Methods

// VerifyPassword verifies if the given password matches the user's password
func (u *UserAggregate) VerifyPassword(plainPassword string) error {
	return u.password.Verify(plainPassword)
}

// ChangePassword changes the user's password
func (u *UserAggregate) ChangePassword(oldPassword, newPassword string) error {
	// Verify old password
	if err := u.VerifyPassword(oldPassword); err != nil {
		return ErrInvalidCredentials
	}

	// Create new password value object
	newPasswordVO, err := valueobjects.NewPassword(newPassword)
	if err != nil {
		return err
	}

	u.password = newPasswordVO
	u.updatedAt = time.Now()

	// Raise UserPasswordChanged event
	u.raiseEvent(events.NewUserPasswordChanged(u.id.String()))

	return nil
}

// ResetPassword resets the user's password (used in forgot password flow)
func (u *UserAggregate) ResetPassword(newPassword string) error {
	newPasswordVO, err := valueobjects.NewPassword(newPassword)
	if err != nil {
		return err
	}

	u.password = newPasswordVO
	u.updatedAt = time.Now()

	// Raise UserPasswordChanged event
	u.raiseEvent(events.NewUserPasswordChanged(u.id.String()))

	return nil
}

// UpdateName updates the user's name
func (u *UserAggregate) UpdateName(newName string) error {
	if newName == "" {
		return errors.New("name cannot be empty")
	}

	u.name = newName
	u.updatedAt = time.Now()

	return nil
}

// Login records a login event for the user
func (u *UserAggregate) Login(ipAddress, userAgent string) {
	u.raiseEvent(events.NewUserLoggedIn(u.id.String(), u.email, ipAddress, userAgent))
}

// Logout records a logout event for the user
func (u *UserAggregate) Logout() {
	u.raiseEvent(events.NewUserLoggedOut(u.id.String()))
}

// Domain Events

// PullDomainEvents returns and clears the domain events
func (u *UserAggregate) PullDomainEvents() []events.DomainEvent {
	domainEvents := u.domainEvents
	u.domainEvents = []events.DomainEvent{}
	return domainEvents
}

// raiseEvent adds a domain event to the user
func (u *UserAggregate) raiseEvent(event events.DomainEvent) {
	u.domainEvents = append(u.domainEvents, event)
}

// Persistence Methods

// ToDTO converts the user aggregate to a DTO for persistence/transport
func (u *UserAggregate) ToDTO() map[string]interface{} {
	return map[string]interface{}{
		"id":         u.id.String(),
		"email":      u.email.String(),
		"name":       u.name,
		"created_at": u.createdAt,
		"updated_at": u.updatedAt,
	}
}

// ToEntity converts the aggregate to the old User entity for backward compatibility
func (u *UserAggregate) ToEntity() User {
	return User{
		BaseEntity: domain.BaseEntity{
			ID:        u.id.String(),
			CreatedAt: u.createdAt,
			UpdatedAt: u.updatedAt,
		},
		Name:     u.name,
		Email:    u.email.String(),
		Password: u.password.Hash(),
	}
}

// User represents a user entity for persistence (backward compatibility)
type User struct {
	domain.BaseEntity
	Name     string `bun:"name,notnull" json:"name" validate:"required"`
	Email    string `bun:"email,notnull,unique" json:"email" validate:"required,email"`
	Password string `bun:"password,notnull" json:"-" validate:"required"`
}

// UserRepository defines the interface for user data operations (backward compatibility)
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*User, error)
}

// UserAggregateRepository defines the interface for user aggregate operations (DDD)
type UserAggregateRepository interface {
	Save(ctx context.Context, aggregate *UserAggregate) error
	GetByID(ctx context.Context, id string) (*UserAggregate, error)
	GetByEmail(ctx context.Context, email string) (*UserAggregate, error)
	Delete(ctx context.Context, id string) error
}
