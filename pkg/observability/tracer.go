package observability

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// otelTracer wraps OpenTelemetry trace.Tracer with our interface
type otelTracer struct {
	tracer         trace.Tracer
	tracerProvider *sdktrace.TracerProvider
}

// NewTracer creates a new tracer instance with the provided configuration and resource
func NewTracer(cfg Config, res *resource.Resource) (Tracer, func(context.Context) error, error) {
	// Create OTLP HTTP trace exporter
	exporter, err := otlptracehttp.New(
		context.Background(),
		otlptracehttp.WithEndpoint(cfg.OTELExporterEndpoint),
		otlptracehttp.WithInsecure(), // TODO: make this configurable
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
	}

	// Set default sampling rates if not provided
	successRate := cfg.SuccessSamplingRate
	errorRate := cfg.ErrorSamplingRate
	if successRate == 0.0 && errorRate == 0.0 {
		// Default: 3% for success, 100% for errors
		successRate = 0.03
		errorRate = 1.0
	} else if successRate == 0.0 {
		successRate = 0.03 // Default success rate
	} else if errorRate == 0.0 {
		errorRate = 1.0 // Default error rate
	}

	// Create custom sampler for error-aware sampling
	sampler := NewErrorAwareSampler(successRate, errorRate)

	// Create span processor for filtering successful spans
	spanProcessor := NewErrorFilteringSpanProcessor(successRate)

	// Create tracer provider with batch processor, sampler, and span processor
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(cfg.BatchTimeout),
			sdktrace.WithExportTimeout(cfg.ExportTimeout),
			sdktrace.WithMaxExportBatchSize(cfg.MaxExportBatch),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
		sdktrace.WithSpanProcessor(spanProcessor),
	)

	// Create tracer from provider
	apiTracer := tracerProvider.Tracer(
		string(cfg.ServiceName),
		trace.WithSchemaURL("https://opentelemetry.io/schemas/1.27.0"),
	)

	shutdown := func(ctx context.Context) error {
		return tracerProvider.Shutdown(ctx)
	}

	return &otelTracer{
		tracer:         apiTracer,
		tracerProvider: tracerProvider,
	}, shutdown, nil
}

// Start starts a new span and returns the context and span
func (t *otelTracer) Start(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span) {
	spanOpts := make([]trace.SpanStartOption, 0, len(opts))
	for _, opt := range opts {
		if spanOpt, ok := opt.(spanOption); ok {
			spanOpts = append(spanOpts, spanOpt.opt)
		}
	}

	ctx, otelSpan := t.tracer.Start(ctx, name, spanOpts...)
	return ctx, &span{span: otelSpan}
}

// StartSpan starts a new span and returns only the span (context must be managed separately)
func (t *otelTracer) StartSpan(ctx context.Context, name string, opts ...SpanOption) Span {
	spanOpts := make([]trace.SpanStartOption, 0, len(opts))
	for _, opt := range opts {
		if spanOpt, ok := opt.(spanOption); ok {
			spanOpts = append(spanOpts, spanOpt.opt)
		}
	}

	_, otelSpan := t.tracer.Start(ctx, name, spanOpts...)
	return &span{span: otelSpan}
}

// GetTracerProvider returns the underlying tracer provider (for advanced use cases)
func (t *otelTracer) GetTracerProvider() *sdktrace.TracerProvider {
	return t.tracerProvider
}
