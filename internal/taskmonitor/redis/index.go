package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

// StoreTaskIndex stores the mapping from taskID to messageID in Redis hash
func (c *Client) StoreTaskIndex(ctx context.Context, taskID int64, messageID string) error {
	start := time.Now()
	taskIDStr := strconv.FormatInt(taskID, 10)

	err := c.client.HSet(ctx, TaskIndexKey, taskIDStr, messageID)
	duration := time.Since(start)

	if err != nil {
		c.logger.Error(ctx, "Failed to store task index",
			observability.Int64("task_id", taskID),
			observability.String("message_id", messageID),
			observability.Duration("duration", duration),
			observability.Error(err))
		return fmt.Errorf("failed to store task index: %w", err)
	}

	// Set TTL on the hash to ensure it expires
	if err := c.client.SetTTL(ctx, TaskIndexKey, TaskIndexTTL); err != nil {
		c.logger.Warn(ctx, "Failed to set TTL on task index",
			observability.Int64("task_id", taskID),
			observability.Error(err))
	}

	c.logger.Debug(ctx, "Task index stored successfully",
		observability.Int64("task_id", taskID),
		observability.String("message_id", messageID),
		observability.Duration("duration", duration))

	return nil
}

// GetTaskMessageID retrieves the messageID for a given taskID
func (c *Client) GetTaskMessageID(ctx context.Context, taskID int64) (string, bool, error) {
	start := time.Now()
	taskIDStr := strconv.FormatInt(taskID, 10)

	messageID, exists, err := c.client.HGetWithExists(ctx, TaskIndexKey, taskIDStr)
	duration := time.Since(start)

	if err != nil {
		c.logger.Error(ctx, "Failed to get task message ID",
			observability.Int64("task_id", taskID),
			observability.Duration("duration", duration),
			observability.Error(err))
		return "", false, fmt.Errorf("failed to get task message ID: %w", err)
	}

	if !exists {
		c.logger.Debug(ctx, "Task not found in index",
			observability.Int64("task_id", taskID),
			observability.Duration("duration", duration))
		return "", false, nil
	}

	c.logger.Debug(ctx, "Task message ID retrieved successfully",
		observability.Int64("task_id", taskID),
		observability.String("message_id", messageID),
		observability.Duration("duration", duration))

	return messageID, true, nil
}

// RemoveTaskIndex removes the taskID to messageID mapping from the index
func (c *Client) RemoveTaskIndex(ctx context.Context, taskID int64) error {
	start := time.Now()
	taskIDStr := strconv.FormatInt(taskID, 10)

	deletedCount, err := c.client.HDelWithCount(ctx, TaskIndexKey, taskIDStr)
	duration := time.Since(start)

	if err != nil {
		c.logger.Error(ctx, "Failed to remove task index",
			observability.Int64("task_id", taskID),
			observability.Duration("duration", duration),
			observability.Error(err))
		return fmt.Errorf("failed to remove task index: %w", err)
	}

	if deletedCount == 0 {
		c.logger.Debug(ctx, "Task index entry not found for removal",
			observability.Int64("task_id", taskID),
			observability.Duration("duration", duration))
	} else {
		c.logger.Debug(ctx, "Task index removed successfully",
			observability.Int64("task_id", taskID),
			observability.Duration("duration", duration))
	}

	return nil
}
