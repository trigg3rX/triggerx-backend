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
	notifier          notify.Notifier
	rpcServer           interface {
		Stop(ctx context.Context) error
	}
}

// NewTaskManager creates a new TaskManager instance
func NewTaskManager(ctx context.Context, logger observability.Logger, tracer observability.Tracer, notifier notify.Notifier) (*TaskManager, error) {
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
		notifier:            notifier,
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

// ReportTaskExecutionStatus handles task execution status reports from keepers
// This is called after the aggregator submission attempt (regardless of success or failure)
// IPFSDataCID contains all execution data (task data, action data, proof, signatures)
func (tm *TaskManager) ReportTaskExecutionStatus(ctx context.Context, req *types.ReportTaskExecutionStatusRequest) (*types.ReportTaskExecutionStatusResponse, error) {
	tm.logger.Info(ctx, "Received task status report",
		observability.Int64("task_id", req.TaskID),
		observability.String("keeper_address", req.KeeperAddress),
		observability.Bool("execution_successful", req.ExecutionSuccessful),
		observability.Bool("aggregator_submitted", req.AggregatorSubmitted),
		observability.String("error", req.Error))

	ctx, span := tm.tracer.Start(ctx, "task.execution.process",
		observability.WithSpanKind(trace.SpanKindServer),
		observability.WithAttributes(
			attribute.Int64("task.id", req.TaskID),
			attribute.String("keeper.address", req.KeeperAddress),
			attribute.Bool("execution.successful", req.ExecutionSuccessful),
			attribute.Bool("aggregator.submitted", req.AggregatorSubmitted),
		),
	)
	defer span.End()

	var ipfsData *types.IPFSData
	var err error
	if req.IPFSDataCID != "" {
		ipfsData, err = tm.fetchIPFSData(ctx, req.IPFSDataCID)
		if err != nil {
			tm.logger.Debug(ctx, "Failed to fetch IPFS data",
				observability.String("ipfs_cid", req.IPFSDataCID),
				observability.Error(err))
		}
	} else {
		ipfsData = nil
	}

	if err := tm.taskRepo.UpdateTaskExecutionData(ctx, req, ipfsData); err != nil {
		tm.logger.Error(ctx, "Failed to update task execution data in database",
			observability.Int64("task_id", req.TaskID),
			observability.Error(err))
		return &types.ReportTaskExecutionStatusResponse{
			Success: false,
			Message: fmt.Sprintf("failed to update task execution data: %v", err),
		}, nil
	}

	span.SetStatus(codes.Ok, "execution event processed")
	tm.logger.Info(ctx, "Execution event processed successfully",
		observability.Int64("task_id", req.TaskID),
		observability.String("tx_hash", req.ExecutionTxHash),
		observability.Bool("execution_successful", req.ExecutionSuccessful))


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
		observability.String("ipfs_data_cid", req.IPFSDataCID))

	return &types.ReportTaskExecutionStatusResponse{
		Success: true,
		Message: "Task status updated, pending on-chain confirmation",
	}, nil
}

// ReportTaskConsensusStatus handles consensus event reports from eventmonitor (TaskSubmitted or TaskRejected)
// The request now contains IPFS data (with trace context) directly from eventmonitor
func (tm *TaskManager) ReportTaskConsensusStatus(ctx context.Context, req *types.ReportTaskConsensusStatusRequest) (*types.ReportTaskConsensusStatusResponse, error) {
	tm.logger.Info(ctx, "Received consensus event report from eventmonitor",
		observability.String("tx_hash", req.TaskSubmissionTxHash),
		observability.Bool("is_accepted", req.IsAccepted),
		observability.Int64("task_id", req.TaskID),
		observability.Int64("task_number", req.TaskNumber))

	// Create span for consensus event processing
	ctx, span := tm.tracer.Start(ctx, "task.consensus.process",
		observability.WithSpanKind(trace.SpanKindServer),
		observability.WithAttributes(
			attribute.String("tx.hash", req.TaskSubmissionTxHash),
			attribute.Bool("is.accepted", req.IsAccepted),
			attribute.Int64("task.id", req.TaskID),
			attribute.Int64("task.number", req.TaskNumber),
		),
	)
	defer span.End()

	// Move task from dispatched to completed stream
	streamHandled := true
	if err := tm.taskStreamManager.MoveTaskToCompleted(ctx, req.TaskID); err != nil {
		span.RecordError(err, observability.WithErrorAttributes(
			attribute.String("error.type", "stream_move_failed"),
		))
		tm.logger.Error(ctx, "Failed to move task to validated stream", observability.Error(err))
		streamHandled = false
		// Stream move failure is critical - task should be in validated stream
		// But continue to attempt DB update anyway
	}

	// Update task submission data in database
	if err := tm.taskRepo.UpdateTaskSubmissionData(ctx, req); err != nil {
		span.RecordError(err, observability.WithErrorAttributes(
			attribute.String("error.type", "database_update_failed"),
		))
		tm.logger.Error(ctx, "Failed to update task submission data in database", observability.Error(err))
		// Don't return error - task is already validated in stream
		// DB update can be retried later if needed, but task should not be rebroadcasted
	} else {
		span.SetStatus(codes.Ok, "task validated and database updated")

		// Schedule IPFS file deletion after 6 hours delay
		// This happens after data is downloaded, DB is updated, and stream is handled
		// Only schedule deletion if both stream handling and DB update succeeded
		if streamHandled && req.IPFSDataCID != "" {
			tm.scheduleIPFSDeletion(req.IPFSDataCID, req.TaskID)
		} else if !streamHandled {
			tm.logger.Warn(ctx, "Skipping IPFS deletion scheduling - stream handling failed",
				observability.String("ipfs_cid", req.IPFSDataCID),
				observability.Int64("task_id", req.TaskID))
		}
	}

	span.AddEvent("task.data.updated", observability.WithEventAttributes(
		attribute.String("database.table", "tasks"),
	))

	// Notify user about task completion/rejection
	if tm.notifier != nil {
		email, err := tm.taskRepo.GetUserEmailByTaskID(ctx, req.TaskID)
		if err != nil {
			tm.logger.Warn(ctx, "Could not fetch user email for task", observability.Int64("task_id", req.TaskID), observability.Error(err))
		} else if email != "" {
			payload := notify.TaskStatusPayload{
				TaskID:          req.TaskID,
				JobID:           0,
				Status:          "completed",
				IsAccepted:      req.IsAccepted,
				SubmissionTx:    req.TaskSubmissionTxHash,
				// ExecutionTxHash: req.ExecutionTxHash,
				ProofOfTask:     req.IPFSDataCID,
				// OccurredAt:      req.ExecutedAt,
			}
			if !req.IsAccepted {
				payload.Status = "failed"
			}
			if err := tm.notifier.NotifyTaskStatus(context.Background(), email, payload); err != nil {
				tm.logger.Warn(ctx, "Failed to notify user", observability.String("email", email), observability.Int64("task_id", req.TaskID), observability.Error(err))
			}
		}
	}

	span.SetStatus(codes.Ok, "consensus event processed")
	return &types.ReportTaskConsensusStatusResponse{
		Success: true,
		Message: "Consensus event processed",
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

// FetchIPFSData fetches IPFS data using the provided CID
func (tm *TaskManager) fetchIPFSData(ctx context.Context, ipfsCID string) (*types.IPFSData, error) {
	ipfsData, err := tm.ipfsClient.Fetch(ctx, ipfsCID)
	if err != nil {
		tm.logger.Error(ctx, "Failed to fetch IPFS data",
			observability.String("ipfs_cid", ipfsCID),
			observability.Error(err))
		return nil, fmt.Errorf("failed to fetch IPFS data: %w", err)
	}
	return &ipfsData, nil
}

// scheduleIPFSDeletion schedules IPFS file deletion after a 6-hour delay
// This is called after data is downloaded, DB is updated, and stream is handled
// Uses a detached context so deletion proceeds even if the request context is cancelled
func (tm *TaskManager) scheduleIPFSDeletion(ipfsCID string, taskID int64) {
	const deletionDelay = 6 * time.Hour

	// Start a goroutine to handle delayed deletion
	go func() {
		// Use a background context that is detached from the request context
		// This ensures deletion proceeds even if the original request context is cancelled
		deleteCtx := context.Background()

		// Wait for the delay - no need to check for context cancellation
		time.Sleep(deletionDelay)

		// Attempt to delete the IPFS file
		if err := tm.ipfsClient.Delete(deleteCtx, ipfsCID); err != nil {
			tm.logger.Error(deleteCtx, "Failed to delete IPFS file",
				observability.String("ipfs_cid", ipfsCID),
				observability.Int64("task_id", taskID),
				observability.Error(err))
		} else {
			tm.logger.Info(deleteCtx, "Successfully deleted IPFS file",
				observability.String("ipfs_cid", ipfsCID),
				observability.Int64("task_id", taskID))
		}
	}()
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
