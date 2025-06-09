package execution

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/core/validation"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/utils"
	"github.com/trigg3rX/triggerx-backend/pkg/client/aggregator"
	"github.com/trigg3rX/triggerx-backend/pkg/cryptography"
	"github.com/trigg3rX/triggerx-backend/pkg/docker"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/proof"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
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
	// check if the scheduler signature is valid
	isSchedulerSignatureTrue, err := e.validator.ValidateSchedulerSignature(task, traceID)
	if !isSchedulerSignatureTrue {
		e.logger.Error("Scheduler signature validation failed", "task_id", task.TaskID, "trace_id", traceID, "error", err)
		return false, err
	}
	e.logger.Info("Scheduler signature validation passed", "task_id", task.TaskID, "trace_id", traceID)

	// check if trigger is valid
	isTriggerTrue, err := e.validator.ValidateTrigger(task.TriggerData, traceID)
	if !isTriggerTrue {
		e.logger.Error("Trigger validation failed", "task_id", task.TaskID, "trace_id", traceID, "error", err)
		return false, err
	}
	e.logger.Info("Trigger validation passed", "task_id", task.TaskID, "trace_id", traceID)

	// create a client for validating event based and performing action
	rpcURL := utils.GetChainRpcUrl(task.TargetData.TargetChainID)
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		e.logger.Error("Failed to connect to chain", "task_id", task.TaskID, "trace_id", traceID, "error", err)
		return false, err
	}
	defer client.Close()

	// execute the action
	var actionData types.PerformerActionData
	switch task.TargetData.TaskDefinitionID {
	case 1, 3, 5:
		actionData, err = e.executeActionWithStaticArgs(task.TargetData, client)
		if err != nil {
			e.logger.Error("Failed to execute action with static args", "task_id", task.TaskID, "trace_id", traceID, "error", err)
			return false, err
		}
	case 2, 4, 6:
		actionData, err = e.executeActionWithDynamicArgs(task.TargetData, client)
		if err != nil {
			e.logger.Error("Failed to execute action with static args", "task_id", task.TaskID, "trace_id", traceID, "error", err)
			return false, err
		}
	default:
		return false, fmt.Errorf("unsupported task definition id: %d", task.TargetData.TaskDefinitionID)
	}
	e.logger.Info("Action execution completed", "task_id", task.TaskID, "trace_id", traceID)

	// generate proof data with real TLS connection
	ipfsData := types.IPFSData{
		TaskData:           task,
		ActionData:         &actionData,
		ProofData:          &types.ProofData{},
		PerformerSignature: &types.PerformerSignatureData{},
	}
	ipfsData.ProofData.TaskID = task.TaskID
	ipfsData.PerformerSignature.TaskID = task.TaskID
	ipfsData.PerformerSignature.PerformerSigningAddress = config.GetConsensusAddress()

	// Generate TLS proof using configured host
	tlsConfig := proof.DefaultTLSProofConfig(config.GetTLSProofHost())
	tlsConfig.TargetPort = config.GetTLSProofPort()
	proofData, err := proof.GenerateProofWithTLSConnection(ipfsData, tlsConfig)
	if err != nil {
		e.logger.Error("Failed to generate TLS proof, falling back to mock", "task_id", task.TaskID, "trace_id", traceID, "error", err)
	} else {
		e.logger.Info("TLS proof generated successfully", "task_id", task.TaskID, "trace_id", traceID)
	}

	// sign the ipfs data
	ipfsData.ProofData = &proofData
	performerSignature, err := cryptography.SignJSONMessage(ipfsData, config.GetPrivateKeyConsensus())
	if err != nil {
		e.logger.Error("Failed to sign the ipfs data", "task_id", task.TaskID, "trace_id", traceID, "error", err)
		return false, err
	}
	ipfsData.PerformerSignature = &types.PerformerSignatureData{
		PerformerSignature:      performerSignature,
		PerformerSigningAddress: config.GetConsensusAddress(),
	}
	e.logger.Info("IPFS data signed", "task_id", task.TaskID, "trace_id", traceID)

	// upload the ipfs data to ipfs
	filename := fmt.Sprintf("proof_of_task_%d_%s.json", task.TaskID, time.Now().Format("20060102150405"))
	ipfsDataBytes, err := json.Marshal(ipfsData)
	if err != nil {
		return false, err
	}
	cid, err := utils.UploadToIPFS(filename, ipfsDataBytes)
	if err != nil {
		e.logger.Error("Failed to upload IPFS data", "task_id", task.TaskID, "trace_id", traceID, "error", err)
		return false, err
	}
	e.logger.Info("IPFS data uploaded", "task_id", task.TaskID, "trace_id", traceID)

	// send the ipfs data to the aggregator
	aggregatorData := types.BroadcastDataForValidators{
		ProofOfTask:        proofData.ProofOfTask,
		Data:               []byte(cid),
		TaskDefinitionID:   task.TargetData.TaskDefinitionID,
		PerformerAddress:   config.GetConsensusAddress(),
		PerformerSignature: ipfsData.PerformerSignature.PerformerSignature,
		SignatureType:      "ecdsa",
		TargetChainID:      parseStringToInt(task.TargetData.TargetChainID),
	}

	success, err := e.aggregatorClient.SendTaskToValidators(ctx, &aggregatorData)
	if !success {
		e.logger.Error("Failed to send task result to aggregator", "task_id", task.TaskID, "error", err, "trace_id", traceID)
		return false, fmt.Errorf("failed to send task result to aggregator")
	}
	e.logger.Info("Task result sent to aggregator", "task_id", task.TaskID, "trace_id", traceID)
	return true, nil
}

func parseStringToInt(str string) int {
	num, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}
	return num
}
