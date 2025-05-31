package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type StatusHandler struct {
	logger logging.Logger
}

func NewStatusHandler(logger logging.Logger) *StatusHandler {
	return &StatusHandler{
		logger: logger,
	}
}

// Status returns the health status of the condition scheduler service
func (h *StatusHandler) Status(c *gin.Context) {
	response := gin.H{
		"status":    "healthy",
		"service":   "condition-scheduler",
		"timestamp": time.Now().UTC(),
		"uptime":    time.Since(time.Now()).String(), // This would be calculated from startup time
	}

	c.JSON(http.StatusOK, response)
}
