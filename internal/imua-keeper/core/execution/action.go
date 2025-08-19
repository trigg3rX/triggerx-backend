package execution

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/trigg3rX/triggerx-backend/internal/imua-keeper/config"
	"github.com/trigg3rX/triggerx-backend/internal/imua-keeper/metrics"
	dockertypes "github.com/trigg3rX/triggerx-backend/pkg/docker/types"
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
		timeToNextTrigger = timeToNextTrigger - 4*time.Second
		if timeToNextTrigger < 0 {
			timeToNextTrigger = 0
		}
	default:
		timeToNextTrigger = 0
	}
	time.Sleep(timeToNextTrigger)

	targetContractAddress := ethcommon.HexToAddress(targetData.TargetContractAddress)
	contractABI, method, err := e.getContractMethodAndABI(targetData.TargetFunction, targetData)
	if err != nil {
		return types.PerformerActionData{}, fmt.Errorf("failed to get contract method and ABI: %v", err)
	}

	var argData []interface{}
	var result *dockertypes.ExecutionResult
	switch targetData.TaskDefinitionID {
	case 2, 4, 6:
		var execErr error
		// Use the DockerManager from the validator to execute the code
		result, execErr = e.validator.GetDockerManager().Execute(context.Background(), targetData.DynamicArgumentsScriptUrl, 1)
		if execErr != nil {
			return types.PerformerActionData{}, fmt.Errorf("failed to execute dynamic arguments script: %v", execErr)
		}

		if !result.Success {
			return types.PerformerActionData{}, fmt.Errorf("failed to execute dynamic arguments script: %v", result.Error)
		}

		argData = e.parseDynamicArgs(result.Output)
		e.logger.Debugf("Parsed dynamic arguments: %+v", argData)
	case 1, 3, 5:
		argData = e.parseStaticArgs(targetData.Arguments)
		result = &dockertypes.ExecutionResult{
			Stats: dockertypes.DockerResourceStats{
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
	e.logger.Debugf("Using nonce: %d", nonce)

	// Pack the execution contract's executeFunction call
	executionABI, err := abi.JSON(strings.NewReader(`[{"inputs":[{"internalType":"uint256","name":"jobId","type":"uint256"},{"internalType":"uint256","name":"tgAmount","type":"uint256"},{"internalType":"address","name":"target","type":"address"},{"internalType":"bytes","name":"data","type":"bytes"}],"name":"executeFunction","outputs":[],"stateMutability":"payable","type":"function"}]`))
	if err != nil {
		return types.PerformerActionData{}, fmt.Errorf("failed to parse execution contract ABI: %v", err)
	}

	// According to the ABI, the function signature is:
	// executeFunction(uint256 jobId, uint256 tgAmount, address target, bytes data)
	// We use jobId from targetData.JobID, and tgAmount is determined by the execution result's total cost.
	var tgAmountBigInt = big.NewInt(0)
	if result != nil {
		// Assuming TotalCost is in float64 and needs to be converted to wei (1e18 multiplier) if it's in ETH
		tgAmountBigInt = new(big.Int).SetInt64(int64(result.Stats.TotalCost * 1e18))
	}

	// Convert *BigInt to *big.Int for ABI packing
	var jobIDBigInt *big.Int
	if targetData.JobID != nil {
		jobIDBigInt = targetData.JobID.ToBigInt()
	} else {
		jobIDBigInt = big.NewInt(0)
	}

	executionInput, err := executionABI.Pack("executeFunction", jobIDBigInt, tgAmountBigInt, targetContractAddress, callData)
	if err != nil {
		return types.PerformerActionData{}, fmt.Errorf("failed to pack execution contract input: %v", err)
	}

	executionContractAddress := config.GetTaskExecutionAddress()
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return types.PerformerActionData{}, fmt.Errorf("failed to get chain ID: %v", err)
	}

	// Get nonce manager for this chain
	nonceManager, err := e.getNonceManager(targetData.TargetChainID)
	if err != nil {
		return types.PerformerActionData{}, fmt.Errorf("failed to get nonce manager: %w", err)
	}

	// Submit transaction with smart retry
	receipt, finalTxHash, err := nonceManager.SubmitTransactionWithSmartRetry(
		context.Background(),
		nonce,
		ethcommon.HexToAddress(executionContractAddress),
		executionInput,
		chainID,
		config.GetPrivateKeyController(),
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
