package handlers

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/core/validation"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

const TraceIDKey = "trace_id"

// TaskExecutorInterface defines the interface for task execution
type TaskExecutorInterface interface {
	ExecuteTask(ctx context.Context, task *types.SendTaskDataToKeeper, traceID string) (bool, error)
}

// IPFSFetcherInterface defines the interface for IPFS content fetching
type IPFSFetcherInterface interface {
	FetchContent(cid string) (types.IPFSData, error)
}

// TaskHandler handles task-related requests
type TaskHandler struct {
	logger      logging.Logger
	executor    TaskExecutorInterface
	validator   validation.TaskValidatorInterface
	ipfsFetcher IPFSFetcherInterface
}

// NewTaskHandler creates a new task handler
func NewTaskHandler(logger logging.Logger, executor TaskExecutorInterface, validator validation.TaskValidatorInterface) *TaskHandler {
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
