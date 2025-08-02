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
	"github.com/trigg3rX/triggerx-backend/internal/imua-keeper/config"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
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

	// Retry configuration
	maxRetries  int
	baseTimeout time.Duration
	priorityFee *big.Int // Base priority fee for EIP-1559
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
		syncInterval: 15 * time.Second, // More frequent sync for low latency
		maxRetries:   5,                // More retries for reliability
		baseTimeout:  3 * time.Second,  // Shorter timeout for faster retries
		priorityFee:  big.NewInt(2e9),  // 2 Gwei base priority fee
		pendingTxs:   make(map[uint64]*PendingTransaction),
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
	pendingNonce, err := nm.client.PendingNonceAt(ctx, nm.address)
	if err != nil {
		return fmt.Errorf("failed to get pending nonce: %w", err)
	}

	// Use the higher of pending nonce or our current nonce
	if pendingNonce > nm.currentNonce {
		nm.currentNonce = pendingNonce
		nm.logger.Infof("Synced nonce with blockchain: %d", nm.currentNonce)
	}

	nm.lastSyncTime = time.Now()
	return nil
}

// SubmitTransactionWithSmartRetry submits a transaction with intelligent retry logic
func (nm *NonceManager) SubmitTransactionWithSmartRetry(
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

	// Get current gas parameters
	gasPrice, maxFeePerGas, maxPriorityFeePerGas, err := nm.getOptimalGasParams(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get gas parameters: %w", err)
	}

	// Create transaction (prefer EIP-1559 if supported)
	var tx *types.Transaction
	var signedTx *types.Transaction
	var signErr error

	// Try EIP-1559 first if supported
	if maxFeePerGas != nil && maxPriorityFeePerGas != nil {
		nm.logger.Debugf("Attempting EIP-1559 transaction with nonce %d", nonce)
		tx = types.NewTx(&types.DynamicFeeTx{
			ChainID:   chainID,
			Nonce:     nonce,
			GasTipCap: maxPriorityFeePerGas,
			GasFeeCap: maxFeePerGas,
			Gas:       300000,
			To:        &to,
			Value:     big.NewInt(0),
			Data:      data,
		})

		// Try to sign EIP-1559 transaction
		signedTx, signErr = types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
		if signErr != nil {
			nm.logger.Warnf("EIP-1559 transaction signing failed, falling back to legacy: %v", signErr)
		}
	}

	// Fallback to legacy transaction if EIP-1559 failed or not supported
	if signedTx == nil || signErr != nil {
		nm.logger.Debugf("Using legacy transaction with nonce %d", nonce)
		if gasPrice == nil {
			return nil, "", fmt.Errorf("gas price is required for legacy transaction")
		}

		tx = types.NewTransaction(nonce, to, big.NewInt(0), 300000, gasPrice, data)
		signedTx, signErr = types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
		if signErr != nil {
			return nil, "", fmt.Errorf("failed to sign legacy transaction: %w", signErr)
		}
	}

	// Track the transaction
	nm.trackTransaction(nonce, signedTx.Hash().Hex(), data, to, chainID, privateKey, gasPrice)

	// Submit with retry logic
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

	// Get higher gas parameters for replacement
	gasPrice, maxFeePerGas, maxPriorityFeePerGas, err := nm.getOptimalGasParams(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get gas parameters: %w", err)
	}

	// Increase fees by 20% for replacement
	if gasPrice != nil {
		gasPrice = new(big.Int).Mul(gasPrice, big.NewInt(120))
		gasPrice = new(big.Int).Div(gasPrice, big.NewInt(100))
	}
	if maxFeePerGas != nil {
		maxFeePerGas = new(big.Int).Mul(maxFeePerGas, big.NewInt(120))
		maxFeePerGas = new(big.Int).Div(maxFeePerGas, big.NewInt(100))
	}
	if maxPriorityFeePerGas != nil {
		maxPriorityFeePerGas = new(big.Int).Mul(maxPriorityFeePerGas, big.NewInt(120))
		maxPriorityFeePerGas = new(big.Int).Div(maxPriorityFeePerGas, big.NewInt(100))
	}

	// Create replacement transaction with fallback logic
	var tx *types.Transaction
	var signedTx *types.Transaction
	var signErr error

	// Try EIP-1559 first if supported
	if maxFeePerGas != nil && maxPriorityFeePerGas != nil {
		nm.logger.Debugf("Attempting EIP-1559 replacement transaction with nonce %d", existingTx.Nonce)
		tx = types.NewTx(&types.DynamicFeeTx{
			ChainID:   chainID,
			Nonce:     existingTx.Nonce,
			GasTipCap: maxPriorityFeePerGas,
			GasFeeCap: maxFeePerGas,
			Gas:       300000,
			To:        &to,
			Value:     big.NewInt(0),
			Data:      data,
		})

		// Try to sign EIP-1559 transaction
		signedTx, signErr = types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
		if signErr != nil {
			nm.logger.Warnf("EIP-1559 replacement transaction signing failed, falling back to legacy: %v", signErr)
		}
	}

	// Fallback to legacy transaction if EIP-1559 failed or not supported
	if signedTx == nil || signErr != nil {
		nm.logger.Debugf("Using legacy replacement transaction with nonce %d", existingTx.Nonce)
		if gasPrice == nil {
			return nil, "", fmt.Errorf("gas price is required for legacy replacement transaction")
		}

		tx = types.NewTransaction(existingTx.Nonce, to, big.NewInt(0), 300000, gasPrice, data)
		signedTx, signErr = types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
		if signErr != nil {
			return nil, "", fmt.Errorf("failed to sign legacy replacement transaction: %w", signErr)
		}
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

	for attempt := 0; attempt < nm.maxRetries; attempt++ {
		// Send transaction
		err := nm.client.SendTransaction(ctx, signedTx)
		if err != nil {
			nm.logger.Warnf("Failed to send transaction (attempt %d): %v", attempt+1, err)

			// Check if it's a nonce too low error - this means we need to sync
			if isNonceTooLowError(err) {
				if syncErr := nm.syncWithBlockchain(ctx); syncErr != nil {
					return nil, "", fmt.Errorf("failed to sync after nonce error: %w", syncErr)
				}
			}

			if attempt == nm.maxRetries-1 {
				return nil, "", fmt.Errorf("failed to send transaction after %d attempts: %v", nm.maxRetries, err)
			}
			continue
		}

		txHash := signedTx.Hash().Hex()
		nm.logger.Infof("Transaction sent (attempt %d): %s", attempt+1, txHash)

		// Wait for confirmation with exponential backoff
		timeout := nm.baseTimeout * time.Duration(1<<attempt) // Exponential backoff
		ctx, cancel := context.WithTimeout(ctx, timeout)
		receipt, err := bind.WaitMined(ctx, nm.client, signedTx)
		cancel()

		if err == nil {
			nm.markTransactionConfirmed(nonce, txHash)
			nm.logger.Infof("Transaction confirmed: %s", txHash)
			return receipt, txHash, nil
		}

		// Handle timeout - create replacement with higher fees
		if ctx.Err() == context.DeadlineExceeded {
			nm.logger.Warnf("Transaction %s timed out after %v, creating replacement", txHash, timeout)

			// Create replacement transaction with higher fees
			replacementTx, err := nm.createReplacementTransaction(signedTx, attempt+1, privateKey)
			if err != nil {
				return nil, "", fmt.Errorf("failed to create replacement transaction: %w", err)
			}
			signedTx = replacementTx
			continue
		}

		// Other error occurred
		nm.logger.Warnf("Error waiting for transaction %s: %v", txHash, err)
		if attempt == nm.maxRetries-1 {
			return nil, "", fmt.Errorf("transaction failed after %d attempts: %v", nm.maxRetries, err)
		}
	}

	return nil, "", fmt.Errorf("transaction failed after %d attempts", nm.maxRetries)
}

// getOptimalGasParams gets optimal gas parameters for current network conditions
func (nm *NonceManager) getOptimalGasParams(ctx context.Context) (*big.Int, *big.Int, *big.Int, error) {

	// Always get legacy gas price as fallback
	gasPrice, err := nm.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	// Add 20% buffer for network congestion
	gasPrice.Mul(gasPrice, big.NewInt(120))
	gasPrice.Div(gasPrice, big.NewInt(100))

	// Try to get EIP-1559 parameters
	head, err := nm.client.HeaderByNumber(ctx, nil)
	if err != nil {
		nm.logger.Warnf("Failed to get latest header for EIP-1559 detection: %v", err)
		return gasPrice, nil, nil, nil
	}

	// Check if EIP-1559 is supported and base fee is reasonable
	if head.BaseFee != nil && head.BaseFee.Cmp(big.NewInt(0)) > 0 {
		nm.logger.Debugf("EIP-1559 detected with base fee: %s", head.BaseFee.String())

		// Calculate optimal EIP-1559 parameters
		baseFee := head.BaseFee
		maxPriorityFeePerGas := new(big.Int).Set(nm.priorityFee)

		// Calculate max fee per gas (base fee + 2x priority fee for safety)
		maxFeePerGas := new(big.Int).Mul(maxPriorityFeePerGas, big.NewInt(2))
		maxFeePerGas.Add(maxFeePerGas, baseFee)

		// Add 20% buffer for network congestion
		maxFeePerGas.Mul(maxFeePerGas, big.NewInt(120))
		maxFeePerGas.Div(maxFeePerGas, big.NewInt(100))

		return gasPrice, maxFeePerGas, maxPriorityFeePerGas, nil
	}

	nm.logger.Debugf("EIP-1559 not supported, using legacy gas price: %s", gasPrice.String())
	return gasPrice, nil, nil, nil
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
	case 2: // EIP-1559 transaction
		to = *originalTx.To()
		data = originalTx.Data()
		chainID = originalTx.ChainId()
	default:
		return nil, fmt.Errorf("unsupported transaction type: %d", tx)
	}

	// Get higher gas parameters
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	gasPrice, maxFeePerGas, maxPriorityFeePerGas, err := nm.getOptimalGasParams(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas parameters for replacement: %w", err)
	}

	// Increase fees by 20% for each attempt
	feeMultiplier := big.NewInt(int64(120 + (attempt * 20))) // 120%, 140%, 160%, etc.
	feeDivisor := big.NewInt(100)

	if maxFeePerGas != nil && maxPriorityFeePerGas != nil {
		// EIP-1559 replacement
		maxFeePerGas.Mul(maxFeePerGas, feeMultiplier)
		maxFeePerGas.Div(maxFeePerGas, feeDivisor)
		maxPriorityFeePerGas.Mul(maxPriorityFeePerGas, feeMultiplier)
		maxPriorityFeePerGas.Div(maxPriorityFeePerGas, feeDivisor)

		tx := types.NewTx(&types.DynamicFeeTx{
			ChainID:   chainID,
			Nonce:     originalTx.Nonce(),
			GasTipCap: maxPriorityFeePerGas,
			GasFeeCap: maxFeePerGas,
			Gas:       300000,
			To:        &to,
			Value:     big.NewInt(0),
			Data:      data,
		})

		// Test if we can sign this transaction type
		_, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
		if err == nil {
			return tx, nil
		}
		nm.logger.Warnf("EIP-1559 replacement transaction signing failed, falling back to legacy: %v", err)
	}

	// Legacy replacement
	if gasPrice == nil {
		return nil, fmt.Errorf("gas price is required for legacy replacement transaction")
	}

	gasPrice.Mul(gasPrice, feeMultiplier)
	gasPrice.Div(gasPrice, feeDivisor)

	return types.NewTransaction(originalTx.Nonce(), to, big.NewInt(0), 300000, gasPrice, data), nil
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

// Utility functions
func isNonceTooLowError(err error) bool {
	errStr := err.Error()
	return strings.Contains(errStr, "nonce too low") ||
		strings.Contains(errStr, "replacement transaction underpriced") ||
		strings.Contains(errStr, "already known")
}
