package observability

import (
	"errors"
	"testing"
	"time"

	"github.com/ilse31/base-repo-be-go/pkg/config"
	"github.com/ilse31/base-repo-be-go/pkg/logger"
)

func TestSentry(t *testing.T) {
	_ = logger.Init("development")

	// Sentry Disabled
	cfg := config.ObservabilityConfig{
		SentryEnabled: false,
	}
	err := SetupSentry(cfg)
	if err != nil {
		t.Errorf("expected no error when sentry disabled, got %v", err)
	}

	// Sentry Enabled without DSN (no-op fallback)
	cfg.SentryEnabled = true
	cfg.SentryDSN = ""
	err = SetupSentry(cfg)
	if err != nil {
		t.Errorf("expected no error when sentry DSN empty, got %v", err)
	}

	CaptureException(nil)
	CaptureException(errors.New("test error"))
	FlushSentry(100 * time.Millisecond)
}
