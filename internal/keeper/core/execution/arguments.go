package execution

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
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
	case abi.BytesTy:
		return ac.convertToBytes(value, targetType)
	case abi.FixedBytesTy:
		return ac.convertToFixedBytes(value, targetType)
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
		// Handle different uint sizes specifically
		switch targetType.Size {
		case 8: // uint8
			switch v := value.(type) {
			case string:
				// Parse as float first, then convert to uint8
				floatVal, err := strconv.ParseFloat(v, 64)
				if err != nil {
					return nil, err
				}
				if floatVal < 0 {
					return nil, fmt.Errorf("cannot convert negative value %f to uint8", floatVal)
				}
				if floatVal > 255 {
					return nil, fmt.Errorf("value %f exceeds maximum uint8 value", floatVal)
				}
				return uint8(floatVal), nil
			case float64:
				if v < 0 {
					return nil, fmt.Errorf("cannot convert negative value %f to uint8", v)
				}
				if v > 255 {
					return nil, fmt.Errorf("value %f exceeds maximum uint8 value", v)
				}
				return uint8(v), nil
			case int:
				if v < 0 {
					return nil, fmt.Errorf("cannot convert negative value %d to uint8", v)
				}
				if v > 255 {
					return nil, fmt.Errorf("value %d exceeds maximum uint8 value", v)
				}
				return uint8(v), nil
			case uint8:
				return v, nil
			case map[string]interface{}:
				// This could be a struct that we need to convert to an integer
				if jsonBytes, err := json.Marshal(v); err == nil {
					var floatVal float64
					if err := json.Unmarshal(jsonBytes, &floatVal); err == nil {
						if floatVal < 0 {
							return nil, fmt.Errorf("cannot convert negative value %f to uint8", floatVal)
						}
						if floatVal > 255 {
							return nil, fmt.Errorf("value %f exceeds maximum uint8 value", floatVal)
						}
						return uint8(floatVal), nil
					}
				}
				return nil, fmt.Errorf("cannot convert map to uint8")
			default:
				return nil, fmt.Errorf("cannot convert type %T to uint8", v)
			}
		case 16: // uint16
			switch v := value.(type) {
			case string:
				floatVal, err := strconv.ParseFloat(v, 64)
				if err != nil {
					return nil, err
				}
				if floatVal < 0 {
					return nil, fmt.Errorf("cannot convert negative value %f to uint16", floatVal)
				}
				if floatVal > 65535 {
					return nil, fmt.Errorf("value %f exceeds maximum uint16 value", floatVal)
				}
				return uint16(floatVal), nil
			case float64:
				if v < 0 {
					return nil, fmt.Errorf("cannot convert negative value %f to uint16", v)
				}
				if v > 65535 {
					return nil, fmt.Errorf("value %f exceeds maximum uint16 value", v)
				}
				return uint16(v), nil
			case int:
				if v < 0 {
					return nil, fmt.Errorf("cannot convert negative value %d to uint16", v)
				}
				if v > 65535 {
					return nil, fmt.Errorf("value %d exceeds maximum uint16 value", v)
				}
				return uint16(v), nil
			case uint16:
				return v, nil
			default:
				return nil, fmt.Errorf("cannot convert type %T to uint16", v)
			}
		case 32: // uint32
			switch v := value.(type) {
			case string:
				// Parse as float first, then convert to uint32
				floatVal, err := strconv.ParseFloat(v, 64)
				if err != nil {
					return nil, err
				}
				if floatVal < 0 {
					return nil, fmt.Errorf("cannot convert negative value %f to uint32", floatVal)
				}
				if floatVal > math.MaxUint32 {
					return nil, fmt.Errorf("value %f exceeds maximum uint32 value", floatVal)
				}
				return uint32(floatVal), nil
			case float64:
				return uint32(v), nil
			case int:
				if v < 0 {
					return 0, fmt.Errorf("cannot convert negative value %d to uint32", v)
				}
				if v > math.MaxUint32 {
					return 0, fmt.Errorf("value %d exceeds maximum uint32 value", v)
				}
				return uint32(v), nil
			case uint32:
				return v, nil
			default:
				return nil, fmt.Errorf("cannot convert type %T to uint32", v)
			}
		case 64: // uint64
			switch v := value.(type) {
			case string:
				floatVal, err := strconv.ParseFloat(v, 64)
				if err != nil {
					return nil, err
				}
				if floatVal < 0 {
					return nil, fmt.Errorf("cannot convert negative value %f to uint64", floatVal)
				}
				if floatVal > math.MaxUint64 {
					return nil, fmt.Errorf("value %f exceeds maximum uint64 value", floatVal)
				}
				return uint64(floatVal), nil
			case float64:
				if v < 0 {
					return nil, fmt.Errorf("cannot convert negative value %f to uint64", v)
				}
				if v > math.MaxUint64 {
					return nil, fmt.Errorf("value %f exceeds maximum uint64 value", v)
				}
				return uint64(v), nil
			case int:
				if v < 0 {
					return nil, fmt.Errorf("cannot convert negative value %d to uint64", v)
				}
				return uint64(v), nil
			case uint64:
				return v, nil
			default:
				return nil, fmt.Errorf("cannot convert type %T to uint64", v)
			}
		default:
			// For other uint sizes, use big.Int
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
	case abi.IntTy:
		// Handle signed integers
		switch targetType.Size {
		case 8: // int8
			switch v := value.(type) {
			case string:
				floatVal, err := strconv.ParseFloat(v, 64)
				if err != nil {
					return nil, err
				}
				if floatVal < -128 || floatVal > 127 {
					return nil, fmt.Errorf("value %f exceeds int8 range", floatVal)
				}
				return int8(floatVal), nil
			case float64:
				if v < -128 || v > 127 {
					return nil, fmt.Errorf("value %f exceeds int8 range", v)
				}
				return int8(v), nil
			case int:
				if v < -128 || v > 127 {
					return nil, fmt.Errorf("value %d exceeds int8 range", v)
				}
				return int8(v), nil
			case int8:
				return v, nil
			default:
				return nil, fmt.Errorf("cannot convert type %T to int8", v)
			}
		case 16: // int16
			switch v := value.(type) {
			case string:
				floatVal, err := strconv.ParseFloat(v, 64)
				if err != nil {
					return nil, err
				}
				if floatVal < -32768 || floatVal > 32767 {
					return nil, fmt.Errorf("value %f exceeds int16 range", floatVal)
				}
				return int16(floatVal), nil
			case float64:
				if v < -32768 || v > 32767 {
					return nil, fmt.Errorf("value %f exceeds int16 range", v)
				}
				return int16(v), nil
			case int:
				if v < -32768 || v > 32767 {
					return nil, fmt.Errorf("value %d exceeds int16 range", v)
				}
				return int16(v), nil
			case int16:
				return v, nil
			default:
				return nil, fmt.Errorf("cannot convert type %T to int16", v)
			}
		case 32: // int32
			switch v := value.(type) {
			case string:
				floatVal, err := strconv.ParseFloat(v, 64)
				if err != nil {
					return nil, err
				}
				if floatVal < math.MinInt32 || floatVal > math.MaxInt32 {
					return nil, fmt.Errorf("value %f exceeds int32 range", floatVal)
				}
				return int32(floatVal), nil
			case float64:
				if v < math.MinInt32 || v > math.MaxInt32 {
					return nil, fmt.Errorf("value %f exceeds int32 range", v)
				}
				return int32(v), nil
			case int:
				if v < math.MinInt32 || v > math.MaxInt32 {
					return nil, fmt.Errorf("value %d exceeds int32 range", v)
				}
				return int32(v), nil
			case int32:
				return v, nil
			default:
				return nil, fmt.Errorf("cannot convert type %T to int32", v)
			}
		case 64: // int64
			switch v := value.(type) {
			case string:
				floatVal, err := strconv.ParseFloat(v, 64)
				if err != nil {
					return nil, err
				}
				if floatVal < math.MinInt64 || floatVal > math.MaxInt64 {
					return nil, fmt.Errorf("value %f exceeds int64 range", floatVal)
				}
				return int64(floatVal), nil
			case float64:
				if v < math.MinInt64 || v > math.MaxInt64 {
					return nil, fmt.Errorf("value %f exceeds int64 range", v)
				}
				return int64(v), nil
			case int:
				return int64(v), nil
			case int64:
				return v, nil
			default:
				return nil, fmt.Errorf("cannot convert type %T to int64", v)
			}
		default:
			// For other int sizes, use big.Int
			switch v := value.(type) {
			case float64:
				return big.NewInt(int64(v)), nil
			case string:
				floatVal, err := strconv.ParseFloat(v, 64)
				if err != nil {
					return nil, fmt.Errorf("cannot convert string to integer: %v", err)
				}
				return big.NewInt(int64(floatVal)), nil
			case map[string]interface{}:
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
	default:
		return nil, fmt.Errorf("unsupported integer type: %v", targetType)
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

func (ac *ArgumentConverter) convertToBytes(value interface{}, targetType abi.Type) ([]byte, error) {
	// Convert to dynamic bytes
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

func (ac *ArgumentConverter) convertToFixedBytes(value interface{}, targetType abi.Type) (interface{}, error) {
	// Convert to fixed-size byte array
	expectedSize := targetType.Size

	switch v := value.(type) {
	case string:
		// Check if it's a hex string
		if strings.HasPrefix(v, "0x") {
			bytes := ethcommon.FromHex(v)
			if len(bytes) != expectedSize {
				return nil, fmt.Errorf("hex string length mismatch: expected %d bytes, got %d", expectedSize, len(bytes))
			}

			// Create a fixed-size array
			arrayType := reflect.ArrayOf(expectedSize, reflect.TypeOf(byte(0)))
			result := reflect.New(arrayType).Elem()

			for i := 0; i < expectedSize; i++ {
				result.Index(i).SetUint(uint64(bytes[i]))
			}

			return result.Interface(), nil
		}
		return nil, fmt.Errorf("invalid hex string for fixed bytes")
	case []byte:
		if len(v) != expectedSize {
			return nil, fmt.Errorf("byte slice length mismatch: expected %d bytes, got %d", expectedSize, len(v))
		}

		// Create a fixed-size array
		arrayType := reflect.ArrayOf(expectedSize, reflect.TypeOf(byte(0)))
		result := reflect.New(arrayType).Elem()

		for i := 0; i < expectedSize; i++ {
			result.Index(i).SetUint(uint64(v[i]))
		}

		return result.Interface(), nil
	default:
		return nil, fmt.Errorf("cannot convert type %T to fixed bytes", v)
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

	// Handle fixed arrays vs slices differently
	if targetType.T == abi.ArrayTy {
		// Fixed-size array
		expectedLength := targetType.Size
		if len(sourceArray) != expectedLength {
			return nil, fmt.Errorf("array length mismatch: expected %d, got %d", expectedLength, len(sourceArray))
		}

		// Create a fixed-size array
		arrayType := reflect.ArrayOf(expectedLength, targetType.Elem.GetType())
		result := reflect.New(arrayType).Elem()

		// Convert each element
		for i, elem := range sourceArray {
			convertedElem, err := ac.convertToType(elem, *targetType.Elem)
			if err != nil {
				return nil, fmt.Errorf("error converting array element %d: %v", i, err)
			}

			// Set the element in the array
			resultElem := reflect.ValueOf(convertedElem)
			result.Index(i).Set(resultElem)
		}

		return result.Interface(), nil
	} else {
		// Dynamic slice
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
