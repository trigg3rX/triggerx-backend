package scheduler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// SchedulerTaskRequest represents the request format for Redis API
type SchedulerTaskRequest struct {
	SendTaskDataToKeeper types.SendTaskDataToKeeper `json:"send_task_data_to_keeper"`
	SchedulerID          int                        `json:"scheduler_id"`
	Source               string                     `json:"source"`
}

// RedisAPIResponse represents the response from Redis API
type RedisAPIResponse struct {
	Success   bool                `json:"success"`
	TaskID    int64               `json:"task_id"`
	Message   string              `json:"message"`
	Performer types.PerformerData `json:"performer"`
	Timestamp string              `json:"timestamp"`
	Error     string              `json:"error,omitempty"`
	Details   string              `json:"details,omitempty"`
}

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

	// Use the first task ID as the primary task ID for the batch
	primaryTaskID := validTaskIDs[0]

	s.logger.Infof("Processing batch of %d tasks, primary task ID: %d", len(validTaskIDs), primaryTaskID)

	// Create scheduler signature data
	schedulerSignatureData := types.SchedulerSignatureData{
		TaskID:      primaryTaskID,
		SchedulerID: s.schedulerID,
	}

	// Create the batch task data
	sendTaskData := types.SendTaskDataToKeeper{
		TaskID:             primaryTaskID,
		TargetData:         targetDataList,
		TriggerData:        triggerDataList,
		SchedulerSignature: &schedulerSignatureData,
	}

	// Create request for Redis API
	request := SchedulerTaskRequest{
		SendTaskDataToKeeper: sendTaskData,
		SchedulerID:          s.schedulerID,
		Source:               "time_scheduler",
	}

	// Submit batch to Redis API
	success := s.submitBatchToRedis(request, primaryTaskID, len(validTaskIDs))

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
func (s *TimeBasedScheduler) submitBatchToRedis(request SchedulerTaskRequest, primaryTaskID int64, taskCount int) bool {
	startTime := time.Now()

	// Marshal request to JSON
	requestBytes, err := json.Marshal(request)
	if err != nil {
		s.logger.Error("Failed to marshal batch request",
			"primary_task_id", primaryTaskID,
			"task_count", taskCount,
			"error", err)
		return false
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/scheduler/submit-task", s.redisAPIURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBytes))
	if err != nil {
		s.logger.Error("Failed to create HTTP request",
			"primary_task_id", primaryTaskID,
			"url", url,
			"error", err)
		return false
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error("Failed to send batch to Redis API",
			"primary_task_id", primaryTaskID,
			"task_count", taskCount,
			"url", url,
			"error", err)
		return false
	}
	defer resp.Body.Close()

	// Parse response
	var apiResponse RedisAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		s.logger.Error("Failed to decode Redis API response",
			"primary_task_id", primaryTaskID,
			"status_code", resp.StatusCode,
			"error", err)
		return false
	}

	duration := time.Since(startTime)

	if resp.StatusCode != http.StatusOK {
		s.logger.Error("Redis API returned error",
			"primary_task_id", primaryTaskID,
			"task_count", taskCount,
			"status_code", resp.StatusCode,
			"error", apiResponse.Error,
			"details", apiResponse.Details,
			"duration", duration)
		return false
	}

	if !apiResponse.Success {
		s.logger.Error("Redis API processing failed",
			"primary_task_id", primaryTaskID,
			"task_count", taskCount,
			"message", apiResponse.Message,
			"error", apiResponse.Error,
			"duration", duration)
		return false
	}

	s.logger.Info("Successfully submitted batch to Redis API",
		"primary_task_id", primaryTaskID,
		"task_count", taskCount,
		"performer_id", apiResponse.Performer.KeeperID,
		"performer_address", apiResponse.Performer.KeeperAddress,
		"duration", duration,
		"message", apiResponse.Message)

	return true
}
