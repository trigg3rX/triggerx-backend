package stream

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/trigg3rX/triggerx-backend/internal/redis/metrics"
)

func (tsm *TaskStreamManager) getTaskStreamData(taskID int64) (*TaskStreamData, error) {
	taskStreamData, err := tsm.readTasksFromStream(TasksReadyStream, "task_stream_manager", "task_stream_manager", 1)
	if err != nil {
		tsm.logger.Error("Failed to read task stream data", "error", err)
		return nil, err
	}

	for _, task := range taskStreamData {
		if task.SendTaskDataToKeeper.TaskID == taskID {
			return &task, nil
		}
	}

	return nil, fmt.Errorf("task not found")
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
				"task_id", task.SendTaskDataToKeeper.TaskID,
				"retry_count", task.RetryCount,
				"scheduled_for", task.ScheduledFor)
		} else {
			tsm.logger.Debug("Task not yet ready for retry",
				"task_id", task.SendTaskDataToKeeper.TaskID,
				"scheduled_for", task.ScheduledFor,
				"time_remaining", task.ScheduledFor.Sub(now))
		}
	}

	tsm.logger.Debug("Read tasks from retry stream",
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
				"task_id", task.SendTaskDataToKeeper.TaskID,
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
