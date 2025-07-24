package tasks

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/taskmanager/config"
	"github.com/trigg3rX/triggerx-backend/internal/taskmanager/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/client/aggregator"
	"github.com/trigg3rX/triggerx-backend/pkg/client/dbserver"
	redisClient "github.com/trigg3rX/triggerx-backend/pkg/client/redis"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type TaskStreamManager struct {
	client         redisClient.RedisClientInterface
	aggClient      *aggregator.AggregatorClient
	dbClient       *dbserver.DBServerClient
	logger         logging.Logger
	consumerGroups map[string]bool
	mu             sync.RWMutex
	startTime      time.Time
	batchProcessor *TaskBatchProcessor
}

func NewTaskStreamManager(logger logging.Logger, client redisClient.RedisClientInterface) (*TaskStreamManager, error) {
	logger.Info("Initializing TaskStreamManager...")

	// Initialize aggregator client
	aggClientCfg := aggregator.AggregatorClientConfig{
		AggregatorRPCUrl: config.GetAggregatorRPCUrl(),
		SenderPrivateKey: config.GetRedisSigningKey(),
		SenderAddress:    config.GetRedisSigningAddress(),
		RetryAttempts:    config.GetMaxRetries(),
		RetryDelay:       config.GetRetryDelay(),
		RequestTimeout:   config.GetRequestTimeout(),
	}
	aggClient, err := aggregator.NewAggregatorClient(logger, aggClientCfg)
	if err != nil {
		logger.Error("Failed to initialize aggregator client", "error", err)
		return nil, fmt.Errorf("failed to initialize aggregator client: %w", err)
	}

	// Initialize dbserver client
	dbserverClient, err := dbserver.NewDBServerClient(logger, config.GetDBServerRPCUrl())
	if err != nil {
		logger.Error("Failed to create DBServer client for TaskStreamManager", "error", err)
		metrics.ServiceStatus.WithLabelValues("task_stream_manager").Set(0)
		return nil, fmt.Errorf("failed to create dbserver client: %w", err)
	}

	tsm := &TaskStreamManager{
		client:         client,
		aggClient:      aggClient,
		dbClient:       dbserverClient,
		logger:         logger,
		consumerGroups: make(map[string]bool),
		startTime:      time.Now(),
	}

	// Initialize batch processor for improved performance
	batchSize := config.GetTaskBatchSize()
	batchTimeout := config.GetTaskBatchTimeout()
	tsm.batchProcessor = NewTaskBatchProcessor(batchSize, batchTimeout, tsm)

	logger.Info("TaskStreamManager initialized successfully")
	metrics.ServiceStatus.WithLabelValues("task_stream_manager").Set(1)
	return tsm, nil
}

func (tsm *TaskStreamManager) Initialize() error {
	tsm.logger.Info("Initializing task streams...")

	ctx, cancel := context.WithTimeout(context.Background(), config.GetInitializationTimeout())
	defer cancel()

	// Initialize task streams with specific expiration rules
	streamConfigs := map[string]time.Duration{
		TasksReadyStream:      0,                 // No expiration until moved
		TasksProcessingStream: 0,                 // Managed by timeout worker
		TasksCompletedStream:  TasksCompletedTTL, // 1 hour expiration
		TasksFailedStream:     TasksFailedTTL,    // 7 days for debugging
		TasksRetryStream:      TasksRetryTTL,     // 24 hours
	}

	for stream, ttl := range streamConfigs {
		tsm.logger.Debug("Creating stream", "stream", stream, "ttl", ttl)
		if err := tsm.client.CreateStreamIfNotExists(ctx, stream, ttl); err != nil {
			tsm.logger.Error("Failed to initialize stream",
				"stream", stream,
				"error", err,
				"ttl", ttl)
			return fmt.Errorf("failed to initialize stream %s: %w", stream, err)
		}
		tsm.logger.Info("Stream initialized successfully", "stream", stream, "ttl", ttl)
	}

	// Register consumer groups for task processing
	if err := tsm.RegisterConsumerGroup(TasksReadyStream, "task-processors"); err != nil {
		return fmt.Errorf("failed to register task-processors group: %w", err)
	}

	tsm.logger.Info("All task streams initialized successfully")

	// Start the batch processor
	tsm.batchProcessor.Start()
	tsm.logger.Info("Batch processor started successfully")

	return nil
}

// MarkTaskCompleted marks a task as completed
func (tsm *TaskStreamManager) MarkTaskCompleted(taskID int64, performerData types.PerformerData) error {
	tsm.logger.Info("Marking task as completed",
		"task_id", taskID,
		"performer_id", performerData.KeeperID)

	// Find and move task from processing to completed
	task, err := tsm.findTaskInProcessing(taskID)
	if err != nil {
		return fmt.Errorf("failed to find task in processing: %w", err)
	}

	task.CompletedAt = &[]time.Time{time.Now()}[0]

	// Add to completed stream
	err = tsm.addTaskToStream(TasksCompletedStream, task)
	if err != nil {
		return fmt.Errorf("failed to add to completed stream: %w", err)
	}

	// Remove from processing stream (acknowledge)
	// Note: In a real implementation, we'd need to track the processing message ID
	tsm.logger.Info("Task marked as completed successfully", "task_id", taskID)
	metrics.TasksAddedToStreamTotal.WithLabelValues("completed", "success").Inc()

	return nil
}

// findTaskInProcessing finds a specific task in the processing stream
func (tsm *TaskStreamManager) findTaskInProcessing(taskID int64) (*TaskStreamData, error) {
	tasks, _, err := tsm.readTasksFromStreamWithIDs(TasksProcessingStream, "task-finder", "finder", 100)
	if err != nil {
		return nil, err
	}

	for _, task := range tasks {
		if task.SendTaskDataToKeeper.TaskID[0] == taskID {
			return &task, nil
		}
	}

	return nil, fmt.Errorf("task %d not found in processing stream", taskID)
}

// RegisterConsumerGroup registers a consumer group for a stream
func (tsm *TaskStreamManager) RegisterConsumerGroup(stream string, group string) error {
	tsm.mu.Lock()
	defer tsm.mu.Unlock()

	key := fmt.Sprintf("%s:%s", stream, group)
	if _, exists := tsm.consumerGroups[key]; exists {
		// tsm.logger.Debug("Consumer group already exists", "stream", stream, "group", group)
		return nil
	}

	tsm.logger.Info("Registering consumer group", "stream", stream, "group", group)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := tsm.client.CreateConsumerGroup(ctx, stream, group); err != nil {
		tsm.logger.Error("Failed to create consumer group",
			"stream", stream,
			"group", group,
			"error", err)
		return fmt.Errorf("failed to create consumer group for %s: %w", stream, err)
	}

	tsm.consumerGroups[key] = true
	tsm.logger.Info("Consumer group created successfully", "stream", stream, "group", group)
	return nil
}

// GetStreamInfo returns information about task streams
func (tsm *TaskStreamManager) GetStreamInfo() map[string]interface{} {
	tsm.logger.Debug("Getting stream information")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	streamLengths := make(map[string]int64)
	streams := []string{TasksReadyStream, TasksRetryStream, TasksProcessingStream, TasksCompletedStream, TasksFailedStream}

	for _, stream := range streams {
		length, err := tsm.client.XLen(ctx, stream)
		if err != nil {
			tsm.logger.Warn("Failed to get stream length",
				"stream", stream,
				"error", err)
			length = -1
		}
		streamLengths[stream] = length

		// Update stream length metrics
		switch stream {
		case TasksReadyStream:
			metrics.TaskStreamLengths.WithLabelValues("ready").Set(float64(length))
		case TasksRetryStream:
			metrics.TaskStreamLengths.WithLabelValues("retry").Set(float64(length))
		case TasksProcessingStream:
			metrics.TaskStreamLengths.WithLabelValues("processing").Set(float64(length))
		case TasksCompletedStream:
			metrics.TaskStreamLengths.WithLabelValues("completed").Set(float64(length))
		case TasksFailedStream:
			metrics.TaskStreamLengths.WithLabelValues("failed").Set(float64(length))
		}
	}

	info := map[string]interface{}{
		"available":            tsm.client != nil,
		"max_length":           10000, // Default value, can be made configurable
		"tasks_processing_ttl": TasksProcessingTTL.String(),
		"tasks_completed_ttl":  TasksCompletedTTL.String(),
		"tasks_failed_ttl":     TasksFailedTTL.String(),
		"tasks_retry_ttl":      TasksRetryTTL.String(),
		"stream_lengths":       streamLengths,
		"max_retries":          MaxRetryAttempts,
		"retry_backoff":        RetryBackoffBase.String(),
		"consumer_groups":      len(tsm.consumerGroups),
	}

	tsm.logger.Debug("Stream information retrieved", "info", info)
	return info
}

// GetBatchProcessorStats returns statistics about the batch processor
func (tsm *TaskStreamManager) GetBatchProcessorStats() map[string]interface{} {
	if tsm.batchProcessor == nil {
		return map[string]interface{}{
			"status": "not_initialized",
		}
	}
	return tsm.batchProcessor.GetBatchStats()
}

func (tsm *TaskStreamManager) Close() error {
	tsm.logger.Info("Closing TaskStreamManager")

	// Stop the batch processor
	if tsm.batchProcessor != nil {
		tsm.batchProcessor.Stop()
		tsm.logger.Info("Batch processor stopped")
	}

	err := tsm.client.Close()
	if err != nil {
		tsm.logger.Error("Failed to close Redis client", "error", err)
		return err
	}

	tsm.logger.Info("TaskStreamManager closed successfully")
	return nil
}
