package api

import (
	"context"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
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
		// Note: We use background context as this is system-level info, not request-scoped
		ctx := context.Background()
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		if metrics.MemoryUsageBytes != nil {
			metrics.MemoryUsageBytes.Set(ctx, float64(memStats.Alloc))
		}
		if metrics.CPUUsagePercent != nil {
			metrics.CPUUsagePercent.Set(ctx, float64(memStats.Sys))
		}
		if metrics.GoroutinesActive != nil {
			metrics.GoroutinesActive.Set(ctx, float64(runtime.NumGoroutine()))
		}
		if metrics.GCDurationSeconds != nil {
			metrics.GCDurationSeconds.Set(ctx, float64(memStats.PauseTotalNs)/1e9)
		}
	}
}

// LoggerMiddleware creates a gin middleware for logging requests
func LoggerMiddleware(logger observability.Logger) gin.HandlerFunc {
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

		logger.Info(c.Request.Context(), "Request processed",
			observability.Any("trace_id", traceID),
			observability.Int("status", statusCode),
			observability.String("method", c.Request.Method),
			observability.String("path", path),
			observability.String("query", raw),
			observability.String("ip", c.ClientIP()),
			observability.Duration("latency", duration),
			observability.String("user-agent", c.Request.UserAgent()),
		)
	}
}

// ErrorMiddleware handles errors in a consistent way
func ErrorMiddleware(logger observability.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			// Get the last error
			err := c.Errors.Last()
			traceID, _ := c.Get(TraceIDKey)

			logger.Error(c.Request.Context(), "Error",
				observability.Any("trace_id", traceID),
				observability.Error(err),
				observability.String("path", c.Request.URL.Path),
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
		ctx := c.Request.Context()

		// Track incoming tasks
		if path == "/p2p/message" && c.Request.Method == "POST" {
			if metrics.TasksReceivedTotal != nil {
				metrics.TasksReceivedTotal.Inc(ctx)
			}
		}

		c.Next()

		duration := time.Since(start)
		statusCode := c.Writer.Status()

		// Track completed tasks based on endpoint and status
		if statusCode >= 200 && statusCode < 300 {
			switch path {
			case "/p2p/message":
				// Task execution endpoint
				if metrics.TasksPerDay != nil {
					metrics.TasksPerDay.WithLabelValues("executed").Inc(ctx)
				}
				if metrics.TasksCompletedTotal != nil {
					metrics.TasksCompletedTotal.WithLabelValues("executed").Inc(ctx)
				}
				if metrics.TaskDurationSeconds != nil {
					metrics.TaskDurationSeconds.WithLabelValues("executed").Record(ctx, duration.Seconds())
				}
				// metrics.AverageTaskCompletionTimeSeconds.WithLabelValues("executed").Set(duration.Seconds())
			case "/task/validate":
				// Task validation endpoint
				if metrics.TasksPerDay != nil {
					metrics.TasksPerDay.WithLabelValues("validated").Inc(ctx)
				}
				if metrics.TasksCompletedTotal != nil {
					metrics.TasksCompletedTotal.WithLabelValues("validated").Inc(ctx)
				}
				if metrics.TaskDurationSeconds != nil {
					metrics.TaskDurationSeconds.WithLabelValues("validated").Record(ctx, duration.Seconds())
				}
				// metrics.AverageTaskCompletionTimeSeconds.WithLabelValues("validated").Set(duration.Seconds())
			}
		}
	}
}

// RestartTrackingMiddleware tracks service restarts
func RestartTrackingMiddleware() gin.HandlerFunc {
	// This should be called once during service startup
	if metrics.RestartsTotal != nil {
		metrics.RestartsTotal.Inc(context.Background())
	}

	return func(c *gin.Context) {
		c.Next()
	}
}
