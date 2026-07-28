package database

import (
	"context"
	"testing"
	"time"

	"github.com/ilse31/base-repo-be-go/pkg/config"
	"github.com/ilse31/base-repo-be-go/pkg/logger"
)

func TestNewPostgresDB(t *testing.T) {
	_ = logger.Init("development")

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "postgres",
			Password: "postgres",
			DBName:   "mydb",
			SSLMode:  "disable",
		},
		Server: config.ServerConfig{
			Env: "development",
		},
		Observability: config.ObservabilityConfig{
			TracingEnabled: true,
		},
	}

	db, err := NewPostgresDB(cfg)
	if err != nil {
		t.Fatalf("NewPostgresDB failed: %v", err)
	}
	if db == nil {
		t.Fatal("expected non-nil DB")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_ = db.Ping(ctx)
	_ = db.Close()
}
