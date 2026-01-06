package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/api"
	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/config"
	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/rpc"
	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/service"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

func main() {
	// Initialize configuration
	configPath := "config/services/eventmonitor.yaml"
	if err := config.Init(configPath); err != nil {
		panic(fmt.Sprintf("Failed to initialize config: %v", err))
	}

	// Initialize observability (logger, tracer, metrics)
	obsCfg := observability.NewConfig(
		observability.EventMonitorService,
		config.GetVersion(),
		config.GetOTELExporterEndpoint(),
		config.IsDevMode(),
	)

	// Initialize observability (all three pillars)
	obs, err := observability.Initialize(obsCfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize observability: %v", err))
	}
	defer func() {
		if err := obs.Shutdown(context.Background()); err != nil {
			panic(fmt.Sprintf("Failed to shutdown observability: %v", err))
		}
	}()

	// Extract individual components
	logger := obs.Logger()
	tracer := obs.Tracer()
	obsMetrics := obs.Metrics()

	// Initialize application metrics
	metrics.InitializeMetrics(obsMetrics)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger.Info(ctx, "[1/3] Dependency: Observability Module Initialised")

	// Initialize service
	svc, err := service.NewService(ctx, logger, tracer)
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize service", observability.Error(err))
	}
	logger.Info(ctx, "[2/3] Dependency: Service Initialised")

	// Setup API server
	apiSrv := api.NewServer(config.GetHTTPPort())
	logger.Info(ctx, "[3/3] Dependency: API Server Initialised")

	// Setup gRPC server
	rpcSrv := rpc.NewServer(logger, tracer, svc.GetRegistryManager(), svc)
	logger.Info(ctx, "[3/3] Dependency: gRPC Server Initialised")

	metrics.StartMetricsCollection()
	logger.Info(ctx, "[1/3] Process: Metrics Collector Started")

	// Start service
	if err := svc.Start(); err != nil {
		logger.Fatal(ctx, "Failed to start service", observability.Error(err))
	}
	logger.Info(ctx, "[2/3] Process: Service Started")

	// Start gRPC server
	go func() {
		if err := rpcSrv.Start(ctx); err != nil {
			logger.Error(ctx, "gRPC server error", observability.Error(err))
		}
	}()
	logger.Info(ctx, "[3/3] Process: gRPC Server Started", observability.String("address", rpcSrv.GetServiceInfo().Address), observability.String("port", config.GetGRPCPort()))

	// Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	sig := <-shutdown
	logger.Info(ctx, "Received shutdown signal", observability.String("signal", sig.String()))

	performGracefulShutdown(ctx, logger, cancel, apiSrv, rpcSrv, svc, obs)
}

func performGracefulShutdown(
	ctx context.Context,
	logger observability.Logger,
	cancel context.CancelFunc,
	apiSrv *api.Server,
	rpcSrv *rpc.Server,
	svc *service.Service,
	obs *observability.Observability,
) {
	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), config.GetShutdownTimeout())
	defer shutdownCancel()

	// Cancel context to stop service and gRPC server
	cancel()

	// Stop service
	svc.Stop()
	logger.Info(ctx, "[1/4] Shutdown: Monitoring Service Stopped")

	// Shutdown gRPC server gracefully
	if err := rpcSrv.Stop(shutdownCtx); err != nil {
		logger.Error(ctx, "gRPC server forced to shutdown", observability.Error(err))
	} else {
		logger.Info(ctx, "[2/4] Shutdown: gRPC Server Stopped")
	}

	// Shutdown API server gracefully
	if err := apiSrv.Stop(shutdownCtx); err != nil {
		logger.Error(ctx, "API server forced to shutdown", observability.Error(err))
	} else {
		logger.Info(ctx, "[3/4] Shutdown: API Server Stopped")
	}

	// Shutdown observability (handles logger, tracer, metrics)
	if err := obs.Shutdown(shutdownCtx); err != nil {
		// Use fmt here since logger is being shut down
		fmt.Printf("Error shutting down observability: %v\n", err)
	} else {
		logger.Info(ctx, "[4/4] Shutdown: Observability Shutdown Complete")
	}

	logger.Info(ctx, "Service shutdown completed successfully")
}
