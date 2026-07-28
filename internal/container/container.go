// Package container is the application's composition root: it assembles all
// bounded-context modules from their infrastructure dependencies and mounts
// their HTTP routes under a versioned API group.
//
// To add a new module:
//  1. Implement its Module (domain/application/presentation + module.go).
//  2. Add its dependencies to Deps.
//  3. Wire it in New and RegisterRoutes.
package container

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"

	authmodule "github.com/ilse31/base-repo-be-go/internal/modules/auth"
	authdomain "github.com/ilse31/base-repo-be-go/internal/modules/auth/domain"
	usermodule "github.com/ilse31/base-repo-be-go/internal/modules/user"
	userdomain "github.com/ilse31/base-repo-be-go/internal/modules/user/domain"
	"github.com/ilse31/base-repo-be-go/pkg/jwt"
	"github.com/ilse31/base-repo-be-go/pkg/mailer"
)

// Deps holds the infrastructure dependencies required to build all modules.
// The composition root (main) is responsible for constructing these.
type Deps struct {
	UserRepo         userdomain.UserRepository
	AuthRepo         authdomain.AuthRepository
	JWTManager       *jwt.JWTManager
	Mailer           mailer.Mailer
	FrontendResetURL string
	Validator        *validator.Validate
}

// Container holds the assembled bounded-context modules.
type Container struct {
	User *usermodule.Module
	Auth *authmodule.Module
}

// New builds every module from the provided dependencies.
func New(deps Deps) *Container {
	return &Container{
		User: usermodule.New(deps.UserRepo, deps.Validator),
		Auth: authmodule.New(deps.UserRepo, deps.AuthRepo, deps.JWTManager, deps.Mailer, deps.FrontendResetURL, deps.Validator),
	}
}

// RegisterRoutes mounts all module routes under /api/v1 on the echo instance.
func (c *Container) RegisterRoutes(e *echo.Echo) {
	v1 := e.Group("/api/v1")
	c.Auth.RegisterRoutes(v1.Group("/auth"))
	c.User.RegisterRoutes(v1.Group("/users"))
}
