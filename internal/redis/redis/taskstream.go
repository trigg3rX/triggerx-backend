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
	if !config.IsRedisAvailable() {
		metrics.ServiceStatus.WithLabelValues("task_stream_manager").Set(0)
		return nil, fmt.Errorf("redis not available")
	}

	client, err := NewRedisClient(logger)
	if err != nil {
		metrics.ServiceStatus.WithLabelValues("task_stream_manager").Set(0)
		return nil, fmt.Errorf("failed to create redis client: %w", err)
	}

	metrics.ServiceStatus.WithLabelValues("task_stream_manager").Set(1)
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

func (tsm *TaskStreamManager) RegisterConsumerGroup(stream string, group string) error {
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

// Ready ot be sent to the performer
func (tsm *TaskStreamManager) AddTaskToReadyStream(task *TaskStreamData) (types.PerformerData, error) {
	performerData := GetPerformerData()
	if performerData.KeeperID == 0 {
		return types.PerformerData{}, fmt.Errorf("no performers available")
	}

	err := tsm.addTaskToStream(TasksReadyStream, task)
	if err != nil {
		return types.PerformerData{}, err
	}
	return performerData, nil
}

// Processing by the performer
func (tsm *TaskStreamManager) AddTaskToProcessingStream(task *TaskStreamData) error {
	err := tsm.addTaskToStream(TasksProcessingStream, task)
	if err != nil {
		return err
	}
	return nil
}

// Completed by the performer, sent to the validators
func (tsm *TaskStreamManager) AddTaskToCompletedStream(task *TaskStreamData) error {
	err := tsm.addTaskToStream(TasksCompletedStream, task)
	if err != nil {
		return err
	}
	return nil
}

// Failed to send to the performer, sent to the retry stream
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
		metrics.TaskMaxRetriesExceededTotal.Inc()
		metrics.TasksMovedToFailedStreamTotal.Inc()

		err := tsm.addTaskToStream(TasksFailedStream, task)
		if err != nil {
			return err
		}
	}

	err := tsm.addTaskToStream(TasksRetryStream, task)
	if err != nil {
		metrics.TasksAddedToStreamTotal.WithLabelValues("retry", "failure").Inc()
		return err
	}
	metrics.TasksAddedToStreamTotal.WithLabelValues("retry", "success").Inc()
	return nil
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
		metrics.TasksAddedToStreamTotal.WithLabelValues(stream, "failure").Inc()
		tsm.logger.Errorf("Failed to add task to stream %s: %v", stream, err)
		return err
	}

	metrics.TasksAddedToStreamTotal.WithLabelValues(stream, "success").Inc()
	tsm.logger.Debugf("Task %d added to stream %s with ID %s", task.TaskID, stream, res)
	return nil
}

func (tsm *TaskStreamManager) ReadTasksFromReadyStream(consumerGroup, consumerName string, count int64) ([]TaskStreamData, error) {
	tasks, err := tsm.readTasksFromStream(TasksReadyStream, consumerGroup, consumerName, count)
	if err != nil {
		return nil, err
	}
	return tasks, nil
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
			metrics.TasksReadFromStreamTotal.WithLabelValues(stream, "empty").Inc()
			return []TaskStreamData{}, nil
		}
		return nil, fmt.Errorf("failed to read from stream: %w", err)
	}

	metrics.TasksReadFromStreamTotal.WithLabelValues(stream, "success").Inc()

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
