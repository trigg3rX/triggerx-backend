package execution

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"

	// "io/ioutil"
	// "net/http"
	// "reflect"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// isAgentJob checks if the given task definition ID is an agent job (TDI 7, 8, or 9)
func isAgentJob(taskDefinitionID int) bool {
	return taskDefinitionID == 7 ||
		taskDefinitionID == 8 ||
		taskDefinitionID == 9
}

func (e *TaskExecutor) getContractMethodAndABI(ctx context.Context, methodName string, targetData *types.TaskTargetData) (*abi.ABI, *abi.Method, error) {
	if targetData.ABI == "" {
		return nil, nil, fmt.Errorf("contract ABI not provided in job data")
	}

	abiData := []byte(targetData.ABI)

	parsed, err := abi.JSON(bytes.NewReader(abiData))
	if err != nil {
		e.logger.Warn(ctx, "Error parsing ABI", observability.Error(err))
		return nil, nil, err
	}

	// e.logger.Debug(ctx, "Using ABI from database", observability.String("contract_address", targetData.TargetContractAddress))

	method, ok := parsed.Methods[methodName]
	if !ok {
		e.logger.Warn(ctx, "Method not found in contract ABI", observability.String("method_name", methodName))
		return nil, nil, fmt.Errorf("method %s not found in contract ABI", methodName)
	}

	// e.logger.Debug(ctx, "Found method", observability.String("method_name", method.Name))
	return &parsed, &method, nil
}

func (e *TaskExecutor) processArguments(ctx context.Context, args interface{}, methodInputs []abi.Argument) ([]interface{}, error) {
	convertedArgs := make([]interface{}, 0)

	// e.logger.Debug(ctx, "Processing arguments for method inputs", observability.Any("args", args), observability.Any("method_inputs", methodInputs))

	// Handle nil or empty args
	if args == nil {
		// e.logger.Warn(ctx, "Received nil arguments", observability.Any("args", args))
		return nil, fmt.Errorf("nil arguments provided")
	}

	// Check if we have any inputs at all
	if len(methodInputs) == 0 {
		// e.logger.Debug(ctx, "Method has no inputs, returning empty args")
		return convertedArgs, nil
	}

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

	e.logger.Debug(ctx, "Successfully converted arguments", observability.Any("converted_args", convertedArgs))
	return convertedArgs, nil
}

func (e *TaskExecutor) parseDynamicArgs(ctx context.Context, output string) []interface{} {
	// First try to parse as JSON
	var argData []interface{}
	if err := json.Unmarshal([]byte(output), &argData); err == nil && len(argData) > 0 {
		e.logger.Debug(ctx, "Successfully parsed dynamic arguments as JSON", observability.Any("arg_data", argData))
		return argData
	}

	// If JSON parsing fails, try to extract values from container logs
	// Look for lines containing "Response:" which typically contain the values we need
	responsePattern := regexp.MustCompile(`Response:\s*([\d\.]+)`)
	matches := responsePattern.FindAllStringSubmatch(output, -1)

	if len(matches) > 0 {
		// Extract all response values
		argData = make([]interface{}, 0, len(matches))
		for _, match := range matches {
			if len(match) >= 2 {
				// Try to parse as float first
				if val, err := strconv.ParseFloat(match[1], 64); err == nil {
					// e.logger.Debug(ctx, "Found numeric response value", observability.Float64("val", val))
					argData = append(argData, val)
					continue
				}

				// If not a number, use as string
				// e.logger.Debug(ctx, "Found string response value", observability.String("value", match[1]))
				argData = append(argData, match[1])
			}
		}

		if len(argData) > 0 {
			return argData
		}
	}

	// If we can't find "Response:" lines, look for any numeric values
	numericPattern := regexp.MustCompile(`[\d\.]+`)
	numMatches := numericPattern.FindAllString(output, -1)

	if len(numMatches) > 0 {
		// Filter out timestamps and other irrelevant numbers
		for _, match := range numMatches {
			if val, err := strconv.ParseFloat(match, 64); err == nil {
				// Only consider "significant" numbers (not small ones that might be timestamps)
				if val > 100 {
					// e.logger.Debug(ctx, "Found significant numeric value", observability.Float64("val", val))
					argData = append(argData, val)
				}
			}
		}

		if len(argData) > 0 {
			return argData
		}
	}

	// As a fallback, check for "Condition satisfied: true" pattern
	if strings.Contains(output, "Condition satisfied: true") {
		// e.logger.Debug(ctx, "Found condition satisfied pattern, using true as argument")
		return []interface{}{true}
	}

	// e.logger.Warn(ctx, "Failed to extract any arguments from output", observability.String("output", output))
	return []interface{}{"0"} // Return a default value as fallback
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
// 		e.logger.Info(ctx, "Method has no outputs to decode", observability.String("method_name", method.Name))
// 		return nil, nil
// 	}

// 	// Single output case
// 	if len(method.Outputs) == 1 {
// 		outputType := method.Outputs[0]
// 		result := reflect.New(outputType.Type.GetType()).Elem()

// 		err := contractABI.UnpackIntoInterface(result.Addr().Interface(), method.Name, output)
// 		if err != nil {
// 			e.logger.Warn(ctx, "Error unpacking single output", observability.Error(err))
// 			return nil, err
// 		}

// 		e.logger.Info(ctx, "Decoded single output", observability.Any("result", result.Interface()))
// 		return result.Interface(), nil
// 	}

// 	// Multiple outputs case
// 	results := make([]interface{}, len(method.Outputs))
// 	err := contractABI.UnpackIntoInterface(&results, method.Name, output)
// 	if err != nil {
// 		e.logger.Warn(ctx, "Error unpacking multiple outputs", observability.Error(err))
// 		return nil, err
// 	}

// 	e.logger.Info(ctx, "Decoded multiple outputs", observability.Any("results", results))
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
// 		logger.Warn(ctx, "Failed to fetch ABI from Blockscout", observability.Error(err))
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
