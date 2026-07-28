package observability

import (
	"time"

	"github.com/getsentry/sentry-go"

	"github.com/ilse31/base-repo-be-go/pkg/config"
	"github.com/ilse31/base-repo-be-go/pkg/logger"
	"go.uber.org/zap"
)

// SetupSentry initializes the global Sentry client. When disabled (or when no
// DSN is configured) it is a no-op; subsequent CaptureException calls become
// safe no-ops because the SDK falls back to a no-op transport when no DSN is
// set.
func SetupSentry(cfg config.ObservabilityConfig) error {
	if !cfg.SentryEnabled || cfg.SentryDSN == "" {
		logger.Info("Sentry disabled")
		return nil
	}

	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.SentryDSN,
		Environment:      cfg.SentryEnv,
		Release:          cfg.ServiceName,
		TracesSampleRate: cfg.SentrySampleRate,
		// Attach panic stack traces and the current request context (set by the
		// Echo integration via hub scope) to captured events.
		AttachStacktrace: true,
	}); err != nil {
		return err
	}

	logger.Info("Sentry enabled",
		zap.String("environment", cfg.SentryEnv),
	)
	return nil
}

// FlushSentry blocks until pending Sentry events are flushed (up to timeout)
// or the deadline elapses. Call during graceful shutdown.
func FlushSentry(timeout time.Duration) {
	sentry.Flush(timeout)
}

// CaptureException reports an error to Sentry with the current scope. It is
// safe to call when Sentry is disabled (no-op).
func CaptureException(err error) {
	if err == nil {
		return
	}
	sentry.CaptureException(err)
}
