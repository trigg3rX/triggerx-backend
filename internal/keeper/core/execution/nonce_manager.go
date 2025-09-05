package execution

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
)

// NonceManager handles nonce allocation and transaction retry logic
type NonceManager struct {
	mu           sync.Mutex
	currentNonce uint64
	client       *ethclient.Client
	address      common.Address
	logger       logging.Logger
	lastSyncTime time.Time
	syncInterval time.Duration

	// Transaction tracking
	pendingTxs map[uint64]*PendingTransaction
	txMutex    sync.RWMutex

	// Retry configurations for different operations
	rpcRetryConfig     *retry.RetryConfig
	submitRetryConfig  *retry.RetryConfig
	confirmRetryConfig *retry.RetryConfig
}

type PendingTransaction struct {
	Nonce        uint64
	TxHash       string
	CreatedAt    time.Time
	Status       string // "pending", "confirmed", "failed", "replaced"
	Attempts     int
	LastGasPrice *big.Int
	Data         []byte
	To           common.Address
	ChainID      *big.Int
	PrivateKey   *ecdsa.PrivateKey
}

// NewNonceManager creates a new nonce manager with optimized retry settings
func NewNonceManager(client *ethclient.Client, logger logging.Logger) *NonceManager {
	return &NonceManager{
		client:       client,
		address:      common.HexToAddress(config.GetKeeperAddress()),
		logger:       logger,
		syncInterval: 10 * time.Second, // More frequent sync for L2 chains
		pendingTxs:   make(map[uint64]*PendingTransaction),

		// Aggressive retry configs optimized for L2 chains (1-2 sec block time)
		rpcRetryConfig: &retry.RetryConfig{
			MaxRetries:      8,                      // More retries for RPC calls
			InitialDelay:    200 * time.Millisecond, // Fast initial retry
			MaxDelay:        5 * time.Second,        // Cap at 5 seconds
			BackoffFactor:   1.5,                    // Moderate backoff
			JitterFactor:    0.3,                    // High jitter to avoid conflicts
			LogRetryAttempt: true,
			ShouldRetry:     shouldRetryRPCError,
		},

		submitRetryConfig: &retry.RetryConfig{
			MaxRetries:      10,                     // Very aggressive for submission
			InitialDelay:    100 * time.Millisecond, // Very fast initial retry
			MaxDelay:        3 * time.Second,        // Shorter max delay for L2
			BackoffFactor:   1.3,                    // Gentle backoff
			JitterFactor:    0.4,                    // High jitter
			LogRetryAttempt: true,
			ShouldRetry:     shouldRetrySubmissionError,
		},

		confirmRetryConfig: &retry.RetryConfig{
			MaxRetries:      15,                     // Very aggressive for confirmation
			InitialDelay:    500 * time.Millisecond, // Start with 500ms for confirmation
			MaxDelay:        8 * time.Second,        // Allow up to 8 seconds
			BackoffFactor:   1.2,                    // Very gentle backoff
			JitterFactor:    0.2,                    // Lower jitter for confirmation
			LogRetryAttempt: true,
			ShouldRetry:     shouldRetryConfirmationError,
		},
	}
}

// Initialize sets up the initial nonce
func (nm *NonceManager) Initialize(ctx context.Context) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	return nm.syncWithBlockchain(ctx)
}

// GetNextNonce returns the next available nonce atomically
func (nm *NonceManager) GetNextNonce(ctx context.Context) (uint64, error) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	// Sync with blockchain if needed
	if time.Since(nm.lastSyncTime) > nm.syncInterval {
		if err := nm.syncWithBlockchain(ctx); err != nil {
			return 0, fmt.Errorf("failed to sync nonce with blockchain: %w", err)
		}
	}

	nonce := nm.currentNonce
	nm.currentNonce++

	nm.logger.Debugf("Allocated nonce: %d", nonce)
	return nonce, nil
}

// syncWithBlockchain updates the current nonce from the blockchain
func (nm *NonceManager) syncWithBlockchain(ctx context.Context) error {
	operation := func() (uint64, error) {
		return nm.client.PendingNonceAt(ctx, nm.address)
	}

	pendingNonce, err := retry.Retry(ctx, operation, nm.rpcRetryConfig, nm.logger)
	if err != nil {
		return fmt.Errorf("failed to get pending nonce after retries: %w", err)
	}

	// Use the higher of pending nonce or our current nonce
	if pendingNonce > nm.currentNonce {
		nm.currentNonce = pendingNonce
		nm.logger.Infof("Synced nonce with blockchain: %d", nm.currentNonce)
	}

	nm.lastSyncTime = time.Now()
	return nil
}

// SubmitTransaction submits a transaction with intelligent retry logic
func (nm *NonceManager) SubmitTransaction(
	ctx context.Context,
	nonce uint64,
	to common.Address,
	data []byte,
	chainID *big.Int,
	privateKey *ecdsa.PrivateKey,
) (*types.Receipt, string, error) {

	// Check if we should replace an existing transaction
	nm.txMutex.RLock()
	if existingTx, exists := nm.pendingTxs[nonce]; exists && existingTx.Status == "pending" {
		nm.txMutex.RUnlock()

		// If existing tx is older than 30 seconds, replace it
		if time.Since(existingTx.CreatedAt) > 30*time.Second {
			return nm.replaceTransaction(ctx, existingTx, data, to, chainID, privateKey)
		}

		// Otherwise, wait for the existing transaction
		return nm.waitForExistingTransaction(ctx, existingTx)
	}
	nm.txMutex.RUnlock()

	// Submit new transaction
	return nm.submitNewTransaction(ctx, nonce, to, data, chainID, privateKey)
}

// submitNewTransaction submits a new transaction with EIP-1559 support
func (nm *NonceManager) submitNewTransaction(
	ctx context.Context,
	nonce uint64,
	to common.Address,
	data []byte,
	chainID *big.Int,
	privateKey *ecdsa.PrivateKey,
) (*types.Receipt, string, error) {

	// Get current gas price
	gasPrice, err := nm.getOptimalGasParams(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get gas price: %w", err)
	}

	// Create and sign transaction
	if gasPrice == nil {
		return nil, "", fmt.Errorf("gas price is required for transaction")
	}

	tx := types.NewTransaction(nonce, to, big.NewInt(0), 300000, gasPrice, data)
	signedTx, signErr := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if signErr != nil {
		return nil, "", fmt.Errorf("failed to sign transaction: %w", signErr)
	}

	// Track the transaction
	nm.trackTransaction(nonce, signedTx.Hash().Hex(), data, to, chainID, privateKey, gasPrice)

	return nm.submitWithRetry(ctx, signedTx, nonce, privateKey)
}

// replaceTransaction replaces a stuck transaction with higher fees
func (nm *NonceManager) replaceTransaction(
	ctx context.Context,
	existingTx *PendingTransaction,
	data []byte,
	to common.Address,
	chainID *big.Int,
	privateKey *ecdsa.PrivateKey,
) (*types.Receipt, string, error) {

	nm.logger.Infof("Replacing stuck transaction with nonce %d", existingTx.Nonce)

	// Get higher gas price for replacement
	gasPrice, err := nm.getOptimalGasParams(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get gas price: %w", err)
	}

	// Increase fees by 20% for replacement
	gasPrice = new(big.Int).Mul(gasPrice, big.NewInt(120))
	gasPrice = new(big.Int).Div(gasPrice, big.NewInt(100))

	// Create legacy replacement transaction
	tx := types.NewTransaction(existingTx.Nonce, to, big.NewInt(0), 300000, gasPrice, data)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return nil, "", fmt.Errorf("failed to sign replacement transaction: %w", err)
	}

	// Update tracking
	nm.updateTransactionStatus(existingTx.Nonce, signedTx.Hash().Hex(), gasPrice)

	return nm.submitWithRetry(ctx, signedTx, existingTx.Nonce, privateKey)
}

// waitForExistingTransaction waits for an existing transaction to be confirmed
func (nm *NonceManager) waitForExistingTransaction(ctx context.Context, existingTx *PendingTransaction) (*types.Receipt, string, error) {
	nm.logger.Infof("Waiting for existing transaction with nonce %d: %s", existingTx.Nonce, existingTx.TxHash)

	// Wait for the transaction to be confirmed
	receipt, err := bind.WaitMined(ctx, nm.client, &types.Transaction{})
	if err != nil {
		return nil, "", fmt.Errorf("failed to wait for existing transaction: %w", err)
	}

	nm.markTransactionConfirmed(existingTx.Nonce, existingTx.TxHash)
	return receipt, existingTx.TxHash, nil
}

// submitWithRetry handles the actual submission with intelligent retry logic
func (nm *NonceManager) submitWithRetry(ctx context.Context, signedTx *types.Transaction, nonce uint64, privateKey *ecdsa.PrivateKey) (*types.Receipt, string, error) {
	// First, submit the transaction with retry logic
	submitOperation := func() (string, error) {
		err := nm.client.SendTransaction(ctx, signedTx)
		if err != nil {
			// Check if it's a nonce too low error - this means we need to sync
			if isNonceTooLowError(err) {
				if syncErr := nm.syncWithBlockchain(ctx); syncErr != nil {
					return "", fmt.Errorf("failed to sync after nonce error: %w", syncErr)
				}
			}
			return "", err
		}

		txHash := signedTx.Hash().Hex()
		nm.logger.Infof("Transaction sent: %s", txHash)
		return txHash, nil
	}

	txHash, err := retry.Retry(ctx, submitOperation, nm.submitRetryConfig, nm.logger)
	if err != nil {
		return nil, "", fmt.Errorf("failed to submit transaction after retries: %w", err)
	}

	// Now wait for confirmation with retry logic
	confirmOperation := func() (*types.Receipt, error) {
		// Create a timeout context for this specific confirmation attempt
		confirmCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		receipt, err := bind.WaitMined(confirmCtx, nm.client, signedTx)
		if err != nil {
			// If we get a timeout, we might need to create a replacement transaction
			if confirmCtx.Err() == context.DeadlineExceeded {
				nm.logger.Warnf("Transaction %s confirmation timed out, will create replacement", txHash)
				return nil, fmt.Errorf("confirmation timeout")
			}
			return nil, err
		}

		return receipt, nil
	}

	receipt, err := retry.Retry(ctx, confirmOperation, nm.confirmRetryConfig, nm.logger)
	if err != nil {
		// If confirmation failed, try to create a replacement transaction
		nm.logger.Warnf("Transaction %s confirmation failed, creating replacement: %v", txHash, err)

		// Create replacement transaction with higher fees
		signedReplacementTx, replaceErr := nm.createReplacementTransaction(signedTx, 1, privateKey)
		if replaceErr != nil {
			return nil, "", fmt.Errorf("failed to create replacement transaction: %w", replaceErr)
		}

		// Update tracking
		nm.updateTransactionStatus(nonce, signedReplacementTx.Hash().Hex(), signedReplacementTx.GasPrice())

		// Recursively try with the replacement transaction (but limit depth)
		return nm.submitWithRetry(ctx, signedReplacementTx, nonce, privateKey)
	}

	// Success!
	nm.markTransactionConfirmed(nonce, txHash)
	nm.logger.Infof("Transaction confirmed: %s", txHash)
	return receipt, txHash, nil
}

// getOptimalGasParams gets optimal gas price for current network conditions with retry logic
func (nm *NonceManager) getOptimalGasParams(ctx context.Context) (*big.Int, error) {
	operation := func() (*big.Int, error) {
		return nm.client.SuggestGasPrice(ctx)
	}

	gasPrice, err := retry.Retry(ctx, operation, nm.rpcRetryConfig, nm.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price after retries: %w", err)
	}

	// Add 20% buffer for network congestion
	gasPrice.Mul(gasPrice, big.NewInt(120))
	gasPrice.Div(gasPrice, big.NewInt(100))

	nm.logger.Debugf("Using legacy gas price: %s", gasPrice.String())
	return gasPrice, nil
}

// createReplacementTransaction creates a replacement transaction with higher fees
func (nm *NonceManager) createReplacementTransaction(originalTx *types.Transaction, attempt int, privateKey *ecdsa.PrivateKey) (*types.Transaction, error) {
	// Get the transaction data
	var to common.Address
	var data []byte
	var chainID *big.Int

	switch tx := originalTx.Type(); tx {
	case 0: // Legacy transaction
		to = *originalTx.To()
		data = originalTx.Data()
		chainID = originalTx.ChainId()
	case 2: // EIP-1559 transaction - convert to legacy
		to = *originalTx.To()
		data = originalTx.Data()
		chainID = originalTx.ChainId()
	default:
		return nil, fmt.Errorf("unsupported transaction type: %d", tx)
	}

	// Get higher gas price with retry logic
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	gasPrice, err := nm.getOptimalGasParams(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price for replacement: %w", err)
	}

	// Increase fees by 20% for each attempt
	feeMultiplier := big.NewInt(int64(120 + (attempt * 20))) // 120%, 140%, 160%, etc.
	feeDivisor := big.NewInt(100)

	gasPrice.Mul(gasPrice, feeMultiplier)
	gasPrice.Div(gasPrice, feeDivisor)

	// Create and sign the replacement transaction
	tx := types.NewTransaction(originalTx.Nonce(), to, big.NewInt(0), 300000, gasPrice, data)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign replacement transaction: %w", err)
	}

	return signedTx, nil
}

// Helper methods for transaction tracking
func (nm *NonceManager) trackTransaction(nonce uint64, txHash string, data []byte, to common.Address, chainID *big.Int, privateKey *ecdsa.PrivateKey, gasPrice *big.Int) {
	nm.txMutex.Lock()
	defer nm.txMutex.Unlock()

	nm.pendingTxs[nonce] = &PendingTransaction{
		Nonce:        nonce,
		TxHash:       txHash,
		CreatedAt:    time.Now(),
		Status:       "pending",
		Attempts:     1,
		LastGasPrice: gasPrice,
		Data:         data,
		To:           to,
		ChainID:      chainID,
		PrivateKey:   privateKey,
	}
}

func (nm *NonceManager) updateTransactionStatus(nonce uint64, txHash string, gasPrice *big.Int) {
	nm.txMutex.Lock()
	defer nm.txMutex.Unlock()

	if tx, exists := nm.pendingTxs[nonce]; exists {
		tx.TxHash = txHash
		tx.Status = "pending"
		tx.Attempts++
		tx.LastGasPrice = gasPrice
		tx.CreatedAt = time.Now()
	}
}

func (nm *NonceManager) markTransactionConfirmed(nonce uint64, txHash string) {
	nm.txMutex.Lock()
	defer nm.txMutex.Unlock()

	if tx, exists := nm.pendingTxs[nonce]; exists {
		tx.Status = "confirmed"
		tx.TxHash = txHash
	}
}

// Custom error predicates for different operation types
func shouldRetryRPCError(err error, attempt int) bool {
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

	return false
}

func shouldRetrySubmissionError(err error, attempt int) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())

	// Retry on network issues
	if shouldRetryRPCError(err, attempt) {
		return true
	}

	// Retry on nonce issues (but not too many times)
	if attempt < 3 && (strings.Contains(errStr, "nonce too low") ||
		strings.Contains(errStr, "replacement transaction underpriced") ||
		strings.Contains(errStr, "already known")) {
		return true
	}

	// Retry on gas estimation issues
	if strings.Contains(errStr, "gas") &&
		(strings.Contains(errStr, "estimate") || strings.Contains(errStr, "limit")) {
		return true
	}

	// Retry on temporary blockchain issues
	if strings.Contains(errStr, "insufficient funds") && attempt < 2 {
		return true // Might be a temporary balance issue
	}

	return false
}

func shouldRetryConfirmationError(err error, attempt int) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())

	// Retry on network issues
	if shouldRetryRPCError(err, attempt) {
		return true
	}

	// Retry on transaction not found (might be pending)
	if strings.Contains(errStr, "not found") ||
		strings.Contains(errStr, "unknown transaction") {
		return true
	}

	// Retry on timeout errors
	if strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "deadline exceeded") {
		return true
	}

	return false
}

// Utility functions
func isNonceTooLowError(err error) bool {
	errStr := err.Error()
	return strings.Contains(errStr, "nonce too low") ||
		strings.Contains(errStr, "replacement transaction underpriced") ||
		strings.Contains(errStr, "already known")
}
