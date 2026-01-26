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

	if len(tasks[0].TaskID) == 0 && len(tasks[1].TaskID) == 0 && len(tasks[2].TaskID) == 0 {
		return
	}

	s.logger.Debug(ctx, "Found tasks to process", observability.Int("task_count", (len(tasks[0].TaskID) + len(tasks[1].TaskID) + len(tasks[2].TaskID))))
	metrics.UpdateTasksCreated(float64(len(tasks[0].TaskID) + len(tasks[1].TaskID) + len(tasks[2].TaskID)))
	metrics.UpdateTaskBatchSize(float64(s.taskBatchSize))

	// Process mainnet tasks in batches
	if len(tasks[0].TaskID) > 0 {
		s.logger.Debug(ctx, "Processing mainnet tasks in batches", observability.Int("task_count", len(tasks[0].TaskID)))
		for i := 0; i < len(tasks[0].TaskID); i += s.taskBatchSize {
			end := i + s.taskBatchSize
			if end > len(tasks[0].TaskID) {
				end = len(tasks[0].TaskID)
			}

			batch := types.SendTaskDataToKeeper{
				TaskID:           tasks[0].TaskID[i:end],
				TargetData:       tasks[0].TargetData[i:end],
				TriggerData:      tasks[0].TriggerData[i:end],
				SchedulerID:      s.schedulerID,
				ManagerSignature: "",
				Network:          types.NetworkMainnet,
			}
			s.processBatch(ctx, batch)
		}
	}

	// Process sepolia tasks in batches
	if len(tasks) > 0 {
		s.logger.Debug(ctx, "Processing sepolia tasks in batches", observability.Int("task_count", len(tasks[1].TaskID)))
		for i := 0; i < len(tasks[1].TaskID); i += s.taskBatchSize {
			end := i + s.taskBatchSize
			if end > len(tasks[1].TaskID) {
				end = len(tasks[1].TaskID)
			}

			batch := types.SendTaskDataToKeeper{
				TaskID:           tasks[1].TaskID[i:end],
				TargetData:       tasks[1].TargetData[i:end],
				TriggerData:      tasks[1].TriggerData[i:end],
				SchedulerID:      s.schedulerID,
				ManagerSignature: "",
				Network:          types.NetworkSepolia,
			}
			s.processBatch(ctx, batch)
		}
	}

	// Process imua tasks in separate batches
	if len(tasks[2].TaskID) > 0 {
		s.logger.Debug(ctx, "Processing imua tasks in separate batches", observability.Int("task_count", len(tasks[2].TaskID)))
		for i := 0; i < len(tasks[2].TaskID); i += s.taskBatchSize {
			end := i + s.taskBatchSize
			if end > len(tasks[2].TaskID) {
				end = len(tasks[2].TaskID)
			}

			batch := types.SendTaskDataToKeeper{
				TaskID:           tasks[2].TaskID[i:end],
				TargetData:       tasks[2].TargetData[i:end],
				TriggerData:      tasks[2].TriggerData[i:end],
				SchedulerID:      s.schedulerID,
				ManagerSignature: "",
				Network:          types.NetworkImua,
			}
			s.processBatch(ctx, batch)
		}
	}
}
