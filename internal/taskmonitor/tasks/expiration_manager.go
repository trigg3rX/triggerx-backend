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
	// StreamExpirationKeyPrefix is the prefix for sorted sets tracking stream entry expiration
	StreamExpirationKeyPrefix = "stream:expiration:"
	// ExpirationTrackingTTL is the TTL for expiration tracking sorted sets
	ExpirationTrackingTTL = 1 * time.Hour
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

// getStreamExpirationKey returns the Redis key for tracking expiration of a specific stream
func (em *ExpirationManager) getStreamExpirationKey(stream string) string {
	return StreamExpirationKeyPrefix + stream
}

// AddTaskTimeout adds a task to the timeout tracking sorted set
func (em *ExpirationManager) AddTaskTimeout(ctx context.Context, taskID int64, timeoutDuration time.Duration) error {
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, config.GetReadTimeout())
	defer cancel()

	timeoutTimestamp := float64(time.Now().Add(timeoutDuration).Unix())
	taskIDStr := strconv.FormatInt(taskID, 10)

	_, err := em.tsm.redisClient.ZAdd(ctx, DispatchedTimeoutsKey, redis.Z{
		Score:  timeoutTimestamp,
		Member: taskIDStr,
	})
	duration := time.Since(start)

	if err != nil {
		metrics.TasksAddedToStreamTotal.WithLabelValues("timeout_add", "failure").Inc()
		em.tsm.logger.Error("Failed to add task timeout",
			"task_id", taskID,
			"timeout_timestamp", timeoutTimestamp,
			"duration", duration,
			"error", err)
		return fmt.Errorf("failed to add task timeout: %w", err)
	}

	// Set TTL on the sorted set to ensure it expires
	err = em.tsm.redisClient.SetTTL(ctx, DispatchedTimeoutsKey, ExpirationTrackingTTL)
	if err != nil {
		em.tsm.logger.Warn("Failed to set TTL on timeout tracking",
			"task_id", taskID,
			"error", err)
	}

	metrics.TasksAddedToStreamTotal.WithLabelValues("timeout_add", "success").Inc()
	em.tsm.logger.Debug("Task timeout added successfully",
		"task_id", taskID,
		"timeout_timestamp", timeoutTimestamp,
		"duration", duration)

	return nil
}

// GetExpiredTasks efficiently retrieves all tasks that have timed out
func (em *ExpirationManager) GetExpiredTasks(ctx context.Context) ([]int64, error) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, config.GetReadTimeout())
	defer cancel()

	currentTimestamp := time.Now().Unix()
	expiredTaskIDs, err := em.tsm.redisClient.ZRangeByScore(ctx, DispatchedTimeoutsKey, "0", strconv.FormatInt(currentTimestamp, 10))
	duration := time.Since(start)

	if err != nil {
		metrics.TasksAddedToStreamTotal.WithLabelValues("timeout_query", "failure").Inc()
		em.tsm.logger.Error("Failed to get expired tasks",
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
			em.tsm.logger.Error("Failed to parse task ID from timeout tracking",
				"task_id_str", taskIDStr,
				"error", err)
			continue
		}
		taskIDs = append(taskIDs, taskID)
	}

	metrics.TasksAddedToStreamTotal.WithLabelValues("timeout_query", "success").Inc()
	return taskIDs, nil
}

// RemoveTaskTimeout removes a task from the timeout tracking sorted set
func (em *ExpirationManager) RemoveTaskTimeout(ctx context.Context, taskID int64) error {
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, config.GetReadTimeout())
	defer cancel()

	taskIDStr := strconv.FormatInt(taskID, 10)
	removed, err := em.tsm.redisClient.ZRem(ctx, DispatchedTimeoutsKey, taskIDStr)
	duration := time.Since(start)

	if err != nil {
		metrics.TasksAddedToStreamTotal.WithLabelValues("timeout_remove", "failure").Inc()
		em.tsm.logger.Error("Failed to remove task timeout",
			"task_id", taskID,
			"duration", duration,
			"error", err)
		return fmt.Errorf("failed to remove task timeout: %w", err)
	}

	if removed == 0 {
		em.tsm.logger.Debug("Task timeout entry not found for removal",
			"task_id", taskID,
			"duration", duration)
	} else {
		metrics.TasksAddedToStreamTotal.WithLabelValues("timeout_remove", "success").Inc()
		em.tsm.logger.Debug("Task timeout removed successfully",
			"task_id", taskID,
			"duration", duration)
	}

	return nil
}

// RemoveMultipleTaskTimeouts removes multiple tasks from the timeout tracking sorted set
func (em *ExpirationManager) RemoveMultipleTaskTimeouts(ctx context.Context, taskIDs []int64) error {
	if len(taskIDs) == 0 {
		return nil
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, config.GetReadTimeout())
	defer cancel()

	var taskIDStrs []string
	for _, taskID := range taskIDs {
		taskIDStrs = append(taskIDStrs, strconv.FormatInt(taskID, 10))
	}

	_, err := em.tsm.redisClient.ExecutePipeline(ctx, func(pipe redis.Pipeliner) error {
		for _, taskIDStr := range taskIDStrs {
			pipe.ZRem(ctx, DispatchedTimeoutsKey, taskIDStr)
		}
		return nil
	})
	duration := time.Since(start)

	if err != nil {
		metrics.TasksAddedToStreamTotal.WithLabelValues("timeout_remove_multiple", "failure").Inc()
		em.tsm.logger.Error("Failed to remove multiple task timeouts",
			"task_count", len(taskIDs),
			"duration", duration,
			"error", err)
		return fmt.Errorf("failed to remove multiple task timeouts: %w", err)
	}

	metrics.TasksAddedToStreamTotal.WithLabelValues("timeout_remove_multiple", "success").Inc()
	em.tsm.logger.Debug("Multiple task timeouts removed successfully",
		"task_count", len(taskIDs),
		"duration", duration)

	return nil
}

// AddMessageExpiration adds a stream entry to the expiration tracking sorted set
func (em *ExpirationManager) AddMessageExpiration(ctx context.Context, stream string, messageID string, ttl time.Duration) error {
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, config.GetReadTimeout())
	defer cancel()

	expirationTimestamp := float64(time.Now().Add(ttl).Unix())
	expirationKey := em.getStreamExpirationKey(stream)
	member := fmt.Sprintf("%s:%s", stream, messageID)

	_, err := em.tsm.redisClient.ZAdd(ctx, expirationKey, redis.Z{
		Score:  expirationTimestamp,
		Member: member,
	})
	duration := time.Since(start)

	if err != nil {
		metrics.TasksAddedToStreamTotal.WithLabelValues("expiration_add", "failure").Inc()
		em.tsm.logger.Error("Failed to add stream entry expiration",
			"stream", stream,
			"message_id", messageID,
			"expiration_timestamp", expirationTimestamp,
			"duration", duration,
			"error", err)
		return fmt.Errorf("failed to add stream entry expiration: %w", err)
	}

	// Set TTL on the sorted set to ensure it expires
	err = em.tsm.redisClient.SetTTL(ctx, expirationKey, ExpirationTrackingTTL)
	if err != nil {
		em.tsm.logger.Warn("Failed to set TTL on expiration tracking",
			"stream", stream,
			"message_id", messageID,
			"error", err)
	}

	metrics.TasksAddedToStreamTotal.WithLabelValues("expiration_add", "success").Inc()
	em.tsm.logger.Debug("Stream entry expiration added successfully",
		"stream", stream,
		"message_id", messageID,
		"expiration_timestamp", expirationTimestamp,
		"ttl", ttl,
		"duration", duration)

	return nil
}

// RemoveMessageExpiration removes a stream entry from the expiration tracking sorted set
func (em *ExpirationManager) RemoveMessageExpiration(ctx context.Context, stream string, messageID string) error {
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, config.GetReadTimeout())
	defer cancel()

	expirationKey := em.getStreamExpirationKey(stream)
	member := fmt.Sprintf("%s:%s", stream, messageID)

	removed, err := em.tsm.redisClient.ZRem(ctx, expirationKey, member)
	duration := time.Since(start)

	if err != nil {
		metrics.TasksAddedToStreamTotal.WithLabelValues("expiration_remove", "failure").Inc()
		em.tsm.logger.Error("Failed to remove stream entry expiration",
			"stream", stream,
			"message_id", messageID,
			"duration", duration,
			"error", err)
		return fmt.Errorf("failed to remove stream entry expiration: %w", err)
	}

	if removed == 0 {
		em.tsm.logger.Debug("Stream entry expiration not found for removal",
			"stream", stream,
			"message_id", messageID,
			"duration", duration)
	} else {
		metrics.TasksAddedToStreamTotal.WithLabelValues("expiration_remove", "success").Inc()
		em.tsm.logger.Debug("Stream entry expiration removed successfully",
			"stream", stream,
			"message_id", messageID,
			"duration", duration)
	}

	return nil
}

// GetExpiredMessages efficiently retrieves all expired entries for a specific stream
func (em *ExpirationManager) GetExpiredMessages(ctx context.Context, stream string) ([]string, error) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, config.GetReadTimeout())
	defer cancel()

	currentTimestamp := time.Now().Unix()
	expirationKey := em.getStreamExpirationKey(stream)

	expiredMembers, err := em.tsm.redisClient.ZRangeByScore(ctx, expirationKey, "0", strconv.FormatInt(currentTimestamp, 10))
	duration := time.Since(start)

	if err != nil {
		metrics.TasksAddedToStreamTotal.WithLabelValues("expiration_query", "failure").Inc()
		em.tsm.logger.Error("Failed to get expired stream entries",
			"stream", stream,
			"current_timestamp", currentTimestamp,
			"duration", duration,
			"error", err)
		return nil, fmt.Errorf("failed to get expired stream entries: %w", err)
	}

	// Extract message IDs from members (format: "stream:messageID")
	var messageIDs []string
	streamPrefix := stream + ":"
	for _, member := range expiredMembers {
		if len(member) > len(streamPrefix) && member[:len(streamPrefix)] == streamPrefix {
			messageID := member[len(streamPrefix):]
			messageIDs = append(messageIDs, messageID)
		}
	}

	metrics.TasksAddedToStreamTotal.WithLabelValues("expiration_query", "success").Inc()
	return messageIDs, nil
}

// GetExpiredMessagesForAllStreams retrieves expired entries for all tracked streams
func (em *ExpirationManager) GetExpiredMessagesForAllStreams(ctx context.Context) (map[string][]string, error) {
	streams := []string{
		StreamTaskDispatched,
		StreamTaskCompleted,
		StreamTaskFailed,
		StreamTaskRetry,
	}

	expiredEntries := make(map[string][]string)
	for _, stream := range streams {
		messageIDs, err := em.GetExpiredMessages(ctx, stream)
		if err != nil {
			em.tsm.logger.Warn("Failed to get expired entries for stream",
				"stream", stream,
				"error", err)
			continue
		}
		if len(messageIDs) > 0 {
			expiredEntries[stream] = messageIDs
		}
	}

	return expiredEntries, nil
}

// RemoveMultipleMessageExpirations removes multiple entries from expiration tracking
func (em *ExpirationManager) RemoveMultipleMessageExpirations(ctx context.Context, stream string, messageIDs []string) error {
	if len(messageIDs) == 0 {
		return nil
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, config.GetReadTimeout())
	defer cancel()

	expirationKey := em.getStreamExpirationKey(stream)

	var members []string
	for _, messageID := range messageIDs {
		members = append(members, fmt.Sprintf("%s:%s", stream, messageID))
	}

	_, err := em.tsm.redisClient.ExecutePipeline(ctx, func(pipe redis.Pipeliner) error {
		for _, member := range members {
			pipe.ZRem(ctx, expirationKey, member)
		}
		return nil
	})
	duration := time.Since(start)

	if err != nil {
		metrics.TasksAddedToStreamTotal.WithLabelValues("expiration_remove_multiple", "failure").Inc()
		em.tsm.logger.Error("Failed to remove multiple stream entry expirations",
			"stream", stream,
			"count", len(messageIDs),
			"duration", duration,
			"error", err)
		return fmt.Errorf("failed to remove multiple stream entry expirations: %w", err)
	}

	metrics.TasksAddedToStreamTotal.WithLabelValues("expiration_remove_multiple", "success").Inc()
	em.tsm.logger.Debug("Multiple stream entry expirations removed successfully",
		"stream", stream,
		"count", len(messageIDs),
		"duration", duration)

	return nil
}
