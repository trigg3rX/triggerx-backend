package executor

import (
	"fmt"
	"github.com/trigg3rX/triggerx-backend/execute/manager"
	"log"
)

type JobExecutor struct {
	// Add any executor-specific configuration or dependencies
}

func NewJobExecutor() *JobExecutor {
	return &JobExecutor{}
}

func (e *JobExecutor) Execute(job *manager.Job) (interface{}, error) {
	log.Printf("Executing job: %s (Function: %s)", job.JobID, job.TargetFunction)

	switch job.TargetFunction {
	case "transfer":
		return e.executeTransfer(job)
	case "execute":
		return e.executeGenericContract(job)
	default:
		return nil, fmt.Errorf("unsupported target function: %s", job.TargetFunction)
	}
}

func (e *JobExecutor) executeTransfer(job *manager.Job) (interface{}, error) {
	// Simulate transfer logic
	log.Printf("Performing transfer for job %s", job.JobID)

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

func (e *JobExecutor) executeGenericContract(job *manager.Job) (interface{}, error) {
	// Generic contract execution logic
	log.Printf("Executing contract call for job %s", job.JobID)

	return map[string]interface{}{
		"status":    "success",
		"contract":  job.ContractAddress,
		"chainID":   job.ChainID,
		"arguments": job.Arguments,
	}, nil
}
