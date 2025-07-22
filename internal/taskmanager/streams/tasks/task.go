package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/trigg3rX/triggerx-backend/internal/taskmanager/config"
	"github.com/trigg3rX/triggerx-backend/internal/taskmanager/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/taskmanager/streams/performers"
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

// ReceiveTaskFromScheduler is the main entry point for schedulers to submit tasks
func (tsm *TaskStreamManager) ReceiveTaskFromScheduler(request *SchedulerTaskRequest) (*types.PerformerData, error) {
	tsm.logger.Info("Receiving task from scheduler",
		"task_id", request.SendTaskDataToKeeper.TaskID,
		"scheduler_id", request.SchedulerID,
		"source", request.Source)

	// Create task stream data
	taskStreamData := TaskStreamData{
		JobID:                request.SendTaskDataToKeeper.TaskID, // Use TaskID as JobID for simple cases
		TaskDefinitionID:     request.SendTaskDataToKeeper.TargetData[0].TaskDefinitionID,
		CreatedAt:            time.Now(),
		RetryCount:           0,
		SendTaskDataToKeeper: request.SendTaskDataToKeeper,
	}

	// Add task to batch processor for improved performance
	err := tsm.batchProcessor.AddTask(&taskStreamData)
	if err != nil {
		tsm.logger.Error("Failed to add task to batch processor",
			"task_id", request.SendTaskDataToKeeper.TaskID,
			"source", request.Source,
			"error", err)
		return nil, fmt.Errorf("failed to add task to batch processor: %w", err)
	}

	// Get performer data for immediate response
	performerData := performers.GetPerformerData()
	if performerData.KeeperID == 0 {
		tsm.logger.Error("No performers available for task", "task_id", request.SendTaskDataToKeeper.TaskID)
		return nil, fmt.Errorf("no performers available")
	}

	tsm.logger.Info("Task received and added to batch processor",
		"task_id", request.SendTaskDataToKeeper.TaskID,
		"performer_id", performerData.KeeperID,
		"source", request.Source)

	return &performerData, nil
}

// StartTaskProcessor starts the main task processing loop with consumer groups
func (tsm *TaskStreamManager) StartTaskProcessor(ctx context.Context, consumerName string) {
	tsm.logger.Info("Starting task processor", "consumer_name", consumerName)

	ticker := time.NewTicker(1 * time.Second) // Process tasks every second
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			tsm.logger.Info("Task processor stopping", "consumer_name", consumerName)
			return
		case <-ticker.C:
			tsm.processReadyTasks(consumerName)
		}
	}
}

// StartTimeoutWorker monitors processing tasks for timeouts
func (tsm *TaskStreamManager) StartTimeoutWorker(ctx context.Context) {
	tsm.logger.Info("Starting task timeout worker")

	ticker := time.NewTicker(30 * time.Second) // Check timeouts every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			tsm.logger.Info("Timeout worker stopping")
			return
		case <-ticker.C:
			tsm.checkProcessingTimeouts()
		}
	}
}

// StartRetryWorker processes tasks from retry stream when they're ready
func (tsm *TaskStreamManager) StartRetryWorker(ctx context.Context) {
	tsm.logger.Info("Starting task retry worker")

	ticker := time.NewTicker(5 * time.Second) // Check retries every 5 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			tsm.logger.Info("Retry worker stopping")
			return
		case <-ticker.C:
			tsm.processRetryTasks()
		}
	}
}

// processReadyTasks processes tasks from the ready stream
func (tsm *TaskStreamManager) processReadyTasks(consumerName string) {
	tasks, messageIDs, err := tsm.readTasksFromStreamWithIDs(TasksReadyStream, "task-processors", consumerName, 10)
	if err != nil {
		tsm.logger.Error("Failed to read tasks from ready stream", "error", err)
		return
	}

	if len(tasks) == 0 {
		return // No tasks available
	}

	tsm.logger.Debug("Processing tasks from ready stream", "count", len(tasks))

	for i, task := range tasks {
		messageID := messageIDs[i]

		// Move task to processing stream
		if err := tsm.moveTaskToProcessing(task, messageID); err != nil {
			tsm.logger.Error("Failed to move task to processing",
				"task_id", task.SendTaskDataToKeeper.TaskID,
				"error", err)
			continue
		}

		// Send task to performer
		go tsm.sendTaskToPerformer(task)
	}
}

// moveTaskToProcessing moves a task from ready to processing stream
func (tsm *TaskStreamManager) moveTaskToProcessing(task TaskStreamData, messageID string) error {
	task.ProcessingStartedAt = &[]time.Time{time.Now()}[0]

	// Add to processing stream
	err := tsm.addTaskToStream(TasksProcessingStream, &task)
	if err != nil {
		return fmt.Errorf("failed to add to processing stream: %w", err)
	}

	// Acknowledge in ready stream
	if err := tsm.AckTaskProcessed(TasksReadyStream, "task-processors", messageID); err != nil {
		tsm.logger.Warn("Failed to acknowledge task in ready stream",
			"task_id", task.SendTaskDataToKeeper.TaskID,
			"message_id", messageID,
			"error", err)
	}

	tsm.logger.Debug("Task moved to processing stream",
		"task_id", task.SendTaskDataToKeeper.TaskID)

	return nil
}

// sendTaskToPerformer sends the task to the assigned performer
func (tsm *TaskStreamManager) sendTaskToPerformer(task TaskStreamData) {
	taskID := task.SendTaskDataToKeeper.TaskID

	tsm.logger.Info("Sending task to performer", "task_id", taskID)

	// Send to aggregator/performer using existing method
	jsonData, err := json.Marshal(task.SendTaskDataToKeeper)
	if err != nil {
		tsm.logger.Errorf("Failed to marshal batch task data: %v", err)
		return
	}
	dataBytes := []byte(jsonData)

	broadcastDataForPerformer := types.BroadcastDataForPerformer{
		TaskID:           task.SendTaskDataToKeeper.TaskID,
		TaskDefinitionID: task.SendTaskDataToKeeper.TargetData[0].TaskDefinitionID,
		PerformerAddress: task.SendTaskDataToKeeper.PerformerData.KeeperAddress,
		Data:             dataBytes,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	success, err := tsm.aggClient.SendTaskToPerformer(ctx, &broadcastDataForPerformer)
	if err != nil {
		tsm.logger.Error("Failed to send task to performer",
			"task_id", taskID,
			"error", err)

		// Move task to failed stream
		if moveErr := tsm.moveTaskToFailed(task, err.Error()); moveErr != nil {
			tsm.logger.Error("Failed to move task to failed stream",
				"task_id", taskID,
				"error", moveErr)
		}
		return
	}

	if success {
		tsm.logger.Info("Task sent to performer successfully", "task_id", taskID)
		metrics.TasksAddedToStreamTotal.WithLabelValues("processing", "success").Inc()
	} else {
		tsm.logger.Warn("Task sending to performer was not successful", "task_id", taskID)
		if moveErr := tsm.moveTaskToFailed(task, "performer send failed"); moveErr != nil {
			tsm.logger.Error("Failed to move task to failed stream",
				"task_id", taskID,
				"error", moveErr)
		}
	}
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

// moveTaskToFailed moves a task to the failed stream or retry stream
func (tsm *TaskStreamManager) moveTaskToFailed(task TaskStreamData, errorMsg string) error {
	task.LastError = errorMsg
	task.RetryCount++

	if task.RetryCount <= MaxRetryAttempts {
		// Move to retry stream with backoff
		delay := time.Duration(task.RetryCount) * RetryBackoffBase
		scheduledFor := time.Now().Add(delay)
		task.ScheduledFor = &scheduledFor

		err := tsm.addTaskToStream(TasksRetryStream, &task)
		if err != nil {
			return fmt.Errorf("failed to add to retry stream: %w", err)
		}

		tsm.logger.Info("Task moved to retry stream",
			"task_id", task.SendTaskDataToKeeper.TaskID,
			"retry_count", task.RetryCount,
			"scheduled_for", scheduledFor)

		metrics.TasksAddedToStreamTotal.WithLabelValues("retry", "success").Inc()
	} else {
		// Move to failed stream permanently
		err := tsm.addTaskToStream(TasksFailedStream, &task)
		if err != nil {
			return fmt.Errorf("failed to add to failed stream: %w", err)
		}

		tsm.logger.Error("Task permanently failed",
			"task_id", task.SendTaskDataToKeeper.TaskID,
			"retry_count", task.RetryCount,
			"error", errorMsg)

		metrics.TasksAddedToStreamTotal.WithLabelValues("failed", "success").Inc()
	}

	return nil
}

// checkProcessingTimeouts checks for tasks that have been processing too long
func (tsm *TaskStreamManager) checkProcessingTimeouts() {
	tsm.logger.Debug("Checking for processing timeouts")

	// Read all processing tasks (simplified - in production would use consumer groups)
	tasks, messageIDs, err := tsm.readTasksFromStreamWithIDs(TasksProcessingStream, "timeout-checker", "timeout-worker", 100)
	if err != nil {
		tsm.logger.Error("Failed to read processing tasks for timeout check", "error", err)
		return
	}

	now := time.Now()
	timeoutCount := 0

	for i, task := range tasks {
		if task.ProcessingStartedAt != nil {
			processingDuration := now.Sub(*task.ProcessingStartedAt)
			if processingDuration > TasksProcessingTTL {
				tsm.logger.Warn("Task processing timeout detected",
					"task_id", task.SendTaskDataToKeeper.TaskID,
					"processing_duration", processingDuration)

				// Move to failed/retry
				if err := tsm.moveTaskToFailed(task, "processing timeout"); err != nil {
					tsm.logger.Error("Failed to handle timeout task",
						"task_id", task.SendTaskDataToKeeper.TaskID,
						"error", err)
				} else {
					// Acknowledge the timed-out task
					messageID := messageIDs[i]
					err := tsm.AckTaskProcessed(TasksProcessingStream, "timeout-checker", messageID)
					if err != nil {
						tsm.logger.Error("Failed to acknowledge timed-out task",
							"task_id", task.SendTaskDataToKeeper.TaskID,
							"error", err)
					}
					timeoutCount++
				}
			}
		}
	}

	if timeoutCount > 0 {
		tsm.logger.Info("Processed task timeouts", "timeout_count", timeoutCount)
	}
}

// processRetryTasks processes tasks that are ready for retry
func (tsm *TaskStreamManager) processRetryTasks() {
	tasks, messageIDs, err := tsm.readTasksFromStreamWithIDs(TasksRetryStream, "retry-processor", "retry-worker", 10)
	if err != nil {
		tsm.logger.Error("Failed to read retry tasks", "error", err)
		return
	}

	now := time.Now()
	retryCount := 0

	for i, task := range tasks {
		if task.ScheduledFor == nil || task.ScheduledFor.Before(now) {
			// Task is ready for retry
			task.LastAttemptAt = &now

			// Move back to ready stream
			err := tsm.addTaskToStream(TasksReadyStream, &task)
			if err != nil {
				tsm.logger.Error("Failed to move retry task to ready",
					"task_id", task.SendTaskDataToKeeper.TaskID,
					"error", err)
				continue
			}

			// Acknowledge in retry stream
			messageID := messageIDs[i]
			err = tsm.AckTaskProcessed(TasksRetryStream, "retry-processor", messageID)
			if err != nil {
				tsm.logger.Error("Failed to acknowledge task in retry stream",
					"task_id", task.SendTaskDataToKeeper.TaskID,
					"message_id", messageID,
					"error", err)
			}

			tsm.logger.Info("Task moved from retry to ready",
				"task_id", task.SendTaskDataToKeeper.TaskID,
				"retry_count", task.RetryCount)

			retryCount++
		}
	}

	if retryCount > 0 {
		tsm.logger.Info("Processed retry tasks", "retry_count", retryCount)
	}
}

// findTaskInProcessing finds a specific task in the processing stream
func (tsm *TaskStreamManager) findTaskInProcessing(taskID int64) (*TaskStreamData, error) {
	tasks, _, err := tsm.readTasksFromStreamWithIDs(TasksProcessingStream, "task-finder", "finder", 100)
	if err != nil {
		return nil, err
	}

	for _, task := range tasks {
		if task.SendTaskDataToKeeper.TaskID == taskID {
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

// AckTaskProcessed acknowledges that a task has been processed
func (tsm *TaskStreamManager) AckTaskProcessed(stream, consumerGroup, messageID string) error {
	tsm.logger.Debug("Acknowledging task processed",
		"stream", stream,
		"consumer_group", consumerGroup,
		"message_id", messageID)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := tsm.client.XAck(ctx, stream, consumerGroup, messageID)
	if err != nil {
		tsm.logger.Error("Failed to acknowledge task",
			"stream", stream,
			"consumer_group", consumerGroup,
			"message_id", messageID,
			"error", err)
		return err
	}

	tsm.logger.Debug("Task acknowledged successfully",
		"stream", stream,
		"message_id", messageID)

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

// readTasksFromStreamWithIDs reads tasks and returns both tasks and message IDs
func (tsm *TaskStreamManager) readTasksFromStreamWithIDs(stream, consumerGroup, consumerName string, count int64) ([]TaskStreamData, []string, error) {
	if err := tsm.RegisterConsumerGroup(stream, consumerGroup); err != nil {
		return nil, nil, err
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	streams, err := tsm.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    consumerGroup,
		Consumer: consumerName,
		Streams:  []string{stream, ">"},
		Count:    count,
		Block:    time.Second,
	})

	duration := time.Since(start)

	if err != nil {
		if err == redis.Nil {
			// tsm.logger.Debug("No tasks available in stream",
			// 	"stream", stream,
			// 	"consumer_group", consumerGroup,
			// 	"duration", duration)
			return []TaskStreamData{}, []string{}, nil
		}
		tsm.logger.Error("Failed to read from stream",
			"stream", stream,
			"consumer_group", consumerGroup,
			"duration", duration,
			"error", err)
		return nil, nil, fmt.Errorf("failed to read from stream: %w", err)
	}

	var tasks []TaskStreamData
	var messageIDs []string

	for _, stream := range streams {
		for _, message := range stream.Messages {
			taskJSON, exists := message.Values["task"].(string)
			if !exists {
				tsm.logger.Warn("Message missing task data",
					"stream", stream.Stream,
					"message_id", message.ID)
				continue
			}

			var task TaskStreamData
			if err := json.Unmarshal([]byte(taskJSON), &task); err != nil {
				tsm.logger.Error("Failed to unmarshal task data",
					"stream", stream.Stream,
					"message_id", message.ID,
					"error", err)
				continue
			}

			tasks = append(tasks, task)
			messageIDs = append(messageIDs, message.ID)

			tsm.logger.Debug("Task read from stream",
				"task_id", task.SendTaskDataToKeeper.TaskID,
				"stream", stream.Stream,
				"message_id", message.ID)
		}
	}

	tsm.logger.Info("Tasks read from stream successfully",
		"stream", stream,
		"task_count", len(tasks),
		"duration", duration)

	return tasks, messageIDs, nil
}
