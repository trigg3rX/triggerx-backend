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
	traceName := fmt.Sprintf("time-%s-%d", s.schedulerID, time.Now().Unix())

	// Create root span for polling cycle BEFORE polling DB
	ctx, pollSpan := s.tracer.Start(ctx, traceName,
		observability.WithSpanKind(trace.SpanKindProducer),
		observability.WithAttributes(
			attribute.String("scheduler.id", s.schedulerID),
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

	pollSpan.SetAttributes(
		attribute.Int("poll.tasks_found", len(tasks)),
	)
	pollSpan.AddEvent("poll.completed", observability.WithEventAttributes(
		attribute.Int("task_count", len(tasks)),
	))

	if len(tasks) == 0 {
		return
	}

	s.logger.Debug(ctx, "Found tasks to dispatch", observability.Int("task_count", len(tasks)))
	metrics.UpdateTasksCreated(float64(len(tasks)))
	metrics.UpdateTaskBatchSize(float64(s.taskBatchSize))

	for _, task := range tasks {
		metrics.TrackTaskByScheduleType(task.TriggerData.TimeScheduleType)
		task.SchedulerID = s.schedulerID
		// Create request for task dispatcher
		request := types.SchedulerTaskRequest{
			SendTaskDataToKeeper: task,
			Source:               "time_scheduler",
		}

		// Submit batch to task dispatcher
		success := s.submitTaskToTaskDispatcher(ctx, request)

		if success {
			s.logger.Info(ctx, "Task dispatch completed successfully", observability.Int64("task_id", task.TaskID))
			metrics.UpdateTasksDispatched("success")
		} else {
			s.logger.Error(ctx, "Task dispatch failed", observability.Int64("task_id", task.TaskID))
			metrics.UpdateTasksDispatched("failed")
		}
	}
}
