package worker

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/metrics"
	nodeclient "github.com/trigg3rX/triggerx-backend/pkg/client/nodeclient"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// EventWorker monitors blockchain events for specific contracts
type EventWorker struct {
	EventWorkerData    *types.EventWorkerData
	ChainClient        *nodeclient.NodeClient
	Logger             observability.Logger
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
func (w *EventWorker) Start(ctx context.Context) {
	startTime := time.Now()

	w.Mutex.Lock()
	w.IsActive = true
	w.Mutex.Unlock()

	// Track worker start
	metrics.TrackWorkerStart(fmt.Sprintf("%d", w.EventWorkerData.JobID))

	// Get current block number
	blockHex, err := w.ChainClient.EthBlockNumber(w.Ctx)
	if err != nil {
		w.Logger.Error(ctx, "Failed to get current block number", observability.Error(err))
		return
	}
	currentBlock, err := hexToUint64(blockHex)
	if err != nil {
		w.Logger.Error(ctx, "Failed to parse block number", observability.Error(err))
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

	w.Logger.Info(ctx, "Event worker will scan from historical block",
		observability.String("job_id", w.EventWorkerData.JobID.String()),
		observability.Uint64("current_block", currentBlock),
		observability.Uint64("starting_from_block", w.LastBlock),
		observability.Uint64("lookback_blocks", lookbackBlocks),
	)

	w.Logger.Info(ctx, "Starting event worker",
		observability.String("job_id", w.EventWorkerData.JobID.String()),
		observability.String("chain_id", w.EventWorkerData.TriggerChainID),
		observability.String("contract", w.EventWorkerData.TriggerContractAddress),
		observability.String("event", w.EventWorkerData.TriggerEvent),
		observability.Uint64("current_block", currentBlock),
		observability.Time("expiration_time", w.EventWorkerData.ExpirationTime),
		observability.Bool("filter_enabled", w.EventWorkerData.EventFilterParaName != "" && w.EventWorkerData.EventFilterValue != ""),
		observability.String("filter_param", w.EventWorkerData.EventFilterParaName),
		observability.String("filter_value", w.EventWorkerData.EventFilterValue),
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

			w.Logger.Info(ctx, "Event worker stopped",
				observability.String("job_id", w.EventWorkerData.JobID.String()),
				observability.Duration("runtime", duration),
				observability.Uint64("final_block", w.LastBlock),
			)
			metrics.TrackJobCompleted("success")
			return
		case <-ticker.C:
			// Check if job has expired
			if time.Now().After(w.EventWorkerData.ExpirationTime) {
				w.Logger.Info(ctx, "Job has expired, stopping worker",
					observability.String("job_id", w.EventWorkerData.JobID.String()),
					observability.Time("expiration_time", w.EventWorkerData.ExpirationTime),
				)
				go w.Stop(ctx) // Stop in a goroutine to avoid deadlock
				return
			}

			if err := w.checkForEvents(ctx, contractAddr, eventSig); err != nil {
				w.Logger.Error(ctx, "Error checking for events",
					observability.String("job_id", w.EventWorkerData.JobID.String()),
					observability.Error(err))
				metrics.TrackJobCompleted("failed")
			}
		}
	}
}

// Stop gracefully stops the event worker
func (w *EventWorker) Stop(ctx context.Context) {
	w.Mutex.Lock()
	defer w.Mutex.Unlock()

	if w.IsActive {
		w.Cancel()
		w.IsActive = false

		// Track worker stop
		metrics.TrackWorkerStop(fmt.Sprintf("%d", w.EventWorkerData.JobID))

		// Clean up job data from scheduler store
		if w.CleanupCallback != nil {
			if err := w.CleanupCallback(ctx, w.EventWorkerData.JobID.ToBigInt()); err != nil {
				w.Logger.Error(ctx, "Failed to clean up job data",
					observability.String("job_id", w.EventWorkerData.JobID.String()),
					observability.Error(err))
			}
		}

		w.Logger.Info(ctx, "Event worker stopped",
			observability.String("job_id", w.EventWorkerData.JobID.String()))
	}
}

// IsRunning returns whether the worker is currently running
func (w *EventWorker) IsRunning() bool {
	w.Mutex.RLock()
	defer w.Mutex.RUnlock()
	return w.IsActive
}

// hexToUint64 converts a hex string (with or without 0x prefix) to uint64
func hexToUint64(hexStr string) (uint64, error) {
	// Remove 0x prefix if present
	if len(hexStr) >= 2 && hexStr[:2] == "0x" {
		hexStr = hexStr[2:]
	}
	return strconv.ParseUint(hexStr, 16, 64)
}
