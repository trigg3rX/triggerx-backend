package execution

import (
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

// ===========================================
// UNIT TESTS FOR ArgumentConverter.convertToType
// ===========================================

func TestConvertToType_StringToUint256(t *testing.T) {
	// Arrange
	converter := &ArgumentConverter{}
	uintType, _ := abi.NewType("uint256", "", nil)

	// Act
	result, err := converter.convertToType("123", uintType)

	// Assert
	assert.NoError(t, err)
	bigIntResult, ok := result.(*big.Int)
	assert.True(t, ok)
	assert.Equal(t, "123", bigIntResult.String())
}

func TestConvertToType_StringToAddress(t *testing.T) {
	// Arrange
	converter := &ArgumentConverter{}
	addressType, _ := abi.NewType("address", "", nil)
	addressStr := "0x742d35cc6632c0532c718d329da9e8e0a7d4b7fa"

	// Act
	result, err := converter.convertToType(addressStr, addressType)

	// Assert
	assert.NoError(t, err)
	addressResult, ok := result.(common.Address)
	assert.True(t, ok)
	// Use case-insensitive comparison since Ethereum addresses are checksummed
	assert.True(t, strings.EqualFold(addressStr, addressResult.Hex()))
}

func TestConvertToType_StringToBool(t *testing.T) {
	// Arrange
	converter := &ArgumentConverter{}
	boolType, _ := abi.NewType("bool", "", nil)

	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
	}

	for _, test := range tests {
		// Act
		result, err := converter.convertToType(test.input, boolType)

		// Assert
		assert.NoError(t, err)
		boolResult, ok := result.(bool)
		assert.True(t, ok)
		assert.Equal(t, test.expected, boolResult)
	}
}

func TestConvertToType_InvalidAddress(t *testing.T) {
	// Arrange
	converter := &ArgumentConverter{}
	addressType, _ := abi.NewType("address", "", nil)

	// Act
	result, err := converter.convertToType("invalid-address", addressType)

	// Assert
	assert.Error(t, err)
	// The result should be a zero address, not nil
	addressResult, ok := result.(common.Address)
	assert.True(t, ok)
	assert.Equal(t, common.Address{}, addressResult)
	assert.Contains(t, err.Error(), "invalid")
}

func TestConvertToType_UnsupportedType(t *testing.T) {
	// Arrange
	converter := &ArgumentConverter{}

	// Create a type that's not handled
	functionType, _ := abi.NewType("function", "", nil)

	// Act
	result, err := converter.convertToType("test", functionType)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unsupported type")
}

// ===========================================
// UNIT TESTS FOR ArgumentConverter.convertToInteger
// ===========================================

func TestConvertToInteger_ValidString(t *testing.T) {
	// Arrange
	converter := &ArgumentConverter{}
	uintType, _ := abi.NewType("uint256", "", nil)

	tests := []struct {
		input    string
		expected string
	}{
		{"123", "123"},
		{"0", "0"},
		{"999", "999"},
	}

	for _, test := range tests {
		// Act
		result, err := converter.convertToInteger(test.input, uintType)

		// Assert
		assert.NoError(t, err)
		bigIntResult, ok := result.(*big.Int)
		assert.True(t, ok)
		assert.Equal(t, test.expected, bigIntResult.String())
	}
}

func TestConvertToInteger_InvalidString(t *testing.T) {
	// Arrange
	converter := &ArgumentConverter{}
	uintType, _ := abi.NewType("uint256", "", nil)

	// Act
	result, err := converter.convertToInteger("not-a-number", uintType)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestConvertToInteger_FromBigInt(t *testing.T) {
	// Arrange
	converter := &ArgumentConverter{}
	uintType, _ := abi.NewType("uint256", "", nil)
	bigInt := big.NewInt(456)

	// Act
	result, err := converter.convertToInteger(bigInt, uintType)

	// Assert
	assert.NoError(t, err)
	bigIntResult, ok := result.(*big.Int)
	assert.True(t, ok)
	assert.Equal(t, "456", bigIntResult.String())
}

func TestConvertToInteger_FromFloat(t *testing.T) {
	// Arrange
	converter := &ArgumentConverter{}
	uintType, _ := abi.NewType("uint256", "", nil)

	// Act
	result, err := converter.convertToInteger(123.456, uintType)

	// Assert
	assert.NoError(t, err)
	bigIntResult, ok := result.(*big.Int)
	assert.True(t, ok)
	assert.Equal(t, "123", bigIntResult.String()) // Should truncate
}

func TestConvertToInteger_Uint32Type(t *testing.T) {
	// Arrange
	converter := &ArgumentConverter{}
	uint32Type, _ := abi.NewType("uint32", "", nil)

	// Act
	result, err := converter.convertToInteger("123", uint32Type)

	// Assert
	assert.NoError(t, err)
	uint32Result, ok := result.(uint32)
	assert.True(t, ok)
	assert.Equal(t, uint32(123), uint32Result)
}

func TestConvertToInteger_UnsupportedType(t *testing.T) {
	// Arrange
	converter := &ArgumentConverter{}
	uintType, _ := abi.NewType("uint256", "", nil)

	// Act
	result, err := converter.convertToInteger([]int{1, 2, 3}, uintType)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "cannot convert")
}

// ===========================================
// UNIT TESTS FOR ArgumentConverter.convertToString
// ===========================================

func TestConvertToString_FromString(t *testing.T) {
	// Arrange
	converter := &ArgumentConverter{}

	// Act
	result, err := converter.convertToString("hello world")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "hello world", result)
}

func TestConvertToString_FromInt(t *testing.T) {
	// Arrange
	converter := &ArgumentConverter{}

	// Act
	result, err := converter.convertToString(42)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "42", result)
}

func TestConvertToString_FromFloat(t *testing.T) {
	// Arrange
	converter := &ArgumentConverter{}

	// Act
	result, err := converter.convertToString(3.14159)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "3.14159", result)
}

func TestConvertToString_UnsupportedType(t *testing.T) {
	// Arrange
	converter := &ArgumentConverter{}

	// Act
	result, err := converter.convertToString([]int{1, 2, 3})

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "", result)
	assert.Contains(t, err.Error(), "cannot convert")
}

// ===========================================
// UNIT TESTS FOR ArgumentConverter.convertToBool
// ===========================================

func TestConvertToBool_FromBool(t *testing.T) {
	// Arrange
	converter := &ArgumentConverter{}

	tests := []struct {
		input    bool
		expected bool
	}{
		{true, true},
		{false, false},
	}

	for _, test := range tests {
		// Act
		result, err := converter.convertToBool(test.input)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, test.expected, result)
	}
}

func TestConvertToBool_FromString(t *testing.T) {
	// Arrange
	converter := &ArgumentConverter{}

	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1", true},
		{"0", false},
	}

	for _, test := range tests {
		// Act
		result, err := converter.convertToBool(test.input)

		// Assert
		assert.NoError(t, err, "Failed for input: %s", test.input)
		assert.Equal(t, test.expected, result, "Failed for input: %s", test.input)
	}
}

func TestConvertToBool_FromFloat(t *testing.T) {
	// Arrange
	converter := &ArgumentConverter{}

	tests := []struct {
		input    float64
		expected bool
	}{
		{1.0, true},
		{0.0, false},
		{-1.0, true},
		{100.5, true},
	}

	for _, test := range tests {
		// Act
		result, err := converter.convertToBool(test.input)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, test.expected, result)
	}
}

func TestConvertToBool_InvalidString(t *testing.T) {
	// Arrange
	converter := &ArgumentConverter{}

	// Act
	result, err := converter.convertToBool("maybe")

	// Assert
	assert.Error(t, err)
	assert.False(t, result)
}

func TestConvertToBool_UnsupportedType(t *testing.T) {
	// Arrange
	converter := &ArgumentConverter{}

	// Act
	result, err := converter.convertToBool([]string{"true"})

	// Assert
	assert.Error(t, err)
	assert.False(t, result)
	assert.Contains(t, err.Error(), "cannot convert")
}

// ===========================================
// UNIT TESTS FOR ArgumentConverter.convertToAddress
// ===========================================

func TestConvertToAddress_ValidHexAddress(t *testing.T) {
	// Arrange
	converter := &ArgumentConverter{}
	addressStr := "0x742d35cc6632c0532c718d329da9e8e0a7d4b7fa"

	// Act
	result, err := converter.convertToAddress(addressStr)

	// Assert
	assert.NoError(t, err)
	// Use case-insensitive comparison since Ethereum addresses are checksummed
	assert.True(t, strings.EqualFold(addressStr, result.Hex()))
}

func TestConvertToAddress_InvalidAddress(t *testing.T) {
	// Arrange
	converter := &ArgumentConverter{}

	// Act
	result, err := converter.convertToAddress("invalid-address")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, common.Address{}, result)
	assert.Contains(t, err.Error(), "invalid")
}

func TestConvertToAddress_UnsupportedType(t *testing.T) {
	// Arrange
	converter := &ArgumentConverter{}

	// Act
	result, err := converter.convertToAddress(123)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, common.Address{}, result)
	assert.Contains(t, err.Error(), "cannot convert")
}
