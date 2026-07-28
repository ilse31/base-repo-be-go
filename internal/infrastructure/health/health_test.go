package health

import (
	"context"
	"errors"
	"testing"

	"github.com/ilse31/base-repo-be-go/pkg/logger"
)

type mockPinger struct {
	err error
}

func (m *mockPinger) Ping(ctx context.Context) error {
	return m.err
}

func TestHealthChecker(t *testing.T) {
	_ = logger.Init("development")

	checker := NewChecker()
	if checker == nil {
		t.Fatal("expected non-nil checker")
	}

	checker.Register("db_ok", &mockPinger{err: nil})
	checker.Register("redis_fail", &mockPinger{err: errors.New("connection failed")})

	// Check pings both registered pingers without panicking
	checker.Check()
}
