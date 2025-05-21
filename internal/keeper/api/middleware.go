package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// LoggerMiddleware creates a gin middleware for logging requests
func LoggerMiddleware(logger logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log request details
		param := gin.LogFormatterParams{
			Request: c.Request,
			Keys:    c.Keys,
		}

		// Stop timer
		param.TimeStamp = time.Now()
		param.Latency = param.TimeStamp.Sub(start)

		logger.Info("Request processed",
			"status", c.Writer.Status(),
			"method", c.Request.Method,
			"path", path,
			"query", raw,
			"ip", c.ClientIP(),
			"latency", param.Latency,
			"user-agent", c.Request.UserAgent(),
		)
	}
}

// ErrorMiddleware handles errors in a consistent way
func ErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			// Get the last error
			err := c.Errors.Last()

			// If the response hasn't been written yet
			if !c.Writer.Written() {
				c.JSON(c.Writer.Status(), gin.H{
					"error": err.Error(),
				})
			}
		}
	}
}
