package scheduler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// pollAndScheduleTasks fetches tasks from database and schedules them for execution
func (s *TimeBasedScheduler) pollAndScheduleTasks() {
	tasks, err := s.dbClient.GetTimeBasedTasks()
	if err != nil {
		s.logger.Errorf("Failed to fetch time-based tasks: %v", err)
		metrics.TrackDBConnectionError()
		return
	}

	if len(tasks) == 0 {
		return
	}

	s.logger.Infof("Found %d tasks to process", len(tasks))
	metrics.TasksScheduled.Set(float64(len(tasks)))
	metrics.TaskBatchSize.Set(float64(s.taskBatchSize))

	// Separate tasks based on is_imua flag
	var imuaTasks []types.ScheduleTimeTaskData
	var nonImuaTasks []types.ScheduleTimeTaskData

	for _, task := range tasks {
		if task.IsImua {
			imuaTasks = append(imuaTasks, task)
		} else {
			nonImuaTasks = append(nonImuaTasks, task)
		}
	}

	// Process non-imua tasks in batches
	if len(nonImuaTasks) > 0 {
		s.logger.Infof("Processing %d non-imua tasks in batches", len(nonImuaTasks))
		for i := 0; i < len(nonImuaTasks); i += s.taskBatchSize {
			end := i + s.taskBatchSize
			if end > len(nonImuaTasks) {
				end = len(nonImuaTasks)
			}

			batch := nonImuaTasks[i:end]
			s.processBatch(batch)
		}
	}

	// Process imua tasks in separate batches
	if len(imuaTasks) > 0 {
		s.logger.Infof("Processing %d imua tasks in separate batches", len(imuaTasks))
		for i := 0; i < len(imuaTasks); i += s.taskBatchSize {
			end := i + s.taskBatchSize
			if end > len(imuaTasks) {
				end = len(imuaTasks)
			}

			batch := imuaTasks[i:end]
			s.processBatch(batch)
		}
	}
}

// processBatch processes a batch of tasks by submitting to task dispatcher
func (s *TimeBasedScheduler) processBatch(tasks []types.ScheduleTimeTaskData) {
	s.logger.Infof("Processing batch of %d time-based tasks", len(tasks))

	var targetDataList []types.TaskTargetData
	var triggerDataList []types.TaskTriggerData
	var validTaskIDs []int64

	for _, task := range tasks {
		// Check if ExpirationTime of the job has passed or not
		if task.ExpirationTime.Before(time.Now()) {
			s.logger.Infof("Task ID %d has expired, skipping execution", task.TaskID)
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
		s.logger.Debug("No valid tasks in batch after filtering expired tasks")
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
	success := s.submitBatchToTaskDispatcher(request, taskIDs, len(validTaskIDs))

	if success {
		s.logger.Infof("Batch processing completed successfully: %d tasks submitted", len(validTaskIDs))
		metrics.TrackTaskCompletion(true, time.Since(time.Now()))
		metrics.TrackTaskBroadcast("task_dispatcher_submitted")
	} else {
		s.logger.Errorf("Batch processing failed: %d tasks", len(validTaskIDs))
		metrics.TrackTaskBroadcast("failed")
	}
}

// submitBatchToTaskDispatcher submits the batch task data to Task Dispatcher via RPC
func (s *TimeBasedScheduler) submitBatchToTaskDispatcher(request types.SchedulerTaskRequest, taskIDs string, taskCount int) bool {
	startTime := time.Now()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Make RPC call to task dispatcher
	var response types.TaskManagerAPIResponse
	err := s.taskDispatcherClient.Call(ctx, "submit-task", &request, &response)
	if err != nil {
		s.logger.Error("Failed to submit batch to task dispatcher via RPC",
			"task_ids", taskIDs,
			"task_count", taskCount,
			"error", err)
		return false
	}

	duration := time.Since(startTime)

	if !response.Success {
		s.logger.Error("Task dispatcher processing failed",
			"task_ids", taskIDs,
			"task_count", taskCount,
			"message", response.Message,
			"error", response.Error,
			"duration", duration)
		return false
	}

	s.logger.Info("Successfully submitted batch to task dispatcher",
		"task_ids", taskIDs,
		"task_count", taskCount,
		"response_task_ids", response.TaskID,
		"duration", duration,
		"message", response.Message)

	return true
}
