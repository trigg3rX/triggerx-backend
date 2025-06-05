package scheduler

import (
	"context"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/metrics"
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

	for i := 0; i < len(jobs); i += batchSize {
		end := i + batchSize
		if end > len(jobs) {
			end = len(jobs)
		}

		batch := jobs[i:end]
		s.processBatch(batch)
	}
}

// processBatch processes a batch of jobs
func (s *TimeBasedScheduler) processBatch(jobs []types.ScheduleTimeJobData) {
	for _, job := range jobs {
		// Add job to execution queue
		select {
		case s.jobQueue <- &job:
			s.executeJob(&job)
			metrics.JobsRunning.Inc()
			s.logger.Debugf("Queued job %d for execution", job.JobID)
		default:
			s.logger.Warnf("Job queue is full, skipping job %d", job.JobID)
			metrics.JobsFailed.Inc()
		}
	}
}

// executeJob executes a single job and updates its next execution time
func (s *TimeBasedScheduler) executeJob(job *types.ScheduleTimeJobData) {
	startTime := time.Now()

	s.logger.Infof("Executing time-based job %d (type: %s)", job.JobID, job.ScheduleType)

	// Check if ExpirationTime of the job has passed or not
	if job.ExpirationTime.Before(time.Now()) {
		s.logger.Infof("Job %d has expired, skipping execution", job.JobID)
		return
	}

	// Get the performer data
	// TODO: Get the performer data from redis service, which gets it from online keepers list from health service, and sets the performerLock in redis
	// For now, I fixed the performer
	performerData := types.GetPerformerData{
		KeeperID: 3,
		KeeperAddress: "0x0a067a261c5f5e8c4c0b9137430b4fe1255eb62e",
	}

	// Execute the actual job
	executionSuccess := s.performJobExecution(job, performerData)

	if executionSuccess {
		metrics.TasksExecuted.Inc()
		s.logger.Infof("Executed task for job %d in %v", job.JobID, time.Since(startTime))

	} else {
		metrics.TasksFailed.Inc()
		s.logger.Errorf("Failed to execute job %d after %v", job.JobID, time.Since(startTime))
	}
}

// performJobExecution handles the actual job execution logic
func (s *TimeBasedScheduler) performJobExecution(job *types.ScheduleTimeJobData, performerData types.GetPerformerData) bool {
	success, err := s.aggClient.SendTimeBasedTaskToPerformer(s.ctx, job, performerData)

	if err != nil {
		s.logger.Errorf("Failed to send task to performer: %v", err)
		return false
	}

	return success
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

	duration := time.Since(startTime)

	s.logger.Info("Time-based scheduler stopped",
		"duration", duration,
		"active_jobs_stopped", activeJobsCount,
		"queue_length", queueLength,
		"workers_stopped", s.maxWorkers,
	)
}
