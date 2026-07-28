package application

import (
	"context"
	"errors"
	"testing"

	userdomain "github.com/ilse31/base-repo-be-go/internal/modules/user/domain"
	"github.com/ilse31/base-repo-be-go/pkg/jwt"
	"github.com/ilse31/base-repo-be-go/pkg/logger"
	"github.com/ilse31/base-repo-be-go/pkg/mailer"
)

type mockUserRepo struct {
	users map[string]*userdomain.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]*userdomain.User)}
}

func (r *mockUserRepo) Create(ctx context.Context, user *userdomain.User) error {
	if user.ID == "" {
		user.ID = "user-123"
	}
	userCopy := *user
	r.users[user.ID] = &userCopy
	return nil
}

func (r *mockUserRepo) GetByID(ctx context.Context, id string) (*userdomain.User, error) {
	u, ok := r.users[id]
	if !ok {
		return nil, errors.New("not found")
	}
	userCopy := *u
	return &userCopy, nil
}

func (r *mockUserRepo) GetByEmail(ctx context.Context, email string) (*userdomain.User, error) {
	for _, u := range r.users {
		if u.Email == email {
			userCopy := *u
			return &userCopy, nil
		}
	}
	return nil, errors.New("not found")
}

func (r *mockUserRepo) Update(ctx context.Context, user *userdomain.User) error {
	if _, ok := r.users[user.ID]; !ok {
		return errors.New("not found")
	}
	userCopy := *user
	r.users[user.ID] = &userCopy
	return nil
}

func (r *mockUserRepo) Delete(ctx context.Context, id string) error {
	delete(r.users, id)
	return nil
}

func (r *mockUserRepo) List(ctx context.Context, limit, offset int) ([]*userdomain.User, error) {
	res := make([]*userdomain.User, 0, len(r.users))
	for _, u := range r.users {
		userCopy := *u
		res = append(res, &userCopy)
	}
	return res, nil
}

type mockAuthRepo struct {
	tokens      map[string]string
	resetTokens map[string]string
}

func newMockAuthRepo() *mockAuthRepo {
	return &mockAuthRepo{
		tokens:      make(map[string]string),
		resetTokens: make(map[string]string),
	}
}

func (r *mockAuthRepo) StorePasswordResetToken(ctx context.Context, token string, userID string) error {
	r.resetTokens[token] = userID
	return nil
}

func (r *mockAuthRepo) GetPasswordResetToken(ctx context.Context, token string) (string, error) {
	uID, ok := r.resetTokens[token]
	if !ok {
		return "", errors.New("token not found")
	}
	return uID, nil
}

func (r *mockAuthRepo) DeletePasswordResetToken(ctx context.Context, token string) error {
	delete(r.resetTokens, token)
	return nil
}

func (r *mockAuthRepo) StoreRefreshToken(ctx context.Context, token string, userID string) error {
	r.tokens[token] = userID
	return nil
}

func (r *mockAuthRepo) GetRefreshToken(ctx context.Context, token string) (string, error) {
	uID, ok := r.tokens[token]
	if !ok {
		return "", errors.New("token not found")
	}
	return uID, nil
}

func (r *mockAuthRepo) DeleteRefreshToken(ctx context.Context, token string) error {
	delete(r.tokens, token)
	return nil
}

func TestAuthService(t *testing.T) {
	_ = logger.Init("development")
	uRepo := newMockUserRepo()
	aRepo := newMockAuthRepo()
	jwtMgr := jwt.NewJWTManager("secret123", 1, 168)
	m := mailer.NewNoop()

	svc := NewAuthService(uRepo, aRepo, jwtMgr, m, "http://localhost:3000/reset-password")
	ctx := context.Background()

	// Register
	u, err := svc.Register(ctx, "Alice", "alice@example.com", "Password123!")
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if u.Email != "alice@example.com" || u.Password != "" {
		t.Errorf("unexpected registered user: %+v", u)
	}

	// Register Duplicate
	_, err = svc.Register(ctx, "Alice2", "alice@example.com", "Password123!")
	if !errors.Is(err, ErrUserAlreadyExists) {
		t.Errorf("expected ErrUserAlreadyExists, got %v", err)
	}

	// Login Invalid Credentials (wrong email)
	_, _, _, err = svc.Login(ctx, "unknown@example.com", "Password123!")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials for unknown user, got %v", err)
	}

	// Login Invalid Credentials (wrong password)
	_, _, _, err = svc.Login(ctx, "alice@example.com", "WrongPassword")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials for wrong password, got %v", err)
	}

	// Login Success
	accessTok, refreshTok, loggedInUser, err := svc.Login(ctx, "alice@example.com", "Password123!")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if accessTok == "" || refreshTok == "" || loggedInUser.ID != u.ID {
		t.Errorf("unexpected login result: %s, %s, %+v", accessTok, refreshTok, loggedInUser)
	}

	// ValidateToken
	claims, err := svc.ValidateToken(accessTok)
	if err != nil || claims.UserID != u.ID {
		t.Errorf("ValidateToken failed: %v, claims: %+v", err, claims)
	}

	// RefreshToken Success
	newAccess, newRefresh, err := svc.RefreshToken(ctx, refreshTok)
	if err != nil {
		t.Fatalf("RefreshToken failed: %v", err)
	}
	if newAccess == "" || newRefresh == "" {
		t.Error("expected non-empty refreshed tokens")
	}

	// RefreshToken Invalid/Used Token
	_, _, err = svc.RefreshToken(ctx, refreshTok)
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken for reused refresh token, got %v", err)
	}

	// ForgotPassword (known user & unknown user)
	err = svc.ForgotPassword(ctx, "unknown@example.com")
	if err != nil {
		t.Errorf("expected nil error for unknown user forgot password, got %v", err)
	}

	err = svc.ForgotPassword(ctx, "alice@example.com")
	if err != nil {
		t.Fatalf("ForgotPassword failed: %v", err)
	}
	// Check reset token in aRepo
	if len(aRepo.resetTokens) != 1 {
		t.Fatalf("expected 1 reset token in repo, got %d", len(aRepo.resetTokens))
	}

	var resetToken string
	for token := range aRepo.resetTokens {
		resetToken = token
	}

	// ResetPassword Invalid Token
	err = svc.ResetPassword(ctx, "invalid-token", "NewPassword123!")
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}

	// ResetPassword Success
	err = svc.ResetPassword(ctx, resetToken, "NewPassword123!")
	if err != nil {
		t.Fatalf("ResetPassword failed: %v", err)
	}

	// Login with new password
	_, _, _, err = svc.Login(ctx, "alice@example.com", "NewPassword123!")
	if err != nil {
		t.Errorf("expected login with new password to succeed, got %v", err)
	}

	// Logout
	err = svc.Logout(ctx, newRefresh)
	if err != nil {
		t.Errorf("Logout failed: %v", err)
	}

	// GenerateSecureToken
	secToken, err := GenerateSecureToken(16)
	if err != nil || secToken == "" {
		t.Errorf("GenerateSecureToken failed: %v", err)
	}
}
