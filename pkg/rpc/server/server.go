package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"sync"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
	rpcpkg "github.com/trigg3rX/triggerx-backend/pkg/rpc"
)

// Server represents a unified RPC server
type Server struct {
	config     Config
	logger     logging.Logger
	handlers   map[string]rpcpkg.RPCHandler
	middleware []rpcpkg.Middleware
	registry   rpcpkg.ServiceRegistry

	// Server state
	httpServer *http.Server
	rpcServer  *rpc.Server
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

// NewServer creates a new RPC server
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

// AddMiddleware adds middleware to the server
func (s *Server) AddMiddleware(middleware rpcpkg.Middleware) {
	s.middleware = append(s.middleware, middleware)
}

// SetRegistry sets the service registry
func (s *Server) SetRegistry(registry rpcpkg.ServiceRegistry) {
	s.registry = registry
}

// Start starts the RPC server
func (s *Server) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		return fmt.Errorf("server is already running")
	}

	// Create RPC server
	s.rpcServer = rpc.NewServer()

	// Register handlers
	for serviceName, handler := range s.handlers {
		wrappedHandler := s.wrapHandler(serviceName, handler)
		if err := s.rpcServer.RegisterName(serviceName, wrappedHandler); err != nil {
			return fmt.Errorf("failed to register handler %s: %w", serviceName, err)
		}
		s.logger.Info("Registered RPC handler", "service", serviceName)
	}

	// Setup HTTP handler
	rpc.HandleHTTP()

	// Create listener
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.config.Address, s.config.Port))
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}
	s.listener = listener

	// Create HTTP server
	s.httpServer = &http.Server{
		Addr:         listener.Addr().String(),
		Handler:      http.DefaultServeMux,
		ReadTimeout:  s.config.Timeout,
		WriteTimeout: s.config.Timeout,
	}

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
		// For registry operations, we generally want to retry on any error
		retryCfg.ShouldRetry = func(err error) bool { return err != nil }

		if err := retry.RetryFunc(ctx, func() error {
			return s.registry.Register(ctx, s.serviceInfo)
		}, retryCfg, s.logger); err != nil {
			s.logger.Warn("Failed to register with service registry after retries", "error", err)
		}
	}

	// Start server
	s.isRunning = true
	s.logger.Info("Starting RPC server",
		"address", s.httpServer.Addr,
		"services", len(s.handlers))

	go func() {
		if err := s.httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			s.logger.Error("RPC server error", "error", err)
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

	s.logger.Info("Stopping RPC server")

	// Deregister from service registry (with retry)
	if s.registry != nil {
		retryCfg := retry.DefaultRetryConfig()
		retryCfg.MaxRetries = 3
		retryCfg.InitialDelay = 300 * time.Millisecond
		retryCfg.MaxDelay = 3 * time.Second
		retryCfg.BackoffFactor = 2.0
		retryCfg.JitterFactor = 0.2
		retryCfg.LogRetryAttempt = false
		retryCfg.ShouldRetry = func(err error) bool { return err != nil }

		if err := retry.RetryFunc(ctx, func() error {
			return s.registry.Deregister(ctx, s.config.Name)
		}, retryCfg, s.logger); err != nil {
			s.logger.Warn("Failed to deregister from service registry after retries", "error", err)
		}
	}

	// Stop HTTP server
	if s.httpServer != nil {
		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			s.logger.Error("Error during server shutdown", "error", err)
			return err
		}
	}

	s.isRunning = false
	s.logger.Info("RPC server stopped")
	return nil
}

// wrapHandler wraps a handler with middleware
func (s *Server) wrapHandler(serviceName string, handler rpcpkg.RPCHandler) interface{} {
	return &HandlerWrapper{
		serviceName: serviceName,
		handler:     handler,
		middleware:  s.middleware,
		logger:      s.logger,
	}
}

// HandlerWrapper wraps an RPC handler with middleware
type HandlerWrapper struct {
	serviceName string
	handler     rpcpkg.RPCHandler
	middleware  []rpcpkg.Middleware
	logger      logging.Logger
}

// Call implements the RPC call interface for Go's built-in RPC system
func (h *HandlerWrapper) Call(args *rpcpkg.RPCArgs, reply *rpcpkg.RPCReply) error {
	// Apply middleware chain
	var next rpcpkg.RPCHandler = h.handler
	for i := len(h.middleware) - 1; i >= 0; i-- {
		middleware := h.middleware[i]
		next = &middlewareWrapper{
			middleware: middleware,
			next:       next,
		}
	}

	// Create context from args if available
	ctx := context.Background()
	if args.Context != nil {
		ctx = args.Context
	}

	result, err := next.Handle(ctx, args.Method, args.Request)
	if err != nil {
		return err
	}

	// Set the result in the reply
	reply.Result = result
	return nil
}

// middlewareWrapper wraps middleware for chaining
type middlewareWrapper struct {
	middleware rpcpkg.Middleware
	next       rpcpkg.RPCHandler
}

func (m *middlewareWrapper) Handle(ctx context.Context, method string, request interface{}) (interface{}, error) {
	return m.middleware.Process(ctx, method, request, m.next)
}

func (m *middlewareWrapper) GetMethods() []rpcpkg.RPCMethod {
	return m.next.GetMethods()
}
