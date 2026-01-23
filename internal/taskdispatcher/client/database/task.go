package database

import (
	"context"

	"github.com/trigg3rX/triggerx-backend/internal/taskdispatcher/client/database/queries"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// UpdateTaskError updates a task with error information
func (dm *DatabaseClient) UpdateTaskStatusToDispatched(ctx context.Context, taskID int64) error {
	if err := dm.db.NewQuery(queries.UpdateTaskStatusQuery, types.TaskStatusDispatched, "", taskID).Exec(); err != nil {
		dm.logger.Error(ctx, "Error updating task status to dispatched for task ID", observability.Int64("task_id", taskID), observability.Error(err))
		return err
	}
	dm.logger.Debug(ctx, "Successfully updated task status to dispatched", observability.Int64("task_id", taskID))
	return nil
}

func (dm *DatabaseClient) UpdateTaskStatusToFailed(ctx context.Context, taskID int64, errorMsg string) error {
	if err := dm.db.NewQuery(queries.UpdateTaskStatusQuery, types.TaskStatusFailed, errorMsg, taskID).Exec(); err != nil {
		dm.logger.Error(ctx, "Error updating task status to failed for task ID", observability.Int64("task_id", taskID), observability.Error(err))
		return err
	}
	dm.logger.Debug(ctx, "Successfully updated task status to failed", observability.Int64("task_id", taskID))
	return nil
}
