package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

// TimeoutMiddleware creates a new middleware that applies a context deadline to each request.
// Handlers should check c.Request.Context().Err() or use context-aware operations to
// respect the timeout. This avoids spawning a goroutine that races on gin.Context.
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip timeout middleware for WebSocket connections
		if c.GetHeader("Upgrade") == "websocket" || c.GetHeader("Connection") == "Upgrade" {
			c.Next()
			return
		}

		// Create a context with timeout and attach it to the request
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// Replace the request context so downstream handlers/DB calls respect the deadline
		c.Request = c.Request.WithContext(ctx)

		// Process request synchronously — context cancellation will propagate
		// to any context-aware operation (DB queries, HTTP calls, etc.)
		c.Next()
	}
}
