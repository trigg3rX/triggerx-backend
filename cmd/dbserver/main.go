package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/api"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/api/handlers"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/api/middleware"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/config"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/database"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/redis"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/database/repository"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/rpc/clients/conditionscheduler"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/websocket"
	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

func main() {
	// 1. Initialize configuration
	configPath := "config/services/dbserver.yaml"
	if err := config.Init(configPath); err != nil {
		panic(fmt.Sprintf("Failed to initialize config: %v", err))
	}

	// 2. Initialize observability (logger, tracer, metrics)
	obsCfg := observability.NewConfig(
		observability.ServerService,
		config.GetVersion(),
		config.GetOTELExporterEndpoint(),
		config.IsDevMode(),
		config.GetServiceID(),
	)

	obs, err := observability.Initialize(obsCfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize observability: %v", err))
	}
	defer func() {
		if err := obs.Shutdown(context.Background()); err != nil {
			panic(fmt.Sprintf("Failed to shutdown observability: %v", err))
		}
	}()

	// Extract components
	logger := obs.Logger()
	tracer := obs.Tracer()
	obsMetrics := obs.Metrics()

	ctx := context.Background()

	// Initialize application metrics
	metrics.InitializeMetrics(obsMetrics)
	logger.Info(ctx, "[1/7] Dependency: Observability Module Initialised")

	// Initialize database connection
	dbConn, err := database.NewConnection(logger)
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize database connection", observability.Error(err))
	}
	defer dbConn.Close()

	mainSession := dbConn.Session()
	if mainSession == nil {
		logger.Fatal(ctx, "Database session cannot be nil")
	}
	logger.Info(ctx, "[2/7] Dependency: Database Connection Initialised")

	// Initialize Redis client
	redisClient, err := redis.NewClient(logger)
	if err != nil {
		logger.Error(ctx, "Failed to initialize Redis client", observability.Error(err))
		redisClient = nil // Continue without Redis
	} else {
		logger.Info(ctx, "[3/7] Dependency: Redis Client Initialised")
	}

	// Initialize rate limiter (requires Redis)
	var rateLimiter *middleware.RateLimiter
	if redisClient != nil {
		var err error
		rateLimiter, err = middleware.NewRateLimiterWithClient(redisClient, logger)
		if err != nil {
			logger.Error(ctx, "Failed to initialize rate limiter", observability.Error(err))
		}
	} else {
		logger.Warn(ctx, "Rate limiter disabled - Redis client not available")
	}

	// Initialize Docker executor
	dockerExecutor, err := dockerexecutor.NewDockerExecutorFromFile("config/services/docker-executor.yaml", logger)
	if err != nil {
		logger.Error(ctx, "Failed to create Docker executor", observability.Error(err))
		dockerExecutor = nil
	} else {
		// Initialize Docker executor with language-specific pools
		if err := dockerExecutor.Initialize(context.Background()); err != nil {
			logger.Error(ctx, "Failed to initialize Docker executor", observability.Error(err))
		} else {
			logger.Info(ctx, "[4/7] Dependency: Docker Executor Initialised")
		}
	}

	// Initialize condition scheduler gRPC client
	conditionSchedulerClient, err := conditionscheduler.NewClient(
		config.GetConditionSchedulerRPCUrl(),
		logger,
		tracer,
	)
	if err != nil {
		logger.Fatal(ctx, "Failed to create condition scheduler gRPC client", observability.Error(err))
	}
	logger.Info(ctx, "[5/7] Dependency: Condition Scheduler Client Initialised")

	// Initialize WebSocket components
	hub := websocket.NewHub(logger)

	// Create the task repository with publisher for WebSocket events
	taskRepo := repository.NewTaskRepositoryWithPublisher(dbConn, nil) // publisher will be set later if needed

	// Create and set the initial data handler for the hub
	initialDataHandler := handlers.NewInitialDataHandler(taskRepo, logger)
	hub.SetInitialDataCallback(initialDataHandler.HandleInitialData)

	// Initialize API key auth (needed for WebSocket auth)
	apiKeyAuth := middleware.NewApiKeyAuth(dbConn, rateLimiter, logger)

	wsConnectionManager := websocket.NewWebSocketConnectionManager(
		websocket.NewWebSocketUpgrader(logger),
		websocket.NewWebSocketAuthMiddleware(apiKeyAuth, logger),
		websocket.NewWebSocketRateLimiter(rateLimiter, 100, logger), // Max 100 connections per IP
		hub,
		logger,
	)

	// Start WebSocket hub
	go hub.Run(ctx)
	logger.Info(ctx, "[6/7] Dependency: WebSocket Hub Initialised")

	// Initialize validator
	validator := middleware.NewValidator(ctx, logger)

	// Initialize API server
	apiCfg := api.Config{
		Port:           config.GetHTTPPort(),
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	apiDeps := &api.Dependencies{
		Logger:                   logger,
		Tracer:                   tracer,
		Metrics:                  obsMetrics,
		DB:                       dbConn,
		RedisClient:              redisClient,
		RateLimiter:              rateLimiter,
		ApiKeyAuth:               apiKeyAuth,
		Validator:                validator,
		Hub:                      hub,
		WSConnectionManager:      wsConnectionManager,
		ConditionSchedulerClient: conditionSchedulerClient,
		DockerExecutor:           dockerExecutor,
	}

	apiSrv := api.NewServer(apiCfg, apiDeps)
	logger.Info(ctx, "[7/7] Dependency: API Server Initialised")

	// Start metrics collector
	collector := metrics.NewCollector(obsMetrics, logger)
	collector.Start()
	logger.Info(ctx, "[1/2] Process: Metrics Collector Started")

	// Start all servers and background processes
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start API server
	go func() {
		if err := apiSrv.Start(); err != nil {
			logger.Error(ctx, "API server error", observability.Error(err))
		}
	}()
	logger.Info(ctx, "[2/2] Process: API Server Started", observability.String("port", config.GetHTTPPort()))

	// 14. Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	sig := <-shutdown
	logger.Info(ctx, "Received shutdown signal", observability.String("signal", sig.String()))

	performGracefulShutdown(ctx, apiSrv, dockerExecutor, hub, obs, logger)
}

func performGracefulShutdown(
	ctx context.Context,
	apiSrv *api.Server,
	dockerExecutor dockerexecutor.DockerExecutorAPI,
	hub *websocket.Hub,
	obs *observability.Observability,
	logger observability.Logger,
) {
	shutdownCtx, cancel := context.WithTimeout(ctx, config.GetShutdownTimeout())
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)

		// Stop WebSocket hub
		if hub != nil {
			hub.Shutdown(shutdownCtx)
		}
		logger.Info(ctx, "[1/4] Shutdown: WebSocket Hub Stopped")

		// Stop API server
		if err := apiSrv.Stop(shutdownCtx); err != nil {
			logger.Error(shutdownCtx, "API server shutdown error", observability.Error(err))
		}
		logger.Info(ctx, "[2/4] Shutdown: API Server Stopped")

		// Close Docker executor
		if dockerExecutor != nil {
			if err := dockerExecutor.Close(shutdownCtx); err != nil {
				logger.Error(shutdownCtx, "Failed to close Docker executor", observability.Error(err))
			}
		}
		logger.Info(ctx, "[3/4] Shutdown: Docker Executor Closed")

		logger.Info(ctx, "Graceful shutdown completed successfully")

		// Shutdown observability
		if err := obs.Shutdown(shutdownCtx); err != nil {
			logger.Error(shutdownCtx, "Observability shutdown error", observability.Error(err))
		}
		logger.Info(ctx, "[4/4] Shutdown: Observability Shutdown Complete")
	}()

	select {
	case <-done:
		// Shutdown completed successfully
	case <-shutdownCtx.Done():
		logger.Warn(ctx, "Shutdown timeout reached, forcing exit")
	}
	os.Exit(0)
}
