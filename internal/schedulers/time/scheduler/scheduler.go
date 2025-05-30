package scheduler

import (
	"context"
	"sort"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/client"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/parser"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

const (
	pollInterval    = 30 * time.Second // Poll every 30 seconds
	executionWindow = 5 * time.Minute  // Look ahead 5 minutes
	workerPoolSize  = 10               // Number of concurrent workers
	batchSize       = 50               // Process jobs in batches
)

type TimeBasedScheduler struct {
	ctx        context.Context
	cancel     context.CancelFunc
	logger     logging.Logger
	workerPool chan struct{}
	activeJobs map[int64]*types.TimeJobData
	jobQueue   chan *types.TimeJobData
	dbClient   *client.DBServerClient
	metrics    *metrics.Collector
	managerID  string
}

// NewTimeBasedScheduler creates a new instance of TimeBasedScheduler
func NewTimeBasedScheduler(managerID string, logger logging.Logger, dbClient *client.DBServerClient) (*TimeBasedScheduler, error) {
	ctx, cancel := context.WithCancel(context.Background())

	scheduler := &TimeBasedScheduler{
		ctx:        ctx,
		cancel:     cancel,
		logger:     logger,
		workerPool: make(chan struct{}, workerPoolSize),
		activeJobs: make(map[int64]*types.TimeJobData),
		jobQueue:   make(chan *types.TimeJobData, 100),
		dbClient:   dbClient,
		metrics:    metrics.NewCollector(),
		managerID:  managerID,
	}

	// Start metrics collection
	scheduler.metrics.Start()

	// Start the worker pool
	for i := 0; i < workerPoolSize; i++ {
		go scheduler.worker()
	}

	return scheduler, nil
}

// Start begins the scheduler's main polling and execution loop
func (s *TimeBasedScheduler) Start(ctx context.Context) {
	s.logger.Info("Starting time-based scheduler", "manager_id", s.managerID)

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Scheduler context cancelled, stopping")
			return
		case <-s.ctx.Done():
			s.logger.Info("Scheduler stopped")
			return
		case <-ticker.C:
			s.pollAndScheduleJobs()
		}
	}
}

// pollAndScheduleJobs fetches jobs from database and schedules them for execution
func (s *TimeBasedScheduler) pollAndScheduleJobs() {
	jobs, err := s.dbClient.GetTimeBasedJobs()
	if err != nil {
		s.logger.Errorf("Failed to fetch time-based jobs: %v", err)
		metrics.JobsFailed.Inc()
		return
	}

	if len(jobs) == 0 {
		s.logger.Debug("No jobs found for execution")
		return
	}

	s.logger.Infof("Found %d jobs to process", len(jobs))
	metrics.JobsScheduled.Set(float64(len(jobs)))

	// Sort jobs by execution time (earliest first)
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].NextExecutionTimestamp.Before(jobs[j].NextExecutionTimestamp)
	})

	// Process jobs in batches
	now := time.Now()
	executionWindow := now.Add(executionWindow)

	for i := 0; i < len(jobs); i += batchSize {
		end := i + batchSize
		if end > len(jobs) {
			end = len(jobs)
		}

		batch := jobs[i:end]
		s.processBatch(batch, now, executionWindow)
	}
}

// processBatch processes a batch of jobs
func (s *TimeBasedScheduler) processBatch(jobs []types.TimeJobData, now, executionWindow time.Time) {
	for _, job := range jobs {
		// Check if job is due for execution (within execution window)
		if job.NextExecutionTimestamp.After(executionWindow) {
			continue // Job is not due yet
		}

		if job.NextExecutionTimestamp.Before(now.Add(-1 * time.Minute)) {
			s.logger.Warnf("Job %d is overdue by %v", job.JobID, now.Sub(job.NextExecutionTimestamp))
		}

		// Add job to execution queue
		select {
		case s.jobQueue <- &job:
			metrics.JobsRunning.Inc()
			s.logger.Debugf("Queued job %d for execution", job.JobID)
		default:
			s.logger.Warnf("Job queue is full, skipping job %d", job.JobID)
			metrics.JobsFailed.Inc()
		}
	}
}

// worker processes jobs from the job queue
func (s *TimeBasedScheduler) worker() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case job := <-s.jobQueue:
			s.workerPool <- struct{}{} // Acquire worker slot
			s.executeJob(job)
			<-s.workerPool // Release worker slot
		}
	}
}

// executeJob executes a single job and updates its next execution time
func (s *TimeBasedScheduler) executeJob(job *types.TimeJobData) {
	startTime := time.Now()
	s.logger.Infof("Executing time-based job %d (type: %s)", job.JobID, job.ScheduleType)

	// Update job status to running
	if err := s.dbClient.UpdateJobStatus(job.JobID, true); err != nil {
		s.logger.Errorf("Failed to update job %d status to running: %v", job.JobID, err)
	}

	// TODO: Implement actual job execution logic here
	// This would involve:
	// 1. Calling the target function/webhook
	// 2. Handling the response
	// 3. Managing retries on failure

	// Simulate job execution for now
	executionSuccess := s.simulateJobExecution(job)

	// Calculate next execution time
	nextExecution, err := parser.CalculateNextExecutionTime(
		job.ScheduleType,
		job.TimeInterval,
		job.CronExpression,
		job.SpecificSchedule,
		job.Timezone,
	)
	if err != nil {
		s.logger.Errorf("Failed to calculate next execution time for job %d: %v", job.JobID, err)
		metrics.JobsFailed.Inc()
		return
	}

	// Update next execution time in database
	if err := s.dbClient.UpdateJobNextExecution(job.JobID, nextExecution); err != nil {
		s.logger.Errorf("Failed to update next execution time for job %d: %v", job.JobID, err)
		metrics.JobsFailed.Inc()
		return
	}

	// Update job status to completed
	if err := s.dbClient.UpdateJobStatus(job.JobID, false); err != nil {
		s.logger.Errorf("Failed to update job %d status to completed: %v", job.JobID, err)
	}

	duration := time.Since(startTime)

	if executionSuccess {
		metrics.JobsCompleted.Inc()
		s.logger.Infof("Completed job %d in %v, next execution at %v",
			job.JobID, duration, nextExecution)
	} else {
		metrics.JobsFailed.Inc()
		s.logger.Errorf("Failed to execute job %d after %v", job.JobID, duration)
	}
}

// simulateJobExecution simulates job execution (replace with actual implementation)
func (s *TimeBasedScheduler) simulateJobExecution(job *types.TimeJobData) bool {
	// TODO: Replace this with actual job execution logic
	// For now, simulate execution with a small delay
	time.Sleep(100 * time.Millisecond)

	// Simulate 95% success rate
	return time.Now().UnixNano()%100 < 95
}

// Stop gracefully stops the scheduler
func (s *TimeBasedScheduler) Stop() {
	s.logger.Info("Stopping time-based scheduler")
	s.cancel()

	// Close job queue
	close(s.jobQueue)

	// Wait for workers to finish (with timeout)
	timeout := time.After(30 * time.Second)
	for len(s.workerPool) < workerPoolSize {
		select {
		case <-timeout:
			s.logger.Warn("Timeout waiting for workers to finish")
			return
		case <-time.After(100 * time.Millisecond):
			// Continue waiting
		}
	}

	s.logger.Info("Time-based scheduler stopped")
}

// GetStats returns current scheduler statistics
func (s *TimeBasedScheduler) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"manager_id":   s.managerID,
		"active_jobs":  len(s.activeJobs),
		"queue_length": len(s.jobQueue),
		"worker_pool":  len(s.workerPool),
		"max_workers":  workerPoolSize,
	}
}
