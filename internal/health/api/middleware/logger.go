package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	healthmetrics "github.com/trigg3rX/triggerx-backend/internal/health/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// Logger creates a gin middleware for logging HTTP requests
func Logger(logger logging.Logger, healthMetrics *healthmetrics.HealthMetrics) gin.HandlerFunc {
	middlewareLogger := logger
	return func(c *gin.Context) {
		// Skip logging for metrics endpoint
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		// Record HTTP metrics
		statusCode := fmt.Sprintf("%d", status)
		healthMetrics.HTTPRequestsTotal.WithLabelValues(method, path, statusCode).Inc()
		healthMetrics.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())

		middlewareLogger.Debug("HTTP Request",
			"method", method,
			"path", path,
			"status", status,
			"duration_ms", duration.Milliseconds(),
			"ip", c.ClientIP(),
		)
	}
}
