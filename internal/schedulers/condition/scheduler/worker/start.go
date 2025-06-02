package worker

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/cache"
	redisx "github.com/trigg3rX/triggerx-backend/internal/redis"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"

	commonTypes "github.com/trigg3rX/triggerx-backend/pkg/types"
	schedulerTypes "github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/scheduler/types"
)

// ConditionWorker represents an individual worker monitoring a specific condition
type ConditionWorker struct {
	Job          *commonTypes.ConditionJobData
	Logger       logging.Logger
	Cache        cache.Cache
	HttpClient   *http.Client
	Ctx          context.Context
	Cancel       context.CancelFunc
	IsActive     bool
	Mutex        sync.RWMutex
	LastValue    float64
	LastCheck    time.Time
	ConditionMet int64 // Count of consecutive condition met checks
	ManagerID    string
}

// start begins the condition worker's monitoring loop
func (w *ConditionWorker) Start() {
	startTime := time.Now()

	w.Mutex.Lock()
	w.IsActive = true
	w.Mutex.Unlock()

	// Try to acquire performer lock
	lockKey := fmt.Sprintf("condition_job_%d", w.Job.JobID)
	lockAcquired := false

	if w.Cache != nil {
		acquired, err := w.Cache.AcquirePerformerLock(lockKey, schedulerTypes.PerformerLockTTL)
		if err != nil {
			w.Logger.Warnf("Failed to acquire performer lock for condition job %d: %v", w.Job.JobID, err)

			// Add lock failure event to Redis stream
			if redisx.IsAvailable() {
				lockFailureEvent := map[string]interface{}{
					"event_type":        "worker_lock_failed",
					"job_id":            w.Job.JobID,
					"manager_id":        w.ManagerID,
					"lock_key":          lockKey,
					"error":             err.Error(),
					"condition_type":    w.Job.ConditionType,
					"value_source_type": w.Job.ValueSourceType,
					"value_source_url":  w.Job.ValueSourceUrl,
					"failed_at":         startTime.Unix(),
				}
				if err := redisx.AddJobToStream(redisx.JobsRetryConditionStream, lockFailureEvent); err != nil {
					w.Logger.Warnf("Failed to add worker lock failure event to Redis stream: %v", err)
				}
			}
		} else if !acquired {
			w.Logger.Warnf("Condition job %d is already being monitored by another worker, stopping", w.Job.JobID)

			// Add lock conflict event to Redis stream
			if redisx.IsAvailable() {
				lockConflictEvent := map[string]interface{}{
					"event_type":        "worker_lock_conflict",
					"job_id":            w.Job.JobID,
					"manager_id":        w.ManagerID,
					"lock_key":          lockKey,
					"condition_type":    w.Job.ConditionType,
					"value_source_type": w.Job.ValueSourceType,
					"value_source_url":  w.Job.ValueSourceUrl,
					"conflict_at":       startTime.Unix(),
				}
				if err := redisx.AddJobToStream(redisx.JobsRetryConditionStream, lockConflictEvent); err != nil {
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
					w.Logger.Warnf("Failed to release performer lock for condition job %d: %v", w.Job.JobID, err)
				}
			}
		}()
	}

	// Add worker start event to Redis stream
	if redisx.IsAvailable() {
		workerStartEvent := map[string]interface{}{
			"event_type":              "worker_started",
			"job_id":                  w.Job.JobID,
			"manager_id":              w.ManagerID,
			"condition_type":          w.Job.ConditionType,
			"upper_limit":             w.Job.UpperLimit,
			"lower_limit":             w.Job.LowerLimit,
			"value_source_type":       w.Job.ValueSourceType,
			"value_source_url":        w.Job.ValueSourceUrl,
			"target_chain_id":         w.Job.TargetChainID,
			"target_contract_address": w.Job.TargetContractAddress,
			"target_function":         w.Job.TargetFunction,
			"lock_acquired":           lockAcquired,
			"cache_available":         w.Cache != nil,
			"started_at":              startTime.Unix(),
			"poll_interval_seconds":   schedulerTypes.PollInterval.Seconds(),
		}

		if err := redisx.AddJobToStream(redisx.JobsReadyConditionStream, workerStartEvent); err != nil {
			w.Logger.Warnf("Failed to add worker start event to Redis stream: %v", err)
		}
	}

	w.Logger.Info("Starting condition worker",
		"job_id", w.Job.JobID,
		"condition_type", w.Job.ConditionType,
		"value_source", w.Job.ValueSourceUrl,
		"upper_limit", w.Job.UpperLimit,
		"lower_limit", w.Job.LowerLimit,
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
					"event_type":        "worker_stopped",
					"job_id":            w.Job.JobID,
					"manager_id":        w.ManagerID,
					"condition_type":    w.Job.ConditionType,
					"value_source_type": w.Job.ValueSourceType,
					"value_source_url":  w.Job.ValueSourceUrl,
					"last_value":        w.LastValue,
					"condition_met":     w.ConditionMet,
					"runtime_seconds":   duration.Seconds(),
					"stopped_at":        stopTime.Unix(),
					"graceful_stop":     true,
				}

				if err := redisx.AddJobToStream(redisx.JobsReadyConditionStream, workerStopEvent); err != nil {
					w.Logger.Warnf("Failed to add worker stop event to Redis stream: %v", err)
				}
			}

			w.Logger.Info("Condition worker stopped",
				"job_id", w.Job.JobID,
				"runtime", duration,
				"last_value", w.LastValue,
				"condition_met_count", w.ConditionMet,
			)
			return
		case <-ticker.C:
			if err := w.checkCondition(); err != nil {
				w.Logger.Error("Error checking condition", "job_id", w.Job.JobID, "error", err)
				metrics.JobsFailed.Inc()

				// Add error event to Redis stream
				if redisx.IsAvailable() {
					errorEvent := map[string]interface{}{
						"event_type":        "worker_error",
						"job_id":            w.Job.JobID,
						"manager_id":        w.ManagerID,
						"error":             err.Error(),
						"condition_type":    w.Job.ConditionType,
						"value_source_type": w.Job.ValueSourceType,
						"value_source_url":  w.Job.ValueSourceUrl,
						"last_value":        w.LastValue,
						"error_at":          time.Now().Unix(),
					}
					if err := redisx.AddJobToStream(redisx.JobsRetryConditionStream, errorEvent); err != nil {
						w.Logger.Warnf("Failed to add worker error event to Redis stream: %v", err)
					}
				}
			}
		}
	}
}

// stop gracefully stops the condition worker
func (w *ConditionWorker) Stop() {
	w.Mutex.Lock()
	defer w.Mutex.Unlock()

	if w.IsActive {
		w.Cancel()
		w.IsActive = false
		w.Logger.Info("Condition worker stopped", "job_id", w.Job.JobID)
	}
}

// IsRunning returns whether the worker is currently running
func (w *ConditionWorker) IsRunning() bool {
	w.Mutex.RLock()
	defer w.Mutex.RUnlock()
	return w.IsActive
}
