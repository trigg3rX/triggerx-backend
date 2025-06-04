package scheduler

import (
	"context"
	"sort"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/parser"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

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

	// Try to acquire performer lock to prevent duplicate execution
	lockAcquired := false

	// Update job status to running
	if err := s.dbClient.UpdateJobStatus(job.JobID, true); err != nil {
		s.logger.Errorf("Failed to update job %d status to running: %v", job.JobID, err)
	}

	// Create comprehensive job execution context for Redis streaming
	jobContext := map[string]interface{}{
		"event_type":               "job_started",
		"job_id":                   job.JobID,
		"schedule_type":            job.ScheduleType,
		"time_interval":            job.TimeInterval,
		"cron_expression":          job.CronExpression,
		"specific_schedule":        job.SpecificSchedule,
		"timezone":                 job.Timezone,
		"manager_id":               s.managerID,
		"lock_acquired":            lockAcquired,
		"scheduled_execution_time": job.NextExecutionTimestamp.Unix(),
		"actual_start_time":        startTime.Unix(),
		"delay_seconds":            startTime.Sub(job.NextExecutionTimestamp).Seconds(),
		"status":                   "processing",
	}

	// Execute the actual job
	executionSuccess := s.performJobExecution(job)

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

		// Add failure event to Redis stream
		failureContext := jobContext
		failureContext["event_type"] = "job_failed"
		failureContext["status"] = "failed"
		failureContext["error"] = err.Error()
		failureContext["error_type"] = "next_execution_calculation"
		failureContext["completed_at"] = time.Now().Unix()
		failureContext["duration_ms"] = time.Since(startTime).Milliseconds()

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

	// Create completion context for Redis streaming
	completionContext := jobContext
	completionContext["duration_ms"] = duration.Milliseconds()
	completionContext["next_execution"] = nextExecution.Unix()
	completionContext["completed_at"] = time.Now().Unix()
	completionContext["execution_success"] = executionSuccess

	if executionSuccess {
		metrics.JobsCompleted.Inc()
		s.logger.Infof("Completed job %d in %v, next execution at %v",
			job.JobID, duration, nextExecution)

		completionContext["event_type"] = "job_completed"
		completionContext["status"] = "completed"

	} else {
		metrics.JobsFailed.Inc()
		s.logger.Errorf("Failed to execute job %d after %v", job.JobID, duration)

		completionContext["event_type"] = "job_failed"
		completionContext["status"] = "failed"
		completionContext["error"] = "job execution failed"
		completionContext["error_type"] = "execution_failure"

	}
}

// performJobExecution handles the actual job execution logic
func (s *TimeBasedScheduler) performJobExecution(job *types.TimeJobData) bool {
	// TODO: Replace this with actual job execution logic
	// This would involve:
	// 1. Calling the target function/webhook
	// 2. Handling the response
	// 3. Managing retries on failure

	// Simulate job execution for now
	time.Sleep(100 * time.Millisecond)

	// Simulate 95% success rate
	return time.Now().UnixNano()%100 < 95
}

// Stop gracefully stops the scheduler
func (s *TimeBasedScheduler) Stop() {
	startTime := time.Now()
	s.logger.Info("Stopping time-based scheduler")

	// Capture statistics before shutdown
	activeJobsCount := len(s.activeJobs)
	queueLength := len(s.jobQueue)

	s.cancel()

	// Close job queue
	close(s.jobQueue)

	// Wait for workers to finish (with timeout)
	timeout := time.After(30 * time.Second)
	for len(s.workerPool) < s.maxWorkers {
		select {
		case <-timeout:
			s.logger.Warn("Timeout waiting for workers to finish")

			return
		case <-time.After(100 * time.Millisecond):
			// Continue waiting
		}
	}

	duration := time.Since(startTime)

	s.logger.Info("Time-based scheduler stopped",
		"duration", duration,
		"active_jobs_stopped", activeJobsCount,
		"queue_length", queueLength,
		"workers_stopped", s.maxWorkers,
	)
}

// GetStats returns current scheduler statistics
func (s *TimeBasedScheduler) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"manager_id":   s.managerID,
		"active_jobs":  len(s.activeJobs),
		"queue_length": len(s.jobQueue),
		"worker_pool":  len(s.workerPool),
		"max_workers":  s.maxWorkers,
	}
}
