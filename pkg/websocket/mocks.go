package websocket

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

// MockWebSocketClient is a mock implementation of the WebSocketClientInterface
type MockWebSocketClient struct {
	mock.Mock
}

// Connect mocks the Connect method
func (m *MockWebSocketClient) Connect(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// ReadMessage mocks the ReadMessage method
func (m *MockWebSocketClient) ReadMessage(ctx context.Context) ([]byte, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

// WriteMessage mocks the WriteMessage method
func (m *MockWebSocketClient) WriteMessage(ctx context.Context, messageType int, data []byte) error {
	args := m.Called(ctx, messageType, data)
	return args.Error(0)
}

// WriteTextMessage mocks the WriteTextMessage method
func (m *MockWebSocketClient) WriteTextMessage(ctx context.Context, data []byte) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

// WriteBinaryMessage mocks the WriteBinaryMessage method
func (m *MockWebSocketClient) WriteBinaryMessage(ctx context.Context, data []byte) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

// IsConnected mocks the IsConnected method
func (m *MockWebSocketClient) IsConnected() bool {
	args := m.Called()
	return args.Bool(0)
}

// GetReconnectCount mocks the GetReconnectCount method
func (m *MockWebSocketClient) GetReconnectCount() int {
	args := m.Called()
	return args.Int(0)
}

// GetLastMessageTime mocks the GetLastMessageTime method
func (m *MockWebSocketClient) GetLastMessageTime() time.Time {
	args := m.Called()
	if args.Get(0) == nil {
		return time.Time{}
	}
	return args.Get(0).(time.Time)
}

// MessageChannel mocks the MessageChannel method
func (m *MockWebSocketClient) MessageChannel() <-chan []byte {
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

// ErrorChannel mocks the ErrorChannel method
func (m *MockWebSocketClient) ErrorChannel() <-chan error {
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

// Close mocks the Close method
func (m *MockWebSocketClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

// GetConn mocks the GetConn method
func (m *MockWebSocketClient) GetConn() interface{} {
	args := m.Called()
	return args.Get(0)
}
