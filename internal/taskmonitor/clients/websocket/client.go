package websocket

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"encoding/json"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	nodeclient "github.com/trigg3rX/triggerx-backend/pkg/client/nodeclient"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	wsclient "github.com/trigg3rX/triggerx-backend/pkg/websocket"
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
	chainID           string
	chainName         string
	nodeClient        *nodeclient.NodeClient
	subscriptions     map[string]*Subscription
	nodeSubscriptions map[string]*NodeSubscription // Maps our subscription ID to nodeclient subscription info
	subManager        *SubscriptionManager
	eventChan         chan *ChainEvent
	logger            logging.Logger
	mu                sync.RWMutex
	websocketURL      string
	rpcURL            string
	wg                sync.WaitGroup
}

// NodeSubscription holds information about a nodeclient subscription
type NodeSubscription struct {
	NodeSubID        string
	NotificationChan <-chan *nodeclient.SubscriptionNotification
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

	// Create node client config with WebSocket support
	nodeCfg := createNodeClientConfig(config, c.logger)

	nodeClient, err := nodeclient.NewNodeClient(nodeCfg)
	if err != nil {
		return fmt.Errorf("failed to create node client for %s: %w", config.Name, err)
	}

	// Create subscription manager
	subManager := NewSubscriptionManager(config.ChainID, c.logger)

	// Create chain connection
	chainConn := &ChainConnection{
		chainID:           config.ChainID,
		chainName:         config.Name,
		nodeClient:        nodeClient,
		subscriptions:     make(map[string]*Subscription),
		nodeSubscriptions: make(map[string]*NodeSubscription),
		subManager:        subManager,
		eventChan:         c.eventChan,
		logger:            c.logger,
		websocketURL:      config.WebSocketURL,
		rpcURL:            config.RPCURL,
		wg:                sync.WaitGroup{},
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

	return nil
}

// Start begins monitoring all configured chains
func (c *Client) Start() error {
	for chainID, chainConn := range c.chains {
		c.wg.Add(1)
		go c.startChainConnection(chainID, chainConn)
	}

	return nil
}

// Stop gracefully stops all chain connections
func (c *Client) Stop() error {
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

	// Connect to WebSocket via nodeclient
	if err := chainConn.connect(c.ctx); err != nil {
		c.logger.Errorf("Failed to connect to chain %s: %v", chainID, err)
		return
	}

	// Start processing events from nodeclient subscriptions
	go chainConn.processEvents(c.ctx)

	// Wait for context cancellation
	<-c.ctx.Done()
}

// processEvents waits for context cancellation (subscriptions are handled individually)
func (cc *ChainConnection) processEvents(ctx context.Context) {
	cc.logger.Infof("Starting event processing for chain %s", cc.chainName)
	<-ctx.Done()
	cc.wg.Wait()
	cc.logger.Infof("Stopping event processing for chain %s", cc.chainName)
}

// processSubscriptionNotifications processes notifications for a specific subscription
func (cc *ChainConnection) processSubscriptionNotifications(ctx context.Context, subID string, notifChan <-chan *nodeclient.SubscriptionNotification) {
	defer cc.wg.Done()

	cc.subManager.mu.RLock()
	sub, exists := cc.subManager.subscriptions[subID]
	cc.subManager.mu.RUnlock()

	if !exists {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case notif, ok := <-notifChan:
			if !ok {
				cc.logger.Infof("Notification channel closed for subscription %s", subID)
				return
			}
			// Process the notification
			if err := cc.processNotification(sub, notif); err != nil {
				cc.logger.Errorf("Failed to process notification for subscription %s: %v", subID, err)
			}
		}
	}
}

// processNotification processes a single subscription notification
func (cc *ChainConnection) processNotification(sub *EventSubscription, notif *nodeclient.SubscriptionNotification) error {
	// Parse the log from the notification result
	var log types.Log
	if err := json.Unmarshal(notif.Result, &log); err != nil {
		return fmt.Errorf("failed to unmarshal log: %w", err)
	}

	// Process the log entry using subscription manager
	return cc.subManager.processLogEntry(log, cc.eventChan)
}

// connect establishes WebSocket connection for the chain via nodeclient
func (cc *ChainConnection) connect(ctx context.Context) error {
	cc.logger.Infof("Connecting to chain %s WebSocket at %s", cc.chainName, cc.websocketURL)

	if cc.websocketURL == "" {
		return fmt.Errorf("WebSocket URL not configured for chain %s", cc.chainName)
	}

	// Connect WebSocket via nodeclient (it handles connection internally)
	if err := cc.nodeClient.ConnectWebSocket(ctx); err != nil {
		return fmt.Errorf("failed to connect WebSocket: %w", err)
	}

	cc.logger.Infof("Successfully connected to chain %s WebSocket", cc.chainName)

	// Subscribe to all active subscriptions using nodeclient
	if err := cc.subscribeAll(ctx); err != nil {
		cc.logger.Errorf("Failed to subscribe for chain %s: %v", cc.chainName, err)
		// Don't fail the connection for subscription errors
	}

	return nil
}

// subscribeAll subscribes to all active subscriptions using nodeclient
func (cc *ChainConnection) subscribeAll(ctx context.Context) error {
	cc.subManager.mu.RLock()
	activeSubs := make([]*EventSubscription, 0)
	for _, sub := range cc.subManager.subscriptions {
		if sub.Active {
			activeSubs = append(activeSubs, sub)
		}
	}
	cc.subManager.mu.RUnlock()

	for _, sub := range activeSubs {
		if err := cc.subscribeToEvent(ctx, sub); err != nil {
			cc.logger.Errorf("Failed to subscribe to %s.%s: %v", sub.ContractType, sub.EventName, err)
			continue
		}
	}

	return nil
}

// subscribeToEvent subscribes to a specific event using nodeclient.EthSubscribe
func (cc *ChainConnection) subscribeToEvent(ctx context.Context, sub *EventSubscription) error {
	// Build filter params for eth_subscribe
	filterParams := map[string]interface{}{
		"address": []string{sub.ContractAddr.Hex()},
		"topics":  [][]string{{sub.EventSig.Hex()}},
	}

	// Subscribe using nodeclient
	nodeSubID, notifChan, err := cc.nodeClient.EthSubscribe(ctx, "logs", filterParams)
	if err != nil {
		return fmt.Errorf("failed to subscribe via nodeclient: %w", err)
	}

	cc.mu.Lock()
	cc.nodeSubscriptions[sub.ID] = &NodeSubscription{
		NodeSubID:        nodeSubID,
		NotificationChan: notifChan,
	}
	cc.mu.Unlock()

	// Start processing notifications for this subscription
	cc.wg.Add(1)
	go cc.processSubscriptionNotifications(ctx, sub.ID, notifChan)

	cc.logger.Infof("Subscribed to %s.%s with nodeclient subscription ID: %s", sub.ContractType, sub.EventName, nodeSubID)
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
		sub, err := cc.subManager.AddContractSubscription(contractAddr, contractType, eventName)
		if err != nil {
			cc.logger.Errorf("Failed to add contract subscription for %s.%s: %v", contractType, eventName, err)
			continue
		}

		// Also add to local subscriptions for tracking
		addr := common.HexToAddress(contractAddr)
		subscription := &Subscription{
			ID:           sub.ID,
			ChainID:      cc.chainID,
			ContractAddr: addr,
			ContractType: contractType,
			EventName:    eventName,
			Active:       true,
			CreatedAt:    time.Now(),
		}

		cc.subscriptions[sub.ID] = subscription

		// Subscribe via nodeclient if WebSocket is connected
		if cc.nodeClient != nil && cc.nodeClient.IsWebSocketConnected() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if err := cc.subscribeToEvent(ctx, sub); err != nil {
				cc.logger.Errorf("Failed to subscribe via nodeclient for %s.%s: %v", contractType, eventName, err)
			}
		}
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
	}

	return nil
}

// getStatus returns the current status of the chain connection
func (cc *ChainConnection) getStatus() ChainStatus {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	var latestBlock uint64
	if cc.nodeClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if blockHex, err := cc.nodeClient.EthBlockNumber(ctx); err == nil {
			if block, err := hexToUint64(blockHex); err == nil {
				latestBlock = block
			}
		}
	}

	var isConnected bool
	if cc.nodeClient != nil {
		isConnected = cc.nodeClient.IsWebSocketConnected()
	}

	return ChainStatus{
		ChainID:        cc.chainID,
		ChainName:      cc.chainName,
		Connected:      isConnected,
		LastMessage:    time.Time{}, // Not tracked separately anymore
		Subscriptions:  len(cc.subscriptions),
		ReconnectCount: 0, // Managed internally by nodeclient
		LatestBlock:    latestBlock,
	}
}

// hexToUint64 converts a hex string (with or without 0x prefix) to uint64
func hexToUint64(hexStr string) (uint64, error) {
	// Remove 0x prefix if present
	if len(hexStr) >= 2 && hexStr[:2] == "0x" {
		hexStr = hexStr[2:]
	}
	return strconv.ParseUint(hexStr, 16, 64)
}

// close closes the chain connection
func (cc *ChainConnection) close() error {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	// Unsubscribe from all nodeclient subscriptions
	for subID, nodeSub := range cc.nodeSubscriptions {
		if cc.nodeClient != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			if err := cc.nodeClient.EthUnsubscribe(ctx, nodeSub.NodeSubID); err != nil {
				cc.logger.Errorf("Failed to unsubscribe %s: %v", subID, err)
			}
			cancel()
		}
	}

	// Wait for all notification processors to finish
	cc.wg.Wait()

	if cc.nodeClient != nil {
		cc.nodeClient.Close()
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

	// Disconnect and reconnect WebSocket via nodeclient
	if chainConn.nodeClient != nil {
		_ = chainConn.nodeClient.DisconnectWebSocket()
	}

	// Reconnect
	return chainConn.connect(c.ctx)
}

// GetReconnectStats returns reconnection statistics for all chains
func (c *Client) GetReconnectStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := make(map[string]interface{})
	for chainID, chainConn := range c.chains {
		chainConn.mu.RLock()
		var isConnected bool
		if chainConn.nodeClient != nil {
			isConnected = chainConn.nodeClient.IsWebSocketConnected()
		}
		chainConn.mu.RUnlock()

		stats[chainID] = map[string]interface{}{
			"is_connected":  isConnected,
			"subscriptions": len(chainConn.nodeSubscriptions),
		}
	}

	return stats
}

// createNodeClientConfig creates a nodeclient.Config from ChainConfig
func createNodeClientConfig(config ChainConfig, logger logging.Logger) *nodeclient.Config {
	// Extract API key from RPC URL if it's an Alchemy/Blast URL
	apiKey := extractAPIKeyFromURL(config.RPCURL)

	nodeCfg := nodeclient.DefaultConfig(apiKey, "", logger)
	nodeCfg.BaseURL = config.RPCURL
	nodeCfg.WebSocketURL = config.WebSocketURL
	nodeCfg.RequestTimeout = 30 * time.Second

	// Create WebSocket config from ReconnectConfig
	wsConfig := wsclient.DefaultWebSocketRetryConfig()
	if config.Reconnect.MaxRetries > 0 {
		wsConfig.ReconnectConfig.MaxRetries = config.Reconnect.MaxRetries
	}
	if config.Reconnect.BaseDelay > 0 {
		wsConfig.ReconnectConfig.BaseDelay = config.Reconnect.BaseDelay
	}
	if config.Reconnect.MaxDelay > 0 {
		wsConfig.ReconnectConfig.MaxDelay = config.Reconnect.MaxDelay
	}
	if config.Reconnect.BackoffFactor > 0 {
		wsConfig.ReconnectConfig.BackoffFactor = config.Reconnect.BackoffFactor
	}
	wsConfig.ReconnectConfig.Jitter = config.Reconnect.Jitter
	nodeCfg.WebSocketConfig = wsConfig

	return nodeCfg
}

// extractAPIKeyFromURL extracts API key from RPC URL
func extractAPIKeyFromURL(url string) string {
	// For Alchemy/Blast URLs, the API key is typically at the end
	// Format: https://base-mainnet.g.alchemy.com/v2/API_KEY
	// or: https://base-mainnet.blastapi.io/API_KEY
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}
