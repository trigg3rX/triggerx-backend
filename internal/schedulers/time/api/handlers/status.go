package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// StatusHandler handles status endpoint requests
type StatusHandler struct {
	logger    logging.Logger
}

// NewStatusHandler creates a new status handler
func NewStatusHandler(logger logging.Logger) *StatusHandler {
	return &StatusHandler{
		logger: logger,
	}
}

// Status handles status endpoint requests
func (h *StatusHandler) Status(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
