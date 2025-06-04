package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

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
		return fmt.Errorf("job %d is already scheduled", jobData.JobID)
	}

	// Check if we've reached the maximum number of workers
	if len(s.workers) >= s.maxWorkers {
		return fmt.Errorf("maximum number of workers (%d) reached, cannot schedule job %d", s.maxWorkers, jobData.JobID)
	}

	// Get chain client
	s.clientsMutex.RLock()
	client, exists := s.chainClients[jobData.TriggerChainID]
	s.clientsMutex.RUnlock()

	if !exists {
		return fmt.Errorf("unsupported chain ID: %s", jobData.TriggerChainID)
	}

	// Create job worker
	worker, err := s.createJobWorker(jobData, client)
	if err != nil {
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

	worker := &worker.EventWorker{
		Job:          jobData,
		Client:       client,
		Logger:       s.logger,
		Ctx:          ctx,
		Cancel:       cancel,
		EventSig:     eventSig,
		ContractAddr: contractAddr,
		LastBlock:    0,
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
