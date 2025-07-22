package taskmanager

import (
	"context"
	"fmt"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/taskmanager/config"
	"github.com/trigg3rX/triggerx-backend/internal/taskmanager/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/taskmanager/streams/jobs"
	"github.com/trigg3rX/triggerx-backend/internal/taskmanager/streams/performers"
	"github.com/trigg3rX/triggerx-backend/internal/taskmanager/streams/tasks"
	redisClient "github.com/trigg3rX/triggerx-backend/pkg/client/redis"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// TaskManager orchestrates all Redis-based task management components
type TaskManager struct {
	logger              logging.Logger
	redisClient         redisClient.RedisClientInterface
	taskStreamManager   *tasks.TaskStreamManager
	jobStreamManager    *jobs.JobStreamManager
	performerManager    *performers.PerformerManager
	metricsUpdateTicker *time.Ticker
	ctx                 context.Context
	cancel              context.CancelFunc
}

// NewTaskManager creates a new TaskManager instance
func NewTaskManager(logger logging.Logger) (*TaskManager, error) {
	logger.Info("Initializing TaskManager...")

	// Check if Redis is available
	if !config.IsRedisAvailable() {
		logger.Error("Redis not available for TaskManager initialization")
		metrics.ServiceStatus.WithLabelValues("task_manager").Set(0)
		return nil, fmt.Errorf("redis not available")
	}

	// Create Redis client with monitoring
	redisConfig := config.GetRedisClientConfig()
	client, err := redisClient.NewRedisClient(logger, redisConfig)
	if err != nil {
		logger.Error("Failed to create Redis client for TaskManager", "error", err)
		metrics.ServiceStatus.WithLabelValues("task_manager").Set(0)
		return nil, fmt.Errorf("failed to create redis client: %w", err)
	}

	// Set up monitoring hooks
	monitoringHooks := metrics.CreateRedisMonitoringHooks()
	client.SetMonitoringHooks(monitoringHooks)

	// Create context for managing background workers
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize stream managers
	taskStreamManager, err := tasks.NewTaskStreamManager(logger, client)
	if err != nil {
		cancel()
		logger.Error("Failed to create TaskStreamManager", "error", err)
		return nil, fmt.Errorf("failed to create task stream manager: %w", err)
	}

	jobStreamManager, err := jobs.NewJobStreamManager(logger, client)
	if err != nil {
		cancel()
		logger.Error("Failed to create JobStreamManager", "error", err)
		return nil, fmt.Errorf("failed to create job stream manager: %w", err)
	}

	performerManager := performers.NewPerformerManager(client, logger)

	tm := &TaskManager{
		logger:              logger,
		redisClient:         client,
		taskStreamManager:   taskStreamManager,
		jobStreamManager:    jobStreamManager,
		performerManager:    performerManager,
		metricsUpdateTicker: time.NewTicker(30 * time.Second), // Update metrics every 30 seconds
		ctx:                 ctx,
		cancel:              cancel,
	}

	logger.Info("TaskManager initialized successfully",
		"redis_type", config.GetRedisType())

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

	// Initialize job streams
	if err := tm.jobStreamManager.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize job stream manager: %w", err)
	}

	// Start background workers
	go tm.startMetricsUpdateWorker()
	go tm.taskStreamManager.StartTaskProcessor(tm.ctx, "taskmanager-processor")
	go tm.taskStreamManager.StartTimeoutWorker(tm.ctx)
	go tm.taskStreamManager.StartRetryWorker(tm.ctx)

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

	jobStreamInfo := tm.jobStreamManager.GetJobStreamInfo()
	if lengths, ok := jobStreamInfo["stream_lengths"].(map[string]int64); ok {
		for stream, length := range lengths {
			switch stream {
			case "jobs:running":
				metrics.JobStreamLengths.WithLabelValues("running").Set(float64(length))
			case "jobs:completed":
				metrics.JobStreamLengths.WithLabelValues("completed").Set(float64(length))
			}
		}
	}

	// Update connection status
	connectionStatus := tm.redisClient.GetConnectionStatus()
	if isRecovering, ok := connectionStatus["is_recovering"].(bool); ok && !isRecovering {
		metrics.RedisConnectionHealth.WithLabelValues("main").Set(1)
	}
}

// GetTaskStreamManager returns the task stream manager
func (tm *TaskManager) GetTaskStreamManager() *tasks.TaskStreamManager {
	return tm.taskStreamManager
}

// GetJobStreamManager returns the job stream manager
func (tm *TaskManager) GetJobStreamManager() *jobs.JobStreamManager {
	return tm.jobStreamManager
}

// GetPerformerManager returns the performer manager
func (tm *TaskManager) GetPerformerManager() *performers.PerformerManager {
	return tm.performerManager
}

// GetRedisClient returns the Redis client
func (tm *TaskManager) GetRedisClient() redisClient.RedisClientInterface {
	return tm.redisClient
}

// HealthCheck performs a comprehensive health check
func (tm *TaskManager) HealthCheck() map[string]interface{} {
	tm.logger.Debug("Performing TaskManager health check")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	healthStatus := map[string]interface{}{
		"redis_available": config.IsRedisAvailable(),
		"redis_type":      config.GetRedisType(),
		"timestamp":       time.Now(),
	}

	// Check Redis connection
	if tm.redisClient != nil {
		redisHealth := tm.redisClient.GetHealthStatus(ctx)
		healthStatus["redis_connection"] = map[string]interface{}{
			"connected":    redisHealth.Connected,
			"last_ping":    redisHealth.LastPing,
			"ping_latency": redisHealth.PingLatency,
			"errors":       redisHealth.Errors,
			"type":         redisHealth.Type,
		}
	}

	// Get stream information
	if tm.taskStreamManager != nil {
		healthStatus["task_streams"] = tm.taskStreamManager.GetStreamInfo()
	}

	if tm.jobStreamManager != nil {
		healthStatus["job_streams"] = tm.jobStreamManager.GetJobStreamInfo()
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

	// Stop metrics ticker
	if tm.metricsUpdateTicker != nil {
		tm.metricsUpdateTicker.Stop()
	}

	// Close stream managers
	var errors []error

	if tm.taskStreamManager != nil {
		if err := tm.taskStreamManager.Close(); err != nil {
			tm.logger.Error("Failed to close TaskStreamManager", "error", err)
			errors = append(errors, err)
		}
	}

	if tm.jobStreamManager != nil {
		if err := tm.jobStreamManager.Close(); err != nil {
			tm.logger.Error("Failed to close JobStreamManager", "error", err)
			errors = append(errors, err)
		}
	}

	// Close Redis client
	if tm.redisClient != nil {
		if err := tm.redisClient.Close(); err != nil {
			tm.logger.Error("Failed to close Redis client", "error", err)
			errors = append(errors, err)
		}
	}

	metrics.ServiceStatus.WithLabelValues("task_manager").Set(0)

	if len(errors) > 0 {
		return fmt.Errorf("errors during TaskManager shutdown: %v", errors)
	}

	tm.logger.Info("TaskManager closed successfully")
	return nil
}
