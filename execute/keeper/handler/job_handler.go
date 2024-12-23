package handler

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

	shell "github.com/ipfs/go-ipfs-api"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/trigg3rX/triggerx-backend/execute/keeper/executor"
	"github.com/trigg3rX/triggerx-backend/execute/manager"
)

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
	// Convert various numeric types to big.Int or big.Rat
	switch v := value.(type) {
	case float64:
		if targetType.Size > 64 {
			// For larger integers, use big.Rat or big.Int
			bigRat := new(big.Rat).SetFloat64(v)
			return bigRat.Num(), nil
		}
		return big.NewInt(int64(v)), nil
	case int, int64, uint, uint64:
		// Convert to big.Int
		return big.NewInt(reflect.ValueOf(v).Convert(reflect.TypeOf(int64(0))).Int()), nil
	case string:
		// Try parsing string to int
		intVal, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("cannot convert string to integer: %v", err)
		}
		return big.NewInt(intVal), nil
	default:
		return nil, fmt.Errorf("cannot convert type %T to integer", v)
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

type JobHandler struct {
	executor        *executor.JobExecutor
	ethClient       *ethclient.Client
	etherscanAPIKey string
	argConverter    *ArgumentConverter
	ipfsShell       *shell.Shell
}

func NewJobHandler(ethClient *ethclient.Client, etherscanAPIKey string) *JobHandler {
	return &JobHandler{
		executor:        executor.NewJobExecutor(),
		ethClient:       ethClient,
		etherscanAPIKey: etherscanAPIKey,
		argConverter:    &ArgumentConverter{},
		ipfsShell:       shell.NewShell("localhost:5001"),
	}
}

func (h *JobHandler) HandleJob(job *manager.Job) error {
	if h.ethClient == nil {
		return fmt.Errorf("ethereum client not initialized")
	}

	log.Printf("🔧 Received job %s for execution", job.JobID)

	// Validate job
	if err := h.validateJob(job); err != nil {
		log.Printf("❌ Job validation failed: %v", err)
		return err
	}

	// Fetch checker_template.go from IPFS if CodeURL is provided
	checkerPath, err := h.fetchCheckerTemplateFromIPFS(job.CodeURL)
	if err != nil {
		log.Printf("Error fetching checker template: %v", err)
		return err
	}

	// Cleanup - remove the temporary checker file at the end
	defer func() {
		if checkerPath != "" {
			os.Remove(checkerPath)
		}
	}()

	// Execute job based on argument type
	switch job.ArgType {
	case "None":
		return h.executeNoArgContract(job)
	case "Static":
		return h.executeStaticArgContract(job)
	case "Dynamic":
		return h.executeDynamicArgContract(job)
	default:
		return fmt.Errorf("unsupported argument type: %s", job.ArgType)
	}
}

func (h *JobHandler) fetchCheckerTemplateFromIPFS(codeURL string) (string, error) {
	if codeURL == "" {
		log.Println("No CodeURL provided, skipping IPFS fetch")
		return "", nil
	}

	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot find home directory: %v", err)
	}

	// Define output path for the checker template
	checkerPath := filepath.Join(homeDir, "checker_template.go")

	// Remove existing file if it exists
	if _, err := os.Stat(checkerPath); err == nil {
		if err := os.Remove(checkerPath); err != nil {
			return "", fmt.Errorf("failed to remove existing file: %v", err)
		}
	}

	// Construct IPFS gateway URL and fetch the file
	ipfsUrl := codeURL
	cmd := exec.Command("wget", "-q", "-O", checkerPath, ipfsUrl)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to fetch file from IPFS: %v\nOutput: %s", err, string(output))
	}

	// fetchedCode, err := ioutil.ReadFile(checkerPath)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to read fetched code: %v", err)
	// }
	// log.Printf("Fetched Checker Template Code:\n%s", string(fetchedCode))
	fmt.Println("Fetched Checker Template Code")
	return checkerPath, nil
}

func (h *JobHandler) executeNoArgContract(job *manager.Job) error {
	log.Printf("Executing contract call for job %s with no arguments", job.JobID)

	// Check IPFS code validation if CodeURL is provided
	if job.CodeURL != "" {
		// Fetch checker template from IPFS
		checkerPath, err := h.fetchCheckerTemplateFromIPFS(job.CodeURL)
		if err != nil {
			return fmt.Errorf("failed to fetch checker template: %v", err)
		}
		defer os.Remove(checkerPath)

		// Compile the checker template
		compiledPath := checkerPath + ".bin"
		buildCmd := exec.Command("go", "build", "-o", compiledPath, checkerPath)
		if output, err := buildCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("	: %v\nOutput: %s", err, string(output))
		}
		defer os.Remove(compiledPath)

		// Execute the compiled checker
		runCmd := exec.Command(compiledPath)
		output, err := runCmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to execute checker: %v\nOutput: %s", err, string(output))
		}

		// Parse the checker output
		var result struct {
			Success bool                   `json:"success"`
			Payload map[string]interface{} `json:"payload"`
		}
		if err := json.Unmarshal(output, &result); err != nil {
			return fmt.Errorf("failed to parse checker output: %v", err)
		}

		if !result.Success {
			return fmt.Errorf("checker validation failed")
		}

		// Note: We don't modify arguments here since it's a no-arg contract
	}

	contractAddress := common.HexToAddress(job.ContractAddress)

	// Prepare method call data
	contractABI, method, err := h.getContractMethodAndABI(job.TargetFunction, job.ContractAddress)
	if err != nil {
		return err
	}

	// Encode method call data
	input, err := contractABI.Pack(method.Name)
	if err != nil {
		log.Printf("Error packing input for method %s: %v", method.Name, err)
		return err
	}
	log.Printf("Packed input for method %s: %x", method.Name, input)

	// Perform contract call
	callResult, err := h.ethClient.CallContract(context.Background(), ethereum.CallMsg{
		To:   &contractAddress,
		Data: input,
	}, nil)
	if err != nil {
		log.Printf("Error calling contract: %v", err)
		return err
	}

	// Decode the result
	decodedResults, err := h.decodeContractOutput(contractABI, method, callResult)
	if err != nil {
		log.Printf("Error decoding contract output: %v", err)
		return err
	}

	log.Printf("✅ Job %s executed successfully. Result: %+v", job.JobID, map[string]interface{}{
		"arguments": job.Arguments,
		"chainID":   job.ChainID,
		"contract":  job.ContractAddress,
		"result":    decodedResults,
		"status":    "success",
	})

	return nil
}

func (h *JobHandler) executeStaticArgContract(job *manager.Job) error {
	log.Printf("Executing contract call for job %s with static arguments", job.JobID)

	// Check IPFS code validation if CodeURL is provided
	if job.CodeURL != "" {
		// Fetch checker template from IPFS
		checkerPath, err := h.fetchCheckerTemplateFromIPFS(job.CodeURL)
		if err != nil {
			return fmt.Errorf("failed to fetch checker template: %v", err)
		}
		defer os.Remove(checkerPath)

		// Compile the checker template
		compiledPath := checkerPath + ".bin"
		buildCmd := exec.Command("go", "build", "-o", compiledPath, checkerPath)
		if output, err := buildCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to compile checker: %v\nOutput: %s", err, string(output))
		}
		defer os.Remove(compiledPath)

		// Execute the compiled checker
		runCmd := exec.Command(compiledPath)
		output, err := runCmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to execute checker: %v\nOutput: %s", err, string(output))
		}

		// Parse the checker output
		var result struct {
			Success bool                   `json:"success"`
			Payload map[string]interface{} `json:"payload"`
		}
		if err := json.Unmarshal(output, &result); err != nil {
			return fmt.Errorf("failed to parse checker output: %v", err)
		}

		if !result.Success {
			return fmt.Errorf("checker validation failed")
		}

		// Note: We don't modify arguments here since they are static
	}

	contractAddress := common.HexToAddress(job.ContractAddress)

	// Prepare method call data
	contractABI, method, err := h.getContractMethodAndABI(job.TargetFunction, job.ContractAddress)
	if err != nil {
		return err
	}

	// Convert arguments to match method input types
	var convertedArgs []interface{}
	for i, inputParam := range method.Inputs {
		// Find corresponding argument from job
		argKey := inputParam.Name
		if argKey == "" {
			argKey = fmt.Sprintf("arg%d", i)
		}

		argValue, exists := job.Arguments[argKey]
		if !exists {
			return fmt.Errorf("missing argument for parameter %s", argKey)
		}

		// Convert argument to required type
		convertedArg, err := h.argConverter.convertToType(argValue, inputParam.Type)
		if err != nil {
			return fmt.Errorf("error converting argument %s: %v", argKey, err)
		}

		convertedArgs = append(convertedArgs, convertedArg)
	}

	// Encode method call data with arguments
	input, err := contractABI.Pack(method.Name, convertedArgs...)
	if err != nil {
		log.Printf("Error packing arguments: %v", err)
		return err
	}

	// Perform contract call
	callResult, err := h.ethClient.CallContract(context.Background(), ethereum.CallMsg{
		To:   &contractAddress,
		Data: input,
	}, nil)
	if err != nil {
		log.Printf("Contract call error: %v", err)
		return err
	}

	// Decode the result
	decodedResults, err := h.decodeContractOutput(contractABI, method, callResult)
	if err != nil {
		log.Printf("Error decoding contract output: %v", err)
		return err
	}

	log.Printf("✅ Job %s executed successfully. Result: %+v", job.JobID, map[string]interface{}{
		"arguments": job.Arguments,
		"chainID":   job.ChainID,
		"contract":  job.ContractAddress,
		"result":    decodedResults,
		"status":    "success",
	})
	return nil
}

// Only showing the modified parts of job_handler.go for brevity

func (h *JobHandler) executeDynamicArgContract(job *manager.Job) error {
	log.Printf("Executing contract call for job %s with dynamic arguments", job.JobID)

	// Fetch and compile the checker template if CodeURL is provided
	var checkerResult map[string]interface{}
	if job.CodeURL != "" {
		// Fetch checker template from IPFS
		checkerPath, err := h.fetchCheckerTemplateFromIPFS(job.CodeURL)
		if err != nil {
			return fmt.Errorf("failed to fetch checker template: %v", err)
		}
		defer os.Remove(checkerPath)

		// Compile the checker template
		compiledPath := checkerPath + ".bin"
		buildCmd := exec.Command("go", "build", "-o", compiledPath, checkerPath)
		if output, err := buildCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to compile checker: %v\nOutput: %s", err, string(output))
		}
		defer os.Remove(compiledPath)

		// Execute the compiled checker
		runCmd := exec.Command(compiledPath)
		output, err := runCmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to execute checker: %v\nOutput: %s", err, string(output))
		}

		// Parse the checker output
		var result struct {
			Success bool                   `json:"success"`
			Payload map[string]interface{} `json:"payload"`
		}
		if err := json.Unmarshal(output, &result); err != nil {
			return fmt.Errorf("failed to parse checker output: %v", err)
		}

		if !result.Success {
			return fmt.Errorf("checker validation failed")
		}

		checkerResult = result.Payload

		// Merge checker results into job arguments
		for k, v := range checkerResult {
			job.Arguments[k] = v
		}
	}

	// Continue with contract execution
	contractAddress := common.HexToAddress(job.ContractAddress)
	contractABI, method, err := h.getContractMethodAndABI(job.TargetFunction, job.ContractAddress)
	if err != nil {
		return err
	}

	// Convert and pack arguments
	var convertedArgs []interface{}
	for i, inputParam := range method.Inputs {
		argKey := inputParam.Name
		if argKey == "" {
			argKey = fmt.Sprintf("arg%d", i)
		}

		argValue, exists := job.Arguments[argKey]
		if !exists {
			return fmt.Errorf("missing argument for parameter %s", argKey)
		}

		convertedArg, err := h.argConverter.convertToType(argValue, inputParam.Type)
		if err != nil {
			return fmt.Errorf("error converting argument %s: %v", argKey, err)
		}

		convertedArgs = append(convertedArgs, convertedArg)
	}

	// Pack arguments and make contract call
	input, err := contractABI.Pack(method.Name, convertedArgs...)
	if err != nil {
		return fmt.Errorf("error packing arguments: %v", err)
	}

	callResult, err := h.ethClient.CallContract(context.Background(), ethereum.CallMsg{
		To:   &contractAddress,
		Data: input,
	}, nil)
	if err != nil {
		return fmt.Errorf("contract call error: %v", err)
	}

	// Decode and log results
	decodedResults, err := h.decodeContractOutput(contractABI, method, callResult)
	if err != nil {
		return fmt.Errorf("error decoding contract output: %v", err)
	}

	log.Printf("✅ Job %s executed successfully. Result: %+v", job.JobID, map[string]interface{}{
		"arguments": job.Arguments,
		"chainID":   job.ChainID,
		"contract":  job.ContractAddress,
		"result":    decodedResults,
		"status":    "success",
	})

	return nil
}

// func (h *JobHandler) loadCheckerFunction(job *manager.Job) (func(*manager.Job) (bool, map[string]interface{}), error) {
//     log.Printf("Loading checker function for job %s", job.JobID)

//     // If no code URL is provided, return a default checker
//     if job.CodeURL == "" {
//         log.Printf("No CodeURL provided, using default checker")
//         return h.defaultChecker(job.ArgType), nil
//     }

//     // Fetch checker template from IPFS and get the file path
//     checkerPath, err := h.fetchCheckerTemplateFromIPFS(job.CodeURL)
//     if err != nil {
//         return nil, fmt.Errorf("failed to fetch checker template: %v", err)
//     }

//     // Cleanup - remove the compiled binary at the end
//     compiledBinaryPath := "/tmp/checker_template"
//     defer func() {
//         if _, err := os.Stat(compiledBinaryPath); err == nil {
//             os.Remove(compiledBinaryPath)
//         }
//     }()

//     // Compile the checker template using the path
//     cmd := exec.Command("go", "build", "-o", compiledBinaryPath, checkerPath)
//     output, err := cmd.CombinedOutput()
//     if err != nil {
//         return nil, fmt.Errorf("failed to compile checker template: %v, output: %s", err, string(output))
//     }

//     // Depending on the argument type, return appropriate checker function
//     switch job.ArgType {
//     case "None", "Static":
//         return func(j *manager.Job) (bool, map[string]interface{}) {
//             // Simple time interval check for static and no-arg jobs
//             if h.intervalChecker.ValidateJobInterval(j) {
//                 return true, make(map[string]interface{})
//             }
//             return false, make(map[string]interface{})
//         }, nil
//     case "Dynamic":
//         // For dynamic jobs, load payload directly from the compiled template
//         return func(j *manager.Job) (bool, map[string]interface{}) {
//             // Run the compiled checker to get payload
//             cmd := exec.Command(compiledBinaryPath)
//             output, err := cmd.CombinedOutput()
//             if err != nil {
//                 log.Printf("Error running checker template: %v", err)
//                 return false, make(map[string]interface{})
//             }

//             // Parse the output
//             var payload map[string]interface{}
//             err = json.Unmarshal(output, &payload)
//             if err != nil {
//                 log.Printf("Error parsing checker output: %v", err)
//                 return false, make(map[string]interface{})
//             }

//             // Check if payload contains a price
//             if _, ok := payload["price"]; !ok {
//                 log.Printf("No price found in payload")
//                 return false, make(map[string]interface{})
//             }

//             return true, payload
//         }, nil
//     default:
//         return nil, fmt.Errorf("unsupported argument type for checker: %s", job.ArgType)
//     }
// }

// func (h *JobHandler) defaultChecker(argType string) func(*manager.Job) (bool, map[string]interface{}) {
//     return func(j *manager.Job) (bool, map[string]interface{}) {
//         switch argType {
//         case "None", "Static":
//             if h.intervalChecker.ValidateJobInterval(j) {
//                 return true, make(map[string]interface{})
//             }
//             return false, make(map[string]interface{})
//         case "Dynamic":
//             return true, map[string]interface{}{
//                 "price": "0", // Default price for dynamic jobs
//             }
//         default:
//             return false, make(map[string]interface{})
//         }
//     }
// }

func (h *JobHandler) sendToQuorumHead(originalValue interface{}) (interface{}, error) {
	log.Printf("Sending value to Quorum Head: %v", originalValue)

	// Placeholder for Quorum Head processing
	// In future, this will involve complex quorum head logic
	receivedValue, err := h.receiveValueFromWorkers(originalValue)
	if err != nil {
		log.Printf("Error in Quorum Head processing: %v", err)
		return nil, err
	}

	return receivedValue, nil
}

func (h *JobHandler) receiveValueFromWorkers(originalValue interface{}) (interface{}, error) {
	log.Printf("Receiving value from workers: %v", originalValue)

	// Placeholder for worker value processing
	consensusValue, err := h.runConsensus(originalValue)
	if err != nil {
		log.Printf("Error in worker value processing: %v", err)
		return nil, err
	}

	return consensusValue, nil
}

func (h *JobHandler) runConsensus(originalValue interface{}) (interface{}, error) {
	log.Printf("Running consensus on value: %v", originalValue)

	// Placeholder for consensus logic
	// Currently just returns the original value
	// In future, this will involve complex consensus mechanism
	return originalValue, nil
}

// New method to fetch dynamic argument from API
func (h *JobHandler) fetchDynamicArgument(arguments map[string]interface{}) (interface{}, error) {
	// Check if URL is provided in arguments
	urlInterface, exists := arguments["url"]
	if !exists {
		log.Printf("No URL provided for dynamic argument, returning default value 0")
		return "0", nil
	}

	// Convert URL to string
	url, ok := urlInterface.(string)
	if !ok {
		log.Printf("Invalid URL format, returning default value 0")
		return "0", nil
	}

	log.Printf("Fetching dynamic argument from URL: %s", url)

	// Attempt to fetch value from API
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error fetching dynamic argument: %v, returning default value 0", err)
		return "0", nil
	}
	defer resp.Body.Close()

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v, returning default value 0", err)
		return "0", nil
	}

	log.Printf("Received API response body: %s", string(body))

	// Flexible response parsing
	var apiResponse map[string]interface{}
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		log.Printf("Error parsing API response: %v, returning default value 0", err)
		return "0", nil
	}

	// Try to extract value from different possible nested structures
	var value interface{}

	// Check for "data.value" structure
	if data, ok := apiResponse["data"].(map[string]interface{}); ok {
		value = data["value"]
	}

	// If not found, check for direct "value" at root
	if value == nil {
		value = apiResponse["value"]

		// If still nil, try checking for "price"
		if value == nil {
			value = apiResponse["price"]
		}
	}

	// Convert value to string
	var stringValue string
	switch v := value.(type) {
	case string:
		stringValue = v
	case float64:
		stringValue = fmt.Sprintf("%d", int(v))
	case int:
		stringValue = strconv.Itoa(v)
	case int64:
		stringValue = strconv.FormatInt(v, 10)
	default:
		log.Printf("Unexpected value type: %T, returning default value 0", value)
		return "0", nil
	}

	// If no value is found or conversion fails, return "0"
	if stringValue == "" {
		log.Printf("No value found in API response, returning default value 0")
		return "0", nil
	}

	// New workflow: Send fetched value through Quorum Head processing
	processedValue, err := h.sendToQuorumHead(stringValue)
	if err != nil {
		log.Printf("Error in Quorum Head processing: %v, using original value", err)
		processedValue = stringValue
	}

	// Update newPrice in arguments if it exists
	if _, exists := arguments["data"]; exists {
		arguments["data"] = processedValue
		log.Printf("Updated data to: %v", processedValue)
	}

	log.Printf("Fetched and processed dynamic argument value: %v", processedValue)
	return processedValue, nil
}

func (h *JobHandler) decodeContractOutput(contractABI *abi.ABI, method *abi.Method, output []byte) (interface{}, error) {
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

func (h *JobHandler) validateJob(job *manager.Job) error {
	if job == nil {
		return fmt.Errorf("received nil job")
	}
	if job.JobID == "" {
		return fmt.Errorf("invalid job: empty job ID")
	}
	return nil
}

func (h *JobHandler) getContractMethodAndABI(methodName, contractAddress string) (*abi.ABI, *abi.Method, error) {
	// Fetch ABI
	abiData, err := h.fetchContractABI(contractAddress)
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

func (h *JobHandler) fetchContractABI(contractAddress string) ([]byte, error) {
	if h.etherscanAPIKey == "" {
		return nil, fmt.Errorf("missing Etherscan API key")
	}

	url := fmt.Sprintf("https://api-sepolia-optimism.etherscan.io/api?module=contract&action=getabi&address=%s&apikey=%s", contractAddress, h.etherscanAPIKey)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
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
