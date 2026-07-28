package scheduler

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/ilse31/base-repo-be-go/pkg/logger"
)

func TestScheduler(t *testing.T) {
	_ = logger.Init("development")

	s := New()
	if s == nil {
		t.Fatal("expected non-nil scheduler")
	}

	var wg sync.WaitGroup
	wg.Add(1)

	executed := false
	err := s.AddJob("test-job", "@every 100ms", func() {
		executed = true
		wg.Done()
	})
	if err != nil {
		t.Fatalf("AddJob failed: %v", err)
	}

	s.Start()

	// Wait for job to execute
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		if !executed {
			t.Error("expected job to be executed")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for scheduled job")
	}

	s.Stop()
}

func TestScheduler_InvalidSchedule(t *testing.T) {
	_ = logger.Init("development")
	s := New()

	err := s.AddJob("invalid-job", "invalid-cron-expr", func() {})
	if err == nil {
		t.Error("expected error for invalid schedule, got nil")
	}
}

func TestCronLogger(t *testing.T) {
	_ = logger.Init("development")
	cl := cronLogger{}
	cl.Info("test cron info", "key", "val")
	cl.Error(errors.New("cron err"), "test cron error", "key", "val")
}
