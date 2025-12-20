// init_condition.go
package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/metrics"
	httppkg "github.com/trigg3rX/triggerx-backend/pkg/http"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// ConditionWorker represents an individual worker monitoring a specific condition
type ConditionWorker struct {
	ConditionWorkerData *types.ConditionWorkerData
	Logger              observability.Logger
	HttpClient          *httppkg.HTTPClient
	Ctx                 context.Context
	Cancel              context.CancelFunc
	IsActive            bool
	Mutex               sync.RWMutex
	LastValue           float64
	LastCheckTimestamp  time.Time
	ConditionMet        int64 // Count of consecutive condition met checks
	TriggerCallback     WorkerTriggerCallback
	CleanupCallback     WorkerCleanupCallback
}

// Start begins the condition worker's monitoring loop
func (w *ConditionWorker) Start(ctx context.Context) {
	startTime := time.Now()

	w.Mutex.Lock()
	w.IsActive = true
	w.Mutex.Unlock()

	metrics.TrackWorkerStart(fmt.Sprintf("%d", w.ConditionWorkerData.JobID))

	w.Logger.Info(ctx, "Starting condition worker",
		observability.String("job_id", w.ConditionWorkerData.JobID.String()),
		observability.String("condition_type", w.ConditionWorkerData.ConditionType),
		observability.String("value_source", w.ConditionWorkerData.ValueSourceUrl),
		observability.String("selected_key_route", w.ConditionWorkerData.SelectedKeyRoute),
		observability.Float64("upper_limit", w.ConditionWorkerData.UpperLimit),
		observability.Float64("lower_limit", w.ConditionWorkerData.LowerLimit),
		observability.Time("expiration_time", w.ConditionWorkerData.ExpirationTime),
	)

	ticker := time.NewTicker(ConditionPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.Ctx.Done():
			stopTime := time.Now()
			duration := stopTime.Sub(startTime)

			w.Logger.Info(ctx, "Condition worker stopped",
				observability.String("job_id", w.ConditionWorkerData.JobID.String()),
				observability.Duration("runtime", duration),
				observability.Float64("last_value", w.LastValue),
				observability.Int64("condition_met_count", w.ConditionMet),
			)
			metrics.JobsCompleted.WithLabelValues("success").Inc()
			return
		case <-ticker.C:
			if time.Now().After(w.ConditionWorkerData.ExpirationTime) {
				w.Logger.Info(ctx, "Job has expired, stopping worker",
					observability.String("job_id", w.ConditionWorkerData.JobID.String()),
					observability.Time("expiration_time", w.ConditionWorkerData.ExpirationTime),
				)
				go w.Stop(ctx)
				return
			}

			if err := w.checkCondition(ctx); err != nil {
				w.Logger.Error(ctx, "Error checking condition",
					observability.String("job_id", w.ConditionWorkerData.JobID.String()),
					observability.Error(err))
				metrics.JobsCompleted.WithLabelValues("failed").Inc()
			}
		}
	}
}

// Stop gracefully stops the condition worker
func (w *ConditionWorker) Stop(ctx context.Context) {
	w.Mutex.Lock()
	defer w.Mutex.Unlock()

	if w.IsActive {
		w.Cancel()
		w.IsActive = false

		metrics.TrackWorkerStop(fmt.Sprintf("%d", w.ConditionWorkerData.JobID))

		if w.CleanupCallback != nil {
			if err := w.CleanupCallback(ctx, w.ConditionWorkerData.JobID.ToBigInt()); err != nil {
				w.Logger.Error(ctx, "Failed to clean up job data",
					observability.String("job_id", w.ConditionWorkerData.JobID.String()),
					observability.Error(err))
			}
		}

		w.Logger.Info(ctx, "Condition worker stopped",
			observability.String("job_id", w.ConditionWorkerData.JobID.String()))
	}
}

// IsRunning returns whether the worker is currently running
func (w *ConditionWorker) IsRunning() bool {
	w.Mutex.RLock()
	defer w.Mutex.RUnlock()
	return w.IsActive
}
