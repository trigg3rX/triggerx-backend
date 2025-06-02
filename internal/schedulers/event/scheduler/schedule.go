package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	redisx "github.com/trigg3rX/triggerx-backend/internal/redis"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/event/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/event/scheduler/worker"
	commonTypes "github.com/trigg3rX/triggerx-backend/pkg/types"
)

// ScheduleJob creates and starts a new job worker
func (s *EventBasedScheduler) ScheduleJob(jobData *commonTypes.EventJobData) error {
	s.workersMutex.Lock()
	defer s.workersMutex.Unlock()

	startTime := time.Now()

	// Check if job is already scheduled
	if _, exists := s.workers[jobData.JobID]; exists {
		// Add job scheduling failure event to Redis stream
		if redisx.IsAvailable() {
			failureEvent := map[string]interface{}{
				"event_type":       "job_schedule_failed",
				"job_id":           jobData.JobID,
				"manager_id":       s.managerID,
				"error":            "job already scheduled",
				"trigger_chain_id": jobData.TriggerChainID,
				"contract_address": jobData.TriggerContractAddress,
				"trigger_event":    jobData.TriggerEvent,
				"failed_at":        startTime.Unix(),
			}
			err := redisx.AddJobToStream(redisx.JobsRetryEventStream, failureEvent)
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
				"event_type":       "job_schedule_failed",
				"job_id":           jobData.JobID,
				"manager_id":       s.managerID,
				"error":            fmt.Sprintf("maximum workers (%d) reached", s.maxWorkers),
				"current_workers":  len(s.workers),
				"max_workers":      s.maxWorkers,
				"trigger_chain_id": jobData.TriggerChainID,
				"contract_address": jobData.TriggerContractAddress,
				"trigger_event":    jobData.TriggerEvent,
				"failed_at":        startTime.Unix(),
			}
			err := redisx.AddJobToStream(redisx.JobsRetryEventStream, failureEvent)
			if err != nil {
				s.logger.Warnf("Failed to add job scheduling failure event to Redis stream: %v", err)
			}
		}
		return fmt.Errorf("maximum number of workers (%d) reached, cannot schedule job %d", s.maxWorkers, jobData.JobID)
	}

	// Get chain client
	s.clientsMutex.RLock()
	client, exists := s.chainClients[jobData.TriggerChainID]
	s.clientsMutex.RUnlock()

	if !exists {
		// Add job scheduling failure event to Redis stream
		if redisx.IsAvailable() {
			failureEvent := map[string]interface{}{
				"event_type":       "job_schedule_failed",
				"job_id":           jobData.JobID,
				"manager_id":       s.managerID,
				"error":            fmt.Sprintf("unsupported chain ID: %s", jobData.TriggerChainID),
				"trigger_chain_id": jobData.TriggerChainID,
				"contract_address": jobData.TriggerContractAddress,
				"trigger_event":    jobData.TriggerEvent,
				"failed_at":        startTime.Unix(),
			}
			err := redisx.AddJobToStream(redisx.JobsRetryEventStream, failureEvent)
			if err != nil {
				s.logger.Warnf("Failed to add job scheduling failure event to Redis stream: %v", err)
			}
		}
		return fmt.Errorf("unsupported chain ID: %s", jobData.TriggerChainID)
	}

	// Create job worker
	worker, err := s.createJobWorker(jobData, client)
	if err != nil {
		// Add job scheduling failure event to Redis stream
		if redisx.IsAvailable() {
			failureEvent := map[string]interface{}{
				"event_type":       "job_schedule_failed",
				"job_id":           jobData.JobID,
				"manager_id":       s.managerID,
				"error":            fmt.Sprintf("failed to create worker: %v", err),
				"trigger_chain_id": jobData.TriggerChainID,
				"contract_address": jobData.TriggerContractAddress,
				"trigger_event":    jobData.TriggerEvent,
				"failed_at":        startTime.Unix(),
			}
			if err := redisx.AddJobToStream(redisx.JobsRetryEventStream, failureEvent); err != nil {
				s.logger.Warnf("Failed to add job scheduling failure event to Redis stream: %v", err)
			}
		}
		return fmt.Errorf("failed to create job worker: %w", err)
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
			"event_type":               "job_scheduled",
			"job_id":                   jobData.JobID,
			"manager_id":               s.managerID,
			"trigger_chain_id":         jobData.TriggerChainID,
			"trigger_contract_address": jobData.TriggerContractAddress,
			"trigger_event":            jobData.TriggerEvent,
			"target_chain_id":          jobData.TargetChainID,
			"target_contract_address":  jobData.TargetContractAddress,
			"target_function":          jobData.TargetFunction,
			"recurring":                jobData.Recurring,
			"active_workers":           len(s.workers),
			"max_workers":              s.maxWorkers,
			"cache_available":          s.cache != nil,
			"scheduled_at":             startTime.Unix(),
			"duration_ms":              duration.Milliseconds(),
			"status":                   "scheduled",
		}

		if err := redisx.AddJobToStream(redisx.JobsReadyEventStream, jobContext); err != nil {
			s.logger.Warnf("Failed to add job scheduling event to Redis stream: %v", err)
		}
	}

	s.logger.Info("Job scheduled successfully",
		"job_id", jobData.JobID,
		"trigger_chain", jobData.TriggerChainID,
		"contract", jobData.TriggerContractAddress,
		"event", jobData.TriggerEvent,
		"target_chain", jobData.TargetChainID,
		"target_contract", jobData.TargetContractAddress,
		"target_function", jobData.TargetFunction,
		"active_workers", len(s.workers),
		"max_workers", s.maxWorkers,
		"duration", duration,
	)

	return nil
}

// createJobWorker creates a new job worker instance
func (s *EventBasedScheduler) createJobWorker(jobData *commonTypes.EventJobData, client *ethclient.Client) (*worker.EventWorker, error) {
	ctx, cancel := context.WithCancel(s.ctx)

	// Validate contract address
	if !common.IsHexAddress(jobData.TriggerContractAddress) {
		cancel()
		return nil, fmt.Errorf("invalid contract address: %s", jobData.TriggerContractAddress)
	}
	contractAddr := common.HexToAddress(jobData.TriggerContractAddress)

	// Calculate event signature
	eventSig := crypto.Keccak256Hash([]byte(jobData.TriggerEvent))

	// Get current block number (with caching)
	currentBlock, err := s.getCachedOrFetchBlockNumber(client, jobData.TriggerChainID)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to get current block number: %w", err)
	}

	worker := &worker.EventWorker{
		Job:          jobData,
		Client:       client,
		Logger:       s.logger,
		Cache:        s.cache,
		Ctx:          ctx,
		Cancel:       cancel,
		EventSig:     eventSig,
		ContractAddr: contractAddr,
		LastBlock:    currentBlock,
		IsActive:    false,
		ManagerID:    s.managerID,
	}

	return worker, nil
}

// UnscheduleJob stops and removes a job worker
func (s *EventBasedScheduler) UnscheduleJob(jobID int64) error {
	s.workersMutex.Lock()
	defer s.workersMutex.Unlock()

	startTime := time.Now()

	worker, exists := s.workers[jobID]
	if !exists {
		// Add job unscheduling failure event to Redis stream
		if redisx.IsAvailable() {
			failureEvent := map[string]interface{}{
				"event_type": "job_unschedule_failed",
				"job_id":     jobID,
				"manager_id": s.managerID,
				"error":      "job not found",
				"failed_at":  startTime.Unix(),
			}
			err := redisx.AddJobToStream(redisx.JobsRetryEventStream, failureEvent)
			if err != nil {
				s.logger.Warnf("Failed to add job unscheduling failure event to Redis stream: %v", err)
			}
		}
		return fmt.Errorf("job %d is not scheduled", jobID)
	}

	// Capture job details before stopping
	jobDetails := map[string]interface{}{
		"trigger_chain_id":         worker.Job.TriggerChainID,
		"trigger_contract_address": worker.Job.TriggerContractAddress,
		"trigger_event":            worker.Job.TriggerEvent,
		"target_chain_id":          worker.Job.TargetChainID,
		"target_contract_address":  worker.Job.TargetContractAddress,
		"target_function":          worker.Job.TargetFunction,
		"last_processed_block":     worker.LastBlock,
		"was_running":              worker.IsRunning(),
	}

	// Stop worker
	worker.Stop()

	// Remove from workers map
	delete(s.workers, jobID)

	// Update metrics
	metrics.JobsRunning.Dec()

	duration := time.Since(startTime)

	// Add comprehensive job unscheduling success event to Redis stream
	if redisx.IsAvailable() {
		jobContext := map[string]interface{}{
			"event_type":               "job_unscheduled",
			"job_id":                   jobID,
			"manager_id":               s.managerID,
			"trigger_chain_id":         jobDetails["trigger_chain_id"],
			"trigger_contract_address": jobDetails["trigger_contract_address"],
			"trigger_event":            jobDetails["trigger_event"],
			"target_chain_id":          jobDetails["target_chain_id"],
			"target_contract_address":  jobDetails["target_contract_address"],
			"target_function":          jobDetails["target_function"],
			"last_processed_block":     jobDetails["last_processed_block"],
			"was_running":              jobDetails["was_running"],
			"remaining_workers":        len(s.workers),
			"max_workers":              s.maxWorkers,
			"unscheduled_at":           startTime.Unix(),
			"duration_ms":              duration.Milliseconds(),
			"status":                   "unscheduled",
		}

		if err := redisx.AddJobToStream(redisx.JobsReadyEventStream, jobContext); err != nil {
			s.logger.Warnf("Failed to add job unscheduling event to Redis stream: %v", err)
		}
	}

	s.logger.Info("Job unscheduled successfully",
		"job_id", jobID,
		"trigger_chain", jobDetails["trigger_chain_id"],
		"contract", jobDetails["trigger_contract_address"],
		"event", jobDetails["trigger_event"],
		"was_running", jobDetails["was_running"],
		"remaining_workers", len(s.workers),
		"duration", duration,
	)

	return nil
}
