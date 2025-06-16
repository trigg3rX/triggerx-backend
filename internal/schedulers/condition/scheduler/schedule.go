package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/scheduler/worker"

	commonTypes "github.com/trigg3rX/triggerx-backend/pkg/types"
)

// ScheduleJob creates and starts a new condition worker
func (s *ConditionBasedScheduler) ScheduleJob(jobData *commonTypes.ScheduleConditionJobData) error {
	s.workersMutex.Lock()
	defer s.workersMutex.Unlock()

	startTime := time.Now()

	// Check if job is already scheduled
	if _, exists := s.workers[jobData.JobID]; exists {
		metrics.TrackCriticalError("duplicate_job_schedule")
		return fmt.Errorf("job %d is already scheduled", jobData.JobID)
	}

	// Check if we've reached the maximum number of workers
	if len(s.workers) >= s.maxWorkers {
		metrics.TrackCriticalError("max_workers_exceeded")
		return fmt.Errorf("maximum number of workers (%d) reached, cannot schedule job %d", s.maxWorkers, jobData.JobID)
	}

	// Validate condition type
	if !isValidConditionType(jobData.ConditionType) {
		metrics.TrackCriticalError("invalid_condition_type")
		return fmt.Errorf("unsupported condition type: %s", jobData.ConditionType)
	}

	// Validate value source type
	if !isValidSourceType(jobData.ValueSourceType) {
		metrics.TrackCriticalError("invalid_source_type")
		return fmt.Errorf("unsupported value source type: %s", jobData.ValueSourceType)
	}

	// Track condition by type and source
	metrics.TrackConditionByType(jobData.ConditionType)
	metrics.TrackConditionBySource(jobData.ValueSourceType)

	// Create condition worker
	worker, err := s.createConditionWorker(jobData)
	if err != nil {
		metrics.TrackCriticalError("worker_creation_failed")
		return fmt.Errorf("failed to create condition worker: %w", err)
	}

	// Store worker
	s.workers[jobData.JobID] = worker

	// Start worker
	go worker.Start()

	// Update metrics
	metrics.TrackJobScheduled()
	metrics.UpdateActiveWorkers(len(s.workers))

	duration := time.Since(startTime)

	s.logger.Info("Condition job scheduled successfully",
		"job_id", jobData.JobID,
		"condition_type", jobData.ConditionType,
		"value_source", jobData.ValueSourceUrl,
		"upper_limit", jobData.UpperLimit,
		"lower_limit", jobData.LowerLimit,
		"target_chain", jobData.TaskTargetData.TargetChainID,
		"target_contract", jobData.TaskTargetData.TargetContractAddress,
		"target_function", jobData.TaskTargetData.TargetFunction,
		"active_workers", len(s.workers),
		"max_workers", s.maxWorkers,
		"duration", duration,
	)

	return nil
}

// createConditionWorker creates a new condition worker instance
func (s *ConditionBasedScheduler) createConditionWorker(jobData *commonTypes.ScheduleConditionJobData) (*worker.ConditionWorker, error) {
	ctx, cancel := context.WithCancel(s.ctx)

	worker := &worker.ConditionWorker{
		Job:        jobData,
		Logger:     s.logger,
		HttpClient: s.httpClient,
		Ctx:        ctx,
		Cancel:     cancel,
		IsActive:   false,
		LastCheck:  time.Now(),
		ManagerID:  s.managerID,
	}

	return worker, nil
}

// UnscheduleJob stops and removes a condition worker
func (s *ConditionBasedScheduler) UnscheduleJob(jobID int64) error {
	s.workersMutex.Lock()
	defer s.workersMutex.Unlock()

	worker, exists := s.workers[jobID]
	if !exists {
		metrics.TrackCriticalError("job_not_found")
		return fmt.Errorf("job %d is not scheduled", jobID)
	}

	// Stop worker
	worker.Stop()

	// Remove from workers map
	delete(s.workers, jobID)

	// Update active workers count
	metrics.UpdateActiveWorkers(len(s.workers))

	// Track job completion
	metrics.TrackJobCompleted("unscheduled")

	s.logger.Info("Condition job unscheduled successfully", "job_id", jobID)
	return nil
}
