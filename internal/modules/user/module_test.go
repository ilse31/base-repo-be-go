package user

import (
	"testing"

	"github.com/labstack/echo/v4"

	userdomain "github.com/ilse31/base-repo-be-go/internal/modules/user/domain"
	"github.com/ilse31/base-repo-be-go/internal/shared/validation"
)

type dummyUserRepo struct {
	userdomain.UserRepository
}

func TestUserModule(t *testing.T) {
	v := validation.New()
	mod := New(&dummyUserRepo{}, v)
	if mod == nil {
		t.Fatal("expected non-nil user module")
	}

	e := echo.New()
	g := e.Group("/users")
	dummyMw := func(next echo.HandlerFunc) echo.HandlerFunc {
		return next
	}

	mod.RegisterRoutes(g, dummyMw)
}
