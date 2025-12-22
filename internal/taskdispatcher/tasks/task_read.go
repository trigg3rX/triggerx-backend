package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/trigg3rX/triggerx-backend/internal/taskdispatcher/config"
	"github.com/trigg3rX/triggerx-backend/internal/taskdispatcher/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

func (tsm *TaskStreamManager) GetTaskDataFromStream(stream string, taskID int64) (*TaskStreamData, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, config.GetRequestTimeout())
	defer cancel()

	// Use XRANGE to read recent messages without adding to PEL
	// This is more efficient for lookup operations
	streams, err := tsm.client.Client().XRange(ctx, stream, "-", "+").Result()
	if err != nil {
		tsm.logger.Error(ctx, "Failed to read task stream data",
			observability.Int64("task_id", taskID),
			observability.Error(err))
		return nil, fmt.Errorf("failed to read task stream data: %w", err)
	}

	// Limit the search to the most recent 100 messages to avoid performance issues
	startIndex := 0
	if len(streams) > 100 {
		startIndex = len(streams) - 100
	}

	for i := startIndex; i < len(streams); i++ {
		message := streams[i]
		taskJSON, exists := message.Values["task"].(string)
		if !exists {
			continue
		}

		var task TaskStreamData
		if err := json.Unmarshal([]byte(taskJSON), &task); err != nil {
			tsm.logger.Error(ctx, "Failed to unmarshal task data",
				observability.String("message_id", message.ID),
				observability.Error(err))
			continue
		}

		if task.SendTaskDataToKeeper.TaskID[0] == taskID {
			return &task, nil
		}
	}

	return nil, fmt.Errorf("task not found: %d", taskID)
}

func (tsm *TaskStreamManager) ReadTasksFromStream(ctx context.Context, stream, consumerGroup, consumerName string, count int64) ([]TaskStreamData, error) {
	if err := tsm.RegisterConsumerGroup(ctx, stream, consumerGroup); err != nil {
		return nil, fmt.Errorf("failed to register consumer group: %w", err)
	}

	start := time.Now()

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
			if metrics.TasksReadFromStreamTotal != nil {
				metrics.TasksReadFromStreamTotal.WithLabelValues(stream, "empty").Inc(ctx)
			}
			tsm.logger.Debug(ctx, "No tasks available in stream",
				observability.String("stream", stream),
				observability.String("consumer_group", consumerGroup),
				observability.Duration("duration", duration))
			return []TaskStreamData{}, nil
		}
		tsm.logger.Error(ctx, "Failed to read from stream",
			observability.String("stream", stream),
			observability.String("consumer_group", consumerGroup),
			observability.Duration("duration", duration),
			observability.Error(err))
		return nil, fmt.Errorf("failed to read from stream: %w", err)
	}

	if metrics.TasksReadFromStreamTotal != nil {
		metrics.TasksReadFromStreamTotal.WithLabelValues(stream, "success").Inc(ctx)
	}

	// Pre-allocate slice for better performance
	var tasks []TaskStreamData
	totalMessages := 0
	for _, stream := range streams {
		totalMessages += len(stream.Messages)
	}
	tasks = make([]TaskStreamData, 0, totalMessages)

	for _, stream := range streams {
		for _, message := range stream.Messages {
			taskJSON, exists := message.Values["task"].(string)
			if !exists {
				tsm.logger.Warn(ctx, "Message missing task data",
					observability.String("stream", stream.Stream),
					observability.String("message_id", message.ID))
				continue
			}

			var task TaskStreamData
			if err := json.Unmarshal([]byte(taskJSON), &task); err != nil {
				tsm.logger.Error(ctx, "Failed to unmarshal task data",
					observability.String("stream", stream.Stream),
					observability.String("message_id", message.ID),
					observability.Error(err))
				continue
			}

			tasks = append(tasks, task)
			tsm.logger.Debug(ctx, "Task read from stream",
				observability.Int64("task_id", task.SendTaskDataToKeeper.TaskID[0]),
				observability.String("stream", stream.Stream),
				observability.String("message_id", message.ID))
		}
	}

	tsm.logger.Info(ctx, "Tasks read from stream successfully",
		observability.String("stream", stream),
		observability.Int("task_count", len(tasks)),
		observability.Duration("duration", duration))

	return tasks, nil
}
