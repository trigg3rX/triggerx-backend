package stream

import (
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
)


func (tsm *TaskStreamManager) UpdateDatabase(ipfsData types.IPFSData) {
	tsm.logger.Info("Updating task stream and database ...")

	taskID := ipfsData.TaskData.TaskID
	taskStreamData, err := tsm.getTaskStreamData(taskID)
	if err != nil {
		tsm.logger.Error("Failed to read task stream data", "error", err)
		return
	}

	if taskStreamData == nil {
		tsm.logger.Error("Task stream data not found", "task_id", taskID)
		return
	}

	now := time.Now()
	taskStreamData.CompletedAt = &now

	// Update task stream
	err = tsm.addTaskToStream(TasksCompletedStream, taskStreamData)
	if err != nil {
		tsm.logger.Error("Failed to add task to completed stream", "error", err)
	}


	// Update task execution data in Database
	updateTaskExecutionData := types.UpdateTaskExecutionDataRequest{
		TaskID: ipfsData.TaskData.TaskID,
		TaskPerformerID: ipfsData.TaskData.PerformerData.KeeperID,
		ExecutionTimestamp: ipfsData.ActionData.ExecutionTimestamp,
		ExecutionTxHash: ipfsData.ActionData.ActionTxHash,
		ProofOfTask: ipfsData.ProofData.ProofOfTask,
		TaskOpXCost: ipfsData.ActionData.TotalFee,
	}

	success, err := tsm.dbClient.UpdateTaskExecutionData(updateTaskExecutionData)
	if err != nil {
		tsm.logger.Error("Failed to update task execution data", "error", err)
	}

	if success {
		tsm.logger.Info("Task execution data updated successfully")
	} else {
		tsm.logger.Error("Failed to update task execution data")
	}

	tsm.logger.Info("Task stream and database updated successfully")
}