// Package scheduler provides a thin wrapper around robfig/cron for running
// periodic background jobs. It supports standard cron expressions and the
// "@every <duration>" descriptor (e.g. "@every 10m").
//
// The Scheduler is safe to start once and stop once. Stop waits for any
// currently running job to finish before returning.
package scheduler

import (
	"fmt"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"

	"github.com/ilse31/base-repo-be-go/pkg/logger"
)

// Job is a unit of periodic work executed by the Scheduler.
type Job func()

// Scheduler wraps *cron.Cron to provide a small, opinionated API and
// consistent logging.
type Scheduler struct {
	cron *cron.Cron
}

// New creates a Scheduler that does not start until Start is called.
func New() *Scheduler {
	return &Scheduler{
		cron: cron.New(
			cron.WithChain(
				cron.Recover(cronLogger{}),
			),
		),
	}
}

// AddJob registers a job to run on the given cron schedule (e.g. "@every 10m"
// or "*/10 * * * *"). Must be called before Start.
func (s *Scheduler) AddJob(name, schedule string, job Job) error {
	_, err := s.cron.AddFunc(schedule, func() {
		logger.Info("scheduler: job started", zap.String("job", name))
		job()
		logger.Info("scheduler: job finished", zap.String("job", name))
	})
	if err != nil {
		return fmt.Errorf("scheduler: invalid schedule %q for job %q: %w", schedule, name, err)
	}
	return nil
}

// Start begins executing scheduled jobs. It is safe to call once.
func (s *Scheduler) Start() {
	s.cron.Start()
	logger.Info("scheduler: started")
}

// Stop halts the scheduler and waits for any running job to complete before
// returning. Call this during graceful shutdown.
func (s *Scheduler) Stop() {
	logger.Info("scheduler: stopping, waiting for running jobs...")
	stopCtx := s.cron.Stop()
	<-stopCtx.Done()
	logger.Info("scheduler: stopped")
}

// cronLogger adapts robfig/cron's Logger interface to the app's zap logger,
// so panics recovered inside jobs are logged instead of swallowed.
type cronLogger struct{}

func (cronLogger) Info(msg string, keysAndValues ...interface{}) {
	logger.Info("scheduler: "+msg, zap.Any("kv", keysAndValues))
}

func (cronLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	logger.Error("scheduler: "+msg, append([]zap.Field{zap.Error(err)}, zap.Any("kv", keysAndValues))...)
}
