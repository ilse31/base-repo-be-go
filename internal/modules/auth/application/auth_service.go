package application

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	authdomain "github.com/ilse31/base-repo-be-go/internal/modules/auth/domain"
	authinfra "github.com/ilse31/base-repo-be-go/internal/modules/auth/infrastructure"
	userdomain "github.com/ilse31/base-repo-be-go/internal/modules/user/domain"
	"github.com/ilse31/base-repo-be-go/pkg/jwt"
	"github.com/ilse31/base-repo-be-go/pkg/logger"
	"github.com/ilse31/base-repo-be-go/pkg/mailer"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidToken       = errors.New("invalid token")
)

type AuthService struct {
	userRepo   userdomain.UserRepository
	authRepo   authdomain.AuthRepository
	jwtManager *jwt.JWTManager
	mailer     mailer.Mailer
	resetURL   string
}

func NewAuthService(
	userRepo userdomain.UserRepository,
	authRepo authdomain.AuthRepository,
	jwtManager *jwt.JWTManager,
	mailer mailer.Mailer,
	resetURL string,
) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		authRepo:   authRepo,
		jwtManager: jwtManager,
		mailer:     mailer,
		resetURL:   resetURL,
	}
}

// Register creates a new user
func (s *AuthService) Register(ctx context.Context, name, email, password string) (*userdomain.User, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &userdomain.User{
		Name:     name,
		Email:    email,
		Password: string(hashedPassword),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Remove password from response
	user.Password = ""
	return user, nil
}

// Login authenticates a user and returns access and refresh tokens
func (s *AuthService) Login(ctx context.Context, email, password string) (string, string, *userdomain.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", "", nil, ErrInvalidCredentials
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", "", nil, ErrInvalidCredentials
	}

	// Generate access token
	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID, user.Email)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token in Redis
	if err := s.authRepo.StoreRefreshToken(ctx, refreshToken, user.ID); err != nil {
		return "", "", nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	// Remove password from response
	user.Password = ""
	return accessToken, refreshToken, user, nil
}

// ForgotPassword initiates a password reset for the given email. It always
// returns nil so callers cannot tell whether the address exists (security).
// When the user exists, a reset token is stored and a reset link is emailed.
func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil // user not found: succeed silently to prevent enumeration
	}

	// Generate reset token
	token := authinfra.GenerateResetToken()

	// Store token in Redis
	if err := s.authRepo.StorePasswordResetToken(ctx, token, user.ID); err != nil {
		return fmt.Errorf("failed to store reset token: %w", err)
	}

	// Email the reset link (best-effort: never fail the request on mail errors)
	resetLink := fmt.Sprintf("%s?token=%s", s.resetURL, token)
	resetData := PasswordResetData{ResetLink: resetLink}

	htmlBody, err := passwordResetHTML(resetData)
	if err != nil {
		logger.Error("failed to render password reset email (html)",
			zap.String("user_id", user.ID),
			zap.Error(err),
		)
		return nil
	}
	plainBody, err := passwordResetPlain(resetData)
	if err != nil {
		logger.Error("failed to render password reset email (plain)",
			zap.String("user_id", user.ID),
			zap.Error(err),
		)
		return nil
	}

	if err := s.mailer.Send(ctx, mailer.Email{
		To:      mailer.NewAddress(user.Email, user.Name),
		Subject: passwordResetSubject,
		HTML:    htmlBody,
		Plain:   plainBody,
	}); err != nil {
		logger.Error("failed to send password reset email",
			zap.String("user_id", user.ID),
			zap.Error(err),
		)
	}

	return nil
}

// ResetPassword resets a user's password using a reset token
func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	// Validate token and get user ID
	userID, err := s.authRepo.GetPasswordResetToken(ctx, token)
	if err != nil {
		return ErrInvalidToken
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return ErrInvalidToken
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	user.Password = string(hashedPassword)
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Delete the reset token
	_ = s.authRepo.DeletePasswordResetToken(ctx, token)

	return nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *AuthService) ValidateToken(tokenString string) (*jwt.Claims, error) {
	return s.jwtManager.ValidateToken(tokenString)
}

// RefreshToken generates new access and refresh tokens using a valid refresh token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	// Validate refresh token
	claims, err := s.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		return "", "", ErrInvalidToken
	}

	// Check if refresh token exists in Redis
	userID, err := s.authRepo.GetRefreshToken(ctx, refreshToken)
	if err != nil || userID != claims.UserID {
		return "", "", ErrInvalidToken
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", "", ErrInvalidToken
	}

	// Generate new access token
	newAccessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate new refresh token (token rotation)
	newRefreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID, user.Email)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store new refresh token in Redis
	if err := s.authRepo.StoreRefreshToken(ctx, newRefreshToken, user.ID); err != nil {
		return "", "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	// Delete old refresh token
	_ = s.authRepo.DeleteRefreshToken(ctx, refreshToken)

	return newAccessToken, newRefreshToken, nil
}

// Logout invalidates the refresh token
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	return s.authRepo.DeleteRefreshToken(ctx, refreshToken)
}

// GenerateSecureToken generates a secure random token
func GenerateSecureToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
