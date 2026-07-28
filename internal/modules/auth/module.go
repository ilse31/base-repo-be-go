// Package auth wires together the authentication bounded context: it depends
// on the user repository for credential lookup, a token store (auth repo) for
// sessions, and a JWT manager for token signing. Module is the single entry
// point used by the composition root (container).
package auth

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"

	"github.com/ilse31/base-repo-be-go/internal/modules/auth/application"
	"github.com/ilse31/base-repo-be-go/internal/modules/auth/domain"
	"github.com/ilse31/base-repo-be-go/internal/modules/auth/presentation"
	userdomain "github.com/ilse31/base-repo-be-go/internal/modules/user/domain"
	"github.com/ilse31/base-repo-be-go/pkg/jwt"
	"github.com/ilse31/base-repo-be-go/pkg/mailer"
)

// Module bundles the auth bounded context. The composition root only needs
// Module.RegisterRoutes to expose its HTTP API.
type Module struct {
	handler *presentation.AuthHandler
}

func New(
	userRepo userdomain.UserRepository,
	authRepo domain.AuthRepository,
	jwtManager *jwt.JWTManager,
	m mailer.Mailer,
	resetURL string,
	v *validator.Validate,
	isProduction bool,
) *Module {
	service := application.NewAuthService(userRepo, authRepo, jwtManager, m, resetURL)
	handler := presentation.NewAuthHandler(service, v, isProduction)
	return &Module{handler: handler}
}

// RegisterRoutes mounts the module's routes onto the given echo group.
func (m *Module) RegisterRoutes(g *echo.Group) {
	m.handler.RegisterRoutes(g)
}
