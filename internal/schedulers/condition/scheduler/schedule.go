package scheduler

import (
	"context"
	"fmt"
	"encoding/json"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/scheduler/worker"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/config"
	"github.com/trigg3rX/triggerx-backend/pkg/cryptography"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// ScheduleJob creates and starts a new condition worker
func (s *ConditionBasedScheduler) ScheduleJob(jobData *types.ScheduleConditionJobData) error {
	s.workersMutex.Lock()
	defer s.workersMutex.Unlock()

	startTime := time.Now()

	switch jobData.TaskDefinitionID {
	case 3, 4:
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
		// Create condition worker
		worker, err := s.createConditionWorker(&jobData.ConditionWorkerData, s.HTTPClient)
		if err != nil {
			metrics.TrackCriticalError("worker_creation_failed")
			return fmt.Errorf("failed to create condition worker: %w", err)
		}
		// Store worker
		s.conditionWorkers[jobData.JobID] = worker
		// Start worker
		go worker.Start()
		duration := time.Since(startTime)
		// Track condition by type and source
		metrics.TrackConditionByType(jobData.ConditionWorkerData.ConditionType)
		metrics.TrackConditionBySource(jobData.ConditionWorkerData.ValueSourceType)
		// metrics.TrackWorkerStartDuration(fmt.Sprintf("%d", jobData.JobID), duration)

		s.logger.Info("Condition job scheduled successfully",
			"job_id", jobData.JobID,
			"condition_type", jobData.ConditionWorkerData.ConditionType,
			"value_source", jobData.ConditionWorkerData.ValueSourceUrl,
			"upper_limit", jobData.ConditionWorkerData.UpperLimit,
			"lower_limit", jobData.ConditionWorkerData.LowerLimit,
			"active_workers", len(s.eventWorkers) + len(s.conditionWorkers),
			"max_workers", s.maxWorkers,
			"duration", duration,
		)
	case 5, 6:
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
		// Create event worker
		worker, err := s.createEventWorker(&jobData.EventWorkerData, s.chainClients[jobData.EventWorkerData.TriggerChainID])
		if err != nil {
			metrics.TrackCriticalError("worker_creation_failed")
			return fmt.Errorf("failed to create event worker: %w", err)
		}
		// Store worker
		s.eventWorkers[jobData.JobID] = worker
		// Start worker
		go worker.Start()
		// Track event by chain and contract
		duration := time.Since(startTime)
		// metrics.TrackEventByChain(jobData.EventWorkerData.TriggerChainID)
		// metrics.TrackWorkerStartDuration(fmt.Sprintf("%d", jobData.JobID), duration)

		s.logger.Info("Job scheduled successfully",
			"job_id", jobData.JobID,
			"trigger_chain", jobData.EventWorkerData.TriggerChainID,
			"contract", jobData.EventWorkerData.TriggerContractAddress,
			"event", jobData.EventWorkerData.TriggerEvent,
			"target_chain", jobData.TaskTargetData.TargetChainID,
			"target_contract", jobData.TaskTargetData.TargetContractAddress,
			"target_function", jobData.TaskTargetData.TargetFunction,
			"active_workers", len(s.eventWorkers) + len(s.conditionWorkers),
			"max_workers", s.maxWorkers,
			"duration", duration,
		)
		
	default:
		return fmt.Errorf("unsupported task definition id: %d", jobData.TaskDefinitionID)
	}

	// Update metrics
	metrics.TrackJobScheduled()
	metrics.UpdateActiveWorkers(len(s.eventWorkers) + len(s.conditionWorkers))
	metrics.TrackWorkerStart(fmt.Sprintf("%d", jobData.JobID))

	return nil
}

// createConditionWorker creates a new condition worker instance
func (s *ConditionBasedScheduler) createConditionWorker(conditionWorkerData *types.ConditionWorkerData, httpClient *retry.HTTPClient) (*worker.ConditionWorker, error) {
	ctx, cancel := context.WithCancel(s.ctx)

	worker := &worker.ConditionWorker{
		ConditionWorkerData: conditionWorkerData,
		Logger:          s.logger,
		HttpClient:      httpClient,
		Ctx:             ctx,
		Cancel:          cancel,
		IsActive:        false,
		LastCheckTimestamp: time.Now(),
		TriggerCallback: s.handleTriggerNotification,
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
		LastBlock: currentBlock,
		IsActive:        false,
		TriggerCallback: s.handleTriggerNotification,
	}

	return worker, nil
}

// handleTriggerNotification processes trigger notifications from workers and sends them to keepers
func (s *ConditionBasedScheduler) handleTriggerNotification(notification *worker.TriggerNotification) error {
	startTime := time.Now()

	s.logger.Info("Processing trigger notification from worker",
		"job_id", notification.JobID,
		"trigger_value", notification.TriggerValue,
		"trigger_tx_hash", notification.TriggerTxHash,
		"triggered_at", notification.TriggeredAt,
	)

	// Get the worker to access job data
	s.workersMutex.RLock()
	worker, exists := s.conditionWorkers[notification.JobID]
	s.workersMutex.RUnlock()

	if !exists {
		return fmt.Errorf("worker not found for job %d", notification.JobID)
	}

	// Send task to keeper for execution
	success, err := s.SendTaskToPerformer(worker.ConditionWorkerData, notification)
	if err != nil {
		s.logger.Error("Failed to send condition task to keeper",
			"job_id", notification.JobID,
			"error", err,
		)
		metrics.TrackCriticalError("keeper_task_send_failed")
		return err
	}

	duration := time.Since(startTime)
	if success {
		s.logger.Info("Successfully sent condition task to keeper",
			"job_id", notification.JobID,
			"duration", duration,
		)
		metrics.TrackActionExecution(fmt.Sprintf("%d", notification.JobID), duration)
	} else {
		s.logger.Error("Failed to send condition task to keeper",
			"job_id", notification.JobID,
			"duration", duration,
		)
		metrics.TrackCriticalError("keeper_task_execution_failed")
	}

	return nil
}

// sendConditionTaskToKeeper sends the triggered condition task to a keeper for execution
func (s *ConditionBasedScheduler) SendTaskToPerformer(jobData *types.ScheduleConditionJobData, notification *types.TriggerNotification) (bool, error) {
	// Get the performer data
	// TODO: Get the performer data from redis service, which gets it from online keepers list from health service, and sets the performerLock in redis
	// For now, I fixed the performer
	performerData := types.PerformerData{
		KeeperID:      3,
		KeeperAddress: "0x0a067a261c5f5e8c4c0b9137430b4fe1255eb62e",
	}

	createTaskRequest := types.CreateTaskRequest{
		JobID: jobData.JobID,
		TaskDefinitionID: jobData.TaskDefinitionID,
		TaskPerformerID: performerData.KeeperID,
	}

	// Create the task data
	response, err := s.dbClient.CreateTask(createTaskRequest)
	if err != nil {
		return false, fmt.Errorf("failed to create task: %w", err)
	}

	// Generate the task data to send to the performer
	targetData := []types.TaskTargetData{
		{
			JobID:                     jobData.JobID,
			TaskID:                    jobData.JobID, // Using jobID as taskID for now
			TaskDefinitionID:          jobData.TaskDefinitionID,
			TargetChainID:             jobData.TaskTargetData.TargetChainID,
			TargetContractAddress:     jobData.TaskTargetData.TargetContractAddress,
			TargetFunction:            jobData.TaskTargetData.TargetFunction,
			ABI:                       jobData.TaskTargetData.ABI,
			ArgType:                   jobData.TaskTargetData.ArgType,
			Arguments:                 jobData.TaskTargetData.Arguments,
			DynamicArgumentsScriptUrl: jobData.TaskTargetData.DynamicArgumentsScriptUrl,
		},
	}

	triggerData := []types.TaskTriggerData{
		{
			TaskID:                  jobData.JobID,
			TaskDefinitionID:        jobData.TaskDefinitionID,
			Recurring:               false,
			ExpirationTime:          time.Now().Add(time.Hour * 24 * 30),
			CurrentTriggerTimestamp: notification.TriggeredAt,
			ConditionType:           worker.ConditionWorkerData.ConditionType,
			ConditionSourceType:     worker.ConditionWorkerData.ValueSourceType,
			ConditionSourceUrl:      worker.ConditionWorkerData.ValueSourceUrl,
			ConditionUpperLimit:     int(worker.ConditionWorkerData.UpperLimit),
			ConditionLowerLimit:     int(worker.ConditionWorkerData.LowerLimit),
			ConditionSatisfiedValue: int(notification.TriggerValue),
		},
	}

	// Create scheduler signature data
	schedulerSignatureData := types.SchedulerSignatureData{
		TaskID:                  jobData.JobID,
		SchedulerSigningAddress: s.schedulerSigningAddress,
	}

	// Create the batch task data
	sendTaskData := types.SendTaskDataToKeeper{
		TaskID:             jobData.JobID,
		PerformerData:      performerData,
		TargetData:         targetData,
		TriggerData:        triggerData,
		SchedulerSignature: &schedulerSignatureData,
	}

	// Sign the task data
	signature, err := cryptography.SignJSONMessage(sendTaskData, config.GetSchedulerSigningKey())
	if err != nil {
		s.logger.Errorf("Failed to sign batch task data: %v", err)
		return false, fmt.Errorf("failed to sign batch task data: %w", err)
	}
	sendTaskData.SchedulerSignature.SchedulerSignature = signature

	jsonData, err := json.Marshal(sendTaskData)
	if err != nil {
		s.logger.Errorf("Failed to marshal batch task data: %v", err)
		return false, fmt.Errorf("failed to marshal batch task data: %w", err)
	}
	dataBytes := []byte(jsonData)

	broadcastDataForPerformer := types.BroadcastDataForPerformer{
		TaskID:           sendTaskData.TaskID,
		TaskDefinitionID: sendTaskData.TargetData[0].TaskDefinitionID, // Use first task's definition ID
		PerformerAddress: sendTaskData.PerformerData.KeeperAddress,
		Data:             dataBytes,
	}

	success, err := s.aggClient.SendTaskToPerformer(s.ctx, &broadcastDataForPerformer)
	if err != nil {
		s.logger.Errorf("Failed to send task to performer: %v", err)
		return false, fmt.Errorf("failed to send task to performer: %w", err)
	}

	// Execute the task
	return success, nil
}

// UnscheduleJob stops and removes a condition worker
func (s *ConditionBasedScheduler) UnscheduleJob(jobID int64) error {
	s.workersMutex.Lock()
	defer s.workersMutex.Unlock()

	worker, exists := s.conditionWorkers[jobID]
	if !exists {
		metrics.TrackCriticalError("job_not_found")
		return fmt.Errorf("job %d is not scheduled", jobID)
	}

	// Stop worker
	worker.Stop()

	// Remove from workers map
	delete(s.conditionWorkers, jobID)

	// Update active workers count
	totalWorkers := len(s.conditionWorkers) + len(s.eventWorkers)
	metrics.UpdateActiveWorkers(totalWorkers)

	// Track job completion
	metrics.TrackJobCompleted("unscheduled")

	s.logger.Info("Condition job unscheduled successfully", "job_id", jobID)
	return nil
}
