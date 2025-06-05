package execution

import (
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// TaskExecutor is the default implementation of TaskExecutor
type TaskExecutor struct {
	ethClient       *ethclient.Client
	etherscanAPIKey string
	argConverter    *ArgumentConverter
	logger          logging.Logger
}

// NewTaskExecutor creates a new instance of TaskExecutor
func NewTaskExecutor(ethClient *ethclient.Client, etherscanAPIKey string, logger logging.Logger) *TaskExecutor {
	return &TaskExecutor{
		ethClient:       ethClient,
		etherscanAPIKey: etherscanAPIKey,
		argConverter:    &ArgumentConverter{},
		// validator:       validation.NewJobValidator(logger, ethClient),
		logger: logger,
	}
}

// Execute implements the TaskExecutor interface
func (e *TaskExecutor) ExecuteTimeBasedTask(timeJobData *types.ScheduleTimeJobData) (types.ActionData, error) {
	e.logger.Info("Executing time based task", "jobID", timeJobData.JobID)

	// TODO: Execute the task
	return types.ActionData{}, nil
}


func (e *TaskExecutor) ExecuteTask(taskTargetData *types.TaskTargetData, triggerData *types.TriggerData) (types.ActionData, error) {
	e.logger.Info("Executing task", "jobID", taskTargetData.JobID)

	switch taskTargetData.TaskDefinitionID {
	case 3, 5:
		return e.executeActionWithStaticArgs(taskTargetData, triggerData)
	case 4, 6:
		return e.executeActionWithDynamicArgs(taskTargetData, triggerData)
	default:
		return types.ActionData{}, fmt.Errorf("unsupported task definition id: %d", taskTargetData.TaskDefinitionID)
	}
}
