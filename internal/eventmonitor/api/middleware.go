package api

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

// LoggerMiddleware creates a gin middleware for logging and metrics
func LoggerMiddleware(logger observability.Logger) gin.HandlerFunc {
	middlewareLogger := logger
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		ctx := c.Request.Context()

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()
		statusCode := fmt.Sprintf("%d", status)

		// Record HTTP metrics
		metrics.RecordHTTPRequest(ctx, method, path, statusCode, duration)

		middlewareLogger.Debug(ctx, "HTTP Request",
			observability.String("method", method),
			observability.String("path", path),
			observability.Int("status", status),
			observability.Int64("duration_ms", duration.Milliseconds()),
			observability.String("ip", c.ClientIP()),
		)
	}
}

