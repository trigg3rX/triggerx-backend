package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/database/repository"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

// JobStatusChecker handles periodic checking of job statuses
// It ensures expired jobs are marked as inactive in the database and
// stops any running workers or unregisters event monitoring for those jobs.
type JobStatusChecker struct {
	eventJobRepo     repository.EventJobRepository
	conditionJobRepo repository.ConditionJobRepository
	scheduler        *ConditionBasedScheduler
	logger           observability.Logger
}

// NewJobStatusChecker creates a new JobStatusChecker instance
func NewJobStatusChecker(
	eventJobRepo repository.EventJobRepository,
	conditionJobRepo repository.ConditionJobRepository,
	scheduler *ConditionBasedScheduler,
	logger observability.Logger,
) *JobStatusChecker {
	return &JobStatusChecker{
		eventJobRepo:     eventJobRepo,
		conditionJobRepo: conditionJobRepo,
		scheduler:        scheduler,
		logger:           logger,
	}
}

// StartStatusCheckLoop begins the periodic job status check
// Checks more frequently to ensure expired jobs are cleaned up promptly
func (c *JobStatusChecker) StartStatusCheckLoop(ctx context.Context) {
	// Check every 30 seconds for faster expiration handling
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Initial check on startup
	c.checkJobStatuses(ctx)

	for {
		select {
		case <-ctx.Done():
			c.logger.Info(ctx, "Job status checker stopping")
			return
		case <-ticker.C:
			c.checkJobStatuses(ctx)
		}
	}
}

// checkJobStatuses checks all active jobs for expiration
func (c *JobStatusChecker) checkJobStatuses(ctx context.Context) {
	var wg sync.WaitGroup
	currentTime := time.Now()

	c.logger.Info(ctx, "Checking for expired jobs",
		observability.Time("current_time", currentTime))

	// Check event jobs
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.checkEventJobs(ctx, currentTime)
	}()

	// Check condition jobs
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.checkConditionJobs(ctx, currentTime)
	}()

	wg.Wait()

	c.logger.Debug(ctx, "Job status check completed")
}

// checkEventJobs checks all active event jobs for expiration
// For expired jobs, it sets is_active=false in DB and unregisters from Event Monitor Service if still registered
func (c *JobStatusChecker) checkEventJobs(ctx context.Context, currentTime time.Time) {
	eventJobs, err := c.eventJobRepo.GetActiveEventJobs()
	if err != nil {
		c.logger.Error(ctx, "Failed to fetch active event jobs", observability.Error(err))
		return
	}

	expiredCount := 0
	var wg sync.WaitGroup

	for _, job := range eventJobs {
		if job.ExpirationTime.Before(currentTime) || job.ExpirationTime.Equal(currentTime) {
			wg.Add(1)
			go func(jobID string, expirationTime time.Time) {
				defer wg.Done()

				// First, check if job is still registered in the scheduler
				// If so, unregister it from Event Monitor Service
				if c.scheduler != nil {
					// Check if job exists in jobDataStore (means it's still registered)
					c.scheduler.workersMutex.RLock()
					_, exists := c.scheduler.jobDataStore[jobID]
					c.scheduler.workersMutex.RUnlock()

					if exists {
						// Job is still registered, unregister it
						if err := c.scheduler.UnregisterEventJob(ctx, jobID); err != nil {
							c.logger.Warn(ctx, "Failed to unregister expired event job from Event Monitor Service",
								observability.String("job_id", jobID),
								observability.Error(err))
							// Continue to update DB status even if unregister fails
						} else {
							c.logger.Info(ctx, "Unregistered expired event job from Event Monitor Service",
								observability.String("job_id", jobID))
						}
					}
				}

				// Set is_active=false in database
				if err := c.eventJobRepo.UpdateEventJobStatus(jobID, false); err != nil {
					c.logger.Error(ctx, "Failed to update event job status in database",
						observability.String("job_id", jobID),
						observability.Error(err))
					return
				}

				c.logger.Info(ctx, "Event job marked as inactive due to expiration",
					observability.String("job_id", jobID),
					observability.Time("expiration_time", expirationTime))
			}(job.JobID, job.ExpirationTime)
			expiredCount++
		}
	}

	wg.Wait()

	if expiredCount > 0 {
		c.logger.Info(ctx, "Event job status check completed",
			observability.Int("expired_count", expiredCount),
			observability.Int("total_checked", len(eventJobs)))
	} else {
		c.logger.Debug(ctx, "Event job status check completed, no expired jobs",
			observability.Int("total_checked", len(eventJobs)))
	}
}

// checkConditionJobs checks all active condition jobs for expiration
// For expired jobs, it sets is_active=false in DB and stops the worker if still running
func (c *JobStatusChecker) checkConditionJobs(ctx context.Context, currentTime time.Time) {
	conditionJobs, err := c.conditionJobRepo.GetActiveConditionJobs()
	if err != nil {
		c.logger.Error(ctx, "Failed to fetch active condition jobs", observability.Error(err))
		return
	}

	expiredCount := 0
	var wg sync.WaitGroup

	for _, job := range conditionJobs {
		if job.ExpirationTime.Before(currentTime) || job.ExpirationTime.Equal(currentTime) {
			wg.Add(1)
			go func(jobID string, expirationTime time.Time) {
				defer wg.Done()

				// First, check if worker is still running and stop it
				if c.scheduler != nil {
					// Check if job data exists (means worker is still scheduled)
					c.scheduler.workersMutex.RLock()
					_, jobDataExists := c.scheduler.jobDataStore[jobID]
					c.scheduler.workersMutex.RUnlock()

					if jobDataExists {
						// Worker is still running or job data exists, unschedule it
						// UnscheduleJob handles stopping workers and cleaning up job data
						if err := c.scheduler.UnscheduleJob(ctx, jobID); err != nil {
							c.logger.Warn(ctx, "Failed to unschedule expired condition job",
								observability.String("job_id", jobID),
								observability.Error(err))
							// Continue to update DB status even if unschedule fails
						} else {
							c.logger.Info(ctx, "Stopped expired condition job worker",
								observability.String("job_id", jobID))
						}
					}
				}

				// Set is_active=false in database
				if err := c.conditionJobRepo.UpdateConditionJobStatus(jobID, false); err != nil {
					c.logger.Error(ctx, "Failed to update condition job status in database",
						observability.String("job_id", jobID),
						observability.Error(err))
					return
				}

				c.logger.Info(ctx, "Condition job marked as inactive due to expiration",
					observability.String("job_id", jobID),
					observability.Time("expiration_time", expirationTime))
			}(job.JobID, job.ExpirationTime)
			expiredCount++
		}
	}

	wg.Wait()

	if expiredCount > 0 {
		c.logger.Info(ctx, "Condition job status check completed",
			observability.Int("expired_count", expiredCount),
			observability.Int("total_checked", len(conditionJobs)))
	} else {
		c.logger.Debug(ctx, "Condition job status check completed, no expired jobs",
			observability.Int("total_checked", len(conditionJobs)))
	}
}
