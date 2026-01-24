package rpc

import (
	"context"
	"fmt"
	"strconv"

	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/config"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	rpcserver "github.com/trigg3rX/triggerx-backend/pkg/rpc/server"
	rpctracing "github.com/trigg3rX/triggerx-backend/pkg/rpc/tracing"
)

// Server represents the gRPC server
type Server struct {
	server *rpcserver.Server
	logger observability.Logger
}

// Dependencies contains dependencies for the RPC server
type Dependencies struct {
	Monitor  TaskMonitorInterface
	DBClient DatabaseClientInterface
}

// NewServer creates a new gRPC server
func NewServer(logger observability.Logger, tracer observability.Tracer, deps *Dependencies) (*Server, error) {
	port, err := strconv.Atoi(config.GetGRPCPort())
	if err != nil {
		return nil, fmt.Errorf("invalid port %q: %w", config.GetGRPCPort(), err)
	}

	serverConfig := rpcserver.Config{
		Name:    config.GetServiceName(),
		Version: config.GetVersion(),
		Address: "0.0.0.0",
		Port:    port,
	}

	rpcSrv := rpcserver.NewServer(serverConfig, logger)
	rpcSrv.AddInterceptor(rpcserver.LoggingInterceptor(logger))
	rpcSrv.AddInterceptor(rpctracing.TraceInterceptor(tracer, config.GetServiceName()))

	handler := NewTaskMonitorHandler(logger, deps.Monitor, deps.DBClient)
	rpcSrv.RegisterHandler("task-monitor", handler)

	return &Server{
		server: rpcSrv,
		logger: logger,
	}, nil
}

// Start starts the gRPC server
func (s *Server) Start(ctx context.Context) error {
	return s.server.Start(ctx)
}

// Stop stops the gRPC server gracefully
func (s *Server) Stop(ctx context.Context) error {
	return s.server.Stop(ctx)
}
