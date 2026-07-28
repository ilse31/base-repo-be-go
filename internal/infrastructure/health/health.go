// Package health provides periodic health checks for the application's
// infrastructure dependencies (database, cache). It is invoked by the
// scheduler at a fixed interval and logs the result of each dependency ping.
package health

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/ilse31/base-repo-be-go/pkg/logger"
)

// Pinger is any dependency that can be health-checked via a context-scoped
// ping. Both the Postgres DB and Redis client implement this.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Checker runs health checks against a set of named dependencies.
type Checker struct {
	checks map[string]Pinger
}

// NewChecker creates a Checker from the given named dependencies.
func NewChecker() *Checker {
	return &Checker{checks: make(map[string]Pinger)}
}

// Register adds (or replaces) a dependency to be checked.
func (c *Checker) Register(name string, p Pinger) {
	c.checks[name] = p
}

// Check pings every registered dependency with a per-call timeout and logs
// the result. It never panics: a failing dependency is logged as an error and
// the remaining dependencies are still checked.
func (c *Checker) Check() {
	for name, p := range c.checks {
		c.checkOne(name, p)
	}
}

func (c *Checker) checkOne(name string, p Pinger) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start := time.Now()
	err := p.Ping(ctx)
	elapsed := time.Since(start)

	if err != nil {
		logger.Error("health check failed",
			zap.String("dependency", name),
			zap.Duration("latency", elapsed),
			zap.Error(err),
		)
		return
	}

	logger.Info("health check ok",
		zap.String("dependency", name),
		zap.Duration("latency", elapsed),
	)
}
