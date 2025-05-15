package execution

// import (
// 	"encoding/json"
// 	"fmt"
// 	"math/big"
// 	"reflect"
// 	"strconv"
// 	"strings"

// 	"github.com/ethereum/go-ethereum/accounts/abi"
// 	ethcommon "github.com/ethereum/go-ethereum/common"
// )

// type ArgumentConverter struct{}

// func (ac *ArgumentConverter) convertToType(value interface{}, targetType abi.Type) (interface{}, error) {
// 	switch targetType.T {
// 	case abi.UintTy, abi.IntTy:
// 		return ac.convertToInteger(value, targetType)
// 	case abi.StringTy:
// 		return ac.convertToString(value)
// 	case abi.BoolTy:
// 		return ac.convertToBool(value)
// 	case abi.AddressTy:
// 		return ac.convertToAddress(value)
// 	case abi.BytesTy, abi.FixedBytesTy:
// 		return ac.convertToBytes(value)
// 	case abi.ArrayTy, abi.SliceTy:
// 		return ac.convertToArray(value, targetType)
// 	case abi.TupleTy:
// 		return ac.convertToStruct(value, targetType)
// 	default:
// 		return nil, fmt.Errorf("unsupported type conversion: %v", targetType)
// 	}
// }
// func (ac *ArgumentConverter) convertToInteger(value interface{}, targetType abi.Type) (interface{}, error) {
// 	if bigInt, ok := value.(*big.Int); ok {
// 		return bigInt, nil
// 	}

// 	switch targetType.T {
// 	case abi.UintTy:
// 		if targetType.Size == 32 {
// 			switch v := value.(type) {
// 			case string:
// 				floatVal, err := strconv.ParseFloat(v, 64)
// 				if err != nil {
// 					return nil, err
// 				}
// 				return uint32(floatVal), nil
// 			case float64:
// 				return uint32(v), nil
// 			case int:
// 				return uint32(v), nil
// 			}
// 		}
// 		fallthrough
// 	default:
// 		switch v := value.(type) {
// 		case float64:
// 			return big.NewInt(int64(v)), nil
// 		case string:
// 			floatVal, err := strconv.ParseFloat(v, 64)
// 			if err != nil {
// 				return nil, fmt.Errorf("cannot convert string to integer: %v", err)
// 			}
// 			return big.NewInt(int64(floatVal)), nil
// 		case map[string]interface{}:
// 			if jsonBytes, err := json.Marshal(v); err == nil {
// 				var floatVal float64
// 				if err := json.Unmarshal(jsonBytes, &floatVal); err == nil {
// 					return big.NewInt(int64(floatVal)), nil
// 				}
// 			}
// 			return nil, fmt.Errorf("cannot convert map to integer")
// 		default:
// 			return nil, fmt.Errorf("cannot convert type %T to integer", v)
// 		}
// 	}
// }

// func (ac *ArgumentConverter) convertToString(value interface{}) (string, error) {
// 	switch v := value.(type) {
// 	case string:
// 		return v, nil
// 	case float64, int, uint:
// 		return fmt.Sprintf("%v", v), nil
// 	case map[string]interface{}:
// 		if jsonBytes, err := json.Marshal(v); err == nil {
// 			return string(jsonBytes), nil
// 		}
// 		return "", fmt.Errorf("cannot convert map to string")
// 	default:
// 		return "", fmt.Errorf("cannot convert type %T to string", v)
// 	}
// }

// func (ac *ArgumentConverter) convertToBool(value interface{}) (bool, error) {
// 	switch v := value.(type) {
// 	case bool:
// 		return v, nil
// 	case string:
// 		return strconv.ParseBool(v)
// 	case float64:
// 		return v != 0, nil
// 	case map[string]interface{}:
// 		if jsonBytes, err := json.Marshal(v); err == nil {
// 			var boolVal bool
// 			if err := json.Unmarshal(jsonBytes, &boolVal); err == nil {
// 				return boolVal, nil
// 			}
// 		}
// 		return false, fmt.Errorf("cannot convert map to bool")
// 	default:
// 		return false, fmt.Errorf("cannot convert type %T to bool", v)
// 	}
// }

// func (ac *ArgumentConverter) convertToAddress(value interface{}) (ethcommon.Address, error) {
// 	switch v := value.(type) {
// 	case string:
// 		if !ethcommon.IsHexAddress(v) {
// 			return ethcommon.Address{}, fmt.Errorf("invalid Ethereum address: %s", v)
// 		}
// 		return ethcommon.HexToAddress(v), nil
// 	case map[string]interface{}:
// 		if addrStr, ok := v["address"].(string); ok {
// 			if !ethcommon.IsHexAddress(addrStr) {
// 				return ethcommon.Address{}, fmt.Errorf("invalid Ethereum address: %s", addrStr)
// 			}
// 			return ethcommon.HexToAddress(addrStr), nil
// 		}
// 		return ethcommon.Address{}, fmt.Errorf("cannot convert map to address")
// 	default:
// 		return ethcommon.Address{}, fmt.Errorf("cannot convert type %T to address", v)
// 	}
// }

// func (ac *ArgumentConverter) convertToBytes(value interface{}) ([]byte, error) {
// 	switch v := value.(type) {
// 	case string:
// 		if strings.HasPrefix(v, "0x") {
// 			return ethcommon.FromHex(v), nil
// 		}
// 		return []byte(v), nil
// 	case []byte:
// 		return v, nil
// 	case map[string]interface{}:
// 		if jsonBytes, err := json.Marshal(v); err == nil {
// 			return jsonBytes, nil
// 		}
// 		return nil, fmt.Errorf("cannot convert map to bytes")
// 	default:
// 		return nil, fmt.Errorf("cannot convert type %T to bytes", v)
// 	}
// }

// func (ac *ArgumentConverter) convertToArray(value interface{}, targetType abi.Type) (interface{}, error) {
// 	var sourceArray []interface{}

// 	switch v := value.(type) {
// 	case []interface{}:
// 		sourceArray = v
// 	case string:
// 		if err := json.Unmarshal([]byte(v), &sourceArray); err != nil {
// 			return nil, fmt.Errorf("failed to parse string as JSON array: %v", err)
// 		}
// 	case map[string]interface{}:
// 		jsonBytes, err := json.Marshal(v)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to marshal map as JSON: %v", err)
// 		}
// 		if err := json.Unmarshal(jsonBytes, &sourceArray); err != nil {
// 			return nil, fmt.Errorf("failed to parse map as JSON array: %v", err)
// 		}
// 	default:
// 		return nil, fmt.Errorf("cannot convert type %T to array/slice", v)
// 	}

// 	sliceType := reflect.SliceOf(targetType.Elem.GetType())
// 	result := reflect.MakeSlice(sliceType, len(sourceArray), len(sourceArray))

// 	for i, elem := range sourceArray {
// 		convertedElem, err := ac.convertToType(elem, *targetType.Elem)
// 		if err != nil {
// 			return nil, fmt.Errorf("error converting array element %d: %v", i, err)
// 		}

// 		resultElem := reflect.ValueOf(convertedElem)
// 		result.Index(i).Set(resultElem)
// 	}

// 	return result.Interface(), nil
// }

// func (ac *ArgumentConverter) convertToStruct(value interface{}, targetType abi.Type) (interface{}, error) {
// 	structType := targetType.GetType()
// 	structValue := reflect.New(structType).Elem()

// 	var sourceMap map[string]interface{}

// 	switch v := value.(type) {
// 	case map[string]interface{}:
// 		sourceMap = v
// 	case string:
// 		if err := json.Unmarshal([]byte(v), &sourceMap); err != nil {
// 			return nil, fmt.Errorf("failed to parse string as JSON object: %v", err)
// 		}
// 	default:
// 		valueVal := reflect.ValueOf(value)
// 		if valueVal.Kind() == reflect.Struct {
// 			jsonBytes, err := json.Marshal(value)
// 			if err != nil {
// 				return nil, fmt.Errorf("failed to marshal struct: %v", err)
// 			}
// 			if err := json.Unmarshal(jsonBytes, &sourceMap); err != nil {
// 				return nil, fmt.Errorf("failed to unmarshal struct to map: %v", err)
// 			}
// 		} else {
// 			return nil, fmt.Errorf("cannot convert type %T to struct", v)
// 		}
// 	}

// 	for i, component := range targetType.TupleElems {
// 		fieldName := targetType.TupleRawNames[i]
// 		fieldValue, exists := sourceMap[fieldName]

// 		if !exists {
// 			for k, v := range sourceMap {
// 				if strings.EqualFold(k, fieldName) {
// 					fieldValue = v
// 					exists = true
// 					break
// 				}
// 			}
// 		}

// 		if !exists {
// 			logger.Warnf("Field %s not found in input data", fieldName)
// 			continue
// 		}

// 		convertedValue, err := ac.convertToType(fieldValue, *component)
// 		if err != nil {
// 			return nil, fmt.Errorf("error converting struct field %s: %v", fieldName, err)
// 		}

// 		var structField reflect.Value
// 		for j := 0; j < structValue.NumField(); j++ {
// 			if strings.EqualFold(structType.Field(j).Name, fieldName) {
// 				structField = structValue.Field(j)
// 				break
// 			}
// 		}

// 		if !structField.IsValid() {
// 			return nil, fmt.Errorf("struct field %s not found", fieldName)
// 		}

// 		convertedValueReflect := reflect.ValueOf(convertedValue)
// 		if structField.Type() != convertedValueReflect.Type() {
// 			if convertedValueReflect.Type().ConvertibleTo(structField.Type()) {
// 				convertedValueReflect = convertedValueReflect.Convert(structField.Type())
// 			} else {
// 				return nil, fmt.Errorf("cannot convert %v to field type %v", convertedValueReflect.Type(), structField.Type())
// 			}
// 		}
// 		structField.Set(convertedValueReflect)
// 	}

// 	return structValue.Interface(), nil
// }
