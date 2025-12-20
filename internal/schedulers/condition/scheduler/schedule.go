package scheduler

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"

	eventmonitorTypes "github.com/trigg3rX/triggerx-backend/internal/eventmonitor/types"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/scheduler/worker"
	httppkg "github.com/trigg3rX/triggerx-backend/pkg/http"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

// ScheduleJob creates and starts a new condition worker for monitoring
func (s *ConditionBasedScheduler) ScheduleJob(ctx context.Context, jobData *types.ScheduleConditionJobData) error {
	s.workersMutex.Lock()
	defer s.workersMutex.Unlock()

	startTime := time.Now()

	switch jobData.TaskDefinitionID {
	case 3, 4: // Event-based jobs
		if err := s.scheduleEventJob(ctx, jobData, startTime); err != nil {
			return err
		}

	case 5, 6: // Condition-based jobs
		if err := s.scheduleConditionJob(ctx, jobData, startTime); err != nil {
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
func (s *ConditionBasedScheduler) scheduleConditionJob(ctx context.Context, jobData *types.ScheduleConditionJobData, startTime time.Time) error {
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
		go websocketWorker.Start(ctx)
		duration := time.Since(startTime)
		s.logger.Info(ctx, "WebSocket job monitoring started",
			observability.String("job_id", jobData.JobID.String()),
			observability.String("condition_type", jobData.ConditionWorkerData.ConditionType),
			observability.String("value_source", jobData.ConditionWorkerData.ValueSourceUrl),
			observability.Int("active_workers", len(s.eventWorkers)+len(s.conditionWorkers)),
			observability.Int("max_workers", s.maxWorkers),
			observability.Duration("duration", duration),
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
	go conditionWorker.Start(ctx)

	duration := time.Since(startTime)

	// Track condition by type and source
	metrics.TrackConditionByType(jobData.ConditionWorkerData.ConditionType)
	metrics.TrackConditionBySource(jobData.ConditionWorkerData.ValueSourceType)

	s.logger.Info(ctx, "Condition job monitoring started",
		observability.String("job_id", jobData.JobID.String()),
		observability.String("condition_type", jobData.ConditionWorkerData.ConditionType),
		observability.String("value_source", jobData.ConditionWorkerData.ValueSourceUrl),
		observability.Float64("upper_limit", jobData.ConditionWorkerData.UpperLimit),
		observability.Float64("lower_limit", jobData.ConditionWorkerData.LowerLimit),
		observability.Int("active_workers", len(s.eventWorkers)+len(s.conditionWorkers)),
		observability.Int("max_workers", s.maxWorkers),
		observability.Duration("duration", duration),
	)

	return nil
}

// scheduleEventJob handles event-based job scheduling using Event Monitor Service
func (s *ConditionBasedScheduler) scheduleEventJob(ctx context.Context, jobData *types.ScheduleConditionJobData, startTime time.Time) error {
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

	if err := s.eventMonitorClient.Register(ctx, monitoringRequest); err != nil {
		metrics.TrackCriticalError("event_monitor_registration_failed")
		return fmt.Errorf("failed to register with Event Monitor Service: %w", err)
	}

	// Store job data (no local event worker needed)
	s.eventWorkers[jobData.JobID] = nil // Mark as using Event Monitor Service
	s.jobDataStore[jobData.JobID.String()] = jobData

	duration := time.Since(startTime)

	s.logger.Info(ctx, "Event job monitoring started via Event Monitor Service",
		observability.String("job_id", jobData.JobID.String()),
		observability.String("trigger_chain", jobData.EventWorkerData.TriggerChainID),
		observability.String("contract", jobData.EventWorkerData.TriggerContractAddress),
		observability.String("event", jobData.EventWorkerData.TriggerEvent),
		observability.String("target_chain", jobData.TaskTargetData.TargetChainID),
		observability.String("target_contract", jobData.TaskTargetData.TargetContractAddress),
		observability.String("target_function", jobData.TaskTargetData.TargetFunction),
		observability.Int("active_workers", len(s.eventWorkers)+len(s.conditionWorkers)),
		observability.Int("max_workers", s.maxWorkers),
		observability.Duration("duration", duration),
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

// cleanupJobData removes job data from the scheduler's store when a worker stops
func (s *ConditionBasedScheduler) cleanupJobData(ctx context.Context, jobID *big.Int) error {
	s.notificationMutex.Lock()
	defer s.notificationMutex.Unlock()

	s.workersMutex.Lock()
	defer s.workersMutex.Unlock()

	// Remove job data from store
	delete(s.jobDataStore, jobID.String())

	s.logger.Debug(ctx, "Cleaned up job data from store", observability.String("job_id", jobID.String()))
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
func (s *ConditionBasedScheduler) UnregisterEventJob(ctx context.Context, jobID *big.Int) error {
	s.workersMutex.Lock()
	defer s.workersMutex.Unlock()

	// Get the original JobID pointer from jobDataStore to match the map key
	jobData, exists := s.jobDataStore[jobID.String()]
	if !exists || jobData == nil {
		return fmt.Errorf("job %d is not an event job", jobID)
	}

	// Use the original JobID pointer to check if this is an event job
	originalJobID := jobData.JobID
	if _, exists := s.eventWorkers[originalJobID]; !exists {
		return fmt.Errorf("job %d is not an event job", jobID)
	}

	// Unregister from Event Monitor Service
	if s.eventMonitorClient != nil {
		if err := s.eventMonitorClient.Unregister(ctx, jobID.String()); err != nil {
			return fmt.Errorf("failed to unregister from Event Monitor Service: %w", err)
		}
		s.logger.Info(ctx, "Unregistered event job from Event Monitor Service", observability.String("job_id", jobID.String()))
	}

	// Remove from event workers map using the original JobID pointer
	delete(s.eventWorkers, originalJobID)

	// Clean up job data
	delete(s.jobDataStore, jobID.String())

	return nil
}

// unregisterEventJobByPointer unregisters an event job using the original *types.BigInt pointer
// This avoids the pointer-to-value conversion issue when using map keys
func (s *ConditionBasedScheduler) unregisterEventJobByPointer(ctx context.Context, jobID *types.BigInt) error {
	s.workersMutex.Lock()
	defer s.workersMutex.Unlock()

	jobIDStr := jobID.String()

	// Check if this job exists in eventWorkers
	if _, exists := s.eventWorkers[jobID]; !exists {
		// Job may have already been unregistered by another goroutine
		return fmt.Errorf("job %s is not in eventWorkers (may already be unregistered)", jobIDStr)
	}

	// Unregister from Event Monitor Service
	// Note: Event Monitor has its own internal expiration cleanup that may have already
	// removed the subscriber. In that case, we get a "not found" error which is fine -
	// the job is already unregistered which is the desired outcome.
	if s.eventMonitorClient != nil {
		if err := s.eventMonitorClient.Unregister(ctx, jobIDStr); err != nil {
			// Check if this is a "not found" error (job already expired/removed by Event Monitor)
			errStr := err.Error()
			if strings.Contains(errStr, "not found") || strings.Contains(errStr, "404") {
				s.logger.Info(ctx, "Event job already removed from Event Monitor Service (expired)",
					observability.String("job_id", jobIDStr))
			} else {
				// This is a real error, return it
				return fmt.Errorf("failed to unregister from Event Monitor Service: %w", err)
			}
		} else {
			s.logger.Info(ctx, "Unregistered event job from Event Monitor Service", observability.String("job_id", jobIDStr))
		}
	}

	// Remove from event workers map using the original pointer
	delete(s.eventWorkers, jobID)

	// Clean up job data from store
	delete(s.jobDataStore, jobIDStr)

	return nil
}

// UnscheduleJob stops and removes a condition worker
func (s *ConditionBasedScheduler) UnscheduleJob(ctx context.Context, jobID *big.Int) error {
	s.notificationMutex.Lock()
	defer s.notificationMutex.Unlock()

	s.workersMutex.Lock()
	defer s.workersMutex.Unlock()

	jobIDStr := jobID.String()
	var originalJobID *types.BigInt

	// Get the original JobID pointer from jobDataStore to match the map keys
	jobData, exists := s.jobDataStore[jobIDStr]
	if exists && jobData != nil {
		originalJobID = jobData.JobID
	} else {
		// If job data doesn't exist, find the key by iterating through the maps
		for k := range s.conditionWorkers {
			if k != nil && k.String() == jobIDStr {
				originalJobID = k
				break
			}
		}
		if originalJobID == nil {
			for k := range s.eventWorkers {
				if k != nil && k.String() == jobIDStr {
					originalJobID = k
					break
				}
			}
		}
	}

	if originalJobID == nil {
		metrics.TrackCriticalError("job_not_found")
		return fmt.Errorf("job %d is not scheduled", jobID)
	}

	// Try condition workers first
	if conditionWorker, exists := s.conditionWorkers[originalJobID]; exists {
		conditionWorker.Stop(ctx)
		delete(s.conditionWorkers, originalJobID)
		delete(s.jobDataStore, jobIDStr) // Clean up job data
	} else if eventWorker, exists := s.eventWorkers[originalJobID]; exists {
		// If event worker exists, unregister from Event Monitor Service
		if s.eventMonitorClient != nil {
			if err := s.eventMonitorClient.Unregister(ctx, jobIDStr); err != nil {
				s.logger.Warn(ctx, "Failed to unregister from Event Monitor Service",
					observability.String("job_id", jobID.String()),
					observability.Error(err))
				// Continue with cleanup even if unregister fails
			}
		}

		// Stop local worker if it exists (for backward compatibility)
		if eventWorker != nil {
			eventWorker.Stop(ctx)
		}

		delete(s.eventWorkers, originalJobID)
		delete(s.jobDataStore, jobIDStr) // Clean up job data
	} else {
		metrics.TrackCriticalError("job_not_found")
		return fmt.Errorf("job %d is not scheduled", jobID)
	}

	// Update active workers count
	totalWorkers := len(s.conditionWorkers) + len(s.eventWorkers)
	metrics.UpdateActiveWorkers(totalWorkers)

	// Track job completion
	metrics.TrackJobCompleted("unscheduled")

	s.logger.Info(ctx, "Job unscheduled successfully", observability.String("job_id", jobID.String()))
	return nil
}
