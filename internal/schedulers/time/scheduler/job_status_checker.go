package scheduler

import (
	"context"
	"math/big"
	"sync"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/repository"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

// JobStatusChecker handles periodic checking of job statuses
type JobStatusChecker struct {
	timeJobRepo repository.TimeJobRepository
	logger      observability.Logger
}

// NewJobStatusChecker creates a new JobStatusChecker instance
func NewJobStatusChecker(
	timeJobRepo repository.TimeJobRepository,
	logger observability.Logger,
) *JobStatusChecker {
	return &JobStatusChecker{
		timeJobRepo: timeJobRepo,
		logger:      logger,
	}
}

// StartStatusCheckLoop begins the periodic job status check
func (c *JobStatusChecker) StartStatusCheckLoop(ctx context.Context) {
	// Check every 15 minutes like the dbserver implementation
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	// Initial check on startup
	c.checkTimeJobs(ctx)

	for {
		select {
		case <-ctx.Done():
			c.logger.Info(ctx, "Job status checker stopping")
			return
		case <-ticker.C:
			c.checkTimeJobs(ctx)
		}
	}
}

// checkTimeJobs checks all active time jobs for expiration
func (c *JobStatusChecker) checkTimeJobs(ctx context.Context) {
	currentTime := time.Now()

	c.logger.Info(ctx, "Checking for expired time jobs",
		observability.Time("current_time", currentTime))

	timeJobs, err := c.timeJobRepo.GetActiveTimeJobs()
	if err != nil {
		c.logger.Error(ctx, "Failed to fetch active time jobs", observability.Error(err))
		return
	}

	expiredCount := 0
	var wg sync.WaitGroup

	for _, job := range timeJobs {
		if job.ExpirationTime.Before(currentTime) {
			wg.Add(1)
			go func(jobID *big.Int) {
				defer wg.Done()
				if err := c.timeJobRepo.UpdateTimeJobStatus(jobID, false); err != nil {
					c.logger.Error(ctx, "Failed to update time job status for job ID",
						observability.String("job_id", jobID.String()),
						observability.Error(err))
					return
				}
				c.logger.Info(ctx, "Time job marked as inactive due to expiration",
					observability.String("job_id", jobID.String()))
			}(job.JobID)
			expiredCount++
		}
	}

	wg.Wait()

	if expiredCount > 0 {
		c.logger.Info(ctx, "Time job status check completed",
			observability.Int("expired_count", expiredCount),
			observability.Int("total_checked", len(timeJobs)))
	} else {
		c.logger.Debug(ctx, "Time job status check completed, no expired jobs",
			observability.Int("total_checked", len(timeJobs)))
	}
}
