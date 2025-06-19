package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/trigg3rX/triggerx-backend/internal/redis/config"
	"github.com/trigg3rX/triggerx-backend/internal/redis/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type TaskStreamManager struct {
	client         *Client
	logger         logging.Logger
	consumerGroups map[string]bool
}

func NewTaskStreamManager(logger logging.Logger) (*TaskStreamManager, error) {
	logger.Info("Initializing TaskStreamManager...")

	if !config.IsRedisAvailable() {
		logger.Error("Redis not available for TaskStreamManager initialization")
		metrics.ServiceStatus.WithLabelValues("task_stream_manager").Set(0)
		return nil, fmt.Errorf("redis not available")
	}

	client, err := NewRedisClient(logger)
	if err != nil {
		logger.Error("Failed to create Redis client for TaskStreamManager", "error", err)
		metrics.ServiceStatus.WithLabelValues("task_stream_manager").Set(0)
		return nil, fmt.Errorf("failed to create redis client: %w", err)
	}

	tsm := &TaskStreamManager{
		client:         client,
		logger:         logger,
		consumerGroups: make(map[string]bool),
	}

	logger.Info("TaskStreamManager initialized successfully",
		"redis_type", config.GetRedisType(),
		"max_stream_length", config.GetStreamMaxLen(),
		"task_stream_ttl", config.GetTaskStreamTTL())

	metrics.ServiceStatus.WithLabelValues("task_stream_manager").Set(1)
	return tsm, nil
}

func (tsm *TaskStreamManager) Initialize() error {
	tsm.logger.Info("Initializing task streams...")

	streams := []string{
		TasksReadyStream, TasksRetryStream, TasksProcessingStream,
		TasksCompletedStream, TasksFailedStream,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, stream := range streams {
		tsm.logger.Debug("Creating stream if not exists", "stream", stream)
		if err := tsm.client.CreateStreamIfNotExists(ctx, stream, config.GetTaskStreamTTL()); err != nil {
			tsm.logger.Error("Failed to initialize stream",
				"stream", stream,
				"error", err,
				"ttl", config.GetTaskStreamTTL())
			return fmt.Errorf("failed to initialize stream %s: %w", stream, err)
		}
		tsm.logger.Info("Stream initialized successfully", "stream", stream)
	}

	tsm.logger.Info("All task streams initialized successfully", "stream_count", len(streams))
	return nil
}

func (tsm *TaskStreamManager) RegisterConsumerGroup(stream string, group string) error {
	key := fmt.Sprintf("%s:%s", stream, group)
	if _, exists := tsm.consumerGroups[key]; exists {
		tsm.logger.Debug("Consumer group already exists", "stream", stream, "group", group)
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

// Ready to be sent to the performer
func (tsm *TaskStreamManager) AddTaskToReadyStream(task *TaskStreamData) (types.PerformerData, error) {
	tsm.logger.Info("Adding task to ready stream",
		"task_id", task.TaskID,
		"job_id", task.JobID,
		"manager_id", task.ManagerID)

	performerData := GetPerformerData()
	if performerData.KeeperID == 0 {
		tsm.logger.Error("No performers available for task", "task_id", task.TaskID)
		return types.PerformerData{}, fmt.Errorf("no performers available")
	}

	// Update task with performer information
	task.PerformerID = int64(performerData.KeeperID)

	err := tsm.addTaskToStream(TasksReadyStream, task)
	if err != nil {
		tsm.logger.Error("Failed to add task to ready stream",
			"task_id", task.TaskID,
			"performer_id", performerData.KeeperID,
			"error", err)
		return types.PerformerData{}, err
	}

	tsm.logger.Info("Task added to ready stream successfully",
		"task_id", task.TaskID,
		"performer_id", performerData.KeeperID,
		"performer_address", performerData.KeeperAddress)

	return performerData, nil
}

// Processing by the performer
func (tsm *TaskStreamManager) AddTaskToProcessingStream(task *TaskStreamData) error {
	tsm.logger.Info("Moving task to processing stream",
		"task_id", task.TaskID,
		"job_id", task.JobID,
		"performer_id", task.PerformerID)

	// Track transition metrics
	metrics.TaskReadyToProcessingTotal.Inc()
	start := time.Now()

	err := tsm.addTaskToStream(TasksProcessingStream, task)
	if err != nil {
		tsm.logger.Error("Failed to add task to processing stream",
			"task_id", task.TaskID,
			"error", err)
		return err
	}

	// Record transition duration
	metrics.TaskLifecycleTransitionDuration.WithLabelValues("ready", "processing").Observe(time.Since(start).Seconds())

	tsm.logger.Info("Task moved to processing stream successfully",
		"task_id", task.TaskID,
		"processing_time", time.Now().Format(time.RFC3339))

	return nil
}

// Completed by the performer, sent to the validators
func (tsm *TaskStreamManager) AddTaskToCompletedStream(task *TaskStreamData) error {
	tsm.logger.Info("Moving task to completed stream",
		"task_id", task.TaskID,
		"job_id", task.JobID,
		"performer_id", task.PerformerID)

	// Track completion metrics
	metrics.TaskProcessingToCompletedTotal.Inc()
	start := time.Now()

	err := tsm.addTaskToStream(TasksCompletedStream, task)
	if err != nil {
		tsm.logger.Error("Failed to add task to completed stream",
			"task_id", task.TaskID,
			"error", err)
		return err
	}

	// Record transition duration
	metrics.TaskLifecycleTransitionDuration.WithLabelValues("processing", "completed").Observe(time.Since(start).Seconds())

	tsm.logger.Info("Task completed successfully",
		"task_id", task.TaskID,
		"completion_time", time.Now().Format(time.RFC3339),
		"total_retry_count", task.RetryCount)

	// TODO: Notify scheduler about task completion
	tsm.notifySchedulerTaskComplete(task, true)

	return nil
}

// Failed to send to the performer, sent to the retry stream
func (tsm *TaskStreamManager) AddTaskToRetryStream(task *TaskStreamData, retryReason string) error {
	tsm.logger.Warn("Adding task to retry stream",
		"task_id", task.TaskID,
		"job_id", task.JobID,
		"retry_count", task.RetryCount,
		"retry_reason", retryReason)

	task.RetryCount++
	now := time.Now()
	task.LastAttemptAt = &now

	// Exponential backoff with jitter
	baseBackoff := time.Duration(task.RetryCount) * RetryBackoffBase
	jitter := time.Duration(rand.Int63n(int64(RetryBackoffBase))) // Up to 1 base backoff
	backoffDuration := baseBackoff + jitter

	scheduledFor := now.Add(backoffDuration)
	task.ScheduledFor = &scheduledFor

	tsm.logger.Info("Task retry scheduled",
		"task_id", task.TaskID,
		"retry_count", task.RetryCount,
		"scheduled_for", scheduledFor.Format(time.RFC3339),
		"backoff_duration", backoffDuration)

	if task.RetryCount >= MaxRetryAttempts {
		tsm.logger.Error("Task exceeded max retry attempts, moving to failed stream",
			"task_id", task.TaskID,
			"retry_count", task.RetryCount,
			"max_attempts", MaxRetryAttempts)

		metrics.TaskMaxRetriesExceededTotal.Inc()
		metrics.TasksMovedToFailedStreamTotal.Inc()

		err := tsm.addTaskToStream(TasksFailedStream, task)
		if err != nil {
			tsm.logger.Error("Failed to add task to failed stream",
				"task_id", task.TaskID,
				"error", err)
			return err
		}

		// Notify scheduler about task failure
		tsm.notifySchedulerTaskComplete(task, false)

		tsm.logger.Error("Task permanently failed",
			"task_id", task.TaskID,
			"final_retry_count", task.RetryCount)

		return nil
	}

	err := tsm.addTaskToStream(TasksRetryStream, task)
	if err != nil {
		tsm.logger.Error("Failed to add task to retry stream",
			"task_id", task.TaskID,
			"error", err)
		metrics.TasksAddedToStreamTotal.WithLabelValues("retry", "failure").Inc()
		return err
	}

	metrics.TasksAddedToStreamTotal.WithLabelValues("retry", "success").Inc()
	return nil
}

func (tsm *TaskStreamManager) addTaskToStream(stream string, task *TaskStreamData) error {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	taskJSON, err := json.Marshal(task)
	if err != nil {
		tsm.logger.Error("Failed to marshal task data",
			"task_id", task.TaskID,
			"stream", stream,
			"error", err)
		return err
	}

	res, err := tsm.client.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		MaxLen: int64(config.GetStreamMaxLen()),
		Approx: true,
		Values: map[string]interface{}{
			"task":       taskJSON,
			"created_at": time.Now().Unix(),
		},
	})

	duration := time.Since(start)

	if err != nil {
		metrics.TasksAddedToStreamTotal.WithLabelValues(stream, "failure").Inc()
		tsm.logger.Error("Failed to add task to stream",
			"task_id", task.TaskID,
			"stream", stream,
			"duration", duration,
			"error", err)
		return err
	}

	metrics.TasksAddedToStreamTotal.WithLabelValues(stream, "success").Inc()
	tsm.logger.Debug("Task added to stream successfully",
		"task_id", task.TaskID,
		"stream", stream,
		"stream_id", res,
		"duration", duration,
		"task_json_size", len(taskJSON))

	return nil
}

func (tsm *TaskStreamManager) ReadTasksFromReadyStream(consumerGroup, consumerName string, count int64) ([]TaskStreamData, error) {
	tsm.logger.Debug("Reading tasks from ready stream",
		"consumer_group", consumerGroup,
		"consumer_name", consumerName,
		"count", count)

	tasks, err := tsm.readTasksFromStream(TasksReadyStream, consumerGroup, consumerName, count)
	if err != nil {
		tsm.logger.Error("Failed to read from ready stream",
			"consumer_group", consumerGroup,
			"error", err)
		return nil, err
	}

	tsm.logger.Info("Read tasks from ready stream",
		"consumer_group", consumerGroup,
		"task_count", len(tasks))

	return tasks, nil
}

func (tsm *TaskStreamManager) ReadTasksFromRetryStream(consumerGroup, consumerName string, count int64) ([]TaskStreamData, error) {
	tsm.logger.Debug("Reading tasks from retry stream",
		"consumer_group", consumerGroup,
		"consumer_name", consumerName,
		"count", count)

	tasks, err := tsm.readTasksFromStream(TasksRetryStream, consumerGroup, consumerName, count)
	if err != nil {
		tsm.logger.Error("Failed to read from retry stream",
			"consumer_group", consumerGroup,
			"error", err)
		return nil, err
	}

	now := time.Now()
	var readyTasks []TaskStreamData
	for _, task := range tasks {
		if task.ScheduledFor == nil || task.ScheduledFor.Before(now) {
			readyTasks = append(readyTasks, task)
			tsm.logger.Debug("Task ready for retry",
				"task_id", task.TaskID,
				"retry_count", task.RetryCount,
				"scheduled_for", task.ScheduledFor)
		} else {
			tsm.logger.Debug("Task not yet ready for retry",
				"task_id", task.TaskID,
				"scheduled_for", task.ScheduledFor,
				"time_remaining", task.ScheduledFor.Sub(now))
		}
	}

	tsm.logger.Info("Read tasks from retry stream",
		"consumer_group", consumerGroup,
		"total_tasks", len(tasks),
		"ready_tasks", len(readyTasks))

	return readyTasks, nil
}

func (tsm *TaskStreamManager) readTasksFromStream(stream, consumerGroup, consumerName string, count int64) ([]TaskStreamData, error) {
	if err := tsm.RegisterConsumerGroup(stream, consumerGroup); err != nil {
		return nil, err
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
			metrics.TasksReadFromStreamTotal.WithLabelValues(stream, "empty").Inc()
			tsm.logger.Debug("No tasks available in stream",
				"stream", stream,
				"consumer_group", consumerGroup,
				"duration", duration)
			return []TaskStreamData{}, nil
		}
		tsm.logger.Error("Failed to read from stream",
			"stream", stream,
			"consumer_group", consumerGroup,
			"duration", duration,
			"error", err)
		return nil, fmt.Errorf("failed to read from stream: %w", err)
	}

	metrics.TasksReadFromStreamTotal.WithLabelValues(stream, "success").Inc()

	var tasks []TaskStreamData
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
			tsm.logger.Debug("Task read from stream",
				"task_id", task.TaskID,
				"stream", stream.Stream,
				"message_id", message.ID)
		}
	}

	tsm.logger.Info("Tasks read from stream successfully",
		"stream", stream,
		"task_count", len(tasks),
		"duration", duration)

	return tasks, nil
}

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
		"available":       config.IsRedisAvailable(),
		"max_length":      config.GetStreamMaxLen(),
		"ttl":             config.GetTaskStreamTTL().String(),
		"stream_lengths":  streamLengths,
		"max_retries":     MaxRetryAttempts,
		"retry_backoff":   RetryBackoffBase.String(),
		"consumer_groups": len(tsm.consumerGroups),
	}

	tsm.logger.Debug("Stream information retrieved", "info", info)
	return info
}

// notifySchedulerTaskComplete notifies schedulers about task completion status
// TODO: Implement actual scheduler notification mechanism
func (tsm *TaskStreamManager) notifySchedulerTaskComplete(task *TaskStreamData, success bool) {
	status := "failed"
	if success {
		status = "completed"
	}

	tsm.logger.Info("Task status notification",
		"task_id", task.TaskID,
		"job_id", task.JobID,
		"manager_id", task.ManagerID,
		"status", status,
		"retry_count", task.RetryCount)

	// TODO: Add actual notification mechanism to schedulers
	// This could be implemented as:
	// 1. HTTP callback to scheduler endpoints
	// 2. Message to scheduler-specific Redis stream
	// 3. Database update that schedulers poll
	// 4. Event publication to message broker
}

func (tsm *TaskStreamManager) Close() error {
	tsm.logger.Info("Closing TaskStreamManager")

	err := tsm.client.Close()
	if err != nil {
		tsm.logger.Error("Failed to close Redis client", "error", err)
		return err
	}

	tsm.logger.Info("TaskStreamManager closed successfully")
	return nil
}
