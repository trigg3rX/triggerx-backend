package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend-imua/internal/keeper/core/execution"
	"github.com/trigg3rX/triggerx-backend-imua/internal/keeper/core/validation"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/logging"
)

const TraceIDKey = "trace_id"

// TaskHandler handles task-related requests
type TaskHandler struct {
	logger    logging.Logger
	executor  execution.TaskExecutor
	validator validation.TaskValidator
}

// NewTaskHandler creates a new task handler
func NewTaskHandler(logger logging.Logger, executor execution.TaskExecutor, validator validation.TaskValidator) *TaskHandler {
	return &TaskHandler{
		logger:    logger,
		executor:  executor,
		validator: validator,
	}
}

func (h *TaskHandler) getTraceID(c *gin.Context) string {
	traceID, exists := c.Get(TraceIDKey)
	if !exists {
		return ""
	}
	return traceID.(string)
}
