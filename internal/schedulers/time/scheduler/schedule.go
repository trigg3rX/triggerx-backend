package scheduler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
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

	for i := 0; i < len(tasks); i += s.taskBatchSize {
		end := i + s.taskBatchSize
		if end > len(tasks) {
			end = len(tasks)
		}

		batch := tasks[i:end]
		s.processBatch(batch)
	}
}

// processBatch processes a batch of tasks by submitting to Redis API
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

	// Create request for Redis API
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

	// Submit batch to Redis API
	success := s.submitBatchToRedis(request, taskIDs, len(validTaskIDs))

	if success {
		s.logger.Infof("Batch processing completed successfully: %d tasks submitted", len(validTaskIDs))
		metrics.TrackTaskCompletion(true, time.Since(time.Now()))
		metrics.TrackTaskBroadcast("redis_submitted")
	} else {
		s.logger.Errorf("Batch processing failed: %d tasks", len(validTaskIDs))
		metrics.TrackTaskBroadcast("failed")
	}
}

// submitBatchToRedis submits the batch task data to Redis API
func (s *TimeBasedScheduler) submitBatchToRedis(request types.SchedulerTaskRequest, taskIDs string, taskCount int) bool {
	startTime := time.Now()

	// Marshal request to JSON
	requestBytes, err := json.Marshal(request)
	if err != nil {
		s.logger.Error("Failed to marshal batch request",
			"task_ids", taskIDs,
			"task_count", taskCount,
			"error", err)
		return false
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/scheduler/submit-task", s.redisAPIURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBytes))
	if err != nil {
		s.logger.Error("Failed to create HTTP request",
			"task_ids", taskIDs,
			"url", url,
			"error", err)
		return false
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error("Failed to send batch to Redis API",
			"task_ids", taskIDs,
			"task_count", taskCount,
			"url", url,
			"error", err)
		return false
	}
	defer func() { _ = resp.Body.Close() }()

	// Parse response
	var apiResponse types.TaskManagerAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		s.logger.Error("Failed to decode TaskManager response",
			"task_ids", taskIDs,
			"status_code", resp.StatusCode,
			"error", err)
		return false
	}

	duration := time.Since(startTime)

	if resp.StatusCode != http.StatusOK {
		s.logger.Error("TaskManager returned error",
			"task_ids", taskIDs,
			"task_count", taskCount,
			"status_code", resp.StatusCode,
			"error", apiResponse.Error,
			"details", apiResponse.Details,
			"duration", duration)
		return false
	}

	if !apiResponse.Success {
		s.logger.Error("TaskManager processing failed",
			"task_ids", taskIDs,
			"task_count", taskCount,
			"message", apiResponse.Message,
			"error", apiResponse.Error,
			"duration", duration)
		return false
	}

	s.logger.Info("Successfully submitted batch to TaskManager",
		"task_ids", taskIDs,
		"task_count", taskCount,
		"response_task_ids", apiResponse.TaskID,
		"duration", duration,
		"message", apiResponse.Message)

	return true
}
