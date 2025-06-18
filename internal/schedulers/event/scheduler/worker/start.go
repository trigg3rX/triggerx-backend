package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/event/metrics"
	schedulerTypes "github.com/trigg3rX/triggerx-backend/internal/schedulers/event/scheduler/types"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	commonTypes "github.com/trigg3rX/triggerx-backend/pkg/types"
)

// JobWorker represents an individual worker watching a specific job
type EventWorker struct {
	Job          *commonTypes.ScheduleEventJobData
	Client       *ethclient.Client
	Logger       logging.Logger
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

	// Track worker start in metrics
	metrics.TrackWorkerStart(fmt.Sprintf("%d", w.Job.JobID))

	// Try to acquire performer lock
	lockAcquired := false

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

			w.Logger.Info("Job worker stopped",
				"job_id", w.Job.JobID,
				"runtime", duration,
				"final_block", w.LastBlock,
			)
			return
		case <-ticker.C:
			if err := w.checkForEvents(); err != nil {
				w.Logger.Error("Error checking for events", "job_id", w.Job.JobID, "error", err)
				metrics.TrackJobCompleted("failed")
				metrics.TrackWorkerError(fmt.Sprintf("%d", w.Job.JobID), "event_check_error")
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

		// Track worker stop in metrics
		metrics.TrackWorkerStop(fmt.Sprintf("%d", w.Job.JobID))

		w.Logger.Info("Job worker stopped", "job_id", w.Job.JobID)
	}
}

// IsRunning returns whether the worker is currently running
func (w *EventWorker) IsRunning() bool {
	w.Mutex.RLock()
	defer w.Mutex.RUnlock()
	return w.IsActive
}
