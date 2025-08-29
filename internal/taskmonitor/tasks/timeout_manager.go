package tasks

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/config"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/metrics"
)

const (
	// DispatchedTimeoutsKey is the Redis sorted set key for tracking dispatched task timeouts
	DispatchedTimeoutsKey = "dispatched_timeouts"
	// DispatchedTimeoutsTTL is the TTL for the timeout tracking sorted set
	DispatchedTimeoutsTTL = 2 * time.Hour
)

// TimeoutManager provides efficient timeout management using Redis sorted sets
type TimeoutManager struct {
	tsm *TaskStreamManager
}

// NewTimeoutManager creates a new timeout manager
func NewTimeoutManager(tsm *TaskStreamManager) *TimeoutManager {
	return &TimeoutManager{
		tsm: tsm,
	}
}

// AddTaskTimeout adds a task to the timeout tracking sorted set
func (tm *TimeoutManager) AddTaskTimeout(ctx context.Context, taskID int64, timeoutDuration time.Duration) error {
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, config.GetReadTimeout())
	defer cancel()

	// Calculate timeout timestamp
	timeoutTimestamp := float64(time.Now().Add(timeoutDuration).Unix())
	taskIDStr := strconv.FormatInt(taskID, 10)

	// Add to sorted set with timeout timestamp as score
	_, err := tm.tsm.redisClient.ZAdd(ctx, DispatchedTimeoutsKey, redis.Z{
		Score:  timeoutTimestamp,
		Member: taskIDStr,
	})
	duration := time.Since(start)

	if err != nil {
		metrics.TasksAddedToStreamTotal.WithLabelValues("timeout_add", "failure").Inc()
		tm.tsm.logger.Error("Failed to add task timeout",
			"task_id", taskID,
			"timeout_timestamp", timeoutTimestamp,
			"duration", duration,
			"error", err)
		return fmt.Errorf("failed to add task timeout: %w", err)
	}

	// Set TTL on the sorted set to ensure it expires
	err = tm.tsm.redisClient.SetTTL(ctx, DispatchedTimeoutsKey, DispatchedTimeoutsTTL)
	if err != nil {
		tm.tsm.logger.Warn("Failed to set TTL on timeout tracking",
			"task_id", taskID,
			"error", err)
		// Don't return error as the main operation succeeded
	}

	metrics.TasksAddedToStreamTotal.WithLabelValues("timeout_add", "success").Inc()
	tm.tsm.logger.Debug("Task timeout added successfully",
		"task_id", taskID,
		"timeout_timestamp", timeoutTimestamp,
		"duration", duration)

	return nil
}

// GetExpiredTasks efficiently retrieves all tasks that have timed out
func (tm *TimeoutManager) GetExpiredTasks(ctx context.Context) ([]int64, error) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, config.GetReadTimeout())
	defer cancel()

	currentTimestamp := time.Now().Unix()

	// Use ZRANGEBYSCORE to get all tasks with timeout timestamp <= current timestamp
	// This is O(log(N) + M) where N is total tasks and M is expired tasks
	expiredTaskIDs, err := tm.tsm.redisClient.ZRangeByScore(ctx, DispatchedTimeoutsKey, "0", strconv.FormatInt(currentTimestamp, 10))
	duration := time.Since(start)

	if err != nil {
		metrics.TasksAddedToStreamTotal.WithLabelValues("timeout_query", "failure").Inc()
		tm.tsm.logger.Error("Failed to get expired tasks",
			"current_timestamp", currentTimestamp,
			"duration", duration,
			"error", err)
		return nil, fmt.Errorf("failed to get expired tasks: %w", err)
	}

	// Convert string task IDs to int64
	var taskIDs []int64
	for _, taskIDStr := range expiredTaskIDs {
		taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
		if err != nil {
			tm.tsm.logger.Error("Failed to parse task ID from timeout tracking",
				"task_id_str", taskIDStr,
				"error", err)
			continue
		}
		taskIDs = append(taskIDs, taskID)
	}

	metrics.TasksAddedToStreamTotal.WithLabelValues("timeout_query", "success").Inc()
	tm.tsm.logger.Debug("Expired tasks retrieved successfully",
		"expired_count", len(taskIDs),
		"current_timestamp", currentTimestamp,
		"duration", duration)

	return taskIDs, nil
}

// RemoveTaskTimeout removes a task from the timeout tracking sorted set
func (tm *TimeoutManager) RemoveTaskTimeout(ctx context.Context, taskID int64) error {
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, config.GetReadTimeout())
	defer cancel()

	taskIDStr := strconv.FormatInt(taskID, 10)

	// Remove from sorted set
	removed, err := tm.tsm.redisClient.ZRemRangeByScore(ctx, DispatchedTimeoutsKey, taskIDStr, taskIDStr)
	duration := time.Since(start)

	if err != nil {
		metrics.TasksAddedToStreamTotal.WithLabelValues("timeout_remove", "failure").Inc()
		tm.tsm.logger.Error("Failed to remove task timeout",
			"task_id", taskID,
			"duration", duration,
			"error", err)
		return fmt.Errorf("failed to remove task timeout: %w", err)
	}

	if removed == 0 {
		tm.tsm.logger.Debug("Task timeout entry not found for removal",
			"task_id", taskID,
			"duration", duration)
	} else {
		metrics.TasksAddedToStreamTotal.WithLabelValues("timeout_remove", "success").Inc()
		tm.tsm.logger.Debug("Task timeout removed successfully",
			"task_id", taskID,
			"duration", duration)
	}

	return nil
}

// RemoveMultipleTaskTimeouts removes multiple tasks from the timeout tracking sorted set
func (tm *TimeoutManager) RemoveMultipleTaskTimeouts(ctx context.Context, taskIDs []int64) error {
	if len(taskIDs) == 0 {
		return nil
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, config.GetReadTimeout())
	defer cancel()

	// Convert task IDs to strings for removal
	var taskIDStrs []string
	for _, taskID := range taskIDs {
		taskIDStrs = append(taskIDStrs, strconv.FormatInt(taskID, 10))
	}

	// Use pipeline to remove multiple tasks efficiently
	_, err := tm.tsm.redisClient.ExecutePipeline(ctx, func(pipe redis.Pipeliner) error {
		for _, taskIDStr := range taskIDStrs {
			pipe.ZRemRangeByScore(ctx, DispatchedTimeoutsKey, taskIDStr, taskIDStr)
		}
		return nil
	})
	duration := time.Since(start)

	if err != nil {
		metrics.TasksAddedToStreamTotal.WithLabelValues("timeout_remove_multiple", "failure").Inc()
		tm.tsm.logger.Error("Failed to remove multiple task timeouts",
			"task_count", len(taskIDs),
			"duration", duration,
			"error", err)
		return fmt.Errorf("failed to remove multiple task timeouts: %w", err)
	}

	metrics.TasksAddedToStreamTotal.WithLabelValues("timeout_remove_multiple", "success").Inc()
	tm.tsm.logger.Debug("Multiple task timeouts removed successfully",
		"task_count", len(taskIDs),
		"duration", duration)

	return nil
}

// GetTimeoutStats returns statistics about the timeout tracking
func (tm *TimeoutManager) GetTimeoutStats(ctx context.Context) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, config.GetReadTimeout())
	defer cancel()

	// Get total count of tracked timeouts
	totalCount, err := tm.tsm.redisClient.ZCard(ctx, DispatchedTimeoutsKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get timeout count: %w", err)
	}

	// Get count of expired tasks
	currentTimestamp := time.Now().Unix()
	expiredCount, err := tm.tsm.redisClient.ZCard(ctx, DispatchedTimeoutsKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get expired count: %w", err)
	}

	stats := map[string]interface{}{
		"total_tracked": totalCount,
		"expired_count": expiredCount,
		"current_time":  currentTimestamp,
		"timeout_key":   DispatchedTimeoutsKey,
		"timeout_ttl":   DispatchedTimeoutsTTL.String(),
	}

	return stats, nil
}
