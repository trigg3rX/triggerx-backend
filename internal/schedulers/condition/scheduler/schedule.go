package scheduler

import (
	"context"
	"fmt"
	"math/big"
	// "strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	// nodeclient "github.com/trigg3rX/triggerx-backend/pkg/client/nodeclient"

	eventmonitorTypes "github.com/trigg3rX/triggerx-backend/internal/eventmonitor/types"
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
	metrics.TrackWorkerStart(fmt.Sprintf("%d", jobData.JobID))

	return nil
}

// scheduleConditionJob handles condition-based job scheduling
func (s *ConditionBasedScheduler) scheduleConditionJob(jobData *types.ScheduleConditionJobData, startTime time.Time) error {
	// Check if job is already scheduled
	if _, exists := s.conditionWorkers[jobData.JobID]; exists {
		metrics.TrackCriticalError("duplicate_job_schedule")
		return fmt.Errorf("job %d is already scheduled", jobData.JobID)
	}
	// WebSocket jobs: check and schedule
	if jobData.ConditionWorkerData.ValueSourceType == worker.SourceTypeWebSocket {
		websocketWorker, err := s.createWebSocketWorker(&jobData.ConditionWorkerData)
		if err != nil {
			metrics.TrackCriticalError("websocket_worker_creation_failed")
			return fmt.Errorf("failed to create websocket worker: %w", err)
		}
		s.conditionWorkers[jobData.JobID] = nil // Or: s.websocketWorkers[jobData.JobID] = websocketWorker (if struct field added)
		s.jobDataStore[jobData.JobID.String()] = jobData
		go websocketWorker.Start()
		duration := time.Since(startTime)
		s.logger.Info("WebSocket job monitoring started",
			"job_id", jobData.JobID,
			"condition_type", jobData.ConditionWorkerData.ConditionType,
			"value_source", jobData.ConditionWorkerData.ValueSourceUrl,
			"active_workers", len(s.eventWorkers)+len(s.conditionWorkers),
			"max_workers", s.maxWorkers,
			"duration", duration,
		)
		return nil
	}
	// Normal jobs:
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
	s.jobDataStore[jobData.JobID.String()] = jobData

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

// scheduleEventJob handles event-based job scheduling using Event Monitor Service
func (s *ConditionBasedScheduler) scheduleEventJob(jobData *types.ScheduleConditionJobData, startTime time.Time) error {
	// Check if job is already scheduled
	if _, exists := s.eventWorkers[jobData.JobID]; exists {
		metrics.TrackCriticalError("duplicate_job_schedule")
		return fmt.Errorf("job %d is already scheduled", jobData.JobID)
	}

	// Validate contract address
	if !common.IsHexAddress(jobData.EventWorkerData.TriggerContractAddress) {
		metrics.TrackCriticalError("invalid_contract_address")
		return fmt.Errorf("invalid contract address: %s", jobData.EventWorkerData.TriggerContractAddress)
	}

	// Register with Event Monitor Service
	monitoringRequest := &eventmonitorTypes.MonitoringRequest{
		RequestID:    jobData.JobID.String(),
		ChainID:      jobData.EventWorkerData.TriggerChainID,
		ContractAddr: jobData.EventWorkerData.TriggerContractAddress,
		EventSig:     jobData.EventWorkerData.TriggerEvent,
		WebhookURL:   s.webhookURL,
		ExpiresAt:    jobData.EventWorkerData.ExpirationTime,
	}

	// Add filter parameters if provided
	if jobData.EventWorkerData.EventFilterParaName != "" && jobData.EventWorkerData.EventFilterValue != "" {
		monitoringRequest.FilterParam = jobData.EventWorkerData.EventFilterParaName
		monitoringRequest.FilterValue = jobData.EventWorkerData.EventFilterValue
	}

	if err := s.eventMonitorClient.Register(monitoringRequest); err != nil {
		metrics.TrackCriticalError("event_monitor_registration_failed")
		return fmt.Errorf("failed to register with Event Monitor Service: %w", err)
	}

	// Store job data (no local event worker needed)
	s.eventWorkers[jobData.JobID] = nil // Mark as using Event Monitor Service
	s.jobDataStore[jobData.JobID.String()] = jobData

	duration := time.Since(startTime)

	s.logger.Info("Event job monitoring started via Event Monitor Service",
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
		TriggerCallback:     s.HandleTriggerNotification,
		CleanupCallback:     s.cleanupJobData,
	}

	return worker, nil
}

// Create a new websocket worker
func (s *ConditionBasedScheduler) createWebSocketWorker(conditionWorkerData *types.ConditionWorkerData) (*worker.WebSocketWorker, error) {
	ctx, cancel := context.WithCancel(s.ctx)
	wsConfig := &worker.WebSocketConfig{
		URL: conditionWorkerData.ValueSourceUrl,
	}
	worker := &worker.WebSocketWorker{
		WebSocketConfig:     wsConfig,
		ConditionWorkerData: conditionWorkerData,
		Logger:              s.logger,
		Ctx:                 ctx,
		Cancel:              cancel,
		IsActive:            false,
		TriggerCallback:     s.HandleTriggerNotification,
		CleanupCallback:     s.cleanupJobData,
	}
	return worker, nil
}

// createEventWorker creates a new event worker instance
// func (s *ConditionBasedScheduler) createEventWorker(eventWorkerData *types.EventWorkerData, client *nodeclient.NodeClient) (*worker.EventWorker, error) {
// 	ctx, cancel := context.WithCancel(s.ctx)

// 	// Get current block number
// 	blockHex, err := client.EthBlockNumber(ctx)
// 	if err != nil {
// 		cancel()
// 		return nil, fmt.Errorf("failed to get current block number: %w", err)
// 	}

// 	// Convert hex to uint64
// 	currentBlock, err := hexToUint64(blockHex)
// 	if err != nil {
// 		cancel()
// 		return nil, fmt.Errorf("failed to parse block number: %w", err)
// 	}

// 	worker := &worker.EventWorker{
// 		EventWorkerData: eventWorkerData,
// 		ChainClient:     client,
// 		Logger:          s.logger,
// 		Ctx:             ctx,
// 		Cancel:          cancel,
// 		LastBlock:       currentBlock,
// 		IsActive:        false,
// 		TriggerCallback: s.HandleTriggerNotification,
// 		CleanupCallback: s.cleanupJobData,
// 	}

// 	return worker, nil
// }

// hexToUint64 converts a hex string (with or without 0x prefix) to uint64
// func hexToUint64(hexStr string) (uint64, error) {
// 	// Remove 0x prefix if present
// 	if len(hexStr) >= 2 && hexStr[:2] == "0x" {
// 		hexStr = hexStr[2:]
// 	}
// 	return strconv.ParseUint(hexStr, 16, 64)
// }

// cleanupJobData removes job data from the scheduler's store when a worker stops
func (s *ConditionBasedScheduler) cleanupJobData(jobID *big.Int) error {
	s.notificationMutex.Lock()
	defer s.notificationMutex.Unlock()

	s.workersMutex.Lock()
	defer s.workersMutex.Unlock()

	// Remove job data from store
	delete(s.jobDataStore, jobID.String())

	s.logger.Debug("Cleaned up job data from store", "job_id", jobID)
	return nil
}

// GetJobData retrieves job data by job ID (thread-safe)
func (s *ConditionBasedScheduler) GetJobData(jobID *big.Int) (*types.ScheduleConditionJobData, error) {
	s.workersMutex.RLock()
	defer s.workersMutex.RUnlock()

	jobData, exists := s.jobDataStore[jobID.String()]
	if !exists || jobData == nil {
		return nil, fmt.Errorf("job data not found for job %d", jobID)
	}

	return jobData, nil
}

// UnregisterEventJob unregisters an event job from Event Monitor Service
func (s *ConditionBasedScheduler) UnregisterEventJob(jobID *big.Int) error {
	s.workersMutex.Lock()
	defer s.workersMutex.Unlock()

	// Check if this is an event job
	if _, exists := s.eventWorkers[types.NewBigInt(jobID)]; !exists {
		return fmt.Errorf("job %d is not an event job", jobID)
	}

	// Unregister from Event Monitor Service
	if s.eventMonitorClient != nil {
		if err := s.eventMonitorClient.Unregister(jobID.String()); err != nil {
			return fmt.Errorf("failed to unregister from Event Monitor Service: %w", err)
		}
		s.logger.Info("Unregistered event job from Event Monitor Service", "job_id", jobID)
	}

	// Remove from event workers map
	delete(s.eventWorkers, types.NewBigInt(jobID))

	// Clean up job data
	delete(s.jobDataStore, jobID.String())

	return nil
}

// UnscheduleJob stops and removes a condition worker
func (s *ConditionBasedScheduler) UnscheduleJob(jobID *big.Int) error {
	s.notificationMutex.Lock()
	defer s.notificationMutex.Unlock()

	s.workersMutex.Lock()
	defer s.workersMutex.Unlock()

	// Try condition workers first
	if conditionWorker, exists := s.conditionWorkers[types.NewBigInt(jobID)]; exists {
		conditionWorker.Stop()
		delete(s.conditionWorkers, types.NewBigInt(jobID))
		delete(s.jobDataStore, jobID.String()) // Clean up job data
	} else if eventWorker, exists := s.eventWorkers[types.NewBigInt(jobID)]; exists {
		// If event worker exists, unregister from Event Monitor Service
		if s.eventMonitorClient != nil {
			if err := s.eventMonitorClient.Unregister(jobID.String()); err != nil {
				s.logger.Warn("Failed to unregister from Event Monitor Service",
					"job_id", jobID,
					"error", err)
				// Continue with cleanup even if unregister fails
			}
		}

		// Stop local worker if it exists (for backward compatibility)
		if eventWorker != nil {
			eventWorker.Stop()
		}

		delete(s.eventWorkers, types.NewBigInt(jobID))
		delete(s.jobDataStore, jobID.String()) // Clean up job data
	} else {
		metrics.TrackCriticalError("job_not_found")
		return fmt.Errorf("job %d is not scheduled", jobID)
	}

	// Update active workers count
	totalWorkers := len(s.conditionWorkers) + len(s.eventWorkers)
	metrics.UpdateActiveWorkers(totalWorkers)

	// Track job completion
	metrics.TrackJobCompleted("unscheduled")

	s.logger.Info("Job unscheduled successfully", "job_id", jobID)
	return nil
}
