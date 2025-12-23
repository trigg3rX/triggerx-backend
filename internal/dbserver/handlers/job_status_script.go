package handlers

import (
	"context"
	// "math/big"
	"sync"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/repository"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

const (
	StatusInactive = "inactive"
)

// JobStatusChecker handles periodic checking of job statuses
type JobStatusChecker struct {
	eventJobRepo     repository.EventJobRepository
	conditionJobRepo repository.ConditionJobRepository
	timeJobRepo      repository.TimeJobRepository
	logger           observability.Logger
}

// NewJobStatusChecker creates a new JobStatusChecker instance
func NewJobStatusChecker(
	eventJobRepo repository.EventJobRepository,
	conditionJobRepo repository.ConditionJobRepository,
	timeJobRepo repository.TimeJobRepository,
	logger observability.Logger,
) *JobStatusChecker {
	return &JobStatusChecker{
		eventJobRepo:     eventJobRepo,
		conditionJobRepo: conditionJobRepo,
		timeJobRepo:      timeJobRepo,
		logger:           logger,
	}
}

// StartStatusCheckLoop begins the periodic job status check
func (c *JobStatusChecker) StartStatusCheckLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.checkJobStatuses(ctx)
	}
}

// checkJobStatuses checks all active jobs for expiration
func (c *JobStatusChecker) checkJobStatuses(ctx context.Context) {
	var wg sync.WaitGroup
	currentTime := time.Now()

	//log the current time and checking for jobs
	// c.logger.Info(ctx, "Checking for jobs at", observability.Time("current_time", currentTime))

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

	// Check time jobs
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.checkTimeJobs(ctx, currentTime)
	}()

	wg.Wait()

	// c.logger.Info("Job status check completed")
}

// checkEventJobs checks all active event jobs for expiration
func (c *JobStatusChecker) checkEventJobs(ctx context.Context, currentTime time.Time) {
	eventJobs, err := c.eventJobRepo.GetActiveEventJobs()
	if err != nil {
		c.logger.Error(ctx, "Failed to fetch active event jobs", observability.Error(err))
		return
	}

	for _, job := range eventJobs {
		if job.ExpirationTime.Before(currentTime) {
			if err := c.eventJobRepo.UpdateEventJobStatus(job.JobID.Int, false); err != nil {
				c.logger.Error(ctx, "Failed to update event job status for job ID", observability.String("job_id", job.JobID.String()), observability.Error(err))
				continue
			}
			c.logger.Info(ctx, "Event job marked as inactive due to expiration", observability.String("job_id", job.JobID.String()))
		}
	}
}

// checkConditionJobs checks all active condition jobs for expiration
func (c *JobStatusChecker) checkConditionJobs(ctx context.Context, currentTime time.Time) {
	conditionJobs, err := c.conditionJobRepo.GetActiveConditionJobs()
	if err != nil {
		c.logger.Error(ctx, "Failed to fetch active condition jobs", observability.Error(err))
		return
	}

	for _, job := range conditionJobs {
		if job.ExpirationTime.Before(currentTime) {
			if err := c.conditionJobRepo.UpdateConditionJobStatus(job.JobID.Int, false); err != nil {
				c.logger.Error(ctx, "Failed to update condition job status for job ID", observability.String("job_id", job.JobID.String()), observability.Error(err))
				continue
			}
			c.logger.Info(ctx, "Condition job marked as inactive due to expiration", observability.String("job_id", job.JobID.String()))
		}
	}
}

// checkTimeJobs checks all active time jobs for expiration
func (c *JobStatusChecker) checkTimeJobs(ctx context.Context, currentTime time.Time) {
	timeJobs, err := c.timeJobRepo.GetActiveTimeJobs()
	if err != nil {
		c.logger.Error(ctx, "Failed to fetch active time jobs", observability.Error(err))
		return
	}

	for _, job := range timeJobs {
		if job.ExpirationTime.Before(currentTime) {
			if err := c.timeJobRepo.UpdateTimeJobStatus(job.JobID.Int, false); err != nil {
				c.logger.Error(ctx, "Failed to update time job status for job ID", observability.String("job_id", job.JobID.String()), observability.Error(err))
				continue
			}
			c.logger.Info(ctx, "Time job marked as inactive due to expiration", observability.String("job_id", job.JobID.String()))
		}
	}
}
