package execution

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (e *TaskExecutor) executeActionWithDynamicArgs(taskTargetData *types.SendTaskTargetDataToKeeper) (types.PerformerActionData, error) {
	chainRpcUrl := e.getChainRpcUrl(taskTargetData.TargetChainID)

	chainClient, err := ethclient.Dial(chainRpcUrl)
	if err != nil {
		return types.PerformerActionData{}, fmt.Errorf("failed to connect to chain: %v", err)
	}
	defer chainClient.Close()

	if taskTargetData.TargetContractAddress == "" {
		e.logger.Errorf("Execution contract address not configured")
		return types.PerformerActionData{}, fmt.Errorf("execution contract address not configured")
	}

	taregtContractAddress := ethcommon.HexToAddress(taskTargetData.TargetContractAddress)
	contractABI, method, err := e.getContractMethodAndABI(taskTargetData.TargetFunction, taskTargetData)
	if err != nil {
		return types.PerformerActionData{}, fmt.Errorf("failed to get contract method and ABI: %v", err)
	}

	codePath, err := e.codeExecutor.Downloader.DownloadFile(context.Background(), taskTargetData.DynamicArgumentsScriptUrl)
	if err != nil {
		return types.PerformerActionData{}, fmt.Errorf("failed to download dynamic arguments script: %v", err)
	}
	defer os.RemoveAll(filepath.Dir(codePath))

	containerID, err := e.codeExecutor.DockerManager.CreateContainer(context.Background(), codePath)
	if err != nil {
		return types.PerformerActionData{}, fmt.Errorf("failed to create container: %v", err)
	}
	defer e.codeExecutor.DockerManager.CleanupContainer(context.Background(), containerID)

	result, err := e.codeExecutor.MonitorExecution(context.Background(), e.codeExecutor.DockerManager.Cli, containerID, 1)
	if err != nil {
		return types.PerformerActionData{}, fmt.Errorf("failed to monitor execution: %v", err)
	}

	if !result.Success {
		return types.PerformerActionData{}, fmt.Errorf("failed to execute dynamic arguments script: %v", result.Error)
	}

	argData := e.parseDynamicArgs(result.Output)

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

	lastUsedNonce := config.GetChainNonce(taskTargetData.TargetChainID)
	nonce := lastUsedNonce + 1
	config.IncrementChainNonce(taskTargetData.TargetChainID)

	gasPrice, err := chainClient.SuggestGasPrice(context.Background())
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

	executionContractAddress := e.getExecutionContractAddress(taskTargetData.TargetChainID)
	chainID, err := chainClient.ChainID(context.Background())
	if err != nil {
		return types.PerformerActionData{}, fmt.Errorf("failed to get chain ID: %v", err)
	}

	// Create and sign transaction
	tx := ethtypes.NewTransaction(nonce, ethcommon.HexToAddress(executionContractAddress), big.NewInt(0), 300000, gasPrice, executionInput)
	signedTx, err := ethtypes.SignTx(tx, ethtypes.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return types.PerformerActionData{}, err
	}

	err = chainClient.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return types.PerformerActionData{}, err
	}

	e.logger.Debugf("Transaction sent to execution contract: %s, tx hash: %s",
		executionContractAddress, signedTx.Hash().Hex())
	// Wait for transaction receipt
	receipt, err := bind.WaitMined(context.Background(), chainClient, signedTx)
	if err != nil {
		e.logger.Warnf("Error waiting for transaction: %v", err)
		return types.PerformerActionData{}, err
	}

	executionResult := types.PerformerActionData{
		TaskID: taskTargetData.TaskID,
		ActionTxHash:  signedTx.Hash().Hex(),
		GasUsed:       strconv.FormatUint(receipt.GasUsed, 10),
		Status:        receipt.Status == ethtypes.ReceiptStatusSuccessful,
		MemoryUsage:   result.Stats.MemoryUsage,
		CPUPercentage: result.Stats.CPUPercentage,
		NetworkRx:     result.Stats.RxBytes,
		NetworkTx:     result.Stats.TxBytes,
		BlockRead:     result.Stats.BlockRead,
		BlockWrite:    result.Stats.BlockWrite,
		BandwidthRate: result.Stats.BandwidthRate,
		TotalFee:      result.Stats.TotalCost,
		StaticComplexity: result.Stats.StaticComplexity,
		DynamicComplexity: result.Stats.DynamicComplexity,
		ExecutionTimestamp: time.Now().UTC(),
	}

	e.logger.Infof("Task ID %d executed successfully. Transaction: %s", taskTargetData.TaskID, signedTx.Hash().Hex())

	return executionResult, nil
}
