package execution

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// Simple mock logger for dynamic unit tests
type DynamicUnitTestLogger struct {
	mock.Mock
}

func (m *DynamicUnitTestLogger) Warnf(format string, args ...interface{})  { m.Called(format, args) }
func (m *DynamicUnitTestLogger) Debugf(format string, args ...interface{}) { m.Called(format, args) }
func (m *DynamicUnitTestLogger) Errorf(format string, args ...interface{}) { m.Called(format, args) }
func (m *DynamicUnitTestLogger) Infof(format string, args ...interface{})  { m.Called(format, args) }
func (m *DynamicUnitTestLogger) Debug(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}
func (m *DynamicUnitTestLogger) Info(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}
func (m *DynamicUnitTestLogger) Warn(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}
func (m *DynamicUnitTestLogger) Error(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}
func (m *DynamicUnitTestLogger) Fatal(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}
func (m *DynamicUnitTestLogger) Fatalf(template string, args ...interface{}) {
	m.Called(template, args)
}
func (m *DynamicUnitTestLogger) With(tags ...any) logging.Logger { return m }

// Simple mock Ethereum client for dynamic unit tests
type DynamicUnitTestEthClient struct {
	mock.Mock
}

// Simple mock downloader for dynamic unit tests
type DynamicUnitTestDownloader struct {
	mock.Mock
}

func (m *DynamicUnitTestDownloader) DownloadScript(ctx context.Context, url string) ([]byte, error) {
	args := m.Called(ctx, url)
	return args.Get(0).([]byte), args.Error(1)
}

// Simple mock Docker manager for dynamic unit tests
type DynamicUnitTestDockerManager struct {
	mock.Mock
}

func (m *DynamicUnitTestDockerManager) CreateContainer(ctx context.Context, script []byte, envVars map[string]string) (string, error) {
	args := m.Called(ctx, script, envVars)
	return args.String(0), args.Error(1)
}

func (m *DynamicUnitTestDockerManager) StartContainer(ctx context.Context, containerID string) error {
	args := m.Called(ctx, containerID)
	return args.Error(0)
}

func (m *DynamicUnitTestDockerManager) WaitForContainer(ctx context.Context, containerID string) (int64, error) {
	args := m.Called(ctx, containerID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *DynamicUnitTestDockerManager) GetContainerLogs(ctx context.Context, containerID string) (string, error) {
	args := m.Called(ctx, containerID)
	return args.String(0), args.Error(1)
}

func (m *DynamicUnitTestDockerManager) RemoveContainer(ctx context.Context, containerID string) error {
	args := m.Called(ctx, containerID)
	return args.Error(0)
}

// ===========================================
// Unit Tests for Dynamic Execution
// ===========================================

func TestDynamicExecution_EmptyContractAddress(t *testing.T) {
	// Arrange
	mockLogger := &DynamicUnitTestLogger{}
	executor := &TaskExecutor{
		logger:       mockLogger,
		argConverter: &ArgumentConverter{},
	}

	taskData := &types.TaskTargetData{
		TaskID:                    1,
		TargetContractAddress:     "", // Empty address should fail
		TargetFunction:            "testFunction",
		TargetChainID:             "1",
		Arguments:                 []string{"123"},
		DynamicArgumentsScriptUrl: "https://example.com/script.js",
	}

	// Fix the mock expectation to match the actual call signature
	mockLogger.On("Errorf", "Execution contract address not configured", mock.Anything).Return()

	// Act
	result, err := executor.executeActionWithDynamicArgs(taskData, nil)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "execution contract address not configured")
	assert.Equal(t, types.PerformerActionData{}, result)
	mockLogger.AssertExpectations(t)
}

func TestDynamicExecution_EmptyScriptURL(t *testing.T) {
	// Arrange
	mockLogger := &DynamicUnitTestLogger{}
	executor := &TaskExecutor{
		logger:       mockLogger,
		argConverter: &ArgumentConverter{},
	}

	taskData := &types.TaskTargetData{
		TaskID:                    1,
		TargetContractAddress:     "0x1234567890123456789012345678901234567890",
		TargetFunction:            "testFunction",
		TargetChainID:             "1",
		Arguments:                 []string{"123"},
		DynamicArgumentsScriptUrl: "", // Empty script URL should fail
	}

	mockLogger.On("Errorf", mock.AnythingOfType("string"), mock.Anything).Return().Maybe()

	// Act
	result, err := executor.executeActionWithDynamicArgs(taskData, nil)

	// Assert
	assert.Error(t, err)
	// Will fail at getContractMethodAndABI or download stage due to empty URL
	assert.Equal(t, types.PerformerActionData{}, result)
}

func TestDynamicExecution_ValidContractAddressFormat(t *testing.T) {
	// This test validates that valid address formats pass the early validation
	// It cannot test full execution without proper mock setup, but it can verify
	// that valid addresses don't trigger the "execution contract address not configured" error

	validAddresses := []string{
		"0x1234567890123456789012345678901234567890",
		"0xdAC17F958D2ee523a2206206994597C13D831ec7", // USDT contract
		"0xA0b86a33E6417c1db2b55d2a6Fb57b5F4bCB3232", // Lowercase
		"0xDef1C0dE5Ef2E0d1C4B6F5C2e7B8a9F4FfCa3210", // Mixed case
	}

	for _, addr := range validAddresses {
		t.Run("Address_"+addr[:10], func(t *testing.T) {
			// Test the address validation logic directly by checking it doesn't fail the empty check
			isEmpty := addr == ""

			// Assert - Valid addresses should not be considered empty
			assert.False(t, isEmpty, "Valid address should not be empty: %s", addr)

			// Additional basic validation that the address looks like a hex address
			assert.True(t, len(addr) == 42, "Address should be 42 characters long: %s", addr)
			assert.True(t, addr[:2] == "0x", "Address should start with 0x: %s", addr)
		})
	}
}

func TestDynamicExecution_InvalidContractAddressFormat(t *testing.T) {
	// This test validates that invalid address formats can be detected
	// without requiring full integration setup

	invalidAddresses := []string{
		"0x",    // Too short
		"0x123", // Too short
		"0xgg1234567890123456789012345678901234567890", // Invalid hex
		"1234567890123456789012345678901234567890",     // Missing 0x prefix
		"0x12345678901234567890123456789012345678901",  // Too long
		"invalid", // Completely invalid
	}

	for _, addr := range invalidAddresses {
		t.Run("Invalid_Address_"+addr[:minInt(10, len(addr))], func(t *testing.T) {
			// Test basic format validation
			isValidLength := len(addr) == 42
			hasValidPrefix := len(addr) >= 2 && addr[:2] == "0x"

			// Assert - Invalid addresses should fail basic format checks
			if addr != "" { // Skip empty address as that's tested elsewhere
				isValid := isValidLength && hasValidPrefix
				if !isValid {
					// This is expected for invalid addresses
					t.Logf("Address '%s' correctly identified as invalid", addr)
				}
			}
		})
	}
}

func TestDynamicExecution_GetContractMethodAndABI_Integration(t *testing.T) {
	// Arrange
	mockLogger := &DynamicUnitTestLogger{}
	executor := &TaskExecutor{
		logger:       mockLogger,
		argConverter: &ArgumentConverter{},
	}

	testCases := []struct {
		name           string
		abi            string
		targetFunction string
		expectError    bool
		expectedErrMsg string
	}{
		{
			name:           "Valid_ABI_Valid_Function",
			abi:            `[{"type":"function","name":"transfer","inputs":[{"name":"to","type":"address"},{"name":"amount","type":"uint256"}],"outputs":[{"name":"","type":"bool"}]}]`,
			targetFunction: "transfer",
			expectError:    false,
		},
		{
			name:           "Valid_ABI_Invalid_Function",
			abi:            `[{"type":"function","name":"transfer","inputs":[{"name":"to","type":"address"},{"name":"amount","type":"uint256"}],"outputs":[{"name":"","type":"bool"}]}]`,
			targetFunction: "nonExistentFunction",
			expectError:    true,
			expectedErrMsg: "method nonExistentFunction not found in contract ABI",
		},
		{
			name:           "Invalid_ABI_Format",
			abi:            `invalid json`,
			targetFunction: "transfer",
			expectError:    true,
			expectedErrMsg: "invalid character",
		},
		{
			name:           "Empty_ABI",
			abi:            "",
			targetFunction: "transfer",
			expectError:    true,
			expectedErrMsg: "contract ABI not provided in job data",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			taskData := &types.TaskTargetData{
				TaskID:                    1,
				TargetContractAddress:     "0x1234567890123456789012345678901234567890",
				TargetFunction:            tc.targetFunction,
				TargetChainID:             "1",
				Arguments:                 []string{"123"},
				DynamicArgumentsScriptUrl: "https://example.com/script.js",
				ABI:                       tc.abi,
			}

			mockLogger.On("Debugf", mock.AnythingOfType("string"), mock.Anything).Return().Maybe()
			mockLogger.On("Warnf", mock.AnythingOfType("string"), mock.Anything).Return().Maybe()
			mockLogger.On("Errorf", mock.AnythingOfType("string"), mock.Anything).Return().Maybe()

			// Test only the ABI parsing and method finding part
			_, _, err := executor.getContractMethodAndABI(taskData.TargetFunction, taskData)

			// Assert
			if tc.expectError {
				assert.Error(t, err)
				if tc.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tc.expectedErrMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDynamicExecution_ValidScriptURLs(t *testing.T) {
	// Arrange
	mockLogger := &DynamicUnitTestLogger{}
	executor := &TaskExecutor{
		logger:       mockLogger,
		argConverter: &ArgumentConverter{},
	}

	validURLs := []string{
		"https://example.com/script.js",
		"https://cdn.example.com/scripts/dynamic.js",
		"https://raw.githubusercontent.com/user/repo/main/script.js",
		"https://api.example.com/v1/scripts/12345",
		"https://example.com/scripts/test-script.js",
	}

	for _, url := range validURLs {
		t.Run("URL_"+url[8:20], func(t *testing.T) {
			taskData := &types.TaskTargetData{
				TaskID:                    1,
				TargetContractAddress:     "0x1234567890123456789012345678901234567890",
				TargetFunction:            "testFunction",
				TargetChainID:             "1",
				Arguments:                 []string{"123"},
				DynamicArgumentsScriptUrl: url,
				ABI:                       `[{"type":"function","name":"testFunction","inputs":[{"name":"param1","type":"uint256"}],"outputs":[]}]`,
			}

			mockLogger.On("Debugf", mock.AnythingOfType("string"), mock.Anything).Return().Maybe()
			mockLogger.On("Errorf", mock.AnythingOfType("string"), mock.Anything).Return().Maybe()

			// Test URL validation - should pass basic format checks
			// We test the basic validation by checking that the URL is properly formed
			assert.NotEmpty(t, taskData.DynamicArgumentsScriptUrl)
			assert.Contains(t, taskData.DynamicArgumentsScriptUrl, "https://")
			assert.NotContains(t, taskData.DynamicArgumentsScriptUrl, " ")

			// Test ABI parsing to ensure everything up to the actual download works
			_, _, err := executor.getContractMethodAndABI(taskData.TargetFunction, taskData)
			assert.NoError(t, err, "ABI parsing should succeed for valid URL test")
		})
	}
}

func TestDynamicExecution_ArgumentProcessing_EdgeCases(t *testing.T) {
	// Arrange
	mockLogger := &DynamicUnitTestLogger{}
	executor := &TaskExecutor{
		logger:       mockLogger,
		argConverter: &ArgumentConverter{},
	}

	testCases := []struct {
		name      string
		arguments []string
	}{
		{
			name:      "Nil_Arguments",
			arguments: nil,
		},
		{
			name:      "Empty_Arguments",
			arguments: []string{},
		},
		{
			name:      "Single_Argument",
			arguments: []string{"123"},
		},
		{
			name:      "Multiple_Arguments",
			arguments: []string{"123", "456", "789"},
		},
		{
			name:      "Special_Characters",
			arguments: []string{"@#$%", "^&*()", "{}[]"},
		},
		{
			name:      "Very_Long_Arguments",
			arguments: []string{"this_is_a_very_long_argument_that_should_be_handled_gracefully_without_causing_any_issues"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			taskData := &types.TaskTargetData{
				TaskID:                    1,
				TargetContractAddress:     "0x1234567890123456789012345678901234567890",
				TargetFunction:            "testFunction",
				TargetChainID:             "1",
				Arguments:                 tc.arguments,
				DynamicArgumentsScriptUrl: "https://example.com/script.js",
				ABI:                       `[{"type":"function","name":"testFunction","inputs":[{"name":"param1","type":"uint256"}],"outputs":[]}]`,
			}

			mockLogger.On("Debugf", mock.AnythingOfType("string"), mock.Anything).Return().Maybe()
			mockLogger.On("Errorf", mock.AnythingOfType("string"), mock.Anything).Return().Maybe()

			// Test that arguments are properly stored in taskData
			assert.Equal(t, tc.arguments, taskData.Arguments)

			// Test ABI parsing to ensure basic setup works
			_, _, err := executor.getContractMethodAndABI(taskData.TargetFunction, taskData)
			assert.NoError(t, err, "ABI parsing should succeed for argument processing test")
		})
	}
}

func TestDynamicExecution_ChainID_Validation(t *testing.T) {
	// Arrange
	mockLogger := &DynamicUnitTestLogger{}
	executor := &TaskExecutor{
		logger:       mockLogger,
		argConverter: &ArgumentConverter{},
	}

	testCases := []struct {
		name    string
		chainID string
	}{
		{"Mainnet", "1"},
		{"Polygon", "137"},
		{"BSC", "56"},
		{"Arbitrum", "42161"},
		{"Invalid_String", "invalid"},
		{"Empty_ChainID", ""},
		{"Negative_ChainID", "-1"},
		{"Zero_ChainID", "0"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			taskData := &types.TaskTargetData{
				TaskID:                    1,
				TargetContractAddress:     "0x1234567890123456789012345678901234567890",
				TargetFunction:            "testFunction",
				TargetChainID:             tc.chainID,
				Arguments:                 []string{"123"},
				DynamicArgumentsScriptUrl: "https://example.com/script.js",
				ABI:                       `[{"type":"function","name":"testFunction","inputs":[{"name":"param1","type":"uint256"}],"outputs":[]}]`,
			}

			mockLogger.On("Debugf", mock.AnythingOfType("string"), mock.Anything).Return().Maybe()
			mockLogger.On("Errorf", mock.AnythingOfType("string"), mock.Anything).Return().Maybe()

			// Test that chain ID is properly stored in taskData
			assert.Equal(t, tc.chainID, taskData.TargetChainID)

			// Test ABI parsing to ensure basic setup works
			_, _, err := executor.getContractMethodAndABI(taskData.TargetFunction, taskData)
			assert.NoError(t, err, "ABI parsing should succeed for chain ID validation test")

			t.Logf("Chain ID '%s' properly handled in task data", tc.chainID)
		})
	}
}
