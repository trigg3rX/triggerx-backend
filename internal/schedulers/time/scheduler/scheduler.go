package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/cache"
	redisx "github.com/trigg3rX/triggerx-backend/internal/redis"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/client"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/config"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/parser"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

const (
	pollInterval       = 30 * time.Second // Poll every 30 seconds
	executionWindow    = 5 * time.Minute  // Look ahead 5 minutes
	batchSize          = 50               // Process jobs in batches
	performerLockTTL   = 10 * time.Minute // Lock duration for job execution
	jobCacheTTL        = 5 * time.Minute  // Cache TTL for job data
	duplicateJobWindow = 1 * time.Minute  // Window to prevent duplicate job execution
)

type TimeBasedScheduler struct {
	ctx        context.Context
	cancel     context.CancelFunc
	logger     logging.Logger
	workerPool chan struct{}
	activeJobs map[int64]*types.TimeJobData
	jobQueue   chan *types.TimeJobData
	dbClient   *client.DBServerClient
	cache      cache.Cache
	metrics    *metrics.Collector
	managerID  string
	maxWorkers int
}

// NewTimeBasedScheduler creates a new instance of TimeBasedScheduler
func NewTimeBasedScheduler(managerID string, logger logging.Logger, dbClient *client.DBServerClient) (*TimeBasedScheduler, error) {
	ctx, cancel := context.WithCancel(context.Background())

	maxWorkers := config.GetMaxWorkers()

	// Initialize cache
	if err := cache.Init(); err != nil {
		logger.Warnf("Failed to initialize cache: %v", err)
	}

	cacheInstance, err := cache.GetCache()
	if err != nil {
		logger.Warnf("Cache not available, running without cache: %v", err)
	}

	// Test Redis connection
	if err := redisx.Ping(); err != nil {
		logger.Warnf("Redis not available, job streaming disabled: %v", err)
	} else {
		logger.Info("Redis connection established, job streaming enabled")
	}

	scheduler := &TimeBasedScheduler{
		ctx:        ctx,
		cancel:     cancel,
		logger:     logger,
		workerPool: make(chan struct{}, maxWorkers),
		activeJobs: make(map[int64]*types.TimeJobData),
		jobQueue:   make(chan *types.TimeJobData, 100),
		dbClient:   dbClient,
		cache:      cacheInstance,
		metrics:    metrics.NewCollector(),
		managerID:  managerID,
		maxWorkers: maxWorkers,
	}

	// Start metrics collection
	scheduler.metrics.Start()

	// Start the worker pool
	for i := 0; i < maxWorkers; i++ {
		go scheduler.worker()
	}

	scheduler.logger.Info("Time-based scheduler initialized",
		"max_workers", maxWorkers,
		"manager_id", managerID,
		"cache_available", cacheInstance != nil,
	)

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
	jobKey := fmt.Sprintf("job_%d", job.JobID)

	s.logger.Infof("Executing time-based job %d (type: %s)", job.JobID, job.ScheduleType)

	// Try to acquire performer lock to prevent duplicate execution
	lockAcquired := false
	if s.cache != nil {
		acquired, err := s.cache.AcquirePerformerLock(jobKey, performerLockTTL)
		if err != nil {
			s.logger.Warnf("Failed to acquire performer lock for job %d: %v", job.JobID, err)
		} else if !acquired {
			s.logger.Warnf("Job %d is already being executed by another instance, skipping", job.JobID)
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

	// Create job execution context for Redis streaming
	jobContext := map[string]interface{}{
		"job_id":        job.JobID,
		"schedule_type": job.ScheduleType,
		"manager_id":    s.managerID,
		"started_at":    startTime.Unix(),
		"lock_acquired": lockAcquired,
	}

	// Add job start event to Redis stream
	if err := redisx.AddJobToStream(redisx.JobsReadyTimeStream, jobContext); err != nil {
		s.logger.Warnf("Failed to add job start event to Redis stream: %v", err)
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
		failureContext["status"] = "failed"
		failureContext["error"] = err.Error()
		failureContext["completed_at"] = time.Now().Unix()
		redisx.AddJobToStream(redisx.JobsRetryTimeStream, failureContext)
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

	if executionSuccess {
		metrics.JobsCompleted.Inc()
		s.logger.Infof("Completed job %d in %v, next execution at %v",
			job.JobID, duration, nextExecution)

		completionContext["status"] = "completed"
		redisx.AddJobToStream(redisx.JobsReadyTimeStream, completionContext)
	} else {
		metrics.JobsFailed.Inc()
		s.logger.Errorf("Failed to execute job %d after %v", job.JobID, duration)

		completionContext["status"] = "failed"
		redisx.AddJobToStream(redisx.JobsRetryTimeStream, completionContext)
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
		s.cache.Set(recentKey, time.Now().Format(time.RFC3339), duplicateJobWindow)
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
func (s *TimeBasedScheduler) getCachedJobData(jobID int64) (*types.TimeJobData, error) {
	if s.cache == nil {
		return nil, fmt.Errorf("cache not available")
	}

	cacheKey := fmt.Sprintf("job_data_%d", jobID)
	cachedData, err := s.cache.Get(cacheKey)
	if err != nil {
		return nil, err
	}

	var jobData map[string]interface{}
	if err := json.Unmarshal([]byte(cachedData), &jobData); err != nil {
		return nil, err
	}

	// This is a simplified example - you'd need to properly reconstruct the job
	job := &types.TimeJobData{
		JobID: jobID,
		// Add other fields as needed from the cached data
	}

	return job, nil
}

// Stop gracefully stops the scheduler
func (s *TimeBasedScheduler) Stop() {
	s.logger.Info("Stopping time-based scheduler")
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

	s.logger.Info("Time-based scheduler stopped")
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
