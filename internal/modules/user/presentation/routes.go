package presentation

import "github.com/labstack/echo/v4"

// RegisterRoutes wires the user module's HTTP routes onto the given group.
// The group is expected to be already scoped (e.g. /api/v1/users).
func (h *UserHandler) RegisterRoutes(g *echo.Group, authMiddleware echo.MiddlewareFunc) {
	protected := g.Group("", authMiddleware)
	protected.POST("", h.CreateUser)
	protected.GET("", h.ListUsers)
	protected.GET("/:id", h.GetUser)
	protected.PUT("/:id", h.UpdateUser)
	protected.DELETE("/:id", h.DeleteUser)
}
