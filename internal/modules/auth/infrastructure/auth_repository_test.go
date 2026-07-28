package infrastructure

import (
	"context"
	"testing"
	"time"

	infraredis "github.com/ilse31/base-repo-be-go/internal/infrastructure/redis"
	goredis "github.com/redis/go-redis/v9"
)

func TestGenerateResetToken(t *testing.T) {
	tok1 := GenerateResetToken()
	tok2 := GenerateResetToken()

	if tok1 == "" || tok2 == "" {
		t.Error("expected non-empty tokens")
	}
	if tok1 == tok2 {
		t.Error("expected unique tokens")
	}
}

func TestAuthRepository_Methods(t *testing.T) {
	rdb := goredis.NewClient(&goredis.Options{Addr: "127.0.0.1:1"})
	redisClient := &infraredis.Client{Client: rdb}
	repo := NewAuthRepository(redisClient)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_ = repo.StorePasswordResetToken(ctx, "reset-tok", "user-1")
	_, _ = repo.GetPasswordResetToken(ctx, "reset-tok")
	_ = repo.DeletePasswordResetToken(ctx, "reset-tok")

	_ = repo.StoreRefreshToken(ctx, "ref-tok", "user-1")
	_, _ = repo.GetRefreshToken(ctx, "ref-tok")
	_ = repo.DeleteRefreshToken(ctx, "ref-tok")

	_ = rdb.Close()
}
