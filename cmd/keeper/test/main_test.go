package test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/api"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/client/health"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

func TestMain(t *testing.T) {
	// Set environment variables for testing
	if err := os.Setenv("ENV", "development"); err != nil {
		t.Fatalf("Failed to set ENV: %v", err)
	}
	if err := os.Setenv("LOG_LEVEL", "debug"); err != nil {
		t.Fatalf("Failed to set LOG_LEVEL: %v", err)
	}
	if err := os.Setenv("PORT", "8080"); err != nil {
		t.Fatalf("Failed to set PORT: %v", err)
	}

	// Test configuration initialization
	t.Run("Config Initialization", func(t *testing.T) {
		// Skip if .env file doesn't exist
		if _, err := os.Stat(".env"); os.IsNotExist(err) {
			t.Skip("Skipping config initialization test as .env file is not present")
		}
		err := config.Init()
		assert.NoError(t, err, "Config initialization should not fail")
	})

	logConfig := logging.LoggerConfig{
		ProcessName:   logging.KeeperProcess,
		IsDevelopment: true,
	}
	logger, err := logging.NewZapLogger(logConfig)
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	// Test logger initialization
	t.Run("Logger Initialization", func(t *testing.T) {
		if logger == nil {
			t.Fatalf("Logger should not be nil")
		}
		assert.NoError(t, err, "Logger initialization should not fail")
	})

	// Test server setup
	t.Run("Server Setup", func(t *testing.T) {
		srv := api.NewServer(api.Config{
			Port: "8080",
		}, api.Dependencies{
			Logger: logger,
		})
		assert.NotNil(t, srv, "Server should be created successfully")
		logger.Info("Server created successfully")

		// Test server start
		errCh := make(chan error, 1)
		go func() {
			errCh <- srv.Start()
		}()

		// Give server time to start
		time.Sleep(100 * time.Millisecond)
		logger.Info("Server started successfully")

		// Test server stop
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := srv.Stop(ctx)
		assert.NoError(t, err, "Server should stop gracefully")
		logger.Info("Server stopped successfully")

		// Check the error from the goroutine
		err = <-errCh
		if err != nil {
			logger.Info("Server returned error on shutdown", "error", err)
		} else {
			logger.Info("Server shutdown returned nil error (graceful shutdown)")
		}
	})

	// Test health check client
	t.Run("Health Check Client", func(t *testing.T) {
		client, err := health.NewClient(logger, health.Config{
			HealthServiceURL: "http://localhost:8080",
			PrivateKey:       "test-private-key",
			KeeperAddress:    "test-keeper-address",
			PeerID:           "test-peer-id",
			Version:          "0.1.2",
			RequestTimeout:   10 * time.Second,
		})
		assert.NoError(t, err, "Health check client creation should not fail")
		assert.NotNil(t, client, "Health check client should be created successfully")
		logger.Info("Health check client created successfully")
	})
}
