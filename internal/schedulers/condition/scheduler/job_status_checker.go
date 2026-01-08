package scheduler

import (
	"context"
	"math/big"
	"sync"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/repository"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

// JobStatusChecker handles periodic checking of job statuses
type JobStatusChecker struct {
	eventJobRepo     repository.EventJobRepository
	conditionJobRepo repository.ConditionJobRepository
	logger           observability.Logger
}

// NewJobStatusChecker creates a new JobStatusChecker instance
func NewJobStatusChecker(
	eventJobRepo repository.EventJobRepository,
	conditionJobRepo repository.ConditionJobRepository,
	logger observability.Logger,
) *JobStatusChecker {
	return &JobStatusChecker{
		eventJobRepo:     eventJobRepo,
		conditionJobRepo: conditionJobRepo,
		logger:           logger,
	}
}

// StartStatusCheckLoop begins the periodic job status check
func (c *JobStatusChecker) StartStatusCheckLoop(ctx context.Context) {
	// Check every 15 minutes like the dbserver implementation
	ticker := time.NewTicker(15 * time.Minute)
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
func (c *JobStatusChecker) checkEventJobs(ctx context.Context, currentTime time.Time) {
	eventJobs, err := c.eventJobRepo.GetActiveEventJobs()
	if err != nil {
		c.logger.Error(ctx, "Failed to fetch active event jobs", observability.Error(err))
		return
	}

	expiredCount := 0
	var wg sync.WaitGroup

	for _, job := range eventJobs {
		if job.ExpirationTime.Before(currentTime) {
			wg.Add(1)
			go func(jobID *big.Int) {
				defer wg.Done()
				if err := c.eventJobRepo.UpdateEventJobStatus(jobID, false); err != nil {
					c.logger.Error(ctx, "Failed to update event job status for job ID",
						observability.String("job_id", jobID.String()),
						observability.Error(err))
					return
				}
				c.logger.Info(ctx, "Event job marked as inactive due to expiration",
					observability.String("job_id", jobID.String()))
			}(job.JobID)
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
func (c *JobStatusChecker) checkConditionJobs(ctx context.Context, currentTime time.Time) {
	conditionJobs, err := c.conditionJobRepo.GetActiveConditionJobs()
	if err != nil {
		c.logger.Error(ctx, "Failed to fetch active condition jobs", observability.Error(err))
		return
	}

	expiredCount := 0
	var wg sync.WaitGroup

	for _, job := range conditionJobs {
		if job.ExpirationTime.Before(currentTime) {
			wg.Add(1)
			go func(jobID *big.Int) {
				defer wg.Done()
				if err := c.conditionJobRepo.UpdateConditionJobStatus(jobID, false); err != nil {
					c.logger.Error(ctx, "Failed to update condition job status for job ID",
						observability.String("job_id", jobID.String()),
						observability.Error(err))
					return
				}
				c.logger.Info(ctx, "Condition job marked as inactive due to expiration",
					observability.String("job_id", jobID.String()))
			}(job.JobID)
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
