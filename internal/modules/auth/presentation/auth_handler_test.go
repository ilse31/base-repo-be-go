package presentation

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/ilse31/base-repo-be-go/internal/modules/auth/application"
	userdomain "github.com/ilse31/base-repo-be-go/internal/modules/user/domain"
	"github.com/ilse31/base-repo-be-go/internal/shared/validation"
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

func TestAuthHandler(t *testing.T) {
	_ = logger.Init("development")
	v := validation.New()
	uRepo := newMockUserRepo()
	aRepo := newMockAuthRepo()
	jwtMgr := jwt.NewJWTManager("secret123", 1, 168)
	m := mailer.NewNoop()
	svc := application.NewAuthService(uRepo, aRepo, jwtMgr, m, "http://localhost:3000/reset-password")

	handler := NewAuthHandler(svc, v, false)
	e := echo.New()

	t.Run("Register invalid json and validation error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString("invalid json"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.Register(c)
		if err == nil {
			t.Error("expected error for invalid json")
		}

		reqPayload, _ := json.Marshal(RegisterRequest{Name: "Alice", Email: "invalid", Password: "short"})
		req2 := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(reqPayload))
		req2.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec2 := httptest.NewRecorder()
		c2 := e.NewContext(req2, rec2)

		err = handler.Register(c2)
		if err == nil {
			t.Error("expected validation error")
		}
	})

	t.Run("Register Success", func(t *testing.T) {
		reqPayload, _ := json.Marshal(RegisterRequest{
			Name:     "Alice",
			Email:    "alice@example.com",
			Password: "Password123!",
		})
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(reqPayload))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.Register(c)
		if err != nil {
			t.Fatalf("Register failed: %v", err)
		}
		if rec.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", rec.Code)
		}
	})

	t.Run("Login invalid json & success", func(t *testing.T) {
		// Invalid JSON
		reqBad := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString("invalid json"))
		reqBad.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		recBad := httptest.NewRecorder()
		cBad := e.NewContext(reqBad, recBad)
		if err := handler.Login(cBad); err == nil {
			t.Error("expected error for invalid json")
		}

		// Validation error
		reqPayloadVal, _ := json.Marshal(LoginRequest{Email: "invalid-email", Password: ""})
		reqVal := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(reqPayloadVal))
		reqVal.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		recVal := httptest.NewRecorder()
		cVal := e.NewContext(reqVal, recVal)
		if err := handler.Login(cVal); err == nil {
			t.Error("expected validation error")
		}

		// Success
		reqPayload, _ := json.Marshal(LoginRequest{
			Email:    "alice@example.com",
			Password: "Password123!",
		})
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(reqPayload))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.Login(c)
		if err != nil {
			t.Fatalf("Login failed: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
		// Verify cookies were set
		cookies := rec.Result().Cookies()
		if len(cookies) < 2 {
			t.Errorf("expected cookies to be set, got %d", len(cookies))
		}
	})

	t.Run("Me", func(t *testing.T) {
		// Missing cookie
		reqNoCookie := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
		recNoCookie := httptest.NewRecorder()
		cNoCookie := e.NewContext(reqNoCookie, recNoCookie)
		if err := handler.Me(cNoCookie); err == nil {
			t.Error("expected unauthorized error for missing cookie")
		}

		// Invalid cookie
		reqBadCookie := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
		reqBadCookie.AddCookie(&http.Cookie{Name: "access_token", Value: "invalid-token"})
		recBadCookie := httptest.NewRecorder()
		cBadCookie := e.NewContext(reqBadCookie, recBadCookie)
		if err := handler.Me(cBadCookie); err == nil {
			t.Error("expected unauthorized error for invalid token cookie")
		}

		// Success
		tok, _ := jwtMgr.GenerateAccessToken("user-123", "alice@example.com")
		req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
		req.AddCookie(&http.Cookie{Name: "access_token", Value: tok})
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.Me(c)
		if err != nil {
			t.Fatalf("Me failed: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("ForgotPassword", func(t *testing.T) {
		// Invalid body
		reqBad := httptest.NewRequest(http.MethodPost, "/auth/forgot-password", bytes.NewBufferString("bad json"))
		recBad := httptest.NewRecorder()
		cBad := e.NewContext(reqBad, recBad)
		if err := handler.ForgotPassword(cBad); err == nil {
			t.Error("expected error for bad json")
		}

		// Validation error
		reqPayloadVal, _ := json.Marshal(ForgotPasswordRequest{Email: "invalid"})
		reqVal := httptest.NewRequest(http.MethodPost, "/auth/forgot-password", bytes.NewBuffer(reqPayloadVal))
		reqVal.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		recVal := httptest.NewRecorder()
		cVal := e.NewContext(reqVal, recVal)
		if err := handler.ForgotPassword(cVal); err == nil {
			t.Error("expected validation error")
		}

		// Success
		reqPayload, _ := json.Marshal(ForgotPasswordRequest{Email: "alice@example.com"})
		req := httptest.NewRequest(http.MethodPost, "/auth/forgot-password", bytes.NewBuffer(reqPayload))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.ForgotPassword(c)
		if err != nil {
			t.Fatalf("ForgotPassword failed: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("ResetPassword", func(t *testing.T) {
		// Invalid body
		reqBad := httptest.NewRequest(http.MethodPost, "/auth/reset-password", bytes.NewBufferString("bad json"))
		recBad := httptest.NewRecorder()
		cBad := e.NewContext(reqBad, recBad)
		if err := handler.ResetPassword(cBad); err == nil {
			t.Error("expected error for bad json")
		}

		// Validation error
		reqPayloadVal, _ := json.Marshal(ResetPasswordRequest{Token: "tok", NewPassword: "short"})
		reqVal := httptest.NewRequest(http.MethodPost, "/auth/reset-password", bytes.NewBuffer(reqPayloadVal))
		reqVal.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		recVal := httptest.NewRecorder()
		cVal := e.NewContext(reqVal, recVal)
		if err := handler.ResetPassword(cVal); err == nil {
			t.Error("expected validation error")
		}

		// Invalid token error
		reqPayloadInvalid, _ := json.Marshal(ResetPasswordRequest{Token: "invalid-token", NewPassword: "NewPassword123!"})
		reqInv := httptest.NewRequest(http.MethodPost, "/auth/reset-password", bytes.NewBuffer(reqPayloadInvalid))
		reqInv.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		recInv := httptest.NewRecorder()
		cInv := e.NewContext(reqInv, recInv)
		if err := handler.ResetPassword(cInv); err == nil {
			t.Error("expected error for invalid reset token")
		}

		// Success
		_ = aRepo.StorePasswordResetToken(context.Background(), "valid-token", "user-123")
		reqPayload, _ := json.Marshal(ResetPasswordRequest{Token: "valid-token", NewPassword: "NewPassword123!"})
		req := httptest.NewRequest(http.MethodPost, "/auth/reset-password", bytes.NewBuffer(reqPayload))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.ResetPassword(c)
		if err != nil {
			t.Fatalf("ResetPassword failed: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("RefreshToken", func(t *testing.T) {
		// Missing cookie
		reqNoCookie := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
		recNoCookie := httptest.NewRecorder()
		cNoCookie := e.NewContext(reqNoCookie, recNoCookie)
		if err := handler.RefreshToken(cNoCookie); err == nil {
			t.Error("expected error for missing refresh token cookie")
		}

		// Invalid refresh token
		reqBadCookie := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
		reqBadCookie.AddCookie(&http.Cookie{Name: "refresh_token", Value: "invalid-token"})
		recBadCookie := httptest.NewRecorder()
		cBadCookie := e.NewContext(reqBadCookie, recBadCookie)
		if err := handler.RefreshToken(cBadCookie); err == nil {
			t.Error("expected error for invalid refresh token")
		}

		// Success
		refreshTok, _ := jwtMgr.GenerateRefreshToken("user-123", "alice@example.com")
		_ = aRepo.StoreRefreshToken(context.Background(), refreshTok, "user-123")
		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: refreshTok})
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.RefreshToken(c)
		if err != nil {
			t.Fatalf("RefreshToken failed: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("Logout", func(t *testing.T) {
		refreshTok, _ := jwtMgr.GenerateRefreshToken("user-123", "alice@example.com")
		_ = aRepo.StoreRefreshToken(context.Background(), refreshTok, "user-123")

		req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: refreshTok})
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.Logout(c)
		if err != nil {
			t.Fatalf("Logout failed: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})
}
