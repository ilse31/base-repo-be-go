package database

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"

	userdomain "github.com/ilse31/base-repo-be-go/internal/modules/user/domain"
	"github.com/ilse31/base-repo-be-go/internal/shared/domain"
)

func TestUserRepository_Methods(t *testing.T) {
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN("postgres://postgres:@localhost:5432/mydb?sslmode=disable")))
	bundb := bun.NewDB(sqldb, pgdialect.New())
	defer bundb.Close()

	repo := NewUserRepository(bundb)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	u := &userdomain.User{
		BaseEntity: domain.BaseEntity{ID: ""},
		Name:       "Alice",
		Email:      "alice@example.com",
		Password:   "hash",
	}

	_ = repo.Create(ctx, u)
	_, _ = repo.GetByID(ctx, "some-id")
	_, _ = repo.GetByEmail(ctx, "alice@example.com")
	_ = repo.Update(ctx, u)
	_ = repo.Delete(ctx, "some-id")
	_, _ = repo.List(ctx, 10, 0)
}
