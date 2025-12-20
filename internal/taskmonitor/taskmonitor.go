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
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/types"
	redisClient "github.com/trigg3rX/triggerx-backend/pkg/client/redis"
	dbClient "github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/ipfs"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
)

const (
// defaultConnectTimeout = 30 * time.Second
// defaultBlockOverlap   = uint64(5)
)

// TaskManager orchestrates all Redis-based task management components
type TaskManager struct {
	logger              observability.Logger
	tracer              observability.Tracer
	redisClient         *redisClient.Client
	taskStreamManager   *tasks.TaskStreamManager
	eventListener       *events.ContractEventListener
	testEventListener   *events.ContractEventListener
	metricsUpdateTicker *time.Ticker
	ctx                 context.Context
	cancel              context.CancelFunc
	shutdownWg          sync.WaitGroup
	startTime           time.Time
	dbClient            *database.DatabaseClient
	rpcServer           interface {
		Stop(ctx context.Context) error
	}
}

// NewTaskManager creates a new TaskManager instance
func NewTaskManager(ctx context.Context, logger observability.Logger, tracer observability.Tracer) (*TaskManager, error) {
	logger.Info(ctx, "Initializing TaskManager...")

	// Create context for managing background workers
	ctx, cancel := context.WithCancel(context.Background())

	// Create Redis client with monitoring
	redisConfig := config.GetRedisClientConfig()
	client, err := redisClient.NewRedisClient(ctx, logger, redisConfig)
	if err != nil {
		cancel() // Clean up context on error
		logger.Error(ctx, "Failed to create Redis client for TaskManager", observability.Error(err))
		metrics.ServiceStatus.WithLabelValues("task_manager").Set(0)
		return nil, fmt.Errorf("failed to create redis client: %w", err)
	}

	// Set up monitoring hooks
	monitoringHooks := metrics.CreateRedisMonitoringHooks()
	client.SetMonitoringHooks(monitoringHooks)

	// Initialize database client
	dbCfg := &dbClient.Config{
		Hosts:       []string{config.GetDatabaseHostAddress() + ":" + config.GetDatabaseHostPort()},
		Keyspace:    "triggerx",
		Consistency: gocql.Quorum,
		Timeout:     10 * time.Second,
		Retries:     3,
		ConnectWait: 5 * time.Second,
		RetryConfig: retry.DefaultRetryConfig(),
	}
	dbConn, err := dbClient.NewConnection(dbCfg, logger)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize database client: %w", err)
	}

	// Initialize database client
	databaseClient := database.NewDatabaseClient(logger, dbConn)

	// Initialize IPFS client
	ipfsCfg := ipfs.NewConfig(config.GetPinataHost(), config.GetPinataJWT())
	ipfsClient, err := ipfs.NewClient(ipfsCfg)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize IPFS client: %w", err)
	}

	// Initialize task stream manager
	taskStreamManager, err := tasks.NewTaskStreamManager(ctx, client, databaseClient, logger)
	if err != nil {
		// Clean up resources on error
		cancel()
		if closeErr := client.Close(); closeErr != nil {
			logger.Error(ctx, "Failed to close Redis client during error cleanup", observability.Error(closeErr))
		}
		logger.Error(ctx, "Failed to create TaskStreamManager", observability.Error(err))
		return nil, fmt.Errorf("failed to create task stream manager: %w", err)
	}

	// Initialize event listener
	eventListener := events.NewContractEventListener(logger, tracer, events.GetMainnetConfig(), databaseClient, ipfsClient, taskStreamManager)
	testEventListener := events.NewContractEventListener(logger, tracer, events.GetTestnetConfig(), databaseClient, ipfsClient, taskStreamManager)

	tm := &TaskManager{
		logger:              logger,
		tracer:              tracer,
		redisClient:         client,
		taskStreamManager:   taskStreamManager,
		eventListener:       eventListener,
		testEventListener:   testEventListener,
		metricsUpdateTicker: time.NewTicker(config.GetMetricsUpdateInterval()),
		ctx:                 ctx,
		cancel:              cancel,
		startTime:           time.Now(),
		dbClient:            databaseClient,
	}

	logger.Info(ctx, "TaskManager initialized successfully",
		observability.Duration("metrics_update_interval", config.GetMetricsUpdateInterval()))

	metrics.ServiceStatus.WithLabelValues("task_manager").Set(1)
	return tm, nil
}

// Initialize initializes all stream managers and starts background workers
func (tm *TaskManager) Initialize() error {
	tm.logger.Info(tm.ctx, "Initializing TaskManager components...")

	// Initialize task streams
	if err := tm.taskStreamManager.Initialize(tm.ctx); err != nil {
		return fmt.Errorf("failed to initialize task stream manager: %w", err)
	}

	// Start event listener
	if err := tm.eventListener.Start(tm.ctx); err != nil {
		tm.logger.Error(tm.ctx, "Failed to start event listener", observability.Error(err))
		tm.logger.Info(tm.ctx, "Falling back to polling mode")
	}

	if err := tm.testEventListener.Start(tm.ctx); err != nil {
		tm.logger.Error(tm.ctx, "Failed to start test event listener", observability.Error(err))
		tm.logger.Info(tm.ctx, "Falling back to polling mode")
	}

	// Start background workers with proper synchronization
	tm.shutdownWg.Add(2) // Track all background goroutines (metrics worker + timeout worker)

	go func() {
		defer tm.shutdownWg.Done()
		tm.startMetricsUpdateWorker()
	}()

	go func() {
		defer tm.shutdownWg.Done()
		tm.taskStreamManager.StartTimeoutWorker(tm.ctx)
	}()

	tm.logger.Info(tm.ctx, "TaskManager initialization completed successfully")
	return nil
}

// ReportTaskStatus handles task status reports from keepers
// This is called after the aggregator submission attempt (regardless of success or failure)
// ProofCID contains all execution data (task data, action data, proof, signatures)
func (tm *TaskManager) ReportTaskStatus(ctx context.Context, req *types.ReportTaskStatusRequest) (*types.ReportTaskStatusResponse, error) {
	tm.logger.Info(ctx, "Received task status report",
		observability.Int64("task_id", req.TaskID),
		observability.String("keeper_address", req.KeeperAddress),
		observability.Bool("execution_successful", req.ExecutionSuccessful),
		observability.Bool("aggregator_submitted", req.AggregatorSubmitted),
		observability.String("execution_tx_hash", req.ExecutionTxHash),
		observability.String("proof_cid", req.ProofCID),
		observability.String("error", req.Error))

	// Case 1: Task failed (execution failed or aggregator submission failed)
	if !req.ExecutionSuccessful || !req.AggregatorSubmitted {
		if err := tm.dbClient.UpdateTaskAggregatorFailed(ctx, req.TaskID, req.Error, req.ExecutionTxHash, req.ProofCID); err != nil {
			tm.logger.Error(ctx, "Failed to update task failure in database",
				observability.Int64("task_id", req.TaskID),
				observability.Error(err))
			return &types.ReportTaskStatusResponse{
				Success: false,
				Message: fmt.Sprintf("failed to update task failure: %v", err),
			}, nil
		}

		// Move task to failed stream
		_ = tm.taskStreamManager.MarkTaskFailed(ctx, req.TaskID, req.Error)

		tm.logger.Info(ctx, "Task failure recorded",
			observability.Int64("task_id", req.TaskID),
			observability.String("keeper_address", req.KeeperAddress),
			observability.Bool("execution_successful", req.ExecutionSuccessful),
			observability.Bool("aggregator_submitted", req.AggregatorSubmitted),
			observability.String("execution_tx_hash", req.ExecutionTxHash),
			observability.String("error", req.Error))

		return &types.ReportTaskStatusResponse{
			Success: true,
			Message: "Task failure recorded",
		}, nil
	}

	// Case 2: Task succeeded (both execution and aggregator submission succeeded)
	// Update task status to pending confirmation (waiting for on-chain event)
	if err := tm.dbClient.UpdateTaskAggregatorSubmitted(ctx, req.TaskID, req.ExecutionTxHash, req.ProofCID); err != nil {
		tm.logger.Error(ctx, "Failed to update task success in database",
			observability.Int64("task_id", req.TaskID),
			observability.Error(err))
		return &types.ReportTaskStatusResponse{
			Success: false,
			Message: fmt.Sprintf("failed to update task success: %v", err),
		}, nil
	}

	tm.logger.Info(ctx, "Task success recorded, pending on-chain confirmation",
		observability.Int64("task_id", req.TaskID),
		observability.String("keeper_address", req.KeeperAddress),
		observability.String("execution_tx_hash", req.ExecutionTxHash),
		observability.String("proof_cid", req.ProofCID))

	return &types.ReportTaskStatusResponse{
		Success: true,
		Message: "Task status updated, pending on-chain confirmation",
	}, nil
}

// SetRPCServer sets the RPC server for graceful shutdown
func (tm *TaskManager) SetRPCServer(server interface {
	Stop(ctx context.Context) error
}) {
	tm.rpcServer = server
}

// startMetricsUpdateWorker periodically updates metrics from Redis client
func (tm *TaskManager) startMetricsUpdateWorker() {
	tm.logger.Info(tm.ctx, "Starting metrics update worker")

	for {
		select {
		case <-tm.ctx.Done():
			tm.logger.Info(tm.ctx, "Metrics update worker stopping")
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
			tm.logger.Error(tm.ctx, "Panic in metrics update", observability.Any("panic", r))
		}
	}()

	// Update Redis client metrics
	operationMetrics := tm.redisClient.GetOperationMetrics()
	metrics.UpdateRedisClientMetrics(operationMetrics)

	// Update stream length metrics
	taskStreamInfo := tm.taskStreamManager.GetStreamInfo(tm.ctx)
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
	tm.logger.Debug(tm.ctx, "Performing TaskManager health check")

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
		healthStatus["task_streams"] = tm.taskStreamManager.GetStreamInfo(tm.ctx)
	}

	return healthStatus
}

// Close gracefully shuts down the TaskManager
func (tm *TaskManager) Close() error {
	tm.logger.Info(tm.ctx, "Closing TaskManager...")

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
		tm.logger.Info(tm.ctx, "All background workers stopped successfully")
	case <-shutdownCtx.Done():
		tm.logger.Warn(tm.ctx, "Timeout waiting for background workers to stop")
	}

	// Stop event listener
	if err := tm.eventListener.Stop(tm.ctx); err != nil {
		tm.logger.Error(tm.ctx, "Error stopping event listener", observability.Error(err))
	}

	if err := tm.testEventListener.Stop(tm.ctx); err != nil {
		tm.logger.Error(tm.ctx, "Error stopping test event listener", observability.Error(err))
	}

	// Stop metrics ticker
	if tm.metricsUpdateTicker != nil {
		tm.metricsUpdateTicker.Stop()
	}

	// Close stream managers
	var errors []error

	if tm.taskStreamManager != nil {
		if err := tm.taskStreamManager.Close(tm.ctx); err != nil {
			tm.logger.Error(tm.ctx, "Failed to close TaskStreamManager", observability.Error(err))
			errors = append(errors, fmt.Errorf("task stream manager: %w", err))
		}
		// TaskStreamManager handles Redis client closure, so we don't need to close it again
		tm.redisClient = nil
	} else {
		// Only close Redis client if TaskStreamManager is nil
		if tm.redisClient != nil {
			if err := tm.redisClient.Close(); err != nil {
				tm.logger.Error(tm.ctx, "Failed to close Redis client", observability.Error(err))
				errors = append(errors, fmt.Errorf("redis client: %w", err))
			}
		}
	}

	metrics.ServiceStatus.WithLabelValues("task_manager").Set(0)

	if len(errors) > 0 {
		tm.logger.Warn(tm.ctx, "Some non-critical errors occurred during shutdown", observability.Int("error_count", len(errors)))
		for i, err := range errors {
			tm.logger.Debug(tm.ctx, "Shutdown error", observability.Int("index", i), observability.Error(err))
		}
		// Don't return error for cleanup issues - shutdown was successful
	}

	tm.logger.Info(tm.ctx, "TaskManager closed successfully")
	return nil
}
