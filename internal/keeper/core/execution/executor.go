package execution

import (
	"context"
	"encoding/json"
	"fmt"
	// "strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/trigg3rX/triggerx-backend-imua/internal/keeper/config"
	"github.com/trigg3rX/triggerx-backend-imua/internal/keeper/core/validation"
	"github.com/trigg3rX/triggerx-backend-imua/internal/keeper/utils"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/client/aggregator"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/cryptography"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/docker"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/logging"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/proof"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/types"
)

// TaskExecutor is the default implementation of TaskExecutor
type TaskExecutor struct {
	alchemyAPIKey    string
	codeExecutor     *docker.CodeExecutor
	argConverter     *ArgumentConverter
	validator        *validation.TaskValidator
	aggregatorClient *aggregator.AggregatorClient
	logger           logging.Logger
}

// NewTaskExecutor creates a new instance of TaskExecutor
func NewTaskExecutor(
	alchemyAPIKey string,
	codeExecutor *docker.CodeExecutor,
	validator *validation.TaskValidator,
	aggregatorClient *aggregator.AggregatorClient,
	logger logging.Logger) *TaskExecutor {
	return &TaskExecutor{
		alchemyAPIKey:    alchemyAPIKey,
		codeExecutor:     codeExecutor,
		argConverter:     &ArgumentConverter{},
		validator:        validator,
		aggregatorClient: aggregatorClient,
		logger:           logger,
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
	isSchedulerSignatureTrue, err := e.validator.ValidateSchedulerSignature(task, traceID)
	if !isSchedulerSignatureTrue {
		e.logger.Error("Scheduler signature validation failed", "task_id", task.TaskID, "trace_id", traceID, "error", err)
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

			nonce, err := client.PendingNonceAt(context.Background(), common.HexToAddress(config.GetKeeperAddress()))
			if err != nil {
				e.logger.Error("Failed to get pending nonce", "task_id", task.TaskID, "trace_id", traceID, "error", err)
				resultCh <- struct {
					success bool
					err     error
				}{false, err}
				return
			}

			// execute the action
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
					TaskID:        task.TargetData[idx].TaskID,
					PerformerData: task.PerformerData,
					TargetData:    []types.TaskTargetData{task.TargetData[idx]},
					TriggerData:   []types.TaskTriggerData{task.TriggerData[idx]},
					SchedulerSignature: task.SchedulerSignature,
				},
				ActionData:         &actionData,
				ProofData:          &types.ProofData{},
				PerformerSignature: &types.PerformerSignatureData{},
			}
			ipfsData.ProofData.TaskID = task.TaskID
			ipfsData.PerformerSignature.TaskID = task.TaskID
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
			performerSignature, err := cryptography.SignJSONMessage(ipfsData, config.GetPrivateKeyConsensus())
			if err != nil {
				e.logger.Error("Failed to sign the ipfs data", "task_id", task.TaskID, "trace_id", traceID, "error", err)
				resultCh <- struct {
					success bool
					err     error
				}{false, err}
				return
			}
			ipfsData.PerformerSignature = &types.PerformerSignatureData{
				PerformerSignature:      performerSignature,
				PerformerSigningAddress: config.GetConsensusAddress(),
			}
			e.logger.Info("IPFS data signed", "task_id", task.TaskID, "trace_id", traceID)

			filename := fmt.Sprintf("proof_of_task_%d_%s.json", task.TaskID, time.Now().Format("20060102150405"))
			ipfsDataBytes, err := json.Marshal(ipfsData)
			if err != nil {
				resultCh <- struct {
					success bool
					err     error
				}{false, err}
				return
			}
			cid, err := utils.UploadToIPFS(filename, ipfsDataBytes)
			if err != nil {
				e.logger.Error("Failed to upload IPFS data", "task_id", task.TaskID, "trace_id", traceID, "error", err)
				resultCh <- struct {
					success bool
					err     error
				}{false, err}
				return
			}
			e.logger.Info("IPFS data uploaded", "task_id", task.TaskID, "trace_id", traceID)

			aggregatorData := types.BroadcastDataForValidators{
				ProofOfTask:        proofData.ProofOfTask,
				Data:               []byte(cid),
				TaskDefinitionID:   task.TargetData[idx].TaskDefinitionID,
				PerformerAddress:   config.GetConsensusAddress(),
			}

			success, err := e.aggregatorClient.SendTaskToValidators(ctx, &aggregatorData)
			if !success {
				e.logger.Error("Failed to send task result to aggregator", "task_id", task.TaskID, "error", err, "trace_id", traceID)
				resultCh <- struct {
					success bool
					err     error
				}{false, fmt.Errorf("failed to send task result to aggregator")}
				return
			}
			e.logger.Info("Task result sent to aggregator", "task_id", task.TaskID, "trace_id", traceID)
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

// func parseStringToInt(str string) int {
// 	num, err := strconv.Atoi(str)
// 	if err != nil {
// 		return 0
// 	}
// 	return num
// }