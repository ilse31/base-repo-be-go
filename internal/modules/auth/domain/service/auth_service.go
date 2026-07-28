package service

import (
	"context"
	"errors"

	authdomain "github.com/ilse31/base-repo-be-go/internal/modules/auth/domain"
	userdomain "github.com/ilse31/base-repo-be-go/internal/modules/user/domain"
	"github.com/ilse31/base-repo-be-go/internal/shared/domain/valueobjects"
)

// AuthenticationDomainService handles authentication logic that spans across aggregates
type AuthenticationDomainService struct {
	userRepository userdomain.UserAggregateRepository
}

func NewAuthenticationDomainService(userRepository userdomain.UserAggregateRepository) *AuthenticationDomainService {
	return &AuthenticationDomainService{
		userRepository: userRepository,
	}
}

// RegisterUser creates a new user in the system
func (s *AuthenticationDomainService) RegisterUser(ctx context.Context, email, password, name string) (*authdomain.AuthAggregate, error) {
	// Create value objects
	emailVO, err := valueobjects.NewEmail(email)
	if err != nil {
		return nil, err
	}

	passwordVO, err := valueobjects.NewPassword(password)
	if err != nil {
		return nil, err
	}

	// Check if user already exists
	existingUser, err := s.userRepository.GetByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, errors.New("user already exists")
	}

	// Create user aggregate
	userAggregate, err := userdomain.NewUserAggregate(emailVO, passwordVO, name)
	if err != nil {
		return nil, err
	}

	// Save user
	if err := s.userRepository.Save(ctx, userAggregate); err != nil {
		return nil, err
	}

	// Create auth aggregate
	authAggregate := authdomain.NewAuthAggregate(userAggregate)

	return authAggregate, nil
}

// AuthenticateUser authenticates a user with email and password
func (s *AuthenticationDomainService) AuthenticateUser(ctx context.Context, email, password string) (*authdomain.AuthAggregate, error) {
	// Create email value object
	emailVO, err := valueobjects.NewEmail(email)
	if err != nil {
		return nil, err
	}

	// Get user by email
	userAggregate, err := s.userRepository.GetByEmail(ctx, email)
	if err != nil {
		return nil, authdomain.ErrInvalidCredentials
	}

	// Verify email matches
	if !userAggregate.Email().Equals(emailVO) {
		return nil, authdomain.ErrInvalidCredentials
	}

	// Create auth aggregate
	authAggregate := authdomain.NewAuthAggregate(userAggregate)

	// Authenticate
	if err := authAggregate.Authenticate(password); err != nil {
		return nil, authdomain.ErrInvalidCredentials
	}

	return authAggregate, nil
}

// ResetUserPassword resets a user's password using a reset token
func (s *AuthenticationDomainService) ResetUserPassword(ctx context.Context, userID, newPassword string) error {
	// Get user by ID
	userAggregate, err := s.userRepository.GetByID(ctx, userID)
	if err != nil {
		return errors.New("user not found")
	}

	// Reset password
	if err := userAggregate.ResetPassword(newPassword); err != nil {
		return err
	}

	// Save user
	return s.userRepository.Save(ctx, userAggregate)
}
