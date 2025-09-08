package websocket

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

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
		ContractType: matchedSub.ContractType,
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
