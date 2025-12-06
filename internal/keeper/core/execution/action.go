package execution

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/metrics"
	dockertypes "github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor/types"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (e *TaskExecutor) executeAction(targetData *types.TaskTargetData, triggerData *types.TaskTriggerData, nonce uint64, client *ethclient.Client) (types.PerformerActionData, error) {
	if targetData.TaskDefinitionID != 7 && targetData.TargetContractAddress == "" {
		e.logger.Errorf("Execution contract address not configured")
		return types.PerformerActionData{}, fmt.Errorf("execution contract address not configured")
	}

	var timeToNextTrigger time.Duration
	switch targetData.TaskDefinitionID {
	case 1:
		timeToNextTrigger = time.Until(triggerData.NextTriggerTimestamp)
		timeToNextTrigger = timeToNextTrigger - 4*time.Second
	case 2:
		timeToNextTrigger = time.Until(triggerData.NextTriggerTimestamp)
		timeToNextTrigger = timeToNextTrigger - 60*time.Second
		if timeToNextTrigger < 0 {
			timeToNextTrigger = 0
		}
	default:
		timeToNextTrigger = 0
	}
	time.Sleep(timeToNextTrigger)

	// Declare variables before switch to avoid goto issues
	targetContractAddress := ethcommon.HexToAddress(targetData.TargetContractAddress)
	var callData []byte
	var convertedArgs []interface{}
	var err error

	// Skip ABI parsing for custom scripts (TaskDefinitionID 7)
	var contractABI *abi.ABI
	var method *abi.Method
	if targetData.TaskDefinitionID != 7 {
		contractABI, method, err = e.getContractMethodAndABI(targetData.TargetFunction, targetData)
		if err != nil {
			return types.PerformerActionData{}, fmt.Errorf("failed to get contract method and ABI: %v", err)
		}
	}

	var argData []interface{}
	var result *dockertypes.ExecutionResult
	var customScriptOutput *types.CustomScriptOutput
	var storageUpdates map[string]string

	switch targetData.TaskDefinitionID {
	case 7:
		// Custom script execution (TaskDefinitionID = 7)
		scriptOutput, updates, err := e.ExecuteCustomScript(context.Background(), targetData, triggerData)
		if err != nil {
			return types.PerformerActionData{}, fmt.Errorf("custom script execution failed: %v", err)
		}
		customScriptOutput = scriptOutput
		storageUpdates = updates

		// If script says don't execute, return early
		if !customScriptOutput.ShouldExecute {
			e.logger.Infof("[CustomScript] Script returned shouldExecute=false, skipping execution")
			return types.PerformerActionData{
				TaskID:         targetData.TaskID,
				Status:         true,
				StorageUpdates: storageUpdates,
			}, nil
		}

		// Override target contract address with script output
		targetContractAddress = ethcommon.HexToAddress(customScriptOutput.TargetContract)
		// Calldata is already built by the script
		callData = ethcommon.FromHex(customScriptOutput.Calldata)

		e.logger.Infof("[CustomScript] Script returned: target=%s, calldata=%s",
			customScriptOutput.TargetContract, customScriptOutput.Calldata[:min(len(customScriptOutput.Calldata), 66)])

		// Create dummy result for resource tracking
		result = &dockertypes.ExecutionResult{
			Stats: dockertypes.DockerResourceStats{
				TotalCost: big.NewInt(int64(e.validator.GetDockerExecutor().GetExecutionFeeConfig().TransactionCost * 1e18)),
			},
		}

		// Skip normal argument processing for custom scripts
		goto skipArgumentProcessing

	case 1, 2, 3, 4, 5, 6:
		var execErr error
		// Use the DockerManager from the validator to execute the code
		metadata := map[string]string{
			"task_definition_id":      fmt.Sprintf("%d", targetData.TaskDefinitionID),
			"target_chain_id":         targetData.TargetChainID,
			"target_contract_address": targetData.TargetContractAddress,
			"target_function":         targetData.TargetFunction,
			"abi":                     targetData.ABI,
			"from_address":            config.GetTaskExecutionAddress(),
		}
		if targetData.TaskDefinitionID == 1 || targetData.TaskDefinitionID == 3 || targetData.TaskDefinitionID == 5 {
			argData = e.parseStaticArgs(targetData.Arguments)
			argDataJSON, err := json.Marshal(argData)
			if err != nil {
				return types.PerformerActionData{}, fmt.Errorf("failed to marshal static args: %v", err)
			}
			metadata["on_chain_args"] = string(argDataJSON)
		}

		// e.logger.Infof("Metadata: %+v", metadata)

		result, execErr = e.validator.GetDockerExecutor().Execute(context.Background(), targetData.DynamicArgumentsScriptUrl, "go", 1, config.GetAlchemyAPIKey(), metadata)
		if execErr != nil {
			return types.PerformerActionData{}, fmt.Errorf("failed to execute script: %v", execErr)
		}

		if !result.Success {
			return types.PerformerActionData{}, fmt.Errorf("failed to execute script: %v", result.Error)
		}

		if targetData.TaskDefinitionID == 2 || targetData.TaskDefinitionID == 4 || targetData.TaskDefinitionID == 6 {
			argData = e.parseDynamicArgs(result.Output)
			e.logger.Debugf("Parsed dynamic arguments: %+v", argData)
		} else {
			argData = e.parseStaticArgs(targetData.Arguments)
		}
	default:
		return types.PerformerActionData{}, fmt.Errorf("unsupported task definition id: %d", targetData.TaskDefinitionID)
	}

	// Handle args as potentially structured data
	convertedArgs, err = e.processArguments(argData, method.Inputs, contractABI)
	if err != nil {
		return types.PerformerActionData{}, fmt.Errorf("error processing (dynamic) arguments: %v", err)
	}

	// Pack the target contract's function call data
	callData, err = contractABI.Pack(method.Name, convertedArgs...)
	if err != nil {
		e.logger.Warnf("Error packing arguments: %v", err)
		return types.PerformerActionData{}, fmt.Errorf("error packing arguments to function call: %v", err)
	}

skipArgumentProcessing:
	// Create transaction data for execution contract
	privateKey, err := crypto.HexToECDSA(config.GetPrivateKeyController())
	if err != nil {
		return types.PerformerActionData{}, fmt.Errorf("failed to parse private key: %v", err)
	}
	e.logger.Debugf("Using nonce: %d", nonce)

	// Pack the execution contract's executeFunction call
	executionABI, err := abi.JSON(strings.NewReader(`[{"inputs":[{"internalType":"uint256","name":"jobId","type":"uint256"},{"internalType":"uint256","name":"tgAmount","type":"uint256"},{"internalType":"address","name":"target","type":"address"},{"internalType":"bytes","name":"data","type":"bytes"}],"name":"executeFunction","outputs":[],"stateMutability":"payable","type":"function"}]`))
	if err != nil {
		return types.PerformerActionData{}, fmt.Errorf("failed to parse execution contract ABI: %v", err)
	}

	// Convert *BigInt to *big.Int for ABI packing
	var jobIDBigInt *big.Int
	if targetData.JobID != nil {
		jobIDBigInt = targetData.JobID.ToBigInt()
	} else {
		jobIDBigInt = big.NewInt(0)
	}

	executionInput, err := executionABI.Pack("executeFunction", jobIDBigInt, result.Stats.CurrentTotalCost, targetContractAddress, callData)
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
	receipt, finalTxHash, err := nonceManager.SubmitTransaction(
		context.Background(),
		nonce,
		ethcommon.HexToAddress(executionContractAddress),
		executionInput,
		chainID,
		privateKey,
	)
	if err != nil {
		return types.PerformerActionData{}, fmt.Errorf("failed to submit transaction: %v", err)
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
		ConvertedArguments: convertedArgs,
		StorageUpdates:     storageUpdates, // Include storage updates for custom scripts
	}
	metrics.TransactionsSentTotal.WithLabelValues(targetData.TargetChainID, "success").Inc()
	metrics.GasUsedTotal.WithLabelValues(targetData.TargetChainID).Add(float64(receipt.GasUsed))
	metrics.TransactionFeesTotal.WithLabelValues(targetData.TargetChainID).Add(float64(receipt.GasUsed))

	e.logger.Infof("Task ID %d executed successfully. Transaction: %s", targetData.TaskID, finalTxHash)

	return executionResult, nil
}
