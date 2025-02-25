package execution

import (
	"fmt"
	"time"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// JobExecutor handles execution of blockchain transactions and contract calls
// Can be extended with additional configuration and dependencies as needed
type JobExecutor struct {}

func NewJobExecutor() *JobExecutor {
	return &JobExecutor{}
}

// Execute routes jobs to appropriate handlers based on the target function
// Currently supports 'transfer' for token transfers and 'execute' for generic contract calls
func (e *JobExecutor) Execute(job *types.HandleCreateJobData) (types.ActionData, error) {
	logger.Infof("Executing job: %d (Function: %s)", job.JobID, job.TargetFunction)

	switch job.TaskDefinitionID {
	case 1, 3, 5:
		return e.executeActionWithStaticArgs(job)
	case 2, 4, 6:
		return e.executeActionWithDynamicArgs(job)
	default:
		return types.ActionData{}, fmt.Errorf("unsupported task definition id: %d", job.TaskDefinitionID)
	}
}

func (e *JobExecutor) executeActionWithStaticArgs(job *types.HandleCreateJobData) (types.ActionData, error) {
	logger.Infof("Executing timed action for job %d", job.JobID)

	var executionResult types.ActionData
	executionResult.TaskID = 0
	executionResult.ActionTxHash = "0x"
	executionResult.GasUsed = "0"
	executionResult.Status = true
	executionResult.Timestamp = time.Now()

	return executionResult, nil
}

func (e *JobExecutor) executeActionWithDynamicArgs(job *types.HandleCreateJobData) (types.ActionData, error) {
	logger.Infof("Executing condition action for job %d", job.JobID)

	var executionResult types.ActionData
	executionResult.TaskID = 0
	executionResult.ActionTxHash = "0x"
	executionResult.GasUsed = "0"
	executionResult.Status = true
	executionResult.Timestamp = time.Now()

	return executionResult, nil
}
