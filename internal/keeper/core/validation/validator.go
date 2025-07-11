package validation

import (
	"context"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/trigg3rX/triggerx-backend-imua/internal/keeper/utils"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/client/aggregator"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/docker"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/logging"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/types"
)

type TaskValidator struct {
	alchemyAPIKey    string
	etherscanAPIKey  string
	codeExecutor     *docker.CodeExecutor
	aggregatorClient *aggregator.AggregatorClient
	logger           logging.Logger
}

func NewTaskValidator(alchemyAPIKey string, etherscanAPIKey string, codeExecutor *docker.CodeExecutor, aggregatorClient *aggregator.AggregatorClient, logger logging.Logger) *TaskValidator {
	return &TaskValidator{
		alchemyAPIKey:    alchemyAPIKey,
		etherscanAPIKey:  etherscanAPIKey,
		codeExecutor:     codeExecutor,
		aggregatorClient: aggregatorClient,
		logger:           logger,
	}
}

func (v *TaskValidator) ValidateTask(ctx context.Context, ipfsData types.IPFSData, traceID string) (bool, error) {
	// check if the scheduler signature is valid
	isSchedulerSignatureTrue, err := v.ValidateSchedulerSignature(ipfsData.TaskData, traceID)
	if !isSchedulerSignatureTrue {
		v.logger.Error("Scheduler signature validation failed", "task_id", ipfsData.TaskData.TaskID, "trace_id", traceID, "error", err)
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
