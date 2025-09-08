package tracing

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// TraceInterceptor provides OpenTelemetry tracing for gRPC requests
func TraceInterceptor(serviceName string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Extract trace context from incoming metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}

		// Extract trace context using OpenTelemetry propagator
		propagator := otel.GetTextMapPropagator()
		ctx = propagator.Extract(ctx, MetadataTextMapCarrier(md))

		// Start a new span for this gRPC request
		tracer := otel.Tracer(serviceName)
		ctx, span := tracer.Start(ctx, info.FullMethod,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				attribute.String("grpc.method", info.FullMethod),
				attribute.String("grpc.service", serviceName),
			),
		)
		defer span.End()

		// Add request metadata as span attributes
		if len(md.Get("user-agent")) > 0 {
			span.SetAttributes(attribute.String("grpc.user_agent", md.Get("user-agent")[0]))
		}
		if len(md.Get("content-type")) > 0 {
			span.SetAttributes(attribute.String("grpc.content_type", md.Get("content-type")[0]))
		}

		// Process the request
		startTime := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(startTime)

		// Set span attributes based on response
		span.SetAttributes(
			attribute.Int64("grpc.duration_ms", duration.Milliseconds()),
		)

		// Handle errors
		if err != nil {
			st, _ := status.FromError(err)
			span.SetStatus(codes.Error, err.Error())
			span.SetAttributes(
				attribute.Int("grpc.status_code", int(st.Code())),
				attribute.String("grpc.status_message", st.Message()),
			)
		} else {
			span.SetStatus(codes.Ok, "")
			span.SetAttributes(attribute.Int("grpc.status_code", int(codes.Ok)))
		}

		return resp, err
	}
}

// TraceClientInterceptor provides OpenTelemetry tracing for gRPC client requests
func TraceClientInterceptor(serviceName string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// Start a new span for this gRPC client request
		tracer := otel.Tracer(serviceName)
		ctx, span := tracer.Start(ctx, method,
			trace.WithSpanKind(trace.SpanKindClient),
			trace.WithAttributes(
				attribute.String("grpc.method", method),
				attribute.String("grpc.service", serviceName),
				attribute.String("grpc.target", cc.Target()),
			),
		)
		defer span.End()

		// Inject trace context into outgoing metadata
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}

		propagator := otel.GetTextMapPropagator()
		propagator.Inject(ctx, MetadataTextMapCarrier(md))

		// Create new context with metadata
		ctx = metadata.NewOutgoingContext(ctx, md)

		// Make the gRPC call
		startTime := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		duration := time.Since(startTime)

		// Set span attributes based on response
		span.SetAttributes(
			attribute.Int64("grpc.duration_ms", duration.Milliseconds()),
		)

		// Handle errors
		if err != nil {
			st, _ := status.FromError(err)
			span.SetStatus(codes.Error, err.Error())
			span.SetAttributes(
				attribute.Int("grpc.status_code", int(st.Code())),
				attribute.String("grpc.status_message", st.Message()),
			)
		} else {
			span.SetStatus(codes.Ok, "")
			span.SetAttributes(attribute.Int("grpc.status_code", int(codes.Ok)))
		}

		return err
	}
}

// MetadataTextMapCarrier implements the OpenTelemetry TextMapCarrier interface for gRPC metadata
type MetadataTextMapCarrier metadata.MD

// Get retrieves a value from the metadata
func (m MetadataTextMapCarrier) Get(key string) string {
	values := metadata.MD(m).Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

// Set sets a value in the metadata
func (m MetadataTextMapCarrier) Set(key, value string) {
	metadata.MD(m).Set(key, value)
}

// Keys returns all keys in the metadata
func (m MetadataTextMapCarrier) Keys() []string {
	keys := make([]string, 0, len(m))
	for k := range metadata.MD(m) {
		keys = append(keys, k)
	}
	return keys
}

// WithTraceContext adds trace context to gRPC metadata
func WithTraceContext(ctx context.Context) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}

	propagator := otel.GetTextMapPropagator()
	propagator.Inject(ctx, MetadataTextMapCarrier(md))

	return metadata.NewOutgoingContext(ctx, md)
}

// ExtractTraceContext extracts trace context from gRPC metadata
func ExtractTraceContext(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}

	propagator := otel.GetTextMapPropagator()
	return propagator.Extract(ctx, MetadataTextMapCarrier(md))
}
