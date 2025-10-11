package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/config"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	gootel "go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

const (
	TraceIDHeader = "X-Trace-ID"
	TraceIDKey    = "trace_id"
	LoggerKey     = "logger"
)

// InitTracer sets up OpenTelemetry tracing with OTLP exporter for Tempo
func InitTracer() (func(context.Context) error, error) {
	exporter, err := otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpoint(config.GetOTTempoEndpoint()),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("triggerx-backend"),
		)),
	)
	gootel.SetTracerProvider(tp)
	return tp.Shutdown, nil
}

// TraceMiddleware enforces trace ID presence, creates a traced logger, and injects both into the Gin context
func TraceMiddleware(baseLogger logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Exempt health check and metrics endpoints from trace ID requirement
		exemptPaths := []string{"/", "/metrics", "/status", "/api/ws/health", "/api/ws/stats"}
		for _, path := range exemptPaths {
			if c.Request.URL.Path == path {
				c.Next()
				return
			}
		}

		// Get trace ID from request header - REQUIRED for all other endpoints
		traceID := c.GetHeader(TraceIDHeader)
		if traceID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Missing trace ID",
				"code":    "MISSING_TRACE_ID",
				"message": "All requests must include X-Trace-ID header",
			})
			c.Abort()
			return
		}

		// Create a traced logger with the traceID for this request
		tracedLogger := baseLogger.WithTraceID(traceID)

		// Get the global tracer
		tracer := gootel.Tracer("triggerx-backend")

		// Start a new span for this request
		ctx, span := tracer.Start(c.Request.Context(), c.Request.URL.Path)
		defer span.End()

		// Set span attributes including the trace ID
		span.SetAttributes(
			semconv.HTTPMethodKey.String(c.Request.Method),
			semconv.HTTPURLKey.String(c.Request.URL.String()),
			semconv.HTTPUserAgentKey.String(c.Request.UserAgent()),
		)

		// Store trace ID and traced logger in context
		c.Set(TraceIDKey, traceID)
		c.Set(LoggerKey, tracedLogger)
		c.Header(TraceIDHeader, traceID)

		// Update request context with span context
		c.Request = c.Request.WithContext(ctx)

		// Process request
		c.Next()

		// Set response status on span
		span.SetAttributes(semconv.HTTPStatusCodeKey.Int(c.Writer.Status()))
	}
}

// GetLogger retrieves the traced logger from the Gin context
func GetLogger(c *gin.Context) logging.Logger {
	logger, exists := c.Get(LoggerKey)
	if !exists {
		// This should never happen if middleware is properly configured
		// Return a no-op logger as fallback to prevent panics
		return logging.NewNoOpLogger()
	}
	return logger.(logging.Logger)
}

// GetTraceID retrieves the trace ID from the Gin context
func GetTraceID(c *gin.Context) string {
	traceID, exists := c.Get(TraceIDKey)
	if !exists {
		return ""
	}
	return traceID.(string)
}
