package auth

import (
	"testing"

	"github.com/labstack/echo/v4"

	authdomain "github.com/ilse31/base-repo-be-go/internal/modules/auth/domain"
	userdomain "github.com/ilse31/base-repo-be-go/internal/modules/user/domain"
	"github.com/ilse31/base-repo-be-go/internal/shared/validation"
	"github.com/ilse31/base-repo-be-go/pkg/jwt"
	"github.com/ilse31/base-repo-be-go/pkg/mailer"
)

type dummyUserRepo struct {
	userdomain.UserRepository
}

type dummyAuthRepo struct {
	authdomain.AuthRepository
}

func TestAuthModule(t *testing.T) {
	v := validation.New()
	jwtMgr := jwt.NewJWTManager("secret", 1, 168)
	m := mailer.NewNoop()

	mod := New(&dummyUserRepo{}, &dummyAuthRepo{}, jwtMgr, m, "http://localhost/reset", v, false)
	if mod == nil {
		t.Fatal("expected non-nil auth module")
	}

	e := echo.New()
	g := e.Group("/auth")
	mod.RegisterRoutes(g)
}
