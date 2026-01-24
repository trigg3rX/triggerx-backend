package execution

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	// "strconv"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
	noncemanager "github.com/trigg3rX/triggerx-backend/internal/keeper/core/nonce_manager"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/core/validation"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/utils"
	"github.com/trigg3rX/triggerx-backend/pkg/client/aggregator"
	"github.com/trigg3rX/triggerx-backend/pkg/cryptography"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/proof"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// TaskMonitorClientInterface defines the interface for taskmonitor client operations
type TaskMonitorClientInterface interface {
	// ReportTaskStatus reports task execution status to taskmonitor (both success and failure)
	ReportTaskStatus(ctx context.Context, taskID int64, executionSuccessful, aggregatorSubmitted bool, executionTxHash, proofCID, errorMsg string) error
}

// TaskExecutor is the default implementation of TaskExecutor
type TaskExecutor struct {
	alchemyAPIKey     string
	argConverter      *ArgumentConverter
	validator         *validation.TaskValidator
	aggregatorClient  *aggregator.AggregatorClient
	taskMonitorClient TaskMonitorClientInterface
	logger            observability.Logger
	tracer            observability.Tracer
	nonceManagers     map[string]*noncemanager.NonceManager // Chain ID -> NonceManager
	nonceMutex        sync.RWMutex
	// Broadcast data storage for rebroadcast capability
	broadcastData      map[int64]*types.BroadcastDataForValidators // Task ID -> BroadcastData
	broadcastDataMutex sync.RWMutex
}

// NewTaskExecutor creates a new instance of TaskExecutor
func NewTaskExecutor(
	alchemyAPIKey string,
	validator *validation.TaskValidator,
	aggregatorClient *aggregator.AggregatorClient,
	taskMonitorClient TaskMonitorClientInterface,
	logger observability.Logger,
	tracer observability.Tracer) *TaskExecutor {
	return &TaskExecutor{
		alchemyAPIKey:     alchemyAPIKey,
		argConverter:      &ArgumentConverter{},
		validator:         validator,
		aggregatorClient:  aggregatorClient,
		taskMonitorClient: taskMonitorClient,
		logger:            logger,
		tracer:            tracer,
		nonceManagers:     make(map[string]*noncemanager.NonceManager),
		broadcastData:     make(map[int64]*types.BroadcastDataForValidators),
	}
}

func (e *TaskExecutor) ExecuteTask(ctx context.Context, task *types.SendTaskDataToKeeper, traceID string) (bool, error) {
	// Extract trace context from incoming request (already extracted from HTTP headers in handler)
	ctx, span := e.tracer.Start(ctx, "task.execute",
		observability.WithSpanKind(trace.SpanKindConsumer),
	)
	defer span.End()

	// Check for nil task
	if task == nil {
		span.RecordError(fmt.Errorf("task data is nil"), observability.WithErrorAttributes(
			attribute.String("error.type", "invalid_request"),
		))
		span.SetStatus(codes.Error, "task data cannot be nil")
		e.logger.Error(ctx, "Task data is nil", observability.String("trace_id", traceID))
		return false, fmt.Errorf("task data cannot be nil")
	}

	// Check for nil TargetData and TriggerData
	if task.TargetData == nil {
		span.RecordError(fmt.Errorf("target data is nil"), observability.WithErrorAttributes(
			attribute.String("error.type", "invalid_request"),
		))
		span.SetStatus(codes.Error, "target data cannot be nil")
		e.logger.Error(ctx, "TargetData is nil", observability.Int64("task_id", task.TaskID[0]), observability.String("trace_id", traceID))
		return false, fmt.Errorf("target data cannot be nil")
	}
	if task.TriggerData == nil {
		span.RecordError(fmt.Errorf("trigger data is nil"), observability.WithErrorAttributes(
			attribute.String("error.type", "invalid_request"),
		))
		span.SetStatus(codes.Error, "trigger data cannot be nil")
		e.logger.Error(ctx, "TriggerData is nil", observability.Int64("task_id", task.TaskID[0]), observability.String("trace_id", traceID))
		return false, fmt.Errorf("trigger data cannot be nil")
	}

	// Set initial span attributes
	span.SetAttributes(
		attribute.Int64("task.id", task.TaskID[0]),
		attribute.String("target.chain_id", task.TargetData[0].TargetChainID),
		attribute.String("target.contract_address", task.TargetData[0].TargetContractAddress),
		attribute.String("target.function", task.TargetData[0].TargetFunction),
		attribute.String("task.network", string(task.Network)),
		attribute.String("keeper.network", config.GetNetwork()),
	)

	// Network matching: Check if the task's network matches the keeper's configured network
	keeperNetwork := config.GetNetwork()
	if string(task.Network) == "" {
		span.RecordError(fmt.Errorf("task network is empty"), observability.WithErrorAttributes(
			attribute.String("error.type", "network_missing"),
		))
		span.SetStatus(codes.Error, "task network is required")
		e.logger.Info(ctx, "Task rejected: network field is required but missing",
			observability.Int64("task_id", task.TaskID[0]),
			observability.String("keeper_network", keeperNetwork),
			observability.String("trace_id", traceID))
		return false, fmt.Errorf("task network is required but missing (keeper network: %s)", keeperNetwork)
	}
	if string(task.Network) != string(keeperNetwork) {
		// Networks don't match - reject the task
		span.RecordError(fmt.Errorf("network mismatch: task network %s does not match keeper network %s", string(task.Network), string(keeperNetwork)), observability.WithErrorAttributes(
			attribute.String("error.type", "network_mismatch"),
		))
		span.SetStatus(codes.Error, "network mismatch")
		e.logger.Info(ctx, "Task rejected due to network mismatch",
			observability.Int64("task_id", task.TaskID[0]),
			observability.String("task_network", string(task.Network)),
			observability.String("keeper_network", string(keeperNetwork)),
			observability.String("trace_id", traceID))
		return false, fmt.Errorf("network mismatch: task network %s does not match keeper network %s", string(task.Network), string(keeperNetwork))
	}

	e.logger.Info(ctx, "[0/7] Network validation passed", observability.Int64("task_id", task.TaskID[0]), observability.String("network", string(task.Network)), observability.String("trace_id", traceID))

	// check if the scheduler signature is valid
	isManagerSignatureTrue, err := e.validator.ValidateManagerSignature(ctx, task, traceID)
	if !isManagerSignatureTrue {
		span.RecordError(err, observability.WithErrorAttributes(
			attribute.String("error.type", "signature_validation_failed"),
		))
		span.SetStatus(codes.Error, "manager signature validation failed")
		e.logger.Error(ctx, "Manager signature validation failed", observability.Int64("task_id", task.TaskID[0]), observability.String("trace_id", traceID), observability.Error(err))
		return false, err
	}
	e.logger.Info(ctx, "[1/7] Scheduler signature validation passed", observability.Int64("task_id", task.TaskID[0]), observability.String("trace_id", traceID))

	var (
		resultCh = make(chan struct {
			success bool
			err     error
		}, len(task.TargetData))
	)

	for i := range len(task.TargetData) {
		go func(idx int) {
			// Create child span for this task execution (since we're in a goroutine, we need to pass context)
			taskCtx, taskSpan := e.tracer.Start(ctx, "task.execute.target",
				observability.WithSpanKind(trace.SpanKindInternal),
				observability.WithAttributes(
					attribute.Int("task.index", idx),
					attribute.Int64("task.id", task.TargetData[idx].TaskID),
					attribute.String("target.chain_id", task.TargetData[idx].TargetChainID),
					attribute.String("target.contract_address", task.TargetData[idx].TargetContractAddress),
					attribute.String("target.function", task.TargetData[idx].TargetFunction),
				),
			)
			defer taskSpan.End()

			// check if trigger is valid
			isTriggerTrue, err := e.validator.ValidateTrigger(taskCtx, &task.TriggerData[idx], traceID)
			if !isTriggerTrue {
				taskSpan.RecordError(err, observability.WithErrorAttributes(
					attribute.String("error.type", "trigger_validation_failed"),
				))
				taskSpan.SetStatus(codes.Error, "trigger validation failed")
				e.logger.Error(taskCtx, "Trigger validation failed", observability.Int64("task_id", task.TaskID[0]), observability.String("trace_id", traceID), observability.Error(err))
				resultCh <- struct {
					success bool
					err     error
				}{false, err}
				return
			}

			// Add event for trigger validation
			taskSpan.AddEvent("trigger.validated", observability.WithEventAttributes(
				attribute.Bool("trigger.valid", isTriggerTrue),
			))
			e.logger.Info(taskCtx, "[2/7] Trigger validation passed", observability.Int64("task_id", task.TaskID[0]), observability.String("trace_id", traceID))

			// create a client for validating event based and performing action
			rpcURL := utils.GetChainRpcUrl(task.TargetData[idx].TargetChainID)
			client, err := ethclient.Dial(rpcURL)
			if err != nil {
				taskSpan.RecordError(err, observability.WithErrorAttributes(
					attribute.String("error.type", "chain_connection_failed"),
				))
				taskSpan.SetStatus(codes.Error, "failed to connect to chain")
				e.logger.Error(taskCtx, "Failed to connect to chain", observability.Int64("task_id", task.TaskID[0]), observability.String("trace_id", traceID), observability.Error(err))
				resultCh <- struct {
					success bool
					err     error
				}{false, err}
				return
			}
			defer client.Close()
			// e.logger.Debug(taskCtx, "Connected to chain", observability.String("rpc_url", rpcURL))

			//simulate the transaction before doing any action

			// execute the action (nonce is allocated inside executeAction just before tx submission)
			var actionData types.PerformerActionData
			var transactionSubmitted bool
			actionData, transactionSubmitted, err = e.executeAction(taskCtx, &task.TargetData[idx], &task.TriggerData[idx], client)
			if err != nil {
				taskSpan.RecordError(err, observability.WithErrorAttributes(
					attribute.String("error.type", "execution_failure"),
				))
				taskSpan.SetStatus(codes.Error, "action execution failed")
				e.logger.Error(taskCtx, "Failed to execute action", observability.Int64("task_id", task.TaskID[0]), observability.String("trace_id", traceID), observability.Error(err))
				// Report execution failure to taskmonitor (no CID yet)
				e.reportTaskStatus(taskCtx, task.TargetData[idx].TaskID, false, false, actionData.ActionTxHash, "", fmt.Sprintf("action execution failed: %v", err))
				resultCh <- struct {
					success bool
					err     error
				}{false, err}
				return
			}

			// Add event for action execution
			taskSpan.SetAttributes(
				attribute.String("execution.transaction_hash", actionData.ActionTxHash),
				attribute.Bool("execution.success", transactionSubmitted),
			)
			taskSpan.AddEvent("action.executed", observability.WithEventAttributes(
				attribute.String("execution.tx_hash", actionData.ActionTxHash),
				attribute.Bool("execution.success", transactionSubmitted),
			))
			e.logger.Info(taskCtx, "[3/7] Action execution completed", observability.Int64("task_id", task.TaskID[0]), observability.Bool("transaction_submitted", transactionSubmitted), observability.String("trace_id", traceID))

			ipfsData := types.IPFSData{
				TaskData: &types.SendTaskDataToKeeper{
					TaskID:           []int64{task.TargetData[idx].TaskID},
					PerformerData:    task.PerformerData,
					TargetData:       []types.TaskTargetData{task.TargetData[idx]},
					TriggerData:      []types.TaskTriggerData{task.TriggerData[idx]},
					SchedulerID:      task.SchedulerID,
					ManagerSignature: task.ManagerSignature,
					Network:          task.Network,
				},
				ActionData:         &actionData,
				ProofData:          &types.ProofData{},
				PerformerSignature: &types.PerformerSignatureData{},
			}
			ipfsData.ProofData.TaskID = task.TaskID[0]
			ipfsData.PerformerSignature.TaskID = task.TaskID[0]
			ipfsData.PerformerSignature.PerformerSigningAddress = config.GetConsensusAddress()

			tlsConfig := proof.DefaultTLSProofConfig(config.GetTLSProofHost())
			tlsConfig.TargetPort = config.GetTLSProofPort()
			proofData, err := proof.GenerateProofWithTLSConnection(taskCtx, e.logger, ipfsData, tlsConfig)
			if err != nil {
				e.logger.Error(taskCtx, "Failed to generate TLS proof, falling back to mock", observability.Int64("task_id", task.TaskID[0]), observability.String("trace_id", traceID), observability.Error(err))
			} else {
				e.logger.Info(taskCtx, "[4/7]TLS proof generated successfully", observability.Int64("task_id", task.TaskID[0]), observability.String("trace_id", traceID))
			}

			ipfsData.ProofData = &proofData

			// Create a copy of ipfsData without the signature for signing
			ipfsDataForSigning := types.IPFSData{
				TaskData:   ipfsData.TaskData,
				ActionData: ipfsData.ActionData,
				ProofData:  ipfsData.ProofData,
				PerformerSignature: &types.PerformerSignatureData{
					TaskID:                  ipfsData.PerformerSignature.TaskID,
					PerformerSigningAddress: ipfsData.PerformerSignature.PerformerSigningAddress,
					// Note: PerformerSignature field is intentionally left empty for signing
				},
			}

			performerSignature, err := cryptography.SignJSONMessage(ipfsDataForSigning, config.GetPrivateKeyConsensus())
			if err != nil {
				taskSpan.RecordError(err, observability.WithErrorAttributes(
					attribute.String("error.type", "signing_failed"),
				))
				taskSpan.SetStatus(codes.Error, "failed to sign IPFS data")
				e.logger.Error(taskCtx, "Failed to sign the ipfs data", observability.Int64("task_id", task.TaskID[0]), observability.String("trace_id", traceID), observability.Error(err))
				// Report failure (execution succeeded, but signing failed)
				e.reportTaskStatus(taskCtx, task.TargetData[idx].TaskID, transactionSubmitted, false, actionData.ActionTxHash, "", fmt.Sprintf("failed to sign IPFS data: %v", err))
				resultCh <- struct {
					success bool
					err     error
				}{false, err}
				return
			}
			ipfsData.PerformerSignature = &types.PerformerSignatureData{
				TaskID:                  task.TaskID[0],
				PerformerSignature:      performerSignature,
				PerformerSigningAddress: config.GetConsensusAddress(),
			}
			e.logger.Info(taskCtx, "[5/7] IPFS data signed", observability.Int64("task_id", task.TaskID[0]), observability.String("trace_id", traceID))

			// Extract trace context and embed in IPFS data before marshaling
			traceContext := observability.GetTraceContext(taskCtx)
			if traceContext != nil {
				ipfsData.TraceID = traceContext.TraceID
				ipfsData.SpanID = traceContext.SpanID
			}

			filename := fmt.Sprintf("proof_of_task_%d_%s.json", task.TaskID, time.Now().Format("20060102150405"))
			ipfsDataBytes, err := json.Marshal(ipfsData)
			if err != nil {
				taskSpan.RecordError(err, observability.WithErrorAttributes(
					attribute.String("error.type", "json_marshal_failed"),
				))
				taskSpan.SetStatus(codes.Error, "failed to marshal IPFS data")
				// Report failure (execution succeeded, but JSON marshal failed)
				e.reportTaskStatus(taskCtx, task.TargetData[idx].TaskID, transactionSubmitted, false, actionData.ActionTxHash, "", fmt.Sprintf("failed to marshal IPFS data: %v", err))
				resultCh <- struct {
					success bool
					err     error
				}{false, err}
				return
			}
			cid, err := e.validator.IpfsClient.Upload(taskCtx, filename, ipfsDataBytes)
			if err != nil {
				taskSpan.RecordError(err, observability.WithErrorAttributes(
					attribute.String("error.type", "ipfs_upload_failed"),
				))
				taskSpan.SetStatus(codes.Error, "IPFS upload failed")
				e.logger.Error(taskCtx, "Failed to upload IPFS data", observability.Int64("task_id", task.TaskID[0]), observability.String("trace_id", traceID), observability.Error(err))
				// Report failure (execution succeeded, but IPFS upload failed)
				e.reportTaskStatus(taskCtx, task.TargetData[idx].TaskID, transactionSubmitted, false, actionData.ActionTxHash, "", fmt.Sprintf("IPFS upload failed: %v", err))
				resultCh <- struct {
					success bool
					err     error
				}{false, err}
				return
			}
			e.logger.Info(taskCtx, "[6/7] IPFS data uploaded", observability.Int64("task_id", task.TaskID[0]), observability.String("trace_id", traceID))

			// Add event for IPFS upload
			taskSpan.AddEvent("ipfs.uploaded", observability.WithEventAttributes(
				attribute.String("ipfs.cid", cid),
			))

			// Create span for aggregator call
			aggCtx, aggSpan := e.tracer.Start(taskCtx, "task.aggregate",
				observability.WithSpanKind(trace.SpanKindClient),
				observability.WithAttributes(
					attribute.Int64("task.id", task.TargetData[idx].TaskID),
					attribute.String("aggregator.proof_of_task", proofData.ProofOfTask),
					attribute.String("ipfs.cid", cid),
				),
			)
			defer aggSpan.End()

			aggregatorData := types.BroadcastDataForValidators{
				ProofOfTask:      proofData.ProofOfTask,
				Data:             []byte(cid),
				TaskDefinitionID: task.TargetData[idx].TaskDefinitionID,
				PerformerAddress: config.GetConsensusAddress(),
			}

			// Get trace context for aggregator span event
			traceContextForAgg := observability.GetTraceContext(taskCtx)
			if traceContextForAgg != nil {
				aggSpan.AddEvent("aggregator.request.sent", observability.WithEventAttributes(
					attribute.String("trace.id", traceContextForAgg.TraceID),
				))
			}

			success, err := e.aggregatorClient.SendTaskToValidators(aggCtx, &aggregatorData)
			if !success {
				aggSpan.RecordError(err, observability.WithErrorAttributes(
					attribute.String("error.type", "aggregator_submission_failed"),
				))
				aggSpan.SetStatus(codes.Error, "failed to send to aggregator")
				e.logger.Error(taskCtx, "Failed to send task result to aggregator", observability.Int64("task_id", task.TaskID[0]), observability.String("trace_id", traceID), observability.Error(err))
				// Report failure with CID (execution succeeded, aggregator failed)
				errorMsg := "failed to send task result to aggregator"
				if err != nil {
					errorMsg = fmt.Sprintf("%s: %v", errorMsg, err)
				}
				e.reportTaskStatus(taskCtx, task.TargetData[idx].TaskID, transactionSubmitted, false, actionData.ActionTxHash, cid, errorMsg)
				resultCh <- struct {
					success bool
					err     error
				}{false, fmt.Errorf("failed to send task result to aggregator")}
				return
			}

			// Store broadcast data for potential rebroadcast
			e.storeBroadcastData(task.TargetData[idx].TaskID, &aggregatorData)

			// Both execution and aggregator submission succeeded
			aggSpan.SetStatus(codes.Ok, "task sent to aggregator successfully")
			e.logger.Info(taskCtx, "[7/7] Task result sent to aggregator", observability.Int64("task_id", task.TaskID[0]), observability.String("trace_id", traceID))
			e.reportTaskStatus(taskCtx, task.TargetData[idx].TaskID, transactionSubmitted, true, actionData.ActionTxHash, cid, "")

			resultCh <- struct {
				success bool
				err     error
			}{true, nil}
		}(i)
	}

	for range task.TargetData {
		res := <-resultCh
		if res.err != nil || !res.success {
			return false, res.err
		}
	}
	return true, nil
}

// getNonceManager returns or creates a nonce manager for the given chain
func (e *TaskExecutor) getNonceManager(chainID string) (*noncemanager.NonceManager, error) {
	e.nonceMutex.RLock()
	if nm, exists := e.nonceManagers[chainID]; exists {
		e.nonceMutex.RUnlock()
		return nm, nil
	}
	e.nonceMutex.RUnlock()

	e.nonceMutex.Lock()
	defer e.nonceMutex.Unlock()

	// Double-check after acquiring write lock
	if nm, exists := e.nonceManagers[chainID]; exists {
		return nm, nil
	}

	// Create new client and nonce manager
	rpcURL := utils.GetChainRpcUrl(chainID)
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create client for chain %s: %w", chainID, err)
	}

	nm := noncemanager.NewNonceManager(client, e.logger)
	if err := nm.Initialize(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to initialize nonce manager for chain %s: %w", chainID, err)
	}

	e.nonceManagers[chainID] = nm
	return nm, nil
}

// storeBroadcastData stores broadcast data for a task to enable rebroadcast
func (e *TaskExecutor) storeBroadcastData(taskID int64, data *types.BroadcastDataForValidators) {
	e.broadcastDataMutex.Lock()
	defer e.broadcastDataMutex.Unlock()
	e.broadcastData[taskID] = data
	e.logger.Debug(context.Background(), "Stored broadcast data for rebroadcast",
		observability.Int64("task_id", taskID))
}

// getBroadcastData retrieves broadcast data for a task
func (e *TaskExecutor) getBroadcastData(taskID int64) (*types.BroadcastDataForValidators, bool) {
	e.broadcastDataMutex.RLock()
	defer e.broadcastDataMutex.RUnlock()
	data, exists := e.broadcastData[taskID]
	return data, exists
}

// RebroadcastTask rebroadcasts a task to the aggregator
func (e *TaskExecutor) RebroadcastTask(ctx context.Context, taskID int64) error {
	// Get stored broadcast data
	data, exists := e.getBroadcastData(taskID)
	if !exists {
		e.logger.Warn(ctx, "Broadcast data not found for task, cannot rebroadcast",
			observability.Int64("task_id", taskID),
			observability.String("reason", "data_not_stored_or_keeper_restarted"))
		return fmt.Errorf("broadcast data not found for task %d (may have been lost due to keeper restart or task was not executed by this keeper)", taskID)
	}

	// Create span for rebroadcast
	ctx, span := e.tracer.Start(ctx, "task.rebroadcast",
		observability.WithSpanKind(trace.SpanKindClient),
		observability.WithAttributes(
			attribute.Int64("task.id", taskID),
		),
	)
	defer span.End()

	e.logger.Info(ctx, "Rebroadcasting task to aggregator",
		observability.Int64("task_id", taskID))

	// Re-send to aggregator
	success, err := e.aggregatorClient.SendTaskToValidators(ctx, data)
	if !success {
		span.RecordError(err)
		span.SetStatus(codes.Error, "rebroadcast failed")
		e.logger.Error(ctx, "Failed to rebroadcast task to aggregator",
			observability.Int64("task_id", taskID),
			observability.Error(err))
		return fmt.Errorf("failed to rebroadcast task: %w", err)
	}

	span.SetStatus(codes.Ok, "task rebroadcast successfully")
	e.logger.Info(ctx, "Task rebroadcast successfully",
		observability.Int64("task_id", taskID))

	return nil
}

// reportTaskStatus reports task execution status to taskmonitor (best-effort, doesn't block)
// This should be called after the aggregator submission attempt (regardless of success or failure)
func (e *TaskExecutor) reportTaskStatus(ctx context.Context, taskID int64, executionSuccessful, aggregatorSubmitted bool, executionTxHash, proofCID, errorMsg string) {
	// Preserve span context before goroutine to maintain trace continuity
	span := trace.SpanFromContext(ctx)
	spanContext := span.SpanContext()

	// Report status asynchronously to avoid blocking
	go func() {
		reportCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Re-inject span context to maintain trace continuity across goroutine
		if spanContext.HasTraceID() {
			reportCtx = trace.ContextWithSpanContext(reportCtx, spanContext)
		}

		if err := e.taskMonitorClient.ReportTaskStatus(reportCtx, taskID, executionSuccessful, aggregatorSubmitted, executionTxHash, proofCID, errorMsg); err != nil {
			e.logger.Debug(ctx, "Failed to report task status to taskmonitor",
				observability.Int64("task_id", taskID),
				observability.Bool("execution_successful", executionSuccessful),
				observability.Bool("aggregator_submitted", aggregatorSubmitted),
				observability.String("execution_tx_hash", executionTxHash),
				observability.String("proof_cid", proofCID),
				observability.Error(err))
		}
	}()
}

// --- DEPRECATED: ---
// Since executor and validator are controlled by us, backward compatibility is unnecessary.
//
// func (e *TaskExecutor) reportTaskError(taskID int64, errorMsg string) {
// 	if e.taskMonitorClient == nil {
// 		e.logger.Debug("TaskMonitor client not available, skipping error report",
// 			"task_id", taskID)
// 		return
// 	}
// 	go func() {
// 		reportCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 		defer cancel()
// 		if err := e.taskMonitorClient.ReportTaskError(reportCtx, taskID, errorMsg); err != nil {
// 			e.logger.Warn("Failed to report task error to taskmonitor",
// 				"task_id", taskID,
// 				"error", err)
// 		}
// 	}()
// }
