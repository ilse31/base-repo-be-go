// Package observability wires up the application's cross-cutting telemetry
// concerns: distributed tracing via OpenTelemetry and error tracking via
// Sentry. Each is opt-in through configuration; when disabled, the
// corresponding setup function is a no-op and the app uses the global
// no-op providers.
//
// Tracing is exported over OTLP/HTTP (the vendor-neutral standard) so the
// same instrumentation ships spans to any compliant backend: Jaeger,
// Grafana Tempo, Honeycomb, Datadog, etc. Pick the backend at deploy time,
// not in code.
package observability

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.uber.org/zap"

	"github.com/ilse31/base-repo-be-go/pkg/config"
	"github.com/ilse31/base-repo-be-go/pkg/logger"
)

// ShutdownFunc flushes pending telemetry and releases exporter resources. It
// must be called during graceful shutdown. It is safe to call when tracing is
// disabled (a no-op).
type ShutdownFunc func(ctx context.Context) error

// noopShutdown does nothing; returned when tracing is disabled.
func noopShutdown(context.Context) error { return nil }

// SetupTracing builds and registers a global TracerProvider that exports
// spans over OTLP/HTTP. When tracing is disabled it returns a no-op
// shutdown without touching any globals.
func SetupTracing(cfg config.ObservabilityConfig) (ShutdownFunc, error) {
	if !cfg.TracingEnabled {
		logger.Info("tracing disabled, using no-op tracer provider")
		return noopShutdown, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(cfg.OTLPEndpoint),
		otlptracehttp.WithInsecure(), // local collector; TLS terminated at the collector
	)
	if err != nil {
		return nil, fmt.Errorf("create OTLP trace exporter: %w", err)
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes("",
			semconv.ServiceName(cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("build trace resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.TracesSampleRate)),
	)

	// Register globally so instrumentation libs (otelecho, bunotel, redisotel)
	// pick up the provider automatically.
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	logger.Info("tracing enabled",
		zap.String("endpoint", cfg.OTLPEndpoint),
		zap.String("service", cfg.ServiceName),
	)

	return tp.Shutdown, nil
}
