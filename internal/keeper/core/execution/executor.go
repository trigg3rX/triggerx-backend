package execution

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"math/big"
// 	"os"
// 	"path/filepath"
// 	"strconv"
// 	"strings"
// 	"time"

// 	dockertypes "github.com/docker/docker/api/types"
// 	"github.com/docker/docker/client"
// 	"github.com/ethereum/go-ethereum/accounts/abi"
// 	"github.com/ethereum/go-ethereum/accounts/abi/bind"
// 	ethcommon "github.com/ethereum/go-ethereum/common"
// 	"github.com/ethereum/go-ethereum/core/types"
// 	"github.com/ethereum/go-ethereum/crypto"
// 	"github.com/ethereum/go-ethereum/ethclient"
// 	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
// 	"github.com/trigg3rX/triggerx-backend/internal/keeper/interfaces"
// 	"github.com/trigg3rX/triggerx-backend/internal/keeper/validation"
// 	"github.com/trigg3rX/triggerx-backend/pkg/resources"
// 	jobtypes "github.com/trigg3rX/triggerx-backend/pkg/types"
// )

// type JobExecutor struct {
// 	ethClient       interfaces.EthClientInterface
// 	etherscanAPIKey string
// 	argConverter    *ArgumentConverter
// 	validator       interfaces.ValidatorInterface
// 	logger          interfaces.Logger
// }

// func NewJobExecutor(ethClient *ethclient.Client, etherscanAPIKey string) *JobExecutor {
// 	return &JobExecutor{
// 		ethClient:       ethClient,
// 		etherscanAPIKey: etherscanAPIKey,
// 		argConverter:    &ArgumentConverter{},
// 		validator:       validation.NewJobValidator(logger, ethClient),
// 		logger:          logger,
// 	}
// }

// func (e *JobExecutor) Execute(job *jobtypes.HandleCreateJobData) (jobtypes.ActionData, error) {
// 	e.logger.Infof("executionContractAddress value: %s", executionContractAddress)

// 	executionResult := jobtypes.ActionData{
// 		TaskID:       0,
// 		ActionTxHash: "0x",
// 		GasUsed:      "0",
// 		Status:       false,
// 		Timestamp:    time.Now().UTC(),
// 	}

// 	e.logger.Infof("Validating job %d (taskDefID: %d) before execution", job.JobID, job.TaskDefinitionID)
// 	shouldExecute, err := e.validator.ValidateAndPrepareJob(job, nil)
// 	if err != nil {
// 		e.logger.Warnf("Job validation error: %v", err)
// 		return executionResult, fmt.Errorf("job validation failed: %v", err)
// 	}

// 	if !shouldExecute {
// 		e.logger.Infof("Job %d validation determined execution should be skipped", job.JobID)
// 		return executionResult, nil
// 	}

// 	e.logger.Infof("Job %d validated successfully, proceeding with execution", job.JobID)
// 	e.logger.Infof("Executing job: %d (Function: %s)", job.JobID, job.TargetFunction)

// 	switch job.TaskDefinitionID {
// 	case 1, 3, 5:
// 		return e.executeActionWithStaticArgs(job)
// 	case 2, 4, 6:
// 		return e.executeActionWithDynamicArgs(job)
// 	default:
// 		return jobtypes.ActionData{}, fmt.Errorf("unsupported task definition id: %d", job.TaskDefinitionID)
// 	}
// }

// func (e *JobExecutor) executeActionWithStaticArgs(job *jobtypes.HandleCreateJobData) (jobtypes.ActionData, error) {

// 	executionResult := jobtypes.ActionData{
// 		TaskID:       0,
// 		ActionTxHash: "0x",
// 		GasUsed:      "0",
// 		Status:       false,
// 		Timestamp:    time.Now().UTC(),
// 	}

// 	e.logger.Infof("DEBUG: In executeActionWithStaticArgs - executionContractAddress: %s", executionContractAddress)

// 	if executionContractAddress == "" {
// 		e.logger.Errorf("Execution contract address not configured")
// 		return executionResult, fmt.Errorf("execution contract address not configured")
// 	}

// 	cli, err := client.NewClientWithOpts(
// 		client.FromEnv,
// 		client.WithAPIVersionNegotiation(),
// 	)
// 	if err != nil {
// 		return executionResult, fmt.Errorf("failed to create Docker client: %v", err)
// 	}
// 	defer cli.Close()

// 	if job.TaskDefinitionID == 5 && job.ScriptTriggerFunction != "" {
// 		codePath, err := resources.DownloadIPFSFile(job.ScriptTriggerFunction)
// 		if err != nil {
// 			logger.Errorf("Failed to download condition script: %v", err)
// 			return executionResult, fmt.Errorf("failed to download condition script: %v", err)
// 		}
// 		defer os.RemoveAll(filepath.Dir(codePath))

// 		containerID, err := resources.CreateDockerContainer(context.Background(), cli, codePath)
// 		if err != nil {
// 			logger.Errorf("Failed to create container for condition script: %v", err)
// 			return executionResult, fmt.Errorf("failed to create container: %v", err)
// 		}
// 		defer cli.ContainerRemove(context.Background(), containerID, dockertypes.ContainerRemoveOptions{Force: true})

// 		stats, err := resources.MonitorResources(context.Background(), cli, containerID)
// 		if err != nil {
// 			logger.Errorf("Failed to monitor condition script resources: %v", err)
// 			return executionResult, fmt.Errorf("failed to monitor resources: %v", err)
// 		}

// 		if !stats.Status {
// 			logger.Infof("Condition not satisfied for job %d, skipping execution", job.JobID)
// 			return executionResult, nil
// 		}
// 		logger.Infof("Condition satisfied for job %d, proceeding with execution", job.JobID)

// 		executionResult.MemoryUsage = stats.MemoryUsage
// 		executionResult.CPUPercentage = stats.CPUPercentage
// 		executionResult.NetworkRx = stats.RxBytes
// 		executionResult.NetworkTx = stats.TxBytes
// 		executionResult.BlockRead = stats.BlockRead
// 		executionResult.BlockWrite = stats.BlockWrite
// 		executionResult.BandwidthRate = stats.BandwidthRate
// 		executionResult.TotalFee = stats.TotalFee
// 		executionResult.StaticComplexity = stats.StaticComplexity
// 		executionResult.DynamicComplexity = stats.DynamicComplexity
// 		executionResult.ComplexityIndex = stats.ComplexityIndex
// 	}

// 	contractAddress := ethcommon.HexToAddress(job.TargetContractAddress)
// 	contractABI, method, err := e.getContractMethodAndABI(job.TargetFunction, job)
// 	if err != nil {
// 		return executionResult, err
// 	}

// 	convertedArgs, err := e.processArguments(job.Arguments, method.Inputs, contractABI)
// 	if err != nil {
// 		return executionResult, fmt.Errorf("error processing arguments: %v", err)
// 	}

// 	var callData []byte
// 	callData, err = contractABI.Pack(method.Name, convertedArgs...)
// 	if err != nil {
// 		e.logger.Warnf("Error packing arguments: %v", err)
// 		return executionResult, err
// 	}

// 	privateKey, err := crypto.HexToECDSA(config.PrivateKeyController)
// 	if err != nil {
// 		return executionResult, fmt.Errorf("failed to parse private key: %v", err)
// 	}

// 	nonce, err := e.ethClient.PendingNonceAt(context.Background(), ethcommon.HexToAddress(config.KeeperAddress))
// 	if err != nil {
// 		return executionResult, err
// 	}

// 	gasPrice, err := e.ethClient.SuggestGasPrice(context.Background())
// 	if err != nil {
// 		return executionResult, err
// 	}

// 	executionABI, err := abi.JSON(strings.NewReader(`[{"inputs":[{"name":"target","type":"address"},{"name":"data","type":"bytes"}],"name":"executeFunction","outputs":[],"stateMutability":"payable","type":"function"}]`))
// 	if err != nil {
// 		return executionResult, fmt.Errorf("failed to parse execution contract ABI: %v", err)
// 	}

// 	executionInput, err := executionABI.Pack("executeFunction", contractAddress, callData)
// 	if err != nil {
// 		return executionResult, fmt.Errorf("failed to pack execution contract input: %v", err)
// 	}

// 	tx := types.NewTransaction(nonce, ethcommon.HexToAddress(executionContractAddress), big.NewInt(0), 300000, gasPrice, executionInput)
// 	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(big.NewInt(11155420)), privateKey)
// 	if err != nil {
// 		return executionResult, err
// 	}

// 	err = e.ethClient.SendTransaction(context.Background(), signedTx)
// 	if err != nil {
// 		return executionResult, err
// 	}

// 	e.logger.Infof("DEBUG: Transaction sent to execution contract: %s, tx hash: %s",
// 		executionContractAddress, signedTx.Hash().Hex())

// 	receipt, err := bind.WaitMined(context.Background(), e.ethClient, signedTx)
// 	if err != nil {
// 		e.logger.Warnf("Error waiting for transaction: %v", err)
// 		return executionResult, err
// 	}

// 	executionResult.Status = receipt.Status == types.ReceiptStatusSuccessful
// 	executionResult.ActionTxHash = signedTx.Hash().Hex()
// 	executionResult.GasUsed = strconv.FormatUint(receipt.GasUsed, 10)

// 	e.logger.Infof("Job %d executed successfully. Transaction: %s", job.JobID, signedTx.Hash().Hex())

// 	return executionResult, nil
// }

// func (e *JobExecutor) executeActionWithDynamicArgs(job *jobtypes.HandleCreateJobData) (jobtypes.ActionData, error) {
// 	executionResult := jobtypes.ActionData{
// 		TaskID:       0,
// 		ActionTxHash: "0x",
// 		GasUsed:      "0",
// 		Status:       false,
// 		Timestamp:    time.Now().UTC(),
// 	}

// 	logger.Infof("Executing job %d with dynamic arguments", job.JobID)
// 	e.logger.Infof("DEBUG: In executeActionWithDynamicArgs - executionContractAddress: %s", executionContractAddress)

// 	cli, err := client.NewClientWithOpts(
// 		client.FromEnv,
// 		client.WithAPIVersionNegotiation(),
// 	)
// 	if err != nil {
// 		return executionResult, fmt.Errorf("failed to create Docker client: %v", err)
// 	}
// 	defer cli.Close()

// 	if job.TaskDefinitionID == 6 && job.ScriptTriggerFunction != "" {
// 		codePath, err := resources.DownloadIPFSFile(job.ScriptTriggerFunction)
// 		if err != nil {
// 			logger.Errorf("Failed to download condition script: %v", err)
// 			return executionResult, fmt.Errorf("failed to download condition script: %v", err)
// 		}
// 		defer os.RemoveAll(filepath.Dir(codePath))

// 		containerID, err := resources.CreateDockerContainer(context.Background(), cli, codePath)
// 		if err != nil {
// 			logger.Errorf("Failed to create container for condition script: %v", err)
// 			return executionResult, fmt.Errorf("failed to create container: %v", err)
// 		}
// 		defer cli.ContainerRemove(context.Background(), containerID, dockertypes.ContainerRemoveOptions{Force: true})

// 		stats, err := resources.MonitorResources(context.Background(), cli, containerID)
// 		if err != nil {
// 			logger.Errorf("Failed to monitor condition script resources: %v", err)
// 			return executionResult, fmt.Errorf("failed to monitor resources: %v", err)
// 		}

// 		if !stats.Status {
// 			logger.Infof("Condition not satisfied for job %d, skipping execution", job.JobID)
// 			return executionResult, nil
// 		}
// 		logger.Infof("Condition satisfied for job %d, proceeding with execution", job.JobID)

// 		executionResult.MemoryUsage = stats.MemoryUsage
// 		executionResult.CPUPercentage = stats.CPUPercentage
// 		executionResult.NetworkRx = stats.RxBytes
// 		executionResult.NetworkTx = stats.TxBytes
// 		executionResult.BlockRead = stats.BlockRead
// 		executionResult.BlockWrite = stats.BlockWrite
// 		executionResult.BandwidthRate = stats.BandwidthRate
// 		executionResult.TotalFee = stats.TotalFee
// 		executionResult.StaticComplexity = stats.StaticComplexity
// 		executionResult.DynamicComplexity = stats.DynamicComplexity
// 		executionResult.ComplexityIndex = stats.ComplexityIndex
// 	}

// 	contractAddress := ethcommon.HexToAddress(job.TargetContractAddress)
// 	contractABI, method, err := e.getContractMethodAndABI(job.TargetFunction, job)
// 	if err != nil {
// 		return executionResult, err
// 	}

// 	var argData interface{}
// 	if job.ScriptIPFSUrl != "" {
// 		codePath, err := resources.DownloadIPFSFile(job.ScriptIPFSUrl)
// 		if err != nil {
// 			return executionResult, fmt.Errorf("failed to download script: %v", err)
// 		}
// 		defer os.RemoveAll(filepath.Dir(codePath))

// 		containerID, err := resources.CreateDockerContainer(context.Background(), cli, codePath)
// 		if err != nil {
// 			return executionResult, fmt.Errorf("failed to create container: %v", err)
// 		}
// 		defer cli.ContainerRemove(context.Background(), containerID, dockertypes.ContainerRemoveOptions{Force: true})

// 		stats, err := resources.MonitorResources(context.Background(), cli, containerID)
// 		if err != nil {
// 			return executionResult, fmt.Errorf("failed to monitor resources: %v", err)
// 		}

// 		if len(stats.Output) == 0 {
// 			return executionResult, fmt.Errorf("script output is empty")
// 		}

// 		logger.Infof("Script output: %s", stats.Output)

// 		if err := json.Unmarshal([]byte(stats.Output), &argData); err != nil {
// 			logger.Infof("Could not parse output as direct JSON, trying to extract value from payload format")

// 			lines := strings.Split(stats.Output, "\n")
// 			for _, line := range lines {
// 				if strings.Contains(line, "Payload received:") {
// 					payloadValue := strings.TrimSpace(strings.TrimPrefix(line, "Payload received:"))
// 					logger.Infof("Extracted payload value: %s", payloadValue)

// 					if err := json.Unmarshal([]byte(payloadValue), &argData); err != nil {
// 						argData = payloadValue
// 						logger.Infof("Using raw string value as argument")
// 						break
// 					} else {
// 						logger.Infof("Successfully parsed extracted value as JSON")
// 						break
// 					}
// 				}
// 			}

// 			if argData == nil {
// 				logger.Errorf("Failed to parse script output as arguments: %v", err)
// 				return executionResult, fmt.Errorf("failed to parse script output: %v", err)
// 			}
// 		}

// 		logger.Infof("Successfully processed script output data")

// 		executionResult.MemoryUsage = stats.MemoryUsage
// 		executionResult.CPUPercentage = stats.CPUPercentage
// 		executionResult.NetworkRx = stats.RxBytes
// 		executionResult.NetworkTx = stats.TxBytes
// 		executionResult.BlockRead = stats.BlockRead
// 		executionResult.BlockWrite = stats.BlockWrite
// 		executionResult.BandwidthRate = stats.BandwidthRate
// 		executionResult.TotalFee = stats.TotalFee
// 		executionResult.StaticComplexity = stats.StaticComplexity
// 		executionResult.DynamicComplexity = stats.DynamicComplexity
// 		executionResult.ComplexityIndex = stats.ComplexityIndex
// 	} else if len(job.Arguments) > 0 {
// 		if err := json.Unmarshal([]byte(job.Arguments[0]), &argData); err != nil {
// 			return executionResult, fmt.Errorf("failed to parse argument: %v", err)
// 		}
// 		logger.Infof("Successfully parsed JSON data from single argument")
// 	} else {
// 		return executionResult, fmt.Errorf("no script URL or arguments provided")
// 	}

// 	convertedArgs, err := e.processArguments(argData, method.Inputs, contractABI)
// 	if err != nil {
// 		return executionResult, fmt.Errorf("error processing arguments: %v", err)
// 	}

// 	var callData []byte
// 	callData, err = contractABI.Pack(method.Name, convertedArgs...)
// 	if err != nil {
// 		return executionResult, err
// 	}

// 	privateKey, err := crypto.HexToECDSA(config.PrivateKeyController)
// 	if err != nil {
// 		return executionResult, fmt.Errorf("failed to parse private key: %v", err)
// 	}

// 	nonce, err := e.ethClient.PendingNonceAt(context.Background(), ethcommon.HexToAddress(config.KeeperAddress))
// 	if err != nil {
// 		return executionResult, err
// 	}

// 	gasPrice, err := e.ethClient.SuggestGasPrice(context.Background())
// 	if err != nil {
// 		return executionResult, err
// 	}

// 	executionABI, err := abi.JSON(strings.NewReader(`[{"inputs":[{"name":"target","type":"address"},{"name":"data","type":"bytes"}],"name":"executeFunction","outputs":[],"stateMutability":"payable","type":"function"}]`))
// 	if err != nil {
// 		return executionResult, fmt.Errorf("failed to parse execution contract ABI: %v", err)
// 	}

// 	executionInput, err := executionABI.Pack("executeFunction", contractAddress, callData)
// 	if err != nil {
// 		return executionResult, fmt.Errorf("failed to pack execution contract input: %v", err)
// 	}

// 	tx := types.NewTransaction(nonce, ethcommon.HexToAddress(executionContractAddress), big.NewInt(0), 300000, gasPrice, executionInput)
// 	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(big.NewInt(11155420)), privateKey)
// 	if err != nil {
// 		return executionResult, err
// 	}

// 	e.logger.Infof("DEBUG: Creating transaction to execution contract: %s", executionContractAddress)

// 	err = e.ethClient.SendTransaction(context.Background(), signedTx)
// 	if err != nil {
// 		return executionResult, err
// 	}

// 	e.logger.Infof("DEBUG: Transaction sent to execution contract: %s, tx hash: %s",
// 		executionContractAddress, signedTx.Hash().Hex())

// 	receipt, err := bind.WaitMined(context.Background(), e.ethClient, signedTx)
// 	if err != nil {
// 		return executionResult, err
// 	}

// 	executionResult.Status = receipt.Status == types.ReceiptStatusSuccessful
// 	executionResult.ActionTxHash = signedTx.Hash().Hex()
// 	executionResult.GasUsed = strconv.FormatUint(receipt.GasUsed, 10)

// 	logger.Infof("✅ Job %d executed successfully with dynamic arguments. Transaction: %s",
// 		job.JobID, signedTx.Hash().Hex())

// 	return executionResult, nil
// }
