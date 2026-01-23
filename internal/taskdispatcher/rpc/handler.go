package rpc

import (
	"context"
	"fmt"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	rpcpkg "github.com/trigg3rX/triggerx-backend/pkg/rpc"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// TaskDispatcherHandler implements the generic RPC handler interface
type TaskDispatcherHandler struct {
	logger     observability.Logger
	dispatcher TaskDispatcherInterface
}

// TaskDispatcherInterface defines the interface for task dispatcher operations
type TaskDispatcherInterface interface {
	SubmitTaskFromScheduler(ctx context.Context, req *types.SchedulerTaskRequest) (*types.TaskManagerAPIResponse, error)
}

// NewTaskDispatcherHandler creates a new RPC handler
func NewTaskDispatcherHandler(logger observability.Logger, dispatcher TaskDispatcherInterface) *TaskDispatcherHandler {
	return &TaskDispatcherHandler{
		logger:     logger,
		dispatcher: dispatcher,
	}
}

// Handle routes incoming RPC requests based on the method name
func (h *TaskDispatcherHandler) Handle(ctx context.Context, method string, request interface{}) (interface{}, error) {
	switch method {
	case "submit-task":
		// Convert request to the expected type
		req, ok := request.(*types.SchedulerTaskRequest)
		if !ok {
			return nil, fmt.Errorf("invalid request type for submit-task: %T", request)
		}

		resp, err := h.dispatcher.SubmitTaskFromScheduler(ctx, req)
		if err != nil {
			return nil, err
		}
		return resp, nil

	default:
		return nil, fmt.Errorf("unknown method: %s", method)
	}
}

// GetMethods advertises the supported RPC methods for discovery/metrics
func (h *TaskDispatcherHandler) GetMethods() []rpcpkg.RPCMethod {
	return []rpcpkg.RPCMethod{
		{
			Name:         "submit-task",
			Description:  "Submit a task from schedulers to the dispatcher",
			RequestType:  &types.SchedulerTaskRequest{},
			ResponseType: &types.TaskManagerAPIResponse{},
			Timeout:      30 * time.Second,
		},
	}
}
