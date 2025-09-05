package server

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
	rpcpkg "github.com/trigg3rX/triggerx-backend/pkg/rpc"
	rpcproto "github.com/trigg3rX/triggerx-backend/pkg/rpc/proto"
)

// Server represents a gRPC server
type Server struct {
	config       Config
	logger       logging.Logger
	handlers     map[string]rpcpkg.RPCHandler
	interceptors []grpc.UnaryServerInterceptor
	registry     rpcpkg.ServiceRegistry

	// Server state
	grpcServer *grpc.Server
	listener   net.Listener
	isRunning  bool
	mu         sync.RWMutex

	// Service info
	serviceInfo rpcpkg.ServiceInfo
}

// Config holds server configuration
type Config struct {
	Name        string
	Version     string
	Address     string
	Port        int
	Timeout     time.Duration
	MaxRequests int
	Metadata    map[string]string
}

// NewServer creates a new gRPC server
func NewServer(config Config, logger logging.Logger) *Server {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxRequests == 0 {
		config.MaxRequests = 1000
	}

	return &Server{
		config:   config,
		logger:   logger,
		handlers: make(map[string]rpcpkg.RPCHandler),
		serviceInfo: rpcpkg.ServiceInfo{
			Name:     config.Name,
			Version:  config.Version,
			Address:  config.Address,
			Port:     config.Port,
			Metadata: config.Metadata,
		},
	}
}

// RegisterHandler registers an RPC handler for a service
func (s *Server) RegisterHandler(serviceName string, handler rpcpkg.RPCHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[serviceName] = handler
}

// AddInterceptor adds a gRPC interceptor to the server
func (s *Server) AddInterceptor(interceptor grpc.UnaryServerInterceptor) {
	s.interceptors = append(s.interceptors, interceptor)
}

// SetRegistry sets the service registry
func (s *Server) SetRegistry(registry rpcpkg.ServiceRegistry) {
	s.registry = registry
}

// Start starts the gRPC server
func (s *Server) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		return fmt.Errorf("server is already running")
	}

	// Create gRPC server with interceptors
	s.grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(s.chainUnaryInterceptors()),
		grpc.MaxConcurrentStreams(uint32(s.config.MaxRequests)),
	)

	// Register handlers
	for serviceName, handler := range s.handlers {
		genericService := NewGenericService(serviceName, handler, s.logger)
		rpcproto.RegisterGenericServiceServer(s.grpcServer, genericService)
		s.logger.Info("Registered gRPC handler", "service", serviceName)
	}

	// Enable reflection for debugging
	reflection.Register(s.grpcServer)

	// Create listener
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.config.Address, s.config.Port))
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}
	s.listener = listener

	// Update service info
	s.serviceInfo.Address = listener.Addr().String()
	s.serviceInfo.LastSeen = time.Now()

	// Register with service registry if available (with retry)
	if s.registry != nil {
		retryCfg := retry.DefaultRetryConfig()
		retryCfg.MaxRetries = 5
		retryCfg.InitialDelay = 500 * time.Millisecond
		retryCfg.MaxDelay = 5 * time.Second
		retryCfg.BackoffFactor = 2.0
		retryCfg.JitterFactor = 0.2
		retryCfg.LogRetryAttempt = true
		retryCfg.ShouldRetry = func(err error, attempt int) bool { return err != nil }

		if err := retry.RetryFunc(ctx, func() error {
			return s.registry.Register(ctx, s.serviceInfo)
		}, retryCfg, s.logger); err != nil {
			s.logger.Warn("Failed to register with service registry after retries", "error", err)
		}
	}

	// Start server
	s.isRunning = true
	s.logger.Info("Starting gRPC server",
		"address", s.listener.Addr().String(),
		"services", len(s.handlers))

	go func() {
		if err := s.grpcServer.Serve(listener); err != nil {
			s.logger.Error("gRPC server error", "error", err)
		}
	}()

	return nil
}

// Stop gracefully stops the server
func (s *Server) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return nil
	}

	s.logger.Info("Stopping gRPC server")

	// Deregister from service registry (with retry)
	if s.registry != nil {
		retryCfg := retry.DefaultRetryConfig()
		retryCfg.MaxRetries = 3
		retryCfg.InitialDelay = 300 * time.Millisecond
		retryCfg.MaxDelay = 3 * time.Second
		retryCfg.BackoffFactor = 2.0
		retryCfg.JitterFactor = 0.2
		retryCfg.LogRetryAttempt = false
		retryCfg.ShouldRetry = func(err error, attempt int) bool { return err != nil }

		if err := retry.RetryFunc(ctx, func() error {
			return s.registry.Deregister(ctx, s.config.Name)
		}, retryCfg, s.logger); err != nil {
			s.logger.Warn("Failed to deregister from service registry after retries", "error", err)
		}
	}

	// Stop gRPC server
	if s.grpcServer != nil {
		// Graceful shutdown
		s.grpcServer.GracefulStop()
	}

	s.isRunning = false
	s.logger.Info("gRPC server stopped")
	return nil
}

// chainUnaryInterceptors chains all unary interceptors
func (s *Server) chainUnaryInterceptors() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Apply interceptors in order
		for i := len(s.interceptors) - 1; i >= 0; i-- {
			handler = s.wrapUnaryHandler(s.interceptors[i], handler)
		}
		return handler(ctx, req)
	}
}

// wrapUnaryHandler wraps a unary handler with an interceptor
func (s *Server) wrapUnaryHandler(interceptor grpc.UnaryServerInterceptor, handler grpc.UnaryHandler) grpc.UnaryHandler {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return interceptor(ctx, req, &grpc.UnaryServerInfo{}, handler)
	}
}

// GetServiceInfo returns the service information
func (s *Server) GetServiceInfo() rpcpkg.ServiceInfo {
	return s.serviceInfo
}

// IsRunning returns whether the server is running
func (s *Server) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isRunning
}
