package execution

import (
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/stretchr/testify/assert"
)

func TestArgumentConverter_ConvertToType_Uint8(t *testing.T) {
	converter := &ArgumentConverter{}

	// Test uint8 type
	uint8Type := abi.Type{T: abi.UintTy, Size: 8}

	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
		hasError bool
	}{
		{
			name:     "string to uint8",
			input:    "0",
			expected: uint8(0),
			hasError: false,
		},
		{
			name:     "float64 to uint8",
			input:    float64(255),
			expected: uint8(255),
			hasError: false,
		},
		{
			name:     "int to uint8",
			input:    128,
			expected: uint8(128),
			hasError: false,
		},
		{
			name:     "uint8 to uint8",
			input:    uint8(42),
			expected: uint8(42),
			hasError: false,
		},
		{
			name:     "negative value should error",
			input:    -1,
			hasError: true,
		},
		{
			name:     "value too large should error",
			input:    256,
			hasError: true,
		},
		{
			name:     "float too large should error",
			input:    float64(256),
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.convertToType(tt.input, uint8Type)

			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestArgumentConverter_ConvertToType_Enum(t *testing.T) {
	converter := &ArgumentConverter{}

	// Test enum type (uint8)
	enumType := abi.Type{T: abi.UintTy, Size: 8}

	// Test the specific case from the user's error
	// operation: 0 (enum value)
	result, err := converter.convertToType(0, enumType)

	assert.NoError(t, err)
	assert.Equal(t, uint8(0), result)

	// Test with string representation
	result, err = converter.convertToType("0", enumType)

	assert.NoError(t, err)
	assert.Equal(t, uint8(0), result)
}

func TestArgumentConverter_ConvertToType_Array(t *testing.T) {
	converter := &ArgumentConverter{}

	// Test fixed-size array (uint8[3])
	fixedArrayType := abi.Type{
		T:    abi.ArrayTy,
		Size: 3,
		Elem: &abi.Type{T: abi.UintTy, Size: 8},
	}

	// Test dynamic slice (uint8[])
	sliceType := abi.Type{
		T:    abi.SliceTy,
		Elem: &abi.Type{T: abi.UintTy, Size: 8},
	}

	tests := []struct {
		name        string
		targetType  abi.Type
		input       interface{}
		expectedLen int
		hasError    bool
	}{
		{
			name:        "fixed array with correct length",
			targetType:  fixedArrayType,
			input:       []interface{}{1, 2, 3},
			expectedLen: 3,
			hasError:    false,
		},
		{
			name:        "fixed array with wrong length",
			targetType:  fixedArrayType,
			input:       []interface{}{1, 2},
			expectedLen: 0,
			hasError:    true,
		},
		{
			name:        "dynamic slice with any length",
			targetType:  sliceType,
			input:       []interface{}{1, 2, 3, 4},
			expectedLen: 4,
			hasError:    false,
		},
		{
			name:        "dynamic slice with empty array",
			targetType:  sliceType,
			input:       []interface{}{},
			expectedLen: 0,
			hasError:    false,
		},
		{
			name:        "string JSON array",
			targetType:  sliceType,
			input:       "[1,2,3]",
			expectedLen: 3,
			hasError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.convertToType(tt.input, tt.targetType)

			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Check the result type and length
				resultValue := reflect.ValueOf(result)
				if tt.targetType.T == abi.ArrayTy {
					assert.Equal(t, reflect.Array, resultValue.Kind())
				} else {
					assert.Equal(t, reflect.Slice, resultValue.Kind())
				}
				assert.Equal(t, tt.expectedLen, resultValue.Len())
			}
		})
	}
}

func TestArgumentConverter_ConvertToType_FixedBytes(t *testing.T) {
	converter := &ArgumentConverter{}

	// Test bytes32 type
	bytes32Type := abi.Type{T: abi.FixedBytesTy, Size: 32}

	// Test bytes4 type
	bytes4Type := abi.Type{T: abi.FixedBytesTy, Size: 4}

	tests := []struct {
		name       string
		targetType abi.Type
		input      interface{}
		hasError   bool
	}{
		{
			name:       "valid bytes32 hex string",
			targetType: bytes32Type,
			input:      "0x2dbb0cb2cb611d2ee68719380e6fe9578bb5b47ee9a562cd6529130a9436cc68",
			hasError:   false,
		},
		{
			name:       "valid bytes4 hex string",
			targetType: bytes4Type,
			input:      "0xd0e30db0",
			hasError:   false,
		},
		{
			name:       "bytes32 with wrong length",
			targetType: bytes32Type,
			input:      "0x1234",
			hasError:   true,
		},
		{
			name:       "bytes4 with wrong length",
			targetType: bytes4Type,
			input:      "0x1234567890",
			hasError:   true,
		},
		{
			name:       "invalid hex string",
			targetType: bytes32Type,
			input:      "not a hex string",
			hasError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.convertToType(tt.input, tt.targetType)

			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Check that the result is a fixed-size array
				resultValue := reflect.ValueOf(result)
				assert.Equal(t, reflect.Array, resultValue.Kind())
				assert.Equal(t, tt.targetType.Size, resultValue.Len())

				// Check that the element type is byte
				assert.Equal(t, reflect.TypeOf(byte(0)), resultValue.Type().Elem())
			}
		})
	}
}
