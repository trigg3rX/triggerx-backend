package tasks

import (
	"fmt"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// ReceiveTaskFromScheduler is the main entry point for schedulers to submit tasks
func (tsm *TaskStreamManager) ReceiveTaskFromScheduler(request *SchedulerTaskRequest) (error) {
	taskCount := len(request.SendTaskDataToKeeper.TaskID)
	tsm.logger.Info("Receiving task from scheduler",
		"task_ids", request.SendTaskDataToKeeper.TaskID,
		"task_count", taskCount,
		"scheduler_id", request.SchedulerID,
		"source", request.Source)

	// Handle batch requests by creating individual task stream data for each task
	if taskCount > 1 {
		// This is a batch request (likely from time scheduler)
		tsm.logger.Info("Processing batch request", "task_count", taskCount)

		for i := 0; i < taskCount; i++ {
			// Create individual task data for each task in the batch
			individualTaskData := types.SendTaskDataToKeeper{
				TaskID:           []int64{request.SendTaskDataToKeeper.TaskID[i]},
				PerformerData:    request.SendTaskDataToKeeper.PerformerData,
				TargetData:       []types.TaskTargetData{request.SendTaskDataToKeeper.TargetData[i]},
				TriggerData:      []types.TaskTriggerData{request.SendTaskDataToKeeper.TriggerData[i]},
				SchedulerID:      request.SendTaskDataToKeeper.SchedulerID,
				ManagerSignature: request.SendTaskDataToKeeper.ManagerSignature,
			}

			taskStreamData := TaskStreamData{
				JobID:                individualTaskData.TargetData[0].JobID,
				TaskDefinitionID:     individualTaskData.TargetData[0].TaskDefinitionID,
				CreatedAt:            time.Now(),
				RetryCount:           0,
				SendTaskDataToKeeper: individualTaskData,
			}

			// Add individual task to batch processor
			_, err := tsm.AddTaskToReadyStream(taskStreamData)
			if err != nil {
				tsm.logger.Error("Failed to add individual task to batch processor",
					"task_id", individualTaskData.TaskID[0],
					"batch_index", i,
					"source", request.Source,
					"error", err)
				// Continue processing other tasks in the batch
				continue
			}

			tsm.logger.Debug("Individual task added to batch processor",
				"task_id", individualTaskData.TaskID[0],
				"batch_index", i)
		}
	} else {
		// This is a single task request (likely from condition scheduler)
		taskStreamData := TaskStreamData{
			JobID:                request.SendTaskDataToKeeper.TargetData[0].JobID,
			TaskDefinitionID:     request.SendTaskDataToKeeper.TargetData[0].TaskDefinitionID,
			CreatedAt:            time.Now(),
			RetryCount:           0,
			SendTaskDataToKeeper: request.SendTaskDataToKeeper,
		}

		// Add task to batch processor for improved performance
		_, err := tsm.AddTaskToReadyStream(taskStreamData)
		if err != nil {
			tsm.logger.Error("Failed to add task to batch processor",
				"task_id", request.SendTaskDataToKeeper.TaskID[0],
				"source", request.Source,
				"error", err)
			return fmt.Errorf("failed to add task to batch processor: %w", err)
		}
	}

	tsm.logger.Info("Tasks received and added to ready stream",
		"task_count", taskCount,
		"source", request.Source)

	return nil
}

func (tsm *TaskStreamManager) UpdateDatabase(ipfsData types.IPFSData) {
	tsm.logger.Info("Updating task stream and database ...")

	taskID := ipfsData.TaskData.TaskID[0]
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
		TaskID: ipfsData.TaskData.TaskID[0],
		TaskPerformerID: ipfsData.TaskData.PerformerData.OperatorID,
		ExecutionTimestamp: ipfsData.ActionData.ExecutionTimestamp,
		ExecutionTxHash: ipfsData.ActionData.ActionTxHash,
		ProofOfTask: ipfsData.ProofData.ProofOfTask,
		TaskOpXCost: ipfsData.ActionData.TotalFee,
	}
	tsm.logger.Infof("UpdateTaskExecutionDataRequest: %+v", updateTaskExecutionData)

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