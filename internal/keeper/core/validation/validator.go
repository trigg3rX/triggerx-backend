package validation

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/trigg3rX/triggerx-backend/pkg/cryptography"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// EthClientInterface defines the interface for Ethereum client operations
type EthClientInterface interface {
	CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error)
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*ethTypes.Receipt, error)
	TransactionByHash(ctx context.Context, hash common.Hash) (tx *ethTypes.Transaction, isPending bool, err error)
	BlockByNumber(ctx context.Context, number *big.Int) (*ethTypes.Block, error)
	BlockByHash(ctx context.Context, hash common.Hash) (*ethTypes.Block, error)
	Close()
}

// TaskValidatorInterface defines the interface for task validation
type TaskValidatorInterface interface {
	ValidateTask(task *types.SendTaskDataToKeeper, traceID string) (bool, error)
	ValidateTrigger(triggerData *types.TaskTriggerData, traceID string) (bool, error)
	ValidateAction(targetData *types.TaskTargetData, actionData *types.PerformerActionData, client EthClientInterface, traceID string) (bool, error)
	ValidateProof(ipfsData types.IPFSData, traceID string) (bool, error)
}

// TaskValidator implements the TaskValidatorInterface
type TaskValidator struct {
	logger         logging.Logger
	ethClientMaker func(url string) (EthClientInterface, error)
	crypto         CryptographyInterface
}

// CryptographyInterface defines the interface for cryptography operations
type CryptographyInterface interface {
	VerifySignatureFromJSON(jsonData interface{}, signature string, signerAddress string) (bool, error)
}

// CryptographyWrapper wraps the cryptography package
type CryptographyWrapper struct{}

// VerifySignatureFromJSON implements the CryptographyInterface
func (c *CryptographyWrapper) VerifySignatureFromJSON(jsonData interface{}, signature string, signerAddress string) (bool, error) {
	return cryptography.VerifySignatureFromJSON(jsonData, signature, signerAddress)
}

// NewTaskValidator creates a new TaskValidator instance
func NewTaskValidator(logger logging.Logger) *TaskValidator {
	return &TaskValidator{
		logger: logger,
		ethClientMaker: func(url string) (EthClientInterface, error) {
			client, err := ethclient.Dial(url)
			if err != nil {
				return nil, err
			}
			return client, nil
		},
		crypto: &CryptographyWrapper{},
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
