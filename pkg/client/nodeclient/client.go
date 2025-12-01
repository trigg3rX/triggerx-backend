package nodeclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	httppkg "github.com/trigg3rX/triggerx-backend/pkg/http"
	wsclient "github.com/trigg3rX/triggerx-backend/pkg/websocket"
)

var (
	// ErrInvalidConfig is returned when the client configuration is invalid
	ErrInvalidConfig = fmt.Errorf("invalid client configuration")
	// ErrRPCError is returned when an RPC error occurs
	ErrRPCError = fmt.Errorf("RPC error")
	// ErrInvalidResponse is returned when the response cannot be parsed
	ErrInvalidResponse = fmt.Errorf("invalid response")
)

// NodeClient handles communication with blockchain node via Alchemy API
type NodeClient struct {
	config          *Config
	httpClient      *httppkg.HTTPClient
	requestID       int // Counter for JSON-RPC request IDs
	wsClient        *wsclient.WebSocketClient
	wsSubManager    *WebSocketSubscriptionManager
	pendingRequests map[int]chan *RPCResponse
	mu              sync.RWMutex
}

// NewNodeClient creates a new instance of NodeClient
func NewNodeClient(cfg *Config) (*NodeClient, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidConfig, err)
	}

	// Create HTTP client if not provided
	var httpClient *httppkg.HTTPClient
	var err error
	if cfg.HTTPConfig != nil {
		httpClient, err = httppkg.NewHTTPClient(cfg.HTTPConfig, cfg.Logger)
	} else {
		// Use default HTTP config
		httpConfig := httppkg.DefaultHTTPRetryConfig()
		// Override timeout if specified in config
		if cfg.RequestTimeout > 0 {
			httpConfig.Timeout = cfg.RequestTimeout
		}
		httpClient, err = httppkg.NewHTTPClient(httpConfig, cfg.Logger)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	return &NodeClient{
		config:          cfg,
		httpClient:      httpClient,
		requestID:       1,
		pendingRequests: make(map[int]chan *RPCResponse),
	}, nil
}

// Close closes the HTTP client's idle connections and WebSocket connection
func (c *NodeClient) Close() {
	if c.httpClient != nil {
		c.httpClient.Close()
	}
	_ = c.DisconnectWebSocket()
}

// GetHTTPClient returns the underlying HTTP client
func (c *NodeClient) GetHTTPClient() *httppkg.HTTPClient {
	return c.httpClient
}

// SetRequestTimeout updates the request timeout for the HTTP client
func (c *NodeClient) SetRequestTimeout(timeout time.Duration) error {
	if c.config.HTTPConfig == nil {
		c.config.HTTPConfig = httppkg.DefaultHTTPRetryConfig()
	}
	c.config.HTTPConfig.Timeout = timeout

	// Recreate HTTP client with new timeout
	newHTTPClient, err := httppkg.NewHTTPClient(c.config.HTTPConfig, c.config.Logger)
	if err != nil {
		return fmt.Errorf("failed to recreate HTTP client: %w", err)
	}

	// Close old client and use new one
	c.httpClient.Close()
	c.httpClient = newHTTPClient

	return nil
}

// getNextRequestID returns the next request ID and increments the counter
func (c *NodeClient) getNextRequestID() int {
	id := c.requestID
	c.requestID++
	return id
}

// call performs a JSON-RPC call and returns the raw result
func (c *NodeClient) call(ctx context.Context, method string, params []interface{}) (json.RawMessage, error) {
	req := RPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      c.getNextRequestID(),
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	c.config.Logger.Debug("Making RPC call",
		"method", method,
		"url", c.config.GetFullURL(),
		"request_id", req.ID,
	)

	// Create request with context
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.config.GetFullURL(), bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Use GetBody to allow retries
	httpReq.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewBuffer(reqBody)), nil
	}

	// Perform the request with retry
	resp, err := c.httpClient.DoWithRetry(ctx, httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status code: %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse JSON-RPC response
	var rpcResp RPCResponse
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		return nil, fmt.Errorf("%w: failed to unmarshal response: %v", ErrInvalidResponse, err)
	}

	// Check for RPC error
	if rpcResp.Error != nil {
		return nil, fmt.Errorf("%w: code=%d, message=%s, data=%v", ErrRPCError, rpcResp.Error.Code, rpcResp.Error.Message, rpcResp.Error.Data)
	}

	// Check if request ID matches (basic validation)
	if rpcResp.ID != req.ID {
		c.config.Logger.Warn("Request ID mismatch",
			"expected", req.ID,
			"received", rpcResp.ID,
		)
	}

	return rpcResp.Result, nil
}
