package scheduler

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// pollAndScheduleTasks fetches tasks from database and schedules them for execution.
// It polls for time-based jobs and custom jobs within the look-ahead window,
// creates task records, and processes them in batches.
func (s *TimeBasedScheduler) pollAndScheduleTasks(ctx context.Context) {
	// Create trace with format "time-{scheduler_id}-{timestamp}"
	// This trace will be propagated through task creation and gRPC calls
	traceName := fmt.Sprintf("time-%d-%d", s.schedulerID, time.Now().Unix())

	// Create root span for polling cycle BEFORE polling DB
	ctx, pollSpan := s.tracer.Start(ctx, traceName,
		observability.WithSpanKind(trace.SpanKindProducer),
		observability.WithAttributes(
			attribute.Int("scheduler.id", s.schedulerID),
			attribute.String("scheduler.type", "time"),
			attribute.String("poll.look_ahead", s.pollingLookAhead.String()),
			attribute.String("trace.name", traceName),
		),
	)
	defer pollSpan.End()

	pollSpan.AddEvent("poll.started")

	// Calculate look ahead time
	lookAheadTime := time.Now().Add(s.pollingLookAhead)

	// Poll time-based jobs from database
	tasks, err := s.timeJobRepository.GetTimeJobsByNextExecutionTimestamp(lookAheadTime)
	if err != nil {
		pollSpan.RecordError(err, observability.WithErrorAttributes(
			attribute.String("error.type", "db_connection_error"),
		))
		pollSpan.SetStatus(codes.Error, "failed to fetch tasks")
		s.logger.Error(ctx, "Failed to fetch time-based tasks", observability.Error(err))
		metrics.TrackDBRequestError("read", "time_jobs")
		return
	}

	// Get custom jobs (TaskDefinitionID = 7) if repository is available
	if s.customJobRepository != nil {
		customJobs, err := s.customJobRepository.GetCustomJobsDueForExecution(lookAheadTime)
		if err != nil {
			s.logger.Warn(ctx, "Error retrieving custom jobs", observability.Error(err))
			// Don't fail, just log and continue with time jobs only
		} else {
			// Convert custom jobs to ScheduleTimeTaskData format
			for _, customJob := range customJobs {
				taskData := s.convertCustomJobToScheduleTimeTaskData(ctx, &customJob)
				tasks = append(tasks, taskData)
			}
		}
	}

	// Create task data for each task and add task IDs to jobs
	for i := range tasks {
		taskID, err := s.taskRepository.CreateTaskDataInDB(ctx, &types.CreateTaskDataRequest{
			JobID:            tasks[i].TaskTargetData.JobID.Int,
			TaskDefinitionID: tasks[i].TaskDefinitionID,
			IsImua:           tasks[i].IsImua,
		})
		if err != nil {
			s.logger.Error(ctx, "Error creating task data", observability.Error(err))
			continue
		}

		err = s.taskRepository.AddTaskIDToJob(tasks[i].TaskTargetData.JobID.Int, taskID)
		if err != nil {
			s.logger.Error(ctx, "Error adding task ID to job", observability.Error(err))
			continue
		}

		tasks[i].TaskID = taskID
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

	s.logger.Debug(ctx, "Found tasks to process", observability.Int("task_count", len(tasks)))
	metrics.UpdateTasksCreated(float64(len(tasks)))
	metrics.UpdateTaskBatchSize(float64(s.taskBatchSize))

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
		s.logger.Debug(ctx, "Processing non-imua tasks in batches", observability.Int("task_count", len(nonImuaTasks)))
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
		s.logger.Debug(ctx, "Processing imua tasks in separate batches", observability.Int("task_count", len(imuaTasks)))
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
