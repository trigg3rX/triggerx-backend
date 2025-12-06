package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/trigg3rX/triggerx-backend/internal/taskdispatcher/config"
	"github.com/trigg3rX/triggerx-backend/internal/taskdispatcher/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (tsm *TaskStreamManager) AddTaskToDispatchedStream(ctx context.Context, task TaskStreamData) (bool, error) {
	// Prepare payload identical to previous implementation
	jsonData, err := json.Marshal(task.SendTaskDataToKeeper)
	if err != nil {
		tsm.logger.Error("Failed to marshal scheduler task data", "task_id", task.SendTaskDataToKeeper.TaskID[0], "error", err)
		return false, fmt.Errorf("failed to marshal task data: %w", err)
	}

	broadcast := types.BroadcastDataForPerformer{
		TaskID:           task.SendTaskDataToKeeper.TaskID[0],
		TaskDefinitionID: task.SendTaskDataToKeeper.TargetData[0].TaskDefinitionID,
		PerformerAddress: task.SendTaskDataToKeeper.PerformerData.KeeperAddress,
		Data:             []byte(jsonData),
	}

	var success bool
	if task.IsMainnet {
		success, err = tsm.aggregatorClient.SendTaskToPerformer(ctx, &broadcast)
	} else {
		success, err = tsm.testAggregatorClient.SendTaskToPerformer(ctx, &broadcast)
	}
	if err != nil {
		tsm.logger.Error("Failed to send task to aggregator", "task_id", task.SendTaskDataToKeeper.TaskID[0], "error", err)
		return false, err
	}
	if !success {
		tsm.logger.Warn("Aggregator send returned unsuccessful", "task_id", task.SendTaskDataToKeeper.TaskID[0])
		return false, fmt.Errorf("aggregator send unsuccessful")
	}

	success, err = tsm.addTaskToStream(ctx, StreamTaskDispatched, &task)
	if err != nil {
		return false, err
	}

	if !success {
		return false, fmt.Errorf("failed to add task to stream")
	}
	return true, nil
}

func (tsm *TaskStreamManager) addTaskToStream(ctx context.Context, stream string, task *TaskStreamData) (bool, error) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, config.GetRequestTimeout())
	defer cancel()

	taskJSON, err := json.Marshal(task)
	if err != nil {
		tsm.logger.Error("Failed to marshal task data",
			"task_id", task.SendTaskDataToKeeper.TaskID[0],
			"stream", stream,
			"error", err)
		return false, fmt.Errorf("failed to marshal task data: %w", err)
	}

	res, err := tsm.client.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		MaxLen: int64(10000),
		Approx: true,
		Values: map[string]interface{}{
			"task":       taskJSON,
			"created_at": time.Now().Unix(),
		},
	})
	duration := time.Since(start)

	if err != nil {
		metrics.TasksAddedToStreamTotal.WithLabelValues(stream, "failure").Inc()
		tsm.logger.Error("Failed to add task to stream",
			"task_id", task.SendTaskDataToKeeper.TaskID[0],
			"stream", stream,
			"duration", duration,
			"error", err)
		return false, fmt.Errorf("failed to add task to stream: %w", err)
	}

	// Store the task index mapping for efficient lookup
	if stream == StreamTaskDispatched {
		taskID := task.SendTaskDataToKeeper.TaskID[0]
		err = tsm.storeTaskIndex(ctx, taskID, res)
		if err != nil {
			tsm.logger.Warn("Failed to store task index, but task was added to stream",
				"task_id", taskID,
				"message_id", res,
				"error", err)
			// Don't fail the entire operation if index storage fails
		}

		// Add task to timeout tracking
		err = tsm.addTaskToTimeoutTracking(ctx, taskID)
		if err != nil {
			tsm.logger.Warn("Failed to add task to timeout tracking, but task was added to stream",
				"task_id", taskID,
				"error", err)
			// Don't fail the entire operation if timeout tracking fails
		}
	}

	// Track expiration for this stream entry with stream-specific TTL
	var entryTTL time.Duration
	switch stream {
	case StreamTaskDispatched:
		entryTTL = TasksProcessingTTL
	case StreamTaskCompleted:
		entryTTL = TasksCompletedTTL
	case StreamTaskFailed:
		entryTTL = TasksFailedTTL
	case StreamTaskRetry:
		entryTTL = TasksRetryTTL
	default:
		entryTTL = TasksProcessingTTL // Default fallback
	}

	// Track expiration using sorted set (similar to timeout tracking)
	expirationKey := fmt.Sprintf("stream:expiration:%s", stream)
	expirationTimestamp := float64(time.Now().Add(entryTTL).Unix())
	member := fmt.Sprintf("%s:%s", stream, res)
	_, err = tsm.client.ZAdd(ctx, expirationKey, redis.Z{
		Score:  expirationTimestamp,
		Member: member,
	})
	if err != nil {
		tsm.logger.Warn("Failed to add stream entry expiration, but task was added to stream",
			"task_id", task.SendTaskDataToKeeper.TaskID[0],
			"stream", stream,
			"message_id", res,
			"error", err)
		// Don't fail the entire operation if expiration tracking fails
	} else {
		// Set TTL on the sorted set (should be longer than max entry TTL)
		_ = tsm.client.SetTTL(ctx, expirationKey, 48*time.Hour)
	}

	metrics.TasksAddedToStreamTotal.WithLabelValues(stream, "success").Inc()
	tsm.logger.Debug("Task added to stream successfully",
		"task_id", task.SendTaskDataToKeeper.TaskID[0],
		"stream", stream,
		"stream_id", res,
		"duration", duration,
		"task_json_size", len(taskJSON),
		"entry_ttl", entryTTL)

	return true, nil
}
