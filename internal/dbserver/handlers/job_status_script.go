package handlers

import (
	"context"
	"fmt"

	// "math/big"
	"sync"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/datastore/interfaces"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

const (
	StatusInactive = "inactive"
)

// JobStatusChecker handles periodic checking of job statuses
type JobStatusChecker struct {
	eventJobRepo     interfaces.GenericRepository[types.EventJobDataEntity]
	conditionJobRepo interfaces.GenericRepository[types.ConditionJobDataEntity]
	timeJobRepo      interfaces.GenericRepository[types.TimeJobDataEntity]
	logger           logging.Logger
}

// NewJobStatusChecker creates a new JobStatusChecker instance
func NewJobStatusChecker(
	eventJobRepo interfaces.GenericRepository[types.EventJobDataEntity],
	conditionJobRepo interfaces.GenericRepository[types.ConditionJobDataEntity],
	timeJobRepo interfaces.GenericRepository[types.TimeJobDataEntity],
	logger logging.Logger,
) *JobStatusChecker {
	return &JobStatusChecker{
		eventJobRepo:     eventJobRepo,
		conditionJobRepo: conditionJobRepo,
		timeJobRepo:      timeJobRepo,
		logger:           logger,
	}
}

// StartStatusCheckLoop begins the periodic job status check
func (c *JobStatusChecker) StartStatusCheckLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.checkJobStatuses()
	}
}

// checkJobStatuses checks all active jobs for expiration
func (c *JobStatusChecker) checkJobStatuses() {
	var wg sync.WaitGroup
	currentTime := time.Now()

	//log the current time and checking for jobs
	c.logger.Info(fmt.Sprintf("Checking for jobs at %s", currentTime.Format(time.RFC3339)))

	// Check event jobs
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.checkEventJobs(currentTime)
	}()

	// Check condition jobs
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.checkConditionJobs(currentTime)
	}()

	// Check time jobs
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.checkTimeJobs(currentTime)
	}()

	wg.Wait()

	c.logger.Info("Job status check completed")
}

// checkEventJobs checks all active event jobs for expiration
func (c *JobStatusChecker) checkEventJobs(currentTime time.Time) {
	ctx := context.Background()

	eventJobs, err := c.eventJobRepo.List(ctx)
	if err != nil {
		c.logger.Error("Failed to fetch active event jobs", err)
		return
	}

	for _, job := range eventJobs {
		if !job.IsCompleted && job.ExpirationTime.Before(currentTime) {
			job.IsCompleted = true
			if err := c.eventJobRepo.Update(ctx, job); err != nil {
				c.logger.Error(fmt.Sprintf("Failed to update event job status for job ID %v", &job.JobID), err)
				continue
			}
			c.logger.Info(fmt.Sprintf("Event job %v marked as inactive due to expiration", &job.JobID))
		}
	}
}

// checkConditionJobs checks all active condition jobs for expiration
func (c *JobStatusChecker) checkConditionJobs(currentTime time.Time) {
	ctx := context.Background()

	conditionJobs, err := c.conditionJobRepo.List(ctx)
	if err != nil {
		c.logger.Error("Failed to fetch active condition jobs", err)
		return
	}

	for _, job := range conditionJobs {
		if !job.IsCompleted && job.ExpirationTime.Before(currentTime) {
			job.IsCompleted = true
			if err := c.conditionJobRepo.Update(ctx, job); err != nil {
				c.logger.Error(fmt.Sprintf("Failed to update condition job status for job ID %v", &job.JobID), err)
				continue
			}
			c.logger.Info(fmt.Sprintf("Condition job %v marked as inactive due to expiration", &job.JobID))
		}
	}
}

// checkTimeJobs checks all active time jobs for expiration
func (c *JobStatusChecker) checkTimeJobs(currentTime time.Time) {
	ctx := context.Background()

	timeJobs, err := c.timeJobRepo.List(ctx)
	if err != nil {
		c.logger.Error("Failed to fetch active time jobs", err)
		return
	}

	for _, job := range timeJobs {
		if !job.IsCompleted && job.ExpirationTime.Before(currentTime) {
			job.IsCompleted = true
			if err := c.timeJobRepo.Update(ctx, job); err != nil {
				c.logger.Error(fmt.Sprintf("Failed to update time job status for job ID %v", &job.JobID), err)
				continue
			}
			c.logger.Info(fmt.Sprintf("Time job %v marked as inactive due to expiration", &job.JobID))
		}
	}
}
