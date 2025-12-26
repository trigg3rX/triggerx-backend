package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/config"
	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/api"
	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/rpc"
	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/service"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

const shutdownTimeout = 30 * time.Second

func main() {
	// Initialize configuration
	if err := config.Init("config/services/event-monitor.yaml"); err != nil {
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
	apiSrv := api.NewServer(config.GetEventMonitorRPCPort(), logger)
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
	logger.Info(ctx, "[3/3] Process: gRPC Server Started", observability.String("address", rpcSrv.GetServiceInfo().Address), observability.String("port", config.GetEventMonitorRPCPort()))

	// Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	sig := <-shutdown
	logger.Info(ctx, "Received shutdown signal", observability.String("signal", sig.String()))

	performGracefulShutdown(cancel, apiSrv, rpcSrv, svc, obs)
}

func performGracefulShutdown(
	cancel context.CancelFunc,
	apiSrv *api.Server,
	rpcSrv *rpc.Server,
	svc *service.Service,
	obs *observability.Observability,
) {
	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	// Cancel context to stop service and gRPC server
	cancel()

	// Stop service
	svc.Stop()
	log.Println("[1/3] Shutdown: Monitoring Service Stopped")

	// Shutdown gRPC server gracefully
	if err := rpcSrv.Stop(shutdownCtx); err != nil {
		log.Fatalf("gRPC server forced to shutdown: %v", err)
	}
	log.Println("[2/3] Shutdown: gRPC Server Stopped")

	// Shutdown API server gracefully
	if err := apiSrv.Stop(shutdownCtx); err != nil {
		log.Fatalf("API server forced to shutdown: %v", err)
	}
	log.Println("[3/3] Shutdown: API Server Stopped")

	// Shutdown observability (handles logger, tracer, metrics)
	if err := obs.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Error shutting down observability: %v", err)
	}
	log.Println("[4/4] Shutdown: Observability Shutdown Complete")

	log.Println("Service shutdown completed successfully")
}
