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
    "reflect"
    "strconv"

    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"      
    "github.com/ethereum/go-ethereum"

    "github.com/trigg3rX/triggerx-keeper/execute/executor"
    "github.com/trigg3rX/go-backend/execute/manager"
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
    executor *executor.JobExecutor
    ethClient *ethclient.Client
    etherscanAPIKey string
    argConverter *ArgumentConverter 
}

func NewJobHandler(ethClient *ethclient.Client, etherscanAPIKey string) *JobHandler {
    return &JobHandler{
        executor:        executor.NewJobExecutor(),
        ethClient:       ethClient,
        etherscanAPIKey: etherscanAPIKey,
        argConverter:    &ArgumentConverter{},
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

    // Execute job based on argument type
    switch job.ArgType {
    case "None":
        return h.executeNoArgContract(job)
    case "Static":
        return h.executeStaticArgContract(job)
    case "Dynamic":
        log.Printf("In Progress: Job %s with dynamic arguments", job.JobID)
        return nil
    default:
        return fmt.Errorf("unsupported argument type: %s", job.ArgType)
    }
}

func (h *JobHandler) executeNoArgContract(job *manager.Job) error {
    log.Printf("Executing contract call for job %s with no arguments", job.JobID)

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