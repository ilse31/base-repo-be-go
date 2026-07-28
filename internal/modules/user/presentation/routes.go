package presentation

import "github.com/labstack/echo/v4"

// RegisterRoutes wires the user module's HTTP routes onto the given group.
// The group is expected to be already scoped (e.g. /api/v1/users).
func (h *UserHandler) RegisterRoutes(g *echo.Group) {
	g.POST("", h.CreateUser)
	g.GET("", h.ListUsers)
	g.GET("/:id", h.GetUser)
	g.PUT("/:id", h.UpdateUser)
	g.DELETE("/:id", h.DeleteUser)
}
