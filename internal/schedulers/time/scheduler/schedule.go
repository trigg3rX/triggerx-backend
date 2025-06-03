package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	redisx "github.com/trigg3rX/triggerx-backend/internal/redis"
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
	pollStart := time.Now()

	jobs, err := s.dbClient.GetTimeBasedJobs()
	if err != nil {
		s.logger.Errorf("Failed to fetch time-based jobs: %v", err)
		metrics.JobsFailed.Inc()

		// Add database fetch failure event to Redis stream
		if redisx.IsAvailable() {
			fetchFailureEvent := map[string]interface{}{
				"event_type": "jobs_fetch_failed",
				"manager_id": s.managerID,
				"error":      err.Error(),
				"failed_at":  pollStart.Unix(),
			}
			if err := redisx.AddJobToStream(redisx.JobsRetryTimeStream, fetchFailureEvent); err != nil {
				s.logger.Errorf("Failed to add job fetch failure event to Redis stream: %v", err)
			}
		}
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

	// Add polling event to Redis stream
	if redisx.IsAvailable() {
		pollingEvent := map[string]interface{}{
			"event_type":       "jobs_poll_completed",
			"manager_id":       s.managerID,
			"jobs_found":       len(jobs),
			"execution_window": executionWindow.Unix(),
			"poll_duration_ms": time.Since(pollStart).Milliseconds(),
			"polled_at":        pollStart.Unix(),
		}

		if err := redisx.AddJobToStream(redisx.JobsReadyTimeStream, pollingEvent); err != nil {
			s.logger.Warnf("Failed to add polling event to Redis stream: %v", err)
		}
	}

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

		// Check cache to prevent duplicate processing
		if s.cache != nil {
			jobKey := fmt.Sprintf("timejob:processing:%d", job.JobID)
			if _, err := s.cache.Get(jobKey); err == nil {
				s.logger.Debugf("Job %d is already being processed (cache hit), skipping", job.JobID)
				continue
			}
			// Mark job as being processed
			if err := s.cache.Set(jobKey, "1", 5*time.Minute); err != nil {
				s.logger.Warnf("Failed to set processing cache for job %d: %v", job.JobID, err)
			}
		}

		// Add job to execution queue
		select {
		case s.jobQueue <- &job:
			metrics.JobsRunning.Inc()
			s.logger.Debugf("Queued job %d for execution", job.JobID)
		default:
			s.logger.Warnf("Job queue is full, skipping job %d", job.JobID)
			metrics.JobsFailed.Inc()

			// Add queue full event to Redis stream
			if redisx.IsAvailable() {
				queueFullEvent := map[string]interface{}{
					"event_type":   "job_queue_full",
					"job_id":       job.JobID,
					"manager_id":   s.managerID,
					"queue_length": len(s.jobQueue),
					"max_queue":    cap(s.jobQueue),
					"failed_at":    time.Now().Unix(),
				}
				if err := redisx.AddJobToStream(redisx.JobsRetryTimeStream, queueFullEvent); err != nil {
					s.logger.Errorf("Failed to add job queue full event to Redis stream: %v", err)
				}
			}
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
	jobKey := fmt.Sprintf("job_%d", job.JobID)

	s.logger.Infof("Executing time-based job %d (type: %s)", job.JobID, job.ScheduleType)

	// Try to acquire performer lock to prevent duplicate execution
	lockAcquired := false
	if s.cache != nil {
		acquired, err := s.cache.AcquirePerformerLock(jobKey, performerLockTTL)
		if err != nil {
			s.logger.Warnf("Failed to acquire performer lock for job %d: %v", job.JobID, err)

			// Add lock failure event to Redis stream
			if redisx.IsAvailable() {
				lockFailureEvent := map[string]interface{}{
					"event_type":    "job_lock_failed",
					"job_id":        job.JobID,
					"manager_id":    s.managerID,
					"schedule_type": job.ScheduleType,
					"error":         err.Error(),
					"failed_at":     startTime.Unix(),
				}
				if err := redisx.AddJobToStream(redisx.JobsRetryTimeStream, lockFailureEvent); err != nil {
					s.logger.Errorf("Failed to add job lock failure event to Redis stream: %v", err)
				}
			}
		} else if !acquired {
			s.logger.Warnf("Job %d is already being executed by another instance, skipping", job.JobID)

			// Add lock conflict event to Redis stream
			if redisx.IsAvailable() {
				lockConflictEvent := map[string]interface{}{
					"event_type":    "job_lock_conflict",
					"job_id":        job.JobID,
					"manager_id":    s.managerID,
					"schedule_type": job.ScheduleType,
					"conflict_at":   startTime.Unix(),
				}
				if err := redisx.AddJobToStream(redisx.JobsRetryTimeStream, lockConflictEvent); err != nil {
					s.logger.Errorf("Failed to add job lock conflict event to Redis stream: %v", err)
				}
			}
			return
		}
		lockAcquired = true
		defer func() {
			if err := s.cache.ReleasePerformerLock(jobKey); err != nil {
				s.logger.Warnf("Failed to release performer lock for job %d: %v", job.JobID, err)
			}
		}()
	}

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
		"cache_available":          s.cache != nil,
		"scheduled_execution_time": job.NextExecutionTimestamp.Unix(),
		"actual_start_time":        startTime.Unix(),
		"delay_seconds":            startTime.Sub(job.NextExecutionTimestamp).Seconds(),
		"status":                   "processing",
	}

	// Add job start event to Redis stream
	if redisx.IsAvailable() {
		if err := redisx.AddJobToStream(redisx.JobsReadyTimeStream, jobContext); err != nil {
			s.logger.Warnf("Failed to add job start event to Redis stream: %v", err)
		}
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

		if redisx.IsAvailable() {
			if err := redisx.AddJobToStream(redisx.JobsRetryTimeStream, failureContext); err != nil {
				s.logger.Errorf("Failed to add job failure event to Redis stream: %v", err)
			}
		}
		return
	}

	// Update next execution time in database
	if err := s.dbClient.UpdateJobNextExecution(job.JobID, nextExecution); err != nil {
		s.logger.Errorf("Failed to update next execution time for job %d: %v", job.JobID, err)
		metrics.JobsFailed.Inc()

		// Add database update failure event to Redis stream
		if redisx.IsAvailable() {
			dbFailureContext := jobContext
			dbFailureContext["event_type"] = "job_db_update_failed"
			dbFailureContext["status"] = "failed"
			dbFailureContext["error"] = err.Error()
			dbFailureContext["error_type"] = "database_update"
			dbFailureContext["next_execution"] = nextExecution.Unix()
			dbFailureContext["completed_at"] = time.Now().Unix()
			dbFailureContext["duration_ms"] = time.Since(startTime).Milliseconds()

			if err := redisx.AddJobToStream(redisx.JobsRetryTimeStream, dbFailureContext); err != nil {
				s.logger.Errorf("Failed to add job database update failure event to Redis stream: %v", err)
			}
		}
		return
	}

	// Update job status to completed
	if err := s.dbClient.UpdateJobStatus(job.JobID, false); err != nil {
		s.logger.Errorf("Failed to update job %d status to completed: %v", job.JobID, err)
	}

	// Cache the updated job data
	if s.cache != nil {
		s.cacheJobData(job, nextExecution)
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

		if redisx.IsAvailable() {
			if err := redisx.AddJobToStream(redisx.JobsReadyTimeStream, completionContext); err != nil {
				s.logger.Errorf("Failed to add job completion event to Redis stream: %v", err)
			}
		}
	} else {
		metrics.JobsFailed.Inc()
		s.logger.Errorf("Failed to execute job %d after %v", job.JobID, duration)

		completionContext["event_type"] = "job_failed"
		completionContext["status"] = "failed"
		completionContext["error"] = "job execution failed"
		completionContext["error_type"] = "execution_failure"

		if redisx.IsAvailable() {
			if err := redisx.AddJobToStream(redisx.JobsRetryTimeStream, completionContext); err != nil {
				s.logger.Errorf("Failed to add job failure event to Redis stream: %v", err)
			}
		}
	}
}

// performJobExecution handles the actual job execution logic
func (s *TimeBasedScheduler) performJobExecution(job *types.TimeJobData) bool {
	// Check cache for recent execution to prevent duplicates
	if s.cache != nil {
		recentKey := fmt.Sprintf("recent_execution_%d", job.JobID)
		if _, err := s.cache.Get(recentKey); err == nil {
			s.logger.Warnf("Job %d was recently executed, skipping duplicate", job.JobID)
			return true
		}

		// Mark as recently executed
		if err := s.cache.Set(recentKey, time.Now().Format(time.RFC3339), duplicateJobWindow); err != nil {
			s.logger.Errorf("Failed to set recent execution cache: %v", err)
		}
	}

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

// cacheJobData caches job data for faster access
func (s *TimeBasedScheduler) cacheJobData(job *types.TimeJobData, nextExecution time.Time) {
	if s.cache == nil {
		return
	}

	jobData := map[string]interface{}{
		"job_id":         job.JobID,
		"schedule_type":  job.ScheduleType,
		"next_execution": nextExecution.Unix(),
		"cached_at":      time.Now().Unix(),
	}

	jsonData, err := json.Marshal(jobData)
	if err != nil {
		s.logger.Warnf("Failed to marshal job data for caching: %v", err)
		return
	}

	cacheKey := fmt.Sprintf("job_data_%d", job.JobID)
	if err := s.cache.Set(cacheKey, string(jsonData), jobCacheTTL); err != nil {
		s.logger.Warnf("Failed to cache job data: %v", err)
	}
}

// getCachedJobData retrieves job data from cache
// func (s *TimeBasedScheduler) getCachedJobData(jobID int64) (*types.TimeJobData, error) {
// 	if s.cache == nil {
// 		return nil, fmt.Errorf("cache not available")
// 	}

// 	cacheKey := fmt.Sprintf("job_data_%d", jobID)
// 	cachedData, err := s.cache.Get(cacheKey)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var jobData map[string]interface{}
// 	if err := json.Unmarshal([]byte(cachedData), &jobData); err != nil {
// 		return nil, err
// 	}

// 	// This is a simplified example - you'd need to properly reconstruct the job
// 	job := &types.TimeJobData{
// 		JobID: jobID,
// 		// Add other fields as needed from the cached data
// 	}

// 	return job, nil
// }

// Stop gracefully stops the scheduler
func (s *TimeBasedScheduler) Stop() {
	startTime := time.Now()
	s.logger.Info("Stopping time-based scheduler")

	// Capture statistics before shutdown
	activeJobsCount := len(s.activeJobs)
	queueLength := len(s.jobQueue)
	workersInUse := len(s.workerPool)

	// Add scheduler shutdown event to Redis stream
	if redisx.IsAvailable() {
		shutdownEvent := map[string]interface{}{
			"event_type":      "scheduler_shutdown",
			"manager_id":      s.managerID,
			"active_jobs":     activeJobsCount,
			"queue_length":    queueLength,
			"workers_in_use":  workersInUse,
			"max_workers":     s.maxWorkers,
			"cache_available": s.cache != nil,
			"shutdown_at":     startTime.Unix(),
			"graceful":        true,
		}

		if err := redisx.AddJobToStream(redisx.JobsReadyTimeStream, shutdownEvent); err != nil {
			s.logger.Warnf("Failed to add scheduler shutdown event to Redis stream: %v", err)
		} else {
			s.logger.Info("Scheduler shutdown event added to Redis stream")
		}
	}

	s.cancel()

	// Close job queue
	close(s.jobQueue)

	// Wait for workers to finish (with timeout)
	timeout := time.After(30 * time.Second)
	for len(s.workerPool) < s.maxWorkers {
		select {
		case <-timeout:
			s.logger.Warn("Timeout waiting for workers to finish")

			// Add timeout event to Redis stream
			if redisx.IsAvailable() {
				timeoutEvent := map[string]interface{}{
					"event_type":        "scheduler_shutdown_timeout",
					"manager_id":        s.managerID,
					"remaining_workers": s.maxWorkers - len(s.workerPool),
					"timeout_seconds":   30,
					"timeout_at":        time.Now().Unix(),
				}
				if err := redisx.AddJobToStream(redisx.JobsRetryTimeStream, timeoutEvent); err != nil {
					s.logger.Warnf("Failed to add scheduler shutdown timeout event to Redis stream: %v", err)
				}
			}
			return
		case <-time.After(100 * time.Millisecond):
			// Continue waiting
		}
	}

	duration := time.Since(startTime)

	// Add final shutdown completion event to Redis stream
	if redisx.IsAvailable() {
		completionEvent := map[string]interface{}{
			"event_type":      "scheduler_shutdown_complete",
			"manager_id":      s.managerID,
			"duration_ms":     duration.Milliseconds(),
			"completed_at":    time.Now().Unix(),
			"workers_stopped": s.maxWorkers,
		}

		if err := redisx.AddJobToStream(redisx.JobsReadyTimeStream, completionEvent); err != nil {
			s.logger.Warnf("Failed to add shutdown completion event to Redis stream: %v", err)
		}
	}

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
