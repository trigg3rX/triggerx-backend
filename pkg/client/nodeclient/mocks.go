package nodeclient

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"

	"github.com/stretchr/testify/mock"
)

// MockHTTPClientInterface is a mock for httppkg.HTTPClientInterface
// This is used for testing NodeClient without making real HTTP requests
type MockHTTPClientInterface struct {
	mock.Mock
}

func (m *MockHTTPClientInterface) DoWithRetry(ctx context.Context, req *http.Request) (*http.Response, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}

func (m *MockHTTPClientInterface) Get(ctx context.Context, url string) (*http.Response, error) {
	args := m.Called(ctx, url)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}

func (m *MockHTTPClientInterface) Post(ctx context.Context, url, contentType string, body io.Reader) (*http.Response, error) {
	args := m.Called(ctx, url, contentType, body)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}

func (m *MockHTTPClientInterface) Put(ctx context.Context, url, contentType string, body io.Reader) (*http.Response, error) {
	args := m.Called(ctx, url, contentType, body)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}

func (m *MockHTTPClientInterface) Delete(ctx context.Context, url string) (*http.Response, error) {
	args := m.Called(ctx, url)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}

func (m *MockHTTPClientInterface) GetClient() *http.Client {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*http.Client)
}

func (m *MockHTTPClientInterface) Close() {
	m.Called()
}

// MockWebSocketClientInterface is a mock for wsclient.WebSocketClientInterface
type MockWebSocketClientInterface struct {
	mock.Mock
}

func (m *MockWebSocketClientInterface) Connect(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockWebSocketClientInterface) ReadMessage(ctx context.Context) ([]byte, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockWebSocketClientInterface) WriteMessage(ctx context.Context, messageType int, data []byte) error {
	args := m.Called(ctx, messageType, data)
	return args.Error(0)
}

func (m *MockWebSocketClientInterface) WriteTextMessage(ctx context.Context, data []byte) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *MockWebSocketClientInterface) WriteBinaryMessage(ctx context.Context, data []byte) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *MockWebSocketClientInterface) IsConnected() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockWebSocketClientInterface) GetReconnectCount() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockWebSocketClientInterface) GetLastMessageTime() time.Time {
	args := m.Called()
	if args.Get(0) == nil {
		return time.Time{}
	}
	return args.Get(0).(time.Time)
}

func (m *MockWebSocketClientInterface) MessageChannel() <-chan []byte {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	// Handle both chan []byte and <-chan []byte
	if ch, ok := args.Get(0).(<-chan []byte); ok {
		return ch
	}
	if ch, ok := args.Get(0).(chan []byte); ok {
		return ch
	}
	return nil
}

func (m *MockWebSocketClientInterface) ErrorChannel() <-chan error {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	// Handle both chan error and <-chan error
	if ch, ok := args.Get(0).(<-chan error); ok {
		return ch
	}
	if ch, ok := args.Get(0).(chan error); ok {
		return ch
	}
	return nil
}

func (m *MockWebSocketClientInterface) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockWebSocketClientInterface) GetConn() interface{} {
	args := m.Called()
	return args.Get(0)
}

// Helper function to create a mock HTTP client that can be used in NodeClient
// Note: This requires modifying NodeClient to accept an interface, or we use the real HTTPClient
// For now, we'll test with real HTTPClient but mock the responses via test server

// MockResponseBuilder helps build mock HTTP responses for testing
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
