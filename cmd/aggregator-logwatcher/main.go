package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/trigg3rX/triggerx-backend/internal/aggregatorlogwatcher/config"
	"github.com/trigg3rX/triggerx-backend/internal/aggregatorlogwatcher/core"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

func main() {
	// Initialize configuration
	err := config.Init()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize configuration: %v", err))
	}

	// Create observability configuration
	obsConfig := observability.NewConfig(
		observability.AggregatorService,
		"1.0.0",
		config.GetOTELExporterEndpoint(),
		config.IsDevMode(),
		config.GetServiceID(),
	)

	// Initialize observability
	obs, err := observability.Initialize(obsConfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize observability: %v", err))
	}
	defer func() {
		if err := obs.Shutdown(context.Background()); err != nil {
			panic(fmt.Sprintf("Failed to shutdown observability: %v", err))
		}
	}()

	logger := obs.Logger()
	tracer := obs.Tracer()

	ctx := context.Background()
	logger.Info(ctx, "Starting aggregator log watcher service...")

	// Create log watcher
	watcher, err := core.NewLogWatcher(ctx, logger, tracer)
	if err != nil {
		logger.Fatal(ctx, "Failed to create log watcher", observability.Error(err))
	}

	// Start watcher
	if err := watcher.Start(); err != nil {
		logger.Fatal(ctx, "Failed to start log watcher", observability.Error(err))
	}

	logger.Info(ctx, "Aggregator log watcher service started")

	// Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	<-shutdown
	logger.Info(ctx, "Received shutdown signal")

	// Stop watcher
	watcher.Stop()

	logger.Info(ctx, "Aggregator log watcher service stopped")
}
