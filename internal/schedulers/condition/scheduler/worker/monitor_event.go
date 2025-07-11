package worker

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/trigg3rX/triggerx-backend-imua/internal/schedulers/condition/metrics"
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
			JobID:       w.EventWorkerData.JobID,
			TriggerTxHash:   log.TxHash.Hex(),
			TriggeredAt: time.Now(),
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
