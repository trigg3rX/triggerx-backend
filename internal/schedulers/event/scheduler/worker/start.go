package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/trigg3rX/triggerx-backend/internal/cache"
	redisx "github.com/trigg3rX/triggerx-backend/internal/redis"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/event/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	commonTypes "github.com/trigg3rX/triggerx-backend/pkg/types"
	schedulerTypes "github.com/trigg3rX/triggerx-backend/internal/schedulers/event/scheduler/types"
)

// JobWorker represents an individual worker watching a specific job
type EventWorker struct {
	Job          *commonTypes.EventJobData
	Client       *ethclient.Client
	Logger       logging.Logger
	Cache        cache.Cache
	Ctx          context.Context
	Cancel       context.CancelFunc
	EventSig     common.Hash
	ContractAddr common.Address
	LastBlock    uint64
	IsActive     bool
	Mutex        sync.RWMutex
	ManagerID    string
}

// start begins the job worker's event monitoring loop
func (w *EventWorker) Start() {
	startTime := time.Now()

	w.Mutex.Lock()
	w.IsActive = true
	w.Mutex.Unlock()

	// Try to acquire performer lock
	lockKey := fmt.Sprintf("event_job_%d_%s", w.Job.JobID, w.Job.TriggerChainID)
	lockAcquired := false

	if w.Cache != nil {
		acquired, err := w.Cache.AcquirePerformerLock(lockKey, schedulerTypes.PerformerLockTTL)
		if err != nil {
			w.Logger.Warnf("Failed to acquire performer lock for job %d: %v", w.Job.JobID, err)

			// Add lock failure event to Redis stream
			if redisx.IsAvailable() {
				lockFailureEvent := map[string]interface{}{
					"event_type":       "worker_lock_failed",
					"job_id":           w.Job.JobID,
					"manager_id":       w.ManagerID,
					"lock_key":         lockKey,
					"error":            err.Error(),
					"trigger_chain_id": w.Job.TriggerChainID,
					"contract_address": w.Job.TriggerContractAddress,
					"trigger_event":    w.Job.TriggerEvent,
					"failed_at":        startTime.Unix(),
				}
				err := redisx.AddJobToStream(redisx.JobsRetryEventStream, lockFailureEvent)
				if err != nil {
					w.Logger.Warnf("Failed to add worker lock failure event to Redis stream: %v", err)
				}
			}
		} else if !acquired {
			w.Logger.Warnf("Job %d is already being monitored by another worker, stopping", w.Job.JobID)

			// Add lock conflict event to Redis stream
			if redisx.IsAvailable() {
				lockConflictEvent := map[string]interface{}{
					"event_type":       "worker_lock_conflict",
					"job_id":           w.Job.JobID,
					"manager_id":       w.ManagerID,
					"lock_key":         lockKey,
					"trigger_chain_id": w.Job.TriggerChainID,
					"contract_address": w.Job.TriggerContractAddress,
					"trigger_event":    w.Job.TriggerEvent,
					"conflict_at":      startTime.Unix(),
				}
				err := redisx.AddJobToStream(redisx.JobsRetryEventStream, lockConflictEvent)
				if err != nil {
					w.Logger.Warnf("Failed to add worker lock conflict event to Redis stream: %v", err)
				}
			}
			return
		} else {
			lockAcquired = true
		}

		defer func() {
			if lockAcquired {
				if err := w.Cache.ReleasePerformerLock(lockKey); err != nil {
					w.Logger.Warnf("Failed to release performer lock for job %d: %v", w.Job.JobID, err)
				}
			}
		}()
	}

	// Add worker start event to Redis stream
	if redisx.IsAvailable() {
		workerStartEvent := map[string]interface{}{
			"event_type":               "worker_started",
			"job_id":                   w.Job.JobID,
			"manager_id":               w.ManagerID,
			"trigger_chain_id":         w.Job.TriggerChainID,
			"trigger_contract_address": w.Job.TriggerContractAddress,
			"trigger_event":            w.Job.TriggerEvent,
			"target_chain_id":          w.Job.TargetChainID,
			"target_contract_address":  w.Job.TargetContractAddress,
			"target_function":          w.Job.TargetFunction,
			"starting_block":           w.LastBlock,
			"lock_acquired":            lockAcquired,
			"cache_available":          w.Cache != nil,
			"started_at":               startTime.Unix(),
			"poll_interval_seconds":    schedulerTypes.PollInterval.Seconds(),
		}

		if err := redisx.AddJobToStream(redisx.JobsReadyEventStream, workerStartEvent); err != nil {
			w.Logger.Warnf("Failed to add worker start event to Redis stream: %v", err)
		}
	}

	w.Logger.Info("Starting job worker",
		"job_id", w.Job.JobID,
		"contract", w.Job.TriggerContractAddress,
		"event", w.Job.TriggerEvent,
		"starting_block", w.LastBlock,
		"lock_acquired", lockAcquired,
	)

	ticker := time.NewTicker(schedulerTypes.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.Ctx.Done():
			stopTime := time.Now()
			duration := stopTime.Sub(startTime)

			// Add worker stop event to Redis stream
			if redisx.IsAvailable() {
				workerStopEvent := map[string]interface{}{
					"event_type":       "worker_stopped",
					"job_id":           w.Job.JobID,
					"manager_id":       w.ManagerID,
					"trigger_chain_id": w.Job.TriggerChainID,
					"contract_address": w.Job.TriggerContractAddress,
					"trigger_event":    w.Job.TriggerEvent,
					"final_block":      w.LastBlock,
					"runtime_seconds":  duration.Seconds(),
					"stopped_at":       stopTime.Unix(),
					"graceful_stop":    true,
				}

				if err := redisx.AddJobToStream(redisx.JobsReadyEventStream, workerStopEvent); err != nil {
					w.Logger.Warnf("Failed to add worker stop event to Redis stream: %v", err)
				}
			}

			w.Logger.Info("Job worker stopped",
				"job_id", w.Job.JobID,
				"runtime", duration,
				"final_block", w.LastBlock,
			)
			return
		case <-ticker.C:
			if err := w.checkForEvents(); err != nil {
				w.Logger.Error("Error checking for events", "job_id", w.Job.JobID, "error", err)
				metrics.JobsFailed.Inc()

				// Add error event to Redis stream
				if redisx.IsAvailable() {
					errorEvent := map[string]interface{}{
						"event_type":       "worker_error",
						"job_id":           w.Job.JobID,
						"manager_id":       w.ManagerID,
						"error":            err.Error(),
						"trigger_chain_id": w.Job.TriggerChainID,
						"contract_address": w.Job.TriggerContractAddress,
						"trigger_event":    w.Job.TriggerEvent,
						"current_block":    w.LastBlock,
						"error_at":         time.Now().Unix(),
					}
					if err := redisx.AddJobToStream(redisx.JobsRetryEventStream, errorEvent); err != nil {
						w.Logger.Warnf("Failed to add worker error event to Redis stream: %v", err)
					}
				}
			}
		}
	}
}

// stop gracefully stops the job worker
func (w *EventWorker) Stop() {
	w.Mutex.Lock()
	defer w.Mutex.Unlock()

	if w.IsActive {
		w.Cancel()
		w.IsActive = false
		w.Logger.Info("Job worker stopped", "job_id", w.Job.JobID)
	}
}

// IsRunning returns whether the worker is currently running
func (w *EventWorker) IsRunning() bool {
	w.Mutex.RLock()
	defer w.Mutex.RUnlock()
	return w.IsActive
}
