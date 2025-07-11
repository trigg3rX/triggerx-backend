package websocket

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gorilla/websocket"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/logging"
)

// ChainConfig represents configuration for a specific blockchain
type ChainConfig struct {
	ChainID      string
	Name         string
	RPCURL       string
	WebSocketURL string
	Contracts    []ContractConfig
}

// ContractConfig represents a contract to monitor
type ContractConfig struct {
	Address      string
	ContractType ContractType
	ABI          string
	Events       []string // Event names to monitor
}

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
		eventChan: make(chan *ChainEvent, 10000), // Large buffer for high-throughput
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

	// Create chain connection
	chainConn := &ChainConnection{
		chainID:       config.ChainID,
		chainName:     config.Name,
		ethClient:     ethClient,
		subscriptions: make(map[string]*Subscription),
		subManager:    subManager,
		reconnectMgr:  NewReconnectManager(config.WebSocketURL, c.logger),
		eventChan:     c.eventChan,
		logger:        c.logger,
		isConnected:   false,
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

	c.logger.Infof("Added chain %s (%s) for monitoring", config.Name, config.ChainID)
	return nil
}

// Start begins monitoring all configured chains
func (c *Client) Start() error {
	c.logger.Info("Starting multi-chain WebSocket client")

	for chainID, chainConn := range c.chains {
		c.wg.Add(1)
		go c.startChainConnection(chainID, chainConn)
	}

	c.logger.Infof("Started monitoring %d chains", len(c.chains))
	return nil
}

// Stop gracefully stops all chain connections
func (c *Client) Stop() error {
	c.logger.Info("Stopping multi-chain WebSocket client")

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
	c.logger.Info("Multi-chain WebSocket client stopped")
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

	// Start the reconnect manager
	go chainConn.reconnectMgr.Start(c.ctx, func() error {
		return chainConn.connect()
	})

	// Start processing events from the subscription manager
	go chainConn.processEvents(c.ctx)

	// Keep the connection alive
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-time.After(30 * time.Second):
			if chainConn.isConnected {
				if err := chainConn.ping(); err != nil {
					c.logger.Warnf("Ping failed for chain %s: %v", chainID, err)
					chainConn.markDisconnected()
				}
			}
		}
	}
}

// processEvents processes events from the subscription manager
func (cc *ChainConnection) processEvents(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Process WebSocket messages and events
			// This is a placeholder - you'll need to implement the actual WebSocket message processing
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// connect establishes WebSocket connection for the chain
func (cc *ChainConnection) connect() error {
	cc.logger.Infof("Connecting to chain %s WebSocket", cc.chainName)

	// Connect to WebSocket (implementation depends on your WebSocket library)
	// This is a placeholder - you'll need to implement based on your WebSocket client

	cc.mu.Lock()
	cc.isConnected = true
	cc.lastMessage = time.Now()
	cc.mu.Unlock()

	// Re-establish all subscriptions
	return cc.reestablishSubscriptions()
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
		cc.logger.Infof("Subscribed to %s.%s events from %s on chain %s", contractType, eventName, contractAddr, cc.chainID)
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
		cc.logger.Infof("Subscribed to %s events from %s on chain %s", eventName, contractAddr, cc.chainID)
	}

	return nil
}

// reestablishSubscriptions re-creates all subscriptions after reconnection
func (cc *ChainConnection) reestablishSubscriptions() error {
	cc.mu.RLock()
	subscriptions := make([]*Subscription, 0, len(cc.subscriptions))
	for _, sub := range cc.subscriptions {
		subscriptions = append(subscriptions, sub)
	}
	cc.mu.RUnlock()

	for _, sub := range subscriptions {
		// Re-establish subscription (implementation specific)
		cc.logger.Debugf("Re-establishing subscription %s", sub.ID)
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
	// Implementation specific ping
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
		cc.wsConn.Close()
	}

	if cc.ethClient != nil {
		cc.ethClient.Close()
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
