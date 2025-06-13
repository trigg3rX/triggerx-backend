package middleware

import (
	"time"

	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
)

// MetricsMiddleware tracks HTTP metrics for all requests
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		path := c.FullPath()
		method := c.Request.Method

		// Increment active requests
		metrics.ActiveRequests.WithLabelValues(path).Inc()
		defer metrics.ActiveRequests.WithLabelValues(path).Dec()

		// Process request
		c.Next()

		// Record request duration
		duration := time.Since(startTime).Seconds()
		metrics.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)

		// Record total requests with status code
		status := c.Writer.Status()
		metrics.HTTPRequestsTotal.WithLabelValues(method, path, fmt.Sprint(rune(status))).Inc()

		// Update average response time
		metrics.AverageResponseTime.WithLabelValues(path).Set(duration)

		// Update requests per second
		metrics.RequestsPerSecond.WithLabelValues(path).Inc()
	}
}
