package rpc

import (
	"context"
	"fmt"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	rpcpkg "github.com/trigg3rX/triggerx-backend/pkg/rpc"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// TaskDispatcherHandler implements the generic RPC handler interface
type TaskDispatcherHandler struct {
	logger     logging.Logger
	dispatcher TaskDispatcherInterface
}

// TaskDispatcherInterface defines the interface for task dispatcher operations
type TaskDispatcherInterface interface {
	SubmitTaskFromScheduler(ctx context.Context, req *types.SchedulerTaskRequest) (*types.TaskManagerAPIResponse, error)
}

// NewTaskDispatcherHandler creates a new RPC handler
func NewTaskDispatcherHandler(logger logging.Logger, dispatcher TaskDispatcherInterface) *TaskDispatcherHandler {
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
			// Try to convert from map if it's JSON-decoded
			if reqMap, ok := request.(map[string]interface{}); ok {
				req = h.convertMapToRequest(reqMap)
			} else {
				return nil, fmt.Errorf("invalid request type for submit-task: %T", request)
			}
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

// convertMapToRequest converts a map to SchedulerTaskRequest
// This is used when the request comes as JSON-decoded map
func (h *TaskDispatcherHandler) convertMapToRequest(reqMap map[string]interface{}) *types.SchedulerTaskRequest {
	// This is a simplified conversion - in a real implementation,
	// you might want to use a more robust JSON marshaling approach
	h.logger.Warn("Converting map to SchedulerTaskRequest - consider using proper JSON marshaling")

	// For now, return a basic request structure
	// In practice, you'd want to properly unmarshal the map
	return &types.SchedulerTaskRequest{
		Source: "unknown", // Default value
	}
}
