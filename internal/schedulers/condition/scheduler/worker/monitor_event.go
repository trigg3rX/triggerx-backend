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

	// Query logs for events
	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(w.LastBlock + 1),
		ToBlock:   new(big.Int).SetUint64(currentBlock),
		Addresses: []common.Address{contractAddr},
		Topics:    [][]common.Hash{{eventSig}},
	}

	logs, err := w.ChainClient.FilterLogs(context.Background(), query)
	if err != nil {
		metrics.TrackCriticalError("rpc_filter_logs_failed")
		return fmt.Errorf("failed to filter logs: %w", err)
	}

	// Process each event
	for _, log := range logs {
		// Apply event filtering if configured
		if w.shouldFilterEvent() {
			if !w.matchesEventFilter(log) {
				w.Logger.Debug("Event filtered out",
					"job_id", w.EventWorkerData.JobID,
					"tx_hash", log.TxHash.Hex(),
					"block", log.BlockNumber,
					"filter_param", w.EventWorkerData.TriggerEventFilterParaName,
					"filter_value", w.EventWorkerData.TriggerEventFilterValue,
				)
				continue
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

	w.Logger.Debug("Processed blocks",
		"job_id", w.EventWorkerData.JobID,
		"from_block", w.LastBlock+1-uint64(len(logs)),
		"to_block", currentBlock,
		"events_found", len(logs),
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
			JobID:         w.EventWorkerData.JobID,
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
	return strings.TrimSpace(w.EventWorkerData.TriggerEventFilterParaName) != "" &&
		strings.TrimSpace(w.EventWorkerData.TriggerEventFilterValue) != ""
}

// matchesEventFilter checks if the event matches the configured filter
func (w *EventWorker) matchesEventFilter(log types.Log) bool {
	filterParam := strings.TrimSpace(w.EventWorkerData.TriggerEventFilterParaName)
	filterValue := strings.TrimSpace(w.EventWorkerData.TriggerEventFilterValue)

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
		filterAddrHash := common.BytesToHash(filterAddr.Bytes())
		for i, topic := range log.Topics {
			if i == 0 {
				continue
			}
			if topic == filterAddrHash {
				w.Logger.Debug("Event filter matched address in topic",
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
