package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
)

// MetricsMiddleware tracks HTTP metrics for all requests
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path // Fallback for unmatched routes
		}
		method := c.Request.Method
		ctx := c.Request.Context()

		// Track active requests (increment on start, decrement on end)
		metrics.IncrementActiveRequests(path)
		defer metrics.DecrementActiveRequests(path)

		// Process request
		c.Next()

		// Record request duration
		duration := time.Since(startTime).Seconds()
		if metrics.HTTPRequestDuration != nil {
			metrics.HTTPRequestDuration.WithLabelValues(method, path).Record(ctx, duration)
		}

		// Record total requests with status code (properly formatted)
		status := c.Writer.Status()
		statusStr := strconv.Itoa(status)
		if metrics.HTTPRequestsTotal != nil {
			metrics.HTTPRequestsTotal.WithLabelValues(method, path, statusStr).Inc(ctx)
		}

		// Track request for RPS and average response time calculation
		metrics.TrackRequest(path)
		metrics.TrackResponseTime(path, duration)
	}
}
