package execution

import (
	"bytes"
	"encoding/json"
	"fmt"
	// "io/ioutil"
	// "net/http"
	// "reflect"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/types"
)

func (e *TaskExecutor) getContractMethodAndABI(methodName string, targetData *types.TaskTargetData) (*abi.ABI, *abi.Method, error) {
	if targetData.ABI == "" {
		return nil, nil, fmt.Errorf("contract ABI not provided in job data")
	}

	abiData := []byte(targetData.ABI)

	parsed, err := abi.JSON(bytes.NewReader(abiData))
	if err != nil {
		e.logger.Warnf("Error parsing ABI: %v", err)
		return nil, nil, err
	}

	e.logger.Debugf("Using ABI from database for contract %s", targetData.TargetContractAddress)

	method, ok := parsed.Methods[methodName]
	if !ok {
		e.logger.Warnf("Method %s not found in contract ABI", methodName)
		return nil, nil, fmt.Errorf("method %s not found in contract ABI", methodName)
	}

	e.logger.Debugf("Found method: %+v", method)
	return &parsed, &method, nil
}

func (e *TaskExecutor) processArguments(args interface{}, methodInputs []abi.Argument, contractABI *abi.ABI) ([]interface{}, error) {
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
	case string:
		// Handle a single string value (like from our script)
		// If there's only one input parameter, use the string value directly
		if len(methodInputs) == 1 {
			// First attempt to remove JSON string quotes if present
			strValue := argData
			if strings.HasPrefix(strValue, "\"") && strings.HasSuffix(strValue, "\"") {
				strValue = strings.Trim(strValue, "\"")
			}

			convertedArg, err := e.argConverter.convertToType(strValue, methodInputs[0].Type)
			if err != nil {
				return nil, fmt.Errorf("error converting string argument: %v", err)
			}
			convertedArgs = append(convertedArgs, convertedArg)
			return convertedArgs, nil
		} else {
			// Try to parse as JSON array for multiple parameters
			var arrayData []interface{}
			if err := json.Unmarshal([]byte(argData), &arrayData); err == nil {
				if len(arrayData) < len(methodInputs) {
					return nil, fmt.Errorf("not enough arguments in JSON array: expected %d, got %d",
						len(methodInputs), len(arrayData))
				}

				for i, inputParam := range methodInputs {
					convertedArg, err := e.argConverter.convertToType(arrayData[i], inputParam.Type)
					if err != nil {
						return nil, fmt.Errorf("error converting argument %d: %v", i, err)
					}
					convertedArgs = append(convertedArgs, convertedArg)
				}
				return convertedArgs, nil
			}

			return nil, fmt.Errorf("cannot convert single string to %d arguments", len(methodInputs))
		}
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

func (e *TaskExecutor) parseDynamicArgs(output string) []interface{} {
	var argData []interface{}

	if err := json.Unmarshal([]byte(output), &argData); err != nil {
		e.logger.Warnf("Error parsing dynamic arguments: %v", err)
		return nil
	}

	return argData
}

func (e *TaskExecutor) parseStaticArgs(args []string) []interface{} {
	var argData []interface{}

	for _, arg := range args {
		argData = append(argData, arg)
	}

	return argData
}

// func (e *TaskExecutor) decodeContractOutput(contractABI *abi.ABI, method *abi.Method, output []byte) (interface{}, error) {
// 	// Handle different output scenarios
// 	if len(method.Outputs) == 0 {
// 		e.logger.Infof("Method %s has no outputs to decode", method.Name)
// 		return nil, nil
// 	}

// 	// Single output case
// 	if len(method.Outputs) == 1 {
// 		outputType := method.Outputs[0]
// 		result := reflect.New(outputType.Type.GetType()).Elem()

// 		err := contractABI.UnpackIntoInterface(result.Addr().Interface(), method.Name, output)
// 		if err != nil {
// 			e.logger.Warnf("Error unpacking single output: %v", err)
// 			return nil, err
// 		}

// 		e.logger.Infof("Decoded single output: %v", result.Interface())
// 		return result.Interface(), nil
// 	}

// 	// Multiple outputs case
// 	results := make([]interface{}, len(method.Outputs))
// 	err := contractABI.UnpackIntoInterface(&results, method.Name, output)
// 	if err != nil {
// 		e.logger.Warnf("Error unpacking multiple outputs: %v", err)
// 		return nil, err
// 	}

// 	e.logger.Infof("Decoded multiple outputs: %+v", results)
// 	return results, nil
// }

// func (e *JobExecutor) fetchContractABI(contractAddress string) ([]byte, error) {
// 	if e.etherscanAPIKey == "" {
// 		return nil, fmt.Errorf("missing Etherscan API key")
// 	}

// 	// Update the URL to use Optimism Sepolia's API endpoint
// 	blockscoutUrl := fmt.Sprintf(
// 		"https://optimism-sepolia.blockscout.com/api?module=contract&action=getabi&address=%s",
// 		contractAddress)

// 	resp, err := http.Get(blockscoutUrl)
// 	if err != nil || resp.StatusCode != http.StatusOK {
// 		logger.Warnf("Failed to fetch ABI from Blockscout: %v", err)
// 		// Fall back to another source or handle accordingly
// 	}

// 	defer resp.Body.Close()

// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var response struct {
// 		Status  string `json:"status"`
// 		Message string `json:"message"`
// 		Result  string `json:"result"`
// 	}

// 	err = json.Unmarshal(body, &response)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if response.Status != "1" {
// 		return nil, fmt.Errorf("error fetching contract ABI: %s", response.Message)
// 	}

// 	return []byte(response.Result), nil
// }
