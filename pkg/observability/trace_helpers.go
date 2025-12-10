package observability

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// TraceContext represents trace context information (for compatibility with old API)
type TraceContext struct {
	TraceID    string            `json:"trace_id"`
	SpanID     string            `json:"span_id"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

// GetTraceContext extracts trace context from the current context
// This replaces pkg/rpc/tracing.GetTraceContext
func GetTraceContext(ctx context.Context) *TraceContext {
	span := trace.SpanFromContext(ctx)
	spanContext := span.SpanContext()

	if !spanContext.HasTraceID() {
		return nil
	}

	attrs := make(map[string]string)
	return &TraceContext{
		TraceID:    spanContext.TraceID().String(),
		SpanID:     spanContext.SpanID().String(),
		Attributes: attrs,
	}
}

// CreateChildSpan creates a child span from the current context
// This replaces pkg/rpc/tracing.CreateChildSpan
func CreateChildSpan(tracer Tracer, ctx context.Context, operationName string, attributes ...attribute.KeyValue) (context.Context, Span) {
	return tracer.Start(ctx, operationName, WithAttributes(attributes...))
}

// AddSpanEvent adds an event to the current span
// This replaces pkg/rpc/tracing.AddSpanEvent
func AddSpanEvent(ctx context.Context, name string, attributes ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.AddEvent(name, trace.WithAttributes(attributes...))
	}
}

// SetSpanAttributes sets attributes on the current span
// This replaces pkg/rpc/tracing.SetSpanAttributes
func SetSpanAttributes(ctx context.Context, attributes ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(attributes...)
	}
}

// RecordSpanError records an error on the current span
// This replaces pkg/rpc/tracing.RecordSpanError
func RecordSpanError(ctx context.Context, err error, attributes ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.RecordError(err, trace.WithAttributes(attributes...))
		span.SetStatus(codes.Error, err.Error())
	}
}

// GetTraceIDFromContext extracts trace ID from context
// This replaces pkg/rpc/tracing.GetTraceIDFromContext
func GetTraceIDFromContext(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	spanContext := span.SpanContext()
	if spanContext.HasTraceID() {
		return spanContext.TraceID().String()
	}
	return ""
}

// GetSpanIDFromContext extracts span ID from context
// This replaces pkg/rpc/tracing.GetSpanIDFromContext
func GetSpanIDFromContext(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	spanContext := span.SpanContext()
	if spanContext.HasSpanID() {
		return spanContext.SpanID().String()
	}
	return ""
}

// IsSampled checks if the current span is sampled
// This replaces pkg/rpc/tracing.IsSampled
func IsSampled(ctx context.Context) bool {
	span := trace.SpanFromContext(ctx)
	spanContext := span.SpanContext()
	return spanContext.TraceFlags().IsSampled()
}

// FormatTraceContext formats trace context for logging
// This replaces pkg/rpc/tracing.FormatTraceContext
func FormatTraceContext(ctx context.Context) string {
	traceID := GetTraceIDFromContext(ctx)
	spanID := GetSpanIDFromContext(ctx)

	if traceID == "" {
		return "no-trace"
	}

	if spanID == "" {
		return fmt.Sprintf("trace_id=%s", traceID)
	}

	return fmt.Sprintf("trace_id=%s,span_id=%s", traceID, spanID)
}
