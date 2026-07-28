package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/ilse31/base-repo-be-go/internal/infrastructure/redis"
	authdomain "github.com/ilse31/base-repo-be-go/internal/modules/auth/domain"
)

// authRepository implements authdomain.AuthRepository using the shared Redis client.
type authRepository struct {
	client *redis.Client
}

func NewAuthRepository(client *redis.Client) authdomain.AuthRepository {
	return &authRepository{client: client}
}

func (r *authRepository) StorePasswordResetToken(ctx context.Context, token string, userID string) error {
	key := fmt.Sprintf("password_reset:%s", token)
	return r.client.Set(ctx, key, userID, 1*time.Hour)
}

func (r *authRepository) GetPasswordResetToken(ctx context.Context, token string) (string, error) {
	key := fmt.Sprintf("password_reset:%s", token)
	userID, err := r.client.Get(ctx, key)
	if err != nil {
		return "", err
	}
	return userID, nil
}

func (r *authRepository) DeletePasswordResetToken(ctx context.Context, token string) error {
	key := fmt.Sprintf("password_reset:%s", token)
	return r.client.Delete(ctx, key)
}

func (r *authRepository) StoreRefreshToken(ctx context.Context, token string, userID string) error {
	key := fmt.Sprintf("refresh_token:%s", token)
	return r.client.Set(ctx, key, userID, 7*24*time.Hour) // 7 days
}

func (r *authRepository) GetRefreshToken(ctx context.Context, token string) (string, error) {
	key := fmt.Sprintf("refresh_token:%s", token)
	userID, err := r.client.Get(ctx, key)
	if err != nil {
		return "", err
	}
	return userID, nil
}

func (r *authRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	key := fmt.Sprintf("refresh_token:%s", token)
	return r.client.Delete(ctx, key)
}

// GenerateResetToken generates a new reset token.
func GenerateResetToken() string {
	return uuid.New().String()
}
