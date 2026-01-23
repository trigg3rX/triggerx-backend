package validation

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/utils"
	"github.com/trigg3rX/triggerx-backend/pkg/client/aggregator"
	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor"
	"github.com/trigg3rX/triggerx-backend/pkg/ipfs"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type TaskValidator struct {
	alchemyAPIKey    string
	etherscanAPIKey  string
	dockerExecutor   dockerexecutor.DockerExecutorAPI
	aggregatorClient *aggregator.AggregatorClient
	logger           observability.Logger
	tracer           observability.Tracer
	IpfsClient       ipfs.IPFSClient
}

func NewTaskValidator(
	alchemyAPIKey string,
	etherscanAPIKey string,
	dockerExecutor dockerexecutor.DockerExecutorAPI,
	aggregatorClient *aggregator.AggregatorClient,
	logger observability.Logger,
	tracer observability.Tracer,
	ipfsClient ipfs.IPFSClient,
) *TaskValidator {
	return &TaskValidator{
		alchemyAPIKey:    alchemyAPIKey,
		etherscanAPIKey:  etherscanAPIKey,
		dockerExecutor:   dockerExecutor,
		aggregatorClient: aggregatorClient,
		logger:           logger,
		tracer:           tracer,
		IpfsClient:       ipfsClient,
	}
}

func (v *TaskValidator) ValidateTask(ctx context.Context, data string, traceID string) (bool, error) {
	// Decode the data if it's hex-encoded (with 0x prefix)
	dataBytes, err := hex.DecodeString(data[2:]) // Remove "0x" prefix before decoding
	if err != nil {
		v.logger.Error(ctx, "Failed to hex-decode data", observability.String("trace_id", traceID), observability.Error(err))
		return false, err
	}
	decodedData := string(dataBytes)
	ipfsData, err := v.IpfsClient.Fetch(ctx, decodedData)
	if err != nil {
		v.logger.Error(ctx, "Failed to fetch IPFS content", observability.String("trace_id", traceID), observability.Error(err))
		return false, err
	}

	// Continue trace from IPFS data if available
	if ipfsData.TraceID != "" {
		ctx = observability.ContinueTrace(ctx, ipfsData.TraceID, ipfsData.SpanID)
		traceID = ipfsData.TraceID

		// Start a new span for validation linked to the performer's trace
		var span observability.Span
		ctx, span = v.tracer.Start(ctx, "ValidateTaskFromIPFS")
		defer span.End()

		v.logger.Info(ctx, "Continuing trace from IPFS data", observability.String("trace_id", traceID))
	}

	// Network matching: Check if the task's network matches the keeper's configured network
	if string(ipfsData.TaskData.Network) != string(config.GetNetwork()) {
		v.logger.Info(ctx, "Task validation rejected due to network mismatch",
			observability.Int64("task_id", ipfsData.TaskData.TaskID[0]),
			observability.String("task_network", string(ipfsData.TaskData.Network)),
			observability.String("keeper_network", string(config.GetNetwork())),
			observability.String("trace_id", traceID))
		return false, fmt.Errorf("network mismatch: task network %s does not match keeper network %s", string(ipfsData.TaskData.Network), string(config.GetNetwork()))
	}

	v.logger.Info(ctx, "[0/5] Network validation passed", observability.Int64("task_id", ipfsData.TaskData.TaskID[0]), observability.String("network", string(ipfsData.TaskData.Network)), observability.String("trace_id", traceID))

	// check if the scheduler signature is valid
	isManagerSignatureTrue, err := v.ValidateManagerSignature(ctx, ipfsData.TaskData, traceID)
	if !isManagerSignatureTrue {
		v.logger.Error(ctx, "Manager signature validation failed", observability.Int64("task_id", ipfsData.TaskData.TaskID[0]), observability.String("trace_id", traceID), observability.Error(err))
		return false, err
	}
	v.logger.Info(ctx, "[1/5] Scheduler signature validation passed", observability.Int64("task_id", ipfsData.TaskData.TaskID[0]), observability.String("trace_id", traceID))

	rpcURL := utils.GetChainRpcUrl(ipfsData.TaskData.TargetData[0].TargetChainID)
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		v.logger.Error(ctx, "Failed to connect to chain", observability.Int64("task_id", ipfsData.TaskData.TaskID[0]), observability.String("trace_id", traceID), observability.Error(err))
		return false, err
	}
	defer client.Close()

	// check if trigger is valid
	isTriggerTrue, err := v.ValidateTrigger(ctx, &ipfsData.TaskData.TriggerData[0], traceID)
	if !isTriggerTrue {
		v.logger.Error(ctx, "Trigger validation failed", observability.Int64("task_id", ipfsData.TaskData.TaskID[0]), observability.String("trace_id", traceID), observability.Error(err))
		return false, err
	}
	v.logger.Info(ctx, "[2/5] Trigger validation successful", observability.String("traceID", traceID))

	// check if the action is valid
	isActionTrue, err := v.ValidateAction(&ipfsData.TaskData.TargetData[0], &ipfsData.TaskData.TriggerData[0], ipfsData.ActionData, client, traceID)
	if !isActionTrue {
		v.logger.Error(ctx, "Action validation failed", observability.Int64("task_id", ipfsData.TaskData.TaskID[0]), observability.String("trace_id", traceID), observability.Error(err))
		return false, err
	}
	v.logger.Info(ctx, "[3/5] Action validation passed", observability.Int64("task_id", ipfsData.TaskData.TaskID[0]), observability.String("trace_id", traceID))

	// validate the proof data
	isProofTrue, err := v.ValidateProof(ctx, ipfsData, traceID)
	if !isProofTrue {
		v.logger.Error(ctx, "Proof validation failed", observability.Int64("task_id", ipfsData.TaskData.TaskID[0]), observability.String("trace_id", traceID), observability.Error(err))
		return false, err
	}
	v.logger.Info(ctx, "[4/5] Proof validation passed", observability.Int64("task_id", ipfsData.TaskData.TaskID[0]), observability.String("trace_id", traceID))

	// check if the performer signature is valid
	isPerformerSignatureTrue, err := v.ValidatePerformerSignature(ctx, ipfsData, traceID)
	if !isPerformerSignatureTrue {
		v.logger.Error(ctx, "Performer signature validation failed", observability.Int64("task_id", ipfsData.TaskData.TaskID[0]), observability.String("trace_id", traceID), observability.Error(err))
		return false, err
	}
	v.logger.Info(ctx, "[5/5] Performer signature validation passed", observability.String("traceID", traceID))

	return true, nil
}

// ValidateTarget validates a target
func (v *TaskValidator) ValidateTarget(ctx context.Context, targetData *types.TaskTargetData, traceID string) (bool, error) {
	// TODO: Implement target validation
	return true, nil
}

// GetDockerManager returns the DockerManager instance
func (v *TaskValidator) GetDockerExecutor() dockerexecutor.DockerExecutorAPI {
	return v.dockerExecutor
}
