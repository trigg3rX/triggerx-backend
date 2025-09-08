package websocket

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gorilla/websocket"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// Client manages WebSocket connections to multiple blockchains
type Client struct {
	logger    logging.Logger
	chains    map[string]*ChainConnection
	eventChan chan *ChainEvent
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

// ChainConnection represents a WebSocket connection to a specific blockchain
type ChainConnection struct {
	chainID       string
	chainName     string
	ethClient     *ethclient.Client
	wsConn        *websocket.Conn
	subscriptions map[string]*Subscription
	subManager    *SubscriptionManager
	reconnectMgr  *ReconnectManager
	eventChan     chan *ChainEvent
	logger        logging.Logger
	mu            sync.RWMutex
	isConnected   bool
	lastMessage   time.Time
	websocketURL  string
}

// Subscription represents an event subscription
type Subscription struct {
	ID           string
	ChainID      string
	ContractAddr common.Address
	ContractType ContractType
	EventName    string
	Query        ethereum.FilterQuery
	Active       bool
	CreatedAt    time.Time
}

// ChainEvent represents an event from any blockchain
type ChainEvent struct {
	ChainID      string       `json:"chain_id"`
	ChainName    string       `json:"chain_name"`
	ContractAddr string       `json:"contract_address"`
	ContractType ContractType `json:"contract_type"`
	EventName    string       `json:"event_name"`
	BlockNumber  uint64       `json:"block_number"`
	TxHash       string       `json:"tx_hash"`
	LogIndex     uint         `json:"log_index"`
	Data         interface{}  `json:"data"`
	RawLog       types.Log    `json:"raw_log"`
	ProcessedAt  time.Time    `json:"processed_at"`
}

// NewClient creates a new multi-chain WebSocket client
func NewClient(logger logging.Logger) *Client {
	ctx, cancel := context.WithCancel(context.Background())

	return &Client{
		logger:    logger,
		chains:    make(map[string]*ChainConnection),
		eventChan: make(chan *ChainEvent, 10000),
		ctx:       ctx,
		cancel:    cancel,
	}
}

// AddChain adds a new blockchain to monitor
func (c *Client) AddChain(config ChainConfig) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.chains[config.ChainID]; exists {
		return fmt.Errorf("chain %s already exists", config.ChainID)
	}

	// Create Ethereum client
	ethClient, err := ethclient.Dial(config.RPCURL)
	if err != nil {
		return fmt.Errorf("failed to connect to %s RPC: %w", config.Name, err)
	}

	// Create subscription manager
	subManager := NewSubscriptionManager(config.ChainID, c.logger)

	// Create reconnection manager
	reconnectMgr := NewReconnectManagerWithConfig(config.WebSocketURL, config.Reconnect, c.logger)

	// Create chain connection
	chainConn := &ChainConnection{
		chainID:       config.ChainID,
		chainName:     config.Name,
		ethClient:     ethClient,
		subscriptions: make(map[string]*Subscription),
		subManager:    subManager,
		reconnectMgr:  reconnectMgr,
		eventChan:     c.eventChan,
		logger:        c.logger,
		isConnected:   false,
		websocketURL:  config.WebSocketURL,
	}

	c.chains[config.ChainID] = chainConn

	// Auto-subscribe to contracts if specified in config
	if len(config.Contracts) > 0 {
		for _, contractConfig := range config.Contracts {
			if err := chainConn.subscribeToContractEvents(contractConfig); err != nil {
				c.logger.Warnf("Failed to subscribe to contract %s: %v", contractConfig.Address, err)
			}
		}
	}

	// c.logger.Infof("Added chain %s (%s) for monitoring", config.Name, config.ChainID)
	return nil
}

// Start begins monitoring all configured chains
func (c *Client) Start() error {
	// c.logger.Info("Starting multi-chain WebSocket client")

	for chainID, chainConn := range c.chains {
		c.wg.Add(1)
		go c.startChainConnection(chainID, chainConn)
	}

	// c.logger.Infof("Started monitoring %d chains", len(c.chains))
	return nil
}

// Stop gracefully stops all chain connections
func (c *Client) Stop() error {
	// c.logger.Info("Stopping multi-chain WebSocket client")

	c.cancel()
	c.wg.Wait()

	// Close all connections
	c.mu.Lock()
	for chainID, chainConn := range c.chains {
		if err := chainConn.close(); err != nil {
			c.logger.Errorf("Error closing connection for chain %s: %v", chainID, err)
		}
	}
	c.mu.Unlock()

	close(c.eventChan)
	// c.logger.Info("Multi-chain WebSocket client stopped")
	return nil
}

// SubscribeToContract subscribes to events from a specific contract using contract type
func (c *Client) SubscribeToContract(chainID string, contractAddr string, contractType ContractType, events []string) error {
	c.mu.RLock()
	chainConn, exists := c.chains[chainID]
	c.mu.RUnlock()

	if !exists {
		return fmt.Errorf("chain %s not found", chainID)
	}

	return chainConn.subscribeToContractWithType(contractAddr, contractType, events)
}

// SubscribeToContractLegacy subscribes to events from a specific contract (legacy method)
func (c *Client) SubscribeToContractLegacy(chainID string, contractAddr string, events []string) error {
	c.mu.RLock()
	chainConn, exists := c.chains[chainID]
	c.mu.RUnlock()

	if !exists {
		return fmt.Errorf("chain %s not found", chainID)
	}

	return chainConn.subscribeToContract(contractAddr, events)
}

// EventChannel returns the channel for receiving events from all chains
func (c *Client) EventChannel() <-chan *ChainEvent {
	return c.eventChan
}

// GetChainStatus returns the status of all chains
func (c *Client) GetChainStatus() map[string]ChainStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()

	status := make(map[string]ChainStatus)
	for chainID, chainConn := range c.chains {
		status[chainID] = chainConn.getStatus()
	}
	return status
}

// ChainStatus represents the status of a chain connection
type ChainStatus struct {
	ChainID        string    `json:"chain_id"`
	ChainName      string    `json:"chain_name"`
	Connected      bool      `json:"connected"`
	LastMessage    time.Time `json:"last_message"`
	Subscriptions  int       `json:"subscriptions"`
	ReconnectCount int       `json:"reconnect_count"`
	LatestBlock    uint64    `json:"latest_block,omitempty"`
}

// startChainConnection starts monitoring a specific chain
func (c *Client) startChainConnection(chainID string, chainConn *ChainConnection) {
	defer c.wg.Done()

	c.logger.Infof("Starting connection for chain %s", chainID)

	// Start the reconnection manager with the connect function
	chainConn.reconnectMgr.Start(c.ctx, chainConn.connect)

	// Start processing events from the subscription manager
	go chainConn.processEvents(c.ctx)

	// Keep the connection alive
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-time.After(30 * time.Second):
			chainConn.mu.RLock()
			isConnected := chainConn.isConnected
			chainConn.mu.RUnlock()

			if isConnected {
				if err := chainConn.ping(); err != nil {
					c.logger.Warnf("Ping failed for chain %s: %v", chainID, err)
					chainConn.markDisconnected()
					// Trigger reconnection
					chainConn.reconnectMgr.TriggerReconnect(c.ctx, chainConn.connect)
				}
			}
		}
	}
}

// processEvents processes events from the WebSocket connection
func (cc *ChainConnection) processEvents(ctx context.Context) {
	cc.logger.Infof("Starting event processing for chain %s", cc.chainName)

	// Start ping ticker to keep connection alive
	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			cc.logger.Infof("Stopping event processing for chain %s", cc.chainName)
			return
		case <-pingTicker.C:
			// Send ping to keep connection alive
			if err := cc.ping(); err != nil {
				cc.logger.Errorf("Failed to ping WebSocket for chain %s: %v", cc.chainName, err)
				cc.markDisconnected()
				// Trigger reconnection
				cc.reconnectMgr.TriggerReconnect(ctx, cc.connect)
				return
			}
		default:
			// Read WebSocket messages
			if err := cc.readMessage(); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					cc.logger.Errorf("WebSocket connection closed unexpectedly for chain %s: %v", cc.chainName, err)
				} else {
					cc.logger.Debugf("WebSocket read error for chain %s: %v", cc.chainName, err)
				}
				cc.markDisconnected()
				// Trigger reconnection
				cc.reconnectMgr.TriggerReconnect(ctx, cc.connect)
				return
			}
		}
	}
}

// readMessage reads and processes a single WebSocket message
func (cc *ChainConnection) readMessage() error {
	cc.mu.RLock()
	conn := cc.wsConn
	cc.mu.RUnlock()

	if conn == nil {
		return fmt.Errorf("WebSocket connection not established")
	}

	// Set read deadline
	err := conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	if err != nil {
		cc.logger.Errorf("Failed to set read deadline: %v", err)
	}

	// Read message
	_, message, err := conn.ReadMessage()
	if err != nil {
		return err
	}

	// Update last message time
	cc.mu.Lock()
	cc.lastMessage = time.Now()
	cc.mu.Unlock()

	// Process the message using subscription manager
	if err := cc.subManager.ProcessWebSocketMessage(message, cc.eventChan); err != nil {
		cc.logger.Errorf("Failed to process WebSocket message for chain %s: %v", cc.chainName, err)
		// Don't return error here as we want to continue processing other messages
	}

	return nil
}

// connect establishes WebSocket connection for the chain
func (cc *ChainConnection) connect() error {
	cc.logger.Infof("Connecting to chain %s WebSocket at %s", cc.chainName, cc.websocketURL)

	if cc.websocketURL == "" {
		return fmt.Errorf("WebSocket URL not configured for chain %s", cc.chainName)
	}

	// Create dialer with timeout and other options
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
		Proxy:            http.ProxyFromEnvironment,
	}

	// Connect to WebSocket
	conn, resp, err := dialer.Dial(cc.websocketURL, nil)
	if err != nil {
		if resp != nil {
			return fmt.Errorf("failed to connect to WebSocket %s (status: %d): %w", cc.websocketURL, resp.StatusCode, err)
		}
		return fmt.Errorf("failed to connect to WebSocket %s: %w", cc.websocketURL, err)
	}

	// Set connection parameters for reliability
	conn.SetReadLimit(512 * 1024) // 512KB max message size
	err = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	if err != nil {
		cc.logger.Errorf("Failed to set read deadline: %v", err)
	}
	conn.SetPongHandler(func(string) error {
		err = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		if err != nil {
			cc.logger.Errorf("Failed to set read deadline: %v", err)
		}
		return nil
	})

	// Set ping handler to keep connection alive
	conn.SetPingHandler(func(appData string) error {
		// cc.logger.Debugf("Received ping from %s", cc.chainName)
		return conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(30*time.Second))
	})

	cc.mu.Lock()
	cc.wsConn = conn
	cc.isConnected = true
	cc.lastMessage = time.Now()
	cc.mu.Unlock()

	cc.logger.Infof("Successfully connected to chain %s WebSocket", cc.chainName)

	// Send subscription message for all active subscriptions
	if err := cc.sendSubscription(); err != nil {
		cc.logger.Errorf("Failed to send subscription for chain %s: %v", cc.chainName, err)
		// Don't fail the connection for subscription errors
	}

	return nil
}

// sendSubscription sends the WebSocket subscription message for all active subscriptions
func (cc *ChainConnection) sendSubscription() error {
	cc.mu.RLock()
	conn := cc.wsConn
	cc.mu.RUnlock()

	if conn == nil {
		return fmt.Errorf("WebSocket connection not established")
	}

	// Build subscription message
	subscriptionMsg, err := cc.subManager.BuildWebSocketSubscription()
	if err != nil {
		return fmt.Errorf("failed to build subscription message: %w", err)
	}

	// Send subscription message
	err = conn.WriteMessage(websocket.TextMessage, subscriptionMsg)
	if err != nil {
		return fmt.Errorf("failed to send subscription message: %w", err)
	}

	cc.logger.Infof("Sent subscription message for chain %s", cc.chainName)
	return nil
}

// subscribeToContractEvents subscribes to events from a contract using ContractConfig
func (cc *ChainConnection) subscribeToContractEvents(contractConfig ContractConfig) error {
	if contractConfig.ContractType != "" {
		return cc.subscribeToContractWithType(contractConfig.Address, contractConfig.ContractType, contractConfig.Events)
	}
	return cc.subscribeToContract(contractConfig.Address, contractConfig.Events)
}

// subscribeToContractWithType subscribes to events from a contract using contract type
func (cc *ChainConnection) subscribeToContractWithType(contractAddr string, contractType ContractType, events []string) error {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	for _, eventName := range events {
		// Add subscription using the subscription manager
		_, err := cc.subManager.AddContractSubscription(contractAddr, contractType, eventName)
		if err != nil {
			cc.logger.Errorf("Failed to add contract subscription for %s.%s: %v", contractType, eventName, err)
			continue
		}

		// Also add to local subscriptions for tracking
		subID := fmt.Sprintf("%s_%s_%s_%s", cc.chainID, contractAddr, contractType, eventName)
		addr := common.HexToAddress(contractAddr)

		subscription := &Subscription{
			ID:           subID,
			ChainID:      cc.chainID,
			ContractAddr: addr,
			ContractType: contractType,
			EventName:    eventName,
			Active:       true,
			CreatedAt:    time.Now(),
		}

		cc.subscriptions[subID] = subscription
		// cc.logger.Infof("Subscribed to %s.%s events from %s on chain %s", contractType, eventName, contractAddr, cc.chainID)
	}

	return nil
}

// subscribeToContract subscribes to events from a contract (legacy method)
func (cc *ChainConnection) subscribeToContract(contractAddr string, events []string) error {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	addr := common.HexToAddress(contractAddr)

	for _, eventName := range events {
		subID := fmt.Sprintf("%s_%s_%s", cc.chainID, contractAddr, eventName)

		query := ethereum.FilterQuery{
			Addresses: []common.Address{addr},
			// Add event signature filtering here
		}

		subscription := &Subscription{
			ID:           subID,
			ChainID:      cc.chainID,
			ContractAddr: addr,
			EventName:    eventName,
			Query:        query,
			Active:       true,
			CreatedAt:    time.Now(),
		}

		cc.subscriptions[subID] = subscription
		// cc.logger.Infof("Subscribed to %s events from %s on chain %s", eventName, contractAddr, cc.chainID)
	}

	return nil
}

// getStatus returns the current status of the chain connection
func (cc *ChainConnection) getStatus() ChainStatus {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	var latestBlock uint64
	if cc.ethClient != nil {
		if block, err := cc.ethClient.BlockNumber(context.Background()); err == nil {
			latestBlock = block
		}
	}

	return ChainStatus{
		ChainID:        cc.chainID,
		ChainName:      cc.chainName,
		Connected:      cc.isConnected,
		LastMessage:    cc.lastMessage,
		Subscriptions:  len(cc.subscriptions),
		ReconnectCount: cc.reconnectMgr.GetReconnectCount(),
		LatestBlock:    latestBlock,
	}
}

// ping sends a ping to keep the connection alive
func (cc *ChainConnection) ping() error {
	cc.mu.RLock()
	conn := cc.wsConn
	cc.mu.RUnlock()

	if conn == nil {
		return fmt.Errorf("WebSocket connection not established")
	}

	// Send ping with current timestamp as data
	pingData := []byte(fmt.Sprintf("ping_%d", time.Now().Unix()))
	err := conn.WriteControl(websocket.PingMessage, pingData, time.Now().Add(10*time.Second))
	if err != nil {
		return fmt.Errorf("failed to send ping: %w", err)
	}

	// cc.logger.Debugf("Sent ping to chain %s", cc.chainName)
	return nil
}

// markDisconnected marks the connection as disconnected
func (cc *ChainConnection) markDisconnected() {
	cc.mu.Lock()
	cc.isConnected = false
	cc.mu.Unlock()
}

// close closes the chain connection
func (cc *ChainConnection) close() error {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.isConnected = false

	if cc.wsConn != nil {
		err := cc.wsConn.Close()
		if err != nil {
			cc.logger.Errorf("Failed to close WebSocket connection: %v", err)
		}
		cc.wsConn = nil
	}

	if cc.ethClient != nil {
		cc.ethClient.Close()
	}

	// Reset reconnection count when closing
	if cc.reconnectMgr != nil {
		cc.reconnectMgr.resetReconnectCount()
	}

	return nil
}

// GetSubscriptionStats returns statistics for all subscriptions on a chain
func (c *Client) GetSubscriptionStats(chainID string) (map[string]interface{}, error) {
	c.mu.RLock()
	chainConn, exists := c.chains[chainID]
	c.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("chain %s not found", chainID)
	}

	return chainConn.subManager.GetSubscriptionStats(), nil
}

// GetAllSubscriptionStats returns statistics for all chains
func (c *Client) GetAllSubscriptionStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := make(map[string]interface{})
	for chainID, chainConn := range c.chains {
		stats[chainID] = chainConn.subManager.GetSubscriptionStats()
	}

	return stats
}

// TriggerReconnect manually triggers reconnection for a specific chain
func (c *Client) TriggerReconnect(chainID string) error {
	c.mu.RLock()
	chainConn, exists := c.chains[chainID]
	c.mu.RUnlock()

	if !exists {
		return fmt.Errorf("chain %s not found", chainID)
	}

	chainConn.logger.Infof("Manual reconnection triggered for chain %s", chainID)
	chainConn.reconnectMgr.TriggerReconnect(c.ctx, chainConn.connect)
	return nil
}

// GetReconnectStats returns reconnection statistics for all chains
func (c *Client) GetReconnectStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := make(map[string]interface{})
	for chainID, chainConn := range c.chains {
		stats[chainID] = map[string]interface{}{
			"reconnect_count": chainConn.reconnectMgr.GetReconnectCount(),
			"is_running":      chainConn.reconnectMgr.IsRunning(),
			"config":          chainConn.reconnectMgr.GetConfig(),
		}
	}

	return stats
}
