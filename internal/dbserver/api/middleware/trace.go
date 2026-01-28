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
// It requires the X-Trace-ID header only for the create-job handler (/api/jobs POST)
// All other endpoints are allowed without traces and will have trace IDs generated if not provided
func TraceMiddleware(tracer observability.Tracer) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Health check endpoints don't require X-Trace-ID header
		healthCheckPaths := []string{"/status", "/api/health", "/metrics"}
		for _, path := range healthCheckPaths {
			if c.Request.URL.Path == path {
				break
			}
		}

		// Require X-Trace-ID for specific endpoints
		isCreateJob := c.Request.URL.Path == "/api/jobs" && c.Request.Method == http.MethodPost
		isGetFees := c.Request.URL.Path == "/api/fees" && c.Request.Method == http.MethodGet

		traceID := c.GetHeader(TraceIDHeader)

		// Require X-Trace-ID header for create-job and get-fees endpoints
		if traceID == "" && (isCreateJob || isGetFees) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "X-Trace-ID header is required",
			})
			c.Abort()
			return
		}

		// Generate a trace ID if not provided (for non-create-job endpoints)
		if traceID == "" {
			// Generate a simple trace ID based on client IP and timestamp
			traceID = "auto-" + c.ClientIP() + "-" + c.Request.URL.Path
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
