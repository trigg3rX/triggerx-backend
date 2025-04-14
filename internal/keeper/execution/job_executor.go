package execution

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/validation"
	"github.com/trigg3rX/triggerx-backend/pkg/common"
	jobtypes "github.com/trigg3rX/triggerx-backend/pkg/types"

	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/trigg3rX/triggerx-backend/pkg/resources"
)

// ValidatorInterface defines the contract for job validation
type ValidatorInterface interface {
	ValidateTimeBasedJob(job *jobtypes.HandleCreateJobData) (bool, error)
	ValidateEventBasedJob(job *jobtypes.HandleCreateJobData, ipfsData *jobtypes.IPFSData) (bool, error)
	ValidateAndPrepareJob(job *jobtypes.HandleCreateJobData, triggerData *jobtypes.TriggerData) (bool, error)
}

// JobExecutor handles execution of blockchain transactions and contract calls
// Can be extended with additional configuration and dependencies as needed
type ArgumentConverter struct{}

// convertToType converts a value to the target ABI type
func (ac *ArgumentConverter) convertToType(value interface{}, targetType abi.Type) (interface{}, error) {
	// Handle different input types and convert to appropriate blockchain types
	switch targetType.T {
	case abi.UintTy, abi.IntTy:
		return ac.convertToInteger(value, targetType)
	case abi.StringTy:
		return ac.convertToString(value)
	case abi.BoolTy:
		return ac.convertToBool(value)
	case abi.AddressTy:
		return ac.convertToAddress(value)
	case abi.BytesTy, abi.FixedBytesTy:
		return ac.convertToBytes(value)
	case abi.ArrayTy, abi.SliceTy:
		return ac.convertToArray(value, targetType)
	case abi.TupleTy:
		return ac.convertToStruct(value, targetType)
	default:
		return nil, fmt.Errorf("unsupported type conversion: %v", targetType)
	}
}

func (ac *ArgumentConverter) convertToInteger(value interface{}, targetType abi.Type) (interface{}, error) {
	// Add this case to handle when the value is already a *big.Int
	if bigInt, ok := value.(*big.Int); ok {
		return bigInt, nil
	}

	switch targetType.T {
	case abi.UintTy:
		if targetType.Size == 32 {
			// Handle uint32 specifically
			switch v := value.(type) {
			case string:
				// Parse as float first, then convert to uint32
				floatVal, err := strconv.ParseFloat(v, 64)
				if err != nil {
					return nil, err
				}
				return uint32(floatVal), nil
			case float64:
				return uint32(v), nil
			case int:
				return uint32(v), nil
			}
		}
		// For other uint sizes, use big.Int
		fallthrough
	default:
		switch v := value.(type) {
		case float64:
			return big.NewInt(int64(v)), nil
		case string:
			// Parse as float first, then convert to big.Int
			floatVal, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return nil, fmt.Errorf("cannot convert string to integer: %v", err)
			}
			return big.NewInt(int64(floatVal)), nil
		case map[string]interface{}:
			// This could be a struct that we need to convert to an integer
			if jsonBytes, err := json.Marshal(v); err == nil {
				var floatVal float64
				if err := json.Unmarshal(jsonBytes, &floatVal); err == nil {
					return big.NewInt(int64(floatVal)), nil
				}
			}
			return nil, fmt.Errorf("cannot convert map to integer")
		default:
			return nil, fmt.Errorf("cannot convert type %T to integer", v)
		}
	}
}

func (ac *ArgumentConverter) convertToString(value interface{}) (string, error) {
	// Convert various types to string
	switch v := value.(type) {
	case string:
		return v, nil
	case float64, int, uint:
		return fmt.Sprintf("%v", v), nil
	case map[string]interface{}:
		// This could be a JSON object that we need to convert to a string
		if jsonBytes, err := json.Marshal(v); err == nil {
			return string(jsonBytes), nil
		}
		return "", fmt.Errorf("cannot convert map to string")
	default:
		return "", fmt.Errorf("cannot convert type %T to string", v)
	}
}

func (ac *ArgumentConverter) convertToBool(value interface{}) (bool, error) {
	// Convert various types to bool
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		return strconv.ParseBool(v)
	case float64:
		return v != 0, nil
	case map[string]interface{}:
		// Try to convert JSON to bool
		if jsonBytes, err := json.Marshal(v); err == nil {
			var boolVal bool
			if err := json.Unmarshal(jsonBytes, &boolVal); err == nil {
				return boolVal, nil
			}
		}
		return false, fmt.Errorf("cannot convert map to bool")
	default:
		return false, fmt.Errorf("cannot convert type %T to bool", v)
	}
}

func (ac *ArgumentConverter) convertToAddress(value interface{}) (ethcommon.Address, error) {
	// Convert to Ethereum address
	switch v := value.(type) {
	case string:
		if !ethcommon.IsHexAddress(v) {
			return ethcommon.Address{}, fmt.Errorf("invalid Ethereum address: %s", v)
		}
		return ethcommon.HexToAddress(v), nil
	case map[string]interface{}:
		// Check if we have a string representation in the map
		if addrStr, ok := v["address"].(string); ok {
			if !ethcommon.IsHexAddress(addrStr) {
				return ethcommon.Address{}, fmt.Errorf("invalid Ethereum address: %s", addrStr)
			}
			return ethcommon.HexToAddress(addrStr), nil
		}
		return ethcommon.Address{}, fmt.Errorf("cannot convert map to address")
	default:
		return ethcommon.Address{}, fmt.Errorf("cannot convert type %T to address", v)
	}
}

func (ac *ArgumentConverter) convertToBytes(value interface{}) ([]byte, error) {
	// Convert to bytes
	switch v := value.(type) {
	case string:
		// Check if it's a hex string
		if strings.HasPrefix(v, "0x") {
			return ethcommon.FromHex(v), nil
		}
		return []byte(v), nil
	case []byte:
		return v, nil
	case map[string]interface{}:
		// Try to convert JSON to bytes
		if jsonBytes, err := json.Marshal(v); err == nil {
			return jsonBytes, nil
		}
		return nil, fmt.Errorf("cannot convert map to bytes")
	default:
		return nil, fmt.Errorf("cannot convert type %T to bytes", v)
	}
}

func (ac *ArgumentConverter) convertToArray(value interface{}, targetType abi.Type) (interface{}, error) {
	// First, ensure the value is actually an array/slice
	var sourceArray []interface{}

	switch v := value.(type) {
	case []interface{}:
		sourceArray = v
	case string:
		// Try to parse as JSON array
		if err := json.Unmarshal([]byte(v), &sourceArray); err != nil {
			return nil, fmt.Errorf("failed to parse string as JSON array: %v", err)
		}
	case map[string]interface{}:
		// Try to parse map as JSON array
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal map as JSON: %v", err)
		}
		if err := json.Unmarshal(jsonBytes, &sourceArray); err != nil {
			return nil, fmt.Errorf("failed to parse map as JSON array: %v", err)
		}
	default:
		return nil, fmt.Errorf("cannot convert type %T to array/slice", v)
	}

	// Create a new slice with the correct element type
	sliceType := reflect.SliceOf(targetType.Elem.GetType())
	result := reflect.MakeSlice(sliceType, len(sourceArray), len(sourceArray))

	// Convert each element
	for i, elem := range sourceArray {
		convertedElem, err := ac.convertToType(elem, *targetType.Elem)
		if err != nil {
			return nil, fmt.Errorf("error converting array element %d: %v", i, err)
		}

		// Set the element in the slice
		resultElem := reflect.ValueOf(convertedElem)
		result.Index(i).Set(resultElem)
	}

	return result.Interface(), nil
}

func (ac *ArgumentConverter) convertToStruct(value interface{}, targetType abi.Type) (interface{}, error) {
	// Create a new instance of the struct type
	structType := targetType.GetType()
	structValue := reflect.New(structType).Elem()

	// Prepare source data
	var sourceMap map[string]interface{}

	switch v := value.(type) {
	case map[string]interface{}:
		sourceMap = v
	case string:
		// Try to parse as JSON object
		if err := json.Unmarshal([]byte(v), &sourceMap); err != nil {
			return nil, fmt.Errorf("failed to parse string as JSON object: %v", err)
		}
	default:
		// If it's already a struct, we can try to convert it directly
		valueVal := reflect.ValueOf(value)
		if valueVal.Kind() == reflect.Struct {
			// Convert struct to map for easier processing
			jsonBytes, err := json.Marshal(value)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal struct: %v", err)
			}
			if err := json.Unmarshal(jsonBytes, &sourceMap); err != nil {
				return nil, fmt.Errorf("failed to unmarshal struct to map: %v", err)
			}
		} else {
			return nil, fmt.Errorf("cannot convert type %T to struct", v)
		}
	}

	// Iterate through the tuple components and set corresponding fields
	for i, component := range targetType.TupleElems {
		fieldName := targetType.TupleRawNames[i]
		fieldValue, exists := sourceMap[fieldName]

		if !exists {
			// Try with case-insensitive match
			for k, v := range sourceMap {
				if strings.EqualFold(k, fieldName) {
					fieldValue = v
					exists = true
					break
				}
			}
		}

		if !exists {
			log.Printf("Warning: field %s not found in input data", fieldName)
			continue
		}

		// Convert the field value to the correct type
		convertedValue, err := ac.convertToType(fieldValue, *component)
		if err != nil {
			return nil, fmt.Errorf("error converting struct field %s: %v", fieldName, err)
		}

		// Find the corresponding field in the struct
		var structField reflect.Value
		for j := 0; j < structValue.NumField(); j++ {
			if strings.EqualFold(structType.Field(j).Name, fieldName) {
				structField = structValue.Field(j)
				break
			}
		}

		if !structField.IsValid() {
			return nil, fmt.Errorf("struct field %s not found", fieldName)
		}

		// Set the field value
		convertedValueReflect := reflect.ValueOf(convertedValue)
		if structField.Type() != convertedValueReflect.Type() {
			// Try to convert the value to the correct type
			if convertedValueReflect.Type().ConvertibleTo(structField.Type()) {
				convertedValueReflect = convertedValueReflect.Convert(structField.Type())
			} else {
				return nil, fmt.Errorf("cannot convert %v to field type %v", convertedValueReflect.Type(), structField.Type())
			}
		}
		structField.Set(convertedValueReflect)
	}

	return structValue.Interface(), nil
}

type JobExecutor struct {
	ethClient       common.EthClientInterface
	etherscanAPIKey string
	argConverter    *ArgumentConverter
	validator       common.ValidatorInterface
	logger          common.Logger
}

func NewJobExecutor(ethClient *ethclient.Client, etherscanAPIKey string) *JobExecutor {
	return &JobExecutor{
		ethClient:       ethClient,
		etherscanAPIKey: etherscanAPIKey,
		argConverter:    &ArgumentConverter{},
		validator:       validation.NewJobValidator(logger, ethClient),
		logger:          logger,
	}
}

// Execute routes jobs to appropriate handlers based on the target function
// Currently supports 'transfer' for token transfers and 'execute' for generic contract calls
func (e *JobExecutor) Execute(job *jobtypes.HandleCreateJobData) (jobtypes.ActionData, error) {

	executionResult := jobtypes.ActionData{
		TaskID:       0,
		ActionTxHash: "0x",
		GasUsed:      "0",
		Status:       false,
		Timestamp:    time.Now().UTC(),
	}

	e.logger.Infof("Validating job %d (taskDefID: %d) before execution", job.JobID, job.TaskDefinitionID)
	shouldExecute, err := e.validator.ValidateAndPrepareJob(job, nil)
	if err != nil {
		e.logger.Errorf("Job validation error: %v", err)
		return executionResult, fmt.Errorf("job validation failed: %v", err)
	}

	if !shouldExecute {
		e.logger.Infof("Job %d validation determined execution should be skipped", job.JobID)
		return executionResult, nil // Return without error, execution was skipped
	}

	e.logger.Infof("Job %d validated successfully, proceeding with execution", job.JobID)
	logger.Infof("Executing job: %d (Function: %s)", job.JobID, job.TargetFunction)

	switch job.TaskDefinitionID {
	case 1, 3, 5:
		return e.executeActionWithStaticArgs(job)
	case 2, 4, 6:
		return e.executeActionWithDynamicArgs(job)
	default:
		return jobtypes.ActionData{}, fmt.Errorf("unsupported task definition id: %d", job.TaskDefinitionID)
	}
}

func (e *JobExecutor) executeActionWithStaticArgs(job *jobtypes.HandleCreateJobData) (jobtypes.ActionData, error) {
	executionResult := jobtypes.ActionData{
		TaskID:       0,
		ActionTxHash: "0x",
		GasUsed:      "0",
		Status:       false,
		Timestamp:    time.Now().UTC(),
	}

	logger.Infof("Executing contract call for job %s with static arguments", job.JobID)

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
			logger.Errorf("Failed to download condition script: %v", err)
			return executionResult, fmt.Errorf("failed to download condition script: %v", err)
		}
		defer os.RemoveAll(filepath.Dir(codePath))

		// Create and execute container for condition script
		containerID, err := resources.CreateDockerContainer(context.Background(), cli, codePath)
		if err != nil {
			logger.Errorf("Failed to create container for condition script: %v", err)
			return executionResult, fmt.Errorf("failed to create container: %v", err)
		}
		defer cli.ContainerRemove(context.Background(), containerID, dockertypes.ContainerRemoveOptions{Force: true})

		// Monitor resources and get script output
		stats, err := resources.MonitorResources(context.Background(), cli, containerID)
		if err != nil {
			logger.Errorf("Failed to monitor condition script resources: %v", err)
			return executionResult, fmt.Errorf("failed to monitor resources: %v", err)
		}

		// Check if condition was satisfied based on script output
		if !stats.Status {
			logger.Infof("Condition not satisfied for job %d, skipping execution", job.JobID)
			return executionResult, nil
		}
		logger.Infof("Condition satisfied for job %d, proceeding with execution", job.JobID)

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
	contractABI, method, err := e.getContractMethodAndABI(job.TargetFunction, job.TargetContractAddress)
	if err != nil {
		return executionResult, err
	}

	// Handle args as potentially structured data
	convertedArgs, err := e.processArguments(job.Arguments, method.Inputs, contractABI)
	if err != nil {
		return executionResult, fmt.Errorf("error processing arguments: %v", err)
	}

	input, err := contractABI.Pack(method.Name, convertedArgs...)
	if err != nil {
		log.Printf("Error packing arguments: %v", err)
		return executionResult, err
	}

	// Create transaction data
	privateKey, err := crypto.HexToECDSA(config.PrivateKeyController)
	if err != nil {
		return executionResult, fmt.Errorf("failed to parse private key: %v", err)
	}

	nonce, err := e.ethClient.PendingNonceAt(context.Background(), ethcommon.HexToAddress(config.KeeperAddress))
	if err != nil {
		return executionResult, err
	}

	gasPrice, err := e.ethClient.SuggestGasPrice(context.Background())
	if err != nil {
		return executionResult, err
	}

	tx := types.NewTransaction(nonce, contractAddress, big.NewInt(0), 300000, gasPrice, input)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(big.NewInt(11155420)), privateKey)
	if err != nil {
		return executionResult, err
	}

	err = e.ethClient.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return executionResult, err
	}

	// Wait for transaction receipt
	receipt, err := bind.WaitMined(context.Background(), e.ethClient, signedTx)
	if err != nil {
		log.Printf("Error waiting for transaction: %v", err)
		return executionResult, err
	}

	executionResult.Status = receipt.Status == types.ReceiptStatusSuccessful
	executionResult.ActionTxHash = signedTx.Hash().Hex()
	executionResult.GasUsed = strconv.FormatUint(receipt.GasUsed, 10)

	log.Printf("✅ Job %d executed successfully. Transaction: %s", job.JobID, signedTx.Hash().Hex())

	return executionResult, nil
}

func (e *JobExecutor) executeActionWithDynamicArgs(job *jobtypes.HandleCreateJobData) (jobtypes.ActionData, error) {
	executionResult := jobtypes.ActionData{
		TaskID:       0,
		ActionTxHash: "0x",
		GasUsed:      "0",
		Status:       false,
		Timestamp:    time.Now().UTC(),
	}

	logger.Infof("Executing job %d with dynamic arguments", job.JobID)

	// Create Docker client for script execution
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return executionResult, fmt.Errorf("failed to create Docker client: %v", err)
	}
	defer cli.Close()

	// Step 1: Check if we need to evaluate a condition from the script
	if job.TaskDefinitionID == 6 && job.ScriptTriggerFunction != "" {
		// Download and execute the condition script
		codePath, err := resources.DownloadIPFSFile(job.ScriptTriggerFunction)
		if err != nil {
			logger.Errorf("Failed to download condition script: %v", err)
			return executionResult, fmt.Errorf("failed to download condition script: %v", err)
		}
		defer os.RemoveAll(filepath.Dir(codePath))

		// Create and execute container for condition script
		containerID, err := resources.CreateDockerContainer(context.Background(), cli, codePath)
		if err != nil {
			logger.Errorf("Failed to create container for condition script: %v", err)
			return executionResult, fmt.Errorf("failed to create container: %v", err)
		}
		defer cli.ContainerRemove(context.Background(), containerID, dockertypes.ContainerRemoveOptions{Force: true})

		// Monitor resources and get script output
		stats, err := resources.MonitorResources(context.Background(), cli, containerID)
		if err != nil {
			logger.Errorf("Failed to monitor condition script resources: %v", err)
			return executionResult, fmt.Errorf("failed to monitor resources: %v", err)
		}

		// Check if condition was satisfied based on script output
		if !stats.Status {
			logger.Infof("Condition not satisfied for job %d, skipping execution", job.JobID)
			return executionResult, nil
		}
		logger.Infof("Condition satisfied for job %d, proceeding with execution", job.JobID)

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

	// Step 2: Get the contract method and ABI
	contractAddress := ethcommon.HexToAddress(job.TargetContractAddress)
	contractABI, method, err := e.getContractMethodAndABI(job.TargetFunction, job.TargetContractAddress)
	if err != nil {
		logger.Errorf("Failed to get contract method and ABI: %v", err)
		return executionResult, err
	}

	// Step 3: Execute the script to get dynamic arguments if ScriptIPFSUrl is provided
	var argData interface{}
	if job.ScriptIPFSUrl != "" {
		// Download and execute the script
		codePath, err := resources.DownloadIPFSFile(job.ScriptIPFSUrl)
		if err != nil {
			logger.Errorf("Failed to download script: %v", err)
			return executionResult, fmt.Errorf("failed to download script: %v", err)
		}
		defer os.RemoveAll(filepath.Dir(codePath))

		// Create and execute container for script
		containerID, err := resources.CreateDockerContainer(context.Background(), cli, codePath)
		if err != nil {
			logger.Errorf("Failed to create container: %v", err)
			return executionResult, fmt.Errorf("failed to create container: %v", err)
		}
		defer cli.ContainerRemove(context.Background(), containerID, dockertypes.ContainerRemoveOptions{Force: true})

		// Monitor resources and get script output
		stats, err := resources.MonitorResources(context.Background(), cli, containerID)
		if err != nil {
			logger.Errorf("Failed to monitor script resources: %v", err)
			return executionResult, fmt.Errorf("failed to monitor resources: %v", err)
		}

		// Parse the script output
		if len(stats.Output) == 0 {
			logger.Errorf("Script output is empty")
			return executionResult, fmt.Errorf("script output is empty")
		}

		// Try to parse the output as JSON
		if err := json.Unmarshal([]byte(stats.Output), &argData); err != nil {
			logger.Errorf("Failed to parse script output as arguments: %v", err)
			return executionResult, fmt.Errorf("failed to parse script output: %v", err)
		}

		logger.Infof("Successfully parsed JSON data from script output")

		// Update execution result with resource usage from script
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
	} else if len(job.Arguments) > 0 {
		// If no script URL but arguments are provided, try to parse the first argument as JSON
		if err := json.Unmarshal([]byte(job.Arguments[0]), &argData); err != nil {
			logger.Errorf("Failed to parse argument as JSON: %v", err)
			return executionResult, fmt.Errorf("failed to parse argument: %v", err)
		}
		logger.Infof("Successfully parsed JSON data from single argument")
	} else {
		logger.Errorf("No script URL or arguments provided")
		return executionResult, fmt.Errorf("no script URL or arguments provided")
	}

	// Step 4: Process the arguments
	convertedArgs, err := e.processArguments(argData, method.Inputs, contractABI)
	if err != nil {
		logger.Errorf("Error processing arguments: %v", err)
		return executionResult, fmt.Errorf("error processing arguments: %v", err)
	}

	// Step 5: Pack the arguments
	input, err := contractABI.Pack(method.Name, convertedArgs...)
	if err != nil {
		logger.Errorf("Error packing arguments: %v", err)
		return executionResult, err
	}

	// Step 6: Create and send transaction
	privateKey, err := crypto.HexToECDSA(config.PrivateKeyController)
	if err != nil {
		logger.Errorf("Failed to parse private key: %v", err)
		return executionResult, fmt.Errorf("failed to parse private key: %v", err)
	}

	nonce, err := e.ethClient.PendingNonceAt(context.Background(), ethcommon.HexToAddress(config.KeeperAddress))
	if err != nil {
		logger.Errorf("Failed to get nonce: %v", err)
		return executionResult, err
	}

	gasPrice, err := e.ethClient.SuggestGasPrice(context.Background())
	if err != nil {
		logger.Errorf("Failed to get gas price: %v", err)
		return executionResult, err
	}

	// Create and sign transaction
	tx := types.NewTransaction(nonce, contractAddress, big.NewInt(0), 300000, gasPrice, input)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(big.NewInt(11155420)), privateKey)
	if err != nil {
		logger.Errorf("Failed to sign transaction: %v", err)
		return executionResult, err
	}

	// Send transaction
	err = e.ethClient.SendTransaction(context.Background(), signedTx)
	if err != nil {
		logger.Errorf("Failed to send transaction: %v", err)
		return executionResult, err
	}

	// Step 7: Wait for transaction receipt
	receipt, err := bind.WaitMined(context.Background(), e.ethClient, signedTx)
	if err != nil {
		logger.Errorf("Error waiting for transaction: %v", err)
		return executionResult, err
	}

	// Update execution result with transaction details
	executionResult.Status = receipt.Status == types.ReceiptStatusSuccessful
	executionResult.ActionTxHash = signedTx.Hash().Hex()
	executionResult.GasUsed = strconv.FormatUint(receipt.GasUsed, 10)

	logger.Infof("✅ Job %d executed successfully with dynamic arguments. Transaction: %s",
		job.JobID, signedTx.Hash().Hex())

	return executionResult, nil
}

func (e *JobExecutor) executeGoScript(scriptContent string) (string, error) {
	// Create a temporary file for the script
	tempFile, err := ioutil.TempFile("", "script-*.go")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.Write([]byte(scriptContent)); err != nil {
		return "", fmt.Errorf("failed to write script to file: %v", err)
	}
	if err := tempFile.Close(); err != nil {
		return "", fmt.Errorf("failed to close temporary file: %v", err)
	}

	// Create a temp directory for the script's build output
	tempDir, err := ioutil.TempDir("", "script-build")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary build directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Compile the script
	outputBinary := filepath.Join(tempDir, "script")
	cmd := exec.Command("go", "build", "-o", outputBinary, tempFile.Name())
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to compile script: %v, stderr: %s", err, stderr.String())
	}

	// Run the compiled script
	result := exec.Command(outputBinary)
	stdout, err := result.Output()
	if err != nil {
		return "", fmt.Errorf("failed to run script: %v", err)
	}

	return string(stdout), nil
}

func (e *JobExecutor) processArguments(args interface{}, methodInputs []abi.Argument, contractABI *abi.ABI) ([]interface{}, error) {
	convertedArgs := make([]interface{}, 0)

	// Handle the case where we have a single struct argument
	if len(methodInputs) == 1 && methodInputs[0].Type.T == abi.TupleTy {
		// Check if the input is a map or JSON string representing the struct
		switch v := args.(type) {
		case map[string]interface{}:
			// Direct map to struct conversion
			convertedArg, err := e.argConverter.convertToStruct(v, methodInputs[0].Type)
			if err != nil {
				return nil, fmt.Errorf("error converting to struct: %v", err)
			}
			convertedArgs = append(convertedArgs, convertedArg)
			return convertedArgs, nil
		case string:
			// Try to parse as JSON struct
			var structData map[string]interface{}
			if err := json.Unmarshal([]byte(v), &structData); err == nil {
				convertedArg, err := e.argConverter.convertToStruct(structData, methodInputs[0].Type)
				if err != nil {
					return nil, fmt.Errorf("error converting JSON string to struct: %v", err)
				}
				convertedArgs = append(convertedArgs, convertedArg)
				return convertedArgs, nil
			}
		case []interface{}:
			// If there's a single array element and it's a map, try to use it as a struct
			if len(v) == 1 {
				if mapVal, ok := v[0].(map[string]interface{}); ok {
					convertedArg, err := e.argConverter.convertToStruct(mapVal, methodInputs[0].Type)
					if err != nil {
						return nil, fmt.Errorf("error converting map from array to struct: %v", err)
					}
					convertedArgs = append(convertedArgs, convertedArg)
					return convertedArgs, nil
				} else if strVal, ok := v[0].(string); ok {
					// Try to parse as JSON struct
					var structData map[string]interface{}
					if err := json.Unmarshal([]byte(strVal), &structData); err == nil {
						convertedArg, err := e.argConverter.convertToStruct(structData, methodInputs[0].Type)
						if err != nil {
							return nil, fmt.Errorf("error converting JSON string to struct: %v", err)
						}
						convertedArgs = append(convertedArgs, convertedArg)
						return convertedArgs, nil
					}
				}
			}
		}
	}

	// Handle multiple arguments or non-struct arguments
	switch argData := args.(type) {
	case []string:
		// Handle simple string array
		if len(argData) < len(methodInputs) {
			return nil, fmt.Errorf("not enough arguments provided: expected %d, got %d",
				len(methodInputs), len(argData))
		}

		for i, inputParam := range methodInputs {
			convertedArg, err := e.argConverter.convertToType(argData[i], inputParam.Type)
			if err != nil {
				return nil, fmt.Errorf("error converting argument %d: %v", i, err)
			}
			convertedArgs = append(convertedArgs, convertedArg)
		}
	case []interface{}:
		// Handle array of mixed types
		if len(argData) < len(methodInputs) {
			return nil, fmt.Errorf("not enough arguments provided: expected %d, got %d",
				len(methodInputs), len(argData))
		}

		for i, inputParam := range methodInputs {
			convertedArg, err := e.argConverter.convertToType(argData[i], inputParam.Type)
			if err != nil {
				return nil, fmt.Errorf("error converting argument %d: %v", i, err)
			}
			convertedArgs = append(convertedArgs, convertedArg)
		}
	case map[string]interface{}:
		// Handle map of named arguments
		for _, inputParam := range methodInputs {
			paramName := inputParam.Name
			if paramName == "" {
				return nil, fmt.Errorf("cannot use map arguments with unnamed parameters")
			}

			argValue, exists := argData[paramName]
			if !exists {
				// Try with case-insensitive match
				for k, v := range argData {
					if strings.EqualFold(k, paramName) {
						argValue = v
						exists = true
						break
					}
				}

				if !exists {
					return nil, fmt.Errorf("argument %s not found in input data", paramName)
				}
			}

			convertedArg, err := e.argConverter.convertToType(argValue, inputParam.Type)
			if err != nil {
				return nil, fmt.Errorf("error converting argument %s: %v", paramName, err)
			}
			convertedArgs = append(convertedArgs, convertedArg)
		}
	default:
		return nil, fmt.Errorf("unsupported argument format: %T", args)
	}

	return convertedArgs, nil
}

func (e *JobExecutor) evaluateConditionScript(scriptUrl string) (bool, error) {
	// Fetch script content from IPFS
	scriptContent, err := e.fetchFromIPFS(scriptUrl)
	if err != nil {
		return false, fmt.Errorf("failed to fetch condition script: %v", err)
	}

	// Create a temporary file for the script
	tempFile, err := ioutil.TempFile("", "condition-*.go")
	if err != nil {
		return false, fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.Write([]byte(scriptContent)); err != nil {
		return false, fmt.Errorf("failed to write script to file: %v", err)
	}
	if err := tempFile.Close(); err != nil {
		return false, fmt.Errorf("failed to close temporary file: %v", err)
	}

	// Create a temp directory for the script's build output
	tempDir, err := ioutil.TempDir("", "condition-build")
	if err != nil {
		return false, fmt.Errorf("failed to create temporary build directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Compile the script
	outputBinary := filepath.Join(tempDir, "condition")
	cmd := exec.Command("go", "build", "-o", outputBinary, tempFile.Name())
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("failed to compile condition script: %v, stderr: %s", err, stderr.String())
	}

	// Run the compiled script
	result := exec.Command(outputBinary)
	stdout, err := result.Output()
	if err != nil {
		return false, fmt.Errorf("failed to run condition script: %v", err)
	}

	// Parse the output to determine if condition is satisfied
	// Look for a line containing "Condition satisfied: true" or "Condition satisfied: false"
	lines := strings.Split(string(stdout), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Condition satisfied: true") {
			return true, nil
		} else if strings.Contains(line, "Condition satisfied: false") {
			return false, nil
		}
	}

	// If no explicit condition found, try parsing as JSON
	var conditionResult struct {
		Satisfied bool `json:"satisfied"`
	}
	if err := json.Unmarshal(stdout, &conditionResult); err != nil {
		return false, fmt.Errorf("could not determine condition result from output: %s", string(stdout))
	}

	return conditionResult.Satisfied, nil
}

func (e *JobExecutor) fetchArgumentsFromIPFS(scriptIPFSUrl string) ([]string, error) {
	// Fetch script content from IPFS
	scriptContent, err := e.fetchFromIPFS(scriptIPFSUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch arguments script: %v", err)
	}

	// Create a temporary file for the script
	tempFile, err := ioutil.TempFile("", "args-*.go")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.Write([]byte(scriptContent)); err != nil {
		return nil, fmt.Errorf("failed to write script to file: %v", err)
	}
	if err := tempFile.Close(); err != nil {
		return nil, fmt.Errorf("failed to close temporary file: %v", err)
	}

	// Create a temp directory for the script's build output
	tempDir, err := ioutil.TempDir("", "args-build")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary build directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Compile the script
	outputBinary := filepath.Join(tempDir, "args")
	cmd := exec.Command("go", "build", "-o", outputBinary, tempFile.Name())
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to compile args script: %v, stderr: %s", err, stderr.String())
	}

	// Run the compiled script
	result := exec.Command(outputBinary)
	stdout, err := result.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run args script: %v", err)
	}

	// Parse the output to get the arguments
	// First try parsing as JSON array
	var jsonOutput []string
	if err := json.Unmarshal(stdout, &jsonOutput); err == nil {
		return jsonOutput, nil
	}

	// If JSON parsing fails, look for checker function payload format
	lines := strings.Split(string(stdout), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Payload received:") {
			payload := strings.TrimSpace(strings.TrimPrefix(line, "Payload received:"))
			return []string{payload}, nil
		}
	}

	// If no structured format is found, use the entire output as a single argument
	return []string{string(stdout)}, nil
}

func (e *JobExecutor) fetchFromIPFS(url string) (string, error) {
	// Convert IPFS URL to gateway URL if needed
	gatewayURL := url
	if strings.HasPrefix(url, "ipfs://") {
		cid := strings.TrimPrefix(url, "ipfs://")
		gatewayURL = fmt.Sprintf("https://ipfs.io/ipfs/%s", cid)
	}

	// Fetch the content
	resp, err := http.Get(gatewayURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch from IPFS: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("IPFS fetch failed with status code: %d", resp.StatusCode)
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read IPFS response: %v", err)
	}

	return string(content), nil
}

func (e *JobExecutor) getContractMethodAndABI(methodName, contractAddress string) (*abi.ABI, *abi.Method, error) {
	// Fetch ABI
	abiData, err := e.fetchContractABI(contractAddress)
	if err != nil {
		return nil, nil, err
	}

	parsed, err := abi.JSON(bytes.NewReader(abiData))
	if err != nil {
		log.Printf("Error parsing ABI: %v", err)
		return nil, nil, err
	}

	log.Printf("Fetched ABI for contract %s: %s", contractAddress, string(abiData))

	method, ok := parsed.Methods[methodName]
	if !ok {
		log.Printf("Method %s not found in contract ABI", methodName)
		return nil, nil, fmt.Errorf("method %s not found in contract ABI", methodName)
	}

	log.Printf("Found method: %+v", method)
	return &parsed, &method, nil
}

func (e *JobExecutor) decodeContractOutput(contractABI *abi.ABI, method *abi.Method, output []byte) (interface{}, error) {
	// Handle different output scenarios
	if len(method.Outputs) == 0 {
		log.Printf("Method %s has no outputs to decode", method.Name)
		return nil, nil
	}

	// Single output case
	if len(method.Outputs) == 1 {
		outputType := method.Outputs[0]
		result := reflect.New(outputType.Type.GetType()).Elem()

		err := contractABI.UnpackIntoInterface(result.Addr().Interface(), method.Name, output)
		if err != nil {
			log.Printf("Error unpacking single output: %v", err)
			return nil, err
		}

		log.Printf("Decoded single output: %v", result.Interface())
		return result.Interface(), nil
	}

	// Multiple outputs case
	results := make([]interface{}, len(method.Outputs))
	err := contractABI.UnpackIntoInterface(&results, method.Name, output)
	if err != nil {
		log.Printf("Error unpacking multiple outputs: %v", err)
		return nil, err
	}

	log.Printf("Decoded multiple outputs: %+v", results)
	return results, nil
}

func (e *JobExecutor) fetchContractABI(contractAddress string) ([]byte, error) {
	if e.etherscanAPIKey == "" {
		return nil, fmt.Errorf("missing Etherscan API key")
	}

	// Update the URL to use Optimism Sepolia's API endpoint
	blockscoutUrl := fmt.Sprintf(
		"https://optimism-sepolia.blockscout.com/api?module=contract&action=getabi&address=%s",
		contractAddress)

	resp, err := http.Get(blockscoutUrl)
	if err != nil || resp.StatusCode != http.StatusOK {
		logger.Warnf("Failed to fetch ABI from Blockscout: %v", err)
		// Fall back to another source or handle accordingly
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  string `json:"result"`
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	if response.Status != "1" {
		return nil, fmt.Errorf("error fetching contract ABI: %s", response.Message)
	}

	return []byte(response.Result), nil
}
