package rpc

import (
	"context"
	"fmt"
	"strconv"

	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/config"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	rpcserver "github.com/trigg3rX/triggerx-backend/pkg/rpc/server"
)

// StartRPCServer creates and starts the Task Monitor gRPC server using the generic approach.
// It registers the task monitor handler with the generic RPC server.
func StartRPCServer(ctx context.Context, logger observability.Logger, monitor TaskMonitorInterface, dbClient DatabaseClientInterface) (*rpcserver.Server, error) {
	port, err := strconv.Atoi(config.GetGRPCPort())
	if err != nil {
		return nil, fmt.Errorf("invalid port %q: %w", config.GetGRPCPort(), err)
	}

	srv := rpcserver.NewServer(rpcserver.Config{
		Name:    "task-monitor",
		Version: config.GetVersion(),
		Address: "0.0.0.0",
		Port:    port,
	}, logger)

	// Add useful middleware (logging)
	srv.AddInterceptor(rpcserver.LoggingInterceptor(logger))

	// Create and register the generic RPC handler
	handler := NewTaskMonitorHandler(logger, monitor, dbClient)
	srv.RegisterHandler("task-monitor", handler)

	if err := srv.Start(ctx); err != nil {
		return nil, err
	}
	return srv, nil
}
