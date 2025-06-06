package test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/api"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/client"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/config"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/scheduler"
	"github.com/trigg3rX/triggerx-backend/pkg/client/aggregator"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

func TestMain(t *testing.T) {
	// Test configuration initialization
	t.Run("Config Initialization", func(t *testing.T) {
		err := config.Init()
		assert.NoError(t, err, "Config initialization should not fail")
	})

	// Test logger initialization
	t.Run("Logger Initialization", func(t *testing.T) {
		logConfig := logging.LoggerConfig{
			LogDir:          logging.BaseDataDir,
			ProcessName:     logging.TimeSchedulerProcess,
			Environment:     logging.Development,
			UseColors:       true,
			MinStdoutLevel:  logging.DebugLevel,
			MinFileLogLevel: logging.DebugLevel,
		}
		err := logging.InitServiceLogger(logConfig)
		assert.NoError(t, err, "Logger initialization should not fail")
	})

	// Test database client
	t.Run("Database Client", func(t *testing.T) {
		logger := logging.GetServiceLogger()
		dbClientCfg := client.Config{
			DBServerURL:    config.GetDBServerURL(),
			RequestTimeout: 10 * time.Second,
			MaxRetries:     3,
			RetryDelay:     2 * time.Second,
		}
		dbClient, err := client.NewDBServerClient(logger, dbClientCfg)
		assert.NoError(t, err, "Database client creation should not fail")
		assert.NotNil(t, dbClient, "Database client should not be nil")
		logger.Info("Database client created successfully")

		// Test health check
		err = dbClient.HealthCheck()
		if err != nil {
			logger.Warn("Database server health check failed", "error", err)
		} else {
			logger.Info("Database server health check passed")
		}

		// Test client close
		dbClient.Close()
		logger.Info("Database client closed successfully")
	})

	// Test aggregator client
	t.Run("Aggregator Client", func(t *testing.T) {
		logger := logging.GetServiceLogger()
		aggClientCfg := aggregator.AggregatorClientConfig{
			AggregatorRPCUrl: config.GetAggregatorRPCUrl(),
			SenderPrivateKey: config.GetSchedulerPrivateKey(),
			SenderAddress:    config.GetSchedulerAddress(),
			RetryAttempts:    3,
			RetryDelay:       2 * time.Second,
			RequestTimeout:   10 * time.Second,
		}
		aggClient, err := aggregator.NewAggregatorClient(logger, aggClientCfg)
		assert.NoError(t, err, "Aggregator client creation should not fail")
		assert.NotNil(t, aggClient, "Aggregator client should not be nil")
		logger.Info("Aggregator client created successfully")
	})

	// Test scheduler setup
	t.Run("Scheduler Setup", func(t *testing.T) {
		logger := logging.GetServiceLogger()
		dbClientCfg := client.Config{
			DBServerURL:    config.GetDBServerURL(),
			RequestTimeout: 10 * time.Second,
			MaxRetries:     3,
			RetryDelay:     2 * time.Second,
		}
		dbClient, err := client.NewDBServerClient(logger, dbClientCfg)
		assert.NoError(t, err, "Database client creation should not fail")
		defer func() {
			dbClient.Close()
			logger.Info("Database client closed successfully")
		}()

		aggClientCfg := aggregator.AggregatorClientConfig{
			AggregatorRPCUrl: config.GetAggregatorRPCUrl(),
			SenderPrivateKey: config.GetSchedulerPrivateKey(),
			SenderAddress:    config.GetSchedulerAddress(),
			RetryAttempts:    3,
			RetryDelay:       2 * time.Second,
			RequestTimeout:   10 * time.Second,
		}
		aggClient, err := aggregator.NewAggregatorClient(logger, aggClientCfg)
		assert.NoError(t, err, "Aggregator client creation should not fail")

		// Create scheduler
		managerID := "test-time-scheduler"
		timeScheduler, err := scheduler.NewTimeBasedScheduler(managerID, logger, dbClient, aggClient)
		assert.NoError(t, err, "Scheduler creation should not fail")
		assert.NotNil(t, timeScheduler, "Scheduler should not be nil")
		logger.Info("Time scheduler created successfully")

		// Test scheduler start
		ctx, cancel := context.WithCancel(context.Background())
		timeScheduler.Start(ctx)
		logger.Info("Time scheduler started successfully")

		// Test scheduler stop
		timeScheduler.Stop()
		cancel()
		logger.Info("Time scheduler stopped successfully")
	})

	// Test server setup
	t.Run("Server Setup", func(t *testing.T) {
		logger := logging.GetServiceLogger()
		dbClientCfg := client.Config{
			DBServerURL:    config.GetDBServerURL(),
			RequestTimeout: 10 * time.Second,
			MaxRetries:     3,
			RetryDelay:     2 * time.Second,
		}
		dbClient, err := client.NewDBServerClient(logger, dbClientCfg)
		assert.NoError(t, err, "Database client creation should not fail")
		defer func() {
			dbClient.Close()
			logger.Info("Database client closed successfully")
		}()

		aggClientCfg := aggregator.AggregatorClientConfig{
			AggregatorRPCUrl: config.GetAggregatorRPCUrl(),
			SenderPrivateKey: config.GetSchedulerPrivateKey(),
			SenderAddress:    config.GetSchedulerAddress(),
			RetryAttempts:    3,
			RetryDelay:       2 * time.Second,
			RequestTimeout:   10 * time.Second,
		}
		aggClient, err := aggregator.NewAggregatorClient(logger, aggClientCfg)
		assert.NoError(t, err, "Aggregator client creation should not fail")

		// Create scheduler
		managerID := "test-time-scheduler"
		timeScheduler, err := scheduler.NewTimeBasedScheduler(managerID, logger, dbClient, aggClient)
		assert.NoError(t, err, "Scheduler creation should not fail")

		// Create server
		srv := api.NewServer(api.Config{
			Port: "8080",
		}, api.Dependencies{
			Logger:    logger,
			Scheduler: timeScheduler,
		})
		assert.NotNil(t, srv, "Server should be created successfully")
		logger.Info("Server created successfully")

		// Test server start
		go func() {
			err := srv.Start()
			assert.ErrorIs(t, err, http.ErrServerClosed, "Server should close gracefully")
		}()

		// Give server time to start
		time.Sleep(100 * time.Millisecond)
		logger.Info("Server started successfully")

		// Test server stop
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err = srv.Stop(ctx)
		assert.NoError(t, err, "Server should stop gracefully")
		logger.Info("Server stopped successfully")
	})

	// Test environment and log level
	t.Run("Environment and Log Level", func(t *testing.T) {
		env := getEnvironment()
		assert.Contains(t, []logging.LogLevel{logging.Development, logging.Production}, env, "Environment should be either Development or Production")

		level := getLogLevel()
		assert.Contains(t, []logging.Level{logging.DebugLevel, logging.InfoLevel}, level, "Log level should be either Debug or Info")
	})
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
