package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/trigg3rX/triggerx-backend/internal/redis/config"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// TaskExecutor interface for stream operations
type TaskStreamManager struct {
	client *Client
	logger logging.Logger
}

// NewTaskStreamManager creates a new task stream manager
func NewTaskStreamManager(logger logging.Logger) (*TaskStreamManager, error) {
	if !config.IsRedisAvailable() {
		return nil, fmt.Errorf("redis not available")
	}

	client, err := NewRedisClient(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create redis client: %w", err)
	}

	return &TaskStreamManager{
		client: client,
		logger: logger,
	}, nil
}

// AddTaskToReadyStream adds a task to the ready stream for keeper processing
func (tsm *TaskStreamManager) AddTaskToReadyStream(task *TaskStreamData) error {
	return tsm.addTaskToStream(TasksReadyStream, task)
}

// AddTaskToRetryStream adds a failed task to the retry stream
func (tsm *TaskStreamManager) AddTaskToRetryStream(task *TaskStreamData, retryReason string) error {
	task.RetryCount++
	now := time.Now()
	task.LastAttemptAt = &now
	
	// Calculate next retry time with exponential backoff
	backoffDuration := time.Duration(task.RetryCount) * RetryBackoffBase
	scheduledFor := now.Add(backoffDuration)
	task.ScheduledFor = &scheduledFor

	if task.RetryCount >= MaxRetryAttempts {
		tsm.logger.Warnf("Task %d exceeded max retry attempts, moving to failed stream", task.TaskID)
		return tsm.addTaskToStream(TasksFailedStream, task)
	}

	return tsm.addTaskToStream(TasksRetryStream, task)
}

// AddTaskToProcessingStream marks a task as being processed
func (tsm *TaskStreamManager) AddTaskToProcessingStream(task *TaskStreamData, performerID int64) error {
	return tsm.addTaskToStream(TasksProcessingStream, task)
}

// AddTaskToCompletedStream marks a task as completed
func (tsm *TaskStreamManager) AddTaskToCompletedStream(task *TaskStreamData, executionResult map[string]interface{}) error {
	return tsm.addTaskToStream(TasksCompletedStream, task)
}

// Private helper method to add task to any stream
func (tsm *TaskStreamManager) addTaskToStream(stream string, task *TaskStreamData) error {
	ctx := context.Background()

	taskJSON, err := json.Marshal(task)
	if err != nil {
		tsm.logger.Errorf("Failed to marshal task data: %v", err)
		return err
	}

	res, err := tsm.client.redisClient.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		MaxLen: int64(config.GetStreamMaxLen()),
		Approx: true,
		Values: map[string]interface{}{
			"task":       taskJSON,
			"created_at": time.Now().Unix(),
		},
	}).Result()

	if err != nil {
		tsm.logger.Errorf("Failed to add task to stream %s: %v", stream, err)
		return err
	}

	tsm.logger.Infof("Task %d added to stream %s with ID %s", task.TaskID, stream, res)

	// Set TTL on the stream
	if err := tsm.client.redisClient.Expire(ctx, stream, config.GetTaskStreamTTL()).Err(); err != nil {
		tsm.logger.Warnf("Failed to set TTL on stream %s: %v", stream, err)
	}

	return nil
}

// ReadTasksFromReadyStream reads tasks from ready stream for keeper processing
func (tsm *TaskStreamManager) ReadTasksFromReadyStream(consumerGroup, consumerName string, count int64) ([]TaskStreamData, error) {
	return tsm.readTasksFromStream(TasksReadyStream, consumerGroup, consumerName, count)
}

// ReadTasksFromRetryStream reads tasks from retry stream that are ready for retry
func (tsm *TaskStreamManager) ReadTasksFromRetryStream(consumerGroup, consumerName string, count int64) ([]TaskStreamData, error) {
	tasks, err := tsm.readTasksFromStream(TasksRetryStream, consumerGroup, consumerName, count)
	if err != nil {
		return nil, err
	}

	// Filter tasks that are ready for retry (past their scheduled time)
	now := time.Now()
	var readyTasks []TaskStreamData
	for _, task := range tasks {
		if task.ScheduledFor == nil || task.ScheduledFor.Before(now) {
			readyTasks = append(readyTasks, task)
		}
	}

	return readyTasks, nil
}

// Private helper method to read tasks from any stream
func (tsm *TaskStreamManager) readTasksFromStream(stream, consumerGroup, consumerName string, count int64) ([]TaskStreamData, error) {
	ctx := context.Background()

	// Create consumer group if it doesn't exist
	err := tsm.client.redisClient.XGroupCreate(ctx, stream, consumerGroup, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		tsm.logger.Warnf("Failed to create consumer group: %v", err)
	}

	// Read from stream
	streams, err := tsm.client.redisClient.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    consumerGroup,
		Consumer: consumerName,
		Streams:  []string{stream, ">"},
		Count:    count,
		Block:    time.Second,
	}).Result()

	if err != nil {
		if err == redis.Nil {
			return []TaskStreamData{}, nil // No new messages
		}
		return nil, fmt.Errorf("failed to read from stream: %w", err)
	}

	var tasks []TaskStreamData
	for _, stream := range streams {
		for _, message := range stream.Messages {
			taskJSON, exists := message.Values["task"].(string)
			if !exists {
				continue
			}

			var task TaskStreamData
			if err := json.Unmarshal([]byte(taskJSON), &task); err != nil {
				tsm.logger.Errorf("Failed to unmarshal task data: %v", err)
				continue
			}

			tasks = append(tasks, task)
		}
	}

	return tasks, nil
}

// AckTaskProcessed acknowledges that a task has been processed
func (tsm *TaskStreamManager) AckTaskProcessed(stream, consumerGroup, messageID string) error {
	ctx := context.Background()
	return tsm.client.redisClient.XAck(ctx, stream, consumerGroup, messageID).Err()
}

// GetStreamInfo returns information about task streams
func (tsm *TaskStreamManager) GetStreamInfo() map[string]interface{} {
	ctx := context.Background()
	
	streamLengths := make(map[string]int64)
	streams := []string{TasksReadyStream, TasksRetryStream, TasksProcessingStream, TasksCompletedStream, TasksFailedStream}
	
	for _, stream := range streams {
		length, err := tsm.client.redisClient.XLen(ctx, stream).Result()
		if err != nil {
			length = -1 // Indicate error
		}
		streamLengths[stream] = length
	}

	return map[string]interface{}{
		"available":       config.IsRedisAvailable(),
		"max_length":      config.GetStreamMaxLen(),
		"ttl":             config.GetTaskStreamTTL().String(),
		"stream_lengths":  streamLengths,
		"max_retries":     MaxRetryAttempts,
		"retry_backoff":   RetryBackoffBase.String(),
	}
}