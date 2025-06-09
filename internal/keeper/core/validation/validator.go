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

// ValidateTask validates a task
func (v *TaskValidator) ValidateTask(task *types.SendTaskDataToKeeper, traceID string) (bool, error) {
	// Validate trigger
	triggerValid, err := v.ValidateTrigger(task.TriggerData, traceID)
	if err != nil {
		v.logger.Error("Failed to validate trigger", "error", err, "traceID", traceID)
		return false, fmt.Errorf("failed to validate trigger: %v", err)
	}
	if !triggerValid {
		v.logger.Error("Trigger validation failed", "traceID", traceID)
		return false, fmt.Errorf("trigger validation failed")
	}
	v.logger.Info("Trigger validation successful", "traceID", traceID)

	// Validate target
	targetValid, err := v.ValidateTarget(task.TargetData, traceID)
	if err != nil {
		v.logger.Error("Failed to validate target", "error", err, "traceID", traceID)
		return false, fmt.Errorf("failed to validate target: %v", err)
	}
	if !targetValid {
		v.logger.Error("Target validation failed", "traceID", traceID)
		return false, fmt.Errorf("target validation failed")
	}
	v.logger.Info("Target validation successful", "traceID", traceID)

	return true, nil
}

// ValidateTarget validates a target
func (v *TaskValidator) ValidateTarget(targetData *types.TaskTargetData, traceID string) (bool, error) {
	// TODO: Implement target validation
	return true, nil
}
