package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/event/scheduler"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type SchedulerHandler struct {
	logger    logging.Logger
	scheduler *scheduler.EventBasedScheduler
}

func NewSchedulerHandler(logger logging.Logger, scheduler *scheduler.EventBasedScheduler) *SchedulerHandler {
	return &SchedulerHandler{
		logger:    logger,
		scheduler: scheduler,
	}
}

// GetStats returns current scheduler statistics
func (h *SchedulerHandler) GetStats(c *gin.Context) {
	stats := h.scheduler.GetStats()

	response := gin.H{
		"status":    "success",
		"data":      stats,
		"timestamp": time.Now().UTC(),
	}

	c.JSON(http.StatusOK, response)
}

// Stop stops the scheduler
func (h *SchedulerHandler) Stop(c *gin.Context) {
	h.logger.Info("Received request to stop scheduler")

	// Stop the scheduler
	h.scheduler.Stop()

	response := gin.H{
		"status":    "success",
		"message":   "Scheduler stopped successfully",
		"timestamp": time.Now().UTC(),
	}

	c.JSON(http.StatusOK, response)
}

// Start starts the scheduler (placeholder for future implementation)
func (h *SchedulerHandler) Start(c *gin.Context) {
	h.logger.Info("Received request to start scheduler")

	// Note: Starting a stopped scheduler would require additional implementation
	// For now, we'll return a message indicating the current state

	response := gin.H{
		"status":    "info",
		"message":   "Scheduler start functionality not implemented - scheduler runs automatically on service start",
		"timestamp": time.Now().UTC(),
	}

	c.JSON(http.StatusOK, response)
}
