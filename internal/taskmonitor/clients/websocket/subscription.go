package websocket

import (
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

// ReconnectConfig holds reconnection configuration
type ReconnectConfig struct {
	MaxRetries    int           `default:"10"`
	BaseDelay     time.Duration `default:"5s"`
	MaxDelay      time.Duration `default:"300s"` // 5 minutes
	BackoffFactor float64       `default:"2.0"`
	Jitter        bool          `default:"true"`
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
