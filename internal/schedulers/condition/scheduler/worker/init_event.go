package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// EventWorker monitors blockchain events for specific contracts
type EventWorker struct {
	EventWorkerData    *types.EventWorkerData
	ChainClient        *ethclient.Client
	Logger             logging.Logger
	Ctx                context.Context
	Cancel             context.CancelFunc
	IsActive           bool
	Mutex              sync.RWMutex
	LastBlock          uint64
	LastBlockTimestamp time.Time
	TriggerCallback    WorkerTriggerCallback // Callback to notify scheduler when event is detected
	CleanupCallback    WorkerCleanupCallback // Callback to clean up job data when worker stops
}

// Start begins the event worker's monitoring loop
func (w *EventWorker) Start() {
	startTime := time.Now()

	w.Mutex.Lock()
	w.IsActive = true
	w.Mutex.Unlock()

	// Track worker start
	metrics.TrackWorkerStart(fmt.Sprintf("%d", w.EventWorkerData.JobID))

	// Get current block number
	currentBlock, err := w.ChainClient.BlockNumber(w.Ctx)
	if err != nil {
		w.Logger.Error("Failed to get current block number", "error", err)
		return
	}
	
	// Start from a few blocks back to catch recent events
	// This helps catch events that might have been missed during worker startup
	// Using smaller lookback for Alchemy free tier (max 10 blocks per query)
	lookbackBlocks := uint64(100) // Look back 100 blocks (~5 minutes on most chains)
	if currentBlock > lookbackBlocks {
		w.LastBlock = currentBlock - lookbackBlocks
	} else {
		w.LastBlock = 0 // Start from genesis if less than lookback blocks exist
	}
	
	w.Logger.Info("Event worker will scan from historical block",
		"job_id", w.EventWorkerData.JobID,
		"current_block", currentBlock,
		"starting_from_block", w.LastBlock,
		"lookback_blocks", lookbackBlocks,
	)

	w.Logger.Info("Starting event worker",
		"job_id", w.EventWorkerData.JobID,
		"chain_id", w.EventWorkerData.TriggerChainID,
		"contract", w.EventWorkerData.TriggerContractAddress,
		"event", w.EventWorkerData.TriggerEvent,
		"current_block", currentBlock,
		"expiration_time", w.EventWorkerData.ExpirationTime,
		"filter_enabled", w.EventWorkerData.EventFilterParaName != "" && w.EventWorkerData.EventFilterValue != "",
		"filter_param", w.EventWorkerData.EventFilterParaName,
		"filter_value", w.EventWorkerData.EventFilterValue,
	)

	contractAddr := common.HexToAddress(w.EventWorkerData.TriggerContractAddress)
	eventSig := crypto.Keccak256Hash([]byte(w.EventWorkerData.TriggerEvent))

	ticker := time.NewTicker(EventPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.Ctx.Done():
			stopTime := time.Now()
			duration := stopTime.Sub(startTime)

			w.Logger.Info("Event worker stopped",
				"job_id", w.EventWorkerData.JobID,
				"runtime", duration,
				"final_block", w.LastBlock,
			)
			metrics.JobsCompleted.WithLabelValues("success").Inc()
			return
		case <-ticker.C:
			// Check if job has expired
			if time.Now().After(w.EventWorkerData.ExpirationTime) {
				w.Logger.Info("Job has expired, stopping worker",
					"job_id", w.EventWorkerData.JobID,
					"expiration_time", w.EventWorkerData.ExpirationTime,
				)
				go w.Stop() // Stop in a goroutine to avoid deadlock
				return
			}

			if err := w.checkForEvents(contractAddr, eventSig); err != nil {
				w.Logger.Error("Error checking for events", "job_id", w.EventWorkerData.JobID, "error", err)
				metrics.JobsCompleted.WithLabelValues("failed").Inc()
			}
		}
	}
}

// Stop gracefully stops the event worker
func (w *EventWorker) Stop() {
	w.Mutex.Lock()
	defer w.Mutex.Unlock()

	if w.IsActive {
		w.Cancel()
		w.IsActive = false

		// Track worker stop
		metrics.TrackWorkerStop(fmt.Sprintf("%d", w.EventWorkerData.JobID))

		// Clean up job data from scheduler store
		if w.CleanupCallback != nil {
			if err := w.CleanupCallback(w.EventWorkerData.JobID.ToBigInt()); err != nil {
				w.Logger.Error("Failed to clean up job data",
					"job_id", w.EventWorkerData.JobID,
					"error", err)
			}
		}

		w.Logger.Info("Event worker stopped", "job_id", w.EventWorkerData.JobID)
	}
}

// IsRunning returns whether the worker is currently running
func (w *EventWorker) IsRunning() bool {
	w.Mutex.RLock()
	defer w.Mutex.RUnlock()
	return w.IsActive
}
