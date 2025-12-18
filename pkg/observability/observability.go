package observability

import (
	"context"
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel/sdk/resource"
)

// Observability provides unified access to logs, traces, and metrics
type Observability struct {
	resource *resource.Resource
	logger   Logger
	tracer   Tracer
	metrics  Metrics
	shutdown func(context.Context) error
}

// Logger returns the logger instance
func (o *Observability) Logger() Logger {
	return o.logger
}

// Tracer returns the tracer instance
func (o *Observability) Tracer() Tracer {
	return o.tracer
}

// Metrics returns the metrics instance
func (o *Observability) Metrics() Metrics {
	return o.metrics
}

// PrometheusHandler returns the HTTP handler for Prometheus metrics scraping
// Returns nil if Prometheus export is not enabled in the config
func (o *Observability) PrometheusHandler() http.Handler {
	if otelMetrics, ok := o.metrics.(*otelMetrics); ok {
		return otelMetrics.PrometheusHandler()
	}
	return nil
}

// Resource returns the OpenTelemetry resource
func (o *Observability) Resource() *resource.Resource {
	return o.resource
}

// Shutdown gracefully shuts down all observability components
func (o *Observability) Shutdown(ctx context.Context) error {
	if o.shutdown != nil {
		return o.shutdown(ctx)
	}
	return nil
}

// HealthCheck checks the health of all observability components
// Returns an error if any component is unhealthy
func (o *Observability) HealthCheck(ctx context.Context) error {
	var errs []error

	// Check resource
	if o.resource == nil {
		errs = append(errs, fmt.Errorf("resource is nil"))
	}

	// Check logger
	if o.logger == nil {
		errs = append(errs, fmt.Errorf("logger is nil"))
	} else {
		// Try to get logger provider to check if it's healthy
		if otelLogger, ok := o.logger.(*otelLogger); ok {
			if otelLogger.loggerProvider == nil {
				errs = append(errs, fmt.Errorf("logger provider is nil"))
			}
		}
	}

	// Check tracer
	if o.tracer == nil {
		errs = append(errs, fmt.Errorf("tracer is nil"))
	} else {
		// Try to get tracer provider to check if it's healthy
		if otelTracer, ok := o.tracer.(*otelTracer); ok {
			if otelTracer.tracerProvider == nil {
				errs = append(errs, fmt.Errorf("tracer provider is nil"))
			}
		}
	}

	// Check metrics
	if o.metrics == nil {
		errs = append(errs, fmt.Errorf("metrics is nil"))
	} else {
		// Try to get meter provider to check if it's healthy
		if otelMetrics, ok := o.metrics.(*otelMetrics); ok {
			if otelMetrics.meterProvider == nil {
				errs = append(errs, fmt.Errorf("meter provider is nil"))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("observability health check failed: %v", errs)
	}

	return nil
}

// Initialize creates and initializes all observability components (logs, traces, metrics)
// This is the main entry point for setting up observability in your application
func Initialize(cfg Config) (*Observability, error) {
	// Step 1: Create OTel resource with service metadata
	res, err := NewResource(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Step 2-3: Initialize all three pillars
	logger, loggerShutdown, err := NewLogger(cfg, res)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	tracer, tracerShutdown, err := NewTracer(cfg, res)
	if err != nil {
		// Try to shutdown logger if tracer fails
		if loggerShutdown != nil {
			_ = loggerShutdown(context.Background())
		}
		return nil, fmt.Errorf("failed to initialize tracer: %w", err)
	}

	metrics, metricsShutdown, err := NewMetrics(cfg, res)
	if err != nil {
		// Try to shutdown logger and tracer if metrics fails
		if loggerShutdown != nil {
			_ = loggerShutdown(context.Background())
		}
		if tracerShutdown != nil {
			_ = tracerShutdown(context.Background())
		}
		return nil, fmt.Errorf("failed to initialize metrics: %w", err)
	}

	// Step 4: Create unified shutdown function
	shutdown := func(ctx context.Context) error {
		var errs []error

		// Shutdown in reverse order: metrics, tracer, logger
		if metricsShutdown != nil {
			if err := metricsShutdown(ctx); err != nil {
				errs = append(errs, fmt.Errorf("metrics shutdown error: %w", err))
			}
		}

		if tracerShutdown != nil {
			if err := tracerShutdown(ctx); err != nil {
				errs = append(errs, fmt.Errorf("tracer shutdown error: %w", err))
			}
		}

		if loggerShutdown != nil {
			if err := loggerShutdown(ctx); err != nil {
				errs = append(errs, fmt.Errorf("logger shutdown error: %w", err))
			}
		}

		if len(errs) > 0 {
			return fmt.Errorf("shutdown errors: %v", errs)
		}

		return nil
	}

	// Step 5: Create unified observability instance
	obs := &Observability{
		resource: res,
		logger:   logger,
		tracer:   tracer,
		metrics:  metrics,
		shutdown: shutdown,
	}

	return obs, nil
}
