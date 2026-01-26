package manager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/config"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/core/events"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/core/tasks"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/database"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/database/repository"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/redis"
	keeperClient "github.com/trigg3rX/triggerx-backend/internal/taskmonitor/rpc/clients/keeper"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/rpc/clients/notify"
	"github.com/trigg3rX/triggerx-backend/pkg/ipfs"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// TaskManager orchestrates all Redis-based task management components
type TaskManager struct {
	logger              observability.Logger
	tracer              observability.Tracer
	redisClient         *redis.Client
	taskStreamManager   *tasks.TaskStreamManager
	metricsUpdateTicker *time.Ticker
	ctx                 context.Context
	cancel              context.CancelFunc
	shutdownWg          sync.WaitGroup
	startTime           time.Time
	taskRepo            repository.TaskRepository
	ipfsClient          ipfs.IPFSClient
	rpcServer           interface {
		Stop(ctx context.Context) error
	}
}

// NewTaskManager creates a new TaskManager instance
func NewTaskManager(ctx context.Context, logger observability.Logger, tracer observability.Tracer) (*TaskManager, error) {
	logger.Info(ctx, "Initializing TaskManager...")

	// Create context for managing background workers
	ctx, cancel := context.WithCancel(context.Background())

	// Create Redis client using local service-specific wrapper
	client, err := redis.NewClient(ctx, logger)
	if err != nil {
		cancel() // Clean up context on error
		logger.Error(ctx, "Failed to create Redis client for TaskManager", observability.Error(err))
		if metrics.ServiceStatus != nil {
			metrics.ServiceStatus.WithLabelValues("task_manager").Set(ctx, 0)
		}
		return nil, fmt.Errorf("failed to create redis client: %w", err)
	}

	// Set up monitoring hooks
	monitoringHooks := metrics.CreateRedisMonitoringHooks()
	client.SetMonitoringHooks(monitoringHooks)

	// Initialize database connection
	dbConn, err := database.NewConnection(logger)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize database connection: %w", err)
	}

	// Initialize task repository
	taskRepo := repository.NewTaskRepository(dbConn, logger)

	// Initialize IPFS client
	ipfsCfg := ipfs.NewConfig(config.GetPinataHost(), config.GetPinataJWT())
	ipfsClient, err := ipfs.NewClient(ipfsCfg)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize IPFS client: %w", err)
	}

	// Initialize keeper client
	keeperClientInstance := keeperClient.NewClient(logger, tracer)

	// Initialize task stream manager with the local Redis client
	taskStreamManager, err := tasks.NewTaskStreamManager(ctx, client, taskRepo, logger, tracer, keeperClientInstance)
	if err != nil {
		// Clean up resources on error
		cancel()
		if closeErr := client.Close(); closeErr != nil {
			logger.Error(ctx, "Failed to close Redis client during error cleanup", observability.Error(closeErr))
		}
		logger.Error(ctx, "Failed to create TaskStreamManager", observability.Error(err))
		return nil, fmt.Errorf("failed to create task stream manager: %w", err)
	}

	tm := &TaskManager{
		logger:              logger,
		tracer:              tracer,
		redisClient:         client,
		taskStreamManager:   taskStreamManager,
		metricsUpdateTicker: time.NewTicker(config.GetMetricsUpdateInterval()),
		ctx:                 ctx,
		cancel:              cancel,
		startTime:           time.Now(),
		taskRepo:            taskRepo,
		ipfsClient:          ipfsClient,
	}

	logger.Info(ctx, "TaskManager initialized successfully",
		observability.Duration("metrics_update_interval", config.GetMetricsUpdateInterval()))

	if metrics.ServiceStatus != nil {
		metrics.ServiceStatus.WithLabelValues("task_manager").Set(ctx, 1)
	}
	return tm, nil
}

// Initialize initializes all stream managers and starts background workers
func (tm *TaskManager) Initialize() error {
	tm.logger.Info(tm.ctx, "Initializing TaskManager components...")

	// Initialize task streams
	if err := tm.taskStreamManager.Initialize(tm.ctx); err != nil {
		return fmt.Errorf("failed to initialize task stream manager: %w", err)
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
func (tm *TaskManager) ReportTaskStatus(ctx context.Context, req *types.ReportTaskExecutionStatusRequest) (*types.ReportTaskExecutionStatusResponse, error) {
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
		if err := tm.taskRepo.UpdateTaskAggregatorFailed(ctx, req.TaskID, req.Error, req.ExecutionTxHash, req.ProofCID); err != nil {
			tm.logger.Error(ctx, "Failed to update task failure in database",
				observability.Int64("task_id", req.TaskID),
				observability.Error(err))
			return &types.ReportTaskExecutionStatusResponse{
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

		return &types.ReportTaskExecutionStatusResponse{
			Success: true,
			Message: "Task failure recorded",
		}, nil
	}

	// Case 2: Task succeeded (both execution and aggregator submission succeeded)
	// Update task status to pending confirmation (waiting for on-chain event)
	if err := tm.taskRepo.UpdateTaskAggregatorSubmitted(ctx, req.TaskID, req.ExecutionTxHash, req.ProofCID); err != nil {
		tm.logger.Error(ctx, "Failed to update task success in database",
			observability.Int64("task_id", req.TaskID),
			observability.Error(err))
		return &types.ReportTaskExecutionStatusResponse{
			Success: false,
			Message: fmt.Sprintf("failed to update task success: %v", err),
		}, nil
	}

	// Mark task as executed and add to executed stream with timeout
	// Timeout: 15 minutes for validation (from TasksExecutedTTL constant)
	if err := tm.taskStreamManager.MarkTaskExecuted(ctx, req.TaskID, types.TasksExecutedTTL); err != nil {
		tm.logger.Error(ctx, "Failed to mark task as executed",
			observability.Int64("task_id", req.TaskID),
			observability.Error(err))
		// Don't fail the request, but log the error
	}

	tm.logger.Info(ctx, "Task success recorded, pending on-chain confirmation",
		observability.Int64("task_id", req.TaskID),
		observability.String("keeper_address", req.KeeperAddress),
		observability.String("execution_tx_hash", req.ExecutionTxHash),
		observability.String("proof_cid", req.ProofCID))

	return &types.ReportTaskExecutionStatusResponse{
		Success: true,
		Message: "Task status updated, pending on-chain confirmation",
	}, nil
}

// ReportConsensusEvent handles consensus event reports from eventmonitor (TaskSubmitted or TaskRejected)
// The request now contains IPFS data (with trace context) directly from eventmonitor
func (tm *TaskManager) ReportConsensusEvent(ctx context.Context, req *types.ReportTaskConsensusStatusRequest) (*types.ReportTaskConsensusStatusResponse, error) {
	// Extract task ID from IPFS data (ActionData has single task ID)
	taskID := int64(0)
	if req.IPFSData != nil && req.IPFSData.ActionData != nil {
		taskID = req.IPFSData.ActionData.TaskID
	}

	tm.logger.Info(ctx, "Received consensus event report from eventmonitor",
		observability.String("tx_hash", req.TaskSubmissionTxHash),
		observability.Bool("is_accepted", req.IsAccepted),
		observability.Int64("task_id", taskID))

	// Continue the trace from IPFS data if available
	if req.IPFSData != nil && req.IPFSData.TraceID != "" {
		ctx = observability.ContinueTrace(ctx, req.IPFSData.TraceID, req.IPFSData.SpanID)
	}

	// Create span for consensus event processing
	ctx, span := tm.tracer.Start(ctx, "task.consensus.process",
		observability.WithSpanKind(trace.SpanKindServer),
		observability.WithAttributes(
			attribute.String("tx.hash", req.TaskSubmissionTxHash),
			attribute.Bool("is.accepted", req.IsAccepted),
			attribute.Int64("task.id", taskID),
		),
	)
	defer span.End()

	// Create a task handler instance to process the event
	taskHandler := tm.createTaskEventHandler()
	// Process the consensus event with IPFS data
	if err := taskHandler.ProcessConsensusEventFromIPFS(ctx, req.TaskSubmissionTxHash, req.IsAccepted, req.IPFSData, req.IPFSCID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to process consensus event")
		return &types.ReportTaskConsensusStatusResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	span.SetStatus(codes.Ok, "consensus event processed")
	return &types.ReportTaskConsensusStatusResponse{
		Success: true,
		Message: "Consensus event processed",
	}, nil
}

// createTaskEventHandler creates a TaskEventHandler instance for processing events
func (tm *TaskManager) createTaskEventHandler() *events.TaskEventHandler {
	// Create notifier similar to how it's done in the event listener
	notifier := notify.NewCompositeNotifier(tm.logger, notify.NewWebhookNotifier(tm.logger), notify.NewSMTPNotifier(tm.logger))

	return events.NewTaskEventHandler(tm.logger, tm.tracer, tm.taskRepo, tm.ipfsClient, tm.taskStreamManager, notifier)
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
				if metrics.TaskStreamLengths != nil {
					metrics.TaskStreamLengths.WithLabelValues("ready").Set(tm.ctx, float64(length))
				}
			case "tasks:processing":
				if metrics.TaskStreamLengths != nil {
					metrics.TaskStreamLengths.WithLabelValues("processing").Set(tm.ctx, float64(length))
				}
			case "tasks:completed":
				if metrics.TaskStreamLengths != nil {
					metrics.TaskStreamLengths.WithLabelValues("completed").Set(tm.ctx, float64(length))
				}
			case "tasks:failed":
				if metrics.TaskStreamLengths != nil {
					metrics.TaskStreamLengths.WithLabelValues("failed").Set(tm.ctx, float64(length))
				}
			case "tasks:retry":
				if metrics.TaskStreamLengths != nil {
					metrics.TaskStreamLengths.WithLabelValues("retry").Set(tm.ctx, float64(length))
				}
			}
		}
	}

	// Update connection status
	connectionStatus := tm.redisClient.GetConnectionStatus()
	if connectionStatus != nil && !connectionStatus.IsRecovering {
		if metrics.RedisConnectionHealth != nil {
			metrics.RedisConnectionHealth.WithLabelValues("main").Set(tm.ctx, 1)
		}
	}
}

// GetTaskStreamManager returns the task stream manager
func (tm *TaskManager) GetTaskStreamManager() *tasks.TaskStreamManager {
	return tm.taskStreamManager
}

// GetTaskRepository returns the task repository
func (tm *TaskManager) GetTaskRepository() repository.TaskRepository {
	return tm.taskRepo
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

	if metrics.ServiceStatus != nil {
		metrics.ServiceStatus.WithLabelValues("task_manager").Set(tm.ctx, 0)
	}

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
