package scheduler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// pollAndScheduleTasks fetches tasks from database and schedules them for execution
func (s *TimeBasedScheduler) pollAndScheduleTasks(ctx context.Context) {
	// Create root span for polling cycle BEFORE polling DB
	ctx, pollSpan := s.tracer.Start(ctx, "task.poll",
		observability.WithSpanKind(trace.SpanKindProducer),
		observability.WithAttributes(
			attribute.Int("scheduler.id", s.schedulerID),
			attribute.String("scheduler.type", "time"),
			attribute.String("poll.look_ahead", s.pollingLookAhead.String()),
		),
	)
	defer pollSpan.End()

	pollSpan.AddEvent("poll.started")

	// NOW poll the DB server
	tasks, err := s.dbClient.GetTimeBasedTasks(ctx)
	if err != nil {
		pollSpan.RecordError(err, observability.WithErrorAttributes(
			attribute.String("error.type", "db_connection_error"),
		))
		pollSpan.SetStatus(codes.Error, "failed to fetch tasks")
		s.logger.Error(ctx, "Failed to fetch time-based tasks", observability.Error(err))
		metrics.TrackDBConnectionError()
		return
	}

	pollSpan.SetAttributes(
		attribute.Int("poll.tasks_found", len(tasks)),
	)
	pollSpan.AddEvent("poll.completed", observability.WithEventAttributes(
		attribute.Int("task_count", len(tasks)),
	))

	if len(tasks) == 0 {
		return
	}

	s.logger.Info(ctx, "Found %d tasks to process", observability.Int("task_count", len(tasks)))
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
		s.logger.Info(ctx, "Processing %d non-imua tasks in batches", observability.Int("task_count", len(nonImuaTasks)))
		for i := 0; i < len(nonImuaTasks); i += s.taskBatchSize {
			end := i + s.taskBatchSize
			if end > len(nonImuaTasks) {
				end = len(nonImuaTasks)
			}

			batch := nonImuaTasks[i:end]
			s.processBatch(ctx, batch)
		}
	}

	// Process imua tasks in separate batches
	if len(imuaTasks) > 0 {
		s.logger.Info(ctx, "Processing %d imua tasks in separate batches", observability.Int("task_count", len(imuaTasks)))
		for i := 0; i < len(imuaTasks); i += s.taskBatchSize {
			end := i + s.taskBatchSize
			if end > len(imuaTasks) {
				end = len(imuaTasks)
			}

			batch := imuaTasks[i:end]
			s.processBatch(ctx, batch)
		}
	}
}

// processBatch processes a batch of tasks by submitting to task dispatcher
func (s *TimeBasedScheduler) processBatch(ctx context.Context, tasks []types.ScheduleTimeTaskData) {
	s.logger.Info(ctx, "Processing batch of %d time-based tasks", observability.Int("task_count", len(tasks)))

	var targetDataList []types.TaskTargetData
	var triggerDataList []types.TaskTriggerData
	var validTaskIDs []int64

	for _, task := range tasks {
		// Check if ExpirationTime of the job has passed or not
		if task.ExpirationTime.Before(time.Now()) {
			s.logger.Info(ctx, "Task ID %d has expired, skipping execution", observability.Int64("task_id", task.TaskID))
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
		s.logger.Info(ctx, "Batch processing completed successfully: %d tasks submitted", observability.Int("task_count", len(validTaskIDs)))
		metrics.TrackTaskCompletion(true, time.Since(time.Now()))
		metrics.TrackTaskBroadcast("task_dispatcher_submitted")
	} else {
		s.logger.Error(ctx, "Batch processing failed: %d tasks", observability.Int("task_count", len(validTaskIDs)))
		metrics.TrackTaskBroadcast("failed")
	}
}

// submitBatchToTaskDispatcher submits the batch task data to Task Dispatcher via RPC
func (s *TimeBasedScheduler) submitBatchToTaskDispatcher(ctx context.Context, request types.SchedulerTaskRequest, taskIDs string, taskCount int) bool {
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
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// Define the operation to retry
	operation := func() (bool, error) {
		// Create context with timeout for individual RPC call
		rpcCtx, rpcCancel := context.WithTimeout(ctx, 30*time.Second)
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
		s.logger.Error(ctx, "Failed to submit batch to task dispatcher after retries",
			observability.String("task_ids", taskIDs),
			observability.Int("task_count", taskCount),
			observability.Error(err),
			observability.Duration("duration", duration))
		return false
	}

	duration := time.Since(startTime)
	s.logger.Info(ctx, "Successfully submitted batch to task dispatcher",
		observability.String("task_ids", taskIDs),
		observability.Int("task_count", taskCount),
		observability.Duration("duration", duration))

	return success
}
