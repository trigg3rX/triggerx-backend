package validation

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/utils"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

const timeTolerance = 5 * time.Second
const expirationTimeTolerance = 8 * time.Second

// getReceiptRetryConfig returns a retry configuration optimized for L2 chains
func getReceiptRetryConfig(logger logging.Logger) *retry.RetryConfig {
	return &retry.RetryConfig{
		MaxRetries:      12,                     // Very aggressive for receipt fetching
		InitialDelay:    300 * time.Millisecond, // Start with 300ms
		MaxDelay:        6 * time.Second,        // Cap at 6 seconds
		BackoffFactor:   1.4,                    // Moderate backoff
		JitterFactor:    0.3,                    // High jitter to avoid conflicts
		LogRetryAttempt: true,
		ShouldRetry:     shouldRetryReceiptError,
	}
}

// shouldRetryReceiptError determines if a receipt fetch error should be retried
func shouldRetryReceiptError(err error, attempt int) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())

	// Retry on network/connection issues
	if strings.Contains(errStr, "connection") ||
		strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "network") ||
		strings.Contains(errStr, "dial") ||
		strings.Contains(errStr, "refused") ||
		strings.Contains(errStr, "unavailable") {
		return true
	}

	// Retry on rate limiting
	if strings.Contains(errStr, "rate limit") ||
		strings.Contains(errStr, "too many requests") ||
		strings.Contains(errStr, "429") {
		return true
	}

	// Retry on temporary server errors
	if strings.Contains(errStr, "500") ||
		strings.Contains(errStr, "502") ||
		strings.Contains(errStr, "503") ||
		strings.Contains(errStr, "504") {
		return true
	}

	// Retry on "not found" errors (transaction might be pending)
	if strings.Contains(errStr, "not found") ||
		strings.Contains(errStr, "unknown transaction") {
		return true
	}

	return false
}

func (v *TaskValidator) ValidateAction(targetData *types.TaskTargetData, triggerData *types.TaskTriggerData, actionData *types.PerformerActionData, client *ethclient.Client, traceID string) (bool, error) {
	// v.logger.Infof("txHash: %s", actionData.ActionTxHash)
	// time.Sleep(10 * time.Second)
	// Fetch the tx details from the action data
	txHash := common.HexToHash(actionData.ActionTxHash)

	// Get retry configuration
	retryConfig := getReceiptRetryConfig(v.logger)

	// Try to get transaction receipt with retry logic
	receiptOperation := func() (*ethtypes.Receipt, error) {
		return client.TransactionReceipt(context.Background(), txHash)
	}

	receipt, err := retry.Retry(context.Background(), receiptOperation, retryConfig, v.logger)
	if err != nil || receipt == nil {
		// If receipt fetch failed, try to get transaction status with retry logic
		txOperation := func() (bool, error) {
			_, isPending, err := client.TransactionByHash(context.Background(), txHash)
			return isPending, err
		}

		isPending, err := retry.Retry(context.Background(), txOperation, retryConfig, v.logger)
		if err != nil {
			return false, fmt.Errorf("failed to get transaction after retries: %v", err)
		}
		if isPending {
			return false, fmt.Errorf("transaction is pending")
		}
		return false, fmt.Errorf("transaction is not found")
	}

	// check if the tx is successful
	if receipt.Status != 1 {
		return false, fmt.Errorf("transaction is not successful")
	}

	// TODO: get the action tx check right
	// check if the tx was made to correct target contract
	// fetch the AA contract address and the transaction from there to complete the flow

	txTimestamp, err := v.getBlockTimestamp(receipt, utils.GetChainRpcUrl(targetData.TargetChainID))
	if err != nil {
		return false, fmt.Errorf("failed to get block timestamp: %v", err)
	}

	// check if the tx was made before expiration time + tolerance
	if txTimestamp.After(triggerData.ExpirationTime.Add(expirationTimeTolerance)) {
		return false, fmt.Errorf("transaction was made after the expiration time by %v", txTimestamp.Sub(triggerData.ExpirationTime.Add(expirationTimeTolerance)))
	}

	// check if the task was time, if yes, check if it was executed within the time interval + tolerance
	if targetData.TaskDefinitionID == 1 || targetData.TaskDefinitionID == 2 {
		if txTimestamp.After(triggerData.NextTriggerTimestamp.Add(timeTolerance)) {
			return false, fmt.Errorf("transaction was made after the next execution timestamp by %v", txTimestamp.Sub(triggerData.NextTriggerTimestamp.Add(timeTolerance)))
		}
		// if txTimestamp.Before(triggerData.NextTriggerTimestamp.Add(-timeTolerance)) {
		// 	return false, fmt.Errorf("transaction was made before the next execution timestamp by %v", triggerData.NextTriggerTimestamp.Add(-timeTolerance).Sub(txTimestamp))
		// }
	}
	return true, nil
}
