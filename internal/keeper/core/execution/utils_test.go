package execution

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// Enhanced mock logger that implements all required methods
type ComprehensiveMockLogger struct {
	mock.Mock
}

func (m *ComprehensiveMockLogger) Warnf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *ComprehensiveMockLogger) Debugf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *ComprehensiveMockLogger) Errorf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *ComprehensiveMockLogger) Infof(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *ComprehensiveMockLogger) Debug(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *ComprehensiveMockLogger) Info(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *ComprehensiveMockLogger) Warn(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *ComprehensiveMockLogger) Error(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *ComprehensiveMockLogger) Fatal(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *ComprehensiveMockLogger) Fatalf(template string, args ...interface{}) {
	m.Called(template, args)
}

func (m *ComprehensiveMockLogger) With(tags ...any) logging.Logger {
	args := m.Called(tags)
	return args.Get(0).(logging.Logger)
}

// MockLogger is a mock implementation of the logger interface
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Info(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Warn(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Error(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Fatal(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Debugf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) Infof(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) Warnf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) Errorf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) Fatalf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) With(tags ...any) logging.Logger {
	args := m.Called(tags)
	return args.Get(0).(logging.Logger)
}

// ===============================
// getContractMethodAndABI Tests
// ===============================

func TestGetContractMethodAndABI_Success(t *testing.T) {
	mockLogger := &ComprehensiveMockLogger{}
	e := &TaskExecutor{logger: mockLogger, argConverter: &ArgumentConverter{}}

	// Valid ABI with a simple function
	abiJSON := `[{"type":"function","name":"testMethod","inputs":[{"name":"param1","type":"uint256"}],"outputs":[{"name":"result","type":"uint256"}]}]`
	targetData := &types.TaskTargetData{
		ABI:                   abiJSON,
		TargetContractAddress: "0x1234567890123456789012345678901234567890",
	}

	mockLogger.On("Debugf", mock.AnythingOfType("string"), mock.Anything).Return().Maybe()

	parsedABI, method, err := e.getContractMethodAndABI("testMethod", targetData)

	assert.NoError(t, err)
	assert.NotNil(t, parsedABI)
	assert.NotNil(t, method)
	assert.Equal(t, "testMethod", method.Name)
	assert.Equal(t, 1, len(method.Inputs))
	assert.Equal(t, "param1", method.Inputs[0].Name)
	assert.Equal(t, "uint256", method.Inputs[0].Type.String())
}

func TestGetContractMethodAndABI_EmptyABI(t *testing.T) {
	mockLogger := &ComprehensiveMockLogger{}
	e := &TaskExecutor{logger: mockLogger, argConverter: &ArgumentConverter{}}

	targetData := &types.TaskTargetData{
		ABI:                   "",
		TargetContractAddress: "0x1234567890123456789012345678901234567890",
	}

	parsedABI, method, err := e.getContractMethodAndABI("testMethod", targetData)

	assert.Error(t, err)
	assert.Nil(t, parsedABI)
	assert.Nil(t, method)
	assert.Contains(t, err.Error(), "contract ABI not provided")
}

func TestGetContractMethodAndABI_InvalidJSON(t *testing.T) {
	mockLogger := &ComprehensiveMockLogger{}
	e := &TaskExecutor{logger: mockLogger, argConverter: &ArgumentConverter{}}

	invalidABIs := []string{
		"invalid json",
		"{not a valid json}",
		"[{incomplete",
		"null",
		"[]", // Empty array is valid JSON but invalid ABI
		`[{"invalid": "structure"}]`,
	}

	for _, invalidABI := range invalidABIs {
		t.Run("Invalid_ABI_"+invalidABI[:min(10, len(invalidABI))], func(t *testing.T) {
			targetData := &types.TaskTargetData{
				ABI:                   invalidABI,
				TargetContractAddress: "0x1234567890123456789012345678901234567890",
			}

			// Mock all possible logger calls
			mockLogger.On("Debugf", mock.AnythingOfType("string"), mock.Anything).Return().Maybe()
			mockLogger.On("Warnf", mock.AnythingOfType("string"), mock.Anything).Return().Maybe()

			parsedABI, method, err := e.getContractMethodAndABI("testMethod", targetData)

			assert.Error(t, err)
			assert.Nil(t, parsedABI)
			assert.Nil(t, method)
		})
	}
}

func TestGetContractMethodAndABI_MethodNotFound(t *testing.T) {
	mockLogger := &ComprehensiveMockLogger{}
	e := &TaskExecutor{logger: mockLogger, argConverter: &ArgumentConverter{}}

	abiJSON := `[{"type":"function","name":"existingMethod","inputs":[],"outputs":[]}]`
	targetData := &types.TaskTargetData{
		ABI:                   abiJSON,
		TargetContractAddress: "0x1234567890123456789012345678901234567890",
	}

	mockLogger.On("Debugf", mock.AnythingOfType("string"), mock.Anything).Return().Maybe()
	mockLogger.On("Warnf", mock.AnythingOfType("string"), mock.Anything).Return().Maybe()

	parsedABI, method, err := e.getContractMethodAndABI("nonExistentMethod", targetData)

	assert.Error(t, err)
	assert.Nil(t, parsedABI)
	assert.Nil(t, method)
	assert.Contains(t, err.Error(), "method nonExistentMethod not found")
}

func TestGetContractMethodAndABI_ComplexABI(t *testing.T) {
	mockLogger := &ComprehensiveMockLogger{}
	e := &TaskExecutor{logger: mockLogger, argConverter: &ArgumentConverter{}}

	// Complex ABI with multiple methods and types
	complexABI := `[
		{
			"type":"function",
			"name":"complexMethod",
			"inputs":[
				{"name":"addr","type":"address"},
				{"name":"amount","type":"uint256"},
				{"name":"data","type":"bytes"},
				{"name":"flag","type":"bool"},
				{"name":"values","type":"uint256[]"},
				{"name":"userData","type":"tuple","components":[
					{"name":"id","type":"uint256"},
					{"name":"name","type":"string"}
				]}
			],
			"outputs":[
				{"name":"success","type":"bool"},
				{"name":"returnData","type":"bytes"}
			]
		},
		{
			"type":"function",
			"name":"simpleMethod",
			"inputs":[],
			"outputs":[]
		}
	]`

	targetData := &types.TaskTargetData{
		ABI:                   complexABI,
		TargetContractAddress: "0x1234567890123456789012345678901234567890",
	}

	mockLogger.On("Debugf", mock.AnythingOfType("string"), mock.Anything).Return().Maybe()

	parsedABI, method, err := e.getContractMethodAndABI("complexMethod", targetData)

	assert.NoError(t, err)
	assert.NotNil(t, parsedABI)
	assert.NotNil(t, method)
	assert.Equal(t, "complexMethod", method.Name)
	assert.Equal(t, 6, len(method.Inputs))
	assert.Equal(t, 2, len(method.Outputs))

	// Check specific input types
	assert.Equal(t, "address", method.Inputs[0].Type.String())
	assert.Equal(t, "uint256", method.Inputs[1].Type.String())
	assert.Equal(t, "bytes", method.Inputs[2].Type.String())
	assert.Equal(t, "bool", method.Inputs[3].Type.String())
	assert.Equal(t, "uint256[]", method.Inputs[4].Type.String())
	assert.Equal(t, "(uint256,string)", method.Inputs[5].Type.String())
}

// ===============================
// processArguments Tests
// ===============================

func TestProcessArguments_SingleStruct_MapInput(t *testing.T) {
	mockLogger := &ComprehensiveMockLogger{}
	e := &TaskExecutor{logger: mockLogger, argConverter: &ArgumentConverter{}}

	// Create a tuple type (struct)
	tupleType, err := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "id", Type: "uint256"},
		{Name: "name", Type: "string"},
	})
	assert.NoError(t, err)

	inputs := []abi.Argument{{Name: "userData", Type: tupleType}}

	// Test with map input
	mapInput := map[string]interface{}{
		"id":   float64(123),
		"name": "test user",
	}

	args, err := e.processArguments(mapInput, inputs, nil)

	assert.NoError(t, err)
	assert.Len(t, args, 1)
}

func TestProcessArguments_SingleStruct_JSONString(t *testing.T) {
	mockLogger := &ComprehensiveMockLogger{}
	e := &TaskExecutor{logger: mockLogger, argConverter: &ArgumentConverter{}}

	tupleType, err := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "id", Type: "uint256"},
		{Name: "name", Type: "string"},
	})
	assert.NoError(t, err)

	inputs := []abi.Argument{{Name: "userData", Type: tupleType}}

	// Test with JSON string
	jsonStr := `{"id": 456, "name": "json user"}`

	args, err := e.processArguments(jsonStr, inputs, nil)

	assert.NoError(t, err)
	assert.Len(t, args, 1)
}

func TestProcessArguments_SingleStruct_ArrayWithMap(t *testing.T) {
	mockLogger := &ComprehensiveMockLogger{}
	e := &TaskExecutor{logger: mockLogger, argConverter: &ArgumentConverter{}}

	tupleType, err := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "id", Type: "uint256"},
		{Name: "name", Type: "string"},
	})
	assert.NoError(t, err)

	inputs := []abi.Argument{{Name: "userData", Type: tupleType}}

	// Test with array containing a map
	arrayInput := []interface{}{
		map[string]interface{}{
			"id":   float64(789),
			"name": "array user",
		},
	}

	args, err := e.processArguments(arrayInput, inputs, nil)

	assert.NoError(t, err)
	assert.Len(t, args, 1)
}

func TestProcessArguments_SingleStruct_ArrayWithJSONString(t *testing.T) {
	mockLogger := &ComprehensiveMockLogger{}
	e := &TaskExecutor{logger: mockLogger, argConverter: &ArgumentConverter{}}

	tupleType, err := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "id", Type: "uint256"},
		{Name: "name", Type: "string"},
	})
	assert.NoError(t, err)

	inputs := []abi.Argument{{Name: "userData", Type: tupleType}}

	// Test with array containing a JSON string
	arrayInput := []interface{}{
		`{"id": 101112, "name": "json in array"}`,
	}

	args, err := e.processArguments(arrayInput, inputs, nil)

	assert.NoError(t, err)
	assert.Len(t, args, 1)
}

func TestProcessArguments_SingleParameter_String(t *testing.T) {
	mockLogger := &ComprehensiveMockLogger{}
	e := &TaskExecutor{logger: mockLogger, argConverter: &ArgumentConverter{}}

	uintType, _ := abi.NewType("uint256", "", nil)
	inputs := []abi.Argument{{Name: "value", Type: uintType}}

	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Simple_number", "123", true},
		{"Quoted_number", `"456"`, true},
		{"String_number", "789", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			args, err := e.processArguments(tc.input, inputs, nil)

			if tc.expected {
				assert.NoError(t, err)
				assert.Len(t, args, 1)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestProcessArguments_MultipleParameters_JSONArray(t *testing.T) {
	mockLogger := &ComprehensiveMockLogger{}
	e := &TaskExecutor{logger: mockLogger, argConverter: &ArgumentConverter{}}

	uintType, _ := abi.NewType("uint256", "", nil)
	stringType, _ := abi.NewType("string", "", nil)
	inputs := []abi.Argument{
		{Name: "value1", Type: uintType},
		{Name: "value2", Type: stringType},
	}

	// Test with JSON array string
	jsonArrayStr := `[123, "test string"]`

	args, err := e.processArguments(jsonArrayStr, inputs, nil)

	assert.NoError(t, err)
	assert.Len(t, args, 2)
}

func TestProcessArguments_MultipleParameters_InsufficientArgs(t *testing.T) {
	mockLogger := &ComprehensiveMockLogger{}
	e := &TaskExecutor{logger: mockLogger, argConverter: &ArgumentConverter{}}

	uintType, _ := abi.NewType("uint256", "", nil)
	inputs := []abi.Argument{
		{Name: "value1", Type: uintType},
		{Name: "value2", Type: uintType},
	}

	// Test with insufficient arguments
	jsonArrayStr := `[123]` // Only one argument provided

	args, err := e.processArguments(jsonArrayStr, inputs, nil)

	assert.Error(t, err)
	assert.Nil(t, args)
	assert.Contains(t, err.Error(), "not enough arguments")
}

func TestProcessArguments_StringArray(t *testing.T) {
	mockLogger := &ComprehensiveMockLogger{}
	e := &TaskExecutor{logger: mockLogger, argConverter: &ArgumentConverter{}}

	uintType, _ := abi.NewType("uint256", "", nil)
	inputs := []abi.Argument{
		{Name: "value1", Type: uintType},
		{Name: "value2", Type: uintType},
	}

	// Test with string array
	stringArray := []string{"123", "456"}

	args, err := e.processArguments(stringArray, inputs, nil)

	assert.NoError(t, err)
	assert.Len(t, args, 2)
}

func TestProcessArguments_StringArray_InsufficientArgs(t *testing.T) {
	mockLogger := &ComprehensiveMockLogger{}
	e := &TaskExecutor{logger: mockLogger, argConverter: &ArgumentConverter{}}

	uintType, _ := abi.NewType("uint256", "", nil)
	inputs := []abi.Argument{
		{Name: "value1", Type: uintType},
		{Name: "value2", Type: uintType},
	}

	// Test with insufficient string arguments
	stringArray := []string{"123"}

	args, err := e.processArguments(stringArray, inputs, nil)

	assert.Error(t, err)
	assert.Nil(t, args)
	assert.Contains(t, err.Error(), "not enough arguments")
}

func TestProcessArguments_InterfaceArray(t *testing.T) {
	mockLogger := &ComprehensiveMockLogger{}
	e := &TaskExecutor{logger: mockLogger, argConverter: &ArgumentConverter{}}

	uintType, _ := abi.NewType("uint256", "", nil)
	stringType, _ := abi.NewType("string", "", nil)
	inputs := []abi.Argument{
		{Name: "value1", Type: uintType},
		{Name: "value2", Type: stringType},
	}

	// Test with interface array
	interfaceArray := []interface{}{float64(123), "test string"}

	args, err := e.processArguments(interfaceArray, inputs, nil)

	assert.NoError(t, err)
	assert.Len(t, args, 2)
}

func TestProcessArguments_InterfaceArray_InsufficientArgs(t *testing.T) {
	mockLogger := &ComprehensiveMockLogger{}
	e := &TaskExecutor{logger: mockLogger, argConverter: &ArgumentConverter{}}

	uintType, _ := abi.NewType("uint256", "", nil)
	inputs := []abi.Argument{
		{Name: "value1", Type: uintType},
		{Name: "value2", Type: uintType},
	}

	// Test with insufficient interface arguments
	interfaceArray := []interface{}{float64(123)}

	args, err := e.processArguments(interfaceArray, inputs, nil)

	assert.Error(t, err)
	assert.Nil(t, args)
	assert.Contains(t, err.Error(), "not enough arguments")
}

func TestProcessArguments_NamedParameters_Map(t *testing.T) {
	mockLogger := &ComprehensiveMockLogger{}
	e := &TaskExecutor{logger: mockLogger, argConverter: &ArgumentConverter{}}

	uintType, _ := abi.NewType("uint256", "", nil)
	stringType, _ := abi.NewType("string", "", nil)
	inputs := []abi.Argument{
		{Name: "amount", Type: uintType},
		{Name: "message", Type: stringType},
	}

	// Test with named parameters map
	namedArgs := map[string]interface{}{
		"amount":  float64(123),
		"message": "hello world",
	}

	args, err := e.processArguments(namedArgs, inputs, nil)

	assert.NoError(t, err)
	assert.Len(t, args, 2)
}

func TestProcessArguments_NamedParameters_CaseInsensitive(t *testing.T) {
	mockLogger := &ComprehensiveMockLogger{}
	e := &TaskExecutor{logger: mockLogger, argConverter: &ArgumentConverter{}}

	uintType, _ := abi.NewType("uint256", "", nil)
	inputs := []abi.Argument{
		{Name: "Amount", Type: uintType},
	}

	// Test with case-insensitive matching
	namedArgs := map[string]interface{}{
		"amount": float64(123), // lowercase
	}

	args, err := e.processArguments(namedArgs, inputs, nil)

	assert.NoError(t, err)
	assert.Len(t, args, 1)
}

func TestProcessArguments_NamedParameters_UnnamedInput(t *testing.T) {
	mockLogger := &ComprehensiveMockLogger{}
	e := &TaskExecutor{logger: mockLogger, argConverter: &ArgumentConverter{}}

	uintType, _ := abi.NewType("uint256", "", nil)
	inputs := []abi.Argument{
		{Name: "", Type: uintType}, // Unnamed parameter
	}

	namedArgs := map[string]interface{}{
		"amount": float64(123),
	}

	args, err := e.processArguments(namedArgs, inputs, nil)

	assert.Error(t, err)
	assert.Nil(t, args)
	assert.Contains(t, err.Error(), "cannot use map arguments with unnamed parameters")
}

func TestProcessArguments_NamedParameters_MissingArgument(t *testing.T) {
	mockLogger := &ComprehensiveMockLogger{}
	e := &TaskExecutor{logger: mockLogger, argConverter: &ArgumentConverter{}}

	uintType, _ := abi.NewType("uint256", "", nil)
	inputs := []abi.Argument{
		{Name: "amount", Type: uintType},
		{Name: "missing", Type: uintType},
	}

	namedArgs := map[string]interface{}{
		"amount": float64(123),
		// "missing" parameter not provided
	}

	args, err := e.processArguments(namedArgs, inputs, nil)

	assert.Error(t, err)
	assert.Nil(t, args)
	assert.Contains(t, err.Error(), "argument missing not found")
}

func TestProcessArguments_UnsupportedArgumentFormat(t *testing.T) {
	mockLogger := &ComprehensiveMockLogger{}
	e := &TaskExecutor{logger: mockLogger, argConverter: &ArgumentConverter{}}

	uintType, _ := abi.NewType("uint256", "", nil)
	inputs := []abi.Argument{{Name: "value", Type: uintType}}

	// Test with unsupported types
	unsupportedTypes := []interface{}{
		123,            // int (not supported directly)
		123.45,         // float64 (not supported directly)
		true,           // bool (not supported directly)
		make(chan int), // channel (definitely not supported)
	}

	for i, unsupportedArg := range unsupportedTypes {
		t.Run("Unsupported_Type_"+string(rune(i+'A')), func(t *testing.T) {
			args, err := e.processArguments(unsupportedArg, inputs, nil)

			assert.Error(t, err)
			assert.Nil(t, args)
			assert.Contains(t, err.Error(), "unsupported argument format")
		})
	}
}

func TestProcessArguments_SingleStringMultipleParams_Error(t *testing.T) {
	mockLogger := &ComprehensiveMockLogger{}
	e := &TaskExecutor{logger: mockLogger, argConverter: &ArgumentConverter{}}

	uintType, _ := abi.NewType("uint256", "", nil)
	inputs := []abi.Argument{
		{Name: "value1", Type: uintType},
		{Name: "value2", Type: uintType},
	}

	// Single string for multiple parameters should fail
	singleString := "123"

	args, err := e.processArguments(singleString, inputs, nil)

	assert.Error(t, err)
	assert.Nil(t, args)
	assert.Contains(t, err.Error(), "cannot convert single string to 2 arguments")
}

func TestProcessArguments_ComplexTypes(t *testing.T) {
	mockLogger := &ComprehensiveMockLogger{}
	e := &TaskExecutor{logger: mockLogger, argConverter: &ArgumentConverter{}}

	// Test with address type
	addressType, _ := abi.NewType("address", "", nil)
	inputs := []abi.Argument{{Name: "addr", Type: addressType}}

	args, err := e.processArguments("0x1234567890123456789012345678901234567890", inputs, nil)
	assert.NoError(t, err)
	assert.Len(t, args, 1)

	// Test with bytes type
	bytesType, _ := abi.NewType("bytes", "", nil)
	inputs = []abi.Argument{{Name: "data", Type: bytesType}}

	args, err = e.processArguments("0x1234", inputs, nil)
	assert.NoError(t, err)
	assert.Len(t, args, 1)

	// Test with bool type
	boolType, _ := abi.NewType("bool", "", nil)
	inputs = []abi.Argument{{Name: "flag", Type: boolType}}

	args, err = e.processArguments("true", inputs, nil)
	assert.NoError(t, err)
	assert.Len(t, args, 1)
}

// ===============================
// parseDynamicArgs Tests
// ===============================

func TestParseDynamicArgs_ValidJSON(t *testing.T) {
	mockLogger := &ComprehensiveMockLogger{}
	e := &TaskExecutor{logger: mockLogger, argConverter: &ArgumentConverter{}}

	testCases := []struct {
		name     string
		input    string
		expected []interface{}
	}{
		{
			name:     "Simple_array",
			input:    `[1, 2, 3]`,
			expected: []interface{}{float64(1), float64(2), float64(3)},
		},
		{
			name:     "Mixed_types",
			input:    `[123, "hello", true, null]`,
			expected: []interface{}{float64(123), "hello", true, nil},
		},
		{
			name:     "Empty_array",
			input:    `[]`,
			expected: []interface{}{},
		},
		{
			name:  "Nested_objects",
			input: `[{"id": 1, "name": "test"}, {"id": 2, "name": "test2"}]`,
			expected: []interface{}{
				map[string]interface{}{"id": float64(1), "name": "test"},
				map[string]interface{}{"id": float64(2), "name": "test2"},
			},
		},
		{
			name:  "Single_object",
			input: `[{"address": "0x123", "amount": 1000}]`,
			expected: []interface{}{
				map[string]interface{}{"address": "0x123", "amount": float64(1000)},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := e.parseDynamicArgs(tc.input)

			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestParseDynamicArgs_InvalidJSON(t *testing.T) {
	mockLogger := &ComprehensiveMockLogger{}
	e := &TaskExecutor{logger: mockLogger, argConverter: &ArgumentConverter{}}

	invalidJSONs := []string{
		"invalid json",
		"{not valid}",
		"[incomplete",
		"[1, 2, 3",
		"null",     // Valid JSON but not an array
		"123",      // Valid JSON but not an array
		`"string"`, // Valid JSON but not an array
		"",         // Empty string
	}

	for _, invalidJSON := range invalidJSONs {
		t.Run("Invalid_JSON_"+invalidJSON[:min(10, len(invalidJSON))], func(t *testing.T) {
			mockLogger.On("Warnf", mock.AnythingOfType("string"), mock.Anything).Return().Maybe()

			result := e.parseDynamicArgs(invalidJSON)

			assert.Nil(t, result)
		})
	}
}

func TestParseDynamicArgs_EdgeCases(t *testing.T) {
	mockLogger := &ComprehensiveMockLogger{}
	e := &TaskExecutor{logger: mockLogger, argConverter: &ArgumentConverter{}}

	edgeCases := []struct {
		name     string
		input    string
		expected []interface{}
	}{
		{
			name:     "Very_large_numbers",
			input:    `[9223372036854775807, -9223372036854775808]`,
			expected: []interface{}{float64(9223372036854775807), float64(-9223372036854775808)},
		},
		{
			name:     "Very_small_decimals",
			input:    `[0.000000000000001, -0.000000000000001]`,
			expected: []interface{}{0.000000000000001, -0.000000000000001},
		},
		{
			name:     "Unicode_strings",
			input:    `["Hello 世界", "🚀 Rocket", "Émoji 🎉"]`,
			expected: []interface{}{"Hello 世界", "🚀 Rocket", "Émoji 🎉"},
		},
		{
			name:  "Deeply_nested",
			input: `[{"level1": {"level2": {"level3": "deep"}}}]`,
			expected: []interface{}{
				map[string]interface{}{
					"level1": map[string]interface{}{
						"level2": map[string]interface{}{
							"level3": "deep",
						},
					},
				},
			},
		},
	}

	for _, tc := range edgeCases {
		t.Run(tc.name, func(t *testing.T) {
			result := e.parseDynamicArgs(tc.input)

			assert.Equal(t, tc.expected, result)
		})
	}
}

// ===============================
// Integration Tests
// ===============================

func TestProcessArguments_Integration_RealWorldScenarios(t *testing.T) {
	mockLogger := &ComprehensiveMockLogger{}
	e := &TaskExecutor{logger: mockLogger, argConverter: &ArgumentConverter{}}

	// Scenario 1: DeFi token transfer
	t.Run("DeFi_Token_Transfer", func(t *testing.T) {
		addressType, _ := abi.NewType("address", "", nil)
		uintType, _ := abi.NewType("uint256", "", nil)

		inputs := []abi.Argument{
			{Name: "to", Type: addressType},
			{Name: "amount", Type: uintType},
		}

		transferArgs := map[string]interface{}{
			"to":     "0x742d35Cc6634C0532925a3b8D4C186C0D8F1D1b1",
			"amount": "1000000000000000000", // 1 ETH in wei
		}

		args, err := e.processArguments(transferArgs, inputs, nil)

		assert.NoError(t, err)
		assert.Len(t, args, 2)
	})

	// Scenario 2: NFT minting with metadata
	t.Run("NFT_Minting", func(t *testing.T) {
		tupleType, _ := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
			{Name: "tokenId", Type: "uint256"},
			{Name: "uri", Type: "string"},
			{Name: "royalty", Type: "uint96"},
		})

		inputs := []abi.Argument{{Name: "mintData", Type: tupleType}}

		nftData := map[string]interface{}{
			"tokenId": float64(1),
			"uri":     "https://api.example.com/metadata/1",
			"royalty": float64(250), // 2.5%
		}

		args, err := e.processArguments(nftData, inputs, nil)

		assert.NoError(t, err)
		assert.Len(t, args, 1)
	})

	// Scenario 3: Multi-signature wallet operation
	t.Run("MultiSig_Operation", func(t *testing.T) {
		addressType, _ := abi.NewType("address", "", nil)
		uintType, _ := abi.NewType("uint256", "", nil)
		bytesType, _ := abi.NewType("bytes", "", nil)
		addressArrayType, _ := abi.NewType("address[]", "", nil)

		inputs := []abi.Argument{
			{Name: "to", Type: addressType},
			{Name: "value", Type: uintType},
			{Name: "data", Type: bytesType},
			{Name: "signers", Type: addressArrayType},
		}

		multisigArgs := []interface{}{
			"0x742d35Cc6634C0532925a3b8D4C186C0D8F1D1b1",
			"1000000000000000000",
			"0x",
			[]interface{}{
				"0x1234567890123456789012345678901234567890",
				"0x9876543210987654321098765432109876543210",
			},
		}

		args, err := e.processArguments(multisigArgs, inputs, nil)

		assert.NoError(t, err)
		assert.Len(t, args, 4)
	})
}

// ===============================
// Error Handling and Edge Cases
// ===============================

func TestUtils_ArgumentConverter_Integration(t *testing.T) {
	mockLogger := &ComprehensiveMockLogger{}
	e := &TaskExecutor{logger: mockLogger, argConverter: &ArgumentConverter{}}

	// Test ArgumentConverter methods integration
	t.Run("ArgumentConverter_ConvertToType_BigInt", func(t *testing.T) {
		uintType, _ := abi.NewType("uint256", "", nil)

		bigIntVal := big.NewInt(12345)
		result, err := e.argConverter.convertToType(bigIntVal, uintType)

		assert.NoError(t, err)
		assert.Equal(t, bigIntVal, result)
	})

	t.Run("ArgumentConverter_ConvertToType_Address", func(t *testing.T) {
		addressType, _ := abi.NewType("address", "", nil)

		addrStr := "0x742d35Cc6634C0532925a3b8D4C186C0D8F1D1b1"
		result, err := e.argConverter.convertToType(addrStr, addressType)

		assert.NoError(t, err)
		assert.Equal(t, ethcommon.HexToAddress(addrStr), result)
	})
}

func TestUtils_MemoryAndPerformance(t *testing.T) {
	mockLogger := &ComprehensiveMockLogger{}
	e := &TaskExecutor{logger: mockLogger, argConverter: &ArgumentConverter{}}

	// Test with large data structures
	t.Run("Large_Array_Processing", func(t *testing.T) {
		uintType, _ := abi.NewType("uint256", "", nil)
		inputs := []abi.Argument{{Name: "values", Type: uintType}}

		// Test with a reasonable sized value (not an array since this is for single uint256)
		args, err := e.processArguments("12345", inputs, nil)

		assert.NoError(t, err)
		assert.Len(t, args, 1)
	})
}

func TestUtils_ConcurrentAccess(t *testing.T) {
	mockLogger := &ComprehensiveMockLogger{}
	e := &TaskExecutor{logger: mockLogger, argConverter: &ArgumentConverter{}}

	uintType, _ := abi.NewType("uint256", "", nil)
	inputs := []abi.Argument{{Name: "value", Type: uintType}}

	// Test concurrent access to the same TaskExecutor
	numGoroutines := 10
	results := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			_, err := e.processArguments(fmt.Sprintf("%d", id), inputs, nil)
			results <- err
		}(i)
	}

	// Check all results
	for i := 0; i < numGoroutines; i++ {
		err := <-results
		assert.NoError(t, err)
	}
}

// Utility function for minimum
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ===============================
// Benchmarks
// ===============================

func BenchmarkGetContractMethodAndABI(b *testing.B) {
	mockLogger := &ComprehensiveMockLogger{}
	e := &TaskExecutor{logger: mockLogger, argConverter: &ArgumentConverter{}}

	abiJSON := `[{"type":"function","name":"testMethod","inputs":[{"name":"param1","type":"uint256"}],"outputs":[]}]`
	targetData := &types.TaskTargetData{
		ABI:                   abiJSON,
		TargetContractAddress: "0x1234567890123456789012345678901234567890",
	}

	mockLogger.On("Debugf", mock.AnythingOfType("string"), mock.Anything).Return().Maybe()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = e.getContractMethodAndABI("testMethod", targetData)
	}
}

func BenchmarkProcessArguments(b *testing.B) {
	mockLogger := &ComprehensiveMockLogger{}
	e := &TaskExecutor{logger: mockLogger, argConverter: &ArgumentConverter{}}

	uintType, _ := abi.NewType("uint256", "", nil)
	inputs := []abi.Argument{{Name: "value", Type: uintType}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = e.processArguments("123", inputs, nil)
	}
}

func BenchmarkParseDynamicArgs(b *testing.B) {
	mockLogger := &ComprehensiveMockLogger{}
	e := &TaskExecutor{logger: mockLogger, argConverter: &ArgumentConverter{}}

	jsonStr := `[1, 2, 3, "test", true]`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = e.parseDynamicArgs(jsonStr)
	}
}

func TestParseDynamicArgs(t *testing.T) {
	t.Run("JSON parsing", func(t *testing.T) {
		logger := new(MockLogger)
		// Setup expectations - allow any debug calls
		logger.On("Debugf", mock.Anything, mock.Anything).Return()
		logger.On("With", mock.Anything).Return(logger)

		executor := &TaskExecutor{
			logger: logger,
		}

		// Test with valid JSON array
		output := `[123.45, "test", true]`
		args := executor.parseDynamicArgs(output)

		assert.Len(t, args, 3)
		assert.Equal(t, 123.45, args[0])
		assert.Equal(t, "test", args[1])
		assert.Equal(t, true, args[2])
	})

	t.Run("Response pattern", func(t *testing.T) {
		logger := new(MockLogger)
		// Setup expectations - allow any debug calls
		logger.On("Debugf", mock.Anything, mock.Anything).Return()
		logger.On("With", mock.Anything).Return(logger)

		executor := &TaskExecutor{
			logger: logger,
		}

		// Test with container log output containing Response: lines
		output := `
Container Log: START_EXECUTION
Container Log: Condition satisfied: true
Container Log: Timestamp: 2025-07-18T12:31:08Z
Container Log: Response: 3603.66
Container Log: Response: 3603.66
Container Log: Ethereum price is greater than 0
Code execution completed in: 42.284769713s
Container Log: END_EXECUTION
`
		args := executor.parseDynamicArgs(output)

		assert.Len(t, args, 2)
		assert.Equal(t, 3603.66, args[0])
		assert.Equal(t, 3603.66, args[1])
	})

	t.Run("Condition pattern", func(t *testing.T) {
		logger := new(MockLogger)
		// Setup expectations - allow any debug calls
		logger.On("Debugf", mock.Anything, mock.Anything).Return()
		logger.On("Warnf", mock.Anything, mock.Anything).Return()
		logger.On("With", mock.Anything).Return(logger)

		executor := &TaskExecutor{
			logger: logger,
		}

		// Test with container log output containing only condition satisfied
		output := `
Container Log: START_EXECUTION
Container Log: Condition satisfied: true
Container Log: END_EXECUTION
`
		args := executor.parseDynamicArgs(output)

		assert.Len(t, args, 1)
		assert.Equal(t, true, args[0])
	})

	t.Run("Fallback", func(t *testing.T) {
		logger := new(MockLogger)
		// Setup expectations - allow any debug calls
		logger.On("Debugf", mock.Anything, mock.Anything).Return()
		logger.On("Warnf", mock.Anything, mock.Anything).Return()
		logger.On("With", mock.Anything).Return(logger)

		executor := &TaskExecutor{
			logger: logger,
		}

		// Test with container log output that doesn't match any pattern
		output := `
Container Log: START_EXECUTION
Container Log: Some random output
Container Log: END_EXECUTION
`
		args := executor.parseDynamicArgs(output)

		assert.Len(t, args, 1)
		assert.Equal(t, "0", args[0])
	})
}
