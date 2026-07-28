package middleware

import (
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/ilse31/base-repo-be-go/internal/shared/apperrors"
	"github.com/ilse31/base-repo-be-go/pkg/jwt"
)

const (
	ContextUserIDKey = "user_id"
	ContextEmailKey  = "email"
)

// JWTAuth creates an Echo middleware that authenticates requests using JWT tokens
// extracted from either an HTTP cookie ("access_token") or the Authorization header ("Bearer <token>").
func JWTAuth(jwtManager *jwt.JWTManager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var tokenStr string

			// 1. Try extracting token from Cookie
			cookie, err := c.Cookie("access_token")
			if err == nil && cookie.Value != "" {
				tokenStr = cookie.Value
			} else {
				// 2. Fallback to Authorization Header (Bearer token)
				authHeader := c.Request().Header.Get("Authorization")
				if authHeader != "" {
					parts := strings.SplitN(authHeader, " ", 2)
					if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
						tokenStr = parts[1]
					}
				}
			}

			if tokenStr == "" {
				return apperrors.Unauthorized("authentication token required")
			}

			claims, err := jwtManager.ValidateToken(tokenStr)
			if err != nil {
				return apperrors.Unauthorized("invalid or expired token")
			}

			// Store claims in Echo context for handlers
			c.Set(ContextUserIDKey, claims.UserID)
			c.Set(ContextEmailKey, claims.Email)

			return next(c)
		}
	}
}
