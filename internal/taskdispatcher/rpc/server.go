package rpc

import (
	"context"
	"fmt"
	"strconv"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	rpcserver "github.com/trigg3rX/triggerx-backend/pkg/rpc/server"
)

// StartRPCServer creates and starts the Task Dispatcher gRPC server using the generic approach.
// It registers the task dispatcher handler with the generic RPC server.
func StartRPCServer(ctx context.Context, logger logging.Logger, dispatcher TaskDispatcherInterface, addr string, portStr string) (*rpcserver.Server, error) {
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
	srv.AddInterceptor(rpcserver.LoggingInterceptor(logger))

	// Create and register the generic RPC handler
	handler := NewTaskDispatcherHandler(logger, dispatcher)
	srv.RegisterHandler("TaskDispatcher", handler)

	if err := srv.Start(ctx); err != nil {
		return nil, err
	}
	return srv, nil
}
