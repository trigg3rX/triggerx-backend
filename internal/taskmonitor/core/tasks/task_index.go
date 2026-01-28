package tasks

import (
	"context"
	"fmt"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
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

	err := tim.tsm.redisClient.StoreTaskIndex(ctx, taskID, messageID)
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
		return err
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

	messageID, exists, err := tim.tsm.redisClient.GetTaskMessageID(ctx, taskID)
	duration := time.Since(start)

	if err != nil {
		if metrics.TasksAddedToStreamTotal != nil {
			metrics.TasksAddedToStreamTotal.WithLabelValues("index_lookup", "failure").Inc(ctx)
		}
		tim.tsm.logger.Error(ctx, "Failed to get task message ID",
			observability.Int64("task_id", taskID),
			observability.Duration("duration", duration),
			observability.Error(err))
		return "", false, err
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

	err := tim.tsm.redisClient.RemoveTaskIndex(ctx, taskID)
	duration := time.Since(start)

	if err != nil {
		if metrics.TasksAddedToStreamTotal != nil {
			metrics.TasksAddedToStreamTotal.WithLabelValues("index_remove", "failure").Inc(ctx)
		}
		tim.tsm.logger.Error(ctx, "Failed to remove task index",
			observability.Int64("task_id", taskID),
			observability.Duration("duration", duration),
			observability.Error(err))
		return err
	}

	if metrics.TasksAddedToStreamTotal != nil {
		metrics.TasksAddedToStreamTotal.WithLabelValues("index_remove", "success").Inc(ctx)
	}
	tim.tsm.logger.Debug(ctx, "Task index removed successfully",
		observability.Int64("task_id", taskID),
		observability.Duration("duration", duration))

	return nil
}

// FindTaskByID efficiently finds a task by its ID using the index (searches dispatched stream by default)
func (tim *TaskIndexManager) FindTaskByID(ctx context.Context, taskID int64) (*types.TaskStreamData, string, error) {
	return tim.FindTaskByIDInStream(ctx, taskID, types.StreamTaskDispatched)
}

// FindTaskByIDInStream efficiently finds a task by its ID in a specific stream
func (tim *TaskIndexManager) FindTaskByIDInStream(ctx context.Context, taskID int64, stream string) (*types.TaskStreamData, string, error) {
	start := time.Now()

	// First, get the messageID from the index
	messageID, exists, err := tim.GetTaskMessageID(ctx, taskID)
	if err != nil {
		return nil, "", err
	}

	if !exists {
		tim.tsm.logger.Debug(ctx, "Task not found in index, falling back to stream scan",
			observability.Int64("task_id", taskID),
			observability.String("stream", stream))
		// Fall back to the old method for backward compatibility
		task, err := tim.tsm.findTaskInStream(taskID, stream)
		if err != nil {
			return nil, "", err
		}
		return task, "", nil
	}

	// Use XRANGE to get the specific message efficiently without adding to PEL
	task, err := tim.getTaskByMessageID(ctx, messageID, stream)
	if err != nil {
		duration := time.Since(start)
		tim.tsm.logger.Warn(ctx, "Task index found messageID but task not found in stream",
			observability.Int64("task_id", taskID),
			observability.String("message_id", messageID),
			observability.String("stream", stream),
			observability.Duration("duration", duration))
		return nil, messageID, fmt.Errorf("task %d not found in stream %s despite having messageID %s: %w", taskID, stream, messageID, err)
	}

	duration := time.Since(start)
	tim.tsm.logger.Debug(ctx, "Task found efficiently using index",
		observability.Int64("task_id", taskID),
		observability.String("message_id", messageID),
		observability.String("stream", stream),
		observability.Duration("duration", duration))
	return task, messageID, nil
}

// getTaskByMessageID retrieves a specific task by its messageID using XRANGE from a specific stream
func (tim *TaskIndexManager) getTaskByMessageID(ctx context.Context, messageID string, stream string) (*types.TaskStreamData, error) {
	start := time.Now()

	task, err := tim.tsm.redisClient.GetTaskByMessageID(ctx, stream, messageID)
	duration := time.Since(start)

	if err != nil {
		tim.tsm.logger.Error(ctx, "Failed to get task by messageID",
			observability.String("message_id", messageID),
			observability.String("stream", stream),
			observability.Duration("duration", duration),
			observability.Error(err))
		return nil, err
	}

	tim.tsm.logger.Debug(ctx, "Task retrieved by messageID successfully",
		observability.Int64("task_id", task.SendTaskDataToKeeper.TaskID),
		observability.String("message_id", messageID),
		observability.String("stream", stream),
		observability.Duration("duration", duration))

	return task, nil
}
