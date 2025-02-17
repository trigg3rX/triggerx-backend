package execution

import (
	"fmt"

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
func (e *JobExecutor) Execute(job *types.Job) (interface{}, error) {
	logger.Info("[Execution] Executing job: %d (Function: %s)", job.JobID, job.TargetFunction)

	switch job.TargetFunction {
	case "transfer":
		return e.executeTransfer(job)
	case "execute":
		return e.executeGenericContract(job)
	default:
		return nil, fmt.Errorf("unsupported target function: %s", job.TargetFunction)
	}
}

// executeTransfer handles token transfer transactions
// Expects 'from', 'to' addresses and 'amount' in job arguments
func (e *JobExecutor) executeTransfer(job *types.Job) (interface{}, error) {
	logger.Info("[Execution] Performing transfer for job %d", job.JobID)

	from, ok1 := job.Arguments["from"].(string)
	to, ok2 := job.Arguments["to"].(string)
	amount, ok3 := job.Arguments["amount"].(float64)

	if !ok1 || !ok2 || !ok3 {
		return nil, fmt.Errorf("invalid transfer arguments")
	}

	return map[string]interface{}{
		"status":  "success", 
		"from":    from,
		"to":      to,
		"amount":  amount,
		"chainID": job.ChainID,
	}, nil
}

// executeGenericContract handles arbitrary contract function calls
// Uses contract address and arguments specified in the job
func (e *JobExecutor) executeGenericContract(job *types.Job) (interface{}, error) {
	logger.Info("[Execution] Executing contract call for job %d", job.JobID)

	return map[string]interface{}{
		"status":    "success",
		"contract":  job.TargetContractAddress,
		"chainID":   job.ChainID,
		"arguments": job.Arguments,
	}, nil
}