package tracing

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

// SpanFromContext extracts the current span from the context
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// TraceIDFromContext extracts the trace ID from the current context
func TraceIDFromContext(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

// SpanIDFromContext extracts the span ID from the current context
func SpanIDFromContext(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().SpanID().String()
	}
	return ""
}

// StartSpan starts a new span with the given name and returns the span and a new context
func StartSpan(ctx context.Context, tracer trace.Tracer, spanName string) (context.Context, trace.Span) {
	return tracer.Start(ctx, spanName)
}

// AddAttributesToCurrentSpan adds attributes to the current span in the context
func AddAttributesToCurrentSpan(ctx context.Context, attrs map[string]interface{}) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		AddBusinessAttributes(span, attrs)
	}
}
