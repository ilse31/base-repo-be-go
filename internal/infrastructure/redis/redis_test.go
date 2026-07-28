package redis

import (
	"context"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/ilse31/base-repo-be-go/pkg/config"
	"github.com/ilse31/base-repo-be-go/pkg/logger"
)

func TestNewClient_ConnectionError(t *testing.T) {
	_ = logger.Init("development")

	cfg := &config.Config{
		Redis: config.RedisConfig{
			Host: "127.0.0.1",
			Port: "1", // invalid/unreachable port
		},
		Observability: config.ObservabilityConfig{
			TracingEnabled: false,
		},
	}

	_, err := NewClient(cfg)
	if err == nil {
		t.Error("expected connection error for invalid redis port, got nil")
	}
}

func TestNewClient_TracingEnabled(t *testing.T) {
	_ = logger.Init("development")

	cfg := &config.Config{
		Redis: config.RedisConfig{
			Host: "127.0.0.1",
			Port: "1",
		},
		Observability: config.ObservabilityConfig{
			TracingEnabled: true,
		},
	}

	_, err := NewClient(cfg)
	if err == nil {
		t.Error("expected connection error for invalid redis port with tracing, got nil")
	}
}

func TestClient_Methods(t *testing.T) {
	rdb := goredis.NewClient(&goredis.Options{Addr: "127.0.0.1:1"})
	client := &Client{rdb}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_ = client.Ping(ctx)
	_ = client.Set(ctx, "key", "val", time.Second)
	_, _ = client.Get(ctx, "key")
	_ = client.Delete(ctx, "key")
	_, _ = client.Exists(ctx, "key")
	_ = client.Close()
}
