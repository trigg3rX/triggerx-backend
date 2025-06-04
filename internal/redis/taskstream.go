package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/trigg3rX/triggerx-backend/internal/redis/config"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type TaskStreamManager struct {
	client         *Client
	logger         logging.Logger
	consumerGroups map[string]bool
}

func NewTaskStreamManager(logger logging.Logger) (*TaskStreamManager, error) {
	if !config.IsRedisAvailable() {
		return nil, fmt.Errorf("redis not available")
	}

	client, err := NewRedisClient(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create redis client: %w", err)
	}

	return &TaskStreamManager{
		client:         client,
		logger:         logger,
		consumerGroups: make(map[string]bool),
	}, nil
}

func (tsm *TaskStreamManager) Initialize() error {
	streams := []string{
		TasksReadyStream, TasksRetryStream, TasksProcessingStream,
		TasksCompletedStream, TasksFailedStream,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, stream := range streams {
		if err := tsm.client.CreateStreamIfNotExists(ctx, stream, config.GetTaskStreamTTL()); err != nil {
			return fmt.Errorf("failed to initialize stream %s: %w", stream, err)
		}
	}
	return nil
}

func (tsm *TaskStreamManager) RegisterConsumerGroup(stream, group string) error {
	key := fmt.Sprintf("%s:%s", stream, group)
	if _, exists := tsm.consumerGroups[key]; exists {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := tsm.client.CreateConsumerGroup(ctx, stream, group); err != nil {
		return fmt.Errorf("failed to create consumer group for %s: %w", stream, err)
	}

	tsm.consumerGroups[key] = true
	tsm.logger.Infof("Created consumer group '%s' for stream '%s'", group, stream)
	return nil
}

func (tsm *TaskStreamManager) AddTaskToReadyStream(task *TaskStreamData) error {
	return tsm.addTaskToStream(TasksReadyStream, task)
}

func (tsm *TaskStreamManager) AddTaskToRetryStream(task *TaskStreamData, retryReason string) error {
	task.RetryCount++
	now := time.Now()
	task.LastAttemptAt = &now

	// Exponential backoff with jitter
	baseBackoff := time.Duration(task.RetryCount) * RetryBackoffBase
	jitter := time.Duration(rand.Int63n(int64(RetryBackoffBase))) // Up to 1 base backoff
	backoffDuration := baseBackoff + jitter

	scheduledFor := now.Add(backoffDuration)
	task.ScheduledFor = &scheduledFor

	if task.RetryCount >= MaxRetryAttempts {
		tsm.logger.Warnf("Task %d exceeded max retry attempts, moving to failed stream", task.TaskID)
		return tsm.addTaskToStream(TasksFailedStream, task)
	}

	return tsm.addTaskToStream(TasksRetryStream, task)
}

func (tsm *TaskStreamManager) AddTaskToProcessingStream(task *TaskStreamData, performerID int64) error {
	task.PerformerID = performerID
	return tsm.addTaskToStream(TasksProcessingStream, task)
}

func (tsm *TaskStreamManager) AddTaskToCompletedStream(task *TaskStreamData, executionResult map[string]interface{}) error {
	return tsm.addTaskToStream(TasksCompletedStream, task)
}

func (tsm *TaskStreamManager) addTaskToStream(stream string, task *TaskStreamData) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	taskJSON, err := json.Marshal(task)
	if err != nil {
		tsm.logger.Errorf("Failed to marshal task data: %v", err)
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

	if err != nil {
		tsm.logger.Errorf("Failed to add task to stream %s: %v", stream, err)
		return err
	}

	tsm.logger.Debugf("Task %d added to stream %s with ID %s", task.TaskID, stream, res)
	return nil
}

func (tsm *TaskStreamManager) ReadTasksFromReadyStream(consumerGroup, consumerName string, count int64) ([]TaskStreamData, error) {
	return tsm.readTasksFromStream(TasksReadyStream, consumerGroup, consumerName, count)
}

func (tsm *TaskStreamManager) ReadTasksFromRetryStream(consumerGroup, consumerName string, count int64) ([]TaskStreamData, error) {
	tasks, err := tsm.readTasksFromStream(TasksRetryStream, consumerGroup, consumerName, count)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	var readyTasks []TaskStreamData
	for _, task := range tasks {
		if task.ScheduledFor == nil || task.ScheduledFor.Before(now) {
			readyTasks = append(readyTasks, task)
		}
	}
	return readyTasks, nil
}

func (tsm *TaskStreamManager) readTasksFromStream(stream, consumerGroup, consumerName string, count int64) ([]TaskStreamData, error) {
	if err := tsm.RegisterConsumerGroup(stream, consumerGroup); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	streams, err := tsm.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    consumerGroup,
		Consumer: consumerName,
		Streams:  []string{stream, ">"},
		Count:    count,
		Block:    time.Second,
	})

	if err != nil {
		if err == redis.Nil {
			return []TaskStreamData{}, nil
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

func (tsm *TaskStreamManager) AckTaskProcessed(stream, consumerGroup, messageID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return tsm.client.XAck(ctx, stream, consumerGroup, messageID)
}

func (tsm *TaskStreamManager) GetStreamInfo() map[string]interface{} {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	streamLengths := make(map[string]int64)
	streams := []string{TasksReadyStream, TasksRetryStream, TasksProcessingStream, TasksCompletedStream, TasksFailedStream}

	for _, stream := range streams {
		length, err := tsm.client.XLen(ctx, stream)
		if err != nil {
			length = -1
		}
		streamLengths[stream] = length
	}

	return map[string]interface{}{
		"available":      config.IsRedisAvailable(),
		"max_length":     config.GetStreamMaxLen(),
		"ttl":            config.GetTaskStreamTTL().String(),
		"stream_lengths": streamLengths,
		"max_retries":    MaxRetryAttempts,
		"retry_backoff":  RetryBackoffBase.String(),
	}
}

func (tsm *TaskStreamManager) Close() error {
	return tsm.client.Close()
}
