package worker

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/event/metrics"
	schedulerTypes "github.com/trigg3rX/triggerx-backend/internal/schedulers/event/scheduler/types"
)

// checkForEvents checks for new events since the last processed block
func (w *EventWorker) checkForEvents() error {
	// Get current block number
	currentBlock, err := w.Client.BlockNumber(context.Background())
	if err != nil {
		metrics.TrackRPCRequest(w.Job.TriggerChainID, "eth_blockNumber", "failed")
		metrics.TrackConnectionFailure(w.Job.TriggerChainID)
		return fmt.Errorf("failed to get current block number: %w", err)
	}
	metrics.TrackRPCRequest(w.Job.TriggerChainID, "eth_blockNumber", "success")

	// Calculate safe block (with confirmations)
	safeBlock := currentBlock
	if currentBlock > schedulerTypes.BlockConfirmations {
		safeBlock = currentBlock - schedulerTypes.BlockConfirmations
	}

	// Check if there are new blocks to process
	if safeBlock <= w.LastBlock {
		return nil // No new blocks to process
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
		metrics.TrackRPCRequest(w.Job.TriggerChainID, "eth_getLogs", "failed")
		metrics.TrackConnectionFailure(w.Job.TriggerChainID)
		return fmt.Errorf("failed to filter logs: %w", err)
	}
	metrics.TrackRPCRequest(w.Job.TriggerChainID, "eth_getLogs", "success")

	// Process each event
	for _, log := range logs {
		startTime := time.Now()

		if err := w.processEvent(log); err != nil {
			w.Logger.Error("Failed to process event",
				"job_id", w.Job.JobID,
				"tx_hash", log.TxHash.Hex(),
				"block", log.BlockNumber,
				"error", err,
			)
			metrics.TrackJobCompleted("failed")
			metrics.TrackWorkerError(fmt.Sprintf("%d", w.Job.JobID), "event_processing_error")

			// Track as critical error if event processing fails
			metrics.TrackCriticalError("event_processing_failure")
		} else {
			// Track successful event processing with comprehensive metrics
			processingDuration := time.Since(startTime)
			metrics.TrackEventWithDuration(w.Job.TriggerChainID, processingDuration, true)
			metrics.TrackJobCompleted("success")
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

	// Create comprehensive event context for Redis streaming
	eventContext := map[string]interface{}{
		"event_type":               "event_detected",
		"job_id":                   w.Job.JobID,
		"manager_id":               w.ManagerID,
		"trigger_chain_id":         w.Job.TriggerChainID,
		"trigger_contract_address": w.Job.TriggerContractAddress,
		"trigger_event":            w.Job.TriggerEvent,
		"target_chain_id":          w.Job.TaskTargetData.TargetChainID,
		"target_contract_address":  w.Job.TaskTargetData.TargetContractAddress,
		"target_function":          w.Job.TaskTargetData.TargetFunction,
		"tx_hash":                  log.TxHash.Hex(),
		"block_number":             log.BlockNumber,
		"log_index":                log.Index,
		"gas_used":                 log.BlockHash.Hex(),
		"detected_at":              startTime.Unix(),
		"status":                   "processing",
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

		w.Logger.Info("Event processed successfully",
			"job_id", w.Job.JobID,
			"tx_hash", log.TxHash.Hex(),
			"block", log.BlockNumber,
			"target_chain", w.Job.TaskTargetData.TargetChainID,
			"target_function", w.Job.TaskTargetData.TargetFunction,
			"duration", duration,
		)
		metrics.TrackActionExecution(fmt.Sprintf("%d", w.Job.JobID), "success")
	} else {
		eventContext["event_type"] = "event_failed"
		eventContext["status"] = "failed"
		eventContext["error"] = "action execution failed"

		w.Logger.Error("Event processing failed",
			"job_id", w.Job.JobID,
			"tx_hash", log.TxHash.Hex(),
			"block", log.BlockNumber,
			"target_chain", w.Job.TaskTargetData.TargetChainID,
			"target_function", w.Job.TaskTargetData.TargetFunction,
			"duration", duration,
		)
		metrics.TrackActionExecution(fmt.Sprintf("%d", w.Job.JobID), "failed")
	}

	return nil
}
