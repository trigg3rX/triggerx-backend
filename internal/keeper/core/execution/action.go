package execution

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/utils"
	"github.com/trigg3rX/triggerx-backend/pkg/docker"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (e *TaskExecutor) executeAction(targetData *types.TaskTargetData, triggerData *types.TaskTriggerData, nonce uint64, client *ethclient.Client) (types.PerformerActionData, error) {
	if targetData.TargetContractAddress == "" {
		e.logger.Errorf("Execution contract address not configured")
		return types.PerformerActionData{}, fmt.Errorf("execution contract address not configured")
	}

	var timeToNextTrigger time.Duration
	switch targetData.TaskDefinitionID {
	case 1:
		timeToNextTrigger = time.Until(triggerData.NextTriggerTimestamp)
		timeToNextTrigger = timeToNextTrigger - 2*time.Second
	case 2:
		timeToNextTrigger = time.Until(triggerData.NextTriggerTimestamp)
		timeToNextTrigger = timeToNextTrigger - 10*time.Second
		if timeToNextTrigger < 0 {
			timeToNextTrigger = 0
		}
	default:
		timeToNextTrigger = 0
	}
	time.Sleep(timeToNextTrigger)

	taregtContractAddress := ethcommon.HexToAddress(targetData.TargetContractAddress)
	contractABI, method, err := e.getContractMethodAndABI(targetData.TargetFunction, targetData)
	if err != nil {
		return types.PerformerActionData{}, fmt.Errorf("failed to get contract method and ABI: %v", err)
	}

	var argData []interface{}
	var result *docker.ExecutionResult
	switch targetData.TaskDefinitionID {
	case 2, 4, 6:
		codePath, err := e.codeExecutor.Downloader.DownloadFile(context.Background(), targetData.DynamicArgumentsScriptUrl, e.logger)
		if err != nil {
			return types.PerformerActionData{}, fmt.Errorf("failed to download dynamic arguments script: %v", err)
		}
		defer func() { _ = os.RemoveAll(filepath.Dir(codePath)) }()

		containerID, err := e.codeExecutor.DockerManager.CreateContainer(context.Background(), codePath)
		if err != nil {
			return types.PerformerActionData{}, fmt.Errorf("failed to create container: %v", err)
		}
		defer func() {
			if err := e.codeExecutor.DockerManager.CleanupContainer(context.Background(), containerID); err != nil {
				e.logger.Errorf("failed to cleanup container %s: %v", containerID, err)
			}
		}()
		metrics.DockerContainersCreatedTotal.WithLabelValues("golang").Inc()

		start := time.Now()
		result, err := e.codeExecutor.MonitorExecution(context.Background(), e.codeExecutor.DockerManager.Cli, containerID, 1)
		if err != nil {
			return types.PerformerActionData{}, fmt.Errorf("failed to monitor execution: %v", err)
		}

		if !result.Success {
			return types.PerformerActionData{}, fmt.Errorf("failed to execute dynamic arguments script: %v", result.Error)
		}
		metrics.DockerContainerDurationSeconds.WithLabelValues("golang").Set(time.Since(start).Seconds())

		argData = e.parseDynamicArgs(result.Output)
	case 1, 3, 5:
		argData = e.parseStaticArgs(targetData.Arguments)
		result = &docker.ExecutionResult{
			Stats: docker.ResourceStats{
				TotalCost: 0.1,
			},
		}
	default:
		return types.PerformerActionData{}, fmt.Errorf("unsupported task definition id: %d", targetData.TaskDefinitionID)
	}

	// Handle args as potentially structured data
	convertedArgs, err := e.processArguments(argData, method.Inputs, contractABI)
	if err != nil {
		return types.PerformerActionData{}, fmt.Errorf("error processing arguments: %v", err)
	}

	// Pack the target contract's function call data
	var callData []byte
	callData, err = contractABI.Pack(method.Name, convertedArgs...)
	if err != nil {
		e.logger.Warnf("Error packing arguments: %v", err)
		return types.PerformerActionData{}, fmt.Errorf("error packing arguments: %v", err)
	}

	// Create transaction data for execution contract
	privateKey, err := crypto.HexToECDSA(config.GetPrivateKeyController())
	if err != nil {
		return types.PerformerActionData{}, fmt.Errorf("failed to parse private key: %v", err)
	}
	e.logger.Debugf("Using nonce: %d", nonce)

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return types.PerformerActionData{}, err
	}

	// Pack the execution contract's executeFunction call
	executionABI, err := abi.JSON(strings.NewReader(`[{"inputs":[{"name":"target","type":"address"},{"name":"data","type":"bytes"}],"name":"executeFunction","outputs":[],"stateMutability":"payable","type":"function"}]`))
	if err != nil {
		return types.PerformerActionData{}, fmt.Errorf("failed to parse execution contract ABI: %v", err)
	}

	executionInput, err := executionABI.Pack("executeFunction", taregtContractAddress, callData)
	if err != nil {
		return types.PerformerActionData{}, fmt.Errorf("failed to pack execution contract input: %v", err)
	}

	executionContractAddress := utils.GetProxyHubAddress(targetData.TargetChainID)
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return types.PerformerActionData{}, fmt.Errorf("failed to get chain ID: %v", err)
	}

	// Create and sign transaction with retry mechanism
	receipt, finalTxHash, err := e.submitTransactionWithRetry(
		client,
		privateKey,
		nonce,
		ethcommon.HexToAddress(executionContractAddress),
		executionInput,
		chainID,
		gasPrice,
	)
	if err != nil {
		return types.PerformerActionData{}, err
	}

	executionResult := types.PerformerActionData{
		TaskID:             targetData.TaskID,
		ActionTxHash:       finalTxHash,
		GasUsed:            strconv.FormatUint(receipt.GasUsed, 10),
		Status:             receipt.Status == ethtypes.ReceiptStatusSuccessful,
		MemoryUsage:        result.Stats.MemoryUsage,
		CPUPercentage:      result.Stats.CPUPercentage,
		NetworkRx:          result.Stats.RxBytes,
		NetworkTx:          result.Stats.TxBytes,
		BlockRead:          result.Stats.BlockRead,
		BlockWrite:         result.Stats.BlockWrite,
		BandwidthRate:      result.Stats.BandwidthRate,
		TotalFee:           result.Stats.TotalCost,
		StaticComplexity:   result.Stats.StaticComplexity,
		DynamicComplexity:  result.Stats.DynamicComplexity,
		ExecutionTimestamp: time.Now().UTC(),
	}
	metrics.TransactionsSentTotal.WithLabelValues(targetData.TargetChainID, "success").Inc()
	metrics.GasUsedTotal.WithLabelValues(targetData.TargetChainID).Add(float64(receipt.GasUsed))
	metrics.TransactionFeesTotal.WithLabelValues(targetData.TargetChainID).Add(result.Stats.TotalCost)

	e.logger.Infof("Task ID %d executed successfully. Transaction: %s", targetData.TaskID, finalTxHash)

	return executionResult, nil
}

// submitTransactionWithRetry handles transaction submission with timeout and fee bumping
func (e *TaskExecutor) submitTransactionWithRetry(
	client *ethclient.Client,
	privateKey *ecdsa.PrivateKey,
	nonce uint64,
	to ethcommon.Address,
	data []byte,
	chainID *big.Int,
	initialGasPrice *big.Int,
) (*ethtypes.Receipt, string, error) {
	const (
		txTimeout     = 5 * time.Second // Wait 5 seconds before resubmitting
		maxRetries    = 3               // Maximum number of retries
		feeBumpFactor = 1.2             // Increase fees by 20% on each retry
	)

	currentGasPrice := new(big.Int).Set(initialGasPrice)
	var lastTxHash string

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Create and sign transaction
		tx := ethtypes.NewTransaction(nonce, to, big.NewInt(0), 300000, currentGasPrice, data)
		signedTx, err := ethtypes.SignTx(tx, ethtypes.NewEIP155Signer(chainID), privateKey)
		if err != nil {
			return nil, "", fmt.Errorf("failed to sign transaction: %v", err)
		}

		// Send transaction
		err = client.SendTransaction(context.Background(), signedTx)
		if err != nil {
			e.logger.Warnf("Failed to send transaction (attempt %d): %v", attempt+1, err)
			if attempt == maxRetries-1 {
				return nil, "", fmt.Errorf("failed to send transaction after %d attempts: %v", maxRetries, err)
			}
			continue
		}

		lastTxHash = signedTx.Hash().Hex()
		e.logger.Infof("Transaction sent (attempt %d): %s with gas price: %s",
			attempt+1, lastTxHash, currentGasPrice.String())

		// Wait for transaction with timeout
		ctx, cancel := context.WithTimeout(context.Background(), txTimeout)
		receipt, err := bind.WaitMined(ctx, client, signedTx)
		cancel()

		if err == nil {
			// Transaction was mined successfully
			e.logger.Infof("Transaction confirmed: %s", lastTxHash)
			return receipt, lastTxHash, nil
		}

		// Check if it's a timeout or other error
		if ctx.Err() == context.DeadlineExceeded {
			e.logger.Warnf("Transaction %s timed out after %v, attempting resubmission with higher fees",
				lastTxHash, txTimeout)

			// Increase gas price for next attempt
			currentGasPrice = new(big.Int).Mul(currentGasPrice, big.NewInt(int64(feeBumpFactor*100)))
			currentGasPrice = new(big.Int).Div(currentGasPrice, big.NewInt(100))

			continue
		}

		// Other error occurred
		e.logger.Warnf("Error waiting for transaction %s: %v", lastTxHash, err)
		if attempt == maxRetries-1 {
			return nil, "", fmt.Errorf("transaction failed after %d attempts: %v", maxRetries, err)
		}
	}

	return nil, "", fmt.Errorf("transaction failed after %d attempts", maxRetries)
}
