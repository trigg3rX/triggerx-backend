package tracing

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"
)

// TraceContext represents trace context information
type TraceContext struct {
	TraceID    string            `json:"trace_id"`
	SpanID     string            `json:"span_id"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

// GetTraceContext extracts trace context from the current context
func GetTraceContext(ctx context.Context) *TraceContext {
	span := trace.SpanFromContext(ctx)
	spanContext := span.SpanContext()

	if !spanContext.HasTraceID() {
		return nil
	}

	attrs := make(map[string]string)
	span.SetAttributes(
		attribute.String("trace.id", spanContext.TraceID().String()),
		attribute.String("span.id", spanContext.SpanID().String()),
	)

	return &TraceContext{
		TraceID:    spanContext.TraceID().String(),
		SpanID:     spanContext.SpanID().String(),
		Attributes: attrs,
	}
}

// SetTraceContext sets trace context in the current context
func SetTraceContext(ctx context.Context, traceCtx *TraceContext) context.Context {
	if traceCtx == nil {
		return ctx
	}

	// Create a new span with the provided trace context
	tracer := otel.Tracer("triggerx-backend")
	ctx, span := tracer.Start(ctx, "trace-context",
		trace.WithAttributes(
			attribute.String("trace.id", traceCtx.TraceID),
			attribute.String("span.id", traceCtx.SpanID),
		),
	)
	defer span.End()

	// Add custom attributes
	for key, value := range traceCtx.Attributes {
		span.SetAttributes(attribute.String(key, value))
	}

	return ctx
}

// InjectTraceContext injects trace context into gRPC metadata
func InjectTraceContext(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}

	// Get the current trace context
	traceCtx := GetTraceContext(ctx)
	if traceCtx != nil {
		md.Set("x-trace-id", traceCtx.TraceID)
		md.Set("x-span-id", traceCtx.SpanID)
	}

	// Use OpenTelemetry propagator to inject trace context
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(ctx, MetadataTextMapCarrier(md))

	return metadata.NewOutgoingContext(ctx, md), nil
}

// ExtractTraceContextFromMetadata extracts trace context from gRPC metadata
func ExtractTraceContextFromMetadata(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}

	// Extract trace context from metadata
	traceID := ""
	spanID := ""

	if values := md.Get("x-trace-id"); len(values) > 0 {
		traceID = values[0]
	}
	if values := md.Get("x-span-id"); len(values) > 0 {
		spanID = values[0]
	}

	// Use OpenTelemetry propagator to extract trace context
	propagator := otel.GetTextMapPropagator()
	ctx = propagator.Extract(ctx, MetadataTextMapCarrier(md))

	// If we have trace context, create a new span
	if traceID != "" {
		tracer := otel.Tracer("triggerx-backend")
		newCtx, span := tracer.Start(ctx, "extracted-trace",
			trace.WithAttributes(
				attribute.String("trace.id", traceID),
				attribute.String("span.id", spanID),
			),
		)
		defer span.End()
		return newCtx
	}

	return ctx
}

// CreateChildSpan creates a child span from the current context
func CreateChildSpan(ctx context.Context, operationName string, attributes ...attribute.KeyValue) (context.Context, trace.Span) {
	tracer := otel.Tracer("triggerx-backend")
	ctx, span := tracer.Start(ctx, operationName,
		trace.WithAttributes(attributes...),
	)
	return ctx, span
}

// AddSpanEvent adds an event to the current span
func AddSpanEvent(ctx context.Context, name string, attributes ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent(name, trace.WithAttributes(attributes...))
}

// SetSpanAttributes sets attributes on the current span
func SetSpanAttributes(ctx context.Context, attributes ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attributes...)
}

// RecordSpanError records an error on the current span
func RecordSpanError(ctx context.Context, err error, attributes ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.RecordError(err, trace.WithAttributes(attributes...))
	span.SetStatus(codes.Error, err.Error())
}

// GetTraceIDFromContext extracts trace ID from context
func GetTraceIDFromContext(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	spanContext := span.SpanContext()
	if spanContext.HasTraceID() {
		return spanContext.TraceID().String()
	}
	return ""
}

// GetSpanIDFromContext extracts span ID from context
func GetSpanIDFromContext(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	spanContext := span.SpanContext()
	if spanContext.HasSpanID() {
		return spanContext.SpanID().String()
	}
	return ""
}

// IsSampled checks if the current span is sampled
func IsSampled(ctx context.Context) bool {
	span := trace.SpanFromContext(ctx)
	spanContext := span.SpanContext()
	return spanContext.TraceFlags().IsSampled()
}

// FormatTraceContext formats trace context for logging
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

// CreateContextWithTrace creates a new context with trace information
func CreateContextWithTrace(parent context.Context, traceID, spanID string) context.Context {
	if traceID == "" {
		return parent
	}

	tracer := otel.Tracer("triggerx-backend")
	ctx, span := tracer.Start(parent, "with-trace-context",
		trace.WithAttributes(
			attribute.String("trace.id", traceID),
			attribute.String("span.id", spanID),
		),
	)
	defer span.End()

	return ctx
}
