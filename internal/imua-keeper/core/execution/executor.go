package execution

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"

	// "strconv"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/trigg3rX/triggerx-backend/internal/imua-keeper/chainio"
	"github.com/trigg3rX/triggerx-backend/internal/imua-keeper/config"
	"github.com/trigg3rX/triggerx-backend/internal/imua-keeper/core/validation"
	"github.com/trigg3rX/triggerx-backend/internal/imua-keeper/utils"
	"github.com/trigg3rX/triggerx-backend/pkg/client/aggregator"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/proof"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// TaskExecutor is the default implementation of TaskExecutor
type TaskExecutor struct {
	alchemyAPIKey    string
	argConverter     *ArgumentConverter
	validator        *validation.TaskValidator
	aggregatorClient *aggregator.AggregatorClient
	avsWriter        chainio.AvsWriter
	logger           logging.Logger
	nonceManagers    map[string]*NonceManager // Chain ID -> NonceManager
	nonceMutex       sync.RWMutex
}

// NewTaskExecutor creates a new instance of TaskExecutor
func NewTaskExecutor(
	alchemyAPIKey string,
	validator *validation.TaskValidator,
	aggregatorClient *aggregator.AggregatorClient,
	avsWriter chainio.AvsWriter,
	logger logging.Logger) *TaskExecutor {
	return &TaskExecutor{
		alchemyAPIKey:    alchemyAPIKey,
		argConverter:     &ArgumentConverter{},
		validator:        validator,
		aggregatorClient: aggregatorClient,
		avsWriter:        avsWriter,
		logger:           logger,
		nonceManagers:    make(map[string]*NonceManager),
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

			// Get nonce manager for this chain
			nonceManager, err := e.getNonceManager(task.TargetData[idx].TargetChainID)
			if err != nil {
				e.logger.Error("Failed to get nonce manager", "task_id", task.TaskID, "trace_id", traceID, "error", err)
				resultCh <- struct {
					success bool
					err     error
				}{false, err}
				return
			}

			// Get next nonce atomically
			nonce, err := nonceManager.GetNextNonce(context.Background())
			if err != nil {
				e.logger.Error("Failed to get nonce", "task_id", task.TaskID, "trace_id", traceID, "error", err)
				resultCh <- struct {
					success bool
					err     error
				}{false, err}
				return
			}

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

			// execute the action with the allocated nonce
			var actionData types.PerformerActionData
			actionData, err = e.executeAction(&task.TargetData[idx], &task.TriggerData[idx], nonce, client)
			if err != nil {
				e.logger.Error("Failed to execute action", "task_id", task.TaskID, "trace_id", traceID, "error", err)
				resultCh <- struct {
					success bool
					err     error
				}{false, err}
				return
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
			ipfsData.PerformerSignature.PerformerSigningAddress = string(config.GetConsensusKeyPair().PublicKey().Marshal())

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
			ipfsDataForSigningBytes, err := json.Marshal(ipfsDataForSigning)
			if err != nil {
				e.logger.Error("Failed to marshal ipfs data for signing", "task_id", task.TaskID, "trace_id", traceID, "error", err)
				resultCh <- struct {
					success bool
					err     error
				}{false, err}
				return
			}
			// Convert the data to 32-byte hash for BLS signing
			var messageHash [32]byte
			copy(messageHash[:], ipfsDataForSigningBytes)

			// Sign using Prysmaticlabs BLS
			performerSignature := config.GetConsensusKeyPair().Sign(messageHash[:])

			// Encode signature as base64 for proper serialization
			signatureBytes := performerSignature.Marshal()
			performerSignatureEncoded := base64.StdEncoding.EncodeToString(signatureBytes)

			// Encode public key as base64 for proper serialization
			publicKeyBytes := config.GetConsensusKeyPair().PublicKey().Marshal()
			performerSigningAddressEncoded := base64.StdEncoding.EncodeToString(publicKeyBytes)

			ipfsData.PerformerSignature = &types.PerformerSignatureData{
				TaskID:                  task.TaskID[0],
				PerformerSignature:      performerSignatureEncoded,
				PerformerSigningAddress: performerSigningAddressEncoded,
			}
			e.logger.Info("IPFS data signed", "task_id", task.TaskID, "trace_id", traceID)
			e.logger.Info("Performer signature", "performerSignature", ipfsData.PerformerSignature)

			// filename := fmt.Sprintf("proof_of_task_%d_%s.json", task.TaskID, time.Now().Format("20060102150405"))
			// ipfsDataBytes, err := json.Marshal(ipfsData)
			// if err != nil {
			// 	resultCh <- struct {
			// 		success bool
			// 		err     error
			// 	}{false, err}
			// 	return
			// }
			// _, err = utils.UploadToIPFS(filename, ipfsDataBytes)
			// if err != nil {
			// 	e.logger.Error("Failed to upload IPFS data", "task_id", task.TaskID, "trace_id", traceID, "error", err)
			// 	resultCh <- struct {
			// 		success bool
			// 		err     error
			// 	}{false, err}
			// 	return
			// }
			// e.logger.Info("IPFS data uploaded", "task_id", task.TaskID, "trace_id", traceID)

			// Create task on-chain instead of sending to aggregator
			taskName := fmt.Sprintf("Task_%d_%s", task.TaskID[0], time.Now().Format("20060102150405"))
			taskDefinitionId := uint8(task.TargetData[idx].TaskDefinitionID)

			// Use default values for task parameters
			taskResponsePeriod := uint64(2)
			taskChallengePeriod := uint64(2)
			thresholdPercentage := uint8(80) // 80% threshold
			taskStatisticalPeriod := uint64(2)

			// Create task data from IPFS data
			taskData, err := json.Marshal(ipfsData)
			if err != nil {
				e.logger.Error("Failed to marshal task data", "task_id", task.TaskID, "trace_id", traceID, "error", err)
				resultCh <- struct {
					success bool
					err     error
				}{false, err}
				return
			}

			// Create task on-chain
			receipt, err := e.avsWriter.CreateNewTask(
				ctx,
				taskName,
				taskDefinitionId,
				taskData,
				taskResponsePeriod,
				taskChallengePeriod,
				thresholdPercentage,
				taskStatisticalPeriod,
			)
			if err != nil {
				e.logger.Error("Failed to create task on-chain", "task_id", task.TaskID, "trace_id", traceID, "error", err)
				resultCh <- struct {
					success bool
					err     error
				}{false, fmt.Errorf("failed to create task on-chain: %w", err)}
				return
			}

			e.logger.Info("Task created on-chain successfully",
				"task_id", task.TaskID,
				"trace_id", traceID,
				"tx_hash", receipt.TxHash.String(),
				"block_number", receipt.BlockNumber)

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

// func parseStringToInt(str string) int {
// 	num, err := strconv.Atoi(str)
// 	if err != nil {
// 		return 0
// 	}
// 	return num
// }
