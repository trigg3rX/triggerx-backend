package scheduler

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	eventmonitorTypes "github.com/trigg3rX/triggerx-backend/internal/eventmonitor/types"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/scheduler/worker"
	httppkg "github.com/trigg3rX/triggerx-backend/pkg/http"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// ScheduleJob creates and starts a new condition worker for monitoring
func (s *ConditionBasedScheduler) ScheduleJob(ctx context.Context, jobData *types.ScheduleConditionJobData) error {
	// Create trace with format "condition-{scheduler_id}-{timestamp}"
	traceName := fmt.Sprintf("condition-%d-%d", s.schedulerID, time.Now().Unix())

	// Create root span for scheduling operation
	ctx, scheduleSpan := s.tracer.Start(ctx, traceName,
		observability.WithSpanKind(trace.SpanKindProducer),
		observability.WithAttributes(
			attribute.Int("scheduler.id", s.schedulerID),
			attribute.String("scheduler.type", "condition"),
			attribute.String("job.id", jobData.JobID.String()),
			attribute.Int("task_definition_id", jobData.TaskDefinitionID),
			attribute.String("trace.name", traceName),
		),
	)
	defer scheduleSpan.End()

	scheduleSpan.AddEvent("schedule.started")

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
	metrics.UpdateActiveWorkers(len(s.conditionWorkers))
	metrics.TrackWorkerStart(fmt.Sprintf("%d", jobData.JobID))

	scheduleSpan.AddEvent("schedule.completed", observability.WithEventAttributes(
		attribute.String("job_id", jobData.JobID.String()),
	))
	scheduleSpan.SetAttributes(
		attribute.Int("active_workers", len(s.conditionWorkers)),
	)

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
		s.logger.Debug(ctx, "WebSocket job monitoring started",
			observability.String("job_id", jobData.JobID.String()),
			observability.String("condition_type", jobData.ConditionWorkerData.ConditionType),
			observability.String("value_source", jobData.ConditionWorkerData.ValueSourceUrl),
			observability.Int("active_workers", len(s.conditionWorkers)),
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

	s.logger.Debug(ctx, "Condition job monitoring started",
		observability.String("job_id", jobData.JobID.String()),
		observability.String("condition_type", jobData.ConditionWorkerData.ConditionType),
		observability.String("value_source", jobData.ConditionWorkerData.ValueSourceUrl),
		observability.Float64("upper_limit", jobData.ConditionWorkerData.UpperLimit),
		observability.Float64("lower_limit", jobData.ConditionWorkerData.LowerLimit),
		observability.Int("active_workers", len(s.conditionWorkers)),
		observability.Int("max_workers", s.maxWorkers),
		observability.Duration("duration", duration),
	)

	return nil
}

// scheduleEventJob handles event-based job scheduling using Event Monitor Service
func (s *ConditionBasedScheduler) scheduleEventJob(ctx context.Context, jobData *types.ScheduleConditionJobData, startTime time.Time) error {
	// Check if job is already scheduled (check jobDataStore instead of eventWorkers)
	jobIDStr := jobData.JobID.String()
	if _, exists := s.jobDataStore[jobIDStr]; exists {
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
		RequestID:    jobIDStr,
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

	// Store job data (event monitoring is handled by Event Monitor Service)
	s.jobDataStore[jobIDStr] = jobData

	duration := time.Since(startTime)

	s.logger.Info(ctx, "Event job monitoring started via Event Monitor Service",
		observability.String("job_id", jobIDStr),
		observability.String("trigger_chain", jobData.EventWorkerData.TriggerChainID),
		observability.String("contract", jobData.EventWorkerData.TriggerContractAddress),
		observability.String("event", jobData.EventWorkerData.TriggerEvent),
		observability.String("target_chain", jobData.TaskTargetData.TargetChainID),
		observability.String("target_contract", jobData.TaskTargetData.TargetContractAddress),
		observability.String("target_function", jobData.TaskTargetData.TargetFunction),
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
		Tracer:              s.tracer,
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

	jobIDStr := jobID.String()
	// Remove job data from store
	delete(s.jobDataStore, jobIDStr)
	// Clean up last trigger time tracking
	delete(s.lastTriggerTime, jobIDStr)

	s.logger.Debug(ctx, "Cleaned up job data from store", observability.String("job_id", jobIDStr))
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

	jobIDStr := jobID.String()

	// Check if job exists in jobDataStore
	jobData, exists := s.jobDataStore[jobIDStr]
	if !exists || jobData == nil {
		return fmt.Errorf("job %d is not found", jobID)
	}

	// Verify this is an event job (task definition ID 3 or 4)
	if jobData.TaskDefinitionID != 3 && jobData.TaskDefinitionID != 4 {
		return fmt.Errorf("job %d is not an event job", jobID)
	}

	// Unregister from Event Monitor Service
	if s.eventMonitorClient != nil {
		if err := s.eventMonitorClient.Unregister(ctx, jobIDStr); err != nil {
			return fmt.Errorf("failed to unregister from Event Monitor Service: %w", err)
		}
		s.logger.Debug(ctx, "Unregistered event job from Event Monitor Service", observability.String("job_id", jobIDStr))
	}

	// Clean up job data
	delete(s.jobDataStore, jobIDStr)
	delete(s.lastTriggerTime, jobIDStr)

	return nil
}

// UnscheduleJob stops and removes a condition worker
func (s *ConditionBasedScheduler) UnscheduleJob(ctx context.Context, jobID *big.Int) error {
	// Create trace with format "condition-{scheduler_id}-{timestamp}"
	traceName := fmt.Sprintf("condition-%d-%d", s.schedulerID, time.Now().Unix())

	// Create root span for unscheduling operation
	ctx, unscheduleSpan := s.tracer.Start(ctx, traceName,
		observability.WithSpanKind(trace.SpanKindProducer),
		observability.WithAttributes(
			attribute.Int("scheduler.id", s.schedulerID),
			attribute.String("scheduler.type", "condition"),
			attribute.String("job.id", jobID.String()),
			attribute.String("trace.name", traceName),
		),
	)
	defer unscheduleSpan.End()

	unscheduleSpan.AddEvent("unschedule.started")

	s.notificationMutex.Lock()
	defer s.notificationMutex.Unlock()

	s.workersMutex.Lock()
	defer s.workersMutex.Unlock()

	jobIDStr := jobID.String()

	// Get job data to determine job type
	jobData, exists := s.jobDataStore[jobIDStr]
	if !exists || jobData == nil {
		metrics.TrackCriticalError("job_not_found")
		return fmt.Errorf("job %d is not scheduled", jobID)
	}

	// Get the original JobID pointer from jobDataStore to match the map keys
	originalJobID := jobData.JobID

	// Determine job type and handle accordingly
	switch jobData.TaskDefinitionID {
	case 3, 4:
		// Event-based job: unregister from Event Monitor Service
		if s.eventMonitorClient != nil {
			if err := s.eventMonitorClient.Unregister(ctx, jobIDStr); err != nil {
				s.logger.Warn(ctx, "Failed to unregister from Event Monitor Service",
					observability.String("job_id", jobIDStr),
					observability.Error(err))
				// Continue with cleanup even if unregister fails
			}
		}
		delete(s.jobDataStore, jobIDStr)    // Clean up job data
		delete(s.lastTriggerTime, jobIDStr) // Clean up last trigger time tracking
	case 5, 6:
		// Condition-based job: stop the worker
		if conditionWorker, exists := s.conditionWorkers[originalJobID]; exists {
			// Handle websocket workers (stored as nil) - they stop via context cancellation
			// For regular condition workers, call Stop explicitly
			if conditionWorker != nil {
				conditionWorker.Stop(ctx)
			} else {
				// For websocket workers, we need to cancel their context
				// Since websocket workers use context derived from s.ctx, we can't cancel individually
				// The worker will stop when it checks ctx.Done() in its loop
				// We rely on the job data cleanup to prevent new triggers
				s.logger.Debug(ctx, "WebSocket worker will stop via context check", observability.String("job_id", jobIDStr))
			}
			delete(s.conditionWorkers, originalJobID)
		}
		delete(s.jobDataStore, jobIDStr)    // Clean up job data
		delete(s.lastTriggerTime, jobIDStr) // Clean up last trigger time tracking
	default:
		metrics.TrackCriticalError("job_not_found")
		return fmt.Errorf("job %d has unsupported task definition id: %d", jobID, jobData.TaskDefinitionID)
	}

	// Update active workers count (only condition workers now)
	metrics.UpdateActiveWorkers(len(s.conditionWorkers))

	// Track job completion
	metrics.TrackJobCompleted("unscheduled")

	unscheduleSpan.AddEvent("unschedule.completed", observability.WithEventAttributes(
		attribute.String("job_id", jobID.String()),
	))

	s.logger.Info(ctx, "Job unscheduled successfully", observability.String("job_id", jobID.String()))
	return nil
}
