package websocket

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"

	// Contract bindings
	contractAttestationCenter "github.com/trigg3rX/triggerx-contracts/bindings/contracts/AttestationCenter"
)

// ContractType represents the type of contract
type ContractType string

const (
	ContractTypeAttestationCenter ContractType = "attestation_center"
)

// ContractConfig represents a contract to monitor
type ContractConfig struct {
	Address      string
	ContractType ContractType
	ABI          string
	Events       []string // Event names to monitor
}

// ChainConfig represents configuration for a specific blockchain
type ChainConfig struct {
	ChainID      string
	Name         string
	RPCURL       string
	WebSocketURL string
	Contracts    []ContractConfig
	Reconnect    ReconnectConfig
}

// SubscriptionManager manages WebSocket event subscriptions for a chain
type SubscriptionManager struct {
	chainID       string
	subscriptions map[string]*EventSubscription
	eventFilters  map[string][]common.Hash // event name -> topic hashes
	contractABIs  map[ContractType]abi.ABI
	logger        logging.Logger
	mu            sync.RWMutex
}

// EventSubscription represents a single event subscription
type EventSubscription struct {
	ID           string
	ChainID      string
	ContractAddr common.Address
	ContractType ContractType
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
	ID     int         `json:"id"`
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

// ContractEventData represents parsed contract event data
type ContractEventData struct {
	EventType    string                 `json:"event_type"`
	ContractType ContractType           `json:"contract_type"`
	ParsedData   map[string]interface{} `json:"parsed_data"`
	RawData      []byte                 `json:"raw_data"`
	Topics       []string               `json:"topics"`
	BlockNumber  uint64                 `json:"block_number"`
	TxHash       string                 `json:"tx_hash"`
	LogIndex     uint                   `json:"log_index"`
}

// NewSubscriptionManager creates a new subscription manager
func NewSubscriptionManager(chainID string, logger logging.Logger) *SubscriptionManager {
	sm := &SubscriptionManager{
		chainID:       chainID,
		subscriptions: make(map[string]*EventSubscription),
		eventFilters:  make(map[string][]common.Hash),
		contractABIs:  make(map[ContractType]abi.ABI),
		logger:        logger,
	}

	// Initialize contract ABIs
	sm.initializeContractABIs()

	return sm
}

// initializeContractABIs initializes the contract ABIs
func (sm *SubscriptionManager) initializeContractABIs() {
	// Initialize AttestationCenter ABI
	if attestationCenterABI, err := contractAttestationCenter.ContractAttestationCenterMetaData.GetAbi(); err == nil {
		sm.contractABIs[ContractTypeAttestationCenter] = *attestationCenterABI
	} else {
		sm.logger.Errorf("Failed to initialize AttestationCenter ABI: %v", err)
	}
}

// AddContractSubscription adds a new contract event subscription
func (sm *SubscriptionManager) AddContractSubscription(contractAddr string, contractType ContractType, eventName string) (*EventSubscription, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Get the ABI for the contract
	contractABI, exists := sm.contractABIs[contractType]
	if !exists {
		return nil, fmt.Errorf("contract type %s not found", contractType)
	}

	// Get the event from the ABI
	event, exists := contractABI.Events[eventName]
	if !exists {
		return nil, fmt.Errorf("event %s not found in contract %s", eventName, contractType)
	}

	// Generate unique subscription ID
	subID := sm.generateSubscriptionID()

	addr := common.HexToAddress(contractAddr)
	eventSig := event.ID

	subscription := &EventSubscription{
		ID:           subID,
		ChainID:      sm.chainID,
		ContractAddr: addr,
		ContractType: contractType,
		EventName:    eventName,
		EventSig:     eventSig,
		FilterQuery: ethereum.FilterQuery{
			Addresses: []common.Address{addr},
			Topics:    [][]common.Hash{{eventSig}},
		},
		Active:    true,
		CreatedAt: time.Now(),
	}

	sm.subscriptions[subID] = subscription

	// Add to event filters
	if sm.eventFilters[eventName] == nil {
		sm.eventFilters[eventName] = make([]common.Hash, 0)
	}
	sm.eventFilters[eventName] = append(sm.eventFilters[eventName], eventSig)

	// sm.logger.Infof("Added subscription %s for %s.%s events from %s on chain %s",
	// 	subID, contractType, eventName, contractAddr, sm.chainID)

	return subscription, nil
}

// AddEventSubscription adds a new event subscription (legacy method)
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

	// sm.logger.Infof("Added subscription %s for %s events from %s on chain %s",
	// 	subID, eventName, contractAddr, sm.chainID)

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

	// sm.logger.Infof("Removed subscription %s for %s events on chain %s",
	// 	subID, subscription.EventName, sm.chainID)

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
