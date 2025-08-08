package tasks

import (
	"context"
	"encoding/json"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/taskdispatcher/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// sendTaskToPerformer sends the task to the assigned performer
func (tsm *TaskStreamManager) sendTaskToPerformer(task TaskStreamData) {
	taskID := task.SendTaskDataToKeeper.TaskID[0]

	tsm.logger.Info("Sending task to performer", "task_id", taskID)

	// Send to aggregator/performer using existing method
	jsonData, err := json.Marshal(task.SendTaskDataToKeeper)
	if err != nil {
		tsm.logger.Errorf("Failed to marshal batch task data: %v", err)
		return
	}
	dataBytes := []byte(jsonData)

	broadcastDataForPerformer := types.BroadcastDataForPerformer{
		TaskID:           task.SendTaskDataToKeeper.TaskID[0],
		TaskDefinitionID: task.SendTaskDataToKeeper.TargetData[0].TaskDefinitionID,
		PerformerAddress: task.SendTaskDataToKeeper.PerformerData.KeeperAddress,
		Data:             dataBytes,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	success, err := tsm.aggClient.SendTaskToPerformer(ctx, &broadcastDataForPerformer)
	if err != nil {
		tsm.logger.Error("Failed to send task to performer",
			"task_id", taskID,
			"error", err)

		// Move task to failed stream
		if moveErr := tsm.moveTaskToFailed(task, err.Error()); moveErr != nil {
			tsm.logger.Error("Failed to move task to failed stream",
				"task_id", taskID,
				"error", moveErr)
		}
		return
	}

	if success {
		tsm.logger.Info("Task sent to performer successfully", "task_id", taskID)
		metrics.TasksAddedToStreamTotal.WithLabelValues("processing", "success").Inc()
	} else {
		tsm.logger.Warn("Task sending to performer was not successful", "task_id", taskID)
		if moveErr := tsm.moveTaskToFailed(task, "performer send failed"); moveErr != nil {
			tsm.logger.Error("Failed to move task to failed stream",
				"task_id", taskID,
				"error", moveErr)
		}
	}
}
