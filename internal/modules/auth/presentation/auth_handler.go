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
	authService  *application.AuthService
	validator    *validator.Validate
	isProduction bool
}

func NewAuthHandler(authService *application.AuthService, v *validator.Validate, isProduction bool) *AuthHandler {
	return &AuthHandler{
		authService:  authService,
		validator:    v,
		isProduction: isProduction,
	}
}

type RegisterRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,password_complexity"`
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
	NewPassword string `json:"new_password" validate:"required,password_complexity"`
}

// Register godoc
// @Summary Register a new user
// @Description Creates a new user account with email, name, and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Register Payload"
// @Success 201 {object} response.Body "user registered successfully"
// @Failure 400 {object} apperrors.AppError "validation / bad request error"
// @Failure 409 {object} apperrors.AppError "user already exists"
// @Router /auth/register [post]
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

// Login godoc
// @Summary Login user
// @Description Authenticates user and returns access/refresh tokens in JSON and HTTP-only cookies
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login Payload"
// @Success 200 {object} response.Body "login successful"
// @Failure 400 {object} apperrors.AppError "invalid request body"
// @Failure 401 {object} apperrors.AppError "invalid credentials"
// @Router /auth/login [post]
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

	h.setAuthCookies(c, accessToken, refreshToken)

	return response.OKWithMessage(c, "login successful", map[string]any{
		"user":          user,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// Logout godoc
// @Summary Logout user
// @Description Clears access and refresh token cookies and invalidates refresh token in Redis
// @Tags Auth
// @Produce json
// @Success 200 {object} response.Body "logged out successfully"
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c echo.Context) error {
	if refreshCookie, err := c.Cookie("refresh_token"); err == nil && refreshCookie.Value != "" {
		_ = h.authService.Logout(c.Request().Context(), refreshCookie.Value)
	}

	h.clearAuthCookies(c)

	return response.Message(c, http.StatusOK, "logged out successfully")
}

// ForgotPassword godoc
// @Summary Forgot password request
// @Description Sends password reset link to user's email address if registered
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body ForgotPasswordRequest true "Forgot Password Payload"
// @Success 200 {object} response.Body "if email exists, reset link sent"
// @Failure 400 {object} apperrors.AppError "invalid request body"
// @Router /auth/forgot-password [post]
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

// ResetPassword godoc
// @Summary Reset password
// @Description Resets password using valid token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body ResetPasswordRequest true "Reset Password Payload"
// @Success 200 {object} response.Body "password reset successfully"
// @Failure 400 {object} apperrors.AppError "invalid or expired token"
// @Router /auth/reset-password [post]
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

// Me godoc
// @Summary Get authenticated user info
// @Description Returns claims of current authenticated user from cookie/token
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Body "user info"
// @Failure 401 {object} apperrors.AppError "unauthorized"
// @Router /auth/me [get]
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

// RefreshToken godoc
// @Summary Refresh access token
// @Description Rotates access and refresh tokens using valid refresh token cookie
// @Tags Auth
// @Produce json
// @Success 200 {object} response.Body "new tokens generated"
// @Failure 401 {object} apperrors.AppError "invalid or expired refresh token"
// @Router /auth/refresh [post]
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

	h.setAuthCookies(c, newAccessToken, newRefreshToken)

	return response.OK(c, map[string]string{
		"access_token":  newAccessToken,
		"refresh_token": newRefreshToken,
	})
}

func (h *AuthHandler) setAuthCookies(c echo.Context, accessToken, refreshToken string) {
	c.SetCookie(&http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		HttpOnly: true,
		Secure:   h.isProduction,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		Expires:  time.Now().Add(1 * time.Hour),
	})
	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   h.isProduction,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		Expires:  time.Now().Add(7 * 24 * time.Hour),
	})
}

func (h *AuthHandler) clearAuthCookies(c echo.Context) {
	for _, name := range []string{"access_token", "refresh_token"} {
		c.SetCookie(&http.Cookie{
			Name:     name,
			Value:    "",
			HttpOnly: true,
			Secure:   h.isProduction,
			SameSite: http.SameSiteStrictMode,
			Path:     "/",
			Expires:  time.Now().Add(-1 * time.Hour),
		})
	}
}
