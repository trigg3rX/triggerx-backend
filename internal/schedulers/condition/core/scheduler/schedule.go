package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/core/scheduler/worker"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/metrics"
	httppkg "github.com/trigg3rX/triggerx-backend/pkg/http"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// ScheduleJob creates and starts a new condition worker for monitoring
func (s *ConditionBasedScheduler) ScheduleJob(ctx context.Context, request *types.ScheduleConditionJobRequest) error {
	// Create trace with format "condition-{scheduler_id}-{timestamp}"
	traceName := fmt.Sprintf("condition-%s-%d", s.schedulerID, time.Now().Unix())

	// Create root span for scheduling operation
	ctx, scheduleSpan := s.tracer.Start(ctx, traceName,
		observability.WithSpanKind(trace.SpanKindProducer),
		observability.WithAttributes(
			attribute.String("scheduler.id", s.schedulerID),
			attribute.String("scheduler.type", "condition"),
			attribute.String("job.id", request.JobID),
			attribute.Int("task_definition_id", request.TaskDefinitionID),
			attribute.String("trace.name", traceName),
		),
	)
	defer scheduleSpan.End()

	scheduleSpan.AddEvent("schedule.started")

	// Read job data from database
	jobData, err := s.readJobDataFromDB(ctx, request.JobID, request.TaskDefinitionID)
	if err != nil {
		scheduleSpan.RecordError(err)
		return fmt.Errorf("failed to read job data from DB: %w", err)
	}

	s.workersMutex.Lock()
	defer s.workersMutex.Unlock()

	startTime := time.Now()

	switch jobData.TaskDefinitionID {
	case 3, 4, 8: // Event-based jobs
		if err := s.scheduleEventJob(ctx, jobData, startTime); err != nil {
			return err
		}

	case 5, 6, 9: // Condition-based jobs
		if err := s.scheduleConditionJob(ctx, jobData, startTime); err != nil {
			return err
		}

	default:
		return fmt.Errorf("unsupported task definition id: %d", jobData.TaskDefinitionID)
	}

	// Store job data for notifications
	s.jobDataStore[request.JobID] = jobData

	// Update metrics
	metrics.TrackJobScheduled()
	metrics.UpdateActiveWorkers(len(s.conditionWorkers))
	metrics.TrackWorkerStart(jobData.JobID)

	scheduleSpan.AddEvent("schedule.completed", observability.WithEventAttributes(
		attribute.String("job_id", jobData.JobID),
	))
	scheduleSpan.SetAttributes(
		attribute.Int("active_workers", len(s.conditionWorkers)),
	)

	return nil
}

// readJobDataFromDB reads job data from database and creates ScheduleConditionJobData
func (s *ConditionBasedScheduler) readJobDataFromDB(ctx context.Context, jobID string, taskDefinitionID int) (*types.ScheduleConditionJobData, error) {
	var jobData *types.ScheduleConditionJobData

	switch taskDefinitionID {
	case 3, 4, 8: // Event-based jobs
		eventEntity, err := s.jobRepository.GetEventJobByJobID(jobID)
		if err != nil {
			return nil, fmt.Errorf("failed to get event job: %w", err)
		}

		// Create EventWorkerData
		eventWorkerData := &types.EventWorkerData{
			JobID:                  eventEntity.JobID,
			ExpirationTime:         eventEntity.ExpirationTime,
			Recurring:              eventEntity.Recurring,
			TriggerChainID:         eventEntity.TriggerChainID,
			TriggerContractAddress: eventEntity.TriggerContractAddress,
			TriggerEvent:           eventEntity.TriggerEvent,
			EventFilterParaName:    eventEntity.EventFilterParaName,
			EventFilterValue:       eventEntity.EventFilterValue,
		}

		// Create TaskTargetData
		taskTargetData := types.TaskTargetData{
			JobID:                eventEntity.JobID,
			TaskID:               0, // Will be set when task is created
			TaskDefinitionID:      eventEntity.TaskDefinitionID,
			TargetChainID:        eventEntity.TargetChainID,
			TargetContractAddress: eventEntity.TargetContractAddress,
			TargetFunction:       eventEntity.TargetFunction,
			ABI:                  eventEntity.ABI,
			ArgType:              eventEntity.ArgType,
			Arguments:            eventEntity.Arguments,
			ExecutionScriptURL:   eventEntity.ExecutionScriptURL,
			ExecutionScriptLanguage: eventEntity.ExecutionScriptLanguage,
			ExecutionScriptHash:   eventEntity.ExecutionScriptHash,
			MaxExecutionTime:     eventEntity.MaxExecutionTime,
			ChallengePeriod:      eventEntity.ChallengePeriod,
		}

		jobData = &types.ScheduleConditionJobData{
			JobID:              eventEntity.JobID,
			TaskDefinitionID:   eventEntity.TaskDefinitionID,
			Network:            types.Network(eventEntity.Network),
			EventWorkerData:    eventWorkerData,
			ConditionWorkerData: nil,
			TaskTargetData:     taskTargetData,
		}

	case 5, 6, 9: // Condition-based jobs
		conditionEntity, err := s.jobRepository.GetConditionJobByJobID(jobID)
		if err != nil {
			return nil, fmt.Errorf("failed to get condition job: %w", err)
		}

		// Create ConditionWorkerData
		conditionWorkerData := &types.ConditionWorkerData{
			JobID:            conditionEntity.JobID,
			ExpirationTime:   conditionEntity.ExpirationTime,
			Recurring:        conditionEntity.Recurring,
			ConditionType:    conditionEntity.ConditionType,
			SelectedKeyRoute: conditionEntity.SelectedKeyRoute,
			UpperLimit:       conditionEntity.UpperLimit,
			LowerLimit:       conditionEntity.LowerLimit,
			ValueSourceType:  conditionEntity.ValueSourceType,
			ValueSourceUrl:   conditionEntity.ValueSourceURL,
		}

		// Create TaskTargetData
		taskTargetData := types.TaskTargetData{
			JobID:                conditionEntity.JobID,
			TaskID:               0, // Will be set when task is created
			TaskDefinitionID:     conditionEntity.TaskDefinitionID,
			TargetChainID:        conditionEntity.TargetChainID,
			TargetContractAddress: conditionEntity.TargetContractAddress,
			TargetFunction:       conditionEntity.TargetFunction,
			ABI:                  conditionEntity.ABI,
			ArgType:              conditionEntity.ArgType,
			Arguments:            conditionEntity.Arguments,
			ExecutionScriptURL:   conditionEntity.ExecutionScriptURL,
			ExecutionScriptLanguage: conditionEntity.ExecutionScriptLanguage,
			ExecutionScriptHash:   conditionEntity.ExecutionScriptHash,
			MaxExecutionTime:     conditionEntity.MaxExecutionTime,
			ChallengePeriod:      conditionEntity.ChallengePeriod,
		}

		jobData = &types.ScheduleConditionJobData{
			JobID:              conditionEntity.JobID,
			TaskDefinitionID:   conditionEntity.TaskDefinitionID,
			Network:            types.Network(conditionEntity.Network),
			EventWorkerData:    nil,
			ConditionWorkerData: conditionWorkerData,
			TaskTargetData:     taskTargetData,
		}

	default:
		return nil, fmt.Errorf("unsupported task definition id: %d", taskDefinitionID)
	}

	return jobData, nil
}

// scheduleConditionJob handles condition-based job scheduling
func (s *ConditionBasedScheduler) scheduleConditionJob(ctx context.Context, jobData *types.ScheduleConditionJobData, startTime time.Time) error {
	// Check if job is already scheduled
	if _, exists := s.conditionWorkers[jobData.JobID]; exists {
		metrics.TrackCriticalError("duplicate_job_schedule")
		return fmt.Errorf("job %s is already scheduled", jobData.JobID)
	}
	// WebSocket jobs: check and schedule
	if jobData.ConditionWorkerData.ValueSourceType == worker.SourceTypeWebSocket {
		websocketWorker, err := s.createWebSocketWorker(jobData.ConditionWorkerData)
		if err != nil {
			metrics.TrackCriticalError("websocket_worker_creation_failed")
			return fmt.Errorf("failed to create websocket worker: %w", err)
		}
		s.conditionWorkers[jobData.JobID] = nil // Or: s.websocketWorkers[jobData.JobID] = websocketWorker (if struct field added)
		go websocketWorker.Start(ctx)
		duration := time.Since(startTime)
		s.logger.Debug(ctx, "WebSocket job monitoring started",
			observability.String("job_id", jobData.JobID),
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
	conditionWorker, err := s.createConditionWorker(jobData.ConditionWorkerData, s.HTTPClient)
	if err != nil {
		metrics.TrackCriticalError("worker_creation_failed")
		return fmt.Errorf("failed to create condition worker: %w", err)
	}

	// Store worker
	s.conditionWorkers[jobData.JobID] = conditionWorker

	// Start worker
	go conditionWorker.Start(ctx)

	duration := time.Since(startTime)

	// Track condition by type and source
	metrics.TrackConditionByType(jobData.ConditionWorkerData.ConditionType)
	metrics.TrackConditionBySource(jobData.ConditionWorkerData.ValueSourceType)

	s.logger.Debug(ctx, "Condition job monitoring started",
		observability.String("job_id", jobData.JobID),
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
	jobIDStr := jobData.JobID
	if _, exists := s.jobDataStore[jobIDStr]; exists {
		metrics.TrackCriticalError("duplicate_job_schedule")
		return fmt.Errorf("job %s is already scheduled", jobData.JobID)
	}

	// Validate contract address
	if !common.IsHexAddress(jobData.EventWorkerData.TriggerContractAddress) {
		metrics.TrackCriticalError("invalid_contract_address")
		return fmt.Errorf("invalid contract address: %s", jobData.EventWorkerData.TriggerContractAddress)
	}

	// Register with Event Monitor Service
	monitoringRequest := &types.MonitoringRequest{
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

	// Job data is already stored in ScheduleJob

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

	w := &worker.ConditionWorker{
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

	return w, nil
}

// Create a new websocket worker
func (s *ConditionBasedScheduler) createWebSocketWorker(conditionWorkerData *types.ConditionWorkerData) (*worker.WebSocketWorker, error) {
	ctx, cancel := context.WithCancel(s.ctx)
	wsConfig := &worker.WebSocketConfig{
		URL: conditionWorkerData.ValueSourceUrl,
	}
	w := &worker.WebSocketWorker{
		WebSocketConfig:     wsConfig,
		ConditionWorkerData: conditionWorkerData,
		Logger:              s.logger,
		Ctx:                 ctx,
		Cancel:              cancel,
		IsActive:            false,
		TriggerCallback:     s.HandleTriggerNotification,
		CleanupCallback:     s.cleanupJobData,
	}
	return w, nil
}

// cleanupJobData removes job data from the scheduler's store when a worker stops
func (s *ConditionBasedScheduler) cleanupJobData(ctx context.Context, jobID string) error {
	s.notificationMutex.Lock()
	defer s.notificationMutex.Unlock()

	s.workersMutex.Lock()
	defer s.workersMutex.Unlock()

	// Remove job data from store
	delete(s.jobDataStore, jobID)
	// Clean up last trigger time tracking
	delete(s.lastTriggerTime, jobID)

	s.logger.Debug(ctx, "Cleaned up job data from store", observability.String("job_id", jobID))
	return nil
}

// GetJobData retrieves job data by job ID (thread-safe)
func (s *ConditionBasedScheduler) GetJobData(jobID string) (*types.ScheduleConditionJobData, error) {
	s.workersMutex.RLock()
	defer s.workersMutex.RUnlock()

	jobData, exists := s.jobDataStore[jobID]
	if !exists || jobData == nil {
		return nil, fmt.Errorf("job data not found for job %s", jobID)
	}

	return jobData, nil
}

// UnregisterEventJob unregisters an event job from Event Monitor Service
func (s *ConditionBasedScheduler) UnregisterEventJob(ctx context.Context, jobID string) error {
	s.workersMutex.Lock()
	defer s.workersMutex.Unlock()

	// Check if job exists in jobDataStore
	jobData, exists := s.jobDataStore[jobID]
	if !exists || jobData == nil {
		return fmt.Errorf("job %s is not found", jobID)
	}

	// Verify this is an event job (task definition ID 3, 4, or 8)
	if jobData.TaskDefinitionID != 3 && jobData.TaskDefinitionID != 4 && jobData.TaskDefinitionID != 8 {
		return fmt.Errorf("job %s is not an event job", jobID)
	}

	// Unregister from Event Monitor Service
	if s.eventMonitorClient != nil {
		if err := s.eventMonitorClient.Unregister(ctx, jobID); err != nil {
			return fmt.Errorf("failed to unregister from Event Monitor Service: %w", err)
		}
		s.logger.Debug(ctx, "Unregistered event job from Event Monitor Service", observability.String("job_id", jobID))
	}

	// Clean up job data
	delete(s.jobDataStore, jobID)
	delete(s.lastTriggerTime, jobID)

	return nil
}

// UnscheduleJob stops and removes a condition worker
func (s *ConditionBasedScheduler) UnscheduleJob(ctx context.Context, jobID string) error {
	// Create trace with format "condition-{scheduler_id}-{timestamp}"
	traceName := fmt.Sprintf("condition-%s-%d", s.schedulerID, time.Now().Unix())

	// Create root span for unscheduling operation
	ctx, unscheduleSpan := s.tracer.Start(ctx, traceName,
		observability.WithSpanKind(trace.SpanKindProducer),
		observability.WithAttributes(
			attribute.String("scheduler.id", s.schedulerID),
			attribute.String("scheduler.type", "condition"),
			attribute.String("job.id", jobID),
			attribute.String("trace.name", traceName),
		),
	)
	defer unscheduleSpan.End()

	unscheduleSpan.AddEvent("unschedule.started")

	s.notificationMutex.Lock()
	defer s.notificationMutex.Unlock()

	s.workersMutex.Lock()
	defer s.workersMutex.Unlock()

	// Get job data to determine job type
	jobData, exists := s.jobDataStore[jobID]
	if !exists || jobData == nil {
		metrics.TrackCriticalError("job_not_found")
		return fmt.Errorf("job %s is not scheduled", jobID)
	}

	// Determine job type and handle accordingly
	switch jobData.TaskDefinitionID {
	case 3, 4, 8:
		// Event-based job: unregister from Event Monitor Service
		if s.eventMonitorClient != nil {
			if err := s.eventMonitorClient.Unregister(ctx, jobID); err != nil {
				s.logger.Warn(ctx, "Failed to unregister from Event Monitor Service",
					observability.String("job_id", jobID),
					observability.Error(err))
				// Continue with cleanup even if unregister fails
			}
		}
		delete(s.jobDataStore, jobID)    // Clean up job data
		delete(s.lastTriggerTime, jobID) // Clean up last trigger time tracking
	case 5, 6, 9:
		// Condition-based job: stop the worker
		if conditionWorker, exists := s.conditionWorkers[jobID]; exists {
			// Handle websocket workers (stored as nil) - they stop via context cancellation
			// For regular condition workers, call Stop explicitly
			if conditionWorker != nil {
				conditionWorker.Stop(ctx)
			} else {
				// For websocket workers, we need to cancel their context
				// Since websocket workers use context derived from s.ctx, we can't cancel individually
				// The worker will stop when it checks ctx.Done() in its loop
				// We rely on the job data cleanup to prevent new triggers
				s.logger.Debug(ctx, "WebSocket worker will stop via context check", observability.String("job_id", jobID))
			}
			delete(s.conditionWorkers, jobID)
		}
		delete(s.jobDataStore, jobID)    // Clean up job data
		delete(s.lastTriggerTime, jobID) // Clean up last trigger time tracking
	default:
		metrics.TrackCriticalError("job_not_found")
		return fmt.Errorf("job %s has unsupported task definition id: %d", jobID, jobData.TaskDefinitionID)
	}

	// Update active workers count (only condition workers now)
	metrics.UpdateActiveWorkers(len(s.conditionWorkers))

	// Track job completion
	metrics.TrackJobCompleted("unscheduled")

	unscheduleSpan.AddEvent("unschedule.completed", observability.WithEventAttributes(
		attribute.String("job_id", jobID),
	))

	s.logger.Info(ctx, "Job unscheduled successfully", observability.String("job_id", jobID))
	return nil
}
