package worker

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"

	schedulerTypes "github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/scheduler/types"
	commonTypes "github.com/trigg3rX/triggerx-backend/pkg/types"
)

// ConditionWorker represents an individual worker monitoring a specific condition
type ConditionWorker struct {
	Job          *commonTypes.ScheduleConditionJobData
	Logger       logging.Logger
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

	// Track worker start
	metrics.TrackWorkerStart(fmt.Sprintf("%d", w.Job.JobID))

	// Try to acquire performer lock
	lockAcquired := false

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

			w.Logger.Info("Condition worker stopped",
				"job_id", w.Job.JobID,
				"runtime", duration,
				"last_value", w.LastValue,
				"condition_met_count", w.ConditionMet,
			)
			metrics.JobsCompleted.WithLabelValues("success").Inc()
			return
		case <-ticker.C:
			if err := w.checkCondition(); err != nil {
				w.Logger.Error("Error checking condition", "job_id", w.Job.JobID, "error", err)
				metrics.JobsCompleted.WithLabelValues("failed").Inc()
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

		// Track worker stop
		metrics.TrackWorkerStop(fmt.Sprintf("%d", w.Job.JobID))

		w.Logger.Info("Condition worker stopped", "job_id", w.Job.JobID)
	}
}

// IsRunning returns whether the worker is currently running
func (w *ConditionWorker) IsRunning() bool {
	w.Mutex.RLock()
	defer w.Mutex.RUnlock()
	return w.IsActive
}
