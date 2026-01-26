package handlers

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// notifyConditionScheduler sends a notification to the condition scheduler via gRPC
func (h *Handler) notifyConditionScheduler(ctx context.Context, scheduleConditionJobRequest types.ScheduleConditionJobRequest) (bool, error) {
	ctx, span := h.tracer.Start(ctx, "rpc.schedule_condition_job",
		observability.WithSpanKind(trace.SpanKindClient),
		observability.WithAttributes(
			attribute.String("rpc.service", "condition-scheduler"),
			attribute.String("rpc.method", "schedule-job"),
			attribute.String("job.id", scheduleConditionJobRequest.JobID),
			attribute.Int("task_definition_id", scheduleConditionJobRequest.TaskDefinitionID),
		),
	)
	defer span.End()

	if err := h.conditionSchedulerClient.ScheduleJob(ctx, &scheduleConditionJobRequest); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to schedule job")
		h.logger.Error(ctx, "[NotifyConditionScheduler] Failed to notify condition scheduler for job", observability.String("job_id", scheduleConditionJobRequest.JobID), observability.Error(err))
		return false, err
	}

	span.SetStatus(codes.Ok, "")
	h.logger.Info(ctx, "Successfully sent data to condition scheduler via gRPC", observability.String("job_id", scheduleConditionJobRequest.JobID))
	return true, nil
}

// notifyPauseToConditionScheduler sends an unschedule notification to the condition scheduler via gRPC
func (h *Handler) notifyPauseToConditionScheduler(ctx context.Context, scheduleConditionJobRequest types.ScheduleConditionJobRequest) (bool, error) {
	ctx, span := h.tracer.Start(ctx, "rpc.unschedule_condition_job",
		observability.WithSpanKind(trace.SpanKindClient),
		observability.WithAttributes(
			attribute.String("rpc.service", "condition-scheduler"),
			attribute.String("rpc.method", "unschedule-job"),
			attribute.String("job.id", scheduleConditionJobRequest.JobID),
			attribute.Int("task_definition_id", scheduleConditionJobRequest.TaskDefinitionID),
		),
	)
	defer span.End()

	if err := h.conditionSchedulerClient.UnscheduleJob(ctx, &scheduleConditionJobRequest); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to unschedule job")
		h.logger.Error(ctx, "[NotifyConditionScheduler] Failed to unschedule job in condition scheduler", observability.String("job_id", scheduleConditionJobRequest.JobID), observability.Error(err))
		return false, err
	}

	span.SetStatus(codes.Ok, "")
	h.logger.Info(ctx, "Successfully unscheduled job in condition scheduler via gRPC", observability.String("job_id", scheduleConditionJobRequest.JobID))
	return true, nil
}
