package rpc

import (
	"context"
	"fmt"
	"strconv"

	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/config"
	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/registry"
	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/service"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	rpcpkg "github.com/trigg3rX/triggerx-backend/pkg/rpc"
	rpcserver "github.com/trigg3rX/triggerx-backend/pkg/rpc/server"
	rpctracing "github.com/trigg3rX/triggerx-backend/pkg/rpc/tracing"
)

// Server wraps the RPC server for event monitor
type Server struct {
	server *rpcserver.Server
	logger observability.Logger
}

// NewServer creates a new RPC server for event monitor
func NewServer(logger observability.Logger, tracer observability.Tracer, registryManager *registry.RegistryManager, svc *service.Service) (*Server, error) {
	// Parse port from string to int
	port, err := strconv.Atoi(config.GetGRPCPort())
	if err != nil {
		return nil, fmt.Errorf("invalid port %q: %w", config.GetGRPCPort(), err)
	}

	// Create RPC server config
	serverConfig := rpcserver.Config{
		Name:        "event-monitor",
		Version:     config.GetVersion(),
		Address:     "0.0.0.0",
		Port:        port,
		Timeout:     30,
		MaxRequests: 1000,
		Metadata: map[string]string{
			"service": "event-monitor",
		},
	}

	// Create RPC server
	rpcSrv := rpcserver.NewServer(serverConfig, logger)

	// Add trace interceptor
	rpcSrv.AddInterceptor(rpctracing.TraceInterceptor(tracer, "event-monitor"))

	// Create and register handler
	handler := NewHandler(logger, tracer, registryManager, svc)
	rpcSrv.RegisterHandler("event-monitor", handler)

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
