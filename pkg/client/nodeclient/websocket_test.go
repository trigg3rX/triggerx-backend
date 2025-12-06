package nodeclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	wsclient "github.com/trigg3rX/triggerx-backend/pkg/websocket"
)

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// TestWebSocketSubscriptionManager_AddSubscription tests AddSubscription
func TestWebSocketSubscriptionManager_AddSubscription(t *testing.T) {
	manager := NewWebSocketSubscriptionManager()
	ch := make(chan *SubscriptionNotification, 10)

	manager.AddSubscription("sub-1", ch)

	retrieved, exists := manager.GetSubscription("sub-1")
	assert.True(t, exists)
	assert.Equal(t, ch, retrieved)
}

// TestWebSocketSubscriptionManager_RemoveSubscription tests RemoveSubscription
func TestWebSocketSubscriptionManager_RemoveSubscription(t *testing.T) {
	manager := NewWebSocketSubscriptionManager()
	ch := make(chan *SubscriptionNotification, 10)

	manager.AddSubscription("sub-1", ch)
	manager.RemoveSubscription("sub-1")

	_, exists := manager.GetSubscription("sub-1")
	assert.False(t, exists)
}

// TestWebSocketSubscriptionManager_GetAllSubscriptions tests GetAllSubscriptions
func TestWebSocketSubscriptionManager_GetAllSubscriptions(t *testing.T) {
	manager := NewWebSocketSubscriptionManager()

	manager.AddSubscription("sub-1", make(chan *SubscriptionNotification, 10))
	manager.AddSubscription("sub-2", make(chan *SubscriptionNotification, 10))

	subs := manager.GetAllSubscriptions()
	assert.Len(t, subs, 2)
	assert.Contains(t, subs, "sub-1")
	assert.Contains(t, subs, "sub-2")
}

// TestNodeClient_ConnectWebSocket_Success tests ConnectWebSocket
func TestNodeClient_ConnectWebSocket_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := wsUpgrader.Upgrade(w, r, nil)
		require.NoError(t, err)
		defer func () {
			if err := conn.Close(); err != nil {
				t.Errorf("Error closing connection: %v", err)
			}
		}()

		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:]
	logger := logging.NewNoOpLogger()
	config := DefaultConfig("test-key", NetworkEthereum, logger)
	config = config.WithWebSocketURL(wsURL)
	wsConfig := wsclient.DefaultWebSocketRetryConfig()
	wsConfig.RetryConfig.MaxRetries = 1
	wsConfig.RetryConfig.InitialDelay = 10 * time.Millisecond
	config = config.WithWebSocketConfig(wsConfig)

	client, err := NewNodeClient(config)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.ConnectWebSocket(ctx)

	assert.NoError(t, err)
	assert.True(t, client.IsWebSocketConnected())

	_ = client.DisconnectWebSocket()
}

// TestNodeClient_ConnectWebSocket_NoURL_ReturnsError tests ConnectWebSocket without URL
// Note: Since config validation requires network or baseURL, we test with an invalid WebSocket URL
func TestNodeClient_ConnectWebSocket_NoURL_ReturnsError(t *testing.T) {
	logger := logging.NewNoOpLogger()
	// Create a config with a baseURL that can't be converted to WebSocket URL
	config := &Config{
		APIKey:  "test-key",
		BaseURL: "invalid-url", // Invalid URL that can't be converted
		Logger:  logger,
	}

	client, err := NewNodeClient(config)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = client.ConnectWebSocket(ctx)

	// Should fail because the derived WebSocket URL is invalid or connection fails
	assert.Error(t, err)
	// The error will be a connection error, not "WebSocket URL not configured"
	// because getWebSocketURL will derive "wsinvalid-url" which is invalid
}

// TestNodeClient_IsWebSocketConnected_ReturnsStatus tests IsWebSocketConnected
func TestNodeClient_IsWebSocketConnected_ReturnsStatus(t *testing.T) {
	logger := logging.NewNoOpLogger()
	config := DefaultConfig("test-key", NetworkEthereum, logger)

	client, err := NewNodeClient(config)
	require.NoError(t, err)

	assert.False(t, client.IsWebSocketConnected())
}

// TestNodeClient_DisconnectWebSocket_ClosesConnection tests DisconnectWebSocket
func TestNodeClient_DisconnectWebSocket_ClosesConnection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := wsUpgrader.Upgrade(w, r, nil)
		require.NoError(t, err)
		defer func () {
			if err := conn.Close(); err != nil {
				t.Errorf("Error closing connection: %v", err)
			}
		}()

		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:]
	logger := logging.NewNoOpLogger()
	config := DefaultConfig("test-key", NetworkEthereum, logger)
	config = config.WithWebSocketURL(wsURL)
	wsConfig := wsclient.DefaultWebSocketRetryConfig()
	wsConfig.RetryConfig.MaxRetries = 1
	wsConfig.RetryConfig.InitialDelay = 10 * time.Millisecond
	config = config.WithWebSocketConfig(wsConfig)

	client, err := NewNodeClient(config)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.ConnectWebSocket(ctx)
	require.NoError(t, err)

	err = client.DisconnectWebSocket()

	assert.NoError(t, err)
	assert.False(t, client.IsWebSocketConnected())
}

// TestNodeClient_EthSubscribe_Success tests EthSubscribe
func TestNodeClient_EthSubscribe_Success(t *testing.T) {
	requestReceived := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := wsUpgrader.Upgrade(w, r, nil)
		require.NoError(t, err)
		defer func () {
			if err := conn.Close(); err != nil {
				t.Errorf("Error closing connection: %v", err)
			}
		}()

		// Read subscription request
		_, message, err := conn.ReadMessage()
		require.NoError(t, err)

		var req RPCRequest
		err = json.Unmarshal(message, &req)
		require.NoError(t, err)

		if req.Method == "eth_subscribe" {
			requestReceived = true
			// Send subscription response
			resp := RPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  json.RawMessage(`"0x1234"`),
			}
			respJSON, _ := json.Marshal(resp)
			err = conn.WriteMessage(websocket.TextMessage, respJSON)
			if err != nil {
				t.Fatalf("Failed to write message: %v", err)
			}
		}

		time.Sleep(200 * time.Millisecond)
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:]
	logger := logging.NewNoOpLogger()
	config := DefaultConfig("test-key", NetworkEthereum, logger)
	config = config.WithWebSocketURL(wsURL)
	wsConfig := wsclient.DefaultWebSocketRetryConfig()
	wsConfig.RetryConfig.MaxRetries = 1
	wsConfig.RetryConfig.InitialDelay = 10 * time.Millisecond
	config = config.WithWebSocketConfig(wsConfig)

	client, err := NewNodeClient(config)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.ConnectWebSocket(ctx)
	require.NoError(t, err)

	// Wait a bit for connection to be ready
	time.Sleep(200 * time.Millisecond)

	subID, notifChan, err := client.EthSubscribe(ctx, "newHeads", nil)

	assert.NoError(t, err)
	assert.NotEmpty(t, subID)
	assert.NotNil(t, notifChan)
	assert.True(t, requestReceived)

	_ = client.DisconnectWebSocket()
}

// TestNodeClient_EthUnsubscribe_Success tests EthUnsubscribe
func TestNodeClient_EthUnsubscribe_Success(t *testing.T) {
	unsubscribeReceived := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := wsUpgrader.Upgrade(w, r, nil)
		require.NoError(t, err)
		defer func () {
			if err := conn.Close(); err != nil {
				t.Errorf("Error closing connection: %v", err)
			}
		}()

		// Read messages
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				break
			}

			var req RPCRequest
			if err := json.Unmarshal(message, &req); err != nil {
				continue
			}

			// Handle both subscribe and unsubscribe
			if req.Method == "eth_subscribe" {
				// Send subscription response
				resp := RPCResponse{
					JSONRPC: "2.0",
					ID:      req.ID,
					Result:  json.RawMessage(`"0x1234"`),
				}
				respJSON, _ := json.Marshal(resp)
				err := conn.WriteMessage(websocket.TextMessage, respJSON)
				if err != nil {
					break
				}
			} else if req.Method == "eth_unsubscribe" {
				unsubscribeReceived = true
				// Send unsubscribe response
				resp := RPCResponse{
					JSONRPC: "2.0",
					ID:      req.ID,
					Result:  json.RawMessage(`true`),
				}
				respJSON, _ := json.Marshal(resp)
				err := conn.WriteMessage(websocket.TextMessage, respJSON)
				if err != nil {
					break
				}
				break
			}
		}
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:]
	logger := logging.NewNoOpLogger()
	config := DefaultConfig("test-key", NetworkEthereum, logger)
	config = config.WithWebSocketURL(wsURL)
	wsConfig := wsclient.DefaultWebSocketRetryConfig()
	wsConfig.RetryConfig.MaxRetries = 1
	wsConfig.RetryConfig.InitialDelay = 10 * time.Millisecond
	config = config.WithWebSocketConfig(wsConfig)

	client, err := NewNodeClient(config)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.ConnectWebSocket(ctx)
	require.NoError(t, err)

	// Wait a bit for connection to be ready
	time.Sleep(200 * time.Millisecond)

	// First subscribe
	subID, _, err := client.EthSubscribe(ctx, "newHeads", nil)
	require.NoError(t, err)

	// Then unsubscribe
	err = client.EthUnsubscribe(ctx, subID)

	assert.NoError(t, err)
	assert.True(t, unsubscribeReceived)

	_ = client.DisconnectWebSocket()
}

// TestNodeClient_EthUnsubscribe_NotConnected_ReturnsError tests EthUnsubscribe when not connected
func TestNodeClient_EthUnsubscribe_NotConnected_ReturnsError(t *testing.T) {
	logger := logging.NewNoOpLogger()
	config := DefaultConfig("test-key", NetworkEthereum, logger)

	client, err := NewNodeClient(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = client.EthUnsubscribe(ctx, "sub-1")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "WebSocket not connected")
}

// TestNodeClient_GetWebSocketURL_DerivesFromBaseURL tests getWebSocketURL derivation
func TestNodeClient_GetWebSocketURL_DerivesFromBaseURL(t *testing.T) {
	logger := logging.NewNoOpLogger()
	config := DefaultConfig("test-key", NetworkEthereum, logger)
	config = config.WithBaseURL("https://eth-mainnet.g.alchemy.com/v2/")

	client, err := NewNodeClient(config)
	require.NoError(t, err)

	wsURL := client.getWebSocketURL()

	// Should derive wss:// from https://
	assert.Contains(t, wsURL, "wss://")
	assert.Contains(t, wsURL, "test-key")
}

// TestNodeClient_GetWebSocketURL_UsesExplicitURL tests getWebSocketURL with explicit URL
func TestNodeClient_GetWebSocketURL_UsesExplicitURL(t *testing.T) {
	logger := logging.NewNoOpLogger()
	config := DefaultConfig("test-key", NetworkEthereum, logger)
	config = config.WithWebSocketURL("wss://custom.example.com/ws")

	client, err := NewNodeClient(config)
	require.NoError(t, err)

	wsURL := client.getWebSocketURL()

	assert.Equal(t, "wss://custom.example.com/ws", wsURL)
}

// TestNodeClient_HandleSubscriptionNotification_RoutesToChannel tests subscription notification handling
func TestNodeClient_HandleSubscriptionNotification_RoutesToChannel(t *testing.T) {
	manager := NewWebSocketSubscriptionManager()
	ch := make(chan *SubscriptionNotification, 10)
	manager.AddSubscription("sub-1", ch)

	logger := logging.NewNoOpLogger()
	config := DefaultConfig("test-key", NetworkEthereum, logger)
	client, err := NewNodeClient(config)
	require.NoError(t, err)

	client.mu.Lock()
	client.wsSubManager = manager
	client.mu.Unlock()

	// Simulate subscription notification
	msg := map[string]interface{}{
		"method": "eth_subscription",
		"params": map[string]interface{}{
			"subscription": "sub-1",
			"result":       map[string]interface{}{"block": "0x1234"},
		},
	}

	client.handleSubscriptionNotification(msg)

	// Check if notification was sent to channel
	select {
	case notif := <-ch:
		assert.Equal(t, "sub-1", notif.Subscription)
		assert.NotNil(t, notif.Result)
	case <-time.After(1 * time.Second):
		t.Fatal("Notification not received in channel")
	}
}

// TestNodeClient_HandleResponse_RoutesToPendingRequest tests response handling
func TestNodeClient_HandleResponse_RoutesToPendingRequest(t *testing.T) {
	logger := logging.NewNoOpLogger()
	config := DefaultConfig("test-key", NetworkEthereum, logger)
	client, err := NewNodeClient(config)
	require.NoError(t, err)

	responseChan := make(chan *RPCResponse, 1)
	client.mu.Lock()
	client.pendingRequests[1] = responseChan
	client.mu.Unlock()

	// Simulate response
	message := []byte(`{"jsonrpc":"2.0","id":1,"result":"0x1234"}`)
	client.handleResponse(1, message)

	// Check if response was sent to channel
	select {
	case resp := <-responseChan:
		assert.Equal(t, 1, resp.ID)
		assert.NotNil(t, resp.Result)
	case <-time.After(1 * time.Second):
		t.Fatal("Response not received in channel")
	}
}
