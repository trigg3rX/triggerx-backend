package execution

// import (
// 	"bytes"
// 	"encoding/json"
// 	"fmt"
// 	"reflect"
// 	"strings"

// 	"github.com/ethereum/go-ethereum/accounts/abi"
// 	jobtypes "github.com/trigg3rX/triggerx-backend/pkg/types"
// )

// func (e *JobExecutor) processArguments(args interface{}, methodInputs []abi.Argument, contractABI *abi.ABI) ([]interface{}, error) {
// 	convertedArgs := make([]interface{}, 0)

// 	if len(methodInputs) == 1 && methodInputs[0].Type.T == abi.TupleTy {
// 		switch v := args.(type) {
// 		case map[string]interface{}:
// 			convertedArg, err := e.argConverter.convertToStruct(v, methodInputs[0].Type)
// 			if err != nil {
// 				return nil, fmt.Errorf("error converting to struct: %v", err)
// 			}
// 			convertedArgs = append(convertedArgs, convertedArg)
// 			return convertedArgs, nil
// 		case string:
// 			var structData map[string]interface{}
// 			if err := json.Unmarshal([]byte(v), &structData); err == nil {
// 				convertedArg, err := e.argConverter.convertToStruct(structData, methodInputs[0].Type)
// 				if err != nil {
// 					return nil, fmt.Errorf("error converting JSON string to struct: %v", err)
// 				}
// 				convertedArgs = append(convertedArgs, convertedArg)
// 				return convertedArgs, nil
// 			}
// 		case []interface{}:
// 			if len(v) == 1 {
// 				if mapVal, ok := v[0].(map[string]interface{}); ok {
// 					convertedArg, err := e.argConverter.convertToStruct(mapVal, methodInputs[0].Type)
// 					if err != nil {
// 						return nil, fmt.Errorf("error converting map from array to struct: %v", err)
// 					}
// 					convertedArgs = append(convertedArgs, convertedArg)
// 					return convertedArgs, nil
// 				} else if strVal, ok := v[0].(string); ok {
// 					var structData map[string]interface{}
// 					if err := json.Unmarshal([]byte(strVal), &structData); err == nil {
// 						convertedArg, err := e.argConverter.convertToStruct(structData, methodInputs[0].Type)
// 						if err != nil {
// 							return nil, fmt.Errorf("error converting JSON string to struct: %v", err)
// 						}
// 						convertedArgs = append(convertedArgs, convertedArg)
// 						return convertedArgs, nil
// 					}
// 				}
// 			}
// 		}
// 	}

// 	switch argData := args.(type) {
// 	case string:
// 		if len(methodInputs) == 1 {
// 			strValue := argData
// 			if strings.HasPrefix(strValue, "\"") && strings.HasSuffix(strValue, "\"") {
// 				strValue = strings.Trim(strValue, "\"")
// 			}

// 			convertedArg, err := e.argConverter.convertToType(strValue, methodInputs[0].Type)
// 			if err != nil {
// 				return nil, fmt.Errorf("error converting string argument: %v", err)
// 			}
// 			convertedArgs = append(convertedArgs, convertedArg)
// 			return convertedArgs, nil
// 		} else {
// 			var arrayData []interface{}
// 			if err := json.Unmarshal([]byte(argData), &arrayData); err == nil {
// 				if len(arrayData) < len(methodInputs) {
// 					return nil, fmt.Errorf("not enough arguments in JSON array: expected %d, got %d",
// 						len(methodInputs), len(arrayData))
// 				}

// 				for i, inputParam := range methodInputs {
// 					convertedArg, err := e.argConverter.convertToType(arrayData[i], inputParam.Type)
// 					if err != nil {
// 						return nil, fmt.Errorf("error converting argument %d: %v", i, err)
// 					}
// 					convertedArgs = append(convertedArgs, convertedArg)
// 				}
// 				return convertedArgs, nil
// 			}

// 			return nil, fmt.Errorf("cannot convert single string to %d arguments", len(methodInputs))
// 		}
// 	case []string:
// 		if len(argData) < len(methodInputs) {
// 			return nil, fmt.Errorf("not enough arguments provided: expected %d, got %d",
// 				len(methodInputs), len(argData))
// 		}

// 		for i, inputParam := range methodInputs {
// 			convertedArg, err := e.argConverter.convertToType(argData[i], inputParam.Type)
// 			if err != nil {
// 				return nil, fmt.Errorf("error converting argument %d: %v", i, err)
// 			}
// 			convertedArgs = append(convertedArgs, convertedArg)
// 		}
// 	case []interface{}:
// 		if len(argData) < len(methodInputs) {
// 			return nil, fmt.Errorf("not enough arguments provided: expected %d, got %d",
// 				len(methodInputs), len(argData))
// 		}

// 		for i, inputParam := range methodInputs {
// 			convertedArg, err := e.argConverter.convertToType(argData[i], inputParam.Type)
// 			if err != nil {
// 				return nil, fmt.Errorf("error converting argument %d: %v", i, err)
// 			}
// 			convertedArgs = append(convertedArgs, convertedArg)
// 		}
// 	case map[string]interface{}:
// 		for _, inputParam := range methodInputs {
// 			paramName := inputParam.Name
// 			if paramName == "" {
// 				return nil, fmt.Errorf("cannot use map arguments with unnamed parameters")
// 			}

// 			argValue, exists := argData[paramName]
// 			if !exists {
// 				for k, v := range argData {
// 					if strings.EqualFold(k, paramName) {
// 						argValue = v
// 						exists = true
// 						break
// 					}
// 				}

// 				if !exists {
// 					return nil, fmt.Errorf("argument %s not found in input data", paramName)
// 				}
// 			}

// 			convertedArg, err := e.argConverter.convertToType(argValue, inputParam.Type)
// 			if err != nil {
// 				return nil, fmt.Errorf("error converting argument %s: %v", paramName, err)
// 			}
// 			convertedArgs = append(convertedArgs, convertedArg)
// 		}
// 	default:
// 		return nil, fmt.Errorf("unsupported argument format: %T", args)
// 	}

// 	return convertedArgs, nil
// }

// func (e *JobExecutor) getContractMethodAndABI(methodName string, job *jobtypes.HandleCreateJobData) (*abi.ABI, *abi.Method, error) {
// 	if job.ABI == "" {
// 		return nil, nil, fmt.Errorf("contract ABI not provided in job data")
// 	}

// 	abiData := []byte(job.ABI)

// 	parsed, err := abi.JSON(bytes.NewReader(abiData))
// 	if err != nil {
// 		logger.Warnf("Error parsing ABI: %v", err)
// 		return nil, nil, err
// 	}

// 	logger.Infof("Using ABI from database for contract %s", job.TargetContractAddress)

// 	method, ok := parsed.Methods[methodName]
// 	if !ok {
// 		logger.Warnf("Method %s not found in contract ABI", methodName)
// 		return nil, nil, fmt.Errorf("method %s not found in contract ABI", methodName)
// 	}

// 	logger.Infof("Found method: %+v", method)
// 	return &parsed, &method, nil
// }

// func (e *JobExecutor) decodeContractOutput(contractABI *abi.ABI, method *abi.Method, output []byte) (interface{}, error) {
// 	if len(method.Outputs) == 0 {
// 		logger.Infof("Method %s has no outputs to decode", method.Name)
// 		return nil, nil
// 	}

// 	if len(method.Outputs) == 1 {
// 		outputType := method.Outputs[0]
// 		result := reflect.New(outputType.Type.GetType()).Elem()

// 		err := contractABI.UnpackIntoInterface(result.Addr().Interface(), method.Name, output)
// 		if err != nil {
// 			logger.Warnf("Error unpacking single output: %v", err)
// 			return nil, err
// 		}

// 		logger.Infof("Decoded single output: %v", result.Interface())
// 		return result.Interface(), nil
// 	}

// 	results := make([]interface{}, len(method.Outputs))
// 	err := contractABI.UnpackIntoInterface(&results, method.Name, output)
// 	if err != nil {
// 		logger.Warnf("Error unpacking multiple outputs: %v", err)
// 		return nil, err
// 	}

// 	logger.Infof("Decoded multiple outputs: %+v", results)
// 	return results, nil
// }
