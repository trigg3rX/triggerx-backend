package observability

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Logger defines the interface for structured logging with context support
type Logger interface {
	Debug(ctx context.Context, msg string, fields ...Field)
	Info(ctx context.Context, msg string, fields ...Field)
	Warn(ctx context.Context, msg string, fields ...Field)
	Error(ctx context.Context, msg string, fields ...Field)
	Fatal(ctx context.Context, msg string, fields ...Field)
	With(fields ...Field) Logger
}

// Tracer defines the interface for creating and managing spans
type Tracer interface {
	Start(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span)
	StartSpan(ctx context.Context, name string, opts ...SpanOption) Span
}

// Span defines the interface for span operations
type Span interface {
	End()
	SetAttributes(attrs ...attribute.KeyValue)
	AddEvent(name string, opts ...EventOption)
	RecordError(err error, opts ...RecordErrorOption)
	SetStatus(code codes.Code, description string)
	SpanContext() trace.SpanContext
}

// SpanOption represents an option for creating a span
type SpanOption interface {
	applySpanOption()
}

// EventOption represents an option for adding an event to a span
type EventOption interface {
	applyEventOption()
}

// RecordErrorOption represents an option for recording an error on a span
type RecordErrorOption interface {
	applyRecordErrorOption()
}

// Metrics defines the interface for creating metric instruments
type Metrics interface {
	Counter(name string, opts ...InstrumentOption) Counter
	Gauge(name string, opts ...InstrumentOption) Gauge
	Histogram(name string, opts ...InstrumentOption) Histogram
}

// Counter defines the interface for counter metrics
type Counter interface {
	Add(ctx context.Context, value float64, attrs ...attribute.KeyValue)
	Inc(ctx context.Context, attrs ...attribute.KeyValue)
}

// Gauge defines the interface for gauge metrics
type Gauge interface {
	Record(ctx context.Context, value float64, attrs ...attribute.KeyValue)
	// Set sets the gauge to a specific value (convenience method for Prometheus compatibility)
	Set(ctx context.Context, value float64, attrs ...attribute.KeyValue)
}

// Histogram defines the interface for histogram metrics
type Histogram interface {
	Record(ctx context.Context, value float64, attrs ...attribute.KeyValue)
}

// InstrumentOption represents an option for creating a metric instrument
type InstrumentOption interface {
	applyInstrumentOption()
}
