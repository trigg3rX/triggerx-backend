package websocket

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
)

// WebSocketRetryConfig holds configuration for WebSocket connection and reconnection
type WebSocketRetryConfig struct {
	// RetryConfig for initial connection attempts
	RetryConfig *retry.RetryConfig
	// ReconnectConfig for automatic reconnection after disconnection
	ReconnectConfig *ReconnectConfig
	// HandshakeTimeout for WebSocket handshake
	HandshakeTimeout time.Duration
	// ReadDeadline for reading messages
	ReadDeadline time.Duration
	// WriteDeadline for writing messages
	WriteDeadline time.Duration
	// PingInterval for sending ping messages to keep connection alive
	PingInterval time.Duration
	// PongWait is the time to wait for pong response before considering connection dead
	PongWait time.Duration
	// MaxMessageSize maximum message size in bytes
	MaxMessageSize int64
	// EnableCompression enables compression
	EnableCompression bool
}

// ReconnectConfig holds configuration for reconnection after disconnection
type ReconnectConfig struct {
	// MaxRetries maximum number of reconnection attempts (0 = unlimited)
	MaxRetries int
	// BaseDelay initial delay between reconnection attempts
	BaseDelay time.Duration
	// MaxDelay maximum delay between reconnection attempts
	MaxDelay time.Duration
	// BackoffFactor multiplier for exponential backoff
	BackoffFactor float64
	// Jitter enables jitter in backoff delays
	Jitter bool
}

// DefaultWebSocketRetryConfig returns default configuration for WebSocket operations
func DefaultWebSocketRetryConfig() *WebSocketRetryConfig {
	return &WebSocketRetryConfig{
		RetryConfig:       retry.DefaultRetryConfig(),
		ReconnectConfig:   DefaultReconnectConfig(),
		HandshakeTimeout:  10 * time.Second,
		ReadDeadline:      60 * time.Second,
		WriteDeadline:     10 * time.Second,
		PingInterval:      30 * time.Second,
		PongWait:          60 * time.Second,
		MaxMessageSize:    512 * 1024, // 512KB
		EnableCompression: false,
	}
}

// DefaultReconnectConfig returns default reconnection configuration
func DefaultReconnectConfig() *ReconnectConfig {
	return &ReconnectConfig{
		MaxRetries:    0, // Unlimited retries by default
		BaseDelay:     5 * time.Second,
		MaxDelay:      5 * time.Minute,
		BackoffFactor: 2.0,
		Jitter:        true,
	}
}

// Validate checks the WebSocket configuration for reasonable values
func (c *WebSocketRetryConfig) Validate() error {
	if c.HandshakeTimeout <= 0 {
		return fmt.Errorf("handshakeTimeout must be positive")
	}
	if c.ReadDeadline <= 0 {
		return fmt.Errorf("readDeadline must be positive")
	}
	if c.WriteDeadline <= 0 {
		return fmt.Errorf("writeDeadline must be positive")
	}
	if c.PingInterval <= 0 {
		return fmt.Errorf("pingInterval must be positive")
	}
	if c.PongWait <= 0 {
		return fmt.Errorf("pongWait must be positive")
	}
	if c.PongWait <= c.PingInterval {
		return fmt.Errorf("pongWait must be greater than pingInterval")
	}
	if c.MaxMessageSize < 0 {
		return fmt.Errorf("maxMessageSize must be >= 0")
	}
	return nil
}

// WebSocketError represents a WebSocket-specific error
type WebSocketError struct {
	Code    int
	Message string
}

func (e *WebSocketError) Error() string {
	return fmt.Sprintf("WebSocket error [%d]: %s", e.Code, e.Message)
}

// WebSocketClient is a wrapper around websocket.Conn that includes retry and reconnection logic
type WebSocketClient struct {
	url            string
	conn           *websocket.Conn
	config         *WebSocketRetryConfig
	logger         logging.Logger
	mu             sync.RWMutex
	isConnected    bool
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	reconnectCount int
	lastMessage    time.Time
	messageChan    chan []byte
	errorChan      chan error
	headers        http.Header
	dialer         *websocket.Dialer
	closed         bool
}

// NewWebSocketClient creates a new WebSocket client with retry and reconnection capabilities
func NewWebSocketClient(url string, config *WebSocketRetryConfig, logger logging.Logger) (*WebSocketClient, error) {
	if config == nil {
		config = DefaultWebSocketRetryConfig()
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid WebSocket retry config: %w", err)
	}

	// Set up default retry predicate if none provided
	if config.RetryConfig.ShouldRetry == nil {
		config.RetryConfig.ShouldRetry = func(err error, attempt int) bool {
			// Retry on network errors and temporary failures
			var wsErr *WebSocketError
			if errors.As(err, &wsErr) {
				// Retry on temporary errors (code 1006, 1011, etc.)
				return wsErr.Code == websocket.CloseAbnormalClosure ||
					wsErr.Code == websocket.CloseGoingAway ||
					wsErr.Code == websocket.CloseInternalServerErr
			}
			// For all other errors (network errors, etc.), assume they are retryable
			return true
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	dialer := &websocket.Dialer{
		HandshakeTimeout:  config.HandshakeTimeout,
		EnableCompression: config.EnableCompression,
		Proxy:             http.ProxyFromEnvironment,
	}

	return &WebSocketClient{
		url:         url,
		config:      config,
		logger:      logger,
		ctx:         ctx,
		cancel:      cancel,
		isConnected: false,
		messageChan: make(chan []byte, 100),
		errorChan:   make(chan error, 10),
		headers:     make(http.Header),
		dialer:      dialer,
	}, nil
}

// SetHeaders sets custom HTTP headers for the WebSocket handshake
func (c *WebSocketClient) SetHeaders(headers http.Header) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.headers = headers
}

// Connect establishes a WebSocket connection with retry logic
func (c *WebSocketClient) Connect(ctx context.Context) error {
	// Check if already connected
	c.mu.RLock()
	if c.isConnected && c.conn != nil {
		c.mu.RUnlock()
		return nil // Already connected
	}
	c.mu.RUnlock()

	operation := func() (*websocket.Conn, error) {
		conn, resp, err := c.dialer.DialContext(ctx, c.url, c.headers)
		if err != nil {
			if resp != nil {
				return nil, &WebSocketError{
					Code:    resp.StatusCode,
					Message: fmt.Sprintf("failed to connect: %v", err),
				}
			}
			return nil, fmt.Errorf("failed to connect to WebSocket %s: %w", c.url, err)
		}
		return conn, nil
	}

	conn, err := retry.Retry(ctx, operation, c.config.RetryConfig, c.logger)
	if err != nil {
		return fmt.Errorf("failed to establish WebSocket connection after retries: %w", err)
	}

	c.mu.Lock()
	// Close existing connection if any
	if c.conn != nil {
		c.conn.Close()
	}
	c.conn = conn
	c.isConnected = true
	c.lastMessage = time.Now()
	// Note: reconnectCount is managed by reconnectLoop, don't reset it here
	c.mu.Unlock()

	// Configure connection
	conn.SetReadLimit(c.config.MaxMessageSize)
	conn.SetReadDeadline(time.Now().Add(c.config.PongWait))
	conn.SetPongHandler(func(string) error {
		c.mu.Lock()
		c.lastMessage = time.Now()
		c.mu.Unlock()
		conn.SetReadDeadline(time.Now().Add(c.config.PongWait))
		return nil
	})
	conn.SetPingHandler(func(appData string) error {
		return conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(c.config.WriteDeadline))
	})

	c.logger.Infof("Successfully connected to WebSocket: %s", c.url)

	// Start background goroutines
	c.wg.Add(2)
	go c.pingLoop()
	go c.readLoop()

	return nil
}

// pingLoop sends periodic ping messages to keep the connection alive
func (c *WebSocketClient) pingLoop() {
	defer c.wg.Done()
	ticker := time.NewTicker(c.config.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.mu.RLock()
			conn := c.conn
			isConnected := c.isConnected
			c.mu.RUnlock()

			if !isConnected || conn == nil {
				continue
			}

			if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(c.config.WriteDeadline)); err != nil {
				c.logger.Warnf("Failed to send ping: %v", err)
				c.handleDisconnection()
				return
			}
		}
	}
}

// readLoop continuously reads messages from the WebSocket connection
func (c *WebSocketClient) readLoop() {
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		c.mu.RLock()
		conn := c.conn
		c.mu.RUnlock()

		if conn == nil {
			c.handleDisconnection()
			return
		}

		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Errorf("WebSocket connection closed unexpectedly: %v", err)
			} else {
				c.logger.Debugf("WebSocket read error: %v", err)
			}
			c.handleDisconnection()
			return
		}

		// Update last message time
		c.mu.Lock()
		c.lastMessage = time.Now()
		c.mu.Unlock()

		// Only forward text and binary messages
		if messageType == websocket.TextMessage || messageType == websocket.BinaryMessage {
			select {
			case c.messageChan <- message:
			case <-c.ctx.Done():
				return
			default:
				c.logger.Warnf("Message channel full, dropping message")
			}
		}
	}
}

// handleDisconnection handles disconnection and triggers reconnection
func (c *WebSocketClient) handleDisconnection() {
	c.mu.Lock()
	if !c.isConnected {
		c.mu.Unlock()
		return
	}
	c.isConnected = false
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	c.mu.Unlock()

	c.logger.Warnf("WebSocket disconnected, attempting to reconnect...")
	c.wg.Add(1)
	go c.reconnectLoop()
}

// reconnectLoop handles automatic reconnection with exponential backoff
func (c *WebSocketClient) reconnectLoop() {
	defer c.wg.Done()

	cfg := c.config.ReconnectConfig
	attempt := 0
	delay := cfg.BaseDelay

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		// Check max retries
		if cfg.MaxRetries > 0 && attempt >= cfg.MaxRetries {
			c.logger.Errorf("Max reconnection attempts (%d) reached for %s", cfg.MaxRetries, c.url)
			select {
			case c.errorChan <- fmt.Errorf("max reconnection attempts reached"):
			default:
			}
			return
		}

		attempt++
		c.mu.Lock()
		c.reconnectCount = attempt
		c.mu.Unlock()

		c.logger.Warnf("Reconnecting to %s in %v (attempt %d/%s)",
			c.url, delay, attempt, formatMaxRetries(cfg.MaxRetries))

		select {
		case <-c.ctx.Done():
			return
		case <-time.After(delay):
		}

		// Attempt to reconnect
		connectCtx, cancel := context.WithTimeout(c.ctx, c.config.HandshakeTimeout*2)
		err := c.Connect(connectCtx)
		cancel()

		if err != nil {
			c.logger.Errorf("Reconnection attempt %d failed: %v", attempt, err)
			// Calculate next delay with exponential backoff
			delay = calculateReconnectDelay(delay, cfg)
			continue
		}

		c.logger.Infof("Reconnection successful after %d attempts", attempt)
		// Reset reconnect count after successful reconnection
		c.mu.Lock()
		c.reconnectCount = 0
		c.mu.Unlock()
		return
	}
}

// calculateReconnectDelay calculates the delay for the next reconnection attempt
func calculateReconnectDelay(currentDelay time.Duration, cfg *ReconnectConfig) time.Duration {
	delay := time.Duration(float64(currentDelay) * cfg.BackoffFactor)
	if delay > cfg.MaxDelay {
		delay = cfg.MaxDelay
	}

	if cfg.Jitter {
		// Add ±25% jitter
		jitterRange := float64(delay) * 0.25
		jitter := time.Duration(float64(time.Now().UnixNano()%int64(jitterRange*2)) - float64(jitterRange))
		delay += jitter
		if delay < cfg.BaseDelay {
			delay = cfg.BaseDelay
		}
	}

	return delay
}

// formatMaxRetries formats max retries for logging
func formatMaxRetries(maxRetries int) string {
	if maxRetries == 0 {
		return "∞"
	}
	return fmt.Sprintf("%d", maxRetries)
}

// ReadMessage reads a message from the WebSocket connection
func (c *WebSocketClient) ReadMessage(ctx context.Context) ([]byte, error) {
	select {
	case <-c.ctx.Done():
		return nil, c.ctx.Err()
	case <-ctx.Done():
		return nil, ctx.Err()
	case message := <-c.messageChan:
		return message, nil
	case err := <-c.errorChan:
		return nil, err
	}
}

// WriteMessage writes a message to the WebSocket connection
func (c *WebSocketClient) WriteMessage(ctx context.Context, messageType int, data []byte) error {
	c.mu.RLock()
	conn := c.conn
	isConnected := c.isConnected
	c.mu.RUnlock()

	if !isConnected || conn == nil {
		return fmt.Errorf("WebSocket not connected")
	}

	conn.SetWriteDeadline(time.Now().Add(c.config.WriteDeadline))
	return conn.WriteMessage(messageType, data)
}

// WriteTextMessage writes a text message to the WebSocket connection
func (c *WebSocketClient) WriteTextMessage(ctx context.Context, data []byte) error {
	return c.WriteMessage(ctx, websocket.TextMessage, data)
}

// WriteBinaryMessage writes a binary message to the WebSocket connection
func (c *WebSocketClient) WriteBinaryMessage(ctx context.Context, data []byte) error {
	return c.WriteMessage(ctx, websocket.BinaryMessage, data)
}

// IsConnected returns whether the WebSocket is currently connected
func (c *WebSocketClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isConnected
}

// GetReconnectCount returns the number of reconnection attempts
func (c *WebSocketClient) GetReconnectCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.reconnectCount
}

// GetLastMessageTime returns the time of the last received message
func (c *WebSocketClient) GetLastMessageTime() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastMessage
}

// MessageChannel returns the channel for receiving messages
func (c *WebSocketClient) MessageChannel() <-chan []byte {
	return c.messageChan
}

// ErrorChannel returns the channel for receiving errors
func (c *WebSocketClient) ErrorChannel() <-chan error {
	return c.errorChan
}

// Close gracefully closes the WebSocket connection
func (c *WebSocketClient) Close() error {
	c.cancel()
	c.wg.Wait()

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil // Already closed
	}

	var err error
	if c.conn != nil {
		// Send close message
		closeMsg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
		c.conn.WriteControl(websocket.CloseMessage, closeMsg, time.Now().Add(c.config.WriteDeadline))
		err = c.conn.Close()
		c.conn = nil
	}

	c.isConnected = false
	c.closed = true
	close(c.messageChan)
	close(c.errorChan)
	return err
}

// GetConn returns the underlying websocket.Conn (use with caution)
func (c *WebSocketClient) GetConn() *websocket.Conn {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn
}
