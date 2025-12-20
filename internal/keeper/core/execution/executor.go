package execution

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	// "strconv"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
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
	nonceManagers     map[string]*NonceManager // Chain ID -> NonceManager
	nonceMutex        sync.RWMutex
}

// NewTaskExecutor creates a new instance of TaskExecutor
func NewTaskExecutor(
	alchemyAPIKey string,
	validator *validation.TaskValidator,
	aggregatorClient *aggregator.AggregatorClient,
	taskMonitorClient TaskMonitorClientInterface,
	logger observability.Logger) *TaskExecutor {
	return &TaskExecutor{
		alchemyAPIKey:     alchemyAPIKey,
		argConverter:      &ArgumentConverter{},
		validator:         validator,
		aggregatorClient:  aggregatorClient,
		taskMonitorClient: taskMonitorClient,
		logger:            logger,
		nonceManagers:     make(map[string]*NonceManager),
	}
}

func (e *TaskExecutor) ExecuteTask(ctx context.Context, task *types.SendTaskDataToKeeper, traceID string) (bool, error) {
	// Check for nil task
	if task == nil {
		e.logger.Error(ctx, "Task data is nil", observability.String("trace_id", traceID))
		return false, fmt.Errorf("task data cannot be nil")
	}

	// Check for nil TargetData and TriggerData
	if task.TargetData == nil {
		e.logger.Error(ctx, "TargetData is nil", observability.Int64("task_id", task.TaskID[0]), observability.String("trace_id", traceID))
		return false, fmt.Errorf("target data cannot be nil")
	}
	if task.TriggerData == nil {
		e.logger.Error(ctx, "TriggerData is nil", observability.Int64("task_id", task.TaskID[0]), observability.String("trace_id", traceID))
		return false, fmt.Errorf("trigger data cannot be nil")
	}

	// check if the scheduler signature is valid
	isManagerSignatureTrue, err := e.validator.ValidateManagerSignature(ctx, task, traceID)
	if !isManagerSignatureTrue {
		e.logger.Error(ctx, "Manager signature validation failed", observability.Int64("task_id", task.TaskID[0]), observability.String("trace_id", traceID), observability.Error(err))
		return false, err
	}
	e.logger.Info(ctx, "Scheduler signature validation passed", observability.Int64("task_id", task.TaskID[0]), observability.String("trace_id", traceID))

	var (
		resultCh = make(chan struct {
			success bool
			err     error
		}, len(task.TargetData))
	)

	for i := range len(task.TargetData) {
		go func(idx int) {
			// check if trigger is valid
			isTriggerTrue, err := e.validator.ValidateTrigger(ctx, &task.TriggerData[idx], traceID)
			if !isTriggerTrue {
				e.logger.Error(ctx, "Trigger validation failed", observability.Int64("task_id", task.TaskID[0]), observability.String("trace_id", traceID), observability.Error(err))
				resultCh <- struct {
					success bool
					err     error
				}{false, err}
				return
			}
			e.logger.Info(ctx, "Trigger validation passed", observability.Int64("task_id", task.TaskID[0]), observability.String("trace_id", traceID))

			// create a client for validating event based and performing action
			rpcURL := utils.GetChainRpcUrl(task.TargetData[idx].TargetChainID)
			client, err := ethclient.Dial(rpcURL)
			if err != nil {
				e.logger.Error(ctx, "Failed to connect to chain", observability.Int64("task_id", task.TaskID[0]), observability.String("trace_id", traceID), observability.Error(err))
				resultCh <- struct {
					success bool
					err     error
				}{false, err}
				return
			}
			defer client.Close()
			e.logger.Debug(ctx, "Connected to chain: %s", observability.String("rpc_url", rpcURL))

			//simulate the transaction before doing any action

			// execute the action (nonce is allocated inside executeAction just before tx submission)
			var actionData types.PerformerActionData
			var transactionSubmitted bool
			actionData, transactionSubmitted, err = e.executeAction(ctx, &task.TargetData[idx], &task.TriggerData[idx], client)
			if err != nil {
				e.logger.Error(ctx, "Failed to execute action", observability.Int64("task_id", task.TaskID[0]), observability.String("trace_id", traceID), observability.Error(err))
				// Report execution failure to taskmonitor (no CID yet)
				e.reportTaskStatus(ctx, task.TargetData[idx].TaskID, false, false, actionData.ActionTxHash, "", fmt.Sprintf("action execution failed: %v", err))
				resultCh <- struct {
					success bool
					err     error
				}{false, err}
				return
			}

			// If execution was skipped (e.g., custom script returned shouldExecute=false)
			if !transactionSubmitted {
				e.logger.Info(ctx, "Execution skipped (no transaction submitted)", observability.Int64("task_id", task.TaskID[0]), observability.String("trace_id", traceID))
			}
			e.logger.Info(ctx, "Action execution completed", observability.Int64("task_id", task.TaskID[0]), observability.String("trace_id", traceID))

			ipfsData := types.IPFSData{
				TaskData: &types.SendTaskDataToKeeper{
					TaskID:           []int64{task.TargetData[idx].TaskID},
					PerformerData:    task.PerformerData,
					TargetData:       []types.TaskTargetData{task.TargetData[idx]},
					TriggerData:      []types.TaskTriggerData{task.TriggerData[idx]},
					SchedulerID:      task.SchedulerID,
					ManagerSignature: task.ManagerSignature,
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
			proofData, err := proof.GenerateProofWithTLSConnection(ipfsData, tlsConfig)
			if err != nil {
				e.logger.Error(ctx, "Failed to generate TLS proof, falling back to mock", observability.Int64("task_id", task.TaskID[0]), observability.String("trace_id", traceID), observability.Error(err))
			} else {
				e.logger.Info(ctx, "TLS proof generated successfully", observability.Int64("task_id", task.TaskID[0]), observability.String("trace_id", traceID))
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
				e.logger.Error(ctx, "Failed to sign the ipfs data", observability.Int64("task_id", task.TaskID[0]), observability.String("trace_id", traceID), observability.Error(err))
				// Report failure (execution succeeded, but signing failed)
				e.reportTaskStatus(ctx, task.TargetData[idx].TaskID, transactionSubmitted, false, actionData.ActionTxHash, "", fmt.Sprintf("failed to sign IPFS data: %v", err))
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
			e.logger.Info(ctx, "IPFS data signed", observability.Int64("task_id", task.TaskID[0]), observability.String("trace_id", traceID))

			filename := fmt.Sprintf("proof_of_task_%d_%s.json", task.TaskID, time.Now().Format("20060102150405"))
			ipfsDataBytes, err := json.Marshal(ipfsData)
			if err != nil {
				// Report failure (execution succeeded, but JSON marshal failed)
				e.reportTaskStatus(ctx, task.TargetData[idx].TaskID, transactionSubmitted, false, actionData.ActionTxHash, "", fmt.Sprintf("failed to marshal IPFS data: %v", err))
				resultCh <- struct {
					success bool
					err     error
				}{false, err}
				return
			}
			cid, err := e.validator.IpfsClient.Upload(ctx, filename, ipfsDataBytes)
			if err != nil {
				e.logger.Error(ctx, "Failed to upload IPFS data", observability.Int64("task_id", task.TaskID[0]), observability.String("trace_id", traceID), observability.Error(err))
				// Report failure (execution succeeded, but IPFS upload failed)
				e.reportTaskStatus(ctx, task.TargetData[idx].TaskID, transactionSubmitted, false, actionData.ActionTxHash, "", fmt.Sprintf("IPFS upload failed: %v", err))
				resultCh <- struct {
					success bool
					err     error
				}{false, err}
				return
			}
			e.logger.Info(ctx, "IPFS data uploaded", observability.Int64("task_id", task.TaskID[0]), observability.String("trace_id", traceID))

			aggregatorData := types.BroadcastDataForValidators{
				ProofOfTask:      proofData.ProofOfTask,
				Data:             []byte(cid),
				TaskDefinitionID: task.TargetData[idx].TaskDefinitionID,
				PerformerAddress: config.GetConsensusAddress(),
			}

			success, err := e.aggregatorClient.SendTaskToValidators(ctx, &aggregatorData)
			if !success {
				e.logger.Error(ctx, "Failed to send task result to aggregator", observability.Int64("task_id", task.TaskID[0]), observability.String("trace_id", traceID), observability.Error(err))
				// Report failure with CID (execution succeeded, aggregator failed)
				errorMsg := "failed to send task result to aggregator"
				if err != nil {
					errorMsg = fmt.Sprintf("%s: %v", errorMsg, err)
				}
				e.reportTaskStatus(ctx, task.TargetData[idx].TaskID, transactionSubmitted, false, actionData.ActionTxHash, cid, errorMsg)
				resultCh <- struct {
					success bool
					err     error
				}{false, fmt.Errorf("failed to send task result to aggregator")}
				return
			}

			// Both execution and aggregator submission succeeded
			e.logger.Info(ctx, "Task result sent to aggregator", observability.Int64("task_id", task.TaskID[0]), observability.String("trace_id", traceID))
			e.reportTaskStatus(ctx, task.TargetData[idx].TaskID, transactionSubmitted, true, actionData.ActionTxHash, cid, "")

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
func (e *TaskExecutor) getNonceManager(chainID string) (*NonceManager, error) {
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

	nm := NewNonceManager(client, e.logger)
	if err := nm.Initialize(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to initialize nonce manager for chain %s: %w", chainID, err)
	}

	e.nonceManagers[chainID] = nm
	return nm, nil
}

// reportTaskStatus reports task execution status to taskmonitor (best-effort, doesn't block)
// This should be called after the aggregator submission attempt (regardless of success or failure)
func (e *TaskExecutor) reportTaskStatus(ctx context.Context, taskID int64, executionSuccessful, aggregatorSubmitted bool, executionTxHash, proofCID, errorMsg string) {
	if e.taskMonitorClient == nil {
		e.logger.Debug(ctx, "TaskMonitor client not available, skipping status report",
			observability.Int64("task_id", taskID))
		return
	}

	// Report status asynchronously to avoid blocking
	go func() {
		reportCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := e.taskMonitorClient.ReportTaskStatus(reportCtx, taskID, executionSuccessful, aggregatorSubmitted, executionTxHash, proofCID, errorMsg); err != nil {
			e.logger.Warn(ctx, "Failed to report task status to taskmonitor",
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
