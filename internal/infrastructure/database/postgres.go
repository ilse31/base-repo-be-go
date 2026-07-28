package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	"github.com/uptrace/bun/extra/bunotel"

	"github.com/ilse31/base-repo-be-go/pkg/config"
)

type DB struct {
	*bun.DB
}

// NewPostgresDB builds a Bun DB connection. When tracing is enabled each query
// emits an OpenTelemetry span via bunotel; in development bundebug logs the
// raw SQL. Both hooks are additive and safe to register together.
func NewPostgresDB(cfg *config.Config) (*DB, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(sqldb, pgdialect.New())

	// OpenTelemetry tracing for every DB query (spans show up in Jaeger/Tempo).
	if cfg.Observability.TracingEnabled {
		db.AddQueryHook(bunotel.NewQueryHook(
			bunotel.WithDBName(cfg.Database.DBName),
		))
	}

	// Verbose SQL logging in development only.
	if cfg.Server.Env == "development" {
		db.AddQueryHook(bundebug.NewQueryHook(
			bundebug.WithVerbose(true),
			bundebug.FromEnv("BUNDEBUG"),
		))
	}

	return &DB{db}, nil
}

func (db *DB) Close() error {
	return db.DB.Close()
}

func (db *DB) Ping(ctx context.Context) error {
	return db.DB.PingContext(ctx)
}
