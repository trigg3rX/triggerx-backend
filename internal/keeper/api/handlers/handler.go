package handlers

import (
	"github.com/trigg3rX/triggerx-backend/internal/keeper/core/execution"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/core/validation"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)


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
