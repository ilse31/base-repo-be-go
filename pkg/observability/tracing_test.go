package observability

import (
	"context"
	"testing"

	"github.com/ilse31/base-repo-be-go/pkg/config"
	"github.com/ilse31/base-repo-be-go/pkg/logger"
)

func TestSetupTracing_Disabled(t *testing.T) {
	_ = logger.Init("development")

	cfg := config.ObservabilityConfig{
		TracingEnabled: false,
	}

	shutdown, err := SetupTracing(cfg)
	if err != nil {
		t.Fatalf("expected no error when tracing disabled, got %v", err)
	}
	if shutdown == nil {
		t.Fatal("expected non-nil shutdown function")
	}

	err = shutdown(context.Background())
	if err != nil {
		t.Errorf("expected no error from noop shutdown, got %v", err)
	}
}

func TestSetupTracing_Enabled(t *testing.T) {
	_ = logger.Init("development")

	cfg := config.ObservabilityConfig{
		TracingEnabled:   true,
		ServiceName:      "test-service",
		OTLPEndpoint:     "localhost:4318",
		TracesSampleRate: 1.0,
	}

	shutdown, err := SetupTracing(cfg)
	if err != nil {
		t.Fatalf("SetupTracing error: %v", err)
	}
	if shutdown == nil {
		t.Fatal("expected non-nil shutdown")
	}

	// Clean up tracing
	_ = shutdown(context.Background())
}
