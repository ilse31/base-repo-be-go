package presentation

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"

	"github.com/ilse31/base-repo-be-go/internal/modules/auth/application"
	"github.com/ilse31/base-repo-be-go/internal/shared/apperrors"
	"github.com/ilse31/base-repo-be-go/internal/shared/response"
	"github.com/ilse31/base-repo-be-go/internal/shared/validation"
)

type AuthHandler struct {
	authService *application.AuthService
	validator   *validator.Validate
}

func NewAuthHandler(authService *application.AuthService, v *validator.Validate) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validator:   v,
	}
}

type RegisterRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=6"`
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return apperrors.BadRequest("invalid request body")
	}
	if err := h.validator.Struct(req); err != nil {
		return validation.ToAppError(err)
	}

	user, err := h.authService.Register(c.Request().Context(), req.Name, req.Email, req.Password)
	if err != nil {
		return err
	}

	return response.Created(c, "user registered successfully", user)
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return apperrors.BadRequest("invalid request body")
	}
	if err := h.validator.Struct(req); err != nil {
		return validation.ToAppError(err)
	}

	accessToken, refreshToken, user, err := h.authService.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return err
	}

	setAuthCookies(c, accessToken, refreshToken)

	return response.OKWithMessage(c, "login successful", map[string]any{
		"user":          user,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (h *AuthHandler) Logout(c echo.Context) error {
	if refreshCookie, err := c.Cookie("refresh_token"); err == nil && refreshCookie.Value != "" {
		_ = h.authService.Logout(c.Request().Context(), refreshCookie.Value)
	}

	clearAuthCookies(c)

	return response.Message(c, http.StatusOK, "logged out successfully")
}

func (h *AuthHandler) ForgotPassword(c echo.Context) error {
	var req ForgotPasswordRequest
	if err := c.Bind(&req); err != nil {
		return apperrors.BadRequest("invalid request body")
	}
	if err := h.validator.Struct(req); err != nil {
		return validation.ToAppError(err)
	}

	// Always respond identically to prevent email enumeration.
	_ = h.authService.ForgotPassword(c.Request().Context(), req.Email)

	return response.OKWithMessage(c, "if the email exists, a password reset link has been sent", nil)
}

func (h *AuthHandler) ResetPassword(c echo.Context) error {
	var req ResetPasswordRequest
	if err := c.Bind(&req); err != nil {
		return apperrors.BadRequest("invalid request body")
	}
	if err := h.validator.Struct(req); err != nil {
		return validation.ToAppError(err)
	}

	if err := h.authService.ResetPassword(c.Request().Context(), req.Token, req.NewPassword); err != nil {
		if errors.Is(err, application.ErrInvalidToken) {
			return apperrors.BadRequest("invalid or expired token")
		}
		return err
	}

	return response.Message(c, http.StatusOK, "password reset successfully")
}

func (h *AuthHandler) Me(c echo.Context) error {
	cookie, err := c.Cookie("access_token")
	if err != nil {
		return apperrors.Unauthorized("not authenticated")
	}

	claims, err := h.authService.ValidateToken(cookie.Value)
	if err != nil {
		return apperrors.Unauthorized("invalid token")
	}

	return response.OK(c, map[string]string{
		"user_id": claims.UserID,
		"email":   claims.Email,
	})
}

func (h *AuthHandler) RefreshToken(c echo.Context) error {
	refreshCookie, err := c.Cookie("refresh_token")
	if err != nil {
		return apperrors.Unauthorized("refresh token not found")
	}

	newAccessToken, newRefreshToken, err := h.authService.RefreshToken(c.Request().Context(), refreshCookie.Value)
	if err != nil {
		if errors.Is(err, application.ErrInvalidToken) {
			return apperrors.Unauthorized("invalid or expired refresh token")
		}
		return err
	}

	setAuthCookies(c, newAccessToken, newRefreshToken)

	return response.OK(c, map[string]string{
		"access_token":  newAccessToken,
		"refresh_token": newRefreshToken,
	})
}

func setAuthCookies(c echo.Context, accessToken, refreshToken string) {
	c.SetCookie(&http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		Expires:  time.Now().Add(1 * time.Hour),
	})
	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		Expires:  time.Now().Add(7 * 24 * time.Hour),
	})
}

func clearAuthCookies(c echo.Context) {
	for _, name := range []string{"access_token", "refresh_token"} {
		c.SetCookie(&http.Cookie{
			Name:     name,
			Value:    "",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			Path:     "/",
			Expires:  time.Now().Add(-1 * time.Hour),
		})
	}
}
