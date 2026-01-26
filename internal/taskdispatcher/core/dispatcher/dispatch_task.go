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
// It receives the scheduler request, optionally enqueues to Redis (skipped for now),
// and forwards the data to the aggregator using the same approach as send_task.go.
func (d *TaskDispatcher) SubmitTaskFromScheduler(ctx context.Context, req *types.SchedulerTaskRequest) (*types.TaskDispatcherRPCResponse, error) {
	// Trace context is automatically extracted by gRPC interceptor
	ctx, span := d.tracer.Start(ctx, "task.dispatch",
		observability.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	if err := d.validateRequest(req, span); err != nil {
		return nil, err
	}

	taskCount := len(req.SendTaskDataToKeeper.TaskID)

	// Set initial span attributes
	span.SetAttributes(
		attribute.Int64("task.id", req.SendTaskDataToKeeper.TaskID[0]),
		attribute.String("task.network", string(req.SendTaskDataToKeeper.Network)),
		attribute.Int("task.count", taskCount),
		attribute.String("scheduler.id", req.SendTaskDataToKeeper.SchedulerID),
		attribute.String("source", req.Source),
	)

	d.logger.Info(ctx, "Receiving task from scheduler",
		observability.Int64("task_ids", req.SendTaskDataToKeeper.TaskID[0]),
		observability.Int("task_count", taskCount),
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
			observability.Int64("task_id", req.SendTaskDataToKeeper.TaskID[0]),
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

	// Sign the task data
	signature, err := cryptography.SignJSONMessage(req.SendTaskDataToKeeper, d.signingKey)
	if err != nil {
		span.RecordError(err, observability.WithErrorAttributes(
			attribute.String("error.type", "signing_failed"),
		))
		span.SetStatus(codes.Error, "failed to sign task data")
		d.logger.Error(ctx, "Failed to sign batch task data",
			observability.Int64("task_id", req.SendTaskDataToKeeper.TaskID[0]),
			observability.Error(err))
		return nil, fmt.Errorf("failed to sign task data: %w", err)
	}
	req.SendTaskDataToKeeper.ManagerSignature = signature

	// Add event when task is signed
	span.AddEvent("task.signed", observability.WithEventAttributes(
		attribute.String("signature", signature),
	))

	// Handle batch requests by creating individual task stream data for each task
	if taskCount > 1 {
		if err := d.dispatchBatchTasks(ctx, req, span, taskCount); err != nil {
			// Partial failures are logged but don't fail the entire request
			d.logger.Warn(ctx, "Some batch tasks failed", observability.Error(err))
		}
	} else {
		if err := d.dispatchSingleTask(ctx, req, span); err != nil {
			return nil, err
		}
	}

	// Add event when task is sent
	span.AddEvent("task.sent", observability.WithEventAttributes(
		attribute.Int("task_count", taskCount),
	))

	d.logger.Info(ctx, "Task forwarded to performer", observability.Int64("task_id", req.SendTaskDataToKeeper.TaskID[0]))

	span.SetStatus(codes.Ok, "task dispatched successfully")
	return &types.TaskDispatcherRPCResponse{
		Success:   true,
		TaskID:    []int64{req.SendTaskDataToKeeper.TaskID[0]},
		Message:   "Task submitted successfully",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// validateRequest validates the incoming scheduler task request
func (d *TaskDispatcher) validateRequest(req *types.SchedulerTaskRequest, span observability.Span) error {
	if req == nil {
		span.RecordError(fmt.Errorf("nil request"), observability.WithErrorAttributes(
			attribute.String("error.type", "invalid_request"),
		))
		span.SetStatus(codes.Error, "nil request")
		return fmt.Errorf("nil request")
	}
	if len(req.SendTaskDataToKeeper.TaskID) == 0 {
		span.RecordError(fmt.Errorf("missing task id"), observability.WithErrorAttributes(
			attribute.String("error.type", "invalid_request"),
		))
		span.SetStatus(codes.Error, "missing task id")
		return fmt.Errorf("missing task id")
	}
	if len(req.SendTaskDataToKeeper.TriggerData) == 0 {
		span.RecordError(fmt.Errorf("missing trigger data"), observability.WithErrorAttributes(
			attribute.String("error.type", "invalid_request"),
		))
		span.SetStatus(codes.Error, "missing trigger data")
		return fmt.Errorf("missing trigger data")
	}
	if len(req.SendTaskDataToKeeper.TargetData) == 0 {
		span.RecordError(fmt.Errorf("missing target data"), observability.WithErrorAttributes(
			attribute.String("error.type", "invalid_request"),
		))
		span.SetStatus(codes.Error, "missing target data")
		return fmt.Errorf("missing target data")
	}
	return nil
}

// dispatchBatchTasks handles dispatching multiple tasks in a batch request
func (d *TaskDispatcher) dispatchBatchTasks(ctx context.Context, req *types.SchedulerTaskRequest, span observability.Span, taskCount int) error {
	d.logger.Debug(ctx, "Processing batch request", observability.Int("task_count", taskCount))
	span.SetAttributes(attribute.Bool("batch.request", true))

	var batchErrors []error
	for i := 0; i < taskCount; i++ {
		// Create individual task data for each task in the batch
		individualTaskData := types.SendTaskDataToKeeper{
			TaskID:           []int64{req.SendTaskDataToKeeper.TaskID[i]},
			PerformerData:    req.SendTaskDataToKeeper.PerformerData,
			TargetData:       []types.TaskTargetData{req.SendTaskDataToKeeper.TargetData[i]},
			TriggerData:      []types.TaskTriggerData{req.SendTaskDataToKeeper.TriggerData[i]},
			SchedulerID:      req.SendTaskDataToKeeper.SchedulerID,
			ManagerSignature: req.SendTaskDataToKeeper.ManagerSignature,
			Network:          req.SendTaskDataToKeeper.Network,
		}

		taskStreamData := types.TaskStreamData{
			JobID:                individualTaskData.TargetData[0].JobID,
			TaskDefinitionID:     individualTaskData.TargetData[0].TaskDefinitionID,
			CreatedAt:            time.Now(),
			RetryCount:           0,
			SendTaskDataToKeeper: individualTaskData,
			Network:              individualTaskData.Network,
		}

		// Add individual task to batch processor
		_, err := d.taskStreamManager.AddTaskToDispatchedStream(ctx, taskStreamData)
		if err != nil {
			batchErrors = append(batchErrors, err)
			span.RecordError(err, observability.WithErrorAttributes(
				attribute.String("error.type", "batch_task_failed"),
				attribute.Int("batch_index", i),
				attribute.Int64("task.id", individualTaskData.TaskID[0]),
			))
			d.logger.Error(ctx, "Failed to add individual task to batch processor",
				observability.Int64("task_id", individualTaskData.TaskID[0]),
				observability.Int("batch_index", i),
				observability.String("source", req.Source),
				observability.Error(err))
			// Continue processing other tasks in the batch
			continue
		}

		d.logger.Debug(ctx, "Individual task added to batch processor",
			observability.Int64("task_id", individualTaskData.TaskID[0]),
			observability.Int("batch_index", i))
	}

	// Record batch processing results
	if len(batchErrors) > 0 {
		span.SetAttributes(
			attribute.Int("batch.errors", len(batchErrors)),
			attribute.Int("batch.success", taskCount-len(batchErrors)),
		)
		return fmt.Errorf("%d tasks failed in batch", len(batchErrors))
	}
	return nil
}

// dispatchSingleTask handles dispatching a single task request
func (d *TaskDispatcher) dispatchSingleTask(ctx context.Context, req *types.SchedulerTaskRequest, span observability.Span) error {
	taskStreamData := types.TaskStreamData{
		JobID:                req.SendTaskDataToKeeper.TargetData[0].JobID,
		TaskDefinitionID:     req.SendTaskDataToKeeper.TargetData[0].TaskDefinitionID,
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
			observability.Int64("task_id", req.SendTaskDataToKeeper.TaskID[0]),
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
