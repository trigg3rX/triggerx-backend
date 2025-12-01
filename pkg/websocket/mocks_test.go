package websocket

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestMockWebSocketClient_Connect_Success(t *testing.T) {
	mockClient := &MockWebSocketClient{}

	// Setup mock expectations
	mockClient.On("Connect", mock.Anything).Return(nil)

	// Execute
	err := mockClient.Connect(context.Background())

	// Assert
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestMockWebSocketClient_Connect_Error(t *testing.T) {
	mockClient := &MockWebSocketClient{}

	// Setup mock expectations
	expectedErr := errors.New("connection failed")
	mockClient.On("Connect", mock.Anything).Return(expectedErr)

	// Execute
	err := mockClient.Connect(context.Background())

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockClient.AssertExpectations(t)
}

func TestMockWebSocketClient_ReadMessage_Success(t *testing.T) {
	mockClient := &MockWebSocketClient{}

	// Setup mock expectations
	expectedMessage := []byte("test message")
	mockClient.On("ReadMessage", mock.Anything).Return(expectedMessage, nil)

	// Execute
	message, err := mockClient.ReadMessage(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedMessage, message)
	mockClient.AssertExpectations(t)
}

func TestMockWebSocketClient_ReadMessage_Error(t *testing.T) {
	mockClient := &MockWebSocketClient{}

	// Setup mock expectations
	expectedErr := errors.New("read error")
	mockClient.On("ReadMessage", mock.Anything).Return(nil, expectedErr)

	// Execute
	message, err := mockClient.ReadMessage(context.Background())

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, message)
	mockClient.AssertExpectations(t)
}

func TestMockWebSocketClient_WriteTextMessage_Success(t *testing.T) {
	mockClient := &MockWebSocketClient{}

	// Setup mock expectations
	testData := []byte("test message")
	mockClient.On("WriteTextMessage", mock.Anything, testData).Return(nil)

	// Execute
	err := mockClient.WriteTextMessage(context.Background(), testData)

	// Assert
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestMockWebSocketClient_WriteTextMessage_Error(t *testing.T) {
	mockClient := &MockWebSocketClient{}

	// Setup mock expectations
	testData := []byte("test message")
	expectedErr := errors.New("write error")
	mockClient.On("WriteTextMessage", mock.Anything, testData).Return(expectedErr)

	// Execute
	err := mockClient.WriteTextMessage(context.Background(), testData)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockClient.AssertExpectations(t)
}

func TestMockWebSocketClient_WriteBinaryMessage_Success(t *testing.T) {
	mockClient := &MockWebSocketClient{}

	// Setup mock expectations
	testData := []byte{0x01, 0x02, 0x03}
	mockClient.On("WriteBinaryMessage", mock.Anything, testData).Return(nil)

	// Execute
	err := mockClient.WriteBinaryMessage(context.Background(), testData)

	// Assert
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestMockWebSocketClient_WriteBinaryMessage_Error(t *testing.T) {
	mockClient := &MockWebSocketClient{}

	// Setup mock expectations
	testData := []byte{0x01, 0x02, 0x03}
	expectedErr := errors.New("write error")
	mockClient.On("WriteBinaryMessage", mock.Anything, testData).Return(expectedErr)

	// Execute
	err := mockClient.WriteBinaryMessage(context.Background(), testData)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockClient.AssertExpectations(t)
}

func TestMockWebSocketClient_WriteMessage_Success(t *testing.T) {
	mockClient := &MockWebSocketClient{}

	// Setup mock expectations
	testData := []byte("test message")
	messageType := 1 // TextMessage
	mockClient.On("WriteMessage", mock.Anything, messageType, testData).Return(nil)

	// Execute
	err := mockClient.WriteMessage(context.Background(), messageType, testData)

	// Assert
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestMockWebSocketClient_IsConnected_ReturnsTrue(t *testing.T) {
	mockClient := &MockWebSocketClient{}

	// Setup mock expectations
	mockClient.On("IsConnected").Return(true)

	// Execute
	connected := mockClient.IsConnected()

	// Assert
	assert.True(t, connected)
	mockClient.AssertExpectations(t)
}

func TestMockWebSocketClient_IsConnected_ReturnsFalse(t *testing.T) {
	mockClient := &MockWebSocketClient{}

	// Setup mock expectations
	mockClient.On("IsConnected").Return(false)

	// Execute
	connected := mockClient.IsConnected()

	// Assert
	assert.False(t, connected)
	mockClient.AssertExpectations(t)
}

func TestMockWebSocketClient_GetReconnectCount_ReturnsCount(t *testing.T) {
	mockClient := &MockWebSocketClient{}

	// Setup mock expectations
	expectedCount := 5
	mockClient.On("GetReconnectCount").Return(expectedCount)

	// Execute
	count := mockClient.GetReconnectCount()

	// Assert
	assert.Equal(t, expectedCount, count)
	mockClient.AssertExpectations(t)
}

func TestMockWebSocketClient_GetLastMessageTime_ReturnsTime(t *testing.T) {
	mockClient := &MockWebSocketClient{}

	// Setup mock expectations
	expectedTime := time.Now()
	mockClient.On("GetLastMessageTime").Return(expectedTime)

	// Execute
	lastMessage := mockClient.GetLastMessageTime()

	// Assert
	assert.Equal(t, expectedTime, lastMessage)
	mockClient.AssertExpectations(t)
}

func TestMockWebSocketClient_GetLastMessageTime_ReturnsZeroTime(t *testing.T) {
	mockClient := &MockWebSocketClient{}

	// Setup mock expectations
	mockClient.On("GetLastMessageTime").Return(time.Time{})

	// Execute
	lastMessage := mockClient.GetLastMessageTime()

	// Assert
	assert.True(t, lastMessage.IsZero())
	mockClient.AssertExpectations(t)
}

func TestMockWebSocketClient_MessageChannel_ReturnsChannel(t *testing.T) {
	mockClient := &MockWebSocketClient{}

	// Setup mock expectations
	expectedChan := make(chan []byte, 10)
	mockClient.On("MessageChannel").Return((<-chan []byte)(expectedChan))

	// Execute
	msgChan := mockClient.MessageChannel()

	// Assert
	// Verify the channel is returned (can't directly compare channel types)
	assert.NotNil(t, msgChan)
	// The actual channel comparison is handled by the mock framework
	mockClient.AssertExpectations(t)
}

func TestMockWebSocketClient_ErrorChannel_ReturnsChannel(t *testing.T) {
	mockClient := &MockWebSocketClient{}

	// Setup mock expectations
	expectedChan := make(chan error, 10)
	mockClient.On("ErrorChannel").Return((<-chan error)(expectedChan))

	// Execute
	errChan := mockClient.ErrorChannel()

	// Assert
	// Verify the channel is returned (can't directly compare channel types)
	assert.NotNil(t, errChan)
	// The actual channel comparison is handled by the mock framework
	mockClient.AssertExpectations(t)
}

func TestMockWebSocketClient_Close_Success(t *testing.T) {
	mockClient := &MockWebSocketClient{}

	// Setup mock expectations
	mockClient.On("Close").Return(nil)

	// Execute
	err := mockClient.Close()

	// Assert
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestMockWebSocketClient_Close_Error(t *testing.T) {
	mockClient := &MockWebSocketClient{}

	// Setup mock expectations
	expectedErr := errors.New("close error")
	mockClient.On("Close").Return(expectedErr)

	// Execute
	err := mockClient.Close()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockClient.AssertExpectations(t)
}

func TestMockWebSocketClient_GetConn_ReturnsConnection(t *testing.T) {
	mockClient := &MockWebSocketClient{}

	// Setup mock expectations
	expectedConn := "mock connection"
	mockClient.On("GetConn").Return(expectedConn)

	// Execute
	conn := mockClient.GetConn()

	// Assert
	assert.Equal(t, expectedConn, conn)
	mockClient.AssertExpectations(t)
}

func TestMockWebSocketClient_GetConn_ReturnsNil(t *testing.T) {
	mockClient := &MockWebSocketClient{}

	// Setup mock expectations
	mockClient.On("GetConn").Return(nil)

	// Execute
	conn := mockClient.GetConn()

	// Assert
	assert.Nil(t, conn)
	mockClient.AssertExpectations(t)
}

func TestMockWebSocketClient_ComplexScenario(t *testing.T) {
	mockClient := &MockWebSocketClient{}

	// Setup complex scenario
	ctx := context.Background()
	testMessage := []byte("test")

	// Connect
	mockClient.On("Connect", ctx).Return(nil).Once()
	// Check connection
	mockClient.On("IsConnected").Return(true).Once()
	// Write message
	mockClient.On("WriteTextMessage", ctx, testMessage).Return(nil).Once()
	// Read response
	response := []byte("response")
	mockClient.On("ReadMessage", ctx).Return(response, nil).Once()
	// Close
	mockClient.On("Close").Return(nil).Once()

	// Execute
	err := mockClient.Connect(ctx)
	require.NoError(t, err)

	connected := mockClient.IsConnected()
	require.True(t, connected)

	err = mockClient.WriteTextMessage(ctx, testMessage)
	require.NoError(t, err)

	msg, err := mockClient.ReadMessage(ctx)
	require.NoError(t, err)
	assert.Equal(t, response, msg)

	err = mockClient.Close()
	require.NoError(t, err)

	// Assert all expectations
	mockClient.AssertExpectations(t)
}
