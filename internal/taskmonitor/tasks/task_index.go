package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/config"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

const (
	// TaskIndexKey is the Redis hash key that maps taskID to messageID
	TaskIndexKey = "task_id_to_message_id"
	// TaskIndexTTL is the TTL for the task index hash (should be longer than stream TTL)
	TaskIndexTTL = 2 * time.Hour
)

// TaskIndexManager provides efficient O(1) lookup of tasks by their ID
type TaskIndexManager struct {
	tsm *TaskStreamManager
}

// NewTaskIndexManager creates a new task index manager
func NewTaskIndexManager(tsm *TaskStreamManager) *TaskIndexManager {
	return &TaskIndexManager{
		tsm: tsm,
	}
}

// StoreTaskIndex stores the mapping from taskID to messageID in Redis hash
func (tim *TaskIndexManager) StoreTaskIndex(ctx context.Context, taskID int64, messageID string) error {
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, config.GetReadTimeout())
	defer cancel()

	taskIDStr := strconv.FormatInt(taskID, 10)

	err := tim.tsm.redisClient.HSet(ctx, TaskIndexKey, taskIDStr, messageID)
	duration := time.Since(start)

	if err != nil {
		if metrics.TasksAddedToStreamTotal != nil {
			metrics.TasksAddedToStreamTotal.WithLabelValues("index_store", "failure").Inc(ctx)
		}
		tim.tsm.logger.Error(ctx, "Failed to store task index",
			observability.Int64("task_id", taskID),
			observability.String("message_id", messageID),
			observability.Duration("duration", duration),
			observability.Error(err))
		return fmt.Errorf("failed to store task index: %w", err)
	}

	// Set TTL on the hash to ensure it expires
	err = tim.tsm.redisClient.SetTTL(ctx, TaskIndexKey, TaskIndexTTL)
	if err != nil {
		tim.tsm.logger.Warn(ctx, "Failed to set TTL on task index",
			observability.Int64("task_id", taskID),
			observability.Error(err))
		// Don't return error as the main operation succeeded
	}

	if metrics.TasksAddedToStreamTotal != nil {
		metrics.TasksAddedToStreamTotal.WithLabelValues("index_store", "success").Inc(ctx)
	}
	tim.tsm.logger.Debug(ctx, "Task index stored successfully",
		observability.Int64("task_id", taskID),
		observability.String("message_id", messageID),
		observability.Duration("duration", duration))

	return nil
}

// GetTaskMessageID retrieves the messageID for a given taskID
func (tim *TaskIndexManager) GetTaskMessageID(ctx context.Context, taskID int64) (string, bool, error) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, config.GetReadTimeout())
	defer cancel()

	taskIDStr := strconv.FormatInt(taskID, 10)

	messageID, exists, err := tim.tsm.redisClient.HGetWithExists(ctx, TaskIndexKey, taskIDStr)
	duration := time.Since(start)

	if err != nil {
		if metrics.TasksAddedToStreamTotal != nil {
			metrics.TasksAddedToStreamTotal.WithLabelValues("index_lookup", "failure").Inc(ctx)
		}
		tim.tsm.logger.Error(ctx, "Failed to get task message ID",
			observability.Int64("task_id", taskID),
			observability.Duration("duration", duration),
			observability.Error(err))
		return "", false, fmt.Errorf("failed to get task message ID: %w", err)
	}

	if !exists {
		tim.tsm.logger.Debug(ctx, "Task not found in index",
			observability.Int64("task_id", taskID),
			observability.Duration("duration", duration))
		return "", false, nil
	}

	if metrics.TasksAddedToStreamTotal != nil {
		metrics.TasksAddedToStreamTotal.WithLabelValues("index_lookup", "success").Inc(ctx)
	}
	tim.tsm.logger.Debug(ctx, "Task message ID retrieved successfully",
		observability.Int64("task_id", taskID),
		observability.String("message_id", messageID),
		observability.Duration("duration", duration))

	return messageID, true, nil
}

// RemoveTaskIndex removes the taskID to messageID mapping from the index
func (tim *TaskIndexManager) RemoveTaskIndex(ctx context.Context, taskID int64) error {
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, config.GetReadTimeout())
	defer cancel()

	taskIDStr := strconv.FormatInt(taskID, 10)

	deletedCount, err := tim.tsm.redisClient.HDelWithCount(ctx, TaskIndexKey, taskIDStr)
	duration := time.Since(start)

	if err != nil {
		if metrics.TasksAddedToStreamTotal != nil {
			metrics.TasksAddedToStreamTotal.WithLabelValues("index_remove", "failure").Inc(ctx)
		}
		tim.tsm.logger.Error(ctx, "Failed to remove task index",
			observability.Int64("task_id", taskID),
			observability.Duration("duration", duration),
			observability.Error(err))
		return fmt.Errorf("failed to remove task index: %w", err)
	}

	if deletedCount == 0 {
		tim.tsm.logger.Debug(ctx, "Task index entry not found for removal",
			observability.Int64("task_id", taskID),
			observability.Duration("duration", duration))
	} else {
		if metrics.TasksAddedToStreamTotal != nil {
			metrics.TasksAddedToStreamTotal.WithLabelValues("index_remove", "success").Inc(ctx)
		}
		tim.tsm.logger.Debug(ctx, "Task index removed successfully",
			observability.Int64("task_id", taskID),
			observability.Duration("duration", duration))
	}

	return nil
}

// FindTaskByID efficiently finds a task by its ID using the index
func (tim *TaskIndexManager) FindTaskByID(ctx context.Context, taskID int64) (*TaskStreamData, string, error) {
	start := time.Now()

	// First, get the messageID from the index
	messageID, exists, err := tim.GetTaskMessageID(ctx, taskID)
	if err != nil {
		return nil, "", err
	}

	if !exists {
		tim.tsm.logger.Debug(ctx, "Task not found in index, falling back to stream scan",
			observability.Int64("task_id", taskID))
		// Fall back to the old method for backward compatibility
		task, err := tim.tsm.findTaskInDispatched(taskID)
		if err != nil {
			return nil, "", err
		}
		return task, "", nil
	}

	// Use XRANGE to get the specific message efficiently without adding to PEL
	task, err := tim.getTaskByMessageID(ctx, messageID)
	if err != nil {
		duration := time.Since(start)
		tim.tsm.logger.Warn(ctx, "Task index found messageID but task not found in stream",
			observability.Int64("task_id", taskID),
			observability.String("message_id", messageID),
			observability.Duration("duration", duration))
		return nil, messageID, fmt.Errorf("task %d not found in stream despite having messageID %s: %w", taskID, messageID, err)
	}

	duration := time.Since(start)
	tim.tsm.logger.Debug(ctx, "Task found efficiently using index",
		observability.Int64("task_id", taskID),
		observability.String("message_id", messageID),
		observability.Duration("duration", duration))
	return task, messageID, nil
}

// getTaskByMessageID retrieves a specific task by its messageID using XRANGE
func (tim *TaskIndexManager) getTaskByMessageID(ctx context.Context, messageID string) (*TaskStreamData, error) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, config.GetReadTimeout())
	defer cancel()

	// Use XRANGE to get the specific message without adding to PEL
	streams, err := tim.tsm.redisClient.Client().XRange(ctx, StreamTaskDispatched, messageID, messageID).Result()
	if err != nil {
		tim.tsm.logger.Error(ctx, "Failed to get task by messageID using XRANGE",
			observability.String("message_id", messageID),
			observability.Error(err))
		return nil, fmt.Errorf("failed to get task by messageID: %w", err)
	}

	if len(streams) == 0 {
		return nil, fmt.Errorf("no message found with ID %s", messageID)
	}

	message := streams[0]
	taskJSON, exists := message.Values["task"].(string)
	if !exists {
		tim.tsm.logger.Error(ctx, "Message missing task data",
			observability.String("message_id", messageID))
		return nil, fmt.Errorf("message %s missing task data", messageID)
	}

	var task TaskStreamData
	if err := json.Unmarshal([]byte(taskJSON), &task); err != nil {
		tim.tsm.logger.Error(ctx, "Failed to unmarshal task data",
			observability.String("message_id", messageID),
			observability.Error(err))
		return nil, fmt.Errorf("failed to unmarshal task data: %w", err)
	}

	duration := time.Since(start)
	tim.tsm.logger.Debug(ctx, "Task retrieved by messageID successfully",
		observability.Int64("task_id", task.SendTaskDataToKeeper.TaskID[0]),
		observability.String("message_id", messageID),
		observability.Duration("duration", duration))

	return &task, nil
}
