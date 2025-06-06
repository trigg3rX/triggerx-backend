package test

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/trigg3rX/triggerx-backend/internal/health"
	"github.com/trigg3rX/triggerx-backend/internal/health/config"
	"github.com/trigg3rX/triggerx-backend/internal/health/keeper"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/mocks"
)

func TestMain(t *testing.T) {
	// Set environment variables for testing
	os.Setenv("ENV", "development")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("PORT", "8080")

	// Test configuration initialization
	t.Run("Config Initialization", func(t *testing.T) {
		// Skip if .env file doesn't exist
		if _, err := os.Stat(".env"); os.IsNotExist(err) {
			t.Skip("Skipping config initialization test as .env file is not present")
		}
		err := config.Init()
		assert.NoError(t, err, "Config initialization should not fail")
	})

	// Test HTTP server setup
	t.Run("HTTP Server Setup", func(t *testing.T) {
		logConfig := logging.LoggerConfig{
			LogDir:          logging.BaseDataDir,
			ProcessName:     "health-test",
			Environment:     logging.Development,
			UseColors:       true,
			MinStdoutLevel:  logging.DebugLevel,
			MinFileLogLevel: logging.DebugLevel,
		}
		_ = logging.InitServiceLogger(logConfig)
		logger := logging.GetServiceLogger()

		// Initialize test dependencies
		mocks.InitializeTestDependencies(logger)

		// Initialize state manager for tests
		_ = keeper.InitializeStateManager(logger)

		// Setup router
		router := gin.New()
		router.Use(gin.Recovery())
		router.Use(health.LoggerMiddleware(logger))
		health.RegisterRoutes(router)

		// Create server
		srv := &http.Server{
			Addr:    ":8080",
			Handler: router,
		}
		assert.NotNil(t, srv, "Server should be created successfully")
		logger.Info("Server created successfully")

		// Test server start
		go func() {
			err := srv.ListenAndServe()
			assert.ErrorIs(t, err, http.ErrServerClosed, "Server should close gracefully")
		}()

		// Give server time to start
		time.Sleep(100 * time.Millisecond)
		logger.Info("Server started successfully")

		// Test server stop
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := srv.Shutdown(ctx)
		assert.NoError(t, err, "Server should stop gracefully")
		logger.Info("Server stopped successfully")
	})

	// Test graceful shutdown
	t.Run("Graceful Shutdown", func(t *testing.T) {
		logConfig := logging.LoggerConfig{
			LogDir:          logging.BaseDataDir,
			ProcessName:     "health-test",
			Environment:     logging.Development,
			UseColors:       true,
			MinStdoutLevel:  logging.DebugLevel,
			MinFileLogLevel: logging.DebugLevel,
		}
		_ = logging.InitServiceLogger(logConfig)
		logger := logging.GetServiceLogger()

		// Initialize test dependencies
		mocks.InitializeTestDependencies(logger)

		// Initialize state manager for tests
		_ = keeper.InitializeStateManager(logger)

		// Setup router
		router := gin.New()
		router.Use(gin.Recovery())
		router.Use(health.LoggerMiddleware(logger))
		health.RegisterRoutes(router)

		// Create server
		srv := &http.Server{
			Addr:    ":8080",
			Handler: router,
		}
		assert.NotNil(t, srv, "Server should be created successfully")

		// Start server
		go func() {
			err := srv.ListenAndServe()
			assert.ErrorIs(t, err, http.ErrServerClosed, "Server should close gracefully")
		}()

		// Give server time to start
		time.Sleep(100 * time.Millisecond)

		// Test graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := srv.Shutdown(ctx)
		assert.NoError(t, err, "Server should stop gracefully")
		logger.Info("Server stopped gracefully")
	})
}
