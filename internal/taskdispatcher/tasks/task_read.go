package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/trigg3rX/triggerx-backend/internal/taskdispatcher/config"
	"github.com/trigg3rX/triggerx-backend/internal/taskdispatcher/metrics"
)

func (tsm *TaskStreamManager) GetTaskDataFromStream(stream string, taskID int64) (*TaskStreamData, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, config.GetRequestTimeout())
	defer cancel()

	// Use XRANGE to read recent messages without adding to PEL
	// This is more efficient for lookup operations
	streams, err := tsm.client.Client().XRange(ctx, stream, "-", "+").Result()
	if err != nil {
		tsm.logger.Error("Failed to read task stream data",
			"task_id", taskID,
			"error", err)
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
			tsm.logger.Error("Failed to unmarshal task data",
				"message_id", message.ID,
				"error", err)
			continue
		}

		if task.SendTaskDataToKeeper.TaskID[0] == taskID {
			return &task, nil
		}
	}

	return nil, fmt.Errorf("task not found: %d", taskID)
}

func (tsm *TaskStreamManager) ReadTasksFromStream(stream, consumerGroup, consumerName string, count int64) ([]TaskStreamData, error) {
	if err := tsm.RegisterConsumerGroup(stream, consumerGroup); err != nil {
		return nil, fmt.Errorf("failed to register consumer group: %w", err)
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), config.GetRequestTimeout())
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
				"task_id", task.SendTaskDataToKeeper.TaskID[0],
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
