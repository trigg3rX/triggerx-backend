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
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/proof"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// TaskMonitorClientInterface defines the interface for taskmonitor client operations
type TaskMonitorClientInterface interface {
	// ReportTaskStatus reports task execution status to taskmonitor (both success and failure)
	// proofCID contains all execution data (task data, action data, proof, signatures)
	ReportTaskStatus(ctx context.Context, taskID int64, success bool, proofCID, errorMsg string) error
}

// TaskExecutor is the default implementation of TaskExecutor
type TaskExecutor struct {
	alchemyAPIKey     string
	argConverter      *ArgumentConverter
	validator         *validation.TaskValidator
	aggregatorClient  *aggregator.AggregatorClient
	taskMonitorClient TaskMonitorClientInterface
	logger            logging.Logger
	nonceManagers     map[string]*NonceManager // Chain ID -> NonceManager
	nonceMutex        sync.RWMutex
}

// NewTaskExecutor creates a new instance of TaskExecutor
func NewTaskExecutor(
	alchemyAPIKey string,
	validator *validation.TaskValidator,
	aggregatorClient *aggregator.AggregatorClient,
	taskMonitorClient TaskMonitorClientInterface,
	logger logging.Logger) *TaskExecutor {
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
		e.logger.Error("Task data is nil", "trace_id", traceID)
		return false, fmt.Errorf("task data cannot be nil")
	}

	// Check for nil TargetData and TriggerData
	if task.TargetData == nil {
		e.logger.Error("TargetData is nil", "task_id", task.TaskID, "trace_id", traceID)
		return false, fmt.Errorf("target data cannot be nil")
	}
	if task.TriggerData == nil {
		e.logger.Error("TriggerData is nil", "task_id", task.TaskID, "trace_id", traceID)
		return false, fmt.Errorf("trigger data cannot be nil")
	}

	// check if the scheduler signature is valid
	isManagerSignatureTrue, err := e.validator.ValidateManagerSignature(task, traceID)
	if !isManagerSignatureTrue {
		e.logger.Error("Manager signature validation failed", "task_id", task.TaskID, "trace_id", traceID, "error", err)
		return false, err
	}
	e.logger.Info("Scheduler signature validation passed", "task_id", task.TaskID, "trace_id", traceID)

	var (
		resultCh = make(chan struct {
			success bool
			err     error
		}, len(task.TargetData))
	)

	for i := range len(task.TargetData) {
		go func(idx int) {
			// check if trigger is valid
			isTriggerTrue, err := e.validator.ValidateTrigger(&task.TriggerData[idx], traceID)
			if !isTriggerTrue {
				e.logger.Error("Trigger validation failed", "task_id", task.TaskID, "trace_id", traceID, "error", err)
				resultCh <- struct {
					success bool
					err     error
				}{false, err}
				return
			}
			e.logger.Info("Trigger validation passed", "task_id", task.TaskID, "trace_id", traceID)

			// create a client for validating event based and performing action
			rpcURL := utils.GetChainRpcUrl(task.TargetData[idx].TargetChainID)
			client, err := ethclient.Dial(rpcURL)
			if err != nil {
				e.logger.Error("Failed to connect to chain", "task_id", task.TaskID, "trace_id", traceID, "error", err)
				resultCh <- struct {
					success bool
					err     error
				}{false, err}
				return
			}
			defer client.Close()
			e.logger.Debugf("Connected to chain: %s", rpcURL)

			//simulate the transaction before doing any action

			// execute the action (nonce is allocated inside executeAction just before tx submission)
			var actionData types.PerformerActionData
			var transactionSubmitted bool
			actionData, transactionSubmitted, err = e.executeAction(&task.TargetData[idx], &task.TriggerData[idx], client)
			if err != nil {
				e.logger.Error("Failed to execute action", "task_id", task.TaskID, "trace_id", traceID, "error", err)
				// Report execution failure to taskmonitor (no CID yet)
				e.reportTaskStatus(task.TargetData[idx].TaskID, false, "", fmt.Sprintf("action execution failed: %v", err))
				resultCh <- struct {
					success bool
					err     error
				}{false, err}
				return
			}

			// If execution was skipped (e.g., custom script returned shouldExecute=false)
			if !transactionSubmitted {
				e.logger.Info("Execution skipped (no transaction submitted)", "task_id", task.TaskID, "trace_id", traceID)
			}
			e.logger.Info("Action execution completed", "task_id", task.TaskID, "trace_id", traceID)

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
				e.logger.Error("Failed to generate TLS proof, falling back to mock", "task_id", task.TaskID, "trace_id", traceID, "error", err)
			} else {
				e.logger.Info("TLS proof generated successfully", "task_id", task.TaskID, "trace_id", traceID)
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
				e.logger.Error("Failed to sign the ipfs data", "task_id", task.TaskID, "trace_id", traceID, "error", err)
				// Report failure (no CID yet)
				e.reportTaskStatus(task.TargetData[idx].TaskID, false, "", fmt.Sprintf("failed to sign IPFS data: %v", err))
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
			e.logger.Info("IPFS data signed", "task_id", task.TaskID, "trace_id", traceID)

			filename := fmt.Sprintf("proof_of_task_%d_%s.json", task.TaskID, time.Now().Format("20060102150405"))
			ipfsDataBytes, err := json.Marshal(ipfsData)
			if err != nil {
				// Report failure (no CID yet)
				e.reportTaskStatus(task.TargetData[idx].TaskID, false, "", fmt.Sprintf("failed to marshal IPFS data: %v", err))
				resultCh <- struct {
					success bool
					err     error
				}{false, err}
				return
			}
			cid, err := e.validator.IpfsClient.Upload(ctx, filename, ipfsDataBytes)
			if err != nil {
				e.logger.Error("Failed to upload IPFS data", "task_id", task.TaskID, "trace_id", traceID, "error", err)
				// Report failure (no CID yet)
				e.reportTaskStatus(task.TargetData[idx].TaskID, false, "", fmt.Sprintf("IPFS upload failed: %v", err))
				resultCh <- struct {
					success bool
					err     error
				}{false, err}
				return
			}
			e.logger.Info("IPFS data uploaded", "task_id", task.TaskID, "trace_id", traceID)

			aggregatorData := types.BroadcastDataForValidators{
				ProofOfTask:      proofData.ProofOfTask,
				Data:             []byte(cid),
				TaskDefinitionID: task.TargetData[idx].TaskDefinitionID,
				PerformerAddress: config.GetConsensusAddress(),
			}

			success, err := e.aggregatorClient.SendTaskToValidators(ctx, &aggregatorData)
			if !success {
				e.logger.Error("Failed to send task result to aggregator", "task_id", task.TaskID, "error", err, "trace_id", traceID)
				// Report failure with CID (execution succeeded, aggregator failed)
				errorMsg := "failed to send task result to aggregator"
				if err != nil {
					errorMsg = fmt.Sprintf("%s: %v", errorMsg, err)
				}
				e.reportTaskStatus(task.TargetData[idx].TaskID, false, cid, errorMsg)
				resultCh <- struct {
					success bool
					err     error
				}{false, fmt.Errorf("failed to send task result to aggregator")}
				return
			}

			// Both execution and aggregator submission succeeded
			e.logger.Info("Task result sent to aggregator", "task_id", task.TaskID, "trace_id", traceID)
			e.reportTaskStatus(task.TargetData[idx].TaskID, true, cid, "")

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
// proofCID contains all execution data (task data, action data, proof, signatures)
func (e *TaskExecutor) reportTaskStatus(taskID int64, success bool, proofCID, errorMsg string) {
	if e.taskMonitorClient == nil {
		e.logger.Debug("TaskMonitor client not available, skipping status report",
			"task_id", taskID)
		return
	}

	// Report status asynchronously to avoid blocking
	go func() {
		reportCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := e.taskMonitorClient.ReportTaskStatus(reportCtx, taskID, success, proofCID, errorMsg); err != nil {
			e.logger.Warn("Failed to report task status to taskmonitor",
				"task_id", taskID,
				"success", success,
				"proof_cid", proofCID,
				"error", err)
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
