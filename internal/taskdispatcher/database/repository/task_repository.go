package repository

import (
	"context"

	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// TaskRepository defines the interface for task data operations
type TaskRepository interface {
	UpdateTaskStatusToDispatched(ctx context.Context, taskID int64) error
	UpdateTaskStatusToFailed(ctx context.Context, taskID int64, errorMsg string) error
}

type taskRepository struct {
	db     *database.Connection
	logger observability.Logger
}

// NewTaskRepository creates a new task repository instance
func NewTaskRepository(db *database.Connection, logger observability.Logger) TaskRepository {
	return &taskRepository{
		db:     db,
		logger: logger,
	}
}

// UpdateTaskStatusToDispatched updates a task status to dispatched
func (r *taskRepository) UpdateTaskStatusToDispatched(ctx context.Context, taskID int64) error {
	if err := r.db.NewQuery(UpdateTaskStatusQuery, types.TaskStatusDispatched, "", taskID).Exec(); err != nil {
		r.logger.Error(ctx, "Error updating task status to dispatched for task ID", observability.Int64("task_id", taskID), observability.Error(err))
		return err
	}
	r.logger.Debug(ctx, "Successfully updated task status to dispatched", observability.Int64("task_id", taskID))
	return nil
}

// UpdateTaskStatusToFailed updates a task status to failed with error message
func (r *taskRepository) UpdateTaskStatusToFailed(ctx context.Context, taskID int64, errorMsg string) error {
	if err := r.db.NewQuery(UpdateTaskStatusQuery, types.TaskStatusFailed, errorMsg, taskID).Exec(); err != nil {
		r.logger.Error(ctx, "Error updating task status to failed for task ID", observability.Int64("task_id", taskID), observability.Error(err))
		return err
	}
	r.logger.Debug(ctx, "Successfully updated task status to failed", observability.Int64("task_id", taskID))
	return nil
}
