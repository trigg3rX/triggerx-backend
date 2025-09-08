package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/challenger"
	"github.com/trigg3rX/triggerx-backend/internal/challenger/config"
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
		ProcessName:   "challenger", // Use string directly instead of constant
		IsDevelopment: config.IsDevMode(),
	}

	logger, err := logging.NewZapLogger(logConfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	logger.Info("Starting challenger service...")

	// Create challenger instance
	challengerInstance := challenger.NewChallenger(logger)
	if challengerInstance == nil {
		logger.Error("Failed to create challenger instance")
		panic("Failed to create challenger instance")
	}

	// Start challenger service
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize service components
	var wg sync.WaitGroup
	serviceErrors := make(chan error, 1)

	// Start challenger service (event monitoring and challenge processing)
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("Starting challenger event monitoring...")
		if err := challengerInstance.Start(ctx); err != nil {
			if err != context.Canceled {
				serviceErrors <- fmt.Errorf("challenger service error: %v", err)
			}
		}
	}()

	logger.Info("Challenger service is ready")
	logger.Info("Event monitoring and challenge processing started")

	// Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serviceErrors:
		logger.Error("Service error received", "error", err)
	case sig := <-shutdown:
		logger.Info("Received shutdown signal", "signal", sig.String())
	}

	performGracefulShutdown(challengerInstance, &wg, logger, cancel)
}

func performGracefulShutdown(
	challengerInstance *challenger.Challenger,
	wg *sync.WaitGroup,
	logger logging.Logger,
	cancel context.CancelFunc,
) {
	logger.Info("Initiating graceful shutdown...")

	// Create shutdown context
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	// Cancel main context to stop challenger workers
	cancel()

	// Stop challenger service
	if challengerInstance != nil {
		logger.Info("Stopping challenger service...")
		challengerInstance.Close()
	}

	// Wait for all goroutines to finish
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("Challenger service stopped gracefully")
	case <-shutdownCtx.Done():
		logger.Warn("Shutdown timeout exceeded, forcing exit")
	}

	logger.Info("Shutdown complete")
}