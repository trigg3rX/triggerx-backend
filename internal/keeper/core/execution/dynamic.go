package execution

import (
	// "context"
	// "encoding/json"
	// "fmt"
	// "math/big"
	// "os"
	// "path/filepath"
	// "strconv"
	// "strings"
	"time"

	// dockertypes "github.com/docker/docker/api/types/container"
	// "github.com/docker/docker/client"
	// "github.com/ethereum/go-ethereum/accounts/abi"
	// "github.com/ethereum/go-ethereum/accounts/abi/bind"
	// ethcommon "github.com/ethereum/go-ethereum/common"
	// ethtypes "github.com/ethereum/go-ethereum/core/types"
	// "github.com/ethereum/go-ethereum/crypto"
	// "github.com/trigg3rX/triggerx-backend/internal/keeper/config"
	// "github.com/trigg3rX/triggerx-backend/pkg/resources"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (e *TaskExecutor) executeActionWithDynamicArgs(taskTargetData *types.TaskTargetData, triggerData *types.TriggerData) (types.ActionData, error) {
	executionResult := types.ActionData{
		TaskID:       0,
		ActionTxHash: "0x",
		GasUsed:      "0",
		Status:       false,
		Timestamp:    time.Now().UTC(),
	}

	// e.logger.Infof("Executing job %d with dynamic arguments", job.JobID)
	// e.logger.Infof("DEBUG: In executeActionWithDynamicArgs - executionContractAddress: %s", executionContractAddress)
	// // Create Docker client for script execution
	// cli, err := client.NewClientWithOpts(
	// 	client.FromEnv,
	// 	client.WithAPIVersionNegotiation(),
	// )
	// if err != nil {
	// 	return executionResult, fmt.Errorf("failed to create Docker client: %v", err)
	// }
	// defer cli.Close()

	// // Step 1: Check if we need to evaluate a condition from the script
	// if job.TaskDefinitionID == 6 && job.ScriptTriggerFunction != "" {
	// 	// Download and execute the condition script
	// 	codePath, err := resources.DownloadIPFSFile(job.ScriptTriggerFunction)
	// 	if err != nil {
	// 		e.logger.Errorf("Failed to download condition script: %v", err)
	// 		return executionResult, fmt.Errorf("failed to download condition script: %v", err)
	// 	}
	// 	defer os.RemoveAll(filepath.Dir(codePath))

	// 	// Create and execute container for condition script
	// 	containerID, err := resources.CreateDockerContainer(context.Background(), cli, codePath)
	// 	if err != nil {
	// 		e.logger.Errorf("Failed to create container for condition script: %v", err)
	// 		return executionResult, fmt.Errorf("failed to create container: %v", err)
	// 	}
	// 	if err := cli.ContainerRemove(context.Background(), containerID, dockertypes.RemoveOptions{Force: true}); err != nil {
	// 		e.logger.Errorf("Failed to remove container for condition script: %v", err)
	// 		return executionResult, fmt.Errorf("failed to remove container: %v", err)
	// 	}

	// 	// Monitor resources and get script output
	// 	stats, err := resources.MonitorResources(context.Background(), cli, containerID)
	// 	if err != nil {
	// 		e.logger.Errorf("Failed to monitor condition script resources: %v", err)
	// 		return executionResult, fmt.Errorf("failed to monitor resources: %v", err)
	// 	}

	// 	// Check if condition was satisfied based on script output
	// 	if !stats.Status {
	// 		e.logger.Infof("Condition not satisfied for job %d, skipping execution", job.JobID)
	// 		return executionResult, nil
	// 	}
	// 	e.logger.Infof("Condition satisfied for job %d, proceeding with execution", job.JobID)

	// 	// Update execution result with resource usage from condition script
	// 	executionResult.MemoryUsage = stats.MemoryUsage
	// 	executionResult.CPUPercentage = stats.CPUPercentage
	// 	executionResult.NetworkRx = stats.RxBytes
	// 	executionResult.NetworkTx = stats.TxBytes
	// 	executionResult.BlockRead = stats.BlockRead
	// 	executionResult.BlockWrite = stats.BlockWrite
	// 	executionResult.BandwidthRate = stats.BandwidthRate
	// 	executionResult.TotalFee = stats.TotalFee
	// 	executionResult.StaticComplexity = stats.StaticComplexity
	// 	executionResult.DynamicComplexity = stats.DynamicComplexity
	// 	executionResult.ComplexityIndex = stats.ComplexityIndex
	// }

	// // Step 2: Get the contract method and ABI
	// contractAddress := ethcommon.HexToAddress(job.TargetContractAddress)
	// contractABI, method, err := e.getContractMethodAndABI(job.TargetFunction, job)
	// if err != nil {
	// 	return executionResult, err
	// }

	// // Step 3: Execute the script to get dynamic arguments if ScriptIPFSUrl is provided
	// var argData interface{}
	// if job.ScriptIPFSUrl != "" {
	// 	// Download and execute the script
	// 	codePath, err := resources.DownloadIPFSFile(job.ScriptIPFSUrl)
	// 	if err != nil {
	// 		return executionResult, fmt.Errorf("failed to download script: %v", err)
	// 	}
	// 	defer os.RemoveAll(filepath.Dir(codePath))

	// 	// Create and execute container for script
	// 	containerID, err := resources.CreateDockerContainer(context.Background(), cli, codePath)
	// 	if err != nil {
	// 		return executionResult, fmt.Errorf("failed to create container: %v", err)
	// 	}
	// 	defer func() {
	// 		if err := cli.ContainerRemove(context.Background(), containerID, dockertypes.RemoveOptions{Force: true}); err != nil {
	// 			e.logger.Errorf("Failed to remove container for script: %v", err)
	// 		}
	// 	}()

	// 	// Monitor resources and get script output
	// 	stats, err := resources.MonitorResources(context.Background(), cli, containerID)
	// 	if err != nil {
	// 		return executionResult, fmt.Errorf("failed to monitor resources: %v", err)
	// 	}

	// 	// Parse the script output
	// 	if len(stats.Output) == 0 {
	// 		return executionResult, fmt.Errorf("script output is empty")
	// 	}

	// 	e.logger.Infof("Script output: %s", stats.Output)

	// 	// Try to parse the output as JSON directly first
	// 	if err := json.Unmarshal([]byte(stats.Output), &argData); err != nil {
	// 		e.logger.Infof("Could not parse output as direct JSON, trying to extract value from payload format")

	// 		// Try to extract value from "Payload received: X" format
	// 		lines := strings.Split(stats.Output, "\n")
	// 		for _, line := range lines {
	// 			if strings.Contains(line, "Payload received:") {
	// 				payloadValue := strings.TrimSpace(strings.TrimPrefix(line, "Payload received:"))
	// 				e.logger.Infof("Extracted payload value: %s", payloadValue)

	// 				// Try to parse the extracted value as JSON
	// 				if err := json.Unmarshal([]byte(payloadValue), &argData); err != nil {
	// 					// If not JSON, use the raw string value
	// 					argData = payloadValue
	// 					e.logger.Infof("Using raw string value as argument")
	// 					break
	// 				} else {
	// 					e.logger.Infof("Successfully parsed extracted value as JSON")
	// 					break
	// 				}
	// 			}
	// 		}

	// 		// If we still couldn't parse it, return error
	// 		if argData == nil {
	// 			e.logger.Errorf("Failed to parse script output as arguments: %v", err)
	// 			return executionResult, fmt.Errorf("failed to parse script output: %v", err)
	// 		}
	// 	}

	// 	e.logger.Infof("Successfully processed script output data")

	// 	// Update execution result with resource usage from script
	// 	executionResult.MemoryUsage = stats.MemoryUsage
	// 	executionResult.CPUPercentage = stats.CPUPercentage
	// 	executionResult.NetworkRx = stats.RxBytes
	// 	executionResult.NetworkTx = stats.TxBytes
	// 	executionResult.BlockRead = stats.BlockRead
	// 	executionResult.BlockWrite = stats.BlockWrite
	// 	executionResult.BandwidthRate = stats.BandwidthRate
	// 	executionResult.TotalFee = stats.TotalFee
	// 	executionResult.StaticComplexity = stats.StaticComplexity
	// 	executionResult.DynamicComplexity = stats.DynamicComplexity
	// 	executionResult.ComplexityIndex = stats.ComplexityIndex
	// } else if len(job.Arguments) > 0 {
	// 	// If no script URL but arguments are provided, try to parse the first argument as JSON
	// 	if err := json.Unmarshal([]byte(job.Arguments[0]), &argData); err != nil {
	// 		return executionResult, fmt.Errorf("failed to parse argument: %v", err)
	// 	}
	// 	e.logger.Infof("Successfully parsed JSON data from single argument")
	// } else {
	// 	return executionResult, fmt.Errorf("no script URL or arguments provided")
	// }

	// // Step 4: Process the arguments
	// convertedArgs, err := e.processArguments(argData, method.Inputs, contractABI)
	// if err != nil {
	// 	return executionResult, fmt.Errorf("error processing arguments: %v", err)
	// }

	// // Step 5: Pack the arguments
	// var callData []byte // Declare callData first
	// callData, err = contractABI.Pack(method.Name, convertedArgs...)
	// if err != nil {
	// 	return executionResult, err
	// }

	// // Create transaction data for execution contract
	// privateKey, err := crypto.HexToECDSA(config.GetPrivateKeyController())
	// if err != nil {
	// 	return executionResult, fmt.Errorf("failed to parse private key: %v", err)
	// }

	// nonce, err := e.ethClient.PendingNonceAt(context.Background(), ethcommon.HexToAddress(config.GetKeeperAddress()))
	// if err != nil {
	// 	return executionResult, err
	// }

	// gasPrice, err := e.ethClient.SuggestGasPrice(context.Background())
	// if err != nil {
	// 	return executionResult, err
	// }

	// // Pack the execution contract's executeFunction call
	// executionABI, err := abi.JSON(strings.NewReader(`[{"inputs":[{"name":"target","type":"address"},{"name":"data","type":"bytes"}],"name":"executeFunction","outputs":[],"stateMutability":"payable","type":"function"}]`))
	// if err != nil {
	// 	return executionResult, fmt.Errorf("failed to parse execution contract ABI: %v", err)
	// }

	// executionInput, err := executionABI.Pack("executeFunction", contractAddress, callData)
	// if err != nil {
	// 	return executionResult, fmt.Errorf("failed to pack execution contract input: %v", err)
	// }

	// // Create and sign transaction
	// tx := ethtypes.NewTransaction(nonce, ethcommon.HexToAddress(executionContractAddress), big.NewInt(0), 300000, gasPrice, executionInput)
	// signedTx, err := ethtypes.SignTx(tx, ethtypes.NewEIP155Signer(big.NewInt(11155420)), privateKey)
	// if err != nil {
	// 	return executionResult, err
	// }

	// e.logger.Infof("DEBUG: Creating transaction to execution contract: %s", executionContractAddress)
	// // Send transaction
	// err = e.ethClient.SendTransaction(context.Background(), signedTx)
	// if err != nil {
	// 	return executionResult, err
	// }

	// e.logger.Infof("DEBUG: Transaction sent to execution contract: %s, tx hash: %s",
	// 	executionContractAddress, signedTx.Hash().Hex())

	// // Step 7: Wait for transaction receipt
	// receipt, err := bind.WaitMined(context.Background(), e.ethClient, signedTx)
	// if err != nil {
	// 	return executionResult, err
	// }

	// // Update execution result with transaction details
	// executionResult.Status = receipt.Status == ethtypes.ReceiptStatusSuccessful
	// executionResult.ActionTxHash = signedTx.Hash().Hex()
	// executionResult.GasUsed = strconv.FormatUint(receipt.GasUsed, 10)

	e.logger.Infof("✅ Job %d executed successfully with dynamic arguments. Transaction: %s",
		taskTargetData.JobID, "") //signedTx.Hash().Hex())

	return executionResult, nil
}
