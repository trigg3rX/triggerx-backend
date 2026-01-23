package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/trigg3rX/triggerx-backend/internal/taskdispatcher/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (tsm *TaskStreamManager) GetTaskDataFromStream(ctx context.Context, stream string, taskID int64) (*types.TaskStreamData, error) {
	// Use XRANGE to read recent messages without adding to PEL
	// This is more efficient for lookup operations
	streams, err := tsm.client.Client().XRange(ctx, stream, "-", "+").Result()
	if err != nil {
		tsm.logger.Error(ctx, "Failed to read task stream data",
			observability.Int64("task_id", taskID),
			observability.Error(err))
		return nil, fmt.Errorf("failed to read task stream data: %w", err)
	}

	// Limit the search to the most recent 500 messages to avoid performance issues
	startIndex := 0
	if len(streams) > 500 {
		startIndex = len(streams) - 500
	}

	for i := startIndex; i < len(streams); i++ {
		message := streams[i]
		taskJSON, exists := message.Values["task"].(string)
		if !exists {
			continue
		}

		var task types.TaskStreamData
		if err := json.Unmarshal([]byte(taskJSON), &task); err != nil {
			tsm.logger.Error(ctx, "Failed to unmarshal task data",
				observability.String("message_id", message.ID),
				observability.Error(err))
			metrics.TasksReadFromStreamTotal.WithLabelValues(stream, "error").Inc(ctx)
			continue
		}

		if task.SendTaskDataToKeeper.TaskID[0] == taskID {
			metrics.TasksReadFromStreamTotal.WithLabelValues(stream, "success").Inc(ctx)
			return &task, nil
		}
	}

	return nil, fmt.Errorf("task not found: %d", taskID)
}
