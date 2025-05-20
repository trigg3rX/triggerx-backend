package execution

import (
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type TaskExecutor interface {
	Execute(job *types.HandleCreateJobData) (types.ActionData, error)
}

// TaskExecutor is the default implementation of TaskExecutor
type TaskExecutorImp struct {
	logger logging.Logger
}

// NewTaskExecutor creates a new instance of TaskExecutor
func NewTaskExecutor(logger logging.Logger) *TaskExecutorImp {
	return &TaskExecutorImp{
		logger: logger,
	}
}

// Execute implements the TaskExecutor interface
func (e *TaskExecutorImp) Execute(job *types.HandleCreateJobData) (types.ActionData, error) {
	e.logger.Info("Executing task", "jobID", job.JobID)

	// TODO: Implement actual task execution logic
	return types.ActionData{
		TaskID:        job.JobID,
		Timestamp:     time.Now(),
		Status:        true,
		ExecutionTime: 0,
	}, nil
}
