package http

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestMockHTTPClient_DoWithRetry_Success(t *testing.T) {
	mockClient := &MockHTTPClient{}

	// Setup mock expectations
	expectedResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(`{"status": "success"}`)),
	}
	mockClient.On("DoWithRetry", mock.AnythingOfType("*http.Request")).Return(expectedResp, nil)

	// Create test request
	req, err := http.NewRequest("GET", "http://example.com", nil)
	require.NoError(t, err)

	// Execute
	resp, err := mockClient.DoWithRetry(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResp, resp)
	mockClient.AssertExpectations(t)
}

func TestMockHTTPClient_DoWithRetry_Error(t *testing.T) {
	mockClient := &MockHTTPClient{}

	// Setup mock expectations
	expectedErr := errors.New("network error")
	mockClient.On("DoWithRetry", mock.AnythingOfType("*http.Request")).Return(nil, expectedErr)

	// Create test request
	req, err := http.NewRequest("GET", "http://example.com", nil)
	require.NoError(t, err)

	// Execute
	resp, err := mockClient.DoWithRetry(req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, resp)
	mockClient.AssertExpectations(t)
}

func TestMockHTTPClient_Get_Success(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		expectedResp *http.Response
		expectedErr  error
	}{
		{
			name: "successful get request",
			url:  "http://example.com/api",
			expectedResp: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"data": "test"}`)),
			},
			expectedErr: nil,
		},
		{
			name:         "get request with error",
			url:          "http://invalid-url.com",
			expectedResp: nil,
			expectedErr:  errors.New("connection refused"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{}
			mockClient.On("Get", tt.url).Return(tt.expectedResp, tt.expectedErr)

			// Execute
			resp, err := mockClient.Get(tt.url)

			// Assert
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResp, resp)
			}
			mockClient.AssertExpectations(t)
		})
	}
}

func TestMockHTTPClient_Post_Success(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		contentType  string
		body         io.Reader
		expectedResp *http.Response
		expectedErr  error
	}{
		{
			name:        "successful post with json",
			url:         "http://example.com/api",
			contentType: "application/json",
			body:        strings.NewReader(`{"key": "value"}`),
			expectedResp: &http.Response{
				StatusCode: http.StatusCreated,
				Body:       io.NopCloser(strings.NewReader(`{"id": "123"}`)),
			},
			expectedErr: nil,
		},
		{
			name:        "post with form data",
			url:         "http://example.com/form",
			contentType: "application/x-www-form-urlencoded",
			body:        strings.NewReader("key=value&other=data"),
			expectedResp: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"status": "submitted"}`)),
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{}
			mockClient.On("Post", tt.url, tt.contentType, tt.body).Return(tt.expectedResp, tt.expectedErr)

			// Execute
			resp, err := mockClient.Post(tt.url, tt.contentType, tt.body)

			// Assert
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResp, resp)
			}
			mockClient.AssertExpectations(t)
		})
	}
}

func TestMockHTTPClient_Put_Success(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		contentType  string
		body         io.Reader
		expectedResp *http.Response
		expectedErr  error
	}{
		{
			name:        "successful put request",
			url:         "http://example.com/api/123",
			contentType: "application/json",
			body:        strings.NewReader(`{"name": "updated"}`),
			expectedResp: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"updated": true}`)),
			},
			expectedErr: nil,
		},
		{
			name:         "put request with error",
			url:          "http://example.com/api/999",
			contentType:  "application/json",
			body:         strings.NewReader(`{"name": "notfound"}`),
			expectedResp: nil,
			expectedErr:  errors.New("resource not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{}
			mockClient.On("Put", tt.url, tt.contentType, tt.body).Return(tt.expectedResp, tt.expectedErr)

			// Execute
			resp, err := mockClient.Put(tt.url, tt.contentType, tt.body)

			// Assert
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResp, resp)
			}
			mockClient.AssertExpectations(t)
		})
	}
}

func TestMockHTTPClient_Delete_Success(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		expectedResp *http.Response
		expectedErr  error
	}{
		{
			name: "successful delete request",
			url:  "http://example.com/api/123",
			expectedResp: &http.Response{
				StatusCode: http.StatusNoContent,
				Body:       io.NopCloser(strings.NewReader("")),
			},
			expectedErr: nil,
		},
		{
			name:         "delete request with error",
			url:          "http://example.com/api/999",
			expectedResp: nil,
			expectedErr:  errors.New("resource not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{}
			mockClient.On("Delete", tt.url).Return(tt.expectedResp, tt.expectedErr)

			// Execute
			resp, err := mockClient.Delete(tt.url)

			// Assert
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResp, resp)
			}
			mockClient.AssertExpectations(t)
		})
	}
}

func TestMockHTTPClient_Close(t *testing.T) {
	mockClient := &MockHTTPClient{}
	mockClient.On("Close").Return()

	// Execute
	mockClient.Close()

	// Assert
	mockClient.AssertExpectations(t)
}

func TestMockHTTPClient_GetTimeout(t *testing.T) {
	tests := []struct {
		name            string
		expectedTimeout time.Duration
	}{
		{
			name:            "default timeout",
			expectedTimeout: 10 * time.Second,
		},
		{
			name:            "custom timeout",
			expectedTimeout: 30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{}
			mockClient.On("GetTimeout").Return(tt.expectedTimeout)

			// Execute
			timeout := mockClient.GetTimeout()

			// Assert
			assert.Equal(t, tt.expectedTimeout, timeout)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestMockHTTPClient_GetIdleConnTimeout(t *testing.T) {
	tests := []struct {
		name            string
		expectedTimeout time.Duration
	}{
		{
			name:            "default idle timeout",
			expectedTimeout: 30 * time.Second,
		},
		{
			name:            "custom idle timeout",
			expectedTimeout: 60 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{}
			mockClient.On("GetIdleConnTimeout").Return(tt.expectedTimeout)

			// Execute
			timeout := mockClient.GetIdleConnTimeout()

			// Assert
			assert.Equal(t, tt.expectedTimeout, timeout)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestMockResponseBuilder_NewMockResponseBuilder(t *testing.T) {
	builder := NewMockResponseBuilder()

	// Assert default values
	assert.Equal(t, http.StatusOK, builder.statusCode)
	assert.Equal(t, "", builder.body)
	assert.NotNil(t, builder.headers)
	assert.Empty(t, builder.headers)
}

func TestMockResponseBuilder_WithStatusCode(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"ok status", http.StatusOK},
		{"created status", http.StatusCreated},
		{"not found status", http.StatusNotFound},
		{"server error status", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewMockResponseBuilder()

			// Execute
			result := builder.WithStatusCode(tt.statusCode)

			// Assert
			assert.Equal(t, builder, result) // Should return self for chaining
			assert.Equal(t, tt.statusCode, builder.statusCode)
		})
	}
}

func TestMockResponseBuilder_WithBody(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"empty body", ""},
		{"json body", `{"key": "value"}`},
		{"html body", "<html><body>Hello</body></html>"},
		{"plain text", "Hello, World!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewMockResponseBuilder()

			// Execute
			result := builder.WithBody(tt.body)

			// Assert
			assert.Equal(t, builder, result) // Should return self for chaining
			assert.Equal(t, tt.body, builder.body)
		})
	}
}

func TestMockResponseBuilder_WithHeader(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{"content type header", "Content-Type", "application/json"},
		{"authorization header", "Authorization", "Bearer token123"},
		{"custom header", "X-Custom-Header", "custom-value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewMockResponseBuilder()

			// Execute
			result := builder.WithHeader(tt.key, tt.value)

			// Assert
			assert.Equal(t, builder, result) // Should return self for chaining
			assert.Equal(t, tt.value, builder.headers[tt.key])
		})
	}
}

func TestMockResponseBuilder_Build_DefaultResponse(t *testing.T) {
	builder := NewMockResponseBuilder()

	// Execute
	resp := builder.Build()

	// Assert
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotNil(t, resp.Body)
	assert.NotNil(t, resp.Header)

	// Read body
	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "{}", string(bodyBytes))
}

func TestMockResponseBuilder_Build_CustomResponse(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		headers    map[string]string
	}{
		{
			name:       "json response",
			statusCode: http.StatusCreated,
			body:       `{"id": "123", "name": "test"}`,
			headers: map[string]string{
				"Content-Type": "application/json",
				"Location":     "/api/123",
			},
		},
		{
			name:       "html response",
			statusCode: http.StatusOK,
			body:       "<html><body>Success</body></html>",
			headers: map[string]string{
				"Content-Type": "text/html",
			},
		},
		{
			name:       "error response",
			statusCode: http.StatusNotFound,
			body:       `{"error": "Resource not found"}`,
			headers: map[string]string{
				"Content-Type": "application/json",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewMockResponseBuilder().
				WithStatusCode(tt.statusCode).
				WithBody(tt.body)

			// Add headers
			for key, value := range tt.headers {
				builder.WithHeader(key, value)
			}

			// Execute
			resp := builder.Build()

			// Assert
			assert.Equal(t, tt.statusCode, resp.StatusCode)
			assert.NotNil(t, resp.Body)
			assert.NotNil(t, resp.Header)

			// Read body
			bodyBytes, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.Equal(t, tt.body, string(bodyBytes))

			// Check headers
			for key, value := range tt.headers {
				assert.Equal(t, value, resp.Header.Get(key))
			}
		})
	}
}

func TestMockResponseBuilder_Build_EmptyBodyUsesDefault(t *testing.T) {
	builder := NewMockResponseBuilder().
		WithStatusCode(http.StatusNoContent).
		WithBody("") // Empty body

	// Execute
	resp := builder.Build()

	// Assert
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Read body - should use default "{}"
	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "{}", string(bodyBytes))
}

func TestMockResponseBuilder_Chaining(t *testing.T) {
	// Test method chaining
	resp := NewMockResponseBuilder().
		WithStatusCode(http.StatusCreated).
		WithBody(`{"message": "created"}`).
		WithHeader("Content-Type", "application/json").
		WithHeader("X-Custom", "value").
		Build()

	// Assert
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	assert.Equal(t, "value", resp.Header.Get("X-Custom"))

	// Read body
	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, `{"message": "created"}`, string(bodyBytes))
}

func TestMockHTTPClient_Integration_WithMockResponseBuilder(t *testing.T) {
	mockClient := &MockHTTPClient{}

	// Create mock response using builder
	mockResp := NewMockResponseBuilder().
		WithStatusCode(http.StatusOK).
		WithBody(`{"status": "success", "data": "test"}`).
		WithHeader("Content-Type", "application/json").
		WithHeader("X-Request-ID", "req-123").
		Build()

	// Setup mock expectations
	mockClient.On("Get", "http://example.com/api").Return(mockResp, nil)

	// Execute
	resp, err := mockClient.Get("http://example.com/api")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	assert.Equal(t, "req-123", resp.Header.Get("X-Request-ID"))

	// Read body
	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, `{"status": "success", "data": "test"}`, string(bodyBytes))

	mockClient.AssertExpectations(t)
}
