package scheduler

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	dbserverTypes "github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/scheduler/worker"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// HandleTriggerNotification processes trigger notifications and submits individual tasks to Redis API
// This is called by workers and the Event Monitor Service webhook handler
func (s *ConditionBasedScheduler) HandleTriggerNotification(ctx context.Context, notification *worker.TriggerNotification) error {
	startTime := time.Now()

	// Continue trace from worker (ctx already contains trace context)
	ctx, scheduleSpan := s.tracer.Start(ctx, "task.schedule.condition",
		observability.WithSpanKind(trace.SpanKindConsumer),
		observability.WithAttributes(
			attribute.String("job.id", notification.JobID.String()),
			attribute.String("trigger.type", "condition_or_event"),
		),
	)
	defer scheduleSpan.End()

	// Add trigger-specific attributes
	if notification.TriggerValue != 0 {
		scheduleSpan.SetAttributes(
			attribute.Float64("trigger.value", notification.TriggerValue),
		)
	}
	if notification.TriggerTxHash != "" {
		scheduleSpan.SetAttributes(
			attribute.String("trigger.tx_hash", notification.TriggerTxHash),
		)
	}

	scheduleSpan.AddEvent("notification.received")

	s.logger.Info(ctx, "Processing trigger notification - submitting task to task dispatcher",
		observability.String("job_id", notification.JobID.String()),
		observability.Float64("trigger_value", notification.TriggerValue),
		observability.String("trigger_tx_hash", notification.TriggerTxHash),
		observability.Time("triggered_at", notification.TriggeredAt),
	)

	// Acquire notification mutex to prevent cleanup during processing
	s.notificationMutex.Lock()
	defer s.notificationMutex.Unlock()

	// Get the job data from storage
	s.workersMutex.RLock()
	jobData, exists := s.jobDataStore[notification.JobID.String()]
	s.workersMutex.RUnlock()

	if !exists || jobData == nil {
		s.logger.Error(ctx, "Job data not found", observability.String("job_id", notification.JobID.String()))
		return fmt.Errorf("job data not found for job %d", notification.JobID)
	}

	createTaskRequest := dbserverTypes.CreateTaskDataRequest{
		JobID:            jobData.JobID.ToBigInt(),
		TaskDefinitionID: jobData.TaskDefinitionID,
	}

	// Create Task in Database
	taskID, err := s.dbClient.CreateTask(ctx, createTaskRequest)
	if err != nil {
		s.logger.Error(ctx, "Failed to create task in database", observability.String("job_id", notification.JobID.String()), observability.Error(err))
		return fmt.Errorf("failed to create task in database: %w", err)
	}
	jobData.TaskTargetData.TaskID = taskID

	// Create individual task and submit to task dispatcher
	success, err := s.submitTriggeredTaskToTaskDispatcher(ctx, jobData, notification)
	if err != nil {
		s.logger.Error(ctx, "Failed to submit triggered task to task dispatcher",
			observability.String("job_id", notification.JobID.String()),
			observability.Error(err),
		)
		metrics.TrackCriticalError("task_dispatcher_submission_failed")
		return err
	}

	duration := time.Since(startTime)
	if success {
		s.logger.Info(ctx, "Successfully submitted triggered task to task dispatcher",
			observability.String("job_id", notification.JobID.String()),
			observability.Duration("duration", duration),
		)
		metrics.TrackActionExecution(duration)
	} else {
		s.logger.Error(ctx, "Failed to submit triggered task to task dispatcher",
			observability.String("job_id", notification.JobID.String()),
			observability.Duration("duration", duration),
		)
		metrics.TrackCriticalError("task_dispatcher_submission_failed")
	}

	return nil
}

// submitTriggeredTaskToTaskManager creates and submits a single task to TaskManager when triggers occur
func (s *ConditionBasedScheduler) submitTriggeredTaskToTaskDispatcher(ctx context.Context, jobData *types.ScheduleConditionJobData, notification *worker.TriggerNotification) (bool, error) {
	s.logger.Info(ctx, "Creating triggered task for task dispatcher submission",
		observability.String("job_id", jobData.JobID.String()),
		observability.Int("task_definition_id", jobData.TaskDefinitionID),
		observability.Float64("trigger_value", notification.TriggerValue),
	)

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
	triggerData := s.createTriggerDataFromNotification(ctx, jobData, notification)

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
func (s *ConditionBasedScheduler) createTriggerDataFromNotification(ctx context.Context, jobData *types.ScheduleConditionJobData, notification *worker.TriggerNotification) types.TaskTriggerData {
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
		s.logger.Info(ctx, "Condition job expiration time", observability.Time("expiration_time", jobData.ConditionWorkerData.ExpirationTime))

	case 3, 4: // Event-based
		baseTriggerData.ExpirationTime = jobData.EventWorkerData.ExpirationTime
		baseTriggerData.EventTxHash = notification.TriggerTxHash
		baseTriggerData.EventChainId = jobData.EventWorkerData.TriggerChainID
		baseTriggerData.EventTriggerContractAddress = jobData.EventWorkerData.TriggerContractAddress
		baseTriggerData.EventTriggerName = jobData.EventWorkerData.TriggerEvent
		s.logger.Info(ctx, "Event job expiration time", observability.Time("expiration_time", jobData.EventWorkerData.ExpirationTime))
	}

	return baseTriggerData
}

// submitTaskToTaskManager submits the task to Task Dispatcher via RPC
func (s *ConditionBasedScheduler) submitTaskToTaskManager(request types.SchedulerTaskRequest, taskID *big.Int) (bool, error) {
	startTime := time.Now()

	// Create retry configuration for task dispatcher calls
	retryConfig := &retry.RetryConfig{
		MaxRetries:    3,
		InitialDelay:  1 * time.Second,
		MaxDelay:      10 * time.Second,
		BackoffFactor: 2.0,
		JitterFactor:  0.2,
		ShouldRetry: func(err error, attempt int) bool {
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
	success, err := retry.Retry(ctx, operation, retryConfig)
	if err != nil {
		duration := time.Since(startTime)
		s.logger.Error(ctx, "Failed to submit task to task dispatcher after retries",
			observability.String("task_id", taskID.String()),
			observability.Error(err),
			observability.Duration("duration", duration),
		)
		return false, err
	}

	duration := time.Since(startTime)
	s.logger.Info(ctx, "Successfully submitted task to task dispatcher",
		observability.String("task_id", taskID.String()),
		observability.Duration("duration", duration),
	)

	return success, nil
}
