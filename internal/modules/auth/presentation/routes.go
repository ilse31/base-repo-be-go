package presentation

import "github.com/labstack/echo/v4"

// RegisterRoutes wires the auth module's HTTP routes onto the given group.
// The group is expected to be already scoped (e.g. /api/v1/auth).
func (h *AuthHandler) RegisterRoutes(g *echo.Group) {
	g.POST("/register", h.Register)
	g.POST("/login", h.Login)
	g.POST("/logout", h.Logout)
	g.POST("/refresh", h.RefreshToken)
	g.POST("/forgot-password", h.ForgotPassword)
	g.POST("/reset-password", h.ResetPassword)
	g.GET("/me", h.Me)
}
