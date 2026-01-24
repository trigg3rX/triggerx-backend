package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

// RecoveryMiddleware creates a new recovery middleware that collects panic metrics
func RecoveryMiddleware(logger observability.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Get the endpoint path
				endpoint := c.FullPath()
				if endpoint == "" {
					endpoint = c.Request.URL.Path
				}

				// Record panic recovery
				metrics.PanicRecoveriesTotal.WithLabelValues(endpoint).Inc(c.Request.Context())

				// Safely convert to error and log
				if e, ok := err.(error); ok {
					logger.Error(c.Request.Context(), "Panic recovered",
						observability.Error(e),
						observability.String("stack", string(debug.Stack())))
				} else {
					logger.Error(c.Request.Context(), "Panic recovered (non-error)",
						observability.String("panic_value", fmt.Sprintf("%v", err)),
						observability.String("stack", string(debug.Stack())))
				}

				// Return 500 Internal Server Error
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
				})
				c.Abort()
			}
		}()

		c.Next()
	}
}
