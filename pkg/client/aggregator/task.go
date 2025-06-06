package aggregator

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// SendTaskResult sends a task result to the aggregator
func (c *AggregatorClient) SendTaskResult(ctx context.Context, taskResult *types.PerformerBroadcastData) error {
	c.logger.Debug("Sending task result to aggregator",
		"taskDefinitionId", taskResult.TaskDefinitionID,
		"proofOfTask", taskResult.ProofOfTask)

	// Sign the task data
	signature, err := c.signMessage([]byte(taskResult.IPFSDataCID))
	if err != nil {
		c.logger.Error("Failed to sign task data", "error", err)
		return fmt.Errorf("failed to sign task data: %w", err)
	}

	c.logger.Debug("Task data signed successfully", "signature", signature)

	// Prepare parameters using consistent structure
	params := struct {
		ProofOfTask      string `json:"proofOfTask"`
		Data             string `json:"data"`
		TaskDefinitionID int    `json:"taskDefinitionId"`
		PerformerAddress string `json:"performerAddress"`
		Signature        string `json:"signature"`
	}{
		ProofOfTask:      taskResult.ProofOfTask,
		Data:             "0x" + hex.EncodeToString([]byte(taskResult.IPFSDataCID)),
		TaskDefinitionID: taskResult.TaskDefinitionID,
		PerformerAddress: taskResult.PerformerAddress,
		Signature:        signature,
	}

	var response interface{}
	err = c.executeWithRetry(ctx, "sendTask", &response, params)
	if err != nil {
		c.logger.Error("Failed to send task result", "error", err)
		return fmt.Errorf("failed to send task result: %w", err)
	}

	c.logger.Info("Successfully sent task result to aggregator",
		"taskDefinitionId", taskResult.TaskDefinitionID,
		"proofOfTask", taskResult.ProofOfTask,
		"response", response)

	return nil
}
