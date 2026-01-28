package dispatcher

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/trigg3rX/triggerx-backend/internal/taskdispatcher/config"
	"github.com/trigg3rX/triggerx-backend/internal/taskdispatcher/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// GetStreamInfo returns information about task streams
func (tsm *TaskStreamManager) GetStreamInfo(ctx context.Context) map[string]interface{} {
	tsm.logger.Debug(ctx, "Getting stream information")

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	streamLengths := make(map[string]int64)
	streams := []string{types.StreamTaskDispatched, types.StreamTaskRetry, types.StreamTaskCompleted, types.StreamTaskFailed}

	for _, stream := range streams {
		length, err := tsm.redisClient.GetStreamLength(ctx, stream)
		if err != nil {
			tsm.logger.Warn(ctx, "Failed to get stream length",
				observability.String("stream", stream),
				observability.Error(err))
			length = -1
		}
		streamLengths[stream] = length

		// Update stream length metrics
		switch stream {
		case types.StreamTaskDispatched:
			if metrics.TaskStreamLengths != nil {
				metrics.TaskStreamLengths.WithLabelValues("dispatched").Set(ctx, float64(length))
			}
		case types.StreamTaskRetry:
			if metrics.TaskStreamLengths != nil {
				metrics.TaskStreamLengths.WithLabelValues("retry").Set(ctx, float64(length))
			}
		case types.StreamTaskCompleted:
			if metrics.TaskStreamLengths != nil {
				metrics.TaskStreamLengths.WithLabelValues("completed").Set(ctx, float64(length))
			}
		case types.StreamTaskFailed:
			if metrics.TaskStreamLengths != nil {
				metrics.TaskStreamLengths.WithLabelValues("failed").Set(ctx, float64(length))
			}
		}
	}

	info := map[string]interface{}{
		"available":            tsm.redisClient != nil,
		"max_length":           10000, // Default value, can be made configurable
		"tasks_processing_ttl": types.TasksProcessingTTL.String(),
		"tasks_completed_ttl":  types.TasksCompletedTTL.String(),
		"tasks_failed_ttl":     types.TasksFailedTTL.String(),
		"tasks_retry_ttl":      types.TasksRetryTTL.String(),
		"stream_lengths":       streamLengths,
		"max_retries":          types.MaxRetryAttempts,
		"consumer_groups":      len(tsm.consumerGroups),
	}

	tsm.logger.Debug(ctx, "Stream information retrieved", observability.Any("info", info))
	return info
}

// StartStreamHealthMonitor monitors the health of Redis streams
func (tsm *TaskStreamManager) StartStreamHealthMonitor(ctx context.Context) {
	tsm.logger.Info(ctx, "Starting stream health monitor")

	ticker := time.NewTicker(30 * time.Second) // Check health every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			tsm.logger.Info(ctx, "Stream health monitor shutting down")
			return
		case <-ticker.C:
			// Get stream information
			taskInfo := tsm.GetStreamInfo(ctx)

			// Log warnings for high stream length
			if taskLengths, ok := taskInfo["stream_lengths"].(map[string]int64); ok {
				for stream, length := range taskLengths {
					if length > 50 && stream != types.StreamTaskFailed { // Warn if more than 50 tasks in any stream, ignore the StreamTaskFailed stream
						tsm.logger.Warn(ctx, "High task stream length detected",
							observability.String("stream", stream),
							observability.Int64("length", length))
					}
				}
			}
		}
	}
}

// addTaskToStream adds a task to a specified stream
func (tsm *TaskStreamManager) addTaskToStream(ctx context.Context, stream string, task *types.TaskStreamData) (bool, error) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, config.GetRequestTimeout())
	defer cancel()

	taskJSON, err := json.Marshal(task)
	if err != nil {
		tsm.logger.Error(ctx, "Failed to marshal task data",
			observability.Int64("task_id", task.SendTaskDataToKeeper.TaskID),
			observability.String("stream", stream),
			observability.Error(err))
		return false, fmt.Errorf("failed to marshal task data: %w", err)
	}

	res, err := tsm.redisClient.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		MaxLen: int64(10000),
		Approx: true,
		Values: map[string]interface{}{
			"task":       string(taskJSON),
			"created_at": time.Now().Unix(),
		},
	})
	duration := time.Since(start)

	if err != nil {
		if metrics.TasksAddedToStreamTotal != nil {
			metrics.TasksAddedToStreamTotal.WithLabelValues(stream, "failure").Inc(ctx)
		}
		tsm.logger.Error(ctx, "Failed to add task to stream",
			observability.Int64("task_id", task.SendTaskDataToKeeper.TaskID),
			observability.String("stream", stream),
			observability.Duration("duration", duration),
			observability.Error(err))
		return false, fmt.Errorf("failed to add task to stream: %w", err)
	}

	// Store the task index mapping for efficient lookup
	if stream == types.StreamTaskDispatched {
		taskID := task.SendTaskDataToKeeper.TaskID
		err = tsm.redisClient.StoreTaskIndex(ctx, taskID, res)
		if err != nil {
			tsm.logger.Warn(ctx, "Failed to store task index, but task was added to stream",
				observability.Int64("task_id", taskID),
				observability.String("message_id", res),
				observability.Error(err))
			// Don't fail the entire operation if index storage fails
		}

		// Add task to timeout tracking
		err = tsm.redisClient.AddTaskToTimeoutTracking(ctx, taskID, types.TasksProcessingTTL)
		if err != nil {
			tsm.logger.Warn(ctx, "Failed to add task to timeout tracking, but task was added to stream",
				observability.Int64("task_id", taskID),
				observability.Error(err))
			// Don't fail the entire operation if timeout tracking fails
		}
	}

	// Track expiration for this stream entry with stream-specific TTL
	var entryTTL time.Duration
	switch stream {
	case types.StreamTaskDispatched:
		entryTTL = types.TasksProcessingTTL
	case types.StreamTaskCompleted:
		entryTTL = types.TasksCompletedTTL
	case types.StreamTaskFailed:
		entryTTL = types.TasksFailedTTL
	case types.StreamTaskRetry:
		entryTTL = types.TasksRetryTTL
	default:
		entryTTL = types.TasksProcessingTTL // Default fallback
	}

	// Track expiration using redis client method
	err = tsm.redisClient.AddStreamEntryExpiration(ctx, stream, res, entryTTL)
	if err != nil {
		tsm.logger.Warn(ctx, "Failed to add stream entry expiration, but task was added to stream",
			observability.Int64("task_id", task.SendTaskDataToKeeper.TaskID),
			observability.String("stream", stream),
			observability.String("message_id", res),
			observability.Error(err))
		// Don't fail the entire operation if expiration tracking fails
	}

	if metrics.TasksAddedToStreamTotal != nil {
		metrics.TasksAddedToStreamTotal.WithLabelValues(stream, "success").Inc(ctx)
	}
	tsm.logger.Debug(ctx, "Task added to stream successfully",
		observability.Int64("task_id", task.SendTaskDataToKeeper.TaskID),
		observability.String("stream", stream),
		observability.String("stream_id", res),
		observability.Duration("duration", duration),
		observability.Int("task_json_size", len(taskJSON)),
		observability.Duration("entry_ttl", entryTTL))

	return true, nil
}
