package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

const TraceIDHeader = "X-Trace-ID"
const TraceIDKey = "trace_id"

// MetricsHandler handles metrics collection and exposure
type MetricsHandler struct {
	collector *metrics.Collector
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler() *MetricsHandler {
	collector := metrics.NewCollector()
	return &MetricsHandler{
		collector: collector,
	}
}

// Metrics exposes Prometheus metrics
func (h *MetricsHandler) Metrics(c *gin.Context) {
	h.collector.Handler().ServeHTTP(c.Writer, c.Request)
}

// Start initializes the metrics collection (call this once during startup)
func (h *MetricsHandler) Start() {
	h.collector.Start()
}

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

// LoggerMiddleware creates a gin middleware for logging requests
func LoggerMiddleware(logger logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip logging for metrics endpoint
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		traceID, _ := c.Get(TraceIDKey)

		// Process request
		c.Next()

		duration := time.Since(start)
		statusCode := c.Writer.Status()

		logger.Info("Request processed",
			"trace_id", traceID,
			"status", statusCode,
			"method", c.Request.Method,
			"path", path,
			"query", raw,
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

// StreamMetricsMiddleware tracks stream-related metrics
func StreamMetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Track incoming stream operations
		switch path {
		case "/task/validate":
			if c.Request.Method == "POST" {
				metrics.TasksReadFromStreamTotal.WithLabelValues("processing", "received").Inc()
			}
		case "/p2p/message":
			if c.Request.Method == "POST" {
				metrics.JobsReadFromStreamTotal.WithLabelValues("running", "received").Inc()
			}
		}

		c.Next()

		statusCode := c.Writer.Status()

		// Track completed operations based on endpoint and status
		if statusCode >= 200 && statusCode < 300 {
			switch path {
			case "/task/validate":
				// Task validation endpoint - tracks task completion
				metrics.TasksReadFromStreamTotal.WithLabelValues("processing", "success").Inc()
				metrics.TaskProcessingToCompletedTotal.Inc()
			case "/p2p/message":
				// P2P message handling - tracks job processing
				metrics.JobsReadFromStreamTotal.WithLabelValues("running", "success").Inc()
			case "/streams/info":
				// Stream info endpoint - tracks monitoring requests
				metrics.ConnectionChecksTotal.WithLabelValues("success").Inc()
			}
		} else if statusCode >= 400 {
			// Track failed operations
			switch path {
			case "/task/validate":
				metrics.TasksReadFromStreamTotal.WithLabelValues("processing", "failure").Inc()
			case "/p2p/message":
				metrics.JobsReadFromStreamTotal.WithLabelValues("running", "failure").Inc()
			case "/streams/info":
				metrics.ConnectionChecksTotal.WithLabelValues("failure").Inc()
			}
		}
	}
}
