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
			s.pollAndScheduleJobs()
		}
	}
}

// pollAndScheduleJobs fetches jobs from database and schedules them for execution
func (s *TimeBasedScheduler) pollAndScheduleJobs() {
	jobs, err := s.dbClient.GetTimeBasedJobs()
	if err != nil {
		s.logger.Errorf("Failed to fetch time-based jobs: %v", err)
		metrics.TasksCompleted.WithLabelValues("failed").Inc()

		return
	}

	if len(jobs) == 0 {
		s.logger.Debug("No tasks found for execution")
		return
	}

	s.logger.Infof("Found %d tasks to process", len(jobs))
	metrics.TasksScheduled.Set(float64(len(jobs)))

	for i := 0; i < len(jobs); i += s.jobBatchSize {
		end := i + s.jobBatchSize
		if end > len(jobs) {
			end = len(jobs)
		}

		batch := jobs[i:end]
		s.processBatch(batch)
	}
}

// processBatch processes a batch of jobs
func (s *TimeBasedScheduler) processBatch(tasks []types.ScheduleTimeTaskData) {
	for _, task := range tasks {
		// Add job to execution queue
		select {
		case s.taskQueue <- &task:
			s.executeJob(&task)
			metrics.TasksCompleted.WithLabelValues("success").Inc()
			s.logger.Debugf("Queued task %d for execution", task.TaskID)
		default:
			s.logger.Warnf("Task queue is full, skipping task %d", task.TaskID)
			metrics.TasksCompleted.WithLabelValues("failed").Inc()
		}
	}
}

// executeJob executes a single job and updates its next execution time
func (s *TimeBasedScheduler) executeJob(task *types.ScheduleTimeTaskData) {
	startTime := time.Now()

	s.logger.Infof("Executing time-based task %d (type: %s) for job %d", task.TaskID, task.ScheduleType, task.TaskTargetData.JobID)

	// Check if ExpirationTime of the job has passed or not
	if task.ExpirationTime.Before(time.Now()) {
		s.logger.Infof("Job for this task ID %d has expired, skipping execution", task.TaskID)
		return
	}

	// Get the performer data
	// TODO: Get the performer data from redis service, which gets it from online keepers list from health service, and sets the performerLock in redis
	// For now, I fixed the performer
	performerData := types.PerformerData{
		KeeperID:      3,
		KeeperAddress: "0x0a067a261c5f5e8c4c0b9137430b4fe1255eb62e",
	}

	// Generate the task data to send to the performer
	targetData := types.TaskTargetData{
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
		ExpirationTime:          task.ExpirationTime,
		CurrentTriggerTimestamp: time.Now(),
		TimeScheduleType:        task.ScheduleType,
		TimeCronExpression:      task.CronExpression,
		TimeSpecificSchedule:    task.SpecificSchedule,
		TimeInterval:            task.TimeInterval,
	}
	schedulerSignatureData := types.SchedulerSignatureData{
		TaskID:                  task.TaskID,
		SchedulerSigningAddress: s.schedulerSigningAddress,
	}
	sendTaskData := types.SendTaskDataToKeeper{
		TaskID:             task.TaskID,
		PerformerData:      performerData,
		TargetData:         []types.TaskTargetData{targetData},
		TriggerData:        []types.TaskTriggerData{triggerData},
		SchedulerSignature: &schedulerSignatureData,
	}

	// Sign the task data
	signature, err := cryptography.SignJSONMessage(sendTaskData, config.GetSchedulerSigningKey())
	if err != nil {
		s.logger.Errorf("Failed to sign task data: %v", err)
		return
	}
	sendTaskData.SchedulerSignature.SchedulerSignature = signature

	jsonData, err := json.Marshal(sendTaskData)
	if err != nil {
		s.logger.Errorf("Failed to marshal task data: %v", err)
		return
	}
	dataBytes := []byte(jsonData)

	broadcastDataForPerformer := types.BroadcastDataForPerformer{
		TaskID:           task.TaskID,
		TaskDefinitionID: task.TaskDefinitionID,
		PerformerAddress: performerData.KeeperAddress,
		Data:             dataBytes,
	}

	// Execute the actual job
	executionSuccess := s.performJobExecution(broadcastDataForPerformer)

	if executionSuccess {
		metrics.TasksCompleted.WithLabelValues("success").Inc()
		metrics.TaskExecutionTime.Observe(time.Since(startTime).Seconds())
		s.logger.Infof("Executed task ID %d for job %d in %v", task.TaskID, task.TaskTargetData.JobID, time.Since(startTime))

	} else {
		metrics.TasksCompleted.WithLabelValues("failed").Inc()
		s.logger.Errorf("Failed to execute task %d for job %d after %v", task.TaskID, task.TaskTargetData.JobID, time.Since(startTime))
	}
}

// performJobExecution handles the actual job execution logic
func (s *TimeBasedScheduler) performJobExecution(broadcastDataForPerformer types.BroadcastDataForPerformer) bool {
	success, err := s.aggClient.SendTaskToPerformer(s.ctx, &broadcastDataForPerformer)

	if err != nil {
		s.logger.Errorf("Failed to send task to performer: %v", err)
		return false
	}

	return success
}

// Stop gracefully stops the scheduler
func (s *TimeBasedScheduler) Stop() {
	startTime := time.Now()
	s.logger.Info("Stopping time-based scheduler")

	// Capture statistics before shutdown
	activeTasksCount := len(s.activeTasks)
	queueLength := len(s.taskQueue)

	s.cancel()

	// Close job queue
	close(s.taskQueue)

	duration := time.Since(startTime)

	s.logger.Info("Time-based scheduler stopped",
		"duration", duration,
		"active_tasks_stopped", activeTasksCount,
		"queue_length", queueLength,
		"performer_lock_ttl", s.performerLockTTL,
		"task_cache_ttl", s.taskCacheTTL,
		"duplicate_task_window", s.duplicateTaskWindow,
	)
}
