package middleware

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	authapp "github.com/ilse31/base-repo-be-go/internal/modules/auth/application"
	authdomain "github.com/ilse31/base-repo-be-go/internal/modules/auth/domain"
	userapp "github.com/ilse31/base-repo-be-go/internal/modules/user/application"
	userdomain "github.com/ilse31/base-repo-be-go/internal/modules/user/domain"
	"github.com/ilse31/base-repo-be-go/internal/shared/apperrors"
	"github.com/ilse31/base-repo-be-go/internal/shared/domain/valueobjects"
	"github.com/ilse31/base-repo-be-go/internal/shared/response"
	"github.com/ilse31/base-repo-be-go/pkg/jwt"
	"github.com/ilse31/base-repo-be-go/pkg/logger"
	"github.com/ilse31/base-repo-be-go/pkg/observability"
)

// HTTPErrorHandler is Echo's central error handler. It maps domain and app
// errors to structured responses, logs unexpected errors, and reports genuine
// (non-client) failures to Sentry so they surface in error tracking.
func HTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		_ = response.Error(c, appErr)
		return
	}

	var echoErr *echo.HTTPError
	if errors.As(err, &echoErr) {
		_ = response.Error(c, mapEchoError(echoErr))
		return
	}

	if resolved := ResolveDomainError(err); resolved != nil {
		_ = response.Error(c, resolved)
		return
	}

	// Genuine unhandled error: log and report to Sentry.
	logger.Error("unhandled error",
		zap.Error(err),
		zap.String("path", c.Path()),
		zap.String("method", c.Request().Method),
	)
	observability.CaptureException(err)
	_ = response.Error(c, apperrors.Internal("internal server error"))
}

func mapEchoError(e *echo.HTTPError) *apperrors.AppError {
	switch e.Code {
	case http.StatusNotFound:
		return apperrors.New(e.Code, apperrors.CodeNotFound, "resource not found")
	case http.StatusMethodNotAllowed:
		return apperrors.New(e.Code, apperrors.CodeMethodNotAllowed, "method not allowed")
	default:
		return apperrors.New(e.Code, apperrors.CodeBadRequest, http.StatusText(e.Code))
	}
}

func ResolveDomainError(err error) *apperrors.AppError {
	switch {
	case errors.Is(err, userapp.ErrUserNotFound), errors.Is(err, userdomain.ErrUserNotFound):
		return apperrors.New(http.StatusNotFound, apperrors.CodeUserNotFound, "user not found")
	case errors.Is(err, userapp.ErrUserExists), errors.Is(err, authapp.ErrUserAlreadyExists):
		return apperrors.New(http.StatusConflict, apperrors.CodeUserAlreadyExists, "user already exists")
	case errors.Is(err, authapp.ErrInvalidCredentials),
		errors.Is(err, authdomain.ErrInvalidCredentials),
		errors.Is(err, userdomain.ErrInvalidCredentials):
		return apperrors.New(http.StatusUnauthorized, apperrors.CodeInvalidCredentials, "invalid credentials")
	case errors.Is(err, jwt.ErrExpiredToken):
		return apperrors.New(http.StatusUnauthorized, apperrors.CodeTokenExpired, "token expired")
	case errors.Is(err, authapp.ErrInvalidToken), errors.Is(err, jwt.ErrInvalidToken):
		return apperrors.New(http.StatusUnauthorized, apperrors.CodeInvalidToken, "invalid or expired token")
	case errors.Is(err, valueobjects.ErrInvalidEmail),
		errors.Is(err, valueobjects.ErrInvalidPassword),
		errors.Is(err, valueobjects.ErrPasswordTooShort),
		errors.Is(err, valueobjects.ErrInvalidUserID):
		return apperrors.BadRequest(err.Error())
	}
	return nil
}
