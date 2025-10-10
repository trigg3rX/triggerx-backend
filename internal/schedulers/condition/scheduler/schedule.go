package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/scheduler/worker"
	httppkg "github.com/trigg3rX/triggerx-backend/pkg/http"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// ScheduleJob creates and starts a new condition worker for monitoring
func (s *ConditionBasedScheduler) ScheduleJob(jobData *types.ScheduleConditionJobData) error {
	s.workersMutex.Lock()
	defer s.workersMutex.Unlock()

	startTime := time.Now()

	switch jobData.TaskDefinitionID {
	case 3, 4: // Event-based jobs
		if err := s.scheduleEventJob(jobData, startTime); err != nil {
			return err
		}

	case 5, 6: // Condition-based jobs
		if err := s.scheduleConditionJob(jobData, startTime); err != nil {
			return err
		}

	default:
		return fmt.Errorf("unsupported task definition id: %d", jobData.TaskDefinitionID)
	}

	// Update metrics
	metrics.TrackJobScheduled()
	metrics.UpdateActiveWorkers(len(s.eventWorkers) + len(s.conditionWorkers))
	metrics.TrackWorkerStart(jobData.JobID)

	return nil
}

// scheduleConditionJob handles condition-based job scheduling
func (s *ConditionBasedScheduler) scheduleConditionJob(jobData *types.ScheduleConditionJobData, startTime time.Time) error {
	// Check if job is already scheduled
	if _, exists := s.conditionWorkers[jobData.JobID]; exists {
		metrics.TrackCriticalError("duplicate_job_schedule")
		return fmt.Errorf("job %s is already scheduled", jobData.JobID)
	}

	// Validate condition type
	if !isValidConditionType(jobData.ConditionWorkerData.ConditionType) {
		metrics.TrackCriticalError("invalid_condition_type")
		return fmt.Errorf("unsupported condition type: %s", jobData.ConditionWorkerData.ConditionType)
	}

	// Validate value source type
	if !isValidSourceType(jobData.ConditionWorkerData.ValueSourceType) {
		metrics.TrackCriticalError("invalid_source_type")
		return fmt.Errorf("unsupported value source type: %s", jobData.ConditionWorkerData.ValueSourceType)
	}

	// Create condition worker with Redis callback
	conditionWorker, err := s.createConditionWorker(&jobData.ConditionWorkerData, s.HTTPClient)
	if err != nil {
		metrics.TrackCriticalError("worker_creation_failed")
		return fmt.Errorf("failed to create condition worker: %w", err)
	}

	// Store worker and job data separately for Redis integration
	s.conditionWorkers[jobData.JobID] = conditionWorker
	s.jobDataStore[jobData.JobID] = jobData

	// Start worker
	go conditionWorker.Start()

	duration := time.Since(startTime)

	// Track condition by type and source
	metrics.TrackConditionByType(jobData.ConditionWorkerData.ConditionType)
	metrics.TrackConditionBySource(jobData.ConditionWorkerData.ValueSourceType)

	s.logger.Info("Condition job monitoring started",
		"job_id", jobData.JobID,
		"condition_type", jobData.ConditionWorkerData.ConditionType,
		"value_source", jobData.ConditionWorkerData.ValueSourceUrl,
		"upper_limit", jobData.ConditionWorkerData.UpperLimit,
		"lower_limit", jobData.ConditionWorkerData.LowerLimit,
		"active_workers", len(s.eventWorkers)+len(s.conditionWorkers),
		"max_workers", s.maxWorkers,
		"duration", duration,
	)

	return nil
}

// scheduleEventJob handles event-based job scheduling
func (s *ConditionBasedScheduler) scheduleEventJob(jobData *types.ScheduleConditionJobData, startTime time.Time) error {
	// Check if job is already scheduled
	if _, exists := s.eventWorkers[jobData.JobID]; exists {
		metrics.TrackCriticalError("duplicate_job_schedule")
		return fmt.Errorf("job %s is already scheduled", jobData.JobID)
	}

	// Check if chain client is available
	if _, exists := s.chainClients[jobData.EventWorkerData.TriggerChainID]; !exists {
		metrics.TrackCriticalError("chain_client_not_found")
		return fmt.Errorf("chain client not found for chain %s", jobData.EventWorkerData.TriggerChainID)
	}

	// Validate contract address
	if !common.IsHexAddress(jobData.EventWorkerData.TriggerContractAddress) {
		metrics.TrackCriticalError("invalid_contract_address")
		return fmt.Errorf("invalid contract address: %s", jobData.EventWorkerData.TriggerContractAddress)
	}

	// Create event worker with Redis callback
	eventWorker, err := s.createEventWorker(&jobData.EventWorkerData, s.chainClients[jobData.EventWorkerData.TriggerChainID])
	if err != nil {
		metrics.TrackCriticalError("worker_creation_failed")
		return fmt.Errorf("failed to create event worker: %w", err)
	}

	// Store worker and job data separately for Redis integration
	s.eventWorkers[jobData.JobID] = eventWorker
	s.jobDataStore[jobData.JobID] = jobData

	// Start worker
	go eventWorker.Start()

	duration := time.Since(startTime)

	s.logger.Info("Event job monitoring started",
		"job_id", jobData.JobID,
		"trigger_chain", jobData.EventWorkerData.TriggerChainID,
		"contract", jobData.EventWorkerData.TriggerContractAddress,
		"event", jobData.EventWorkerData.TriggerEvent,
		"target_chain", jobData.TaskTargetData.TargetChainID,
		"target_contract", jobData.TaskTargetData.TargetContractAddress,
		"target_function", jobData.TaskTargetData.TargetFunction,
		"active_workers", len(s.eventWorkers)+len(s.conditionWorkers),
		"max_workers", s.maxWorkers,
		"duration", duration,
	)

	return nil
}

// createConditionWorker creates a new condition worker instance
func (s *ConditionBasedScheduler) createConditionWorker(conditionWorkerData *types.ConditionWorkerData, httpClient *httppkg.HTTPClient) (*worker.ConditionWorker, error) {
	ctx, cancel := context.WithCancel(s.ctx)

	worker := &worker.ConditionWorker{
		ConditionWorkerData: conditionWorkerData,
		Logger:              s.logger,
		HttpClient:          httpClient,
		Ctx:                 ctx,
		Cancel:              cancel,
		IsActive:            false,
		LastCheckTimestamp:  time.Now(),
		TriggerCallback:     s.handleTriggerNotification,
		CleanupCallback:     s.cleanupJobData,
	}

	return worker, nil
}

// createEventWorker creates a new event worker instance
func (s *ConditionBasedScheduler) createEventWorker(eventWorkerData *types.EventWorkerData, client *ethclient.Client) (*worker.EventWorker, error) {
	ctx, cancel := context.WithCancel(s.ctx)

	// Get current block number
	currentBlock, err := client.BlockNumber(ctx)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to get current block number: %w", err)
	}

	worker := &worker.EventWorker{
		EventWorkerData: eventWorkerData,
		ChainClient:     client,
		Logger:          s.logger,
		Ctx:             ctx,
		Cancel:          cancel,
		LastBlock:       currentBlock,
		IsActive:        false,
		TriggerCallback: s.handleTriggerNotification,
		CleanupCallback: s.cleanupJobData,
	}

	return worker, nil
}

// cleanupJobData removes job data from the scheduler's store when a worker stops
func (s *ConditionBasedScheduler) cleanupJobData(jobID string) error {
	s.notificationMutex.Lock()
	defer s.notificationMutex.Unlock()

	s.workersMutex.Lock()
	defer s.workersMutex.Unlock()

	// Remove job data from store
	delete(s.jobDataStore, jobID)

	s.logger.Debug("Cleaned up job data from store", "job_id", jobID)
	return nil
}

// UnscheduleJob stops and removes a condition worker
func (s *ConditionBasedScheduler) UnscheduleJob(jobID string) error {
	s.notificationMutex.Lock()
	defer s.notificationMutex.Unlock()

	s.workersMutex.Lock()
	defer s.workersMutex.Unlock()

	// Try condition workers first
	if conditionWorker, exists := s.conditionWorkers[jobID]; exists {
		conditionWorker.Stop()
		delete(s.conditionWorkers, jobID)
		delete(s.jobDataStore, jobID) // Clean up job data
	} else if eventWorker, exists := s.eventWorkers[jobID]; exists {
		eventWorker.Stop()
		delete(s.eventWorkers, jobID)
		delete(s.jobDataStore, jobID) // Clean up job data
	} else {
		metrics.TrackCriticalError("job_not_found")
		return fmt.Errorf("job %s is not scheduled", jobID)
	}

	// Update active workers count
	totalWorkers := len(s.conditionWorkers) + len(s.eventWorkers)
	metrics.UpdateActiveWorkers(totalWorkers)

	// Track job completion
	metrics.TrackJobCompleted("unscheduled")

	s.logger.Info("Job unscheduled successfully", "job_id", jobID)
	return nil
}
