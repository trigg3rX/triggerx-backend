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

	"github.com/trigg3rX/triggerx-backend/internal/manager"
	// "github.com/trigg3rX/triggerx-backend/internal/manager/cache"
	"github.com/trigg3rX/triggerx-backend/internal/manager/client/aggregator"
	"github.com/trigg3rX/triggerx-backend/internal/manager/client/database"
	"github.com/trigg3rX/triggerx-backend/internal/manager/config"
	"github.com/trigg3rX/triggerx-backend/internal/manager/scheduler"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

const (
	shutdownTimeout = 30 * time.Second
	defaultTimeout  = 10 * time.Second
)

func main() {
	// Initialize configuration
	if err := config.Init(); err != nil {
		panic(fmt.Sprintf("Failed to initialize config: %v", err))
	}

	// Initialize logger
	logConfig := logging.LoggerConfig{
		LogDir:          logging.BaseDataDir,
		ProcessName:     logging.ManagerProcess,
		Environment:     getEnvironment(),
		UseColors:       true,
		MinStdoutLevel:  getLogLevel(),
		MinFileLogLevel: getLogLevel(),
	}

	if err := logging.InitServiceLogger(logConfig); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	logger := logging.GetServiceLogger()

	logger.Info("Starting manager service...",
		"mode", getEnvironment(),
		"port", config.GetManagerRPCPort(),
	)

	// Initialize database client
	dbConfig := database.DatabaseClientConfig{
		RPCAddress:  config.GetDatabaseRPCAddress(),
		HTTPTimeout: defaultTimeout,
	}
	dbClient, err := database.NewDatabaseClient(logger, dbConfig)
	if err != nil {
		logger.Fatal("Failed to initialize database client:", err)
	}
	defer dbClient.Close()

	// Initialize aggregator client
	aggregatorConfig := aggregator.AggregatorClientConfig{
		RPCAddress: config.GetAggregatorRPCAddress(),
		PrivateKey: config.GetDeployerPrivateKey(),
		RPCTimeout: defaultTimeout,
	}
	aggregatorClient, err := aggregator.NewAggregatorClient(logger, aggregatorConfig)
	if err != nil {
		logger.Fatal("Failed to initialize aggregator client:", err)
	}
	defer aggregatorClient.Close()

	// Initialize job scheduler
	// jobCache := cache.NewCache()
	// jobScheduler, err := scheduler.NewJobScheduler(logger, jobCache, dbClient, aggregatorClient)
	jobScheduler, err := scheduler.NewJobScheduler(logger, dbClient, aggregatorClient)
	if err != nil {
		logger.Fatal("Failed to initialize job scheduler:", err)
	}

	var wg sync.WaitGroup
	serverErrors := make(chan error, 1)

	ready := make(chan struct{})

	// Setup HTTP server
	srv := setupHTTPServer(logger, jobScheduler)

	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("Starting HTTP server...")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- fmt.Errorf("HTTP server error: %v", err)
		}
	}()

	close(ready)
	logger.Infof("Manager Server initialized, starting on port %s...", config.GetManagerRPCPort())

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		logger.Error("Server error received", "error", err)
	case sig := <-shutdown:
		logger.Info("Received shutdown signal", "signal", sig.String())
	}

	performGracefulShutdown(srv, &wg, logger)
}

func getEnvironment() logging.LogLevel {
	if config.IsDevMode() {
		return logging.Development
	}
	return logging.Production
}

func getLogLevel() logging.Level {
	if config.IsDevMode() {
		return logging.DebugLevel
	}
	return logging.InfoLevel
}

func setupHTTPServer(logger logging.Logger, jobScheduler *scheduler.JobScheduler) *http.Server {
	if !config.IsDevMode() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(manager.LoggerMiddleware(logger))

	manager.RegisterRoutes(router, jobScheduler)

	return &http.Server{
		Addr:    fmt.Sprintf(":%s", config.GetManagerRPCPort()),
		Handler: router,
	}
}

func performGracefulShutdown(srv *http.Server, wg *sync.WaitGroup, logger logging.Logger) {
	logger.Info("Initiating graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("HTTP server shutdown error", "error", err)
		if err := srv.Close(); err != nil {
			logger.Error("Forced HTTP server close error", "error", err)
		}
	}

	// Ensure logger is properly shutdown
	if err := logging.Shutdown(); err != nil {
		fmt.Printf("Error shutting down logger: %v\n", err)
	}

	wg.Wait()
	logger.Info("Shutdown complete")
}
