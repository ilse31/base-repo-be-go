package container

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

func TestContainer(t *testing.T) {
	jwtMgr := jwt.NewJWTManager("secret", 1, 168)
	v := validation.New()
	m := mailer.NewNoop()

	deps := Deps{
		UserRepo:         &dummyUserRepo{},
		AuthRepo:         &dummyAuthRepo{},
		JWTManager:       jwtMgr,
		Mailer:           m,
		FrontendResetURL: "http://localhost/reset",
		Validator:        v,
		IsProduction:     false,
	}

	c := New(deps)
	if c == nil || c.User == nil || c.Auth == nil {
		t.Fatal("expected initialized container with modules")
	}

	e := echo.New()
	c.RegisterRoutes(e)
}
