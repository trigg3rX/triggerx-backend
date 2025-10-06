package worker

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/metrics"
)

// checkForEvents checks for new events since the last processed block
func (w *EventWorker) checkForEvents(contractAddr common.Address, eventSig common.Hash) error {
	// Get current block number
	currentBlock, err := w.ChainClient.BlockNumber(context.Background())
	if err != nil {
		metrics.TrackCriticalError("rpc_block_number_failed")
		return fmt.Errorf("failed to get current block number: %w", err)
	}

	// Check if there are new blocks to process
	if currentBlock <= w.LastBlock {
		return nil // No new blocks to process
	}

	// Query logs for events in chunks to respect API limits
	// Alchemy free tier allows max 10 blocks per eth_getLogs request
	const maxBlockRange = 10
	fromBlock := w.LastBlock + 1
	allLogs := []types.Log{}

	for fromBlock <= currentBlock {
		toBlock := fromBlock + maxBlockRange - 1
		if toBlock > currentBlock {
			toBlock = currentBlock
		}

		query := ethereum.FilterQuery{
			FromBlock: new(big.Int).SetUint64(fromBlock),
			ToBlock:   new(big.Int).SetUint64(toBlock),
			Addresses: []common.Address{contractAddr},
			Topics:    [][]common.Hash{{eventSig}},
		}

		w.Logger.Debug("Querying block range",
			"job_id", w.EventWorkerData.JobID,
			"from_block", fromBlock,
			"to_block", toBlock,
			"range_size", toBlock-fromBlock+1,
		)

		logs, err := w.ChainClient.FilterLogs(context.Background(), query)
		if err != nil {
			metrics.TrackCriticalError("rpc_filter_logs_failed")
			return fmt.Errorf("failed to filter logs for blocks %d-%d: %w", fromBlock, toBlock, err)
		}

		allLogs = append(allLogs, logs...)
		fromBlock = toBlock + 1
	}

	logs := allLogs

	// Process each event
	for _, log := range logs {
		w.Logger.Info("Found raw event",
			"job_id", w.EventWorkerData.JobID,
			"tx_hash", log.TxHash.Hex(),
			"block", log.BlockNumber,
			"topics", len(log.Topics),
			"contract", log.Address.Hex(),
		)
		
		// Apply event filtering if configured
		if w.shouldFilterEvent() {
			w.Logger.Info("Applying event filter",
				"job_id", w.EventWorkerData.JobID,
				"filter_param", w.EventWorkerData.EventFilterParaName,
				"filter_value", w.EventWorkerData.EventFilterValue,
			)
			if !w.matchesEventFilter(log) {
				w.Logger.Info("Event filtered out",
					"job_id", w.EventWorkerData.JobID,
					"tx_hash", log.TxHash.Hex(),
					"block", log.BlockNumber,
					"filter_param", w.EventWorkerData.EventFilterParaName,
					"filter_value", w.EventWorkerData.EventFilterValue,
				)
				continue
			} else {
				w.Logger.Info("Event passed filter",
					"job_id", w.EventWorkerData.JobID,
					"tx_hash", log.TxHash.Hex(),
				)
			}
		}

		if err := w.processEvent(log); err != nil {
			w.Logger.Error("Failed to process event",
				"job_id", w.EventWorkerData.JobID,
				"tx_hash", log.TxHash.Hex(),
				"block", log.BlockNumber,
				"error", err,
			)
			metrics.TrackCriticalError("event_processing_failed")
		}
	}

	// Update last processed block
	w.LastBlock = currentBlock

	fromBlock = w.LastBlock + 1
	if currentBlock <= w.LastBlock {
		fromBlock = w.LastBlock
	}
	
	w.Logger.Info("Processed blocks",
		"job_id", w.EventWorkerData.JobID,
		"from_block", fromBlock,
		"to_block", currentBlock,
		"blocks_scanned", int64(currentBlock) - int64(w.LastBlock),
		"events_found", len(logs),
		"contract_address", contractAddr.Hex(),
		"event_signature", eventSig.Hex(),
	)

	return nil
}

// processEvent processes a single event and notifies the scheduler
func (w *EventWorker) processEvent(log types.Log) error {
	w.Logger.Info("Event detected",
		"job_id", w.EventWorkerData.JobID,
		"tx_hash", log.TxHash.Hex(),
		"block", log.BlockNumber,
		"log_index", log.Index,
		"chain_id", w.EventWorkerData.TriggerChainID,
		"event", w.EventWorkerData.TriggerEvent,
	)

	// Notify scheduler about the event
	if w.TriggerCallback != nil {
		notification := &TriggerNotification{
			JobID:         w.EventWorkerData.JobID.ToBigInt(),
			TriggerTxHash: log.TxHash.Hex(),
			TriggeredAt:   time.Now(),
		}

		if err := w.TriggerCallback(notification); err != nil {
			w.Logger.Error("Failed to notify scheduler about event",
				"job_id", w.EventWorkerData.JobID,
				"tx_hash", log.TxHash.Hex(),
				"error", err,
			)
			metrics.TrackCriticalError("event_notification_failed")
			return err
		} else {
			w.Logger.Info("Successfully notified scheduler about event",
				"job_id", w.EventWorkerData.JobID,
				"tx_hash", log.TxHash.Hex(),
			)
		}
	} else {
		w.Logger.Warn("No trigger callback configured for event worker",
			"job_id", w.EventWorkerData.JobID,
		)
	}

	// For non-recurring jobs, stop the worker after triggering
	if !w.EventWorkerData.Recurring {
		w.Logger.Info("Non-recurring job triggered, stopping worker", "job_id", w.EventWorkerData.JobID)
		go w.Stop() // Stop in a goroutine to avoid deadlock
	}

	return nil
}

// shouldFilterEvent checks if event filtering is enabled
func (w *EventWorker) shouldFilterEvent() bool {
	return strings.TrimSpace(w.EventWorkerData.EventFilterParaName) != "" && 
		   strings.TrimSpace(w.EventWorkerData.EventFilterValue) != ""
}

// matchesEventFilter checks if the event matches the configured filter
func (w *EventWorker) matchesEventFilter(log types.Log) bool {
	filterParam := strings.TrimSpace(w.EventWorkerData.EventFilterParaName)
	filterValue := strings.TrimSpace(w.EventWorkerData.EventFilterValue)

	if filterParam == "" || filterValue == "" {
		return true // No filtering configured
	}

	// For basic filtering, we'll check indexed topics and data fields
	// This is a simplified implementation that works with common event patterns
	
	// Convert filter value to compare with event data
	filterValueLower := strings.ToLower(filterValue)

	// Check indexed topics (topics[1], topics[2], topics[3] are indexed parameters)
	// topics[0] is the event signature
	for i, topic := range log.Topics {
		if i == 0 {
			continue // Skip event signature
		}
		
		// Convert topic to string representation for comparison
		topicStr := strings.ToLower(topic.Hex())
		topicBigInt := topic.Big()
		
		// Check if this might be the parameter we're looking for
		// We'll do a fuzzy match since parameter names aren't directly available in logs
		if strings.Contains(topicStr, filterValueLower) {
			w.Logger.Debug("Event filter matched in topic",
				"job_id", w.EventWorkerData.JobID,
				"topic_index", i,
				"topic_value", topicStr,
				"filter_value", filterValue,
			)
			return true
		}
		
		// Also check if the filter value matches the big int representation
		if topicBigInt.String() == filterValue {
			w.Logger.Debug("Event filter matched in topic (big int)",
				"job_id", w.EventWorkerData.JobID,
				"topic_index", i,
				"topic_value", topicBigInt.String(),
				"filter_value", filterValue,
			)
			return true
		}
	}

	// Check event data for non-indexed parameters
	dataStr := strings.ToLower(common.Bytes2Hex(log.Data))
	if strings.Contains(dataStr, filterValueLower) {
		w.Logger.Debug("Event filter matched in data",
			"job_id", w.EventWorkerData.JobID,
			"data_snippet", dataStr[:min(len(dataStr), 100)], // Log first 100 chars for debugging
			"filter_value", filterValue,
		)
		return true
	}

	// Check if the filter value is a hex address
	if common.IsHexAddress(filterValue) {
		filterAddr := common.HexToAddress(filterValue)
		// Convert address to hash for topic comparison (addresses are left-padded to 32 bytes in topics)
		addressHash := common.HexToHash(filterAddr.Hex())
		w.Logger.Info("Checking address filter",
			"job_id", w.EventWorkerData.JobID,
			"filter_address", filterAddr.Hex(),
			"address_hash", addressHash.Hex(),
			"num_topics", len(log.Topics),
		)
		for i, topic := range log.Topics {
			if i == 0 {
				continue
			}
			w.Logger.Info("Comparing topic",
				"job_id", w.EventWorkerData.JobID,
				"topic_index", i,
				"topic_value", topic.Hex(),
				"expected_hash", addressHash.Hex(),
				"matches", topic == addressHash,
			)
			if topic == addressHash {
				w.Logger.Info("Event filter matched address in topic",
					"job_id", w.EventWorkerData.JobID,
					"topic_index", i,
					"address", filterAddr.Hex(),
				)
				return true
			}
		}
	}

	w.Logger.Debug("Event filter did not match",
		"job_id", w.EventWorkerData.JobID,
		"filter_param", filterParam,
		"filter_value", filterValue,
		"tx_hash", log.TxHash.Hex(),
	)
	
	return false
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
