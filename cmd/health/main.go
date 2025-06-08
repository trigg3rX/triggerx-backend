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

	"github.com/trigg3rX/triggerx-backend/internal/health"
	"github.com/trigg3rX/triggerx-backend/internal/health/client"
	"github.com/trigg3rX/triggerx-backend/internal/health/config"
	"github.com/trigg3rX/triggerx-backend/internal/health/keeper"
	"github.com/trigg3rX/triggerx-backend/internal/health/telegram"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
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
		ProcessName:     logging.HealthProcess,
		IsDevelopment:   config.IsDevMode(),
	}

	logger, err := logging.NewZapLogger(logConfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	logger.Info("Starting health service...")

	// Initialize server components
	var wg sync.WaitGroup
	serverErrors := make(chan error, 3)
	ready := make(chan struct{})

	// Initialize database connection
	dbConfig := &database.Config{
		Hosts:    []string{config.GetDatabaseHostAddress() + ":" + config.GetDatabaseHostPort()},
		Keyspace: "triggerx",
	}
	dbConn, err := database.NewConnection(dbConfig, logger)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize database connection: %v", err))
	}

	// Initialize Telegram bot
	telegramBot, err := telegram.NewBot(config.GetBotToken(), logger, dbConn)
	if err != nil {
		logger.Errorf("Failed to initialize Telegram bot: %v", err)
	}

	// Initialize database manager
	client.InitDatabaseManager(logger, dbConn, telegramBot)
	logger.Info("Database manager initialized")

	// Initialize state manager
	stateManager := keeper.InitializeStateManager(logger)
	logger.Info("Keeper state manager initialized")

	// Load verified keepers from database
	if err := stateManager.LoadVerifiedKeepers(); err != nil {
		logger.Debug("Failed to load verified keepers from database", "error", err)
		// Continue anyway, as we can still operate with an empty state
	}

	// Setup HTTP server
	srv := setupHTTPServer(logger)

	// Start server
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("Starting HTTP server...")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- fmt.Errorf("HTTP server error: %v", err)
		}
	}()

	close(ready)
	logger.Infof("Health service is ready on port %s", config.GetHealthRPCPort())

	// Handle graceful shutdown
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

func setupHTTPServer(logger logging.Logger) *http.Server {
	if !config.IsDevMode() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(health.LoggerMiddleware(logger))

	// Register routes
	health.RegisterRoutes(router, logger)

	return &http.Server{
		Addr:    fmt.Sprintf(":%s", config.GetHealthRPCPort()),
		Handler: router,
	}
}

func performGracefulShutdown(srv *http.Server, wg *sync.WaitGroup, logger logging.Logger) {
	logger.Info("Initiating graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Update all keepers to inactive in database
	stateManager := keeper.GetStateManager()
	if err := stateManager.DumpState(); err != nil {
		logger.Error("Failed to dump keeper state", "error", err)
	}

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("HTTP server shutdown error", "error", err)
		if err := srv.Close(); err != nil {
			logger.Error("Forced HTTP server close error", "error", err)
		}
	}

	wg.Wait()
	logger.Info("Shutdown complete")
}
