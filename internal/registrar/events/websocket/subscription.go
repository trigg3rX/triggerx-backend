package websocket

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"

	// Contract bindings
	// "github.com/trigg3rX/triggerx-contracts/bindings/contracts/AVSGovernanceLogic"
	"github.com/trigg3rX/triggerx-contracts/bindings/contracts/AttestationCenter"
	"github.com/trigg3rX/triggerx-contracts/bindings/contracts/AvsGovernance"
	// "github.com/trigg3rX/triggerx-contracts/bindings/contracts/OBLS"
)

// ContractType represents the type of contract
type ContractType string

const (
	ContractTypeAttestationCenter  ContractType = "attestation_center"
	ContractTypeAvsGovernance      ContractType = "avs_governance"
	// ContractTypeAVSGovernanceLogic ContractType = "avs_governance_logic"
	// ContractTypeOBLS               ContractType = "obls"
)

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
		sm.logger.Infof("Initialized AttestationCenter ABI")
	} else {
		sm.logger.Errorf("Failed to initialize AttestationCenter ABI: %v", err)
	}

	// Initialize AvsGovernance ABI
	if avsGovernanceABI, err := contractAvsGovernance.ContractAvsGovernanceMetaData.GetAbi(); err == nil {
		sm.contractABIs[ContractTypeAvsGovernance] = *avsGovernanceABI
		sm.logger.Infof("Initialized AvsGovernance ABI")
	} else {
		sm.logger.Errorf("Failed to initialize AvsGovernance ABI: %v", err)
	}

	// // Initialize AVSGovernanceLogic ABI
	// if avsGovernanceLogicABI, err := contractAVSGovernanceLogic.ContractAVSGovernanceLogicMetaData.GetAbi(); err == nil {
	// 	sm.contractABIs[ContractTypeAVSGovernanceLogic] = *avsGovernanceLogicABI
	// 	sm.logger.Infof("Initialized AVSGovernanceLogic ABI")
	// } else {
	// 	sm.logger.Errorf("Failed to initialize AVSGovernanceLogic ABI: %v", err)
	// }

	// // Initialize OBLS ABI
	// if oblsABI, err := contractOBLS.ContractOBLSMetaData.GetAbi(); err == nil {
	// 	sm.contractABIs[ContractTypeOBLS] = *oblsABI
	// 	sm.logger.Infof("Initialized OBLS ABI")
	// } else {
	// 	sm.logger.Errorf("Failed to initialize OBLS ABI: %v", err)
	// }
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

	sm.logger.Infof("Added subscription %s for %s.%s events from %s on chain %s",
		subID, contractType, eventName, contractAddr, sm.chainID)

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

// parseEventData parses event data based on the contract type and event
func (sm *SubscriptionManager) parseEventData(sub *EventSubscription, log types.Log) interface{} {
	// Check if we have a contract type for proper parsing
	if sub.ContractType != "" {
		return sm.parseContractEventData(sub, log)
	}

	// Fallback to basic parsing for legacy subscriptions
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

	return eventData
}

// parseContractEventData parses contract event data using the proper ABI
func (sm *SubscriptionManager) parseContractEventData(sub *EventSubscription, log types.Log) interface{} {
	contractABI, exists := sm.contractABIs[sub.ContractType]
	if !exists {
		sm.logger.Errorf("Contract ABI not found for type %s", sub.ContractType)
		return sm.parseBasicEventData(sub, log)
	}

	event, exists := contractABI.Events[sub.EventName]
	if !exists {
		sm.logger.Errorf("Event %s not found in contract %s ABI", sub.EventName, sub.ContractType)
		return sm.parseBasicEventData(sub, log)
	}

	// Parse the event data
	parsedData := make(map[string]interface{})

	// Parse indexed parameters from topics
	topicIndex := 1 // Skip the event signature (topics[0])
	for _, input := range event.Inputs {
		if input.Indexed {
			if topicIndex < len(log.Topics) {
				parsedData[input.Name] = sm.parseTopicData(input, log.Topics[topicIndex])
				topicIndex++
			}
		}
	}

	// Parse non-indexed parameters from data
	if len(log.Data) > 0 {
		nonIndexedInputs := make([]abi.Argument, 0)
		for _, input := range event.Inputs {
			if !input.Indexed {
				nonIndexedInputs = append(nonIndexedInputs, input)
			}
		}

		if len(nonIndexedInputs) > 0 {
			values, err := contractABI.Unpack(sub.EventName, log.Data)
			if err != nil {
				sm.logger.Errorf("Failed to unpack event data for %s: %v", sub.EventName, err)
			} else {
				for i, input := range nonIndexedInputs {
					if i < len(values) {
						parsedData[input.Name] = sm.formatValue(values[i])
					}
				}
			}
		}
	}

	return &ContractEventData{
		EventType:    sub.EventName,
		ContractType: sub.ContractType,
		ParsedData:   parsedData,
		RawData:      log.Data,
		Topics:       sm.formatTopics(log.Topics),
		BlockNumber:  log.BlockNumber,
		TxHash:       log.TxHash.Hex(),
		LogIndex:     log.Index,
	}
}

// parseBasicEventData provides basic event data parsing as fallback
func (sm *SubscriptionManager) parseBasicEventData(sub *EventSubscription, log types.Log) interface{} {
	return map[string]interface{}{
		"event_type":      sub.EventName,
		"contract_type":   sub.ContractType,
		"event_signature": sub.EventSig.Hex(),
		"topics":          sm.formatTopics(log.Topics),
		"data":            hex.EncodeToString(log.Data),
		"block_number":    log.BlockNumber,
		"tx_hash":         log.TxHash.Hex(),
		"log_index":       log.Index,
	}
}

// parseTopicData parses topic data based on the input type
func (sm *SubscriptionManager) parseTopicData(input abi.Argument, topic common.Hash) interface{} {
	switch input.Type.String() {
	case "address":
		return common.HexToAddress(topic.Hex()).Hex()
	case "uint256", "uint128", "uint64", "uint32", "uint16", "uint8":
		return new(big.Int).SetBytes(topic.Bytes()).String()
	case "int256", "int128", "int64", "int32", "int16", "int8":
		// For signed integers, we need to handle two's complement
		value := new(big.Int).SetBytes(topic.Bytes())
		if value.Bit(255) == 1 { // Check if the sign bit is set
			// Convert from two's complement
			max := new(big.Int).Lsh(big.NewInt(1), 256)
			value.Sub(value, max)
		}
		return value.String()
	case "bytes32":
		return topic.Hex()
	case "bool":
		return topic.Big().Cmp(big.NewInt(0)) != 0
	default:
		return topic.Hex()
	}
}

// formatValue formats values for JSON serialization
func (sm *SubscriptionManager) formatValue(value interface{}) interface{} {
	switch v := value.(type) {
	case *big.Int:
		return v.String()
	case common.Address:
		return v.Hex()
	case common.Hash:
		return v.Hex()
	case []byte:
		return hex.EncodeToString(v)
	default:
		return v
	}
}

// formatTopics formats topic slice for JSON serialization
func (sm *SubscriptionManager) formatTopics(topics []common.Hash) []string {
	result := make([]string, len(topics))
	for i, topic := range topics {
		result[i] = topic.Hex()
	}
	return result
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
	_, err := rand.Read(bytes)
	if err != nil {
		sm.logger.Errorf("Failed to generate subscription ID: %v", err)
		return ""
	}
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
