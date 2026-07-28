package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	authapp "github.com/ilse31/base-repo-be-go/internal/modules/auth/application"
	authdomain "github.com/ilse31/base-repo-be-go/internal/modules/auth/domain"
	userapp "github.com/ilse31/base-repo-be-go/internal/modules/user/application"
	userdomain "github.com/ilse31/base-repo-be-go/internal/modules/user/domain"
	"github.com/ilse31/base-repo-be-go/internal/shared/apperrors"
	"github.com/ilse31/base-repo-be-go/internal/shared/domain/valueobjects"
	"github.com/ilse31/base-repo-be-go/pkg/jwt"
	"github.com/ilse31/base-repo-be-go/pkg/logger"
)

func TestJWTAuthMiddleware(t *testing.T) {
	_ = logger.Init("development")
	jwtMgr := jwt.NewJWTManager("secret", 1, 168)
	validToken, _ := jwtMgr.GenerateAccessToken("user-123", "user@example.com")

	e := echo.New()
	mw := JWTAuth(jwtMgr)

	handler := mw(func(c echo.Context) error {
		userID := c.Get(ContextUserIDKey)
		email := c.Get(ContextEmailKey)
		return c.String(http.StatusOK, userID.(string)+":"+email.(string))
	})

	t.Run("Missing Token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler(c)
		if err == nil {
			t.Fatal("expected unauthorized error, got nil")
		}
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			if appErr.HTTPStatus() != http.StatusUnauthorized {
				t.Errorf("expected status 401, got %d", appErr.HTTPStatus())
			}
		}
	})

	t.Run("Valid Token via Cookie", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{Name: "access_token", Value: validToken})
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler(c)
		if err != nil {
			t.Fatalf("expected success, got %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
		if rec.Body.String() != "user-123:user@example.com" {
			t.Errorf("unexpected body: %s", rec.Body.String())
		}
	})

	t.Run("Valid Token via Authorization Header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler(c)
		if err != nil {
			t.Fatalf("expected success, got %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("Invalid Token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler(c)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestHTTPErrorHandler(t *testing.T) {
	_ = logger.Init("development")
	e := echo.New()

	t.Run("Committed Response", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Response().Committed = true

		HTTPErrorHandler(errors.New("some error"), c)
		// Should do nothing if committed
	})

	t.Run("AppError", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		appErr := apperrors.NotFound("item not found")
		HTTPErrorHandler(appErr, c)
		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("Echo HTTPError NotFound and MethodNotAllowed and default", func(t *testing.T) {
		rec1 := httptest.NewRecorder()
		c1 := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), rec1)
		HTTPErrorHandler(echo.ErrNotFound, c1)
		if rec1.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec1.Code)
		}

		rec2 := httptest.NewRecorder()
		c2 := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), rec2)
		HTTPErrorHandler(echo.ErrMethodNotAllowed, c2)
		if rec2.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected 405, got %d", rec2.Code)
		}

		rec3 := httptest.NewRecorder()
		c3 := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), rec3)
		HTTPErrorHandler(echo.NewHTTPError(http.StatusBadRequest, "bad request"), c3)
		if rec3.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec3.Code)
		}
	})

	t.Run("Domain Errors Resolution", func(t *testing.T) {
		domainErrs := []struct {
			err        error
			wantStatus int
			wantCode   string
		}{
			{userapp.ErrUserNotFound, http.StatusNotFound, apperrors.CodeUserNotFound},
			{userdomain.ErrUserNotFound, http.StatusNotFound, apperrors.CodeUserNotFound},
			{userapp.ErrUserExists, http.StatusConflict, apperrors.CodeUserAlreadyExists},
			{authapp.ErrUserAlreadyExists, http.StatusConflict, apperrors.CodeUserAlreadyExists},
			{authapp.ErrInvalidCredentials, http.StatusUnauthorized, apperrors.CodeInvalidCredentials},
			{authdomain.ErrInvalidCredentials, http.StatusUnauthorized, apperrors.CodeInvalidCredentials},
			{userdomain.ErrInvalidCredentials, http.StatusUnauthorized, apperrors.CodeInvalidCredentials},
			{jwt.ErrExpiredToken, http.StatusUnauthorized, apperrors.CodeTokenExpired},
			{authapp.ErrInvalidToken, http.StatusUnauthorized, apperrors.CodeInvalidToken},
			{jwt.ErrInvalidToken, http.StatusUnauthorized, apperrors.CodeInvalidToken},
			{valueobjects.ErrInvalidEmail, http.StatusBadRequest, apperrors.CodeBadRequest},
			{valueobjects.ErrInvalidPassword, http.StatusBadRequest, apperrors.CodeBadRequest},
			{valueobjects.ErrPasswordTooShort, http.StatusBadRequest, apperrors.CodeBadRequest},
			{valueobjects.ErrInvalidUserID, http.StatusBadRequest, apperrors.CodeBadRequest},
		}

		for _, item := range domainErrs {
			rec := httptest.NewRecorder()
			c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), rec)

			HTTPErrorHandler(item.err, c)
			if rec.Code != item.wantStatus {
				t.Errorf("for error %v, expected status %d, got %d", item.err, item.wantStatus, rec.Code)
			}
		}
	})

	t.Run("Unhandled Error", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), rec)

		HTTPErrorHandler(errors.New("unhandled internal error"), c)
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", rec.Code)
		}
	})

	t.Run("ResolveDomainError fallback nil", func(t *testing.T) {
		res := ResolveDomainError(errors.New("random unknown error"))
		if res != nil {
			t.Errorf("expected nil for unknown error, got %v", res)
		}
	})
}
