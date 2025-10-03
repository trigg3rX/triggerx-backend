package taskmonitor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gocql/gocql"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/clients/database"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/config"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/events"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/tasks"
	redisClient "github.com/trigg3rX/triggerx-backend/pkg/client/redis"
	"github.com/trigg3rX/triggerx-backend/pkg/datastore"
	"github.com/trigg3rX/triggerx-backend/pkg/datastore/infrastructure/connection"
	"github.com/trigg3rX/triggerx-backend/pkg/ipfs"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
)

const (
// defaultConnectTimeout = 30 * time.Second
// defaultBlockOverlap   = uint64(5)
)

// TaskManager orchestrates all Redis-based task management components
type TaskManager struct {
	logger              logging.Logger
	redisClient         *redisClient.Client
	taskStreamManager   *tasks.TaskStreamManager
	eventListener       *events.ContractEventListener
	testEventListener   *events.ContractEventListener
	metricsUpdateTicker *time.Ticker
	ctx                 context.Context
	cancel              context.CancelFunc
	shutdownWg          sync.WaitGroup
	startTime           time.Time
}

// NewTaskManager creates a new TaskManager instance
func NewTaskManager(logger logging.Logger) (*TaskManager, error) {
	logger.Info("Initializing TaskManager...")

	// Create context for managing background workers
	ctx, cancel := context.WithCancel(context.Background())

	// Create Redis client with monitoring
	redisConfig := config.GetRedisClientConfig()
	client, err := redisClient.NewRedisClient(logger, redisConfig)
	if err != nil {
		cancel() // Clean up context on error
		logger.Error("Failed to create Redis client for TaskManager", "error", err)
		metrics.ServiceStatus.WithLabelValues("task_manager").Set(0)
		return nil, fmt.Errorf("failed to create redis client: %w", err)
	}

	// Set up monitoring hooks
	monitoringHooks := metrics.CreateRedisMonitoringHooks()
	client.SetMonitoringHooks(monitoringHooks)

	// Initialize database client
	dbCfg := &connection.Config{
		Hosts:       []string{config.GetDatabaseHostAddress() + ":" + config.GetDatabaseHostPort()},
		Keyspace:    "triggerx",
		Consistency: gocql.Quorum,
		Timeout:     10 * time.Second,
		Retries:     3,
		ConnectWait: 5 * time.Second,
		RetryConfig: retry.DefaultRetryConfig(),
	}
	datastore, err := datastore.NewService(dbCfg, logger)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize datastore: %w", err)
	}
	taskRepo := datastore.Task()
	keeperRepo := datastore.Keeper()
	userRepo := datastore.User()
	jobRepo := datastore.Job()

	// Initialize database client
	databaseClient := database.NewDatabaseClient(logger, taskRepo, keeperRepo, userRepo, jobRepo)

	// Initialize IPFS client
	ipfsCfg := ipfs.NewConfig(config.GetPinataHost(), config.GetPinataJWT())
	ipfsClient, err := ipfs.NewClient(ipfsCfg, logger)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize IPFS client: %w", err)
	}

	// Initialize task stream manager
	taskStreamManager, err := tasks.NewTaskStreamManager(client, databaseClient, logger)
	if err != nil {
		// Clean up resources on error
		cancel()
		if closeErr := client.Close(); closeErr != nil {
			logger.Error("Failed to close Redis client during error cleanup", "error", closeErr)
		}
		logger.Error("Failed to create TaskStreamManager", "error", err)
		return nil, fmt.Errorf("failed to create task stream manager: %w", err)
	}

	// Initialize event listener
	eventListener := events.NewContractEventListener(logger, events.GetMainnetConfig(), databaseClient, ipfsClient, taskStreamManager)
	testEventListener := events.NewContractEventListener(logger, events.GetTestnetConfig(), databaseClient, ipfsClient, taskStreamManager)

	tm := &TaskManager{
		logger:              logger,
		redisClient:         client,
		taskStreamManager:   taskStreamManager,
		eventListener:       eventListener,
		testEventListener:   testEventListener,
		metricsUpdateTicker: time.NewTicker(config.GetMetricsUpdateInterval()),
		ctx:                 ctx,
		cancel:              cancel,
		startTime:           time.Now(),
	}

	logger.Info("TaskManager initialized successfully",
		"metrics_update_interval", config.GetMetricsUpdateInterval())

	metrics.ServiceStatus.WithLabelValues("task_manager").Set(1)
	return tm, nil
}

// Initialize initializes all stream managers and starts background workers
func (tm *TaskManager) Initialize() error {
	tm.logger.Info("Initializing TaskManager components...")

	// Initialize task streams
	if err := tm.taskStreamManager.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize task stream manager: %w", err)
	}

	// Start event listener
	if err := tm.eventListener.Start(); err != nil {
		tm.logger.Errorf("Failed to start event listener: %v", err)
		tm.logger.Info("Falling back to polling mode")
	}

	if err := tm.testEventListener.Start(); err != nil {
		tm.logger.Errorf("Failed to start test event listener: %v", err)
		tm.logger.Info("Falling back to polling mode")
	}

	// Start background workers with proper synchronization
	tm.shutdownWg.Add(3) // Track all background goroutines

	go func() {
		defer tm.shutdownWg.Done()
		tm.startMetricsUpdateWorker()
	}()

	go func() {
		defer tm.shutdownWg.Done()
		tm.taskStreamManager.StartTimeoutWorker(tm.ctx)
	}()

	tm.logger.Info("TaskManager initialization completed successfully")
	return nil
}

// startMetricsUpdateWorker periodically updates metrics from Redis client
func (tm *TaskManager) startMetricsUpdateWorker() {
	tm.logger.Info("Starting metrics update worker")

	for {
		select {
		case <-tm.ctx.Done():
			tm.logger.Info("Metrics update worker stopping")
			return
		case <-tm.metricsUpdateTicker.C:
			tm.updateMetrics()
		}
	}
}

// updateMetrics updates Prometheus metrics from Redis client
func (tm *TaskManager) updateMetrics() {
	defer func() {
		if r := recover(); r != nil {
			tm.logger.Error("Panic in metrics update", "panic", r)
		}
	}()

	// Update Redis client metrics
	operationMetrics := tm.redisClient.GetOperationMetrics()
	metrics.UpdateRedisClientMetrics(operationMetrics)

	// Update stream length metrics
	taskStreamInfo := tm.taskStreamManager.GetStreamInfo()
	if lengths, ok := taskStreamInfo["stream_lengths"].(map[string]int64); ok {
		for stream, length := range lengths {
			switch stream {
			case "tasks:ready":
				metrics.TaskStreamLengths.WithLabelValues("ready").Set(float64(length))
			case "tasks:processing":
				metrics.TaskStreamLengths.WithLabelValues("processing").Set(float64(length))
			case "tasks:completed":
				metrics.TaskStreamLengths.WithLabelValues("completed").Set(float64(length))
			case "tasks:failed":
				metrics.TaskStreamLengths.WithLabelValues("failed").Set(float64(length))
			case "tasks:retry":
				metrics.TaskStreamLengths.WithLabelValues("retry").Set(float64(length))
			}
		}
	}

	// Update connection status
	connectionStatus := tm.redisClient.GetConnectionStatus()
	if connectionStatus != nil && !connectionStatus.IsRecovering {
		metrics.RedisConnectionHealth.WithLabelValues("main").Set(1)
	}
}

// GetTaskStreamManager returns the task stream manager
func (tm *TaskManager) GetTaskStreamManager() *tasks.TaskStreamManager {
	return tm.taskStreamManager
}

// HealthCheck performs a comprehensive health check
func (tm *TaskManager) HealthCheck() map[string]interface{} {
	tm.logger.Debug("Performing TaskManager health check")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	healthStatus := map[string]interface{}{
		"timestamp":      time.Now(),
		"uptime_seconds": time.Since(tm.startTime).Seconds(),
		"start_time":     tm.startTime.Format(time.RFC3339),
	}

	// Check Redis connection
	if tm.redisClient != nil {
		redisHealth := tm.redisClient.GetHealthStatus(ctx)
		healthStatus["redis_connection"] = map[string]interface{}{
			"connected":    redisHealth.Connected,
			"last_ping":    redisHealth.LastPing,
			"ping_latency": redisHealth.PingLatency,
			"errors":       redisHealth.Errors,
		}
	}

	// Get stream information
	if tm.taskStreamManager != nil {
		healthStatus["task_streams"] = tm.taskStreamManager.GetStreamInfo()
	}

	return healthStatus
}

// Close gracefully shuts down the TaskManager
func (tm *TaskManager) Close() error {
	tm.logger.Info("Closing TaskManager...")

	// Cancel context to stop all workers
	if tm.cancel != nil {
		tm.cancel()
	}

	// Wait for background workers to finish with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	shutdownDone := make(chan struct{})
	go func() {
		tm.shutdownWg.Wait()
		close(shutdownDone)
	}()

	select {
	case <-shutdownDone:
		tm.logger.Info("All background workers stopped successfully")
	case <-shutdownCtx.Done():
		tm.logger.Warn("Timeout waiting for background workers to stop")
	}

	// Stop event listener
	if err := tm.eventListener.Stop(); err != nil {
		tm.logger.Errorf("Error stopping event listener: %v", err)
	}

	if err := tm.testEventListener.Stop(); err != nil {
		tm.logger.Errorf("Error stopping test event listener: %v", err)
	}

	// Stop metrics ticker
	if tm.metricsUpdateTicker != nil {
		tm.metricsUpdateTicker.Stop()
	}

	// Close stream managers
	var errors []error

	if tm.taskStreamManager != nil {
		if err := tm.taskStreamManager.Close(); err != nil {
			tm.logger.Error("Failed to close TaskStreamManager", "error", err)
			errors = append(errors, fmt.Errorf("task stream manager: %w", err))
		}
		// TaskStreamManager handles Redis client closure, so we don't need to close it again
		tm.redisClient = nil
	} else {
		// Only close Redis client if TaskStreamManager is nil
		if tm.redisClient != nil {
			if err := tm.redisClient.Close(); err != nil {
				tm.logger.Error("Failed to close Redis client", "error", err)
				errors = append(errors, fmt.Errorf("redis client: %w", err))
			}
		}
	}

	metrics.ServiceStatus.WithLabelValues("task_manager").Set(0)

	if len(errors) > 0 {
		tm.logger.Warn("Some non-critical errors occurred during shutdown", "error_count", len(errors))
		for i, err := range errors {
			tm.logger.Debug("Shutdown error", "index", i, "error", err)
		}
		// Don't return error for cleanup issues - shutdown was successful
	}

	tm.logger.Info("TaskManager closed successfully")
	return nil
}
