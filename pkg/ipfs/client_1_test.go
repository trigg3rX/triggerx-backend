package ipfs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	httppkg "github.com/trigg3rX/triggerx-backend/pkg/http"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// Test helper functions
func createTestConfig() *Config {
	return &Config{
		PinataHost:    "gateway.pinata.cloud",
		PinataJWT:     "test-jwt-token",
		PinataBaseURL: "https://uploads.pinata.cloud/v3/files",
	}
}

// Unit Tests for NewConfig
func TestNewConfig_ValidInputs_ReturnsConfig(t *testing.T) {
	tests := []struct {
		name        string
		pinataHost  string
		pinataJWT   string
		expectedURL string
	}{
		{
			name:        "standard inputs",
			pinataHost:  "gateway.pinata.cloud",
			pinataJWT:   "test-jwt-token",
			expectedURL: "https://uploads.pinata.cloud/v3/files",
		},
		{
			name:        "custom host",
			pinataHost:  "custom.pinata.cloud",
			pinataJWT:   "custom-jwt-token",
			expectedURL: "https://uploads.pinata.cloud/v3/files",
		},
		{
			name:        "empty strings (validation will be done later)",
			pinataHost:  "",
			pinataJWT:   "",
			expectedURL: "https://uploads.pinata.cloud/v3/files",
		},
		{
			name:        "special characters in JWT",
			pinataHost:  "gateway.pinata.cloud",
			pinataJWT:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			expectedURL: "https://uploads.pinata.cloud/v3/files",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewConfig(tt.pinataHost, tt.pinataJWT)

			assert.NotNil(t, config)
			assert.Equal(t, tt.pinataHost, config.PinataHost)
			assert.Equal(t, tt.pinataJWT, config.PinataJWT)
			assert.Equal(t, tt.expectedURL, config.PinataBaseURL)
		})
	}
}

func TestNewConfig_AlwaysSetsCorrectBaseURL(t *testing.T) {
	// Test that the base URL is always set correctly regardless of inputs
	config1 := NewConfig("host1", "jwt1")
	config2 := NewConfig("host2", "jwt2")
	config3 := NewConfig("", "")

	expectedBaseURL := "https://uploads.pinata.cloud/v3/files"

	assert.Equal(t, expectedBaseURL, config1.PinataBaseURL)
	assert.Equal(t, expectedBaseURL, config2.PinataBaseURL)
	assert.Equal(t, expectedBaseURL, config3.PinataBaseURL)
}

// Unit Tests for Config.Validate
func TestConfig_Validate_ValidConfig_ReturnsNoError(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{
			name: "valid config with standard values",
			config: &Config{
				PinataHost:    "gateway.pinata.cloud",
				PinataJWT:     "test-jwt-token",
				PinataBaseURL: "https://uploads.pinata.cloud/v3/files",
			},
		},
		{
			name: "valid config with custom values",
			config: &Config{
				PinataHost:    "custom.pinata.cloud",
				PinataJWT:     "custom-jwt-token",
				PinataBaseURL: "https://custom.pinata.cloud/v3/files",
			},
		},
		{
			name: "valid config with long JWT",
			config: &Config{
				PinataHost:    "gateway.pinata.cloud",
				PinataJWT:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
				PinataBaseURL: "https://uploads.pinata.cloud/v3/files",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			assert.NoError(t, err)
		})
	}
}

func TestConfig_Validate_InvalidConfig_ReturnsError(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectedErr string
	}{
		{
			name: "empty PinataHost",
			config: &Config{
				PinataHost:    "",
				PinataJWT:     "test-jwt-token",
				PinataBaseURL: "https://uploads.pinata.cloud/v3/files",
			},
			expectedErr: "PinataHost is required",
		},
		{
			name: "empty PinataJWT",
			config: &Config{
				PinataHost:    "gateway.pinata.cloud",
				PinataJWT:     "",
				PinataBaseURL: "https://uploads.pinata.cloud/v3/files",
			},
			expectedErr: "PinataJWT is required",
		},
		{
			name: "both empty",
			config: &Config{
				PinataHost:    "",
				PinataJWT:     "",
				PinataBaseURL: "https://uploads.pinata.cloud/v3/files",
			},
			expectedErr: "PinataHost is required",
		},
		{
			name: "whitespace only PinataHost",
			config: &Config{
				PinataHost:    "   ",
				PinataJWT:     "test-jwt-token",
				PinataBaseURL: "https://uploads.pinata.cloud/v3/files",
			},
			expectedErr: "PinataHost is required",
		},
		{
			name: "whitespace only PinataJWT",
			config: &Config{
				PinataHost:    "gateway.pinata.cloud",
				PinataJWT:     "   ",
				PinataBaseURL: "https://uploads.pinata.cloud/v3/files",
			},
			expectedErr: "PinataJWT is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestConfig_Validate_EdgeCases(t *testing.T) {
	t.Run("nil config should panic", func(t *testing.T) {
		// This test documents the expected behavior if someone calls Validate on a nil config
		// In practice, this should be avoided by proper initialization
		assert.Panics(t, func() {
			var config *Config
			_ = config.Validate()
		})
	})

	t.Run("config with only PinataBaseURL set", func(t *testing.T) {
		config := &Config{
			PinataBaseURL: "https://uploads.pinata.cloud/v3/files",
		}

		err := config.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "PinataHost is required")
	})

	t.Run("config with tabs and newlines in PinataHost", func(t *testing.T) {
		config := &Config{
			PinataHost:    "\t\n\r   \t\n",
			PinataJWT:     "test-jwt-token",
			PinataBaseURL: "https://uploads.pinata.cloud/v3/files",
		}

		err := config.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "PinataHost is required")
	})

	t.Run("config with tabs and newlines in PinataJWT", func(t *testing.T) {
		config := &Config{
			PinataHost:    "gateway.pinata.cloud",
			PinataJWT:     "\t\n\r   \t\n",
			PinataBaseURL: "https://uploads.pinata.cloud/v3/files",
		}

		err := config.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "PinataJWT is required")
	})
}

func createTestClient() (*ipfsClient, *httppkg.MockHTTPClient, *logging.MockLogger) {
	config := createTestConfig()
	mockHTTPClient := &httppkg.MockHTTPClient{}
	mockLogger := &logging.MockLogger{}

	return &ipfsClient{
		config:     config,
		logger:     mockLogger,
		httpClient: mockHTTPClient,
	}, mockHTTPClient, mockLogger
}

// Unit Tests for NewClient
func TestNewClient_ValidConfig_ReturnsClient(t *testing.T) {
	config := createTestConfig()
	mockLogger := &logging.MockLogger{}

	client, err := NewClient(config, mockLogger)

	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Implements(t, (*IPFSClient)(nil), client)
}

func TestNewClient_InvalidConfig_ReturnsError(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectedErr string
	}{
		{
			name:        "empty host",
			config:      &Config{PinataHost: "", PinataJWT: "token"},
			expectedErr: "PinataHost is required",
		},
		{
			name:        "empty jwt",
			config:      &Config{PinataHost: "host", PinataJWT: ""},
			expectedErr: "PinataJWT is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger := &logging.MockLogger{}

			client, err := NewClient(tt.config, mockLogger)

			assert.Error(t, err)
			assert.Nil(t, client)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

// Unit Tests for Upload method
func TestUpload_ValidData_ReturnsCID(t *testing.T) {
	client, mockHTTP, mockLogger := createTestClient()

	// Mock successful response
	responseBody := `{"data": {"cid": "QmTestCID123"}}`
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(responseBody)),
	}

	mockHTTP.On("DoWithRetry", mock.Anything, mock.AnythingOfType("*http.Request")).Return(mockResponse, nil)
	mockLogger.On("Info", "Successfully uploaded to IPFS", mock.Anything, mock.Anything).Return()

	filename := "test.txt"
	data := []byte("test data")

	cid, err := client.Upload(context.Background(), filename, data)

	assert.NoError(t, err)
	assert.Equal(t, "QmTestCID123", cid)
	mockHTTP.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestUpload_EmptyFilename_ReturnsError(t *testing.T) {
	client, _, _ := createTestClient()

	cid, err := client.Upload(context.Background(), "", []byte("test data"))

	assert.Error(t, err)
	assert.Empty(t, cid)
	assert.Contains(t, err.Error(), "filename cannot be empty")
}

func TestUpload_EmptyData_ReturnsError(t *testing.T) {
	client, _, _ := createTestClient()

	cid, err := client.Upload(context.Background(), "test.txt", []byte{})

	assert.Error(t, err)
	assert.Empty(t, cid)
	assert.Contains(t, err.Error(), "data cannot be empty")
}

func TestUpload_HTTPError_ReturnsError(t *testing.T) {
	client, mockHTTP, _ := createTestClient()

	mockHTTP.On("DoWithRetry", mock.Anything, mock.AnythingOfType("*http.Request")).Return(nil, fmt.Errorf("network error"))

	cid, err := client.Upload(context.Background(), "test.txt", []byte("test data"))

	assert.Error(t, err)
	assert.Empty(t, cid)
	assert.Contains(t, err.Error(), "failed to send request")
	mockHTTP.AssertExpectations(t)
}

func TestUpload_NonOKStatus_ReturnsError(t *testing.T) {
	client, mockHTTP, _ := createTestClient()

	mockResponse := &http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       io.NopCloser(strings.NewReader("error")),
	}

	mockHTTP.On("DoWithRetry", mock.Anything, mock.AnythingOfType("*http.Request")).Return(mockResponse, nil)

	cid, err := client.Upload(context.Background(), "test.txt", []byte("test data"))

	assert.Error(t, err)
	assert.Empty(t, cid)
	assert.Contains(t, err.Error(), "http error: status code 400")
	mockHTTP.AssertExpectations(t)
}

func TestUpload_InvalidJSONResponse_ReturnsError(t *testing.T) {
	client, mockHTTP, _ := createTestClient()

	responseBody := `{"invalid": "json"`
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(responseBody)),
	}

	mockHTTP.On("DoWithRetry", mock.Anything, mock.AnythingOfType("*http.Request")).Return(mockResponse, nil)

	cid, err := client.Upload(context.Background(), "test.txt", []byte("test data"))

	assert.Error(t, err)
	assert.Empty(t, cid)
	assert.Contains(t, err.Error(), "failed to unmarshal IPFS response")
	mockHTTP.AssertExpectations(t)
}

func TestUpload_EmptyCID_ReturnsError(t *testing.T) {
	client, mockHTTP, _ := createTestClient()

	responseBody := `{"data": {"cid": ""}}`
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(responseBody)),
	}

	mockHTTP.On("DoWithRetry", mock.Anything, mock.AnythingOfType("*http.Request")).Return(mockResponse, nil)

	cid, err := client.Upload(context.Background(), "test.txt", []byte("test data"))

	assert.Error(t, err)
	assert.Empty(t, cid)
	assert.Contains(t, err.Error(), "received empty CID from IPFS")
	mockHTTP.AssertExpectations(t)
}

// Unit Tests for Fetch method
func TestFetch_ValidCID_ReturnsIPFSData(t *testing.T) {
	client, mockHTTP, _ := createTestClient()

	// Mock IPFS data response
	ipfsData := types.IPFSData{
		TaskData: &types.SendTaskDataToKeeper{
			TaskID: []int64{123},
		},
	}
	responseBody, _ := json.Marshal(ipfsData)

	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(responseBody)),
	}

	mockHTTP.On("Get", mock.Anything, "https://gateway.pinata.cloud/ipfs/QmTestCID").Return(mockResponse, nil)

	result, err := client.Fetch(context.Background(), "QmTestCID")

	assert.NoError(t, err)
	assert.Equal(t, ipfsData.TaskData.TaskID, result.TaskData.TaskID)
	mockHTTP.AssertExpectations(t)
}

func TestFetch_EmptyCID_ReturnsError(t *testing.T) {
	client, _, _ := createTestClient()

	result, err := client.Fetch(context.Background(), "")

	assert.Error(t, err)
	assert.Equal(t, types.IPFSData{}, result)
	assert.Contains(t, err.Error(), "CID cannot be empty")
}

func TestFetch_HTTPError_ReturnsError(t *testing.T) {
	client, mockHTTP, _ := createTestClient()

	mockHTTP.On("Get", mock.Anything, "https://gateway.pinata.cloud/ipfs/QmTestCID").Return(nil, fmt.Errorf("network error"))

	result, err := client.Fetch(context.Background(), "QmTestCID")

	assert.Error(t, err)
	assert.Equal(t, types.IPFSData{}, result)
	assert.Contains(t, err.Error(), "failed to fetch IPFS content")
	mockHTTP.AssertExpectations(t)
}

func TestFetch_NonOKStatus_ReturnsError(t *testing.T) {
	client, mockHTTP, _ := createTestClient()

	mockResponse := &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(strings.NewReader("not found")),
	}

	mockHTTP.On("Get", mock.Anything, "https://gateway.pinata.cloud/ipfs/QmTestCID").Return(mockResponse, nil)

	result, err := client.Fetch(context.Background(), "QmTestCID")

	assert.Error(t, err)
	assert.Equal(t, types.IPFSData{}, result)
	assert.Contains(t, err.Error(), "http error: status code 404")
	mockHTTP.AssertExpectations(t)
}

func TestFetch_InvalidJSONResponse_ReturnsError(t *testing.T) {
	client, mockHTTP, _ := createTestClient()

	responseBody := `{"invalid": "json`
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(responseBody)),
	}

	mockHTTP.On("Get", mock.Anything, "https://gateway.pinata.cloud/ipfs/QmTestCID").Return(mockResponse, nil)

	result, err := client.Fetch(context.Background(), "QmTestCID")

	assert.Error(t, err)
	assert.Equal(t, types.IPFSData{}, result)
	assert.Contains(t, err.Error(), "failed to unmarshal IPFS data")
	mockHTTP.AssertExpectations(t)
}

// Unit Tests for Close method
func TestClose_Success(t *testing.T) {
	client, mockHTTP, _ := createTestClient()

	mockHTTP.On("Close").Return()

	err := client.Close()

	assert.NoError(t, err)
	mockHTTP.AssertExpectations(t)
}

// Benchmark Tests
func BenchmarkUpload_SmallFile(b *testing.B) {
	filename := "benchmark.txt"
	data := []byte("benchmark test data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create fresh mocks for each iteration
		client, mockHTTP, mockLogger := createTestClient()

		// Create fresh mock response for each iteration
		responseBody := `{"data": {"cid": "QmBenchmarkCID"}}`
		mockResponse := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(responseBody)),
		}

		mockHTTP.On("DoWithRetry", mock.Anything, mock.AnythingOfType("*http.Request")).Return(mockResponse, nil)
		mockLogger.On("Info", "Successfully uploaded to IPFS", mock.Anything, mock.Anything).Return()

		_, err := client.Upload(context.Background(), filename, data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFetch_SmallData(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create fresh mocks for each iteration
		client, mockHTTP, _ := createTestClient()

		// Create fresh mock response for each iteration
		ipfsData := types.IPFSData{
			TaskData: &types.SendTaskDataToKeeper{
				TaskID: []int64{789},
			},
		}
		responseBody, _ := json.Marshal(ipfsData)

		mockResponse := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(responseBody)),
		}

		mockHTTP.On("Get", mock.Anything, "https://gateway.pinata.cloud/ipfs/QmBenchmarkCID").Return(mockResponse, nil)

		_, err := client.Fetch(context.Background(), "QmBenchmarkCID")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Table-driven tests for edge cases
func TestUpload_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		data        []byte
		expectError bool
		errorMsg    string
	}{
		{
			name:        "very large filename",
			filename:    strings.Repeat("a", 1000),
			data:        []byte("test"),
			expectError: false,
		},
		{
			name:        "very large data",
			filename:    "test.txt",
			data:        bytes.Repeat([]byte("a"), 1000000), // 1MB
			expectError: false,
		},
		{
			name:        "special characters in filename",
			filename:    "test@#$%^&*().txt",
			data:        []byte("test"),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh mocks for each test case
			client, mockHTTP, mockLogger := createTestClient()

			if !tt.expectError {
				responseBody := `{"data": {"cid": "QmTestCID"}}`
				mockResponse := &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(responseBody)),
				}
				mockHTTP.On("DoWithRetry", mock.Anything, mock.AnythingOfType("*http.Request")).Return(mockResponse, nil)
				mockLogger.On("Info", "Successfully uploaded to IPFS", mock.Anything, mock.Anything).Return()
			}

			cid, err := client.Upload(context.Background(), tt.filename, tt.data)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, cid)
			}

			mockHTTP.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestFetch_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		cid         string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "very long CID",
			cid:         "Qm" + strings.Repeat("a", 100),
			expectError: false,
		},
		{
			name:        "CID with special characters",
			cid:         "Qm@#$%^&*()",
			expectError: false,
		},
		{
			name:        "CID with spaces",
			cid:         "Qm Test CID",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh mocks for each test case
			client, mockHTTP, _ := createTestClient()

			if !tt.expectError {
				ipfsData := types.IPFSData{
					TaskData: &types.SendTaskDataToKeeper{
						TaskID: []int64{999},
					},
				}
				responseBody, _ := json.Marshal(ipfsData)
				mockResponse := &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(responseBody)),
				}
				mockHTTP.On("Get", mock.Anything, "https://gateway.pinata.cloud/ipfs/"+tt.cid).Return(mockResponse, nil)
			}

			result, err := client.Fetch(context.Background(), tt.cid)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockHTTP.AssertExpectations(t)
		})
	}
}
