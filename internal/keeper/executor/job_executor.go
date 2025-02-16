package executor

import (
	"log"
	"fmt"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type JobExecutor struct {
	// Add any executor-specific configuration or dependencies
}

func NewJobExecutor() *JobExecutor {
	return &JobExecutor{}
}

func (e *JobExecutor) Execute(job *types.Job) (interface{}, error) {
	log.Printf("Executing job: %d (Function: %s)", job.JobID, job.TargetFunction)

	switch job.TargetFunction {
	case "transfer":
		return e.executeTransfer(job)
	case "execute":
		return e.executeGenericContract(job)
	default:
		return nil, fmt.Errorf("unsupported target function: %s", job.TargetFunction)
	}
}

func (e *JobExecutor) executeTransfer(job *types.Job) (interface{}, error) {
	// Simulate transfer logic
	log.Printf("Performing transfer for job %d", job.JobID)

	// Extract transfer details from arguments
	from, ok1 := job.Arguments["from"].(string)
	to, ok2 := job.Arguments["to"].(string)
	amount, ok3 := job.Arguments["amount"].(float64)

	if !ok1 || !ok2 || !ok3 {
		return nil, fmt.Errorf("invalid transfer arguments")
	}

	// Simulated transfer logic
	return map[string]interface{}{
		"status":  "success",
		"from":    from,
		"to":      to,
		"amount":  amount,
		"chainID": job.ChainID,
	}, nil
}

func (e *JobExecutor) executeGenericContract(job *types.Job) (interface{}, error) {
	// Generic contract execution logic
	log.Printf("Executing contract call for job %d", job.JobID)

	return map[string]interface{}{
		"status":    "success",
		"contract":  job.TargetContractAddress,
		"chainID":   job.ChainID,
		"arguments": job.Arguments,
	}, nil
}