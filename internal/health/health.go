package health

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"

	"github.com/trigg3rX/triggerx-backend/internal/health/api"
	"github.com/trigg3rX/triggerx-backend/internal/health/config"
	"github.com/trigg3rX/triggerx-backend/internal/health/keeper"
	"github.com/trigg3rX/triggerx-backend/internal/health/rewards"
	"github.com/trigg3rX/triggerx-backend/internal/health/rpc"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// stateManager is a package-level singleton for accessing the state manager
var stateManager *keeper.StateManager

// rewardsService is a package-level singleton for accessing the rewards service
var rewardsService *rewards.Service

// SetStateManager sets the global state manager instance
func SetStateManager(sm *keeper.StateManager) {
	stateManager = sm
}

// GetStateManager returns the global state manager instance
func GetStateManager() *keeper.StateManager {
	return stateManager
}

// SetRewardsService sets the global rewards service instance
func SetRewardsService(rs *rewards.Service) {
	rewardsService = rs
}

// GetRewardsService returns the global rewards service instance
func GetRewardsService() *rewards.Service {
	return rewardsService
}

// Service orchestrates the HTTP API and gRPC servers for the health service
type Service struct {
	httpServer *http.Server
	rpcServer  *rpc.Server
	logger     logging.Logger
	config     *Config
}

// Config holds configuration for the health service
type Config struct {
	HTTPPort string
	GRPCPort string
	GRPCHost string
}

// NewService creates a new health service orchestrator
func NewService(logger logging.Logger, cfg *Config) *Service {
	return &Service{
		logger: logger,
		config: cfg,
	}
}

// Start starts both the HTTP API and gRPC servers
func (s *Service) Start(ctx context.Context) error {
	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	// Start HTTP server
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := s.startHTTPServer(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("HTTP server error: %w", err)
		}
	}()

	// Start gRPC server
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := s.startRPCServer(); err != nil {
			errChan <- fmt.Errorf("gRPC server error: %w", err)
		}
	}()

	s.logger.Info("Health service started",
		"http_port", s.config.HTTPPort,
		"grpc_port", s.config.GRPCPort,
	)

	// Wait for any startup errors
	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

// Stop gracefully stops both servers
func (s *Service) Stop(ctx context.Context) error {
	s.logger.Info("Stopping health service...")

	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	// Stop HTTP server
	if s.httpServer != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := s.httpServer.Shutdown(ctx); err != nil {
				errChan <- fmt.Errorf("HTTP shutdown error: %w", err)
			}
		}()
	}

	// Stop gRPC server
	if s.rpcServer != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := s.rpcServer.Stop(ctx); err != nil {
				errChan <- fmt.Errorf("gRPC shutdown error: %w", err)
			}
		}()
	}

	wg.Wait()
	close(errChan)

	// Collect any errors
	for err := range errChan {
		s.logger.Error("Shutdown error", "error", err)
	}

	s.logger.Info("Health service stopped")
	return nil
}

// startHTTPServer initializes and starts the HTTP server
func (s *Service) startHTTPServer() error {
	if !config.IsDevMode() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())

	// Register HTTP routes with rewards service
	api.RegisterRoutes(router, s.logger, GetRewardsService())

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%s", s.config.HTTPPort),
		Handler: router,
	}

	s.logger.Info("Starting HTTP server", "port", s.config.HTTPPort)
	return s.httpServer.ListenAndServe()
}

// startRPCServer initializes and starts the gRPC server
func (s *Service) startRPCServer() error {
	// Import keeper state manager
	stateManager := GetStateManager()
	if stateManager == nil {
		return fmt.Errorf("state manager not initialized")
	}

	s.rpcServer = rpc.NewServer(stateManager, s.logger, s.config.GRPCHost, s.config.GRPCPort)

	s.logger.Info("Starting gRPC server",
		"host", s.config.GRPCHost,
		"port", s.config.GRPCPort,
	)
	return s.rpcServer.Start()
}
