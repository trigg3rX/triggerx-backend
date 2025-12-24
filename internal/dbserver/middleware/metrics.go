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
		ctx := c.Request.Context() // Use request context, though background context is also acceptable for metrics usually

		// Increment active requests
		if metrics.ActiveRequests != nil {
			metrics.ActiveRequests.WithLabelValues(path).Set(ctx, 1)
			defer metrics.ActiveRequests.WithLabelValues(path).Set(ctx, -1) // Use background context for defer to ensure execution
		}

		// Process request
		c.Next()

		// Record request duration
		duration := time.Since(startTime).Seconds()
		if metrics.HTTPRequestDuration != nil {
			metrics.HTTPRequestDuration.WithLabelValues(method, path).Record(ctx, duration)
		}

		// Record total requests with status code
		status := c.Writer.Status()
		if metrics.HTTPRequestsTotal != nil {
			metrics.HTTPRequestsTotal.WithLabelValues(method, path, fmt.Sprint(rune(status))).Inc(ctx)
		}

		// Update average response time
		if metrics.AverageResponseTime != nil {
			metrics.AverageResponseTime.WithLabelValues(path).Set(ctx, duration)
		}

		// Update requests per second
		// Note: RequestsPerSecond in Prometheus was likely a Gauge calculated/set periodically or a Counter.
		// In the new definition it is a GaugeVec. Incrementing a Gauge directly as a rate is unusual but we follow the pattern.
		// If it was meant to be a rate, usually we use a Counter and let Prometheus calculate rate().
		// The original code had `.Inc()`, which suggests it might have been a Counter or a Gauge treated as one?
		// Checking previous definition: `RequestsPerSecond = promauto.NewGaugeVec(...)`
		// So it was a Gauge. Incrementing a Gauge... okay.
		if metrics.RequestsPerSecond != nil {
			metrics.RequestsPerSecond.WithLabelValues(path).Set(ctx, 1)
		}
	}
}
