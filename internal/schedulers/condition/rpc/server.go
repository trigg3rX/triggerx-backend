package rpc

import (
	"context"
	"fmt"
	"strconv"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/config"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/core/scheduler"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	rpcserver "github.com/trigg3rX/triggerx-backend/pkg/rpc/server"
	rpctracing "github.com/trigg3rX/triggerx-backend/pkg/rpc/tracing"
)

// Server wraps the RPC server for condition scheduler
type Server struct {
	server *rpcserver.Server
	logger observability.Logger
}

// Dependencies contains dependencies for the RPC server
type Dependencies struct {
	Scheduler *scheduler.ConditionBasedScheduler
}

// NewServer creates a new RPC server for condition scheduler
func NewServer(logger observability.Logger, tracer observability.Tracer, deps *Dependencies) (*Server, error) {
	// Parse port from string to int
	port, err := strconv.Atoi(config.GetGRPCPort())
	if err != nil {
		return nil, fmt.Errorf("invalid port %q: %w", config.GetGRPCPort(), err)
	}

	// Create RPC server config
	serverConfig := rpcserver.Config{
		Name:    config.GetServiceName(),
		Version: config.GetVersion(),
		Address: "0.0.0.0",
		Port:    port,
	}

	// Create RPC server
	rpcSrv := rpcserver.NewServer(serverConfig, logger)

	// Add trace interceptor
	rpcSrv.AddInterceptor(rpctracing.TraceInterceptor(tracer, config.GetServiceName()))

	// Create and register handler
	handler := NewHandler(logger, tracer, deps.Scheduler)
	rpcSrv.RegisterHandler("condition-scheduler", handler)

	return &Server{
		server: rpcSrv,
		logger: logger,
	}, nil
}

// Start starts the RPC server
func (s *Server) Start(ctx context.Context) error {
	return s.server.Start(ctx)
}

// Stop stops the RPC server
func (s *Server) Stop(ctx context.Context) error {
	return s.server.Stop(ctx)
}
