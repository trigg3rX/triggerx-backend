package scheduler

import (
	"context"
	"fmt"
	"time"

	redisx "github.com/trigg3rX/triggerx-backend/internal/redis"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/scheduler/worker"

	commonTypes "github.com/trigg3rX/triggerx-backend/pkg/types"
)

// ScheduleJob creates and starts a new condition worker
func (s *ConditionBasedScheduler) ScheduleJob(jobData *commonTypes.ConditionJobData) error {
	s.workersMutex.Lock()
	defer s.workersMutex.Unlock()

	startTime := time.Now()

	// Check if job is already scheduled
	if _, exists := s.workers[jobData.JobID]; exists {
		// Add job scheduling failure event to Redis stream
		if redisx.IsAvailable() {
			failureEvent := map[string]interface{}{
				"event_type":        "job_schedule_failed",
				"job_id":            jobData.JobID,
				"manager_id":        s.managerID,
				"error":             "job already scheduled",
				"condition_type":    jobData.ConditionType,
				"value_source_type": jobData.ValueSourceType,
				"value_source_url":  jobData.ValueSourceUrl,
				"failed_at":         startTime.Unix(),
			}
			err := redisx.AddJobToStream(redisx.JobsRetryConditionStream, failureEvent)
			if err != nil {
				s.logger.Warnf("Failed to add job scheduling failure event to Redis stream: %v", err)
			}
		}
		return fmt.Errorf("job %d is already scheduled", jobData.JobID)
	}

	// Check if we've reached the maximum number of workers
	if len(s.workers) >= s.maxWorkers {
		// Add job scheduling failure event to Redis stream
		if redisx.IsAvailable() {
			failureEvent := map[string]interface{}{
				"event_type":        "job_schedule_failed",
				"job_id":            jobData.JobID,
				"manager_id":        s.managerID,
				"error":             fmt.Sprintf("maximum workers (%d) reached", s.maxWorkers),
				"current_workers":   len(s.workers),
				"max_workers":       s.maxWorkers,
				"condition_type":    jobData.ConditionType,
				"value_source_type": jobData.ValueSourceType,
				"value_source_url":  jobData.ValueSourceUrl,
				"failed_at":         startTime.Unix(),
			}
			err := redisx.AddJobToStream(redisx.JobsRetryConditionStream, failureEvent)
			if err != nil {
				s.logger.Warnf("Failed to add job scheduling failure event to Redis stream: %v", err)
			}
		}
		return fmt.Errorf("maximum number of workers (%d) reached, cannot schedule job %d", s.maxWorkers, jobData.JobID)
	}

	// Validate condition type
	if !isValidConditionType(jobData.ConditionType) {
		// Add validation failure event to Redis stream
		if redisx.IsAvailable() {
			failureEvent := map[string]interface{}{
				"event_type":        "job_schedule_failed",
				"job_id":            jobData.JobID,
				"manager_id":        s.managerID,
				"error":             fmt.Sprintf("unsupported condition type: %s", jobData.ConditionType),
				"condition_type":    jobData.ConditionType,
				"value_source_type": jobData.ValueSourceType,
				"value_source_url":  jobData.ValueSourceUrl,
				"failed_at":         startTime.Unix(),
			}
			err := redisx.AddJobToStream(redisx.JobsRetryConditionStream, failureEvent)
			if err != nil {
				s.logger.Warnf("Failed to add job scheduling failure event to Redis stream: %v", err)
			}
		}
		return fmt.Errorf("unsupported condition type: %s", jobData.ConditionType)
	}

	// Validate value source type
	if !isValidSourceType(jobData.ValueSourceType) {
		// Add validation failure event to Redis stream
		if redisx.IsAvailable() {
			failureEvent := map[string]interface{}{
				"event_type":        "job_schedule_failed",
				"job_id":            jobData.JobID,
				"manager_id":        s.managerID,
				"error":             fmt.Sprintf("unsupported value source type: %s", jobData.ValueSourceType),
				"condition_type":    jobData.ConditionType,
				"value_source_type": jobData.ValueSourceType,
				"value_source_url":  jobData.ValueSourceUrl,
				"failed_at":         startTime.Unix(),
			}
			if err := redisx.AddJobToStream(redisx.JobsRetryConditionStream, failureEvent); err != nil {
				s.logger.Warnf("Failed to add job scheduling failure event to Redis stream: %v", err)
			}
		}
		return fmt.Errorf("unsupported value source type: %s", jobData.ValueSourceType)
	}

	// Create condition worker
	worker, err := s.createConditionWorker(jobData)
	if err != nil {
		// Add worker creation failure event to Redis stream
		if redisx.IsAvailable() {
			failureEvent := map[string]interface{}{
				"event_type":        "job_schedule_failed",
				"job_id":            jobData.JobID,
				"manager_id":        s.managerID,
				"error":             fmt.Sprintf("failed to create worker: %v", err),
				"condition_type":    jobData.ConditionType,
				"value_source_type": jobData.ValueSourceType,
				"value_source_url":  jobData.ValueSourceUrl,
				"failed_at":         startTime.Unix(),
			}
			if err := redisx.AddJobToStream(redisx.JobsRetryConditionStream, failureEvent); err != nil {
				s.logger.Warnf("Failed to add job scheduling failure event to Redis stream: %v", err)
			}
		}
		return fmt.Errorf("failed to create condition worker: %w", err)
	}

	// Store worker
	s.workers[jobData.JobID] = worker

	// Start worker
	go worker.Start()

	// Update metrics
	metrics.JobsScheduled.Inc()
	metrics.JobsRunning.Inc()

	duration := time.Since(startTime)

	// Add comprehensive job scheduling success event to Redis stream
	if redisx.IsAvailable() {
		jobContext := map[string]interface{}{
			"event_type":              "job_scheduled",
			"job_id":                  jobData.JobID,
			"manager_id":              s.managerID,
			"condition_type":          jobData.ConditionType,
			"upper_limit":             jobData.UpperLimit,
			"lower_limit":             jobData.LowerLimit,
			"value_source_type":       jobData.ValueSourceType,
			"value_source_url":        jobData.ValueSourceUrl,
			"target_chain_id":         jobData.TargetChainID,
			"target_contract_address": jobData.TargetContractAddress,
			"target_function":         jobData.TargetFunction,
			"recurring":               jobData.Recurring,
			"active_workers":          len(s.workers),
			"max_workers":             s.maxWorkers,
			"cache_available":         s.cache != nil,
			"scheduled_at":            startTime.Unix(),
			"duration_ms":             duration.Milliseconds(),
			"status":                  "scheduled",
		}

		if err := redisx.AddJobToStream(redisx.JobsReadyConditionStream, jobContext); err != nil {
			s.logger.Warnf("Failed to add condition job scheduling event to Redis stream: %v", err)
		}
	}

	s.logger.Info("Condition job scheduled successfully",
		"job_id", jobData.JobID,
		"condition_type", jobData.ConditionType,
		"value_source", jobData.ValueSourceUrl,
		"upper_limit", jobData.UpperLimit,
		"lower_limit", jobData.LowerLimit,
		"target_chain", jobData.TargetChainID,
		"target_contract", jobData.TargetContractAddress,
		"target_function", jobData.TargetFunction,
		"active_workers", len(s.workers),
		"max_workers", s.maxWorkers,
		"duration", duration,
	)

	return nil
}

// createConditionWorker creates a new condition worker instance
func (s *ConditionBasedScheduler) createConditionWorker(jobData *commonTypes.ConditionJobData) (*worker.ConditionWorker, error) {
	ctx, cancel := context.WithCancel(s.ctx)

	worker := &worker.ConditionWorker{
		Job:        jobData,
		Logger:     s.logger,
		Cache:      s.cache,
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
		return fmt.Errorf("job %d is not scheduled", jobID)
	}

	// Stop worker
	worker.Stop()

	// Remove from workers map
	delete(s.workers, jobID)

	// Update metrics
	metrics.JobsRunning.Dec()

	// Add job unscheduling event to Redis stream
	jobContext := map[string]interface{}{
		"job_id":         jobID,
		"manager_id":     s.managerID,
		"unscheduled_at": time.Now().Unix(),
		"status":         "unscheduled",
	}

	if err := redisx.AddJobToStream(redisx.JobsReadyConditionStream, jobContext); err != nil {
		s.logger.Warnf("Failed to add condition job unscheduling event to Redis stream: %v", err)
	}

	s.logger.Info("Condition job unscheduled successfully", "job_id", jobID)
	return nil
}
