package rpc

import (
	"context"
	"fmt"
	"strconv"

	"github.com/trigg3rX/triggerx-backend/internal/health/config"
	"github.com/trigg3rX/triggerx-backend/internal/health/core/keeper"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	rpcpkg "github.com/trigg3rX/triggerx-backend/pkg/rpc"
	rpcserver "github.com/trigg3rX/triggerx-backend/pkg/rpc/server"
	rpctracing "github.com/trigg3rX/triggerx-backend/pkg/rpc/tracing"
)

// Server wraps the RPC server for health service
type Server struct {
	server *rpcserver.Server
	logger observability.Logger
}

// NewServer creates a new RPC server for health service
func NewServer(logger observability.Logger, tracer observability.Tracer, stateManager *keeper.StateManager, performerSelector *keeper.PerformerSelector) (*Server, error) {
	// Parse port from string to int
	port, err := strconv.Atoi(config.GetGRPCPort())
	if err != nil {
		return nil, fmt.Errorf("invalid port %q: %w", config.GetGRPCPort(), err)
	}

	// Create RPC server config
	serverConfig := rpcserver.Config{
		Name:        "health",
		Version:     config.GetVersion(),
		Address:     "0.0.0.0",
		Port:        port,
		Timeout:     30,
		MaxRequests: 1000,
		Metadata: map[string]string{
			"service": "health",
		},
	}

	// Create RPC server
	rpcSrv := rpcserver.NewServer(serverConfig, logger)

	// Add trace interceptor
	rpcSrv.AddInterceptor(rpctracing.TraceInterceptor(tracer, "health"))

	// Create and register handler
	handler := NewHandler(logger, tracer, stateManager, performerSelector)
	rpcSrv.RegisterHandler("health", handler)

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

// GetServiceInfo returns the service information
func (s *Server) GetServiceInfo() rpcpkg.ServiceInfo {
	return s.server.GetServiceInfo()
}
