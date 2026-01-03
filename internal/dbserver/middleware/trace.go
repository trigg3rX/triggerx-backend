package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
// It requires the X-Trace-ID header to be present in the request and rejects requests without it
func TraceMiddleware(tracer observability.Tracer) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Require X-Trace-ID header
		traceID := c.GetHeader(TraceIDHeader)
		if traceID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "X-Trace-ID header is required",
			})
			c.Abort()
			return
		}

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
				attribute.String("trace.id", traceID),
			),
		)
		defer span.End()

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

// httpStatusText returns a text representation of the HTTP status code
func httpStatusText(code int) string {
	switch code {
	case http.StatusOK:
		return "OK"
	case http.StatusBadRequest:
		return "Bad Request"
	case http.StatusUnauthorized:
		return "Unauthorized"
	case http.StatusForbidden:
		return "Forbidden"
	case http.StatusNotFound:
		return "Not Found"
	case http.StatusInternalServerError:
		return "Internal Server Error"
	default:
		return http.StatusText(code)
	}
}
