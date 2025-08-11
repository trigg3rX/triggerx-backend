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
	success, err := tsm.addTaskToStream(ctx, StreamTaskDispatched, &task)
	if err != nil {
		return false, err
	}

	if !success {
		return false, fmt.Errorf("failed to add task to stream")
	}

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
		Data:             jsonData,
	}

	success, err = tsm.aggregatorClient.SendTaskToPerformer(ctx, &broadcast)
	if err != nil {
		tsm.logger.Error("Failed to send task to aggregator", "task_id", task.SendTaskDataToKeeper.TaskID[0], "error", err)
		return false, err
	}
	if !success {
		tsm.logger.Warn("Aggregator send returned unsuccessful", "task_id", task.SendTaskDataToKeeper.TaskID[0])
		return false, fmt.Errorf("aggregator send unsuccessful")
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

	metrics.TasksAddedToStreamTotal.WithLabelValues(stream, "success").Inc()
	tsm.logger.Debug("Task added to stream successfully",
		"task_id", task.SendTaskDataToKeeper.TaskID[0],
		"stream", stream,
		"stream_id", res,
		"duration", duration,
		"task_json_size", len(taskJSON))

	return true, nil
}
