package execution

import (
	"fmt"

	"github.com/trigg3rX/triggerx-backend/pkg/client/aggregator"
	"github.com/trigg3rX/triggerx-backend/pkg/docker"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// TaskExecutor is the default implementation of TaskExecutor
type TaskExecutor struct {
	alchemyAPIKey   string
	etherscanAPIKey string
	codeExecutor    *docker.CodeExecutor
	argConverter    *ArgumentConverter
	aggregatorClient *aggregator.AggregatorClient
	logger          logging.Logger
}

// NewTaskExecutor creates a new instance of TaskExecutor
func NewTaskExecutor(alchemyAPIKey string, etherscanAPIKey string, codeExecutor *docker.CodeExecutor, aggregatorClient *aggregator.AggregatorClient, logger logging.Logger) *TaskExecutor {
	return &TaskExecutor{
		alchemyAPIKey:   alchemyAPIKey,
		etherscanAPIKey: etherscanAPIKey,
		codeExecutor:    codeExecutor,
		argConverter:    &ArgumentConverter{},
		aggregatorClient: aggregatorClient,
		logger: logger,
	}
}

// Execute implements the TaskExecutor interface
func (e *TaskExecutor) ExecuteTimeBasedTask(timeJobData *types.ScheduleTimeJobData) (types.PerformerActionData, error) {
	e.logger.Info("Executing time based task", "jobID", timeJobData.JobID)

	// TODO: Execute the task
	return types.PerformerActionData{}, nil
}


func (e *TaskExecutor) ExecuteTask(taskTargetData *types.SendTaskTargetDataToKeeper, triggerData *types.SendTaskTriggerDataToKeeper) (types.PerformerActionData, error) {
	e.logger.Info("Executing task", "jobID", taskTargetData.TaskID)

	switch taskTargetData.TaskDefinitionID {
	case 3, 5:
		return e.executeActionWithStaticArgs(taskTargetData, triggerData)
	case 4, 6:
		return e.executeActionWithDynamicArgs(taskTargetData, triggerData)
	default:
		return types.PerformerActionData{}, fmt.Errorf("unsupported task definition id: %d", taskTargetData.TaskDefinitionID)
	}
}
