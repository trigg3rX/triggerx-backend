package api

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/metrics"
)

// MetricsMiddleware tracks HTTP request metrics
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(startTime)

		// Track metrics
		method := c.Request.Method
		endpoint := c.Request.URL.Path
		statusCode := strconv.Itoa(c.Writer.Status())

		// Track HTTP request
		metrics.TrackHTTPRequest(method, endpoint, statusCode)

		// Log slow requests (optional)
		if duration > 5*time.Second {
			metrics.TrackTimeout("http_request")
		}
	}
}
