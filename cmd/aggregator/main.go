package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/trigg3rX/triggerx-backend/internal/aggregator"
	"github.com/trigg3rX/triggerx-backend/internal/aggregator/api"
	"github.com/trigg3rX/triggerx-backend/internal/aggregator/config"
	"github.com/trigg3rX/triggerx-backend/internal/aggregator/rpc"
	"github.com/trigg3rX/triggerx-backend/internal/aggregator/types"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

const shutdownTimeout = 30 * time.Second

func main() {
	// Initialize configuration
	if err := config.Init(); err != nil {
		panic(fmt.Sprintf("Failed to initialize config: %v", err))
	}

	// Initialize logger
	logConfig := logging.LoggerConfig{
		ProcessName:   logging.AggregatorProcess,
		IsDevelopment: config.IsDevMode(),
	}

	logger, err := logging.NewZapLogger(logConfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	logger.Info("Starting aggregator service...")

	// Initialize aggregator configuration
	aggregatorConfig := &types.AggregatorConfig{
		MaxConcurrentTasks: 100,
		DefaultTimeout:     5 * time.Minute,
		MinOperators:       1,
		MaxOperators:       100,
	}

	// Create aggregator instance
	aggregatorInstance := aggregator.NewAggregator(logger, aggregatorConfig)

	// Start aggregator service
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := aggregatorInstance.Start(ctx); err != nil {
		logger.Error("Failed to start aggregator service", "error", err)
		panic(fmt.Sprintf("Failed to start aggregator service: %v", err))
	}

	// Initialize server components
	var wg sync.WaitGroup
	serverErrors := make(chan error, 3)
	ready := make(chan struct{})

	// Setup HTTP server
	httpServer := setupHTTPServer(aggregatorInstance, logger)

	// Setup RPC server
	rpcServerAddr := fmt.Sprintf(":%s", config.GetAggregatorP2PPort())
	rpcServer := rpc.NewRPCServer(aggregatorInstance, logger, rpcServerAddr)

	// Start HTTP server
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("Starting HTTP API server...")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- fmt.Errorf("HTTP server error: %v", err)
		}
	}()

	// Start RPC server
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("Starting RPC server...")
		if err := rpcServer.Start(ctx); err != nil {
			serverErrors <- fmt.Errorf("RPC server error: %v", err)
		}
	}()

	close(ready)
	logger.Infof("Aggregator service is ready")
	logger.Infof("HTTP API server running on port %s", config.GetAggregatorRPCPort())
	logger.Infof("RPC server running on port %s", config.GetAggregatorP2PPort())

	// Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		logger.Error("Server error received", "error", err)
	case sig := <-shutdown:
		logger.Info("Received shutdown signal", "signal", sig.String())
	}

	performGracefulShutdown(httpServer, rpcServer, aggregatorInstance, &wg, logger, cancel)
}

func setupHTTPServer(aggregatorInstance *aggregator.Aggregator, logger logging.Logger) *http.Server {
	if !config.IsDevMode() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(api.LoggerMiddleware(logger))

	// Register routes with aggregator instance
	api.RegisterRoutesWithAggregator(router, aggregatorInstance, logger)

	return &http.Server{
		Addr:         fmt.Sprintf(":%s", config.GetAggregatorRPCPort()),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

func performGracefulShutdown(
	httpServer *http.Server,
	rpcServer *rpc.RPCServer,
	aggregatorInstance *aggregator.Aggregator,
	wg *sync.WaitGroup,
	logger logging.Logger,
	cancel context.CancelFunc,
) {
	logger.Info("Initiating graceful shutdown...")

	// Create shutdown context
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	// Cancel main context to stop aggregator workers
	cancel()

	// Shutdown RPC server
	if rpcServer != nil {
		logger.Info("Shutting down RPC server...")
		if err := rpcServer.Stop(shutdownCtx); err != nil {
			logger.Error("RPC server shutdown error", "error", err)
		}
	}

	// Shutdown HTTP server
	if httpServer != nil {
		logger.Info("Shutting down HTTP server...")
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			logger.Error("HTTP server shutdown error", "error", err)
			if err := httpServer.Close(); err != nil {
				logger.Error("Forced HTTP server close error", "error", err)
			}
		}
	}

	// Stop aggregator service
	if aggregatorInstance != nil {
		logger.Info("Stopping aggregator service...")
		if err := aggregatorInstance.Stop(); err != nil {
			logger.Error("Aggregator shutdown error", "error", err)
		}
	}

	// Wait for all goroutines to finish
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("All servers stopped gracefully")
	case <-shutdownCtx.Done():
		logger.Warn("Shutdown timeout exceeded, forcing exit")
	}

	logger.Info("Shutdown complete")
}
