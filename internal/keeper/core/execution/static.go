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

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	dockertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
	"github.com/trigg3rX/triggerx-backend/pkg/resources"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

const executionContractAddress = "0x68605feB94a8FeBe5e1fBEF0A9D3fE6e80cEC126"

func (e *TaskExecutor) executeActionWithStaticArgs(job *types.HandleCreateJobData) (types.ActionData, error) {

	executionResult := types.ActionData{
		TaskID:       0,
		ActionTxHash: "0x",
		GasUsed:      "0",
		Status:       false,
		Timestamp:    time.Now().UTC(),
	}

	e.logger.Infof("DEBUG: In executeActionWithStaticArgs - executionContractAddress: %s", executionContractAddress)

	if executionContractAddress == "" {
		e.logger.Errorf("Execution contract address not configured")
		return executionResult, fmt.Errorf("execution contract address not configured")
	}

	// Create Docker client for script execution if needed
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return executionResult, fmt.Errorf("failed to create Docker client: %v", err)
	}
	defer cli.Close()

	if job.TaskDefinitionID == 5 && job.ScriptTriggerFunction != "" {
		// Download and execute the condition script
		codePath, err := resources.DownloadIPFSFile(job.ScriptTriggerFunction)
		if err != nil {
			e.logger.Errorf("Failed to download condition script: %v", err)
			return executionResult, fmt.Errorf("failed to download condition script: %v", err)
		}
		defer os.RemoveAll(filepath.Dir(codePath))

		// Create and execute container for condition script
		containerID, err := resources.CreateDockerContainer(context.Background(), cli, codePath)
		if err != nil {
			e.logger.Errorf("Failed to create container for condition script: %v", err)
			return executionResult, fmt.Errorf("failed to create container: %v", err)
		}
		defer func() {
			if err := cli.ContainerRemove(context.Background(), containerID, dockertypes.RemoveOptions{Force: true}); err != nil {
				e.logger.Errorf("Failed to remove container for condition script: %v", err)
			}
		}()

		// Monitor resources and get script output
		stats, err := resources.MonitorResources(context.Background(), cli, containerID)
		if err != nil {
			e.logger.Errorf("Failed to monitor condition script resources: %v", err)
			return executionResult, fmt.Errorf("failed to monitor resources: %v", err)
		}

		// Check if condition was satisfied based on script output
		if !stats.Status {
			e.logger.Infof("Condition not satisfied for job %d, skipping execution", job.JobID)
			return executionResult, nil
		}
		e.logger.Infof("Condition satisfied for job %d, proceeding with execution", job.JobID)

		// Update execution result with resource usage from condition script
		executionResult.MemoryUsage = stats.MemoryUsage
		executionResult.CPUPercentage = stats.CPUPercentage
		executionResult.NetworkRx = stats.RxBytes
		executionResult.NetworkTx = stats.TxBytes
		executionResult.BlockRead = stats.BlockRead
		executionResult.BlockWrite = stats.BlockWrite
		executionResult.BandwidthRate = stats.BandwidthRate
		executionResult.TotalFee = stats.TotalFee
		executionResult.StaticComplexity = stats.StaticComplexity
		executionResult.DynamicComplexity = stats.DynamicComplexity
		executionResult.ComplexityIndex = stats.ComplexityIndex
	}

	contractAddress := ethcommon.HexToAddress(job.TargetContractAddress)
	contractABI, method, err := e.getContractMethodAndABI(job.TargetFunction, job)
	if err != nil {
		return executionResult, err
	}

	// Handle args as potentially structured data
	convertedArgs, err := e.processArguments(job.Arguments, method.Inputs, contractABI)
	if err != nil {
		return executionResult, fmt.Errorf("error processing arguments: %v", err)
	}

	// Pack the target contract's function call data
	var callData []byte // Declare callData first
	callData, err = contractABI.Pack(method.Name, convertedArgs...)
	if err != nil {
		e.logger.Warnf("Error packing arguments: %v", err)
		return executionResult, err
	}

	// Create transaction data for execution contract
	privateKey, err := crypto.HexToECDSA(config.GetPrivateKeyController())
	if err != nil {
		return executionResult, fmt.Errorf("failed to parse private key: %v", err)
	}

	nonce, err := e.ethClient.PendingNonceAt(context.Background(), ethcommon.HexToAddress(config.GetKeeperAddress()))
	if err != nil {
		return executionResult, err
	}

	gasPrice, err := e.ethClient.SuggestGasPrice(context.Background())
	if err != nil {
		return executionResult, err
	}

	// Pack the execution contract's executeFunction call
	executionABI, err := abi.JSON(strings.NewReader(`[{"inputs":[{"name":"target","type":"address"},{"name":"data","type":"bytes"}],"name":"executeFunction","outputs":[],"stateMutability":"payable","type":"function"}]`))
	if err != nil {
		return executionResult, fmt.Errorf("failed to parse execution contract ABI: %v", err)
	}

	executionInput, err := executionABI.Pack("executeFunction", contractAddress, callData)
	if err != nil {
		return executionResult, fmt.Errorf("failed to pack execution contract input: %v", err)
	}

	// Create and sign transaction
	tx := ethtypes.NewTransaction(nonce, ethcommon.HexToAddress(executionContractAddress), big.NewInt(0), 300000, gasPrice, executionInput)
	signedTx, err := ethtypes.SignTx(tx, ethtypes.NewEIP155Signer(big.NewInt(11155420)), privateKey)
	if err != nil {
		return executionResult, err
	}

	err = e.ethClient.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return executionResult, err
	}

	e.logger.Infof("DEBUG: Transaction sent to execution contract: %s, tx hash: %s",
		executionContractAddress, signedTx.Hash().Hex())
	// Wait for transaction receipt
	receipt, err := bind.WaitMined(context.Background(), e.ethClient, signedTx)
	if err != nil {
		e.logger.Warnf("Error waiting for transaction: %v", err)
		return executionResult, err
	}

	executionResult.Status = receipt.Status == ethtypes.ReceiptStatusSuccessful
	executionResult.ActionTxHash = signedTx.Hash().Hex()
	executionResult.GasUsed = strconv.FormatUint(receipt.GasUsed, 10)

	e.logger.Infof("Job %d executed successfully. Transaction: %s", job.JobID, signedTx.Hash().Hex())

	return executionResult, nil
}
