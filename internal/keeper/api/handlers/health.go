package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// HealthService defines the interface for health checks
type HealthService interface {
	CheckIn(ctx context.Context) error
}

// HealthHandler handles health check requests
type HealthHandler struct {
	logger  logging.Logger
	service HealthService
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(logger logging.Logger, service HealthService) *HealthHandler {
	return &HealthHandler{
		logger:  logger,
		service: service,
	}
}

// Check handles health check requests
func (h *HealthHandler) Check(c *gin.Context) {
	err := h.service.CheckIn(c.Request.Context())
	if err != nil {
		h.logger.Error("Health check failed", "error", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "unhealthy",
			"error":     err.Error(),
			"timestamp": time.Now().UTC(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
	})
}
