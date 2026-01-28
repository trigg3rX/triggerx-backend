package tasks

import (
	"context"
	"fmt"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (tsm *TaskStreamManager) ReadTasksFromStream(ctx context.Context, stream, consumerGroup, consumerName string, count int64) ([]types.TaskStreamData, []string, error) {
	if err := tsm.RegisterConsumerGroup(ctx, stream, consumerGroup); err != nil {
		return nil, nil, fmt.Errorf("failed to register consumer group: %w", err)
	}

	start := time.Now()

	// Use redis client to read tasks from stream
	tasks, messageIDs, err := tsm.redisClient.ReadTasksFromStream(ctx, stream, consumerGroup, consumerName, count, time.Second)
	duration := time.Since(start)

	if err != nil {
		tsm.logger.Error(ctx, "Failed to read from stream",
			observability.String("stream", stream),
			observability.String("consumer_group", consumerGroup),
			observability.Duration("duration", duration),
			observability.Error(err))
		return nil, nil, fmt.Errorf("failed to read from stream: %w", err)
	}

	if len(tasks) == 0 {
		if metrics.TasksReadFromStreamTotal != nil {
			metrics.TasksReadFromStreamTotal.WithLabelValues(stream, "empty").Inc(ctx)
		}
		return []types.TaskStreamData{}, []string{}, nil
	}

	if metrics.TasksReadFromStreamTotal != nil {
		metrics.TasksReadFromStreamTotal.WithLabelValues(stream, "success").Inc(ctx)
	}

	tsm.logger.Info(ctx, "Tasks read from stream successfully",
		observability.String("stream", stream),
		observability.Int("task_count", len(tasks)),
		observability.Duration("duration", duration))

	return tasks, messageIDs, nil
}
