package tasks

import (
	"context"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

// ExpirationManager provides unified expiration management for both task timeouts and stream entries
type ExpirationManager struct {
	tsm *TaskStreamManager
}

// NewExpirationManager creates a new unified expiration manager
func NewExpirationManager(tsm *TaskStreamManager) *ExpirationManager {
	return &ExpirationManager{
		tsm: tsm,
	}
}

// AddTaskTimeout adds a task to the timeout tracking sorted set (for executed tasks)
func (em *ExpirationManager) AddTaskTimeout(ctx context.Context, taskID int64, timeoutDuration time.Duration) error {
	return em.AddExecutedTaskTimeout(ctx, taskID, timeoutDuration)
}

// AddExecutedTaskTimeout adds an executed task to the timeout tracking sorted set
func (em *ExpirationManager) AddExecutedTaskTimeout(ctx context.Context, taskID int64, timeoutDuration time.Duration) error {
	start := time.Now()

	err := em.tsm.redisClient.AddTaskToTimeoutTracking(ctx, taskID, timeoutDuration)
	duration := time.Since(start)

	if err != nil {
		if metrics.TasksAddedToStreamTotal != nil {
			metrics.TasksAddedToStreamTotal.WithLabelValues("timeout_add", "failure").Inc(ctx)
		}
		em.tsm.logger.Error(ctx, "Failed to add executed task timeout",
			observability.Int64("task_id", taskID),
			observability.Duration("duration", duration),
			observability.Error(err))
		return err
	}

	if metrics.TasksAddedToStreamTotal != nil {
		metrics.TasksAddedToStreamTotal.WithLabelValues("timeout_add", "success").Inc(ctx)
	}
	em.tsm.logger.Debug(ctx, "Executed task timeout added successfully",
		observability.Int64("task_id", taskID),
		observability.Duration("duration", duration))

	return nil
}

// GetExpiredTasks efficiently retrieves all executed tasks that have timed out (pending validation)
func (em *ExpirationManager) GetExpiredTasks(ctx context.Context) ([]int64, error) {
	return em.GetExpiredExecutedTasks(ctx)
}

// GetExpiredExecutedTasks efficiently retrieves all executed tasks that have timed out
func (em *ExpirationManager) GetExpiredExecutedTasks(ctx context.Context) ([]int64, error) {
	start := time.Now()

	taskIDs, err := em.tsm.redisClient.GetExpiredTasks(ctx)
	duration := time.Since(start)

	if err != nil {
		if metrics.TasksAddedToStreamTotal != nil {
			metrics.TasksAddedToStreamTotal.WithLabelValues("timeout_query", "failure").Inc(ctx)
		}
		em.tsm.logger.Error(ctx, "Failed to get expired executed tasks",
			observability.Duration("duration", duration),
			observability.Error(err))
		return nil, err
	}

	if metrics.TasksAddedToStreamTotal != nil {
		metrics.TasksAddedToStreamTotal.WithLabelValues("timeout_query", "success").Inc(ctx)
	}
	return taskIDs, nil
}

// RemoveTaskTimeout removes a task from the executed timeout tracking sorted set
func (em *ExpirationManager) RemoveTaskTimeout(ctx context.Context, taskID int64) error {
	return em.RemoveExecutedTaskTimeout(ctx, taskID)
}

// RemoveExecutedTaskTimeout removes an executed task from the timeout tracking sorted set
func (em *ExpirationManager) RemoveExecutedTaskTimeout(ctx context.Context, taskID int64) error {
	start := time.Now()

	err := em.tsm.redisClient.RemoveTaskFromTimeoutTracking(ctx, taskID)
	duration := time.Since(start)

	if err != nil {
		if metrics.TasksAddedToStreamTotal != nil {
			metrics.TasksAddedToStreamTotal.WithLabelValues("timeout_remove", "failure").Inc(ctx)
		}
		em.tsm.logger.Error(ctx, "Failed to remove executed task timeout",
			observability.Int64("task_id", taskID),
			observability.Duration("duration", duration),
			observability.Error(err))
		return err
	}

	if metrics.TasksAddedToStreamTotal != nil {
		metrics.TasksAddedToStreamTotal.WithLabelValues("timeout_remove", "success").Inc(ctx)
	}
	em.tsm.logger.Debug(ctx, "Executed task timeout removed successfully",
		observability.Int64("task_id", taskID),
		observability.Duration("duration", duration))

	return nil
}

// RemoveMultipleTaskTimeouts removes multiple tasks from the timeout tracking sorted set
func (em *ExpirationManager) RemoveMultipleTaskTimeouts(ctx context.Context, taskIDs []int64) error {
	if len(taskIDs) == 0 {
		return nil
	}

	start := time.Now()

	err := em.tsm.redisClient.RemoveMultipleTaskTimeouts(ctx, taskIDs)
	duration := time.Since(start)

	if err != nil {
		if metrics.TasksAddedToStreamTotal != nil {
			metrics.TasksAddedToStreamTotal.WithLabelValues("timeout_remove_multiple", "failure").Inc(ctx)
		}
		em.tsm.logger.Error(ctx, "Failed to remove multiple task timeouts",
			observability.Int("task_count", len(taskIDs)),
			observability.Duration("duration", duration),
			observability.Error(err))
		return err
	}

	if metrics.TasksAddedToStreamTotal != nil {
		metrics.TasksAddedToStreamTotal.WithLabelValues("timeout_remove_multiple", "success").Inc(ctx)
	}
	em.tsm.logger.Debug(ctx, "Multiple task timeouts removed successfully",
		observability.Int("task_count", len(taskIDs)),
		observability.Duration("duration", duration))

	return nil
}

// AddMessageExpiration adds a stream entry to the expiration tracking sorted set
func (em *ExpirationManager) AddMessageExpiration(ctx context.Context, stream string, messageID string, ttl time.Duration) error {
	start := time.Now()

	err := em.tsm.redisClient.AddStreamEntryExpiration(ctx, stream, messageID, ttl)
	duration := time.Since(start)

	if err != nil {
		if metrics.TasksAddedToStreamTotal != nil {
			metrics.TasksAddedToStreamTotal.WithLabelValues("expiration_add", "failure").Inc(ctx)
		}
		em.tsm.logger.Error(ctx, "Failed to add stream entry expiration",
			observability.String("stream", stream),
			observability.String("message_id", messageID),
			observability.Duration("duration", duration),
			observability.Error(err))
		return err
	}

	if metrics.TasksAddedToStreamTotal != nil {
		metrics.TasksAddedToStreamTotal.WithLabelValues("expiration_add", "success").Inc(ctx)
	}
	em.tsm.logger.Debug(ctx, "Stream entry expiration added successfully",
		observability.String("stream", stream),
		observability.String("message_id", messageID),
		observability.Duration("duration", duration),
		observability.Duration("ttl", ttl))

	return nil
}

// RemoveMessageExpiration removes a stream entry from the expiration tracking sorted set
func (em *ExpirationManager) RemoveMessageExpiration(ctx context.Context, stream string, messageID string) error {
	start := time.Now()

	err := em.tsm.redisClient.RemoveStreamEntryExpiration(ctx, stream, messageID)
	duration := time.Since(start)

	if err != nil {
		if metrics.TasksAddedToStreamTotal != nil {
			metrics.TasksAddedToStreamTotal.WithLabelValues("expiration_remove", "failure").Inc(ctx)
		}
		em.tsm.logger.Error(ctx, "Failed to remove stream entry expiration",
			observability.String("stream", stream),
			observability.String("message_id", messageID),
			observability.Duration("duration", duration),
			observability.Error(err))
		return err
	}

	if metrics.TasksAddedToStreamTotal != nil {
		metrics.TasksAddedToStreamTotal.WithLabelValues("expiration_remove", "success").Inc(ctx)
	}
	em.tsm.logger.Debug(ctx, "Stream entry expiration removed successfully",
		observability.String("stream", stream),
		observability.String("message_id", messageID),
		observability.Duration("duration", duration))

	return nil
}

// GetExpiredMessages efficiently retrieves all expired entries for a specific stream
func (em *ExpirationManager) GetExpiredMessages(ctx context.Context, stream string) ([]string, error) {
	start := time.Now()

	messageIDs, err := em.tsm.redisClient.GetExpiredStreamEntries(ctx, stream)
	duration := time.Since(start)

	if err != nil {
		if metrics.TasksAddedToStreamTotal != nil {
			metrics.TasksAddedToStreamTotal.WithLabelValues("expiration_query", "failure").Inc(ctx)
		}
		em.tsm.logger.Error(ctx, "Failed to get expired stream entries",
			observability.String("stream", stream),
			observability.Duration("duration", duration),
			observability.Error(err))
		return nil, err
	}

	if metrics.TasksAddedToStreamTotal != nil {
		metrics.TasksAddedToStreamTotal.WithLabelValues("expiration_query", "success").Inc(ctx)
	}
	return messageIDs, nil
}

// GetExpiredMessagesForAllStreams retrieves expired entries for all tracked streams
func (em *ExpirationManager) GetExpiredMessagesForAllStreams(ctx context.Context) (map[string][]string, error) {
	return em.tsm.redisClient.GetExpiredEntriesForAllStreams(ctx)
}

// RemoveMultipleMessageExpirations removes multiple entries from expiration tracking
func (em *ExpirationManager) RemoveMultipleMessageExpirations(ctx context.Context, stream string, messageIDs []string) error {
	if len(messageIDs) == 0 {
		return nil
	}

	start := time.Now()

	err := em.tsm.redisClient.RemoveMultipleStreamExpirations(ctx, stream, messageIDs)
	duration := time.Since(start)

	if err != nil {
		if metrics.TasksAddedToStreamTotal != nil {
			metrics.TasksAddedToStreamTotal.WithLabelValues("expiration_remove_multiple", "failure").Inc(ctx)
		}
		em.tsm.logger.Error(ctx, "Failed to remove multiple stream entry expirations",
			observability.String("stream", stream),
			observability.Int("count", len(messageIDs)),
			observability.Duration("duration", duration),
			observability.Error(err))
		return err
	}

	if metrics.TasksAddedToStreamTotal != nil {
		metrics.TasksAddedToStreamTotal.WithLabelValues("expiration_remove_multiple", "success").Inc(ctx)
	}
	em.tsm.logger.Debug(ctx, "Multiple stream entry expirations removed successfully",
		observability.String("stream", stream),
		observability.Int("count", len(messageIDs)),
		observability.Duration("duration", duration))

	return nil
}
