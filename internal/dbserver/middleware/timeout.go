package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
)

// TimeoutMiddleware creates a new middleware that tracks request timeouts
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a context with timeout
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// Create a channel to track request completion
		done := make(chan struct{})

		// Start a goroutine to process the request
		go func() {
			c.Next()
			close(done)
		}()

		// Wait for either the request to complete or timeout
		select {
		case <-done:
			// Request completed successfully
			return
		case <-ctx.Done():
			// Request timed out
			endpoint := c.FullPath()
			if endpoint == "" {
				endpoint = c.Request.URL.Path
			}

			// Record timeout
			metrics.RequestTimeoutsTotal.WithLabelValues(endpoint).Inc()

			// Abort the request
			c.AbortWithStatusJSON(http.StatusGatewayTimeout, gin.H{
				"error": "Request timeout",
			})
		}
	}
}
