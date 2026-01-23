package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/config"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (tsm *TaskStreamManager) GetTaskDataFromStream(ctx context.Context, stream string, taskID int64) (*types.TaskStreamData, error) {
	taskStreamData, _, err := tsm.ReadTasksFromStream(ctx, stream, "task_stream_manager", "task_stream_manager", 1000)
	if err != nil {
		tsm.logger.Error(ctx, "Failed to read task stream data",
			observability.Int64("task_id", taskID),
			observability.Error(err))
		return nil, fmt.Errorf("failed to read task stream data: %w", err)
	}

	for _, task := range taskStreamData {
		if task.SendTaskDataToKeeper.TaskID[0] == taskID {
			return &task, nil
		}
	}

	return nil, fmt.Errorf("task not found: %d", taskID)
}

func (tsm *TaskStreamManager) ReadTasksFromStream(ctx context.Context, stream, consumerGroup, consumerName string, count int64) ([]types.TaskStreamData, []string, error) {
	if err := tsm.RegisterConsumerGroup(ctx, stream, consumerGroup); err != nil {
		return nil, nil, fmt.Errorf("failed to register consumer group: %w", err)
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, config.GetRequestTimeout())
	defer cancel()

	streams, err := tsm.redisClient.XReadGroup(ctx, &redis.XReadGroupArgs{
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
			// tsm.logger.Debug("No tasks available in stream",
			// 	"stream", stream,
			// 	"consumer_group", consumerGroup,
			// 	"duration", duration)
			return []types.TaskStreamData{}, []string{}, nil
		}
		tsm.logger.Error(ctx, "Failed to read from stream",
			observability.String("stream", stream),
			observability.String("consumer_group", consumerGroup),
			observability.Duration("duration", duration),
			observability.Error(err))
		return nil, nil, fmt.Errorf("failed to read from stream: %w", err)
	}

	if metrics.TasksReadFromStreamTotal != nil {
		metrics.TasksReadFromStreamTotal.WithLabelValues(stream, "success").Inc(ctx)
	}

	// Pre-allocate slice for better performance
	var tasks []types.TaskStreamData
	var messageIDs []string
	totalMessages := 0
	for _, stream := range streams {
		totalMessages += len(stream.Messages)
	}
	tasks = make([]types.TaskStreamData, 0, totalMessages)

	for _, stream := range streams {
		for _, message := range stream.Messages {
			taskJSON, exists := message.Values["task"].(string)
			if !exists {
				tsm.logger.Warn(ctx, "Message missing task data",
					observability.String("stream", stream.Stream),
					observability.String("message_id", message.ID))
				continue
			}

			var task types.TaskStreamData
			if err := json.Unmarshal([]byte(taskJSON), &task); err != nil {
				tsm.logger.Error(ctx, "Failed to unmarshal task data",
					observability.String("stream", stream.Stream),
					observability.String("message_id", message.ID),
					observability.Error(err))
				continue
			}

			tasks = append(tasks, task)
			messageIDs = append(messageIDs, message.ID)

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

	return tasks, messageIDs, nil
}
