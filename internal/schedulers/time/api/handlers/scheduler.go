package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend-imua/internal/schedulers/time/scheduler"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/logging"
)

type SchedulerHandler struct {
	logger    logging.Logger
	scheduler *scheduler.TimeBasedScheduler
}

func NewSchedulerHandler(logger logging.Logger, scheduler *scheduler.TimeBasedScheduler) *SchedulerHandler {
	return &SchedulerHandler{
		logger:    logger,
		scheduler: scheduler,
	}
}

// getTraceID retrieves the trace ID from the Gin context
func getTraceID(c *gin.Context) string {
	traceID, exists := c.Get("trace_id")
	if !exists {
		return ""
	}
	return traceID.(string)
}

// GetStats returns current scheduler statistics
func (h *SchedulerHandler) GetStats(c *gin.Context) {
	traceID := getTraceID(c)
	h.logger.Info("[GetStats] trace_id=" + traceID + " - Getting scheduler statistics")
	stats := h.scheduler.GetStats()

	response := gin.H{
		"status":    "success",
		"data":      stats,
		"timestamp": time.Now().UTC(),
	}

	c.JSON(http.StatusOK, response)
}
