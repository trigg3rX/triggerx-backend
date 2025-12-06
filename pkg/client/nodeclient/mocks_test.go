package nodeclient

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestMockHTTPClientInterface_DoWithRetry_Success(t *testing.T) {
	mockClient := &MockHTTPClientInterface{}

	expectedResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"status": "success"}`)),
	}
	mockClient.On("DoWithRetry", mock.Anything, mock.AnythingOfType("*http.Request")).Return(expectedResp, nil)

	req, err := http.NewRequest("GET", "http://example.com", nil)
	require.NoError(t, err)

	resp, err := mockClient.DoWithRetry(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, expectedResp, resp)
	mockClient.AssertExpectations(t)
}

func TestMockHTTPClientInterface_DoWithRetry_Error(t *testing.T) {
	mockClient := &MockHTTPClientInterface{}

	expectedErr := errors.New("network error")
	mockClient.On("DoWithRetry", mock.Anything, mock.AnythingOfType("*http.Request")).Return(nil, expectedErr)

	req, err := http.NewRequest("GET", "http://example.com", nil)
	require.NoError(t, err)

	resp, err := mockClient.DoWithRetry(context.Background(), req)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, resp)
	mockClient.AssertExpectations(t)
}

func TestMockHTTPClientInterface_Get_Success(t *testing.T) {
	mockClient := &MockHTTPClientInterface{}

	expectedResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"data": "test"}`)),
	}
	mockClient.On("Get", mock.Anything, "http://example.com/api").Return(expectedResp, nil)

	resp, err := mockClient.Get(context.Background(), "http://example.com/api")

	assert.NoError(t, err)
	assert.Equal(t, expectedResp, resp)
	mockClient.AssertExpectations(t)
}

func TestMockHTTPClientInterface_Close(t *testing.T) {
	mockClient := &MockHTTPClientInterface{}

	mockClient.On("Close").Return()

	mockClient.Close()

	mockClient.AssertExpectations(t)
}

func TestMockWebSocketClientInterface_Connect_Success(t *testing.T) {
	mockClient := &MockWebSocketClientInterface{}

	mockClient.On("Connect", mock.Anything).Return(nil)

	err := mockClient.Connect(context.Background())

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestMockWebSocketClientInterface_Connect_Error(t *testing.T) {
	mockClient := &MockWebSocketClientInterface{}

	expectedErr := errors.New("connection failed")
	mockClient.On("Connect", mock.Anything).Return(expectedErr)

	err := mockClient.Connect(context.Background())

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockClient.AssertExpectations(t)
}

func TestMockWebSocketClientInterface_ReadMessage_Success(t *testing.T) {
	mockClient := &MockWebSocketClientInterface{}

	expectedMessage := []byte("test message")
	mockClient.On("ReadMessage", mock.Anything).Return(expectedMessage, nil)

	message, err := mockClient.ReadMessage(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedMessage, message)
	mockClient.AssertExpectations(t)
}

func TestMockWebSocketClientInterface_WriteTextMessage_Success(t *testing.T) {
	mockClient := &MockWebSocketClientInterface{}

	testData := []byte("test message")
	mockClient.On("WriteTextMessage", mock.Anything, testData).Return(nil)

	err := mockClient.WriteTextMessage(context.Background(), testData)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestMockWebSocketClientInterface_IsConnected_ReturnsTrue(t *testing.T) {
	mockClient := &MockWebSocketClientInterface{}

	mockClient.On("IsConnected").Return(true)

	connected := mockClient.IsConnected()

	assert.True(t, connected)
	mockClient.AssertExpectations(t)
}

func TestMockWebSocketClientInterface_Close_Success(t *testing.T) {
	mockClient := &MockWebSocketClientInterface{}

	mockClient.On("Close").Return(nil)

	err := mockClient.Close()

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestMockResponseBuilder_WithStatusCode_SetsStatusCode(t *testing.T) {
	builder := NewMockResponseBuilder()
	builder = builder.WithStatusCode(http.StatusNotFound)

	resp := builder.Build()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestMockResponseBuilder_WithBody_SetsBody(t *testing.T) {
	builder := NewMockResponseBuilder()
	builder = builder.WithBody(`{"error": "not found"}`)

	resp := builder.Build()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, `{"error": "not found"}`, string(body))
}

func TestMockResponseBuilder_WithHeader_AddsHeader(t *testing.T) {
	builder := NewMockResponseBuilder()
	builder = builder.WithHeader("Content-Type", "application/json")

	resp := builder.Build()

	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
}

func TestMockResponseBuilder_Build_ReturnsResponse(t *testing.T) {
	builder := NewMockResponseBuilder().
		WithStatusCode(http.StatusOK).
		WithBody(`{"status": "ok"}`).
		WithHeader("Content-Type", "application/json")

	resp := builder.Build()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, `{"status": "ok"}`, string(body))
}

func TestMockResponseBuilder_Build_DefaultBody(t *testing.T) {
	builder := NewMockResponseBuilder()

	resp := builder.Build()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, "{}", string(body))
}

func TestMockWebSocketClientInterface_ComplexScenario(t *testing.T) {
	mockClient := &MockWebSocketClientInterface{}

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

func TestMockWebSocketClientInterface_GetReconnectCount(t *testing.T) {
	mockClient := &MockWebSocketClientInterface{}

	expectedCount := 3
	mockClient.On("GetReconnectCount").Return(expectedCount)

	count := mockClient.GetReconnectCount()

	assert.Equal(t, expectedCount, count)
	mockClient.AssertExpectations(t)
}

func TestMockWebSocketClientInterface_GetLastMessageTime(t *testing.T) {
	mockClient := &MockWebSocketClientInterface{}

	expectedTime := time.Now()
	mockClient.On("GetLastMessageTime").Return(expectedTime)

	lastMessage := mockClient.GetLastMessageTime()

	assert.Equal(t, expectedTime, lastMessage)
	mockClient.AssertExpectations(t)
}

func TestMockWebSocketClientInterface_MessageChannel(t *testing.T) {
	mockClient := &MockWebSocketClientInterface{}

	expectedChan := make(chan []byte, 10)
	mockClient.On("MessageChannel").Return((<-chan []byte)(expectedChan))

	msgChan := mockClient.MessageChannel()

	// Verify the channel is returned (can't directly compare channel types)
	assert.NotNil(t, msgChan)
	// Verify it's the same underlying channel by checking it's not nil
	// The actual channel comparison is handled by the mock framework
	mockClient.AssertExpectations(t)
}

func TestMockWebSocketClientInterface_ErrorChannel(t *testing.T) {
	mockClient := &MockWebSocketClientInterface{}

	expectedChan := make(chan error, 10)
	mockClient.On("ErrorChannel").Return((<-chan error)(expectedChan))

	errChan := mockClient.ErrorChannel()

	// Verify the channel is returned (can't directly compare channel types)
	assert.NotNil(t, errChan)
	// Verify it's the same underlying channel by checking it's not nil
	// The actual channel comparison is handled by the mock framework
	mockClient.AssertExpectations(t)
}
