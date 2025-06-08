package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// Handler encapsulates the dependencies for health handlers
type Handler struct {
	logger       logging.Logger
}

// NewHandler creates a new instance of Handler
func NewHandler(logger logging.Logger) *Handler {
	return &Handler{
		logger:       logger,
	}
}

// LoggerMiddleware creates a gin middleware for logging
func LoggerMiddleware(logger logging.Logger) gin.HandlerFunc {
	middlewareLogger := logger
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		if status >= 500 {
			middlewareLogger.Error("HTTP Request",
				"method", method,
				"path", path,
				"status", status,
				"duration_ms", duration.Milliseconds(),
				"ip", c.ClientIP(),
			)
		} else if status >= 400 {
			middlewareLogger.Warn("HTTP Request",
				"method", method,
				"path", path,
				"status", status,
				"duration_ms", duration.Milliseconds(),
				"ip", c.ClientIP(),
			)
		} else {
			middlewareLogger.Info("HTTP Request",
				"method", method,
				"path", path,
				"status", status,
				"duration_ms", duration.Milliseconds(),
				"ip", c.ClientIP(),
			)
		}
	}
}

// RegisterRoutes registers all HTTP routes for the health service
func RegisterRoutes(router *gin.Engine, logger logging.Logger) {
	handler := NewHandler(logger)

	router.GET("/", handler.handleRoot)
	router.POST("/task/validate", handler.ValidateTask)
	router.POST("/p2p/message", handler.HandleP2PMessage)
}

func (h *Handler) handleRoot(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service":   "TriggerX Redis Service",
		"status":    "running",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

