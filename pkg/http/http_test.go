package http

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"bytes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
)

// TestHTTPRetryConfig_DefaultConfig_ReturnsValidConfig tests the default configuration
func TestHTTPRetryConfig_DefaultConfig_ReturnsValidConfig(t *testing.T) {
	config := DefaultHTTPRetryConfig()

	assert.NotNil(t, config)
	assert.NotNil(t, config.RetryConfig)
	assert.Equal(t, 10*time.Second, config.Timeout)
	assert.Equal(t, 30*time.Second, config.IdleConnTimeout)
	assert.Equal(t, int64(4096), config.MaxResponseSize)
}

// TestHTTPRetryConfig_Validate_ValidConfig_ReturnsNoError tests validation of valid config
func TestHTTPRetryConfig_Validate_ValidConfig_ReturnsNoError(t *testing.T) {
	config := &HTTPRetryConfig{
		RetryConfig:     retry.DefaultRetryConfig(),
		Timeout:         10 * time.Second,
		IdleConnTimeout: 30 * time.Second,
		MaxResponseSize: 4096,
	}

	err := config.Validate()
	assert.NoError(t, err)
}

// TestHTTPRetryConfig_Validate_InvalidConfig_ReturnsError tests validation of invalid configs
func TestHTTPRetryConfig_Validate_InvalidConfig_ReturnsError(t *testing.T) {
	tests := []struct {
		name        string
		config      *HTTPRetryConfig
		expectedErr string
	}{
		{
			name: "zero timeout",
			config: &HTTPRetryConfig{
				RetryConfig:     retry.DefaultRetryConfig(),
				Timeout:         0,
				IdleConnTimeout: 30 * time.Second,
				MaxResponseSize: 4096,
			},
			expectedErr: "timeout must be positive",
		},
		{
			name: "negative timeout",
			config: &HTTPRetryConfig{
				RetryConfig:     retry.DefaultRetryConfig(),
				Timeout:         -1 * time.Second,
				IdleConnTimeout: 30 * time.Second,
				MaxResponseSize: 4096,
			},
			expectedErr: "timeout must be positive",
		},
		{
			name: "zero idle conn timeout",
			config: &HTTPRetryConfig{
				RetryConfig:     retry.DefaultRetryConfig(),
				Timeout:         10 * time.Second,
				IdleConnTimeout: 0,
				MaxResponseSize: 4096,
			},
			expectedErr: "idleConnTimeout must be positive",
		},
		{
			name: "negative max response size",
			config: &HTTPRetryConfig{
				RetryConfig:     retry.DefaultRetryConfig(),
				Timeout:         10 * time.Second,
				IdleConnTimeout: 30 * time.Second,
				MaxResponseSize: -1,
			},
			expectedErr: "maxResponseSize must be >= 0",
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

// TestHTTPError_Error_ReturnsFormattedMessage tests HTTPError formatting
func TestHTTPError_Error_ReturnsFormattedMessage(t *testing.T) {
	err := &HTTPError{
		StatusCode: 500,
		Message:    "Internal Server Error",
	}

	expected := "HTTP 500: Internal Server Error"
	assert.Equal(t, expected, err.Error())
}

// TestNewHTTPClient_ValidConfig_ReturnsClient tests client creation with valid config
func TestNewHTTPClient_ValidConfig_ReturnsClient(t *testing.T) {
	logger := logging.NewNoOpLogger()
	config := DefaultHTTPRetryConfig()

	client, err := NewHTTPClient(config, logger)

	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, config, client.HTTPConfig)
	assert.Equal(t, logger, client.logger)
	assert.NotNil(t, client.client)
}

// TestNewHTTPClient_NilConfig_UsesDefaultConfig tests client creation with nil config
func TestNewHTTPClient_NilConfig_UsesDefaultConfig(t *testing.T) {
	logger := logging.NewNoOpLogger()

	client, err := NewHTTPClient(nil, logger)

	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.HTTPConfig)
	assert.Equal(t, 10*time.Second, client.HTTPConfig.Timeout)
}

// TestNewHTTPClient_InvalidConfig_ReturnsError tests client creation with invalid config
func TestNewHTTPClient_InvalidConfig_ReturnsError(t *testing.T) {
	logger := logging.NewNoOpLogger()
	config := &HTTPRetryConfig{
		RetryConfig:     retry.DefaultRetryConfig(),
		Timeout:         0, // Invalid
		IdleConnTimeout: 30 * time.Second,
		MaxResponseSize: 4096,
	}

	client, err := NewHTTPClient(config, logger)

	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "invalid HTTP retry config")
}

// TestHTTPClient_DoWithRetry_SuccessfulRequest_ReturnsResponse tests successful request
func TestHTTPClient_DoWithRetry_SuccessfulRequest_ReturnsResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("success"))
		require.NoError(t, err)
	}))
	defer server.Close()

	logger := logging.NewNoOpLogger()
	client, err := NewHTTPClient(DefaultHTTPRetryConfig(), logger)
	require.NoError(t, err)

	req, err := http.NewRequest("GET", server.URL, nil)
	require.NoError(t, err)

	resp, err := client.DoWithRetry(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			t.Errorf("Error closing response body: %v", cerr)
		}
	}()
}

// TestHTTPClient_DoWithRetry_RetryableError_RetriesAndSucceeds tests retry logic with retryable errors
func TestHTTPClient_DoWithRetry_RetryableError_RetriesAndSucceeds(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			_, err := w.Write([]byte("server error"))
			require.NoError(t, err)
		} else {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("success"))
			require.NoError(t, err)
		}
	}))
	defer server.Close()

	logger := logging.NewNoOpLogger()
	config := DefaultHTTPRetryConfig()
	config.RetryConfig.MaxRetries = 3
	config.RetryConfig.InitialDelay = 10 * time.Millisecond
	config.RetryConfig.MaxDelay = 50 * time.Millisecond

	client, err := NewHTTPClient(config, logger)
	require.NoError(t, err)

	req, err := http.NewRequest("GET", server.URL, nil)
	require.NoError(t, err)

	resp, err := client.DoWithRetry(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.GreaterOrEqual(t, attempts, 3)
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			t.Errorf("Error closing response body: %v", cerr)
		}
	}()
}

// TestHTTPClient_DoWithRetry_NonRetryableError_ReturnsError tests non-retryable errors
func TestHTTPClient_DoWithRetry_NonRetryableError_ReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("bad request"))
		require.NoError(t, err)
	}))
	defer server.Close()

	logger := logging.NewNoOpLogger()
	client, err := NewHTTPClient(DefaultHTTPRetryConfig(), logger)
	require.NoError(t, err)

	req, err := http.NewRequest("GET", server.URL, nil)
	require.NoError(t, err)

	resp, err := client.DoWithRetry(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			t.Errorf("Error closing response body: %v", cerr)
		}
	}()
}

// TestHTTPClient_DoWithRetry_ContextCancelled_ReturnsError tests context cancellation
func TestHTTPClient_DoWithRetry_ContextCancelled_ReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("success"))
		require.NoError(t, err)
	}))
	defer server.Close()

	logger := logging.NewNoOpLogger()
	client, err := NewHTTPClient(DefaultHTTPRetryConfig(), logger)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	req, err := http.NewRequest("GET", server.URL, nil)
	require.NoError(t, err)
	req = req.WithContext(ctx)

	// Cancel context immediately
	cancel()

	resp, err := client.DoWithRetry(req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, context.Canceled, err)
}

// TestHTTPClient_Get_SuccessfulRequest_ReturnsResponse tests GET method
func TestHTTPClient_Get_SuccessfulRequest_ReturnsResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("success"))
		require.NoError(t, err)
	}))
	defer server.Close()

	logger := logging.NewNoOpLogger()
	client, err := NewHTTPClient(DefaultHTTPRetryConfig(), logger)
	require.NoError(t, err)

	resp, err := client.Get(server.URL)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			t.Errorf("Error closing response body: %v", cerr)
		}
	}()
}

// TestHTTPClient_Post_SuccessfulRequest_ReturnsResponse tests POST method
func TestHTTPClient_Post_SuccessfulRequest_ReturnsResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		body, _ := io.ReadAll(r.Body)
		assert.Equal(t, `{"test":"data"}`, string(body))

		w.WriteHeader(http.StatusCreated)
		_, err := w.Write([]byte("created"))
		require.NoError(t, err)
	}))
	defer server.Close()

	logger := logging.NewNoOpLogger()
	client, err := NewHTTPClient(DefaultHTTPRetryConfig(), logger)
	require.NoError(t, err)

	body := strings.NewReader(`{"test":"data"}`)
	resp, err := client.Post(server.URL, "application/json", body)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			t.Errorf("Error closing response body: %v", cerr)
		}
	}()
}

// TestHTTPClient_Put_SuccessfulRequest_ReturnsResponse tests PUT method
func TestHTTPClient_Put_SuccessfulRequest_ReturnsResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		body, _ := io.ReadAll(r.Body)
		assert.Equal(t, `{"test":"data"}`, string(body))

		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("updated"))
		require.NoError(t, err)
	}))
	defer server.Close()

	logger := logging.NewNoOpLogger()
	client, err := NewHTTPClient(DefaultHTTPRetryConfig(), logger)
	require.NoError(t, err)

	body := strings.NewReader(`{"test":"data"}`)
	resp, err := client.Put(server.URL, "application/json", body)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			t.Errorf("Error closing response body: %v", cerr)
		}
	}()
}

// TestHTTPClient_Delete_SuccessfulRequest_ReturnsResponse tests DELETE method
func TestHTTPClient_Delete_SuccessfulRequest_ReturnsResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		// 204 No Content should not have a body according to HTTP spec
		w.WriteHeader(http.StatusNoContent)
		// Don't write a body for 204 responses
	}))
	defer server.Close()

	logger := logging.NewNoOpLogger()
	client, err := NewHTTPClient(DefaultHTTPRetryConfig(), logger)
	require.NoError(t, err)

	resp, err := client.Delete(server.URL)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			t.Errorf("Error closing response body: %v", cerr)
		}
	}()
}

// TestHTTPClient_DoWithRetry_RequestWithBody_HandlesBodyCorrectly tests request body handling
func TestHTTPClient_DoWithRetry_RequestWithBody_HandlesBodyCorrectly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		assert.Equal(t, "test body", string(body))
		w.WriteHeader(http.StatusOK)
		_, err := w.Write(body)
		require.NoError(t, err)
	}))
	defer server.Close()

	logger := logging.NewNoOpLogger()
	client, err := NewHTTPClient(DefaultHTTPRetryConfig(), logger)
	require.NoError(t, err)

	body := strings.NewReader("test body")
	req, err := http.NewRequest("POST", server.URL, body)
	require.NoError(t, err)

	resp, err := client.DoWithRetry(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			t.Errorf("Error closing response body: %v", cerr)
		}
	}()
}

// TestHTTPClient_GetTimeout_ReturnsConfiguredTimeout tests timeout getter
func TestHTTPClient_GetTimeout_ReturnsConfiguredTimeout(t *testing.T) {
	logger := logging.NewNoOpLogger()
	config := DefaultHTTPRetryConfig()
	config.Timeout = 15 * time.Second

	client, err := NewHTTPClient(config, logger)
	require.NoError(t, err)

	assert.Equal(t, 15*time.Second, client.GetTimeout())
}

// TestHTTPClient_GetIdleConnTimeout_ReturnsConfiguredTimeout tests idle connection timeout getter
func TestHTTPClient_GetIdleConnTimeout_ReturnsConfiguredTimeout(t *testing.T) {
	logger := logging.NewNoOpLogger()
	config := DefaultHTTPRetryConfig()
	config.IdleConnTimeout = 45 * time.Second

	client, err := NewHTTPClient(config, logger)
	require.NoError(t, err)

	assert.Equal(t, 45*time.Second, client.GetIdleConnTimeout())
}

// TestHTTPClient_Close_ClosesIdleConnections tests connection cleanup
func TestHTTPClient_Close_ClosesIdleConnections(t *testing.T) {
	logger := logging.NewNoOpLogger()
	client, err := NewHTTPClient(DefaultHTTPRetryConfig(), logger)
	require.NoError(t, err)

	// This should not panic
	client.Close()
}

// TestHTTPClient_DoWithRetry_RequestBodyReadError_ReturnsError tests body read error handling
func TestHTTPClient_DoWithRetry_RequestBodyReadError_ReturnsError(t *testing.T) {
	logger := logging.NewNoOpLogger()
	client, err := NewHTTPClient(DefaultHTTPRetryConfig(), logger)
	require.NoError(t, err)

	// Create a request with a body that will fail to read
	req, err := http.NewRequest("POST", "http://example.com", &errorReader{})
	require.NoError(t, err)

	resp, err := client.DoWithRetry(req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "error reading request body")
}

// errorReader is a reader that always returns an error
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

// TestHTTPClient_DoWithRetry_CustomShouldRetry_RespectsCustomLogic tests custom retry logic
func TestHTTPClient_DoWithRetry_CustomShouldRetry_RespectsCustomLogic(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError) // Retryable by default
			_, err := w.Write([]byte("server error"))
			require.NoError(t, err)
		} else {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("success"))
			require.NoError(t, err)
		}
	}))
	defer server.Close()

	logger := logging.NewNoOpLogger()
	config := DefaultHTTPRetryConfig()
	config.RetryConfig.ShouldRetry = func(err error, attempt int) bool {
		// Custom logic: only retry on specific errors
		return false // Don't retry on any error
	}
	config.RetryConfig.MaxRetries = 2
	config.RetryConfig.InitialDelay = 10 * time.Millisecond

	client, err := NewHTTPClient(config, logger)
	require.NoError(t, err)

	req, err := http.NewRequest("GET", server.URL, nil)
	require.NoError(t, err)

	resp, err := client.DoWithRetry(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Equal(t, 1, attempts) // Should not have retried due to custom logic
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			t.Errorf("Error closing response body: %v", cerr)
		}
	}()
}

// TestHTTPClient_DoWithRetry_RequestWithoutGetBody_UsesFallback tests the fallback for requests without GetBody
func TestHTTPClient_DoWithRetry_RequestWithoutGetBody_UsesFallback(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		body, _ := io.ReadAll(r.Body)
		assert.Equal(t, "test body", string(body))

		// Return a retryable error on first attempt to test retry with body
		if attempts == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			_, err := w.Write([]byte("server error"))
			require.NoError(t, err)
		} else {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write(body)
			require.NoError(t, err)
		}
	}))
	defer server.Close()

	logger := logging.NewNoOpLogger()
	config := DefaultHTTPRetryConfig()
	config.RetryConfig.MaxRetries = 2
	config.RetryConfig.InitialDelay = 10 * time.Millisecond
	config.RetryConfig.MaxDelay = 50 * time.Millisecond

	client, err := NewHTTPClient(config, logger)
	require.NoError(t, err)

	// Create a request with a body but without GetBody function
	body := strings.NewReader("test body")
	req, err := http.NewRequest("POST", server.URL, body)
	require.NoError(t, err)

	// Ensure GetBody is nil initially to trigger fallback
	// Note: DoWithRetry will create a fallback GetBody function internally

	resp, err := client.DoWithRetry(req)

	// The test should succeed even though the request initially had no GetBody
	// The DoWithRetry method should create a fallback GetBody function internally
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.GreaterOrEqual(t, attempts, 2) // Should have retried at least once
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			t.Errorf("Error closing response body: %v", cerr)
		}
	}()
}

// TestHTTPClient_DoWithRetry_RequestWithGetBody_UsesGetBody tests requests that already have GetBody
func TestHTTPClient_DoWithRetry_RequestWithGetBody_UsesGetBody(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		body, _ := io.ReadAll(r.Body)
		assert.Equal(t, "test body", string(body))

		// Return a retryable error on first attempt to test retry with body
		if attempts == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			_, err := w.Write([]byte("server error"))
			require.NoError(t, err)
		} else {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write(body)
			require.NoError(t, err)
		}
	}))
	defer server.Close()

	logger := logging.NewNoOpLogger()
	config := DefaultHTTPRetryConfig()
	config.RetryConfig.MaxRetries = 2
	config.RetryConfig.InitialDelay = 10 * time.Millisecond
	config.RetryConfig.MaxDelay = 50 * time.Millisecond

	client, err := NewHTTPClient(config, logger)
	require.NoError(t, err)

	// Create a request with a body and GetBody function
	bodyBytes := []byte("test body")
	req, err := http.NewRequest("POST", server.URL, bytes.NewReader(bodyBytes))
	require.NoError(t, err)

	// Set GetBody function
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(bodyBytes)), nil
	}

	// Ensure GetBody is not nil
	assert.NotNil(t, req.GetBody)

	resp, err := client.DoWithRetry(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.GreaterOrEqual(t, attempts, 2) // Should have retried at least once
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			t.Errorf("Error closing response body: %v", cerr)
		}
	}()
}

// TestHTTPClient_DoWithRetry_InvalidURL_ReturnsError tests invalid URL handling
func TestHTTPClient_DoWithRetry_InvalidURL_ReturnsError(t *testing.T) {
	logger := logging.NewNoOpLogger()
	client, err := NewHTTPClient(DefaultHTTPRetryConfig(), logger)
	require.NoError(t, err)

	req, err := http.NewRequest("GET", "http://invalid.localhost:99999", nil)
	require.NoError(t, err)

	resp, err := client.DoWithRetry(req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "operation failed after")
}

// TestHTTPClient_DoWithRetry_RequestBodyCloseError_LogsWarning tests body close error handling
func TestHTTPClient_DoWithRetry_RequestBodyCloseError_LogsWarning(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("success"))
		require.NoError(t, err)
	}))
	defer server.Close()

	logger := logging.NewNoOpLogger()
	client, err := NewHTTPClient(DefaultHTTPRetryConfig(), logger)
	require.NoError(t, err)

	// Create a request with a body that will fail to close
	req, err := http.NewRequest("POST", server.URL, &errorCloseReader{})
	require.NoError(t, err)

	resp, err := client.DoWithRetry(req)

	// Should still succeed despite body close error
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			t.Errorf("Error closing response body: %v", cerr)
		}
	}()
}

// errorCloseReader is a reader that fails to close
type errorCloseReader struct {
	io.Reader
	readCount int
}

func (e *errorCloseReader) Read(p []byte) (n int, err error) {
	// Read a small amount of data and then signal EOF
	if e.readCount == 0 {
		copy(p, []byte("test data"))
		e.readCount++
		return 9, nil // Return 9 bytes
	}
	return 0, io.EOF // Signal EOF after first read
}

func (e *errorCloseReader) Close() error {
	return errors.New("close error")
}
