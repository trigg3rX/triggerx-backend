package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	redisx "github.com/trigg3rX/triggerx-backend/internal/redis"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/event/metrics"
	schedulerTypes "github.com/trigg3rX/triggerx-backend/internal/schedulers/event/scheduler/types"
)

// checkForEvents checks for new events since the last processed block
func (w *EventWorker) checkForEvents() error {
	// Get current block number
	currentBlock, err := w.Client.BlockNumber(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get current block number: %w", err)
	}

	// Calculate safe block (with confirmations)
	safeBlock := currentBlock
	if currentBlock > schedulerTypes.BlockConfirmations {
		safeBlock = currentBlock - schedulerTypes.BlockConfirmations
	}

	// Check if there are new blocks to process
	if safeBlock <= w.LastBlock {
		return nil // No new blocks to process
	}

	// Check cache for recent events in this block range to avoid reprocessing
	blockRangeKey := fmt.Sprintf("events_%d_%d_%d", w.Job.JobID, w.LastBlock+1, safeBlock)
	if w.Cache != nil {
		if _, err := w.Cache.Get(blockRangeKey); err == nil {
			w.Logger.Debug("Block range already processed", "job_id", w.Job.JobID, "from", w.LastBlock+1, "to", safeBlock)
			w.LastBlock = safeBlock
			return nil
		}
	}

	// Query logs for events
	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(w.LastBlock + 1),
		ToBlock:   new(big.Int).SetUint64(safeBlock),
		Addresses: []common.Address{w.ContractAddr},
		Topics:    [][]common.Hash{{w.EventSig}},
	}

	logs, err := w.Client.FilterLogs(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to filter logs: %w", err)
	}

	// Process each event
	for _, log := range logs {
		metrics.EventsDetected.Inc()
		if err := w.processEvent(log); err != nil {
			w.Logger.Error("Failed to process event",
				"job_id", w.Job.JobID,
				"tx_hash", log.TxHash.Hex(),
				"block", log.BlockNumber,
				"error", err,
			)
			metrics.JobsFailed.Inc()
		} else {
			metrics.EventsProcessed.Inc()
		}
	}

	// Cache that this block range has been processed
	if w.Cache != nil {
		processedData := map[string]interface{}{
			"job_id":       w.Job.JobID,
			"from_block":   w.LastBlock + 1,
			"to_block":     safeBlock,
			"events_found": len(logs),
			"processed_at": time.Now().Unix(),
		}
		if jsonData, err := json.Marshal(processedData); err == nil {
			if err := w.Cache.Set(blockRangeKey, string(jsonData), schedulerTypes.EventCacheTTL); err != nil {
				w.Logger.Errorf("Failed to set block range cache: %v", err)
			}
		}
	}

	// Update last processed block
	w.LastBlock = safeBlock

	w.Logger.Debug("Processed blocks",
		"job_id", w.Job.JobID,
		"from_block", w.LastBlock+1-uint64(len(logs)),
		"to_block", safeBlock,
		"events_found", len(logs),
	)

	return nil
}

// processEvent processes a single event and triggers the action
func (w *EventWorker) processEvent(log types.Log) error {
	startTime := time.Now()

	w.Logger.Info("Event detected",
		"job_id", w.Job.JobID,
		"tx_hash", log.TxHash.Hex(),
		"block", log.BlockNumber,
		"log_index", log.Index,
	)

	// Check for duplicate event processing
	eventKey := fmt.Sprintf("event_%s_%d", log.TxHash.Hex(), log.Index)
	if w.Cache != nil {
		if cachedValue, err := w.Cache.Get(eventKey); err == nil {
			w.Logger.Debug("Event already processed, skipping",
				"tx_hash", log.TxHash.Hex(),
				"cached_at", cachedValue,
			)

			// Add duplicate event detection to Redis stream
			if redisx.IsAvailable() {
				duplicateEvent := map[string]interface{}{
					"event_type":   "event_duplicate_detected",
					"job_id":       w.Job.JobID,
					"manager_id":   w.ManagerID,
					"chain_id":     w.Job.TriggerChainID,
					"contract":     w.Job.TriggerContractAddress,
					"event":        w.Job.TriggerEvent,
					"tx_hash":      log.TxHash.Hex(),
					"block_number": log.BlockNumber,
					"log_index":    log.Index,
					"cached_at":    cachedValue,
					"detected_at":  startTime.Unix(),
				}
				if err := redisx.AddJobToStream(redisx.JobsReadyEventStream, duplicateEvent); err != nil {
					w.Logger.Warnf("Failed to add event duplicate detection to Redis stream: %v", err)
				}
			}
			return nil
		}

		// Mark event as processed
		if err := w.Cache.Set(eventKey, time.Now().Format(time.RFC3339), schedulerTypes.DuplicateEventWindow); err != nil {
			w.Logger.Errorf("Failed to set event cache: %v", err)
		}
	}

	// Create comprehensive event context for Redis streaming
	eventContext := map[string]interface{}{
		"event_type":               "event_detected",
		"job_id":                   w.Job.JobID,
		"manager_id":               w.ManagerID,
		"trigger_chain_id":         w.Job.TriggerChainID,
		"trigger_contract_address": w.Job.TriggerContractAddress,
		"trigger_event":            w.Job.TriggerEvent,
		"target_chain_id":          w.Job.TargetChainID,
		"target_contract_address":  w.Job.TargetContractAddress,
		"target_function":          w.Job.TargetFunction,
		"tx_hash":                  log.TxHash.Hex(),
		"block_number":             log.BlockNumber,
		"log_index":                log.Index,
		"gas_used":                 log.BlockHash.Hex(), // Block hash for reference
		"cache_available":          w.Cache != nil,
		"detected_at":              startTime.Unix(),
		"status":                   "processing",
	}

	// Add event detection to Redis stream
	if redisx.IsAvailable() {
		if err := redisx.AddJobToStream(redisx.JobsReadyEventStream, eventContext); err != nil {
			w.Logger.Warnf("Failed to add event detection to Redis stream: %v", err)
		}
	}

	// Execute the action
	executionSuccess := w.performActionExecution(log)

	duration := time.Since(startTime)

	// Update event context with completion info
	eventContext["duration_ms"] = duration.Milliseconds()
	eventContext["completed_at"] = time.Now().Unix()

	if executionSuccess {
		eventContext["event_type"] = "event_completed"
		eventContext["status"] = "completed"

		if redisx.IsAvailable() {
			if err := redisx.AddJobToStream(redisx.JobsReadyEventStream, eventContext); err != nil {
				w.Logger.Warnf("Failed to add event completion to Redis stream: %v", err)
			}
		}

		w.Logger.Info("Event processed successfully",
			"job_id", w.Job.JobID,
			"tx_hash", log.TxHash.Hex(),
			"block", log.BlockNumber,
			"target_chain", w.Job.TargetChainID,
			"target_function", w.Job.TargetFunction,
			"duration", duration,
		)
	} else {
		eventContext["event_type"] = "event_failed"
		eventContext["status"] = "failed"
		eventContext["error"] = "action execution failed"

		if redisx.IsAvailable() {
			if err := redisx.AddJobToStream(redisx.JobsRetryEventStream, eventContext); err != nil {
				w.Logger.Warnf("Failed to add event failure to Redis stream: %v", err)
			}
		}

		w.Logger.Error("Event processing failed",
			"job_id", w.Job.JobID,
			"tx_hash", log.TxHash.Hex(),
			"block", log.BlockNumber,
			"target_chain", w.Job.TargetChainID,
			"target_function", w.Job.TargetFunction,
			"duration", duration,
		)
	}

	return nil
}