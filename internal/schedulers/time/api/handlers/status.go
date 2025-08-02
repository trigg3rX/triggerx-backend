package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// StatusHandler handles status endpoint requests
type StatusHandler struct {
	logger logging.Logger
}

// NewStatusHandler creates a new status handler
func NewStatusHandler(logger logging.Logger) *StatusHandler {
	return &StatusHandler{
		logger: logger,
	}
}

// Status handles status endpoint requests
func (h *StatusHandler) Status(c *gin.Context) {
	traceID := getTraceID(c)
	h.logger.Info("[Status] trace_id=" + traceID + " - Checking service health")
	response := gin.H{
		"status":    "healthy",
		"service":   "condition-scheduler",
		"timestamp": time.Now().UTC(),
		"uptime":    time.Since(time.Now()).String(), // This would be calculated from startup time
	}

	c.JSON(http.StatusOK, response)
}
