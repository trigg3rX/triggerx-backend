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
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
)

// NonceManager handles nonce allocation and transaction retry logic
type NonceManager struct {
	mu           sync.Mutex
	currentNonce uint64
	client       *ethclient.Client
	address      common.Address
	logger       observability.Logger
	lastSyncTime time.Time
	syncInterval time.Duration

	// Transaction tracking
	pendingTxs map[uint64]*PendingTransaction
	txMutex    sync.RWMutex

	// Retry configurations
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

// NewNonceManager creates a new nonce manager optimized for L2 chains
func NewNonceManager(client *ethclient.Client, logger observability.Logger) *NonceManager {
	return &NonceManager{
		client:       client,
		address:      common.HexToAddress(config.GetKeeperAddress()),
		logger:       logger,
		syncInterval: 10 * time.Second,
		pendingTxs:   make(map[uint64]*PendingTransaction),

		rpcRetryConfig: &retry.RetryConfig{
			MaxRetries:      8,
			InitialDelay:    200 * time.Millisecond,
			MaxDelay:        5 * time.Second,
			BackoffFactor:   1.5,
			JitterFactor:    0.3,
			ShouldRetry:     shouldRetryRPCError,
		},
		submitRetryConfig: &retry.RetryConfig{
			MaxRetries:      10,
			InitialDelay:    100 * time.Millisecond,
			MaxDelay:        3 * time.Second,
			BackoffFactor:   1.3,
			JitterFactor:    0.4,
			ShouldRetry:     shouldRetrySubmissionError,
		},
		confirmRetryConfig: &retry.RetryConfig{
			MaxRetries:      15,
			InitialDelay:    500 * time.Millisecond,
			MaxDelay:        8 * time.Second,
			BackoffFactor:   1.2,
			JitterFactor:    0.2,
			ShouldRetry:     shouldRetryConfirmationError,
		},
	}
}

// Initialize sets up the initial nonce from blockchain
func (nm *NonceManager) Initialize(ctx context.Context) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	return nm.syncWithBlockchain(ctx)
}

// GetNextNonce returns the next available nonce atomically
func (nm *NonceManager) GetNextNonce(ctx context.Context) (uint64, error) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if time.Since(nm.lastSyncTime) > nm.syncInterval {
		if err := nm.syncWithBlockchain(ctx); err != nil {
			return 0, fmt.Errorf("failed to sync nonce: %w", err)
		}
	}

	nonce := nm.currentNonce
	nm.currentNonce++
	// nm.logger.Debug(ctx, "Allocated nonce", observability.Uint64("nonce", nonce))
	return nonce, nil
}

// ReleaseNonce handles failed nonce allocation by either filling the gap or syncing
func (nm *NonceManager) ReleaseNonce(ctx context.Context, nonce uint64, privateKey *ecdsa.PrivateKey) {
	nm.mu.Lock()
	snapshotNonce := nm.currentNonce
	nm.mu.Unlock()

	if nonce+1 < snapshotNonce {
		// Higher nonces allocated - must fill the gap
		// nm.logger.Warn(ctx, "Filling nonce gap", observability.Uint64("nonce", nonce), observability.Uint64("snapshot_nonce", snapshotNonce-1))
		nm.fillNonceGap(ctx, nonce, privateKey)
	} else {
		// No higher nonces - safe to sync with blockchain
		// nm.logger.Debug(ctx, "Nonce released, syncing with blockchain", observability.Uint64("nonce", nonce))
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := nm.safeSyncFromBlockchain(ctx, snapshotNonce); err != nil {
			nm.logger.Error(ctx, "Failed to sync after release", observability.Error(err))
		}
	}
}

// fillNonceGap submits a self-transaction to fill a nonce gap
func (nm *NonceManager) fillNonceGap(ctx context.Context, nonce uint64, privateKey *ecdsa.PrivateKey) {
	if privateKey == nil {
		nm.logger.Error(ctx, "Cannot fill nonce gap", observability.Uint64("nonce", nonce))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	chainID, err := retry.Retry(ctx, func() (*big.Int, error) {
		return nm.client.ChainID(ctx)
	}, nm.rpcRetryConfig)
	if err != nil {
		nm.logger.Error(ctx, "Failed to get chain ID for gap fill", observability.Error(err))
		return
	}

	_, err = retry.Retry(ctx, func() (string, error) {
		return nm.submitGapFillTx(ctx, nonce, chainID, privateKey)
	}, nm.submitRetryConfig)

	if err != nil {
		nm.logger.Error(ctx, "Failed to fill nonce gap", observability.Uint64("nonce", nonce), observability.Error(err))
	}
}

// submitGapFillTx creates and submits a minimal self-transaction
func (nm *NonceManager) submitGapFillTx(ctx context.Context, nonce uint64, chainID *big.Int, privateKey *ecdsa.PrivateKey) (string, error) {
	gasPrice, err := nm.getGasPrice(ctx)
	if err != nil {
		return "", err
	}

	// Bump gas if replacing existing tx
	nm.txMutex.RLock()
	if existing, exists := nm.pendingTxs[nonce]; exists && existing.LastGasPrice != nil {
		minPrice := new(big.Int).Mul(existing.LastGasPrice, big.NewInt(120))
		minPrice.Div(minPrice, big.NewInt(100))
		if gasPrice.Cmp(minPrice) < 0 {
			gasPrice = minPrice
		}
	}
	nm.txMutex.RUnlock()

	// Minimal self-transaction: 0 ETH, 21k gas
	tx := types.NewTransaction(nonce, nm.address, big.NewInt(0), 21000, gasPrice, nil)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %w", err)
	}

	if err := nm.client.SendTransaction(ctx, signedTx); err != nil {
		if strings.Contains(err.Error(), "already known") {
			nm.trackPendingTx(nonce, signedTx.Hash().Hex(), chainID, privateKey, gasPrice)
			return signedTx.Hash().Hex(), nil
		}
		return "", err
	}

	nm.trackPendingTx(nonce, signedTx.Hash().Hex(), chainID, privateKey, gasPrice)
	return signedTx.Hash().Hex(), nil
}

// syncWithBlockchain updates nonce from blockchain (only increases)
func (nm *NonceManager) syncWithBlockchain(ctx context.Context) error {
	pendingNonce, err := retry.Retry(ctx, func() (uint64, error) {
		return nm.client.PendingNonceAt(ctx, nm.address)
	}, nm.rpcRetryConfig)
	if err != nil {
		return fmt.Errorf("failed to get pending nonce: %w", err)
	}

	if pendingNonce > nm.currentNonce {
		nm.currentNonce = pendingNonce
		// nm.logger.Debug(ctx, "Synced nonce", observability.Uint64("nonce", nm.currentNonce))
	}
	nm.lastSyncTime = time.Now()
	return nil
}

// safeSyncFromBlockchain syncs only if no new allocations since snapshot
func (nm *NonceManager) safeSyncFromBlockchain(ctx context.Context, snapshotNonce uint64) error {
	pendingNonce, err := retry.Retry(ctx, func() (uint64, error) {
		return nm.client.PendingNonceAt(ctx, nm.address)
	}, nm.rpcRetryConfig)
	if err != nil {
		return fmt.Errorf("failed to get pending nonce: %w", err)
	}

	nm.mu.Lock()
	defer nm.mu.Unlock()

	// Skip if other threads allocated nonces
	if nm.currentNonce > snapshotNonce {
		// nm.logger.Debug(ctx, "Skipping sync: nonces allocated", observability.Uint64("snapshot_nonce", snapshotNonce), observability.Uint64("current_nonce", nm.currentNonce))
		return nil
	}

	oldNonce := nm.currentNonce
	nm.currentNonce = pendingNonce
	nm.lastSyncTime = time.Now()

	if pendingNonce != oldNonce {
		nm.logger.Debug(ctx, "Safe synced nonce", observability.Uint64("old_nonce", oldNonce), observability.Uint64("pending_nonce", pendingNonce))
	}
	return nil
}

// SubmitTransaction submits a transaction with retry and replacement logic
func (nm *NonceManager) SubmitTransaction(
	ctx context.Context,
	nonce uint64,
	to common.Address,
	data []byte,
	chainID *big.Int,
	privateKey *ecdsa.PrivateKey,
) (*types.Receipt, string, error) {
	// Check for stuck transaction to replace
	nm.txMutex.RLock()
	existingTx, exists := nm.pendingTxs[nonce]
	nm.txMutex.RUnlock()

	if exists && existingTx.Status == "pending" && time.Since(existingTx.CreatedAt) > 30*time.Second {
		return nm.replaceTransaction(ctx, existingTx, data, to, chainID, privateKey)
	}

	return nm.submitNewTransaction(ctx, nonce, to, data, chainID, privateKey)
}

func (nm *NonceManager) submitNewTransaction(
	ctx context.Context,
	nonce uint64,
	to common.Address,
	data []byte,
	chainID *big.Int,
	privateKey *ecdsa.PrivateKey,
) (*types.Receipt, string, error) {
	gasPrice, err := nm.getGasPrice(ctx)
	if err != nil {
		return nil, "", err
	}

	tx := types.NewTransaction(nonce, to, big.NewInt(0), 600000, gasPrice, data)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return nil, "", fmt.Errorf("failed to sign: %w", err)
	}

	nm.trackTransaction(nonce, signedTx.Hash().Hex(), data, to, chainID, privateKey, gasPrice)
	return nm.submitWithRetry(ctx, signedTx, nonce, privateKey)
}

func (nm *NonceManager) replaceTransaction(
	ctx context.Context,
	existingTx *PendingTransaction,
	data []byte,
	to common.Address,
	chainID *big.Int,
	privateKey *ecdsa.PrivateKey,
) (*types.Receipt, string, error) {
	// nm.logger.Debug(ctx, "Replacing stuck tx", observability.Uint64("nonce", existingTx.Nonce))

	gasPrice, err := nm.getGasPrice(ctx)
	if err != nil {
		return nil, "", err
	}

	// 20% bump for replacement
	gasPrice = new(big.Int).Mul(gasPrice, big.NewInt(120))
	gasPrice.Div(gasPrice, big.NewInt(100))

	tx := types.NewTransaction(existingTx.Nonce, to, big.NewInt(0), 600000, gasPrice, data)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return nil, "", fmt.Errorf("failed to sign replacement: %w", err)
	}

	nm.updateTxStatus(existingTx.Nonce, signedTx.Hash().Hex(), gasPrice)
	return nm.submitWithRetry(ctx, signedTx, existingTx.Nonce, privateKey)
}

func (nm *NonceManager) submitWithRetry(ctx context.Context, signedTx *types.Transaction, nonce uint64, privateKey *ecdsa.PrivateKey) (*types.Receipt, string, error) {
	// Submit with retry
	txHash, err := retry.Retry(ctx, func() (string, error) {
		if err := nm.client.SendTransaction(ctx, signedTx); err != nil {
			if isNonceTooLowError(err) {
				nm.mu.Lock()
				_ = nm.syncWithBlockchain(ctx)
				nm.mu.Unlock()
			}
			return "", err
		}
		nm.logger.Debug(ctx, "Transaction sent", observability.String("tx_hash", signedTx.Hash().Hex()))
		return signedTx.Hash().Hex(), nil
	}, nm.submitRetryConfig)
	if err != nil {
		return nil, "", fmt.Errorf("submit failed: %w", err)
	}

	// Wait for confirmation with retry
	receipt, err := retry.Retry(ctx, func() (*types.Receipt, error) {
		confirmCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		receipt, err := bind.WaitMined(confirmCtx, nm.client, signedTx)
		if confirmCtx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("confirmation timeout")
		}
		return receipt, err
	}, nm.confirmRetryConfig)

	if err != nil {
		// Try replacement
		nm.logger.Warn(ctx, "Confirmation failed, replacing", observability.Error(err))
		replacementTx, replaceErr := nm.createReplacementTx(signedTx, 1, privateKey)
		if replaceErr != nil {
			return nil, "", fmt.Errorf("replacement failed: %w", replaceErr)
		}
		nm.updateTxStatus(nonce, replacementTx.Hash().Hex(), replacementTx.GasPrice())
		return nm.submitWithRetry(ctx, replacementTx, nonce, privateKey)
	}

	nm.markConfirmed(nonce, txHash)
	nm.logger.Debug(ctx, "Transaction confirmed", observability.String("tx_hash", txHash))
	return receipt, txHash, nil
}

func (nm *NonceManager) getGasPrice(ctx context.Context) (*big.Int, error) {
	gasPrice, err := retry.Retry(ctx, func() (*big.Int, error) {
		return nm.client.SuggestGasPrice(ctx)
	}, nm.rpcRetryConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	// 20% buffer
	gasPrice.Mul(gasPrice, big.NewInt(120))
	gasPrice.Div(gasPrice, big.NewInt(100))
	return gasPrice, nil
}

func (nm *NonceManager) createReplacementTx(originalTx *types.Transaction, attempt int, privateKey *ecdsa.PrivateKey) (*types.Transaction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	gasPrice, err := nm.getGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	// Increase by 20% per attempt
	multiplier := big.NewInt(int64(120 + (attempt * 20)))
	gasPrice.Mul(gasPrice, multiplier)
	gasPrice.Div(gasPrice, big.NewInt(100))

	tx := types.NewTransaction(originalTx.Nonce(), *originalTx.To(), big.NewInt(0), 600000, gasPrice, originalTx.Data())
	return types.SignTx(tx, types.NewEIP155Signer(originalTx.ChainId()), privateKey)
}

// Transaction tracking helpers
func (nm *NonceManager) trackTransaction(nonce uint64, txHash string, data []byte, to common.Address, chainID *big.Int, privateKey *ecdsa.PrivateKey, gasPrice *big.Int) {
	nm.txMutex.Lock()
	defer nm.txMutex.Unlock()
	nm.pendingTxs[nonce] = &PendingTransaction{
		Nonce: nonce, TxHash: txHash, CreatedAt: time.Now(), Status: "pending",
		Attempts: 1, LastGasPrice: gasPrice, Data: data, To: to, ChainID: chainID, PrivateKey: privateKey,
	}
}

func (nm *NonceManager) trackPendingTx(nonce uint64, txHash string, chainID *big.Int, privateKey *ecdsa.PrivateKey, gasPrice *big.Int) {
	nm.txMutex.Lock()
	defer nm.txMutex.Unlock()
	if existing, exists := nm.pendingTxs[nonce]; exists {
		existing.TxHash = txHash
		existing.LastGasPrice = gasPrice
		existing.Attempts++
	} else {
		nm.pendingTxs[nonce] = &PendingTransaction{
			Nonce: nonce, TxHash: txHash, CreatedAt: time.Now(), Status: "pending",
			Attempts: 1, LastGasPrice: gasPrice, To: nm.address, ChainID: chainID, PrivateKey: privateKey,
		}
	}
}

func (nm *NonceManager) updateTxStatus(nonce uint64, txHash string, gasPrice *big.Int) {
	nm.txMutex.Lock()
	defer nm.txMutex.Unlock()
	if tx, exists := nm.pendingTxs[nonce]; exists {
		tx.TxHash = txHash
		tx.Status = "pending"
		tx.Attempts++
		tx.LastGasPrice = gasPrice
	}
}

func (nm *NonceManager) markConfirmed(nonce uint64, txHash string) {
	nm.txMutex.Lock()
	defer nm.txMutex.Unlock()
	if tx, exists := nm.pendingTxs[nonce]; exists {
		tx.Status = "confirmed"
		tx.TxHash = txHash
	}
}

// Error retry predicates
func shouldRetryRPCError(err error, _ int) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "connection") || strings.Contains(s, "timeout") ||
		strings.Contains(s, "network") || strings.Contains(s, "dial") ||
		strings.Contains(s, "refused") || strings.Contains(s, "unavailable") ||
		strings.Contains(s, "rate limit") || strings.Contains(s, "429") ||
		strings.Contains(s, "500") || strings.Contains(s, "502") ||
		strings.Contains(s, "503") || strings.Contains(s, "504")
}

func shouldRetrySubmissionError(err error, attempt int) bool {
	if shouldRetryRPCError(err, attempt) {
		return true
	}
	s := strings.ToLower(err.Error())
	if attempt < 3 && (strings.Contains(s, "nonce too low") ||
		strings.Contains(s, "replacement transaction underpriced") ||
		strings.Contains(s, "already known")) {
		return true
	}
	return strings.Contains(s, "gas") && (strings.Contains(s, "estimate") || strings.Contains(s, "limit"))
}

func shouldRetryConfirmationError(err error, attempt int) bool {
	if shouldRetryRPCError(err, attempt) {
		return true
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "not found") || strings.Contains(s, "unknown transaction") ||
		strings.Contains(s, "timeout") || strings.Contains(s, "deadline exceeded")
}

func isNonceTooLowError(err error) bool {
	s := err.Error()
	return strings.Contains(s, "nonce too low") ||
		strings.Contains(s, "replacement transaction underpriced") ||
		strings.Contains(s, "already known")
}
