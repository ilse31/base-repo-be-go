package middleware

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisRateLimiterStore implements echo middleware.RateLimiterStore backed by
// Redis, so the limit is shared across all instances of the application
// instead of being tracked per-process.
//
// It uses a fixed-window counter: each identifier gets a counter key that
// expires after Window; once the counter exceeds Requests within that
// window, further requests are denied until the window resets.
type RedisRateLimiterStore struct {
	client   *redis.Client
	prefix   string
	requests int64
	window   time.Duration
}

// NewRedisRateLimiterStore creates a store allowing up to requests calls per
// identifier within the given window.
func NewRedisRateLimiterStore(client *redis.Client, requests int, window time.Duration) *RedisRateLimiterStore {
	return &RedisRateLimiterStore{
		client:   client,
		prefix:   "ratelimit:",
		requests: int64(requests),
		window:   window,
	}
}

func (s *RedisRateLimiterStore) Allow(identifier string) (bool, error) {
	ctx := context.Background()
	key := s.prefix + identifier

	count, err := s.client.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}
	if count == 1 {
		if err := s.client.Expire(ctx, key, s.window).Err(); err != nil {
			return false, err
		}
	}

	return count <= s.requests, nil
}
