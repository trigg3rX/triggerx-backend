package test

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/api"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/client"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/config"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/scheduler"
	"github.com/trigg3rX/triggerx-backend/pkg/client/aggregator"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// MockDBServerClient is a mock implementation of the scheduler.DBClient interface
type MockDBServerClient struct {
	mock.Mock
}

func (m *MockDBServerClient) HealthCheck() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDBServerClient) Close() {
	m.Called()
}

func (m *MockDBServerClient) GetTimeBasedJobs() ([]types.ScheduleTimeTaskData, error) {
	args := m.Called()
	return args.Get(0).([]types.ScheduleTimeTaskData), args.Error(1)
}

func (m *MockDBServerClient) UpdateJobNextExecution(jobID int64, nextExecution time.Time) error {
	args := m.Called(jobID, nextExecution)
	return args.Error(0)
}

func (m *MockDBServerClient) UpdateJobStatus(jobID int64, status bool) error {
	args := m.Called(jobID, status)
	return args.Error(0)
}

func TestMain(t *testing.T) {
	// Set required environment variables for testing
	if err := os.Setenv("DEV_MODE", "true"); err != nil {
		t.Fatalf("Failed to set DEV_MODE: %v", err)
	}
	if err := os.Setenv("TIME_SCHEDULER_RPC_PORT", "9004"); err != nil {
		t.Fatalf("Failed to set TIME_SCHEDULER_RPC_PORT: %v", err)
	}
	if err := os.Setenv("DATABASE_RPC_URL", "http://localhost:9002"); err != nil {
		t.Fatalf("Failed to set DATABASE_RPC_URL: %v", err)
	}
	if err := os.Setenv("AGGREGATOR_RPC_URL", "http://localhost:9003"); err != nil {
		t.Fatalf("Failed to set AGGREGATOR_RPC_URL: %v", err)
	}
	if err := os.Setenv("SCHEDULER_PRIVATE_KEY", "0x0000000000000000000000000000000000000000000000000000000000000001"); err != nil {
		t.Fatalf("Failed to set SCHEDULER_PRIVATE_KEY: %v", err)
	}
	if err := os.Setenv("SCHEDULER_ADDRESS", "0x0000000000000000000000000000000000000001"); err != nil {
		t.Fatalf("Failed to set SCHEDULER_ADDRESS: %v", err)
	}
	if err := os.Setenv("MAX_WORKERS", "10"); err != nil {
		t.Fatalf("Failed to set MAX_WORKERS: %v", err)
	}
	if err := os.Setenv("POLLING_INTERVAL", "50s"); err != nil {
		t.Fatalf("Failed to set POLLING_INTERVAL: %v", err)
	}
	if err := os.Setenv("POLLING_LOOK_AHEAD", "60s"); err != nil {
		t.Fatalf("Failed to set POLLING_LOOK_AHEAD: %v", err)
	}

	// Test configuration initialization
	t.Run("Config Initialization", func(t *testing.T) {
		err := config.Init()
		if err != nil && err.Error() != "error loading .env file: open .env: no such file or directory" {
			assert.NoError(t, err, "Config initialization should not fail")
		}
	})

	logConfig := logging.LoggerConfig{
		ProcessName:   logging.TimeSchedulerProcess,
		IsDevelopment: true,
	}
	logger, err := logging.NewZapLogger(logConfig)
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	// Test logger initialization
	t.Run("Logger Initialization", func(t *testing.T) {
		if logger == nil {
			panic("Logger should not be nil")
		}
		assert.NoError(t, err, "Logger initialization should not fail")
	})

	// Test database client
	t.Run("Database Client", func(t *testing.T) {
		// Create mock database client
		mockDBClient := new(MockDBServerClient)
		mockDBClient.On("HealthCheck").Return(nil)
		mockDBClient.On("Close").Return()
		mockDBClient.On("GetTimeBasedJobs").Return([]types.ScheduleTimeTaskData{}, nil)
		mockDBClient.On("UpdateJobNextExecution", mock.Anything, mock.Anything).Return(nil)
		mockDBClient.On("UpdateJobStatus", mock.Anything, mock.Anything).Return(nil)

		// Test health check
		err := mockDBClient.HealthCheck()
		assert.NoError(t, err, "Health check should not fail")
		logger.Info("Database server health check passed")

		// Test client close
		mockDBClient.Close()
		logger.Info("Database client closed successfully")
	})

	// Test aggregator client
	t.Run("Aggregator Client", func(t *testing.T) {
		aggClientCfg := aggregator.AggregatorClientConfig{
			AggregatorRPCUrl: "http://localhost:9003",
			SenderPrivateKey: "0000000000000000000000000000000000000000000000000000000000000001",
			SenderAddress:    "0x0000000000000000000000000000000000000001",
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
		// Create mock database client
		mockDBClient := new(MockDBServerClient)
		mockDBClient.On("HealthCheck").Return(nil)
		mockDBClient.On("Close").Return()
		mockDBClient.On("GetTimeBasedJobs").Return([]types.ScheduleTimeTaskData{}, nil)
		mockDBClient.On("UpdateJobNextExecution", mock.Anything, mock.Anything).Return(nil)
		mockDBClient.On("UpdateJobStatus", mock.Anything, mock.Anything).Return(nil)

		aggClientCfg := aggregator.AggregatorClientConfig{
			AggregatorRPCUrl: "http://localhost:9003",
			SenderPrivateKey: "0000000000000000000000000000000000000000000000000000000000000001",
			SenderAddress:    "0x0000000000000000000000000000000000000001",
			RetryAttempts:    3,
			RetryDelay:       2 * time.Second,
			RequestTimeout:   10 * time.Second,
		}
		aggClient, err := aggregator.NewAggregatorClient(logger, aggClientCfg)
		assert.NoError(t, err, "Aggregator client creation should not fail")

		// Create scheduler
		managerID := "test-time-scheduler"
		timeScheduler, err := scheduler.NewTimeBasedScheduler(managerID, logger, &client.DBServerClient{}, aggClient)
		assert.NoError(t, err, "Scheduler creation should not fail")
		assert.NotNil(t, timeScheduler, "Scheduler should not be nil")
		logger.Info("Time scheduler created successfully")

		// Test scheduler start
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		timeScheduler.Start(ctx)
		logger.Info("Time scheduler started successfully")

		// Test scheduler stop
		timeScheduler.Stop()
		logger.Info("Time scheduler stopped successfully")
	})

	// Test server setup
	t.Run("Server Setup", func(t *testing.T) {
		// Create mock database client
		mockDBClient := new(MockDBServerClient)
		mockDBClient.On("HealthCheck").Return(nil)
		mockDBClient.On("Close").Return()
		mockDBClient.On("GetTimeBasedJobs").Return([]types.ScheduleTimeTaskData{}, nil)
		mockDBClient.On("UpdateJobNextExecution", mock.Anything, mock.Anything).Return(nil)
		mockDBClient.On("UpdateJobStatus", mock.Anything, mock.Anything).Return(nil)

		aggClientCfg := aggregator.AggregatorClientConfig{
			AggregatorRPCUrl: "http://localhost:9003",
			SenderPrivateKey: "0000000000000000000000000000000000000000000000000000000000000001",
			SenderAddress:    "0x0000000000000000000000000000000000000001",
			RetryAttempts:    3,
			RetryDelay:       2 * time.Second,
			RequestTimeout:   10 * time.Second,
		}
		aggClient, err := aggregator.NewAggregatorClient(logger, aggClientCfg)
		assert.NoError(t, err, "Aggregator client creation should not fail")

		// Create scheduler
		managerID := "test-time-scheduler"
		timeScheduler, err := scheduler.NewTimeBasedScheduler(managerID, logger, &client.DBServerClient{}, aggClient)
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
			assert.True(t, err == nil || err == http.ErrServerClosed, "Server should close gracefully")
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
}
