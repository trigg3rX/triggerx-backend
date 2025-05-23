package scheduler

import (
	"context"
	"sync"

	"github.com/robfig/cron/v3"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// TimeBasedWorker implements the Worker interface for time-based jobs
type TimeBasedWorker struct {
	jobData     types.HandleCreateJobData
	schedule    string
	scheduler   *TimeBasedScheduler
	cronEntryID cron.EntryID
	stopChan    chan struct{}
	mu          sync.RWMutex
}

// NewTimeBasedWorker creates a new time-based worker
func NewTimeBasedWorker(jobData types.HandleCreateJobData, schedule string, scheduler *TimeBasedScheduler) *TimeBasedWorker {
	return &TimeBasedWorker{
		jobData:   jobData,
		schedule:  schedule,
		scheduler: scheduler,
		stopChan:  make(chan struct{}),
	}
}

// Start starts the time-based worker
func (w *TimeBasedWorker) Start(ctx context.Context) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Add the job to the cron scheduler
	entryID, err := w.scheduler.GetCronScheduler().AddFunc(w.schedule, w.execute)
	if err != nil {
		w.scheduler.Logger().Errorf("Failed to schedule job %d: %v", w.jobData.JobID, err)
		return
	}
	w.cronEntryID = entryID

	// Wait for context cancellation or stop signal
	go func() {
		select {
		case <-ctx.Done():
			w.Stop()
		case <-w.stopChan:
			return
		}
	}()
}

// Stop stops the time-based worker
func (w *TimeBasedWorker) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.cronEntryID != 0 {
		w.scheduler.GetCronScheduler().Remove(w.cronEntryID)
		w.cronEntryID = 0
	}

	close(w.stopChan)
}

// GetJobData returns the job data
func (w *TimeBasedWorker) GetJobData() types.HandleCreateJobData {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.jobData
}

// execute executes the time-based job
func (w *TimeBasedWorker) execute() {
	w.mu.RLock()
	defer w.mu.RUnlock()

	logger := w.scheduler.Logger()
	logger.Infof("Executing time-based job %d", w.jobData.JobID)

	// TODO: Implement the actual job execution logic here
	// 1. Call the target function
	// 2. Update job status in the database
	// 3. Handle any errors
	// 4. Update last executed time

	logger.Infof("Completed time-based job %d", w.jobData.JobID)
}
