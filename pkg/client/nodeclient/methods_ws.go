package nodeclient

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	wsclient "github.com/trigg3rX/triggerx-backend/pkg/websocket"
)

// SubscriptionNotification represents a subscription notification from the WebSocket
type SubscriptionNotification struct {
	Subscription string          `json:"subscription"`
	Result       json.RawMessage `json:"result"`
}

// SubscriptionResponse represents the response from eth_subscribe
type SubscriptionResponse struct {
	SubscriptionID string `json:"subscription_id"`
}

// WebSocketSubscriptionManager manages WebSocket subscriptions
type WebSocketSubscriptionManager struct {
	subscriptions map[string]chan *SubscriptionNotification
	mu            sync.RWMutex
}

// NewWebSocketSubscriptionManager creates a new subscription manager
func NewWebSocketSubscriptionManager() *WebSocketSubscriptionManager {
	return &WebSocketSubscriptionManager{
		subscriptions: make(map[string]chan *SubscriptionNotification),
	}
}

// AddSubscription adds a subscription channel
func (sm *WebSocketSubscriptionManager) AddSubscription(subID string, ch chan *SubscriptionNotification) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.subscriptions[subID] = ch
}

// RemoveSubscription removes a subscription channel
func (sm *WebSocketSubscriptionManager) RemoveSubscription(subID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if ch, exists := sm.subscriptions[subID]; exists {
		close(ch)
		delete(sm.subscriptions, subID)
	}
}

// GetSubscription returns a subscription channel
func (sm *WebSocketSubscriptionManager) GetSubscription(subID string) (chan *SubscriptionNotification, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	ch, exists := sm.subscriptions[subID]
	return ch, exists
}

// GetAllSubscriptions returns all subscription IDs
func (sm *WebSocketSubscriptionManager) GetAllSubscriptions() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	ids := make([]string, 0, len(sm.subscriptions))
	for id := range sm.subscriptions {
		ids = append(ids, id)
	}
	return ids
}

// ConnectWebSocket connects to the WebSocket endpoint
func (c *NodeClient) ConnectWebSocket(ctx context.Context) error {
	// Get WebSocket URL from config
	wsURL := c.getWebSocketURL()
	if wsURL == "" {
		return fmt.Errorf("WebSocket URL not configured")
	}

	// Create WebSocket config if not exists
	if c.config.WebSocketConfig == nil {
		c.config.WebSocketConfig = wsclient.DefaultWebSocketRetryConfig()
	}

	// Create WebSocket client
	wsClient, err := wsclient.NewWebSocketClient(wsURL, c.config.WebSocketConfig, c.config.Logger)
	if err != nil {
		return fmt.Errorf("failed to create WebSocket client: %w", err)
	}

	// Connect
	if err := wsClient.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect WebSocket: %w", err)
	}

	// Store WebSocket client
	c.mu.Lock()
	c.wsClient = wsClient
	if c.wsSubManager == nil {
		c.wsSubManager = NewWebSocketSubscriptionManager()
	}
	c.mu.Unlock()

	// Start message handler
	go c.handleWebSocketMessages(ctx)

	c.config.Logger.Infof("WebSocket connected to %s", wsURL)
	return nil
}

// DisconnectWebSocket disconnects from the WebSocket endpoint
func (c *NodeClient) DisconnectWebSocket() error {
	c.mu.Lock()
	wsClient := c.wsClient
	c.wsClient = nil

	// Close all pending requests
	for requestID, responseChan := range c.pendingRequests {
		close(responseChan)
		delete(c.pendingRequests, requestID)
	}

	// Close all subscriptions
	if c.wsSubManager != nil {
		for _, subID := range c.wsSubManager.GetAllSubscriptions() {
			c.wsSubManager.RemoveSubscription(subID)
		}
	}
	c.mu.Unlock()

	if wsClient != nil {
		return wsClient.Close()
	}
	return nil
}

// IsWebSocketConnected returns whether the WebSocket is connected
func (c *NodeClient) IsWebSocketConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.wsClient != nil && c.wsClient.IsConnected()
}

// getWebSocketURL returns the WebSocket URL from config
func (c *NodeClient) getWebSocketURL() string {
	if c.config.WebSocketURL != "" {
		return c.config.WebSocketURL
	}
	// Derive from HTTP URL if possible
	baseURL := c.config.GetBaseURL()
	if baseURL != "" {
		// Convert https:// to wss:// or http:// to ws://
		wsURL := baseURL
		if len(wsURL) >= 5 && wsURL[:5] == "https" {
			wsURL = "wss" + wsURL[5:]
		} else if len(wsURL) >= 4 && wsURL[:4] == "http" {
			wsURL = "ws" + wsURL[4:]
		}
		return wsURL + c.config.APIKey
	}
	return ""
}

// handleWebSocketMessages handles incoming WebSocket messages
func (c *NodeClient) handleWebSocketMessages(ctx context.Context) {
	c.mu.RLock()
	wsClient := c.wsClient
	c.mu.RUnlock()

	if wsClient == nil {
		return
	}

	msgChan := wsClient.MessageChannel()
	errChan := wsClient.ErrorChannel()

	for {
		select {
		case <-ctx.Done():
			return
		case message, ok := <-msgChan:
			if !ok {
				return
			}
			// Parse message
			var msg map[string]interface{}
			if err := json.Unmarshal(message, &msg); err != nil {
				c.config.Logger.Warnf("Failed to parse WebSocket message: %v", err)
				continue
			}

			// Check if it's a subscription notification (no ID field)
			if method, ok := msg["method"].(string); ok && method == "eth_subscription" {
				c.handleSubscriptionNotification(msg)
				continue
			}

			// Otherwise, it's a response to a request (has ID field)
			if id, ok := msg["id"].(float64); ok {
				requestID := int(id)
				c.handleResponse(requestID, message)
			}
		case err, ok := <-errChan:
			if !ok {
				return
			}
			c.config.Logger.Errorf("WebSocket error: %v", err)
		}
	}
}

// handleResponse routes a response to the appropriate pending request
func (c *NodeClient) handleResponse(requestID int, message []byte) {
	c.mu.RLock()
	responseChan, exists := c.pendingRequests[requestID]
	c.mu.RUnlock()

	if !exists {
		c.config.Logger.Warnf("Received response for unknown request ID: %d", requestID)
		return
	}

	// Parse response
	var rpcResp RPCResponse
	if err := json.Unmarshal(message, &rpcResp); err != nil {
		c.config.Logger.Warnf("Failed to parse response: %v", err)
		return
	}

	// Send to waiting channel
	select {
	case responseChan <- &rpcResp:
	default:
		c.config.Logger.Warnf("Response channel full for request ID: %d", requestID)
	}

	// Clean up
	c.mu.Lock()
	delete(c.pendingRequests, requestID)
	c.mu.Unlock()
}

// handleSubscriptionNotification handles subscription notifications
func (c *NodeClient) handleSubscriptionNotification(msg map[string]interface{}) {
	params, ok := msg["params"].(map[string]interface{})
	if !ok {
		c.config.Logger.Warnf("Invalid subscription notification format")
		return
	}

	subID, ok := params["subscription"].(string)
	if !ok {
		c.config.Logger.Warnf("Missing subscription ID in notification")
		return
	}

	result, ok := params["result"]
	if !ok {
		c.config.Logger.Warnf("Missing result in subscription notification")
		return
	}

	resultBytes, err := json.Marshal(result)
	if err != nil {
		c.config.Logger.Warnf("Failed to marshal subscription result: %v", err)
		return
	}

	notification := &SubscriptionNotification{
		Subscription: subID,
		Result:       resultBytes,
	}

	// Route to subscription channel
	if c.wsSubManager != nil {
		if ch, exists := c.wsSubManager.GetSubscription(subID); exists {
			select {
			case ch <- notification:
			default:
				c.config.Logger.Warnf("Subscription channel full for %s", subID)
			}
		}
	}
}

// EthSubscribe subscribes to Ethereum events via WebSocket
// subscriptionType can be "newHeads", "logs", "newPendingTransactions", etc.
// params are the subscription parameters (e.g., filter object for logs)
func (c *NodeClient) EthSubscribe(ctx context.Context, subscriptionType string, params interface{}) (string, <-chan *SubscriptionNotification, error) {
	// Ensure WebSocket is connected
	if !c.IsWebSocketConnected() {
		if err := c.ConnectWebSocket(ctx); err != nil {
			return "", nil, fmt.Errorf("failed to connect WebSocket: %w", err)
		}
	}

	c.mu.RLock()
	wsClient := c.wsClient
	c.mu.RUnlock()

	if wsClient == nil {
		return "", nil, fmt.Errorf("WebSocket client not initialized")
	}

	// Build subscription parameters
	var rpcParams []interface{}
	if params != nil {
		rpcParams = []interface{}{subscriptionType, params}
	} else {
		rpcParams = []interface{}{subscriptionType}
	}

	// Create JSON-RPC request
	req := RPCRequest{
		JSONRPC: "2.0",
		Method:  "eth_subscribe",
		Params:  rpcParams,
		ID:      c.getNextRequestID(),
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal subscription request: %w", err)
	}

	// Create response channel
	responseChan := make(chan *RPCResponse, 1)

	// Register pending request
	c.mu.Lock()
	if c.pendingRequests == nil {
		c.pendingRequests = make(map[int]chan *RPCResponse)
	}
	c.pendingRequests[req.ID] = responseChan
	c.mu.Unlock()

	// Send request
	if err := wsClient.WriteTextMessage(ctx, reqBody); err != nil {
		c.mu.Lock()
		delete(c.pendingRequests, req.ID)
		c.mu.Unlock()
		return "", nil, fmt.Errorf("failed to send subscription request: %w", err)
	}

	// Wait for response with timeout
	responseCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var rpcResp *RPCResponse
	select {
	case <-responseCtx.Done():
		c.mu.Lock()
		delete(c.pendingRequests, req.ID)
		c.mu.Unlock()
		return "", nil, fmt.Errorf("timeout waiting for subscription response: %w", responseCtx.Err())
	case rpcResp = <-responseChan:
		// Response received
	}

	// Check for RPC error
	if rpcResp.Error != nil {
		return "", nil, fmt.Errorf("%w: code=%d, message=%s, data=%v", ErrRPCError, rpcResp.Error.Code, rpcResp.Error.Message, rpcResp.Error.Data)
	}

	// Extract subscription ID
	var subscriptionID string
	if err := json.Unmarshal(rpcResp.Result, &subscriptionID); err != nil {
		return "", nil, fmt.Errorf("failed to unmarshal subscription ID: %w", err)
	}

	// Create notification channel
	notifChan := make(chan *SubscriptionNotification, 100)

	// Register subscription
	if c.wsSubManager == nil {
		c.mu.Lock()
		c.wsSubManager = NewWebSocketSubscriptionManager()
		c.mu.Unlock()
	}
	c.wsSubManager.AddSubscription(subscriptionID, notifChan)

	c.config.Logger.Infof("Subscribed to %s with ID: %s", subscriptionType, subscriptionID)

	return subscriptionID, notifChan, nil
}

// EthUnsubscribe unsubscribes from a WebSocket subscription
func (c *NodeClient) EthUnsubscribe(ctx context.Context, subscriptionID string) error {
	if !c.IsWebSocketConnected() {
		return fmt.Errorf("WebSocket not connected")
	}

	c.mu.RLock()
	wsClient := c.wsClient
	c.mu.RUnlock()

	if wsClient == nil {
		return fmt.Errorf("WebSocket client not initialized")
	}

	// Create JSON-RPC request
	req := RPCRequest{
		JSONRPC: "2.0",
		Method:  "eth_unsubscribe",
		Params:  []interface{}{subscriptionID},
		ID:      c.getNextRequestID(),
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal unsubscribe request: %w", err)
	}

	// Create response channel
	responseChan := make(chan *RPCResponse, 1)

	// Register pending request
	c.mu.Lock()
	if c.pendingRequests == nil {
		c.pendingRequests = make(map[int]chan *RPCResponse)
	}
	c.pendingRequests[req.ID] = responseChan
	c.mu.Unlock()

	// Send request
	if err := wsClient.WriteTextMessage(ctx, reqBody); err != nil {
		c.mu.Lock()
		delete(c.pendingRequests, req.ID)
		c.mu.Unlock()
		return fmt.Errorf("failed to send unsubscribe request: %w", err)
	}

	// Wait for response with timeout
	responseCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var rpcResp *RPCResponse
	select {
	case <-responseCtx.Done():
		c.mu.Lock()
		delete(c.pendingRequests, req.ID)
		c.mu.Unlock()
		return fmt.Errorf("timeout waiting for unsubscribe response: %w", responseCtx.Err())
	case rpcResp = <-responseChan:
		// Response received
	}

	// Check for RPC error
	if rpcResp.Error != nil {
		return fmt.Errorf("%w: code=%d, message=%s, data=%v", ErrRPCError, rpcResp.Error.Code, rpcResp.Error.Message, rpcResp.Error.Data)
	}

	// Check result (should be true if successful)
	var result bool
	if err := json.Unmarshal(rpcResp.Result, &result); err != nil {
		return fmt.Errorf("failed to unmarshal unsubscribe result: %w", err)
	}

	if !result {
		return fmt.Errorf("unsubscribe failed: subscription %s not found", subscriptionID)
	}

	// Remove subscription from manager
	if c.wsSubManager != nil {
		c.wsSubManager.RemoveSubscription(subscriptionID)
	}

	c.config.Logger.Infof("Unsubscribed from subscription: %s", subscriptionID)

	return nil
}
