package websocket

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// SubscriptionManager manages WebSocket event subscriptions for a chain
type SubscriptionManager struct {
	chainID       string
	subscriptions map[string]*EventSubscription
	eventFilters  map[string][]common.Hash // event name -> topic hashes
	logger        logging.Logger
	mu            sync.RWMutex
}

// EventSubscription represents a single event subscription
type EventSubscription struct {
	ID           string
	ChainID      string
	ContractAddr common.Address
	EventName    string
	EventSig     common.Hash
	FilterQuery  ethereum.FilterQuery
	Active       bool
	CreatedAt    time.Time
	LastEvent    time.Time
	EventCount   uint64
}

// ContractABI represents the ABI for a contract
type ContractABI struct {
	Address string           `json:"address"`
	Events  map[string]Event `json:"events"`
}

// Event represents an ABI event definition
type Event struct {
	Name      string       `json:"name"`
	Signature string       `json:"signature"`
	Inputs    []EventInput `json:"inputs"`
}

// EventInput represents an event input parameter
type EventInput struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Indexed bool   `json:"indexed"`
}

// WebSocketMessage represents incoming WebSocket messages
type WebSocketMessage struct {
	ID     string      `json:"id"`
	Method string      `json:"method"`
	Params interface{} `json:"params"`
}

// SubscriptionResult represents a subscription result
type SubscriptionResult struct {
	Subscription string      `json:"subscription"`
	Result       interface{} `json:"result"`
}

// LogsSubscription represents the logs subscription parameters
type LogsSubscription struct {
	Address []string   `json:"address,omitempty"`
	Topics  [][]string `json:"topics,omitempty"`
}

// NewSubscriptionManager creates a new subscription manager
func NewSubscriptionManager(chainID string, logger logging.Logger) *SubscriptionManager {
	return &SubscriptionManager{
		chainID:       chainID,
		subscriptions: make(map[string]*EventSubscription),
		eventFilters:  make(map[string][]common.Hash),
		logger:        logger,
	}
}

// AddEventSubscription adds a new event subscription
func (sm *SubscriptionManager) AddEventSubscription(contractAddr string, eventName string, eventSig string) (*EventSubscription, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Generate unique subscription ID
	subID := sm.generateSubscriptionID()

	addr := common.HexToAddress(contractAddr)
	sigHash := crypto.Keccak256Hash([]byte(eventSig))

	subscription := &EventSubscription{
		ID:           subID,
		ChainID:      sm.chainID,
		ContractAddr: addr,
		EventName:    eventName,
		EventSig:     sigHash,
		FilterQuery: ethereum.FilterQuery{
			Addresses: []common.Address{addr},
			Topics:    [][]common.Hash{{sigHash}},
		},
		Active:    true,
		CreatedAt: time.Now(),
	}

	sm.subscriptions[subID] = subscription

	// Add to event filters
	if sm.eventFilters[eventName] == nil {
		sm.eventFilters[eventName] = make([]common.Hash, 0)
	}
	sm.eventFilters[eventName] = append(sm.eventFilters[eventName], sigHash)

	sm.logger.Infof("Added subscription %s for %s events from %s on chain %s",
		subID, eventName, contractAddr, sm.chainID)

	return subscription, nil
}

// RemoveSubscription removes an event subscription
func (sm *SubscriptionManager) RemoveSubscription(subID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	subscription, exists := sm.subscriptions[subID]
	if !exists {
		return fmt.Errorf("subscription %s not found", subID)
	}

	subscription.Active = false
	delete(sm.subscriptions, subID)

	sm.logger.Infof("Removed subscription %s for %s events on chain %s",
		subID, subscription.EventName, sm.chainID)

	return nil
}

// GetActiveSubscriptions returns all active subscriptions
func (sm *SubscriptionManager) GetActiveSubscriptions() map[string]*EventSubscription {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	active := make(map[string]*EventSubscription)
	for id, sub := range sm.subscriptions {
		if sub.Active {
			active[id] = sub
		}
	}
	return active
}

// BuildWebSocketSubscription creates a WebSocket subscription message
func (sm *SubscriptionManager) BuildWebSocketSubscription() ([]byte, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Collect all contract addresses and topics
	addresses := make(map[common.Address]bool)
	var allTopics []common.Hash

	for _, sub := range sm.subscriptions {
		if sub.Active {
			addresses[sub.ContractAddr] = true
			allTopics = append(allTopics, sub.EventSig)
		}
	}

	// Convert to string arrays for JSON
	addressStrings := make([]string, 0, len(addresses))
	for addr := range addresses {
		addressStrings = append(addressStrings, addr.Hex())
	}

	topicStrings := make([]string, 0, len(allTopics))
	for _, topic := range allTopics {
		topicStrings = append(topicStrings, topic.Hex())
	}

	// Build subscription parameters
	params := map[string]interface{}{
		"id":     1,
		"method": "eth_subscribe",
		"params": []interface{}{
			"logs",
			LogsSubscription{
				Address: addressStrings,
				Topics:  [][]string{topicStrings},
			},
		},
	}

	return json.Marshal(params)
}

// ProcessWebSocketMessage processes incoming WebSocket messages
func (sm *SubscriptionManager) ProcessWebSocketMessage(data []byte, eventChan chan<- *ChainEvent) error {
	var msg WebSocketMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal WebSocket message: %w", err)
	}

	// Handle subscription notifications
	if msg.Method == "eth_subscription" {
		return sm.processSubscriptionNotification(msg.Params, eventChan)
	}

	return nil
}

// processSubscriptionNotification processes subscription notification messages
func (sm *SubscriptionManager) processSubscriptionNotification(params interface{}, eventChan chan<- *ChainEvent) error {
	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid subscription params format")
	}

	result, ok := paramsMap["result"]
	if !ok {
		return fmt.Errorf("no result in subscription notification")
	}

	// Parse the log entry
	logData, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal log data: %w", err)
	}

	var log types.Log
	if err := json.Unmarshal(logData, &log); err != nil {
		return fmt.Errorf("failed to unmarshal log: %w", err)
	}

	// Process the log entry
	return sm.processLogEntry(log, eventChan)
}

// processLogEntry processes a single log entry and routes it to the appropriate handler
func (sm *SubscriptionManager) processLogEntry(log types.Log, eventChan chan<- *ChainEvent) error {
	if len(log.Topics) == 0 {
		return fmt.Errorf("log entry has no topics")
	}

	eventSig := log.Topics[0]

	// Find matching subscription
	sm.mu.RLock()
	var matchedSub *EventSubscription
	for _, sub := range sm.subscriptions {
		if sub.Active && sub.EventSig == eventSig && sub.ContractAddr == log.Address {
			matchedSub = sub
			break
		}
	}
	sm.mu.RUnlock()

	if matchedSub == nil {
		sm.logger.Debugf("No subscription found for event %s from %s", eventSig.Hex(), log.Address.Hex())
		return nil
	}

	// Update subscription stats
	sm.updateSubscriptionStats(matchedSub.ID)

	// Create chain event
	chainEvent := &ChainEvent{
		ChainID:      sm.chainID,
		ChainName:    sm.getChainName(sm.chainID),
		ContractAddr: log.Address.Hex(),
		EventName:    matchedSub.EventName,
		BlockNumber:  log.BlockNumber,
		TxHash:       log.TxHash.Hex(),
		LogIndex:     log.Index,
		Data:         sm.parseEventData(matchedSub, log),
		RawLog:       log,
		ProcessedAt:  time.Now(),
	}

	// Send to event channel (non-blocking)
	select {
	case eventChan <- chainEvent:
		sm.logger.Debugf("Processed %s event from %s at block %d",
			matchedSub.EventName, log.Address.Hex(), log.BlockNumber)
	default:
		sm.logger.Warnf("Event channel full, dropping event %s from %s",
			matchedSub.EventName, log.Address.Hex())
	}

	return nil
}

// parseEventData parses event data based on the event type
func (sm *SubscriptionManager) parseEventData(sub *EventSubscription, log types.Log) interface{} {
	// Basic parsing - you can enhance this based on your specific event types
	eventData := map[string]interface{}{
		"event_signature": sub.EventSig.Hex(),
		"topics":          make([]string, len(log.Topics)),
		"data":            log.Data,
		"block_number":    log.BlockNumber,
		"tx_hash":         log.TxHash.Hex(),
		"log_index":       log.Index,
	}

	for i, topic := range log.Topics {
		eventData["topics"].([]string)[i] = topic.Hex()
	}

	// Add specific parsing for known events
	switch sub.EventName {
	case "OperatorRegistered":
		return sm.parseOperatorRegisteredEvent(log)
	case "OperatorUnregistered":
		return sm.parseOperatorUnregisteredEvent(log)
	case "TaskSubmitted":
		return sm.parseTaskSubmittedEvent(log)
	case "TaskRejected":
		return sm.parseTaskRejectedEvent(log)
	default:
		return eventData
	}
}

// parseOperatorRegisteredEvent parses OperatorRegistered events
func (sm *SubscriptionManager) parseOperatorRegisteredEvent(log types.Log) interface{} {
	// Implement specific parsing for OperatorRegistered events
	return map[string]interface{}{
		"event_type": "OperatorRegistered",
		"operator":   log.Topics[1].Hex(), // Assuming operator is first indexed parameter
		"raw_data":   log.Data,
		"block":      log.BlockNumber,
		"tx_hash":    log.TxHash.Hex(),
	}
}

// parseOperatorUnregisteredEvent parses OperatorUnregistered events
func (sm *SubscriptionManager) parseOperatorUnregisteredEvent(log types.Log) interface{} {
	return map[string]interface{}{
		"event_type": "OperatorUnregistered",
		"operator":   log.Topics[1].Hex(),
		"raw_data":   log.Data,
		"block":      log.BlockNumber,
		"tx_hash":    log.TxHash.Hex(),
	}
}

// parseTaskSubmittedEvent parses TaskSubmitted events
func (sm *SubscriptionManager) parseTaskSubmittedEvent(log types.Log) interface{} {
	return map[string]interface{}{
		"event_type": "TaskSubmitted",
		"task_id":    log.Topics[1].Hex(), // Assuming task ID is first indexed parameter
		"raw_data":   log.Data,
		"block":      log.BlockNumber,
		"tx_hash":    log.TxHash.Hex(),
	}
}

// parseTaskRejectedEvent parses TaskRejected events
func (sm *SubscriptionManager) parseTaskRejectedEvent(log types.Log) interface{} {
	return map[string]interface{}{
		"event_type": "TaskRejected",
		"task_id":    log.Topics[1].Hex(),
		"raw_data":   log.Data,
		"block":      log.BlockNumber,
		"tx_hash":    log.TxHash.Hex(),
	}
}

// updateSubscriptionStats updates statistics for a subscription
func (sm *SubscriptionManager) updateSubscriptionStats(subID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sub, exists := sm.subscriptions[subID]; exists {
		sub.EventCount++
		sub.LastEvent = time.Now()
	}
}

// generateSubscriptionID generates a unique subscription ID
func (sm *SubscriptionManager) generateSubscriptionID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return fmt.Sprintf("%s_%s", sm.chainID, hex.EncodeToString(bytes))
}

// getChainName returns a human-readable chain name
func (sm *SubscriptionManager) getChainName(chainID string) string {
	chainNames := map[string]string{
		"17000":    "Ethereum Holesky",
		"11155111": "Ethereum Sepolia",
		"11155420": "Optimism Sepolia",
		"84532":    "Base Sepolia",
	}

	if name, exists := chainNames[chainID]; exists {
		return name
	}
	return fmt.Sprintf("Chain %s", chainID)
}

// GetSubscriptionStats returns statistics for all subscriptions
func (sm *SubscriptionManager) GetSubscriptionStats() map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	stats := map[string]interface{}{
		"chain_id":             sm.chainID,
		"total_subscriptions":  len(sm.subscriptions),
		"active_subscriptions": 0,
		"subscriptions":        make([]map[string]interface{}, 0),
	}

	activeCount := 0
	for _, sub := range sm.subscriptions {
		if sub.Active {
			activeCount++
		}

		subStats := map[string]interface{}{
			"id":          sub.ID,
			"event_name":  sub.EventName,
			"contract":    sub.ContractAddr.Hex(),
			"active":      sub.Active,
			"created_at":  sub.CreatedAt,
			"last_event":  sub.LastEvent,
			"event_count": sub.EventCount,
		}

		stats["subscriptions"] = append(stats["subscriptions"].([]map[string]interface{}), subStats)
	}

	stats["active_subscriptions"] = activeCount
	return stats
}
