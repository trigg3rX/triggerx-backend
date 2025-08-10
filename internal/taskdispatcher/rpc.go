package taskdispatcher

import (
	"context"
	"fmt"
	"strconv"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	rpcpkg "github.com/trigg3rX/triggerx-backend/pkg/rpc"
	rpcserver "github.com/trigg3rX/triggerx-backend/pkg/rpc/server"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// DispatcherRPCHandler implements the generic RPC handler interface
// and routes the single supported method: "submit-task".
type DispatcherRPCHandler struct {
	logger     logging.Logger
	dispatcher *TaskDispatcher
}

// NewDispatcherRPCHandler constructs a new RPC handler using the provided dispatcher.
func NewDispatcherRPCHandler(logger logging.Logger, dispatcher *TaskDispatcher) *DispatcherRPCHandler {
	return &DispatcherRPCHandler{
		logger:     logger,
		dispatcher: dispatcher,
	}
}

// Handle routes incoming RPC requests based on the method name.
func (h *DispatcherRPCHandler) Handle(ctx context.Context, method string, request interface{}) (interface{}, error) {
	switch method {
	case "submit-task":
		// Be tolerant to different concrete types when decoding over RPC
		var req types.SchedulerTaskRequest
		req = request.(types.SchedulerTaskRequest)

		resp, err := h.dispatcher.SubmitTaskFromScheduler(ctx, &req)
		if err != nil {
			return nil, err
		}
		return resp, nil

	default:
		return nil, fmt.Errorf("unknown method: %s", method)
	}
}

// GetMethods advertises the supported RPC methods for discovery/metrics.
func (h *DispatcherRPCHandler) GetMethods() []rpcpkg.RPCMethod {
	return []rpcpkg.RPCMethod{
		{
			Name:        "submit-task",
			Description: "Submit a task from schedulers to the dispatcher",
		},
	}
}

// StartRPCServer creates and starts the Task Dispatcher RPC server using the shared server package.
// It registers the dispatcher handler under service name "TaskDispatcher".
func StartRPCServer(ctx context.Context, logger logging.Logger, dispatcher *TaskDispatcher, addr string, portStr string) (*rpcserver.Server, error) {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid port %q: %w", portStr, err)
	}

	srv := rpcserver.NewServer(rpcserver.Config{
		Name:    "TaskDispatcher",
		Version: "1.0.0",
		Address: addr,
		Port:    port,
	}, logger)

	// Add useful middleware (logging)
	srv.AddMiddleware(rpcserver.NewLoggingMiddleware(logger))

	// Register our single handler
	handler := NewDispatcherRPCHandler(logger, dispatcher)
	srv.RegisterHandler("TaskDispatcher", handler)

	if err := srv.Start(ctx); err != nil {
		return nil, err
	}
	return srv, nil
}
