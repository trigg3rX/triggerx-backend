package aggregator

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// SendTaskToPerformer sends a task to the specified performer through the aggregator
func (c *AggregatorClient) SendTaskToPerformer(ctx context.Context, taskData *types.BroadcastDataForPerformer) (bool, error) {
	c.logger.Debug("Sending task to performer",
		"TaskID", taskData.TaskID,
		"PerformerAddress", taskData.PerformerAddress)

	// Prepare parameters using consistent structure
	params := struct {
		Data             string `json:"data"`
		TaskDefinitionID int    `json:"taskDefinitionId"`
	}{
		Data:             "0x" + hex.EncodeToString(taskData.Data),
		TaskDefinitionID: taskData.TaskDefinitionID,
	}

	var result interface{}
	err := c.executeWithRetry(ctx, "sendCustomMessage", &result, params)
	if err != nil {
		c.logger.Error("Failed to send custom task", "error", err)
		return false, fmt.Errorf("failed to send custom task: %w", err)
	}

	c.logger.Info("Task sent successfully",
		"TaskID", taskData.TaskID,
		"result", result)
	return true, nil
}
