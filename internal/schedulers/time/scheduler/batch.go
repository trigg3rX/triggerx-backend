package scheduler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// processBatch processes a batch of tasks by submitting them to the task dispatcher.
// It filters out expired tasks, builds the task data structures, and submits
// the batch via RPC to the task dispatcher service.
func (s *TimeBasedScheduler) processBatch(ctx context.Context, tasks []types.ScheduleTimeTaskData) {
	s.logger.Debug(ctx, "Processing batch of time-based tasks", observability.Int("task_count", len(tasks)))

	var targetDataList []types.TaskTargetData
	var triggerDataList []types.TaskTriggerData
	var validTaskIDs []int64

	for _, task := range tasks {
		// Check if ExpirationTime of the job has passed or not
		if task.ExpirationTime.Before(time.Now()) {
			s.logger.Info(ctx, "Task has expired, skipping execution", observability.Int64("task_id", task.TaskID))
			metrics.TrackTaskExpired()
			continue
		}

		// Track task by schedule type
		metrics.TrackTaskByScheduleType(task.ScheduleType)

		// Generate the task data to send to the performer
		targetData := types.TaskTargetData{
			JobID:                     task.TaskTargetData.JobID,
			TaskID:                    task.TaskID,
			TaskDefinitionID:          task.TaskDefinitionID,
			TargetChainID:             task.TaskTargetData.TargetChainID,
			TargetContractAddress:     task.TaskTargetData.TargetContractAddress,
			TargetFunction:            task.TaskTargetData.TargetFunction,
			ABI:                       task.TaskTargetData.ABI,
			ArgType:                   task.TaskTargetData.ArgType,
			Arguments:                 task.TaskTargetData.Arguments,
			DynamicArgumentsScriptUrl: task.TaskTargetData.DynamicArgumentsScriptUrl,
			IsImua:                    task.IsImua,
		}
		triggerData := types.TaskTriggerData{
			TaskID:                  task.TaskID,
			TaskDefinitionID:        task.TaskDefinitionID,
			ExpirationTime:          task.ExpirationTime,
			CurrentTriggerTimestamp: task.LastExecutedAt,
			NextTriggerTimestamp:    task.NextExecutionTimestamp,
			TimeScheduleType:        task.ScheduleType,
			TimeCronExpression:      task.CronExpression,
			TimeSpecificSchedule:    task.SpecificSchedule,
			TimeInterval:            task.TimeInterval,
		}

		targetDataList = append(targetDataList, targetData)
		triggerDataList = append(triggerDataList, triggerData)
		validTaskIDs = append(validTaskIDs, task.TaskID)
	}

	// If no valid tasks, return early
	if len(validTaskIDs) == 0 {
		s.logger.Debug(ctx, "No valid tasks in batch after filtering expired tasks")
		return
	}

	// Create the batch task data
	sendTaskData := types.SendTaskDataToKeeper{
		TaskID:           validTaskIDs,
		TargetData:       targetDataList,
		TriggerData:      triggerDataList,
		SchedulerID:      s.schedulerID,
		ManagerSignature: "",
	}

	// Create request for task dispatcher
	request := types.SchedulerTaskRequest{
		SendTaskDataToKeeper: sendTaskData,
		Source:               "time_scheduler",
	}

	// Convert validTaskIDs ([]int64) to []string for joining
	taskIDStrs := make([]string, len(validTaskIDs))
	for i, id := range validTaskIDs {
		taskIDStrs[i] = fmt.Sprintf("%d", id)
	}
	taskIDs := strings.Join(taskIDStrs, ", ")

	// Submit batch to task dispatcher
	success := s.submitBatchToTaskDispatcher(ctx, request, taskIDs, len(validTaskIDs))

	if success {
		s.logger.Info(ctx, "Batch processing completed successfully", observability.Int("task_count", len(validTaskIDs)))
		metrics.UpdateTasksDispatched("success")
	} else {
		s.logger.Error(ctx, "Batch processing failed", observability.Int("task_count", len(validTaskIDs)))
		metrics.UpdateTasksDispatched("failed")
	}
}
