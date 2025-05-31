package aggregator

import (
	"context"
	"encoding/hex"
	"fmt"
)

// TaskResult represents the data to be sent to the aggregator
type TaskResult struct {
	ProofOfTask      string
	Data             string
	TaskDefinitionID int
	PerformerAddress string
}

// SendTaskResult sends a task result to the aggregator
func (c *AggregatorClient) SendTaskResult(ctx context.Context, result *TaskResult) error {
	// Sign the task data
	signature, err := c.signMessage([]byte(result.Data))
	if err != nil {
		return fmt.Errorf("failed to sign task data: %w", err)
	}

	// Prepare parameters
	params := struct {
		ProofOfTask      string `json:"proofOfTask"`
		Data             string `json:"data"`
		TaskDefinitionID int    `json:"taskDefinitionId"`
		PerformerAddress string `json:"performerAddress"`
		Signature        string `json:"signature"`
	}{
		ProofOfTask:      result.ProofOfTask,
		Data:             "0x" + hex.EncodeToString([]byte(result.Data)),
		TaskDefinitionID: result.TaskDefinitionID,
		PerformerAddress: result.PerformerAddress,
		Signature:        signature,
	}

	var response interface{}
	err = c.executeWithRetry(ctx, "sendTask", &response, params)
	if err != nil {
		return fmt.Errorf("failed to send task result: %w", err)
	}

	c.logger.Info("Successfully sent task result to aggregator",
		"taskDefinitionId", result.TaskDefinitionID,
		"proofOfTask", result.ProofOfTask)

	return nil
}
