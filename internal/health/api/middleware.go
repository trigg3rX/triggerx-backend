package api

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/trigg3rX/triggerx-backend/internal/health/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

const TraceIDHeader = "X-Trace-ID"
const TraceIDKey = "trace_id"

// TraceMiddleware creates a gin middleware for OpenTelemetry tracing
func TraceMiddleware(tracer observability.Tracer) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract trace context from HTTP headers using OpenTelemetry propagator
		propagator := otel.GetTextMapPropagator()
		ctx := propagator.Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

		// Start a new span for this HTTP request
		ctx, span := tracer.Start(ctx, c.Request.URL.Path,
			observability.WithSpanKind(trace.SpanKindServer),
			observability.WithAttributes(
				semconv.HTTPMethodKey.String(c.Request.Method),
				semconv.HTTPURLKey.String(c.Request.URL.String()),
				semconv.HTTPUserAgentKey.String(c.Request.UserAgent()),
				attribute.String("http.client_ip", c.ClientIP()),
			),
		)
		defer span.End()

		// Get or generate trace ID
		traceID := c.GetHeader(TraceIDHeader)
		if traceID == "" {
			spanContext := span.SpanContext()
			if spanContext.HasTraceID() {
				traceID = spanContext.TraceID().String()
			} else {
				traceID = uuid.New().String()
			}
		}

		// Store trace ID in context for handlers
		c.Set(TraceIDKey, traceID)
		c.Header(TraceIDHeader, traceID)

		// Update request context with span context
		c.Request = c.Request.WithContext(ctx)

		// Process request
		c.Next()

		// Set response status on span
		statusCode := c.Writer.Status()
		span.SetAttributes(
			semconv.HTTPStatusCodeKey.Int(statusCode),
		)

		// Set span status based on HTTP status code
		if statusCode >= 400 {
			span.SetStatus(codes.Error, httpStatusText(statusCode))
		} else {
			span.SetStatus(codes.Ok, "")
		}
	}
}

// LoggerMiddleware creates a gin middleware for logging
func LoggerMiddleware(logger observability.Logger) gin.HandlerFunc {
	middlewareLogger := logger
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		ctx := c.Request.Context()

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()
		statusCode := fmt.Sprintf("%d", status)

		// Record HTTP metrics
		metrics.RecordHTTPRequest(ctx, method, path, statusCode, duration)

		middlewareLogger.Debug(ctx, "HTTP Request",
			observability.String("method", method),
			observability.String("path", path),
			observability.Int("status", status),
			observability.Int64("duration_ms", duration.Milliseconds()),
			observability.String("ip", c.ClientIP()),
		)
	}
}

// httpStatusText returns a text representation of HTTP status code
func httpStatusText(code int) string {
	switch {
	case code >= 500:
		return "Internal Server Error"
	case code >= 400:
		return "Client Error"
	case code >= 300:
		return "Redirect"
	case code >= 200:
		return "Success"
	default:
		return "Unknown"
	}
}
