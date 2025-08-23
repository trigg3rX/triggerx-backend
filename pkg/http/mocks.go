package http

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/stretchr/testify/mock"
)

// MockHTTPClient is a mock implementation of the HTTPClient interface
type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) OnGet(url string) *mock.Call {
    return m.On("Get", url)
}

// DoWithRetry mocks the DoWithRetry method
func (m *MockHTTPClient) DoWithRetry(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}

// Get mocks the Get method
func (m *MockHTTPClient) Get(url string) (*http.Response, error) {
	args := m.Called(url)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}

// Post mocks the Post method
func (m *MockHTTPClient) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	args := m.Called(url, contentType, body)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}

// Put mocks the Put method
func (m *MockHTTPClient) Put(url, contentType string, body io.Reader) (*http.Response, error) {
	args := m.Called(url, contentType, body)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}

// Delete mocks the Delete method
func (m *MockHTTPClient) Delete(url string) (*http.Response, error) {
	args := m.Called(url)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}

// Close mocks the Close method
func (m *MockHTTPClient) Close() {
	m.Called()
}

// GetTimeout mocks the GetTimeout method
func (m *MockHTTPClient) GetTimeout() time.Duration {
	args := m.Called()
	return args.Get(0).(time.Duration)
}

// GetIdleConnTimeout mocks the GetIdleConnTimeout method
func (m *MockHTTPClient) GetIdleConnTimeout() time.Duration {
	args := m.Called()
	return args.Get(0).(time.Duration)
}

// GetClient mocks the GetClient method
func (m *MockHTTPClient) GetClient() *http.Client {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*http.Client)
}

// Test helper functions for creating mock responses
type MockResponseBuilder struct {
	statusCode int
	body       string
	headers    map[string]string
}

// NewMockResponseBuilder creates a new mock response builder
func NewMockResponseBuilder() *MockResponseBuilder {
	return &MockResponseBuilder{
		statusCode: http.StatusOK,
		headers:    make(map[string]string),
	}
}

// WithStatusCode sets the status code for the mock response
func (b *MockResponseBuilder) WithStatusCode(statusCode int) *MockResponseBuilder {
	b.statusCode = statusCode
	return b
}

// WithBody sets the response body for the mock response
func (b *MockResponseBuilder) WithBody(body string) *MockResponseBuilder {
	b.body = body
	return b
}

// WithHeader adds a header to the mock response
func (b *MockResponseBuilder) WithHeader(key, value string) *MockResponseBuilder {
	b.headers[key] = value
	return b
}

// Build creates the final mock response
func (b *MockResponseBuilder) Build() *http.Response {
	body := b.body
	if body == "" {
		body = "{}"
	}

	headers := make(http.Header)
	for key, value := range b.headers {
		headers.Add(key, value)
	}

	return &http.Response{
		StatusCode: b.statusCode,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
		Header:     headers,
	}
}
