// Package main demonstrates the nonce gap issue in concurrent transaction scenarios.
// This test uses the ACTUAL NonceManager from internal/keeper/core/execution
// to test real-world nonce gap scenarios and recovery.
//
// Run: go run scripts/tests/nonce_gap_demo.go
//
// Prerequisites:
//   - Set TEST_PRIVATE_KEY env variable (account with testnet ETH)
//   - Set TEST_RPC_URL env variable (e.g., Sepolia RPC)

package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	// Import the ACTUAL NonceManager from the execution package
	"github.com/trigg3rX/triggerx-backend/internal/keeper/core/execution"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// ============================================================================
// Test Configuration
// ============================================================================

const (
	maxRetries    = 3                      // Number of retry attempts for simulated failures
	initialDelay  = 500 * time.Millisecond // Initial delay between retries
	backoffFactor = 1.5                    // Exponential backoff multiplier
)

// ============================================================================
// Transaction Result Tracking
// ============================================================================

type TransactionResult struct {
	TaskID      int
	Nonce       uint64
	TxHash      string
	Success     bool
	Error       error
	StartTime   time.Time
	EndTime     time.Time
	BlockNumber uint64
}

// ============================================================================
// Main Test Runner
// ============================================================================

func main() {
	// Initialize logger using the correct logging package function
	logConfig := logging.LoggerConfig{
		ProcessName:   logging.TestProcess,
		IsDevelopment: true,
	}
	logger, err := logging.NewZapLogger(logConfig)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// Get configuration from environment
	privateKeyHex := os.Getenv("TEST_PRIVATE_KEY")
	rpcURL := os.Getenv("TEST_RPC_URL")

	if privateKeyHex == "" || rpcURL == "" {
		logger.Fatal("Please set TEST_PRIVATE_KEY and TEST_RPC_URL environment variables")
	}

	// Remove 0x prefix if present
	privateKeyHex = strings.TrimPrefix(privateKeyHex, "0x")

	// Parse private key
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		logger.Fatalf("Failed to parse private key: %v", err)
	}

	// Derive address
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		logger.Fatal("Failed to get public key")
	}
	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	logger.Infof("Test address: %s", address.Hex())
	logger.Infof("RPC URL: %s", rpcURL)

	// Connect to Ethereum
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		logger.Fatalf("Failed to connect to Ethereum: %v", err)
	}
	defer client.Close()

	// Get chain ID
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		logger.Fatalf("Failed to get chain ID: %v", err)
	}
	logger.Infof("Chain ID: %s", chainID.String())

	// Check balance
	balance, err := client.BalanceAt(context.Background(), address, nil)
	if err != nil {
		logger.Fatalf("Failed to get balance: %v", err)
	}
	ethBalance := new(big.Float).Quo(new(big.Float).SetInt(balance), big.NewFloat(1e18))
	logger.Infof("Balance: %s ETH", ethBalance.Text('f', 6))

	if balance.Cmp(big.NewInt(1e16)) < 0 { // Less than 0.01 ETH
		logger.Fatal("Insufficient balance for test (need at least 0.01 ETH)")
	}

	// Create the ACTUAL NonceManager from the execution package with TEST address
	// Note: In production, config.GetPrivateKeyController() provides the key for gap filling
	// In this test, we'll handle gap filling manually since config isn't set up
	nonceManager := execution.NewNonceManager(client, logger)

	if err := nonceManager.Initialize(context.Background()); err != nil {
		logger.Fatalf("Failed to initialize nonce manager: %v", err)
	}

	// Get starting nonce
	startNonce, err := nonceManager.GetNextNonce(context.Background())
	if err != nil {
		logger.Fatalf("Failed to get initial nonce: %v", err)
	}
	// Return the nonce since we just wanted to check
	nonceManager.ReleaseNonce(startNonce, privateKey)
	logger.Infof("Starting nonce: %d", startNonce)

	// =========================================================================
	// SCENARIO: Test Case B - Transaction 6 fails to submit
	// =========================================================================
	//
	// We'll allocate nonces 5, 6, 7, 8 (relative to current)
	// Nonce 6 (relative, i.e., startNonce + 1) will fail
	//
	// Expected behavior:
	// - Transaction 5 (nonce=startNonce): SUCCESS
	// - Transaction 6 (nonce=startNonce+1): FAIL TO SUBMIT (simulated)
	// - Transaction 7 (nonce=startNonce+2): STUCK waiting for nonce+1
	// - Transaction 8 (nonce=startNonce+3): STUCK waiting for nonce+1
	// =========================================================================

	logger.Info("")
	logger.Info("╔══════════════════════════════════════════════════════════════════╗")
	logger.Info("║           NONCE GAP TEST - CASE B: SUBMISSION FAILURE            ║")
	logger.Info("╠══════════════════════════════════════════════════════════════════╣")
	logger.Infof("║  Nonce %d (Task 5): Should SUCCEED                              ║", startNonce)
	logger.Infof("║  Nonce %d (Task 6): Will FAIL TO SUBMIT (simulated)             ║", startNonce+1)
	logger.Infof("║  Nonce %d (Task 7): Will be STUCK                               ║", startNonce+2)
	logger.Infof("║  Nonce %d (Task 8): Will be STUCK                               ║", startNonce+3)
	logger.Info("╚══════════════════════════════════════════════════════════════════╝")
	logger.Info("")

	// The nonce to fail
	failNonce := startNonce + 1
	logger.Infof("🎯 Will simulate failure for nonce %d", failNonce)

	// Track failed nonces for manual gap filling
	failedNonces := make(chan uint64, 10)

	// Results collector
	results := make(chan TransactionResult, 4)
	var wg sync.WaitGroup

	// Simulate 4 concurrent task executions
	for taskID := 5; taskID <= 8; taskID++ {
		wg.Add(1)
		go func(taskID int) {
			defer wg.Done()

			result := TransactionResult{
				TaskID:    taskID,
				StartTime: time.Now(),
			}

			logger.Infof("[Task-%d] Starting task execution...", taskID)

			// Add small delay between starting tasks to ensure order
			time.Sleep(time.Duration((taskID-5)*100) * time.Millisecond)

			// Step 1: Allocate nonce using the REAL NonceManager
			nonce, err := nonceManager.GetNextNonce(context.Background())
			if err != nil {
				logger.Errorf("[Task-%d] Failed to allocate nonce: %v", taskID, err)
				result.Error = err
				result.EndTime = time.Now()
				results <- result
				return
			}
			result.Nonce = nonce
			logger.Infof("[Task-%d] Allocated nonce: %d", taskID, nonce)

			// Step 2: Submit transaction
			ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
			defer cancel()

			// Calculate amount: taskID * 1e12 wei (Task 5 = 1e12, Task 6 = 2e12, etc.)
			amount := big.NewInt(int64(taskID-4) * 1_000_000_000_000)
			logger.Infof("[Task-%d] Sending %s wei", taskID, amount.String())

			// Check if this is the nonce to fail
			if nonce == failNonce {
				// Simulate network failure with retries
				txHash, submitErr := simulateFailedSubmission(ctx, nonce, logger)
				result.EndTime = time.Now()

				if submitErr != nil {
					logger.Errorf("[Task-%d] Transaction failed: %v", taskID, submitErr)
					result.Error = submitErr
					result.Success = false

					// Call REAL NonceManager.ReleaseNonce
					// This will trigger gap filling using the provided private key
					nonceManager.ReleaseNonce(nonce, privateKey)
					logger.Infof("[Task-%d] Released nonce %d after failure", taskID, nonce)

					// Signal that this nonce needs gap filling (test-only backup)
					failedNonces <- nonce
				} else {
					result.TxHash = txHash
					result.Success = true
				}
			} else {
				// Normal transaction submission
				receipt, txHash, submitErr := submitETHTransfer(ctx, client, nonce, amount, chainID, privateKey, address, logger)
				result.EndTime = time.Now()

				if submitErr != nil {
					logger.Errorf("[Task-%d] Transaction failed: %v", taskID, submitErr)
					result.Error = submitErr
					result.Success = false

					// Call REAL NonceManager.ReleaseNonce
					nonceManager.ReleaseNonce(nonce, privateKey)
					logger.Infof("[Task-%d] Released nonce %d after failure", taskID, nonce)
				} else {
					result.TxHash = txHash
					result.Success = true
					result.BlockNumber = receipt.BlockNumber.Uint64()
					logger.Infof("[Task-%d] Transaction completed: %s (block: %d)", taskID, txHash, result.BlockNumber)
				}
			}

			results <- result
		}(taskID)
	}

	// Gap filling goroutine for test environment
	// In production, NonceManager.fillNonceGap() handles this via config.GetPrivateKeyController()
	// In this test, we manually fill gaps since config isn't set up
	go func() {
		for gapNonce := range failedNonces {
			logger.Infof("🔧 [TEST] Filling nonce gap with self-transaction for nonce %d", gapNonce)

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

			// Create minimal self-transaction
			gasPrice, err := client.SuggestGasPrice(ctx)
			if err != nil {
				logger.Errorf("[TEST] Failed to get gas price for gap fill: %v", err)
				cancel()
				continue
			}

			tx := types.NewTransaction(
				gapNonce,
				address,       // Send to self
				big.NewInt(0), // 0 ETH
				21000,         // Minimum gas
				gasPrice,
				nil,
			)

			signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
			if err != nil {
				logger.Errorf("[TEST] Failed to sign gap-fill transaction: %v", err)
				cancel()
				continue
			}

			err = client.SendTransaction(ctx, signedTx)
			if err != nil {
				logger.Errorf("[TEST] Failed to submit gap-fill transaction: %v", err)
				cancel()
				continue
			}

			logger.Infof("✅ [TEST] Gap-fill transaction submitted: %s (nonce: %d)", signedTx.Hash().Hex(), gapNonce)
			cancel()
		}
	}()

	// Wait with timeout for all tasks
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
		close(failedNonces) // Stop the gap-filling goroutine
	}()

	// Wait for completion or timeout
	timeout := time.After(200 * time.Second)

	logger.Info("")
	logger.Info("Waiting for transactions to complete (timeout: 200s)...")
	logger.Info("")

	select {
	case <-done:
		logger.Info("All tasks completed")
	case <-timeout:
		logger.Warn("Test timed out - some transactions are likely stuck!")
	}

	// Collect and display results
	close(results)

	logger.Info("")
	logger.Info("╔══════════════════════════════════════════════════════════════════╗")
	logger.Info("║                         TEST RESULTS                              ║")
	logger.Info("╚══════════════════════════════════════════════════════════════════╝")

	allResults := make([]TransactionResult, 0)
	for result := range results {
		allResults = append(allResults, result)
	}

	// Sort by task ID
	for taskID := 5; taskID <= 8; taskID++ {
		for _, r := range allResults {
			if r.TaskID == taskID {
				duration := r.EndTime.Sub(r.StartTime)
				if r.Success {
					logger.Infof("✅ Task %d: SUCCESS | Nonce: %d | TxHash: %s | Block: %d | Duration: %v",
						r.TaskID, r.Nonce, r.TxHash[:16]+"...", r.BlockNumber, duration.Round(time.Millisecond))
				} else {
					logger.Errorf("❌ Task %d: FAILED  | Nonce: %d | Error: %v | Duration: %v",
						r.TaskID, r.Nonce, r.Error, duration.Round(time.Millisecond))
				}
				break
			}
		}
	}

	// Show nonce manager state
	logger.Info("")
	logger.Info("╔══════════════════════════════════════════════════════════════════╗")
	logger.Info("║                    NONCE STATE COMPARISON                        ║")
	logger.Info("╚══════════════════════════════════════════════════════════════════╝")

	logger.Infof("Starting nonce: %d", startNonce)

	// Get blockchain nonce for comparison
	blockchainNonce, err := client.PendingNonceAt(context.Background(), address)
	if err != nil {
		logger.Errorf("Failed to get blockchain nonce: %v", err)
	} else {
		logger.Infof("Blockchain pending nonce: %d", blockchainNonce)

		expectedNonce := startNonce + 4 // We allocated 4 nonces
		if blockchainNonce < expectedNonce {
			gapSize := expectedNonce - blockchainNonce
			logger.Warnf("⚠️  NONCE GAP DETECTED! Expected: %d, Blockchain: %d, Gap: %d",
				expectedNonce, blockchainNonce, gapSize)
		} else {
			logger.Infof("✅ No nonce gap - blockchain caught up to expected nonce")
		}
	}

	logger.Info("")
	logger.Info("╔══════════════════════════════════════════════════════════════════╗")
	logger.Info("║                         ANALYSIS                                  ║")
	logger.Info("╚══════════════════════════════════════════════════════════════════╝")
	logger.Info("")
	logger.Info("This test uses the ACTUAL NonceManager from:")
	logger.Info("  internal/keeper/core/execution/nonce_manager.go")
	logger.Info("")
	logger.Info("Key behaviors tested:")
	logger.Info("1. GetNextNonce() - atomically allocates nonces")
	logger.Info("2. ReleaseNonce() - releases unused nonces and fills gaps")
	logger.Info("3. fillNonceGap() - submits self-tx when higher nonces are pending")
	logger.Info("")
}

// ============================================================================
// Helper Functions
// ============================================================================

// simulateFailedSubmission simulates a network failure with retries
func simulateFailedSubmission(ctx context.Context, nonce uint64, logger logging.Logger) (string, error) {
	delay := initialDelay

	for attempt := 1; attempt <= maxRetries; attempt++ {
		logger.Warnf("🔄 Attempt %d/%d for nonce %d...", attempt, maxRetries, nonce)

		// Simulate network failure
		logger.Errorf("💥 Attempt %d FAILED for nonce %d: simulated network failure", attempt, nonce)

		if attempt < maxRetries {
			logger.Infof("⏳ Waiting %v before retry...", delay)
			select {
			case <-time.After(delay):
				delay = time.Duration(float64(delay) * backoffFactor)
			case <-ctx.Done():
				return "", ctx.Err()
			}
		}
	}

	logger.Errorf("❌ All %d retries exhausted for nonce %d", maxRetries, nonce)
	return "", fmt.Errorf("failed after %d retries: simulated network failure", maxRetries)
}

// submitETHTransfer submits an ETH transfer transaction
func submitETHTransfer(
	ctx context.Context,
	client *ethclient.Client,
	nonce uint64,
	amount *big.Int,
	chainID *big.Int,
	privateKey *ecdsa.PrivateKey,
	fromAddress common.Address,
	logger logging.Logger,
) (*types.Receipt, string, error) {

	// Destination address (can be any address for testing)
	toAddress := common.HexToAddress("0xE20727548Dcb92f5484AE647CdDbB8DeA2a68aB3")

	// Get gas price
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get gas price: %w", err)
	}

	// Add 30% buffer
	gasPrice = new(big.Int).Mul(gasPrice, big.NewInt(130))
	gasPrice = new(big.Int).Div(gasPrice, big.NewInt(100))

	// Create transaction
	tx := types.NewTransaction(nonce, toAddress, amount, 21000, gasPrice, nil)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return nil, "", fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send transaction
	err = client.SendTransaction(ctx, signedTx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to send transaction: %w", err)
	}

	txHash := signedTx.Hash().Hex()
	logger.Infof("✅ Transaction sent: %s (nonce: %d)", txHash, nonce)

	// Wait for confirmation
	logger.Infof("⏳ Waiting for confirmation of nonce %d...", nonce)

	confirmCtx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	receipt, err := bind.WaitMined(confirmCtx, client, signedTx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to confirm transaction: %w", err)
	}

	if receipt.Status == types.ReceiptStatusSuccessful {
		logger.Infof("✅ Transaction confirmed: %s (nonce: %d, block: %d)", txHash, nonce, receipt.BlockNumber.Uint64())
	} else {
		logger.Warnf("⚠️ Transaction reverted: %s (nonce: %d)", txHash, nonce)
	}

	return receipt, txHash, nil
}
