package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// ConditionWorker represents an individual worker monitoring a specific condition
type ConditionWorker struct {
	ConditionWorkerData *types.ConditionWorkerData
	Logger          logging.Logger
	HttpClient      *retry.HTTPClient
	Ctx             context.Context
	Cancel          context.CancelFunc
	IsActive        bool
	Mutex           sync.RWMutex
	LastValue       float64
	LastCheckTimestamp time.Time
	ConditionMet    int64 // Count of consecutive condition met checks
	TriggerCallback WorkerTriggerCallback // Callback to notify scheduler when condition is satisfied
	CleanupCallback WorkerCleanupCallback // Callback to clean up job data when worker stops
}

// Start begins the condition worker's monitoring loop
func (w *ConditionWorker) Start() {
	startTime := time.Now()

	w.Mutex.Lock()
	w.IsActive = true
	w.Mutex.Unlock()

	// Track worker start
	metrics.TrackWorkerStart(fmt.Sprintf("%d", w.ConditionWorkerData.JobID))

	w.Logger.Info("Starting condition worker",
		"job_id", w.ConditionWorkerData.JobID,
		"condition_type", w.ConditionWorkerData.ConditionType,
		"value_source", w.ConditionWorkerData.ValueSourceUrl,
		"upper_limit", w.ConditionWorkerData.UpperLimit,
		"lower_limit", w.ConditionWorkerData.LowerLimit,
		"expiration_time", w.ConditionWorkerData.ExpirationTime,
	)

	ticker := time.NewTicker(ConditionPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.Ctx.Done():
			stopTime := time.Now()
			duration := stopTime.Sub(startTime)

			w.Logger.Info("Condition worker stopped",
				"job_id", w.ConditionWorkerData.JobID,
				"runtime", duration,
				"last_value", w.LastValue,
				"condition_met_count", w.ConditionMet,
			)
			metrics.JobsCompleted.WithLabelValues("success").Inc()
			return
		case <-ticker.C:
			// Check if job has expired
			if time.Now().After(w.ConditionWorkerData.ExpirationTime) {
				w.Logger.Info("Job has expired, stopping worker",
					"job_id", w.ConditionWorkerData.JobID,
					"expiration_time", w.ConditionWorkerData.ExpirationTime,
				)
				go w.Stop() // Stop in a goroutine to avoid deadlock
				return
			}

			if err := w.checkCondition(); err != nil {
				w.Logger.Error("Error checking condition", "job_id", w.ConditionWorkerData.JobID, "error", err)
				metrics.JobsCompleted.WithLabelValues("failed").Inc()
			}
		}
	}
}

// Stop gracefully stops the condition worker
func (w *ConditionWorker) Stop() {
	w.Mutex.Lock()
	defer w.Mutex.Unlock()

	if w.IsActive {
		w.Cancel()
		w.IsActive = false

		// Track worker stop
		metrics.TrackWorkerStop(fmt.Sprintf("%d", w.ConditionWorkerData.JobID))

		// Clean up job data from scheduler store
		if w.CleanupCallback != nil {
			if err := w.CleanupCallback(w.ConditionWorkerData.JobID); err != nil {
				w.Logger.Error("Failed to clean up job data",
					"job_id", w.ConditionWorkerData.JobID,
					"error", err)
			}
		}

		w.Logger.Info("Condition worker stopped", "job_id", w.ConditionWorkerData.JobID)
	}
}

// IsRunning returns whether the worker is currently running
func (w *ConditionWorker) IsRunning() bool {
	w.Mutex.RLock()
	defer w.Mutex.RUnlock()
	return w.IsActive
}
