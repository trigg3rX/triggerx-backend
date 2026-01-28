package dispatcher

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/trigg3rX/triggerx-backend/pkg/cryptography"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// TaskDispatcher encapsulates dependencies for handling scheduler submissions
// and forwarding them to the aggregator.
type TaskDispatcher struct {
	logger            observability.Logger
	tracer            observability.Tracer
	taskStreamManager *TaskStreamManager
	performerFetcher  *PerformerFetcher
	signingKey        string
	signingAddress    string
}

// NewTaskDispatcher constructs a new dispatcher with an initialized aggregator client.
func NewTaskDispatcher(
	logger observability.Logger,
	tracer observability.Tracer,
	taskStreamManager *TaskStreamManager,
	performerFetcher *PerformerFetcher,
	signingKey string,
	signingAddress string,
) (*TaskDispatcher, error) {
	return &TaskDispatcher{
		logger:            logger,
		tracer:            tracer,
		taskStreamManager: taskStreamManager,
		performerFetcher:  performerFetcher,
		signingKey:        signingKey,
		signingAddress:    signingAddress,
	}, nil
}

// SubmitTaskFromScheduler is the core business method used by the RPC handler.
// It receives the scheduler request, optionally enqueues to Redis,
// and forwards the data to the aggregator
func (d *TaskDispatcher) SubmitTaskFromScheduler(ctx context.Context, req *types.SchedulerTaskRequest) (*types.TaskDispatcherRPCResponse, error) {
	// Trace context is automatically extracted by gRPC interceptor
	ctx, span := d.tracer.Start(ctx, "task.dispatch",
		observability.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	// Set initial span attributes
	span.SetAttributes(
		attribute.Int64("task.id", req.SendTaskDataToKeeper.TaskID),
		attribute.String("task.network", string(req.SendTaskDataToKeeper.Network)),
		attribute.String("scheduler.id", req.SendTaskDataToKeeper.SchedulerID),
		attribute.String("source", req.Source),
	)

	d.logger.Info(ctx, "Receiving task from scheduler",
		observability.Int64("task_ids", req.SendTaskDataToKeeper.TaskID),
		observability.String("scheduler_id", req.SendTaskDataToKeeper.SchedulerID),
		observability.String("source", req.Source))

	// Use dynamic performer selection
	performer, err := d.performerFetcher.FetchPerformer(ctx, req.SendTaskDataToKeeper.Network)
	if err != nil {
		span.RecordError(err, observability.WithErrorAttributes(
			attribute.String("error.type", "performer_selection_failed"),
		))
		span.SetStatus(codes.Error, "failed to get performer")
		d.logger.Error(ctx, "Failed to get performer data dynamically",
			observability.Int64("task_id", req.SendTaskDataToKeeper.TaskID),
			observability.Error(err))
		return nil, fmt.Errorf("failed to get performer: %w", err)
	}
	// Update task with performer information
	req.SendTaskDataToKeeper.PerformerData = performer

	// Add event when performer is selected
	span.SetAttributes(
		attribute.String("performer.address", performer.KeeperAddress),
	)
	span.AddEvent("performer.selected", observability.WithEventAttributes(
		attribute.String("performer.address", performer.KeeperAddress),
	))

	// Single task - generate signature normally
	signature, err := cryptography.SignJSONMessage(req.SendTaskDataToKeeper, d.signingKey)
	if err != nil {
		span.RecordError(err, observability.WithErrorAttributes(
			attribute.String("error.type", "signing_failed"),
		))
		span.SetStatus(codes.Error, "failed to sign task data")
		d.logger.Error(ctx, "Failed to sign task data",
			observability.Int64("task_id", req.SendTaskDataToKeeper.TaskID),
			observability.Error(err))
		return nil, fmt.Errorf("failed to sign task data: %w", err)
	}
	req.SendTaskDataToKeeper.ManagerSignature = signature
	
	span.AddEvent("task.signed", observability.WithEventAttributes(
		attribute.String("signature", signature),
	))
	
	if err := d.dispatchSingleTask(ctx, req, span); err != nil {
		return nil, err
	}

	// Add event when task is sent
	span.AddEvent("task.sent", observability.WithEventAttributes())

	d.logger.Info(ctx, "Task forwarded to performer", observability.Int64("task_id", req.SendTaskDataToKeeper.TaskID))

	span.SetStatus(codes.Ok, "task dispatched successfully")
	return &types.TaskDispatcherRPCResponse{
		Success:   true,
		TaskID:    req.SendTaskDataToKeeper.TaskID,
		Message:   "Task submitted successfully",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// dispatchSingleTask handles dispatching a single task request
func (d *TaskDispatcher) dispatchSingleTask(ctx context.Context, req *types.SchedulerTaskRequest, span observability.Span) error {
	taskStreamData := types.TaskStreamData{
		JobID:                req.SendTaskDataToKeeper.TargetData.JobID,
		TaskDefinitionID:     req.SendTaskDataToKeeper.TargetData.TaskDefinitionID,
		CreatedAt:            time.Now(),
		RetryCount:           0,
		SendTaskDataToKeeper: req.SendTaskDataToKeeper,
		Network:              req.SendTaskDataToKeeper.Network,
	}

	// Add task to batch processor for improved performance
	_, err := d.taskStreamManager.AddTaskToDispatchedStream(ctx, taskStreamData)
	if err != nil {
		span.RecordError(err, observability.WithErrorAttributes(
			attribute.String("error.type", "task_dispatch_failed"),
		))
		span.SetStatus(codes.Error, "failed to add task to batch processor")
		d.logger.Error(ctx, "Failed to add task to batch processor",
			observability.Int64("task_id", req.SendTaskDataToKeeper.TaskID),
			observability.String("source", req.Source),
			observability.Error(err))
		return fmt.Errorf("failed to add task to batch processor: %w", err)
	}
	return nil
}

// Close closes all resources held by the TaskDispatcher
func (d *TaskDispatcher) Close(ctx context.Context) error {
	if d.performerFetcher != nil {
		if err := d.performerFetcher.Close(ctx); err != nil {
			d.logger.Warn(ctx, "Failed to close performer fetcher", observability.Error(err))
		}
	}
	return d.taskStreamManager.Close(ctx)
}
