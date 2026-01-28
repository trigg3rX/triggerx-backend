package dispatcher

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/ethereum/go-ethereum/common"
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
	signatureDeadlineBuffer int64 // Deadline buffer in seconds for contract signatures
}

// NewTaskDispatcher constructs a new dispatcher with an initialized aggregator client.
func NewTaskDispatcher(
	logger observability.Logger,
	tracer observability.Tracer,
	taskStreamManager *TaskStreamManager,
	performerFetcher *PerformerFetcher,
	signingKey string,
	signingAddress string,
	signatureDeadlineBuffer int64,
) (*TaskDispatcher, error) {
	return &TaskDispatcher{
		logger:            logger,
		tracer:            tracer,
		taskStreamManager: taskStreamManager,
		performerFetcher:  performerFetcher,
		signingKey:        signingKey,
		signingAddress:    signingAddress,
		signatureDeadlineBuffer: signatureDeadlineBuffer,
	}, nil
}

// signContractExecution signs execution parameters for contract verification.
// This creates a signature that the TaskExecutionHub contract can verify,
// authorizing a specific keeper to execute a specific job.
func (d *TaskDispatcher) signContractExecution(
	jobId *big.Int,
	target common.Address,
	data []byte,
	deadline *big.Int,
	keeperAddress common.Address,
	chainId *big.Int,
) ([]byte, error) {
	return cryptography.SignContractExecution(
		d.signingKey,
		jobId,
		target,
		data,
		deadline,
		keeperAddress,
		chainId,
	)
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

	// Generate contract signatures for target data item
	// The contract signature authorizes the specific keeper to execute on the contract
	deadline := big.NewInt(time.Now().Unix() + d.signatureDeadlineBuffer)
	keeperAddress := common.HexToAddress(performer.KeeperAddress)

	// Set deadline (same for all targets in this batch)
	req.SendTaskDataToKeeper.TargetData.Deadline = deadline.String()

	// Parse target chain ID for contract signing
	chainIdInt, err := strconv.ParseInt(req.SendTaskDataToKeeper.TargetData.TargetChainID, 10, 64)
	if err != nil {
		d.logger.Warn(ctx, "Failed to parse chain ID for contract signing, skipping contract signature",
			observability.Int64("task_id", req.SendTaskDataToKeeper.TargetData.TaskID),
			observability.String("chain_id", req.SendTaskDataToKeeper.TargetData.TargetChainID),
			observability.Error(err))
	}
	chainId := big.NewInt(chainIdInt)
	targetAddress := common.HexToAddress(req.SendTaskDataToKeeper.TargetData.TargetContractAddress)

	// For contract signature, we need the calldata - but at this stage we don't have it yet
	// The calldata is generated by the keeper. So we sign with empty data placeholder
	// and let the keeper populate the actual calldata before execution.
	// NOTE: In future, if we need full calldata signing, we'll need to pre-compute it here.
	// For now, we sign the job parameters which provide sufficient security verification.
	jobId := types.ConvertToBigInt(req.SendTaskDataToKeeper.TargetData.JobID)

	// Create contract signature with empty data - the contract will verify using keccak256(data)
	// where data is the actual calldata at execution time
	contractSig, err := d.signContractExecution(
		jobId,
		targetAddress,
		[]byte{}, // Empty data - will be verified with actual calldata at execution
		deadline,
		keeperAddress,
		chainId,
	)
	if err != nil {
		d.logger.Warn(ctx, "Failed to generate contract signature",
			observability.Int64("task_id", req.SendTaskDataToKeeper.TargetData.TaskID),
			observability.Error(err))
		return nil, fmt.Errorf("failed to generate contract signature: %w", err)
	}
	req.SendTaskDataToKeeper.TargetData.ContractSignature = contractSig

	d.logger.Debug(ctx, "Generated contract signature for task",
		observability.Int64("task_id", req.SendTaskDataToKeeper.TargetData.TaskID),
		observability.Int64("deadline", deadline.Int64()),
		observability.String("keeper", performer.KeeperAddress))

	span.AddEvent("contract.signatures.generated", observability.WithEventAttributes(
		attribute.Int64("deadline", deadline.Int64()),
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
