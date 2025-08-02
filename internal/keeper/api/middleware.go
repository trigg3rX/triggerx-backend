package api

import (
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

const TraceIDHeader = "X-Trace-ID"
const TraceIDKey = "trace_id"

func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader(TraceIDHeader)
		if traceID == "" {
			traceID = uuid.New().String()
		}

		c.Set(TraceIDKey, traceID)
		c.Header(TraceIDHeader, traceID)
		c.Next()
	}
}

// MetricsMiddleware collects HTTP request metrics
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process request
		c.Next()

		// Update system metrics
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		metrics.MemoryUsageBytes.Set(float64(memStats.Alloc))
		metrics.CPUUsagePercent.Set(float64(memStats.Sys))
		metrics.GoroutinesActive.Set(float64(runtime.NumGoroutine()))
		metrics.GCDurationSeconds.Set(float64(memStats.PauseTotalNs) / 1e9)
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

// TaskMetricsMiddleware tracks task-related metrics
func TaskMetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		// Track incoming tasks
		if path == "/p2p/message" && c.Request.Method == "POST" {
			metrics.TasksReceivedTotal.Inc()
		}

		c.Next()

		duration := time.Since(start)
		statusCode := c.Writer.Status()

		// Track completed tasks based on endpoint and status
		if statusCode >= 200 && statusCode < 300 {
			switch path {
			case "/p2p/message":
				// Task execution endpoint
				metrics.TasksPerDay.WithLabelValues("executed").Inc()
				metrics.TasksCompletedTotal.WithLabelValues("executed").Inc()
				metrics.TaskDurationSeconds.WithLabelValues("executed").Observe(duration.Seconds())
				// metrics.AverageTaskCompletionTimeSeconds.WithLabelValues("executed").Set(duration.Seconds())
			case "/task/validate":
				// Task validation endpoint
				metrics.TasksPerDay.WithLabelValues("validated").Inc()
				metrics.TasksCompletedTotal.WithLabelValues("validated").Inc()
				metrics.TaskDurationSeconds.WithLabelValues("validated").Observe(duration.Seconds())
				// metrics.AverageTaskCompletionTimeSeconds.WithLabelValues("validated").Set(duration.Seconds())
			}
		}
	}
}

// RestartTrackingMiddleware tracks service restarts
func RestartTrackingMiddleware() gin.HandlerFunc {
	// This should be called once during service startup
	metrics.RestartsTotal.Inc()

	return func(c *gin.Context) {
		c.Next()
	}
}
