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

	// "github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
	jobtypes "github.com/trigg3rX/triggerx-backend/pkg/types"
)

// JobExecutor handles execution of blockchain transactions and contract calls
// Can be extended with additional configuration and dependencies as needed
type ArgumentConverter struct{}

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
	default:
		return nil, fmt.Errorf("unsupported type conversion: %v", targetType)
	}
}

func (ac *ArgumentConverter) convertToInteger(value interface{}, targetType abi.Type) (interface{}, error) {
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
	default:
		return false, fmt.Errorf("cannot convert type %T to bool", v)
	}
}

func (ac *ArgumentConverter) convertToAddress(value interface{}) (common.Address, error) {
	// Convert to Ethereum address
	switch v := value.(type) {
	case string:
		if !common.IsHexAddress(v) {
			return common.Address{}, fmt.Errorf("invalid Ethereum address: %s", v)
		}
		return common.HexToAddress(v), nil
	default:
		return common.Address{}, fmt.Errorf("cannot convert type %T to address", v)
	}
}

func (ac *ArgumentConverter) convertToBytes(value interface{}) ([]byte, error) {
	// Convert to bytes
	switch v := value.(type) {
	case string:
		return []byte(v), nil
	case []byte:
		return v, nil
	default:
		return nil, fmt.Errorf("cannot convert type %T to bytes", v)
	}
}

type JobExecutor struct {
	ethClient       *ethclient.Client
	etherscanAPIKey string
	argConverter    *ArgumentConverter
}

func NewJobExecutor(ethClient *ethclient.Client, etherscanAPIKey string) *JobExecutor {
	return &JobExecutor{
		ethClient:       ethClient,
		etherscanAPIKey: etherscanAPIKey,
		argConverter:    &ArgumentConverter{},
	}
}

// Execute routes jobs to appropriate handlers based on the target function
// Currently supports 'transfer' for token transfers and 'execute' for generic contract calls
func (e *JobExecutor) Execute(job *jobtypes.HandleCreateJobData) (jobtypes.ActionData, error) {
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
		Timestamp:    time.Now(),
	}

	logger.Infof("Executing contract call for job %s with static arguments", job.JobID)
	if job.TaskDefinitionID == 5 {

		if job.ScriptTriggerFunction != "" {
			satisfied, err := e.evaluateConditionScript(job.ScriptTriggerFunction)
			if err != nil {
				logger.Errorf("Failed to evaluate condition script: %v", err)
				return executionResult, fmt.Errorf("condition script evaluation failed: %v", err)
			}

			if !satisfied {
				logger.Infof("Condition not satisfied for job %d, skipping execution", job.JobID)
				return executionResult, nil // Return without error, but execution was skipped
			}
			logger.Infof("Condition satisfied for job %d, proceeding with execution", job.JobID)
		}
	}
	contractAddress := common.HexToAddress(job.TargetContractAddress)
	contractABI, method, err := e.getContractMethodAndABI(job.TargetFunction, job.TargetContractAddress)
	if err != nil {
		return executionResult, err
	}
	var convertedArgs []interface{}
	for i, inputParam := range method.Inputs {
		// Find corresponding argument from job
		argKey := inputParam.Name
		if argKey == "" {
			argKey = fmt.Sprintf("arg%d", i)
		}

		if i >= len(job.Arguments) {
			return executionResult, fmt.Errorf("missing argument at index %d", i)
		}
		argValue := job.Arguments[i]

		// Convert argument to required type
		convertedArg, err := e.argConverter.convertToType(argValue, inputParam.Type)
		if err != nil {
			return executionResult, fmt.Errorf("error converting argument %s: %v", argKey, err)
		}

		convertedArgs = append(convertedArgs, convertedArg)
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

	nonce, err := e.ethClient.PendingNonceAt(context.Background(), common.HexToAddress(config.KeeperAddress))
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
		Timestamp:    time.Now(),
	}

	logger.Infof("Executing job %d with dynamic arguments", job.JobID)

	// Step 1: Check if we need to evaluate a condition from the script
	if job.TaskDefinitionID == 6 {
		if job.ScriptTriggerFunction != "" {
			satisfied, err := e.evaluateConditionScript(job.ScriptTriggerFunction)
			if err != nil {
				logger.Errorf("Failed to evaluate condition script: %v", err)
				return executionResult, fmt.Errorf("condition script evaluation failed: %v", err)
			}

			if !satisfied {
				logger.Infof("Condition not satisfied for job %d, skipping execution", job.JobID)
				return executionResult, nil // Return without error, but execution was skipped
			}
			logger.Infof("Condition satisfied for job %d, proceeding with execution", job.JobID)
		}
	}

	// Step 2: Fetch dynamic arguments from IPFS if specified
	var dynamicArgs []string
	if job.ScriptIPFSUrl != "" {
		args, err := e.fetchArgumentsFromIPFS(job.ScriptIPFSUrl)
		if err != nil {
			logger.Errorf("Failed to fetch arguments from IPFS: %v", err)
			return executionResult, fmt.Errorf("failed to fetch arguments from IPFS: %v", err)
		}
		dynamicArgs = args
		logger.Infof("Successfully fetched %d arguments from IPFS: %v", len(dynamicArgs), dynamicArgs)
	} else {
		// If no script URL is provided, use the static arguments provided in the job
		dynamicArgs = job.Arguments
	}

	// Step 3: Prepare and execute the contract call with dynamic arguments
	contractAddress := common.HexToAddress(job.TargetContractAddress)
	contractABI, method, err := e.getContractMethodAndABI(job.TargetFunction, job.TargetContractAddress)
	if err != nil {
		logger.Errorf("Failed to get contract method and ABI: %v", err)
		return executionResult, err
	}

	// Convert arguments to the correct types based on the contract ABI
	var convertedArgs []interface{}
	for i, inputParam := range method.Inputs {
		if i >= len(dynamicArgs) {
			return executionResult, fmt.Errorf("missing dynamic argument at index %d", i)
		}

		// Convert argument to required type
		argValue := dynamicArgs[i]
		convertedArg, err := e.argConverter.convertToType(argValue, inputParam.Type)
		if err != nil {
			return executionResult, fmt.Errorf("error converting dynamic argument at index %d: %v", i, err)
		}

		convertedArgs = append(convertedArgs, convertedArg)
	}

	// Pack arguments for the contract call
	input, err := contractABI.Pack(method.Name, convertedArgs...)
	if err != nil {
		logger.Errorf("Error packing arguments: %v", err)
		return executionResult, err
	}

	// Create and send transaction
	privateKey, err := crypto.HexToECDSA(config.PrivateKeyController)
	if err != nil {
		return executionResult, fmt.Errorf("failed to parse private key: %v", err)
	}

	nonce, err := e.ethClient.PendingNonceAt(context.Background(), common.HexToAddress(config.KeeperAddress))
	if err != nil {
		return executionResult, err
	}

	gasPrice, err := e.ethClient.SuggestGasPrice(context.Background())
	if err != nil {
		return executionResult, err
	}

	// Create and sign transaction
	tx := types.NewTransaction(nonce, contractAddress, big.NewInt(0), 300000, gasPrice, input)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(big.NewInt(11155420)), privateKey)
	if err != nil {
		return executionResult, err
	}

	// Send transaction
	err = e.ethClient.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return executionResult, err
	}

	// Wait for transaction receipt
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
