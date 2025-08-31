package scheduler

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	dbserverTypes "github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/scheduler/worker"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// handleTriggerNotification processes trigger notifications and submits individual tasks to Redis API
func (s *ConditionBasedScheduler) handleTriggerNotification(notification *worker.TriggerNotification) error {
	startTime := time.Now()

	s.logger.Info("Processing trigger notification - submitting task to task dispatcher",
		"job_id", notification.JobID,
		"trigger_value", notification.TriggerValue,
		"trigger_tx_hash", notification.TriggerTxHash,
		"triggered_at", notification.TriggeredAt,
	)

	// Acquire notification mutex to prevent cleanup during processing
	s.notificationMutex.Lock()
	defer s.notificationMutex.Unlock()

	// Get the job data from storage
	s.workersMutex.RLock()
	jobData, exists := s.jobDataStore[notification.JobID.String()]
	s.workersMutex.RUnlock()

	if !exists || jobData == nil {
		s.logger.Error("Job data not found", "job_id", notification.JobID)
		return fmt.Errorf("job data not found for job %d", notification.JobID)
	}

	createTaskRequest := dbserverTypes.CreateTaskDataRequest{
		JobID:            jobData.JobID.ToBigInt(),
		TaskDefinitionID: jobData.TaskDefinitionID,
	}

	// Create Task in Database
	taskID, err := s.dbClient.CreateTask(createTaskRequest)
	if err != nil {
		s.logger.Error("Failed to create task in database", "job_id", notification.JobID, "error", err)
		return fmt.Errorf("failed to create task in database: %w", err)
	}
	jobData.TaskTargetData.TaskID = taskID

	// Create individual task and submit to task dispatcher
	success, err := s.submitTriggeredTaskToTaskDispatcher(jobData, notification)
	if err != nil {
		s.logger.Error("Failed to submit triggered task to task dispatcher",
			"job_id", notification.JobID,
			"error", err,
		)
		metrics.TrackCriticalError("task_dispatcher_submission_failed")
		return err
	}

	duration := time.Since(startTime)
	if success {
		s.logger.Info("Successfully submitted triggered task to task dispatcher",
			"job_id", notification.JobID,
			"duration", duration,
		)
		metrics.TrackActionExecution(fmt.Sprintf("%d", notification.JobID), duration)
	} else {
		s.logger.Error("Failed to submit triggered task to task dispatcher",
			"job_id", notification.JobID,
			"duration", duration,
		)
		metrics.TrackCriticalError("task_dispatcher_submission_failed")
	}

	return nil
}

// submitTriggeredTaskToTaskManager creates and submits a single task to TaskManager when triggers occur
func (s *ConditionBasedScheduler) submitTriggeredTaskToTaskDispatcher(jobData *types.ScheduleConditionJobData, notification *worker.TriggerNotification) (bool, error) {
	s.logger.Info("Creating triggered task for task dispatcher submission",
		"job_id", jobData.JobID,
		"task_definition_id", jobData.TaskDefinitionID,
		"trigger_value", notification.TriggerValue)

	// Create single task data (not batch like time scheduler)
	targetData := types.TaskTargetData{
		JobID:                     jobData.JobID,
		TaskID:                    jobData.TaskTargetData.TaskID,
		TaskDefinitionID:          jobData.TaskDefinitionID,
		TargetChainID:             jobData.TaskTargetData.TargetChainID,
		TargetContractAddress:     jobData.TaskTargetData.TargetContractAddress,
		TargetFunction:            jobData.TaskTargetData.TargetFunction,
		ABI:                       jobData.TaskTargetData.ABI,
		ArgType:                   jobData.TaskTargetData.ArgType,
		Arguments:                 jobData.TaskTargetData.Arguments,
		DynamicArgumentsScriptUrl: jobData.TaskTargetData.DynamicArgumentsScriptUrl,
		IsImua:                    jobData.IsImua,
	}

	// Create trigger data based on job type
	triggerData := s.createTriggerDataFromNotification(jobData, notification)

	// Create single task data for keeper
	sendTaskData := types.SendTaskDataToKeeper{
		TaskID:           []int64{jobData.TaskTargetData.TaskID},
		TargetData:       []types.TaskTargetData{targetData}, // Single task, not batch
		TriggerData:      []types.TaskTriggerData{triggerData},
		SchedulerID:      s.schedulerID,
		ManagerSignature: "",
	}

	// Create request for Redis API
	request := types.SchedulerTaskRequest{
		SendTaskDataToKeeper: sendTaskData,
		Source:               "condition_scheduler",
	}

	// Submit to TaskManager
	return s.submitTaskToTaskManager(request, notification.JobID)
}

// createTriggerDataFromNotification creates appropriate trigger data based on job type
func (s *ConditionBasedScheduler) createTriggerDataFromNotification(jobData *types.ScheduleConditionJobData, notification *worker.TriggerNotification) types.TaskTriggerData {
	baseTriggerData := types.TaskTriggerData{
		TaskID:                  jobData.TaskTargetData.TaskID,
		TaskDefinitionID:        jobData.TaskDefinitionID,
		CurrentTriggerTimestamp: notification.TriggeredAt,
	}

	switch jobData.TaskDefinitionID {
	case 5, 6: // Condition-based
		baseTriggerData.ExpirationTime = jobData.ConditionWorkerData.ExpirationTime
		baseTriggerData.ConditionSatisfiedValue = int(notification.TriggerValue)
		baseTriggerData.ConditionType = jobData.ConditionWorkerData.ConditionType
		baseTriggerData.ConditionSourceType = jobData.ConditionWorkerData.ValueSourceType
		baseTriggerData.ConditionSourceUrl = jobData.ConditionWorkerData.ValueSourceUrl
		baseTriggerData.ConditionUpperLimit = int(jobData.ConditionWorkerData.UpperLimit)
		baseTriggerData.ConditionLowerLimit = int(jobData.ConditionWorkerData.LowerLimit)
		s.logger.Info("Condition job expiration time", "expiration_time", jobData.ConditionWorkerData.ExpirationTime)

	case 3, 4: // Event-based
		baseTriggerData.ExpirationTime = jobData.EventWorkerData.ExpirationTime
		baseTriggerData.EventTxHash = notification.TriggerTxHash
		baseTriggerData.EventChainId = jobData.EventWorkerData.TriggerChainID
		baseTriggerData.EventTriggerContractAddress = jobData.EventWorkerData.TriggerContractAddress
		baseTriggerData.EventTriggerName = jobData.EventWorkerData.TriggerEvent
		s.logger.Info("Event job expiration time", "expiration_time", jobData.EventWorkerData.ExpirationTime)
	}

	return baseTriggerData
}

// submitTaskToTaskManager submits the task to Task Dispatcher via RPC
func (s *ConditionBasedScheduler) submitTaskToTaskManager(request types.SchedulerTaskRequest, taskID *big.Int) (bool, error) {
	startTime := time.Now()

	// Create retry configuration for task dispatcher calls
	retryConfig := &retry.RetryConfig{
		MaxRetries:      3,
		InitialDelay:    1 * time.Second,
		MaxDelay:        10 * time.Second,
		BackoffFactor:   2.0,
		JitterFactor:    0.2,
		LogRetryAttempt: true,
		ShouldRetry: func(err error) bool {
			// Retry on network errors, timeouts, and temporary failures
			// Don't retry on permanent errors like invalid requests
			return err != nil && !strings.Contains(err.Error(), "invalid") &&
				!strings.Contains(err.Error(), "permission denied")
		},
	}

	// Create context with timeout for the entire retry operation
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Define the operation to retry
	operation := func() (bool, error) {
		// Create context with timeout for individual RPC call
		rpcCtx, rpcCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer rpcCancel()

		// Make RPC call to task dispatcher
		var response types.TaskManagerAPIResponse
		err := s.taskDispatcherClient.Call(rpcCtx, "submit-task", &request, &response)
		if err != nil {
			return false, fmt.Errorf("RPC call failed: %w", err)
		}

		if !response.Success {
			return false, fmt.Errorf("task dispatcher processing failed: %s - %s", response.Message, response.Error)
		}

		return true, nil
	}

	// Execute with retry logic
	success, err := retry.Retry(ctx, operation, retryConfig, s.logger)
	if err != nil {
		duration := time.Since(startTime)
		s.logger.Error("Failed to submit task to task dispatcher after retries",
			"task_id", taskID,
			"error", err,
			"duration", duration)
		return false, err
	}

	duration := time.Since(startTime)
	s.logger.Info("Successfully submitted task to task dispatcher",
		"task_id", taskID,
		"duration", duration)

	return success, nil
}
