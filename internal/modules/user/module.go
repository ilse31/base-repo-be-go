// Package user wires together the user bounded context: repository,
// application service and HTTP presentation layer. The Module is the single
// entry point used by the composition root (container) to build and mount the
// module onto the HTTP server.
package user

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"

	"github.com/ilse31/base-repo-be-go/internal/modules/user/application"
	"github.com/ilse31/base-repo-be-go/internal/modules/user/domain"
	"github.com/ilse31/base-repo-be-go/internal/modules/user/presentation"
)

// Module bundles the user bounded context. It is self-contained: the
// composition root only needs Module.RegisterRoutes to expose its HTTP API.
type Module struct {
	handler *presentation.UserHandler
}

func New(userRepo domain.UserRepository, v *validator.Validate) *Module {
	service := application.NewUserService(userRepo)
	handler := presentation.NewUserHandler(service, v)
	return &Module{handler: handler}
}

// RegisterRoutes mounts the module's routes onto the given echo group.
func (m *Module) RegisterRoutes(g *echo.Group, authMw echo.MiddlewareFunc) {
	m.handler.RegisterRoutes(g, authMw)
}
