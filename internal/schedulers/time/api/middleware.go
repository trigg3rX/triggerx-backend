package api

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/trigg3rX/triggerx-backend-imua/internal/schedulers/time/metrics"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/logging"
	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

const TraceIDHeader = "X-Trace-ID"
const TraceIDKey = "trace_id"

// TraceMiddleware adds trace ID to requests and starts an OpenTelemetry span
func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tracer := otel.Tracer("triggerx-backend")
		ctx, span := tracer.Start(c.Request.Context(), c.Request.URL.Path)
		defer span.End()

		span.SetAttributes(
			semconv.HTTPMethodKey.String(c.Request.Method),
			semconv.HTTPURLKey.String(c.Request.URL.String()),
			semconv.HTTPUserAgentKey.String(c.Request.UserAgent()),
		)

		traceID := c.GetHeader(TraceIDHeader)
		if traceID == "" {
			spanContext := span.SpanContext()
			if spanContext.HasTraceID() {
				traceID = spanContext.TraceID().String()
			} else {
				traceID = uuid.New().String()
			}
		}

		c.Set(TraceIDKey, traceID)
		c.Header(TraceIDHeader, traceID)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
		span.SetAttributes(semconv.HTTPStatusCodeKey.Int(c.Writer.Status()))
	}
}

// MetricsMiddleware collects HTTP request and system metrics
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process request
		c.Next()

		// Collect HTTP metrics
		statusCode := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method
		endpoint := c.Request.URL.Path

		metrics.TrackHTTPRequest(method, endpoint, statusCode)
	}
}

// LoggerMiddleware creates a gin middleware for logging API group requests only
func LoggerMiddleware(logger logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only log requests under the /api/ path
		if len(c.Request.URL.Path) < 5 || c.Request.URL.Path[:5] != "/api/" {
			c.Next()
			return
		}

		startTime := time.Now()
		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery
		traceID, _ := c.Get(TraceIDKey)

		// Process request
		c.Next()

		duration := time.Since(startTime)
		statusCode := c.Writer.Status()

		logger.Info("Request processed",
			"trace_id", traceID,
			"status", statusCode,
			"method", c.Request.Method,
			"path", path,
			"query", rawQuery,
			"ip", c.ClientIP(),
			"latency", duration,
			"user-agent", c.Request.UserAgent(),
		)
	}
}

// ErrorMiddleware handles errors in a consistent way
func ErrorMiddleware(logger logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			// Get the last error
			err := c.Errors.Last()
			traceID, _ := c.Get(TraceIDKey)

			logger.Error("Error",
				"trace_id", traceID,
				"error", err.Error(),
				"path", c.Request.URL.Path,
			)

			// If the response hasn't been written yet
			if !c.Writer.Written() {
				c.JSON(c.Writer.Status(), gin.H{
					"error":    err.Error(),
					"trace_id": traceID,
				})
			}
		}
	}
}
