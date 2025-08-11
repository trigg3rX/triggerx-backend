package handlers

import (
	"fmt"
	// "math/big"
	"sync"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/repository"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

const (
	StatusInactive = "inactive"
)

// JobStatusChecker handles periodic checking of job statuses
type JobStatusChecker struct {
	eventJobRepo     repository.EventJobRepository
	conditionJobRepo repository.ConditionJobRepository
	timeJobRepo      repository.TimeJobRepository
	logger           logging.Logger
}

// NewJobStatusChecker creates a new JobStatusChecker instance
func NewJobStatusChecker(
	eventJobRepo repository.EventJobRepository,
	conditionJobRepo repository.ConditionJobRepository,
	timeJobRepo repository.TimeJobRepository,
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
	ticker := time.NewTicker(15 * time.Minute)
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
	eventJobs, err := c.eventJobRepo.GetActiveEventJobs()
	if err != nil {
		c.logger.Error("Failed to fetch active event jobs", err)
		return
	}

	for _, job := range eventJobs {
		if job.ExpirationTime.Before(currentTime) {
			if err := c.eventJobRepo.UpdateEventJobStatus(job.JobID, false); err != nil {
				c.logger.Error(fmt.Sprintf("Failed to update event job status for job ID %s", job.JobID.String()), err)
				continue
			}
			c.logger.Info(fmt.Sprintf("Event job %s marked as inactive due to expiration", job.JobID.String()))
		}
	}
}

// checkConditionJobs checks all active condition jobs for expiration
func (c *JobStatusChecker) checkConditionJobs(currentTime time.Time) {
	conditionJobs, err := c.conditionJobRepo.GetActiveConditionJobs()
	if err != nil {
		c.logger.Error("Failed to fetch active condition jobs", err)
		return
	}

	for _, job := range conditionJobs {
		if job.ExpirationTime.Before(currentTime) {
			if err := c.conditionJobRepo.UpdateConditionJobStatus(job.JobID, false); err != nil {
				c.logger.Error(fmt.Sprintf("Failed to update condition job status for job ID %s", job.JobID.String()), err)
				continue
			}
			c.logger.Info(fmt.Sprintf("Condition job %s marked as inactive due to expiration", job.JobID.String()))
		}
	}
}

// checkTimeJobs checks all active time jobs for expiration
func (c *JobStatusChecker) checkTimeJobs(currentTime time.Time) {
	timeJobs, err := c.timeJobRepo.GetActiveTimeJobs()
	if err != nil {
		c.logger.Error("Failed to fetch active time jobs", err)
		return
	}

	for _, job := range timeJobs {
		if job.ExpirationTime.Before(currentTime) {
			if err := c.timeJobRepo.UpdateTimeJobStatus(job.JobID, false); err != nil {
				c.logger.Error(fmt.Sprintf("Failed to update time job status for job ID %s", job.JobID.String()), err)
				continue
			}
			c.logger.Info(fmt.Sprintf("Time job %s marked as inactive due to expiration", job.JobID.String()))
		}
	}
}
