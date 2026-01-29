package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// TaskRepository defines the interface for task data operations
type TaskRepository interface {
	UpdateTaskExecutionData(ctx context.Context, req *types.ReportTaskExecutionStatusRequest, ipfsData *types.IPFSData) error
	UpdateTaskSubmissionData(ctx context.Context, taskData *types.ReportTaskConsensusStatusRequest) error
	GetUserEmailByTaskID(ctx context.Context, taskID int64) (string, error)
	UpdateTaskAttestationTimeoutFailure(ctx context.Context, taskID int64, errorMsg string) error
}

type taskRepository struct {
	db     *database.Connection
	logger observability.Logger
}

// NewTaskRepository creates a new task repository instance
func NewTaskRepository(db *database.Connection, logger observability.Logger) TaskRepository {
	return &taskRepository{
		db:     db,
		logger: logger,
	}
}

// UpdateTaskExecutionData updates task execution data in database
func (r *taskRepository) UpdateTaskExecutionData(ctx context.Context, req *types.ReportTaskExecutionStatusRequest, ipfsData *types.IPFSData) error {
	// Add the req data to task_data table
	var taskStatus string
	if req.ExecutionSuccessful {
		taskStatus = string(types.TaskStatusExecuted)
		req.Error = fmt.Sprintf("execution successful: %s", req.Error)
	} else {
		taskStatus = string(types.TaskStatusFailed)
		req.Error = fmt.Sprintf("execution failed: %s", req.Error)
	}
	var convertedArguments []string
	if ipfsData != nil && ipfsData.ActionData != nil {
		convertedArguments = make([]string, 0, len(ipfsData.ActionData.ConvertedArguments))
		for _, arg := range ipfsData.ActionData.ConvertedArguments {
			convertedArguments = append(convertedArguments, fmt.Sprintf("%v", arg))
		}
	}

	// Update the task_data table
	if err := r.db.NewQuery(UpdateTaskExecutionData,
		[]string{req.KeeperAddress}, req.IPFSDataCID, req.ExecutionSuccessful, req.Error, taskStatus, 
		req.ExecutionTxHash, req.ExecutedAt, req.TaskOpxActualCost, convertedArguments, req.TaskID).Exec(); err != nil {
		r.logger.Error(ctx, "Error updating task execution details for task ID", observability.Int64("task_id", req.TaskID), observability.Error(err))
		return err
	}

	// If there's no IPFS data (task failed before IPFS upload), skip IPFS-dependent operations
	if ipfsData == nil {
		r.logger.Debug(ctx, "Skipping IPFS-dependent updates - no IPFS data available", 
			observability.Int64("task_id", req.TaskID),
			observability.String("ipfs_cid", req.IPFSDataCID))
	} else {
		// Update the specific job_data table
		switch ipfsData.TaskData.TriggerData.TaskDefinitionID {
			case 1, 2, 7: // Time-based jobs
				if err := r.db.NewQuery(UpdateTimeJobLastExecutedAt, ipfsData.TaskData.TriggerData.NextTriggerTimestamp, ipfsData.TaskData.TargetData.JobID).Exec(); err != nil {
					r.logger.Error(ctx, "Failed to update time job last_executed_at", observability.String("job_id", ipfsData.TaskData.TargetData.JobID), observability.Error(err))
					return fmt.Errorf("failed to update time job last_executed_at: %w", err)
				}
			case 3, 4, 8: // Event-based jobs
				if err := r.db.NewQuery(UpdateEventJobLastExecutedAt, ipfsData.TaskData.TriggerData.NextTriggerTimestamp, ipfsData.TaskData.TargetData.JobID).Exec(); err != nil {
					r.logger.Error(ctx, "Failed to update event job last_executed_at", observability.String("job_id", ipfsData.TaskData.TargetData.JobID), observability.Error(err))
					return fmt.Errorf("failed to update event job last_executed_at: %w", err)
				}
			case 5, 6, 9: // Condition-based jobs
				if err := r.db.NewQuery(UpdateConditionJobLastExecutedAt, ipfsData.TaskData.TriggerData.NextTriggerTimestamp, ipfsData.TaskData.TargetData.JobID).Exec(); err != nil {
					r.logger.Error(ctx, "Failed to update condition job last_executed_at", observability.String("job_id", ipfsData.TaskData.TargetData.JobID), observability.Error(err))
					return fmt.Errorf("failed to update condition job last_executed_at: %w", err)
				}
			default:
				r.logger.Warn(ctx, "Unknown task definition ID, skipping last_executed_at update", observability.Int("task_definition_id", ipfsData.TaskData.TriggerData.TaskDefinitionID), observability.String("job_id", ipfsData.TaskData.TargetData.JobID))
		}

		// For agent script jobs (TaskDefinitionID = 7/8/9), update storage
		if ipfsData.TaskData.TriggerData.TaskDefinitionID == 7 && ipfsData.ActionData != nil && ipfsData.ActionData.StorageUpdates != nil && len(ipfsData.ActionData.StorageUpdates) > 0 {
			// Update each storage key-value pair
			for key, value := range ipfsData.ActionData.StorageUpdates {
				if err := r.db.NewQuery(UpdateScriptStorageQuery,
					value, time.Now().UTC(), ipfsData.TaskData.TargetData.JobID, key).Exec(); err != nil {
					r.logger.Error(ctx, "Failed to update storage key for job", observability.String("key", key), observability.String("job_id", ipfsData.TaskData.TargetData.JobID), observability.Error(err))
					return fmt.Errorf("failed to update storage: %w", err)
				}
				r.logger.Debug(ctx, "Updated storage", observability.String("job_id", ipfsData.TaskData.TargetData.JobID), observability.String("key", key))
			}
			r.logger.Debug(ctx, "Successfully updated storage keys for job", observability.Int("storage_count", len(ipfsData.ActionData.StorageUpdates)), observability.String("job_id", ipfsData.TaskData.TargetData.JobID))
		}

		// Update the job_cost_actual in job_data table
		var existingJobCostActual string
		err := r.db.Session().Query(GetJobCostActual, ipfsData.TaskData.TargetData.JobID).Scan(&existingJobCostActual)
		if err != nil {
			r.logger.Error(ctx, "Failed to get job cost actual for job ID", observability.String("job_id", ipfsData.TaskData.TargetData.JobID), observability.Error(err))
			return fmt.Errorf("failed to get job cost actual for job ID %s: %w", ipfsData.TaskData.TargetData.JobID, err)
		}

		jobCostActual := types.Add(existingJobCostActual, req.TaskOpxActualCost)
		if err = r.db.Session().Query(UpdateJobCostActual, jobCostActual, ipfsData.TaskData.TargetData.JobID).Exec(); err != nil {
			r.logger.Error(ctx, "Failed to update job cost actual for job ID", observability.String("job_id", ipfsData.TaskData.TargetData.JobID), observability.Error(err))
			return fmt.Errorf("failed to update job cost actual for job ID %s: %w", ipfsData.TaskData.TargetData.JobID, err)
		}

		// Update the user_points in user_data table
		var userAddress string
		err = r.db.Session().Query(GetUserAddressByJobId, ipfsData.TaskData.TargetData.JobID).Scan(&userAddress)
		if err != nil {
			r.logger.Error(ctx, "Failed to get user address for job ID", observability.String("job_id", ipfsData.TaskData.TargetData.JobID), observability.Error(err))
			return fmt.Errorf("failed to get user address for job ID %s: %w", ipfsData.TaskData.TargetData.JobID, err)
		}
		var existingUserPoints string
		err = r.db.Session().Query(GetUserPoints, userAddress).Scan(&existingUserPoints)
		if err != nil {
			r.logger.Error(ctx, "Failed to get user points for user address", observability.String("user_address", userAddress), observability.Error(err))
			return fmt.Errorf("failed to get user points for user address %s: %w", userAddress, err)
		}
		userPoints := types.Add(existingUserPoints, req.TaskOpxActualCost)
		if err = r.db.Session().Query(UpdateUserPoints, userPoints, time.Now().UTC(), userAddress).Exec(); err != nil {
			r.logger.Error(ctx, "Failed to update user points for user address", observability.String("user_address", userAddress), observability.Error(err))
			return fmt.Errorf("failed to update user points for user address %s: %w", userAddress, err)
		}
	}

	// Update the no_executed_tasks and keeper_points in keeper_data table
	var keeperPoints string
	var noExecutedTasks, noAttestedTasks int64
	var err error
	err = r.db.Session().Query(GetKeeperPointsAndNoOfTasks, req.KeeperAddress).Scan(&keeperPoints, &noExecutedTasks, &noAttestedTasks)
	if err != nil {
		r.logger.Error(ctx, "Failed to get keeper points and no of executed tasks for keeper address", observability.String("keeper_address", req.KeeperAddress), observability.Error(err))
		return fmt.Errorf("failed to get keeper points and no of executed tasks for keeper address %s: %w", req.KeeperAddress, err)
	}
	keeperPoints = types.Add(keeperPoints, req.TaskOpxActualCost)
	noExecutedTasks = noExecutedTasks + 1
	if err = r.db.Session().Query(UpdateKeeperPointsAndNoOfTasks, keeperPoints, noExecutedTasks, noAttestedTasks, req.KeeperAddress).Exec(); err != nil {
		r.logger.Error(ctx, "Failed to update keeper points and no of executed tasks for keeper address", observability.String("keeper_address", req.KeeperAddress), observability.Error(err))
		return fmt.Errorf("failed to update keeper points and no of executed tasks for keeper address %s: %w", req.KeeperAddress, err)
	}

	return nil
}

// UpdateTaskSubmissionData updates task number, success status and execution details in database
func (r *taskRepository) UpdateTaskSubmissionData(ctx context.Context, req *types.ReportTaskConsensusStatusRequest) error {
	// Convert attester operator_ids to keeper_addresses using network-aware lookup
	attesterAddresses := make([]string, 0, len(req.AttesterIds))
	for _, attesterID := range req.AttesterIds {
		var attesterAddress string
		err := r.db.Session().Query(GetKeeperAddressByOperatorIDAndNetwork, attesterID, req.Network).Scan(&attesterAddress)
		if err != nil {
			r.logger.Error(ctx, "Failed to get keeper address for attester ID and network", observability.Int64("attester_id", attesterID), observability.String("network", req.Network), observability.Error(err))
			return fmt.Errorf("failed to get keeper address for attester ID %d and network %s: %w", attesterID, req.Network, err)
		}
		attesterAddresses = append(attesterAddresses, attesterAddress)
	}

	// Update the task_data table
	if err := r.db.Session().Query(UpdateTaskSubmissionData,
		req.TaskNumber, req.IsAccepted, req.TaskSubmissionTxHash, attesterAddresses, time.Now().UTC(), req.TaskID).Exec(); err != nil {
		r.logger.Error(ctx, "Error updating task execution details for task ID", observability.Int64("task_id", req.TaskID), observability.Error(err))
		return err
	}

	// Update the no_attested_tasks and keeper_points in keeper_data table
	for _, attesterAddress := range attesterAddresses {
		var keeperPoints string
		var noExecutedTasks, noAttestedTasks int64
		err := r.db.Session().Query(GetKeeperPointsAndNoOfTasks, attesterAddress).Scan(&keeperPoints, &noExecutedTasks, &noAttestedTasks)
		if err != nil {
			r.logger.Error(ctx, "Failed to get keeper points and no of attested tasks for user address", observability.String("user_address", attesterAddress), observability.Error(err))
			return fmt.Errorf("failed to get keeper points and no of attested tasks for user address %s: %w", attesterAddress, err)
		}
		keeperPoints = types.Add(keeperPoints, req.TaskOpxActualCost)
		noAttestedTasks = noAttestedTasks + 1
		if err = r.db.Session().Query(UpdateKeeperPointsAndNoOfTasks, keeperPoints, noExecutedTasks, noAttestedTasks, attesterAddress).Exec(); err != nil {
			r.logger.Error(ctx, "Failed to update keeper points and no of attested tasks for user address", observability.String("user_address", attesterAddress), observability.Error(err))
			return fmt.Errorf("failed to update keeper points and no of attested tasks for user address %s: %w", attesterAddress, err)
		}
	}

	r.logger.Info(ctx, "Successfully updated task with submission details", observability.Int64("task_id", req.TaskID))
	return nil
}

// GetUserEmailByTaskID returns the user's email_id for a given task_id
func (r *taskRepository) GetUserEmailByTaskID(ctx context.Context, taskID int64) (string, error) {
	var jobID string
	err := r.db.Session().Query(GetJobIDByTaskID, taskID).Scan(&jobID)
	if err != nil {
		r.logger.Error(ctx, "Failed to get job ID for task ID", observability.Int64("task_id", taskID), observability.Error(err))
		return "", fmt.Errorf("failed to get job ID for task ID %d: %w", taskID, err)
	}

	var userAddress string
	err = r.db.Session().Query(GetUserAddressByJobId, jobID).Scan(&userAddress)
	if err != nil {
		r.logger.Error(ctx, "Failed to get user address for job ID", observability.String("job_id", jobID), observability.Error(err))
		return "", fmt.Errorf("failed to get user address for job ID %s: %w", jobID, err)
	}

	var email string
	err = r.db.Session().Query(GetUserEmailByUserAddress, userAddress).Scan(&email)
	if err != nil {
		r.logger.Error(ctx, "Failed to get email for user address", observability.String("user_address", userAddress), observability.Error(err))
		return "", fmt.Errorf("failed to get email for user address %s: %w", userAddress, err)
	}

	return email, nil
}

// UpdateTaskAttestationTimeoutFailure updates task attestation timeout failure in database
func (r *taskRepository) UpdateTaskAttestationTimeoutFailure(ctx context.Context, taskID int64, errorMsg string) error {
	if err := r.db.Session().Query(UpdateTaskAttestationTimeoutFailure, taskID, errorMsg).Exec(); err != nil {
		r.logger.Error(ctx, "Error updating task attestation timeout failure for task ID", observability.Int64("task_id", taskID), observability.Error(err))
		return err
	}
	return nil
}
