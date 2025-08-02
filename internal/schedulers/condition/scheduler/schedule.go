package scheduler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/scheduler/worker"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
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

// scheduleEventJob handles event-based job scheduling
func (s *ConditionBasedScheduler) scheduleEventJob(jobData *types.ScheduleConditionJobData, startTime time.Time) error {
	// Check if job is already scheduled
	if _, exists := s.eventWorkers[jobData.JobID]; exists {
		metrics.TrackCriticalError("duplicate_job_schedule")
		return fmt.Errorf("job %d is already scheduled", jobData.JobID)
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
	s.jobDataStore[jobData.JobID.String()] = jobData

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
func (s *ConditionBasedScheduler) createConditionWorker(conditionWorkerData *types.ConditionWorkerData, httpClient *retry.HTTPClient) (*worker.ConditionWorker, error) {
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

// handleTriggerNotification processes trigger notifications and submits individual tasks to Redis API
func (s *ConditionBasedScheduler) handleTriggerNotification(notification *worker.TriggerNotification) error {
	startTime := time.Now()

	s.logger.Info("Processing trigger notification - submitting task to Redis API",
		"job_id", notification.JobID,
		"trigger_value", notification.TriggerValue,
		"trigger_tx_hash", notification.TriggerTxHash,
		"triggered_at", notification.TriggeredAt,
	)

	// Acquire notification mutex to prevent cleanup during processing
	s.notificationMutex.Lock()
	defer s.notificationMutex.Unlock()

	// Get the job data from storage
	s.workersMutex.RLock()
	jobData, exists := s.jobDataStore[notification.JobID.String()]
	s.workersMutex.RUnlock()

	if !exists || jobData == nil {
		s.logger.Error("Job data not found", "job_id", notification.JobID)
		return fmt.Errorf("job data not found for job %d", notification.JobID)
	}

	createTaskRequest := types.CreateTaskRequest{
		JobID:            jobData.JobID,
		TaskDefinitionID: jobData.TaskDefinitionID,
	}

	// Create Task in Database
	taskID, err := s.dbClient.CreateTask(createTaskRequest)
	if err != nil {
		s.logger.Error("Failed to create task in database", "job_id", notification.JobID, "error", err)
		return fmt.Errorf("failed to create task in database: %w", err)
	}
	jobData.TaskTargetData.TaskID = taskID

	// Create individual task and submit to Redis API
	success, err := s.submitTriggeredTaskToTaskManager(jobData, notification)
	if err != nil {
		s.logger.Error("Failed to submit triggered task to Redis API",
			"job_id", notification.JobID,
			"error", err,
		)
		metrics.TrackCriticalError("redis_task_submission_failed")
		return err
	}

	duration := time.Since(startTime)
	if success {
		s.logger.Info("Successfully submitted triggered task to Redis API",
			"job_id", notification.JobID,
			"duration", duration,
		)
		metrics.TrackActionExecution(fmt.Sprintf("%d", notification.JobID), duration)
	} else {
		s.logger.Error("Failed to submit triggered task to Redis API",
			"job_id", notification.JobID,
			"duration", duration,
		)
		metrics.TrackCriticalError("redis_task_submission_failed")
	}

	return nil
}

// submitTriggeredTaskToTaskManager creates and submits a single task to TaskManager when triggers occur
func (s *ConditionBasedScheduler) submitTriggeredTaskToTaskManager(jobData *types.ScheduleConditionJobData, notification *worker.TriggerNotification) (bool, error) {
	s.logger.Info("Creating triggered task for Redis API submission",
		"job_id", jobData.JobID,
		"task_definition_id", jobData.TaskDefinitionID,
		"trigger_value", notification.TriggerValue)

	// Create single task data (not batch like time scheduler)
	targetData := types.TaskTargetData{
		JobID:                     jobData.JobID,
		TaskID:                    jobData.TaskTargetData.TaskID,
		TaskDefinitionID:          jobData.TaskDefinitionID,
		TargetChainID:             jobData.TaskTargetData.TargetChainID,
		TargetContractAddress:     jobData.TaskTargetData.TargetContractAddress,
		TargetFunction:            jobData.TaskTargetData.TargetFunction,
		ABI:                       jobData.TaskTargetData.ABI,
		ArgType:                   jobData.TaskTargetData.ArgType,
		Arguments:                 jobData.TaskTargetData.Arguments,
		DynamicArgumentsScriptUrl: jobData.TaskTargetData.DynamicArgumentsScriptUrl,
		IsImua:                    jobData.IsImua,
	}

	// Create trigger data based on job type
	triggerData := s.createTriggerDataFromNotification(jobData, notification)

	// Create single task data for keeper
	sendTaskData := types.SendTaskDataToKeeper{
		TaskID:           []int64{jobData.TaskTargetData.TaskID},
		TargetData:       []types.TaskTargetData{targetData}, // Single task, not batch
		TriggerData:      []types.TaskTriggerData{triggerData},
		SchedulerID:      s.schedulerID,
		ManagerSignature: "",
	}

	// Create request for Redis API
	request := types.SchedulerTaskRequest{
		SendTaskDataToKeeper: sendTaskData,
		Source:               "condition_scheduler",
	}

	// Submit to TaskManager
	return s.submitTaskToTaskManager(request, notification.JobID)
}

// createTriggerDataFromNotification creates appropriate trigger data based on job type
func (s *ConditionBasedScheduler) createTriggerDataFromNotification(jobData *types.ScheduleConditionJobData, notification *worker.TriggerNotification) types.TaskTriggerData {
	baseTriggerData := types.TaskTriggerData{
		TaskID:                  jobData.TaskTargetData.TaskID,
		TaskDefinitionID:        jobData.TaskDefinitionID,
		CurrentTriggerTimestamp: notification.TriggeredAt,
	}

	switch jobData.TaskDefinitionID {
	case 5, 6: // Condition-based
		baseTriggerData.ExpirationTime = jobData.ConditionWorkerData.ExpirationTime
		baseTriggerData.ConditionSatisfiedValue = int(notification.TriggerValue)
		baseTriggerData.ConditionType = jobData.ConditionWorkerData.ConditionType
		baseTriggerData.ConditionSourceType = jobData.ConditionWorkerData.ValueSourceType
		baseTriggerData.ConditionSourceUrl = jobData.ConditionWorkerData.ValueSourceUrl
		baseTriggerData.ConditionUpperLimit = int(jobData.ConditionWorkerData.UpperLimit)
		baseTriggerData.ConditionLowerLimit = int(jobData.ConditionWorkerData.LowerLimit)
		s.logger.Info("Condition job expiration time", "expiration_time", jobData.ConditionWorkerData.ExpirationTime)

	case 3, 4: // Event-based
		baseTriggerData.ExpirationTime = jobData.EventWorkerData.ExpirationTime
		baseTriggerData.EventTxHash = notification.TriggerTxHash
		baseTriggerData.EventChainId = jobData.EventWorkerData.TriggerChainID
		baseTriggerData.EventTriggerContractAddress = jobData.EventWorkerData.TriggerContractAddress
		baseTriggerData.EventTriggerName = jobData.EventWorkerData.TriggerEvent
		s.logger.Info("Event job expiration time", "expiration_time", jobData.EventWorkerData.ExpirationTime)
	}

	return baseTriggerData
}

// submitTaskToTaskManager submits the task to TaskManager via HTTP
func (s *ConditionBasedScheduler) submitTaskToTaskManager(request types.SchedulerTaskRequest, taskID *big.Int) (bool, error) {
	startTime := time.Now()

	// Marshal request to JSON
	requestBytes, err := json.Marshal(request)
	if err != nil {
		s.logger.Error("Failed to marshal task request",
			"task_id", taskID,
			"error", err)
		return false, err
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/scheduler/submit-task", s.redisAPIURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBytes))
	if err != nil {
		s.logger.Error("Failed to create HTTP request",
			"task_id", taskID,
			"url", url,
			"error", err)
		return false, err
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error("Failed to send task to Redis API",
			"task_id", taskID,
			"url", url,
			"error", err)
		return false, err
	}
	defer func() { _ = resp.Body.Close() }()

	// Parse response
	var apiResponse types.TaskManagerAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		s.logger.Error("Failed to decode TaskManager response",
			"task_id", taskID,
			"status_code", resp.StatusCode,
			"error", err)
		return false, err
	}

	duration := time.Since(startTime)

	if resp.StatusCode != http.StatusOK {
		s.logger.Error("TaskManager returned error",
			"task_id", taskID,
			"status_code", resp.StatusCode,
			"error", apiResponse.Error,
			"details", apiResponse.Details,
			"duration", duration)
		return false, fmt.Errorf("taskManager error: %s", apiResponse.Error)
	}

	if !apiResponse.Success {
		s.logger.Error("TaskManager processing failed",
			"task_id", taskID,
			"message", apiResponse.Message,
			"error", apiResponse.Error,
			"duration", duration)
		return false, fmt.Errorf("taskManager processing failed: %s", apiResponse.Error)
	}

	s.logger.Info("Successfully submitted task to TaskManager",
		"task_id", taskID,
		"response_task_ids", apiResponse.TaskID,
		"duration", duration,
		"message", apiResponse.Message)

	return true, nil
}

// UnscheduleJob stops and removes a condition worker
func (s *ConditionBasedScheduler) UnscheduleJob(jobID *big.Int) error {
	s.notificationMutex.Lock()
	defer s.notificationMutex.Unlock()

	s.workersMutex.Lock()
	defer s.workersMutex.Unlock()

	// Try condition workers first
	if conditionWorker, exists := s.conditionWorkers[jobID]; exists {
		conditionWorker.Stop()
		delete(s.conditionWorkers, jobID)
		delete(s.jobDataStore, jobID.String()) // Clean up job data
	} else if eventWorker, exists := s.eventWorkers[jobID]; exists {
		eventWorker.Stop()
		delete(s.eventWorkers, jobID)
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
