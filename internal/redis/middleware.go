package redis

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/redis/metrics"
)

// MetricsMiddleware collects Redis system metrics
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process request
		c.Next()

		// Update system metrics
		UpdateSystemMetrics()
	}
}

// StartBackgroundMetricsCollection starts periodic metrics collection
func StartBackgroundMetricsCollection() {
	// Start the metrics collection
	metrics.StartMetricsCollection()

	// Update system metrics every 30 seconds
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			UpdateSystemMetrics()
		}
	}()
}
