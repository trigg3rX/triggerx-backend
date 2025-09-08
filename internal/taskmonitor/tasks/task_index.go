package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/config"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/metrics"
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
		metrics.TasksAddedToStreamTotal.WithLabelValues("index_store", "failure").Inc()
		tim.tsm.logger.Error("Failed to store task index",
			"task_id", taskID,
			"message_id", messageID,
			"duration", duration,
			"error", err)
		return fmt.Errorf("failed to store task index: %w", err)
	}

	// Set TTL on the hash to ensure it expires
	err = tim.tsm.redisClient.SetTTL(ctx, TaskIndexKey, TaskIndexTTL)
	if err != nil {
		tim.tsm.logger.Warn("Failed to set TTL on task index",
			"task_id", taskID,
			"error", err)
		// Don't return error as the main operation succeeded
	}

	metrics.TasksAddedToStreamTotal.WithLabelValues("index_store", "success").Inc()
	tim.tsm.logger.Debug("Task index stored successfully",
		"task_id", taskID,
		"message_id", messageID,
		"duration", duration)

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
		metrics.TasksAddedToStreamTotal.WithLabelValues("index_lookup", "failure").Inc()
		tim.tsm.logger.Error("Failed to get task message ID",
			"task_id", taskID,
			"duration", duration,
			"error", err)
		return "", false, fmt.Errorf("failed to get task message ID: %w", err)
	}

	if !exists {
		tim.tsm.logger.Debug("Task not found in index",
			"task_id", taskID,
			"duration", duration)
		return "", false, nil
	}

	metrics.TasksAddedToStreamTotal.WithLabelValues("index_lookup", "success").Inc()
	tim.tsm.logger.Debug("Task message ID retrieved successfully",
		"task_id", taskID,
		"message_id", messageID,
		"duration", duration)

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
		metrics.TasksAddedToStreamTotal.WithLabelValues("index_remove", "failure").Inc()
		tim.tsm.logger.Error("Failed to remove task index",
			"task_id", taskID,
			"duration", duration,
			"error", err)
		return fmt.Errorf("failed to remove task index: %w", err)
	}

	if deletedCount == 0 {
		tim.tsm.logger.Debug("Task index entry not found for removal",
			"task_id", taskID,
			"duration", duration)
	} else {
		metrics.TasksAddedToStreamTotal.WithLabelValues("index_remove", "success").Inc()
		tim.tsm.logger.Debug("Task index removed successfully",
			"task_id", taskID,
			"duration", duration)
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
		tim.tsm.logger.Debug("Task not found in index, falling back to stream scan",
			"task_id", taskID)
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
		tim.tsm.logger.Warn("Task index found messageID but task not found in stream",
			"task_id", taskID,
			"message_id", messageID,
			"duration", duration)
		return nil, messageID, fmt.Errorf("task %d not found in stream despite having messageID %s: %w", taskID, messageID, err)
	}

	duration := time.Since(start)
	tim.tsm.logger.Debug("Task found efficiently using index",
		"task_id", taskID,
		"message_id", messageID,
		"duration", duration)
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
		tim.tsm.logger.Error("Failed to get task by messageID using XRANGE",
			"message_id", messageID,
			"error", err)
		return nil, fmt.Errorf("failed to get task by messageID: %w", err)
	}

	if len(streams) == 0 {
		return nil, fmt.Errorf("no message found with ID %s", messageID)
	}

	message := streams[0]
	taskJSON, exists := message.Values["task"].(string)
	if !exists {
		tim.tsm.logger.Error("Message missing task data",
			"message_id", messageID)
		return nil, fmt.Errorf("message %s missing task data", messageID)
	}

	var task TaskStreamData
	if err := json.Unmarshal([]byte(taskJSON), &task); err != nil {
		tim.tsm.logger.Error("Failed to unmarshal task data",
			"message_id", messageID,
			"error", err)
		return nil, fmt.Errorf("failed to unmarshal task data: %w", err)
	}

	duration := time.Since(start)
	tim.tsm.logger.Debug("Task retrieved by messageID successfully",
		"task_id", task.SendTaskDataToKeeper.TaskID[0],
		"message_id", messageID,
		"duration", duration)

	return &task, nil
}
