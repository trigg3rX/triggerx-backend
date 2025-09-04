package validation

import (
	"context"
	"encoding/hex"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/utils"
	"github.com/trigg3rX/triggerx-backend/pkg/client/aggregator"
	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor"
	"github.com/trigg3rX/triggerx-backend/pkg/ipfs"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type TaskValidator struct {
	alchemyAPIKey    string
	etherscanAPIKey  string
	dockerExecutor   dockerexecutor.DockerExecutorAPI
	aggregatorClient *aggregator.AggregatorClient
	logger           logging.Logger
	IpfsClient       ipfs.IPFSClient
}

func NewTaskValidator(
	alchemyAPIKey string,
	etherscanAPIKey string,
	dockerExecutor dockerexecutor.DockerExecutorAPI,
	aggregatorClient *aggregator.AggregatorClient,
	logger logging.Logger,
	ipfsClient ipfs.IPFSClient,
) *TaskValidator {
	return &TaskValidator{
		alchemyAPIKey:    alchemyAPIKey,
		etherscanAPIKey:  etherscanAPIKey,
		dockerExecutor:   dockerExecutor,
		aggregatorClient: aggregatorClient,
		logger:           logger,
		IpfsClient:       ipfsClient,
	}
}

func (v *TaskValidator) ValidateTask(ctx context.Context, data string, traceID string) (bool, error) {
	// Decode the data if it's hex-encoded (with 0x prefix)
	dataBytes, err := hex.DecodeString(data[2:]) // Remove "0x" prefix before decoding
	if err != nil {
		v.logger.Error("Failed to hex-decode data", "trace_id", traceID, "error", err)
		return false, err
	}
	decodedData := string(dataBytes)
	ipfsData, err := v.IpfsClient.Fetch(decodedData)
	if err != nil {
		v.logger.Error("Failed to fetch IPFS content", "trace_id", traceID, "error", err)
		return false, err
	}

	// check if the scheduler signature is valid
	isManagerSignatureTrue, err := v.ValidateManagerSignature(ipfsData.TaskData, traceID)
	if !isManagerSignatureTrue {
		v.logger.Error("Manager signature validation failed", "task_id", ipfsData.TaskData.TaskID, "trace_id", traceID, "error", err)
		return false, err
	}
	v.logger.Info("Scheduler signature validation passed", "task_id", ipfsData.TaskData.TaskID, "trace_id", traceID)

	rpcURL := utils.GetChainRpcUrl(ipfsData.TaskData.TargetData[0].TargetChainID)
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		v.logger.Error("Failed to connect to chain", "task_id", ipfsData.TaskData.TaskID, "trace_id", traceID, "error", err)
		return false, err
	}
	defer client.Close()

	// check if trigger is valid
	isTriggerTrue, err := v.ValidateTrigger(&ipfsData.TaskData.TriggerData[0], traceID)
	if !isTriggerTrue {
		v.logger.Error("Trigger validation failed", "task_id", ipfsData.TaskData.TaskID, "trace_id", traceID, "error", err)
		return false, err
	}
	v.logger.Info("Trigger validation successful", "traceID", traceID)

	// check if the action is valid
	isActionTrue, err := v.ValidateAction(&ipfsData.TaskData.TargetData[0], &ipfsData.TaskData.TriggerData[0], ipfsData.ActionData, client, traceID)
	if !isActionTrue {
		v.logger.Error("Action validation failed", "task_id", ipfsData.TaskData.TaskID, "trace_id", traceID, "error", err)
		return false, err
	}
	v.logger.Info("Action validation passed", "task_id", ipfsData.TaskData.TaskID, "trace_id", traceID)

	// validate the proof data
	isProofTrue, err := v.ValidateProof(ipfsData, traceID)
	if !isProofTrue {
		v.logger.Error("Proof validation failed", "task_id", ipfsData.TaskData.TaskID, "trace_id", traceID, "error", err)
		return false, err
	}
	v.logger.Info("Proof validation passed", "task_id", ipfsData.TaskData.TaskID, "trace_id", traceID)

	// check if the performer signature is valid
	isPerformerSignatureTrue, err := v.ValidatePerformerSignature(ipfsData, traceID)
	if !isPerformerSignatureTrue {
		v.logger.Error("Performer signature validation failed", "task_id", ipfsData.TaskData.TaskID, "trace_id", traceID, "error", err)
		return false, err
	}
	v.logger.Info("Target validation successful", "traceID", traceID)

	return true, nil
}

// ValidateTarget validates a target
func (v *TaskValidator) ValidateTarget(targetData *types.TaskTargetData, traceID string) (bool, error) {
	// TODO: Implement target validation
	return true, nil
}

// GetDockerManager returns the DockerManager instance
func (v *TaskValidator) GetDockerExecutor() dockerexecutor.DockerExecutorAPI {
	return v.dockerExecutor
}
