package aggregator

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

// SendTaskToPerformer sends a task to the specified performer through the aggregator
func (c *AggregatorClient) SendTaskToPerformer(ctx context.Context, taskData *types.BroadcastDataForPerformer) (bool, error) {
	c.logger.Debug(ctx, "Sending task to performer",
		observability.Int("TaskID", int(taskData.TaskID)),
		observability.String("PerformerAddress", taskData.PerformerAddress))

	// Prepare parameters using consistent structure
	params := CallParams{
		Data:             "0x" + hex.EncodeToString(taskData.Data),
		TaskDefinitionID: taskData.TaskDefinitionID,
	}

	var result interface{}
	err := c.executeWithRetry(ctx, "sendCustomMessage", &result, params)
	if err != nil {
		c.logger.Error(ctx, "Failed to send custom task", observability.Error(err))
		return false, fmt.Errorf("failed to send custom task: %w", err)
	}

	c.logger.Debug(ctx, "Task sent successfully",
		observability.Int("TaskID", int(taskData.TaskID)),
		observability.Any("result", result))
	return true, nil
}
