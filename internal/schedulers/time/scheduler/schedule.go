package scheduler

import (
	"context"
	"encoding/json"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/config"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/cryptography"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// Start begins the scheduler's main polling and execution loop
func (s *TimeBasedScheduler) Start(ctx context.Context) {
	s.logger.Info("Starting time-based scheduler", "scheduler_signing_address", s.schedulerSigningAddress)

	ticker := time.NewTicker(s.pollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Scheduler context cancelled, stopping")
			return
		case <-s.ctx.Done():
			s.logger.Info("Scheduler stopped")
			return
		case <-ticker.C:
			s.pollAndScheduleTasks()
		}
	}
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
		s.logger.Debug("No tasks found for execution")
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

// processBatch processes a batch of jobs
func (s *TimeBasedScheduler) processBatch(tasks []types.ScheduleTimeTaskData) {
	// Get the performer data
	// TODO: Get the performer data from redis service, which gets it from online keepers list from health service, and sets the performerLock in redis
	// For now, I fixed the performer
	performerData := types.PerformerData{
		KeeperID:      3,
		KeeperAddress: "0x0a067a261c5f5e8c4c0b9137430b4fe1255eb62e",
	}

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
		TaskID:                  primaryTaskID,
		SchedulerSigningAddress: s.schedulerSigningAddress,
	}

	// Create the batch task data
	sendTaskData := types.SendTaskDataToKeeper{
		TaskID:             primaryTaskID,
		PerformerData:      performerData,
		TargetData:         targetDataList,
		TriggerData:        triggerDataList,
		SchedulerSignature: &schedulerSignatureData,
	}

	// Execute the batch
	s.executeTaskBatch(sendTaskData, validTaskIDs)
}

// executeTaskBatch executes a batch of tasks and updates their next execution time
func (s *TimeBasedScheduler) executeTaskBatch(sendTaskData types.SendTaskDataToKeeper, taskIDs []int64) {
	startTime := time.Now()

	s.logger.Infof("Executing batch of %d time-based tasks", len(taskIDs))

	// Sign the task data
	signature, err := cryptography.SignJSONMessage(sendTaskData, config.GetSchedulerSigningKey())
	if err != nil {
		s.logger.Errorf("Failed to sign batch task data: %v", err)
		return
	}
	sendTaskData.SchedulerSignature.SchedulerSignature = signature

	jsonData, err := json.Marshal(sendTaskData)
	if err != nil {
		s.logger.Errorf("Failed to marshal batch task data: %v", err)
		return
	}
	dataBytes := []byte(jsonData)

	broadcastDataForPerformer := types.BroadcastDataForPerformer{
		TaskID:           sendTaskData.TaskID,
		TaskDefinitionID: sendTaskData.TargetData[0].TaskDefinitionID, // Use first task's definition ID
		PerformerAddress: sendTaskData.PerformerData.KeeperAddress,
		Data:             dataBytes,
	}

	// Execute the actual batch job
	executionSuccess := s.performTaskExecution(broadcastDataForPerformer)

	// Track task completion with timing and success status
	executionDuration := time.Since(startTime)
	metrics.TrackTaskCompletion(executionSuccess, executionDuration)

	if executionSuccess {
		s.logger.Infof("Executed batch of %d tasks (IDs: %v) in %v", len(taskIDs), taskIDs, executionDuration)
	} else {
		s.logger.Errorf("Failed to execute batch of %d tasks (IDs: %v) after %v", len(taskIDs), taskIDs, executionDuration)
	}
}

// performTaskExecution handles the actual task execution logic
func (s *TimeBasedScheduler) performTaskExecution(broadcastDataForPerformer types.BroadcastDataForPerformer) bool {
	success, err := s.aggClient.SendTaskToPerformer(s.ctx, &broadcastDataForPerformer)

	if err != nil {
		s.logger.Errorf("Failed to send task to performer: %v", err)
		metrics.TrackTaskBroadcast("failed")
		return false
	}

	metrics.TrackTaskBroadcast("success")
	return success
}

// Stop gracefully stops the scheduler
func (s *TimeBasedScheduler) Stop() {
	startTime := time.Now()
	s.logger.Info("Stopping time-based scheduler")

	// Capture statistics before shutdown
	activeTasksCount := len(s.activeTasks)

	s.cancel()

	duration := time.Since(startTime)

	s.logger.Info("Time-based scheduler stopped",
		"duration", duration,
		"active_tasks_stopped", activeTasksCount,
		"performer_lock_ttl", s.performerLockTTL,
		"task_cache_ttl", s.taskCacheTTL,
		"duplicate_task_window", s.duplicateTaskWindow,
	)
}
