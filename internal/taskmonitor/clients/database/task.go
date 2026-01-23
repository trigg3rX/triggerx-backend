package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/clients/database/queries"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// UpdateTaskSubmissionData updates task number, success status and execution details in database
func (dm *DatabaseClient) UpdateTaskSubmissionData(ctx context.Context, data types.TaskSubmissionData) error {
	// Get performer address (PerformerAddress is a string containing consensus_address)
	if data.PerformerAddress == "" {
		dm.logger.Error(ctx, "Performer address is empty in task submission data",
			observability.Int64("task_id", data.TaskID))
		return fmt.Errorf("performer address is empty for task %d", data.TaskID)
	}

	// PerformerAddress from IPFS contains consensus_address (as stored by keeper)
	performerAddress, err := dm.GetKeeperAddressByConsensusAddress(ctx, data.PerformerAddress)
	if err != nil {
		dm.logger.Error(ctx, "Failed to get performer address by consensus address",
			observability.String("consensus_address", data.PerformerAddress),
			observability.Int64("task_id", data.TaskID),
			observability.Error(err))
		return fmt.Errorf("keeper not found for consensus address %s: %w", data.PerformerAddress, err)
	}

	// Convert attester operator_ids to keeper_addresses
	attesterAddresses := make([]string, 0, len(data.AttesterIds))
	for _, operatorID := range data.AttesterIds {
		attesterAddress, err := dm.GetKeeperAddressByOperatorID(ctx, operatorID)
		if err != nil {
			dm.logger.Warn(ctx, "Failed to get keeper address for operator_id, skipping",
				observability.Int64("operator_id", operatorID),
				observability.Int64("task_id", data.TaskID),
				observability.Error(err))
			continue
		}
		attesterAddresses = append(attesterAddresses, attesterAddress)
	}

	// Convert []interface{} to []string for Cassandra list<text>
	// Ensure we always pass a []string (empty slice if no arguments) for Cassandra list<text>
	convertedArgsStrings := make([]string, 0, len(data.ConvertedArguments))
	for _, arg := range data.ConvertedArguments {
		convertedArgsStrings = append(convertedArgsStrings, fmt.Sprintf("%v", arg))
	}

	if err := dm.db.NewQuery(queries.UpdateTaskSubmissionData,
		data.TaskNumber,
		data.IsAccepted,
		data.TaskSubmissionTxHash,
		[]string{performerAddress}, // task_performer_address is list<text>
		attesterAddresses,          // task_attester_address is list<text>
		data.ExecutionTxHash,
		data.ExecutionTimestamp,
		data.TaskOpxCost, // Already a string (Wei)
		data.ProofOfTask,
		convertedArgsStrings,
		data.TaskID).Exec(); err != nil {
		dm.logger.Error(ctx, "Error updating task execution details for task ID", observability.Int64("task_id", data.TaskID), observability.Error(err))
		return err
	}

	dm.logger.Info(ctx, "Successfully updated task with submission details", observability.Int64("task_id", data.TaskID))
	return nil
}

func (dm *DatabaseClient) UpdateTaskFailed(ctx context.Context, taskID int64) error {
	// Check if the task already has a final status (completed or failed)
	var existingStatus string
	iter := dm.db.NewQuery(queries.GetTaskStatusByID, taskID).Iter()
	defer func() {
		if cerr := iter.Close(); cerr != nil {
			dm.logger.Error(ctx, "Error closing iterator", observability.Error(cerr))
		}
	}()
	if iter.Scan(&existingStatus) && (existingStatus == "completed" || existingStatus == "failed") {
		dm.logger.Info(ctx, "Task already has final status, not updating to failed.", observability.Int64("task_id", taskID), observability.String("status", existingStatus))
		return nil
	}
	// Allow updating from pending_confirmation to failed (timeout case)
	if err := dm.db.NewQuery(queries.UpdateTaskFailed, taskID).Exec(); err != nil {
		dm.logger.Error(ctx, "Error updating task failed for task ID", observability.Int64("task_id", taskID), observability.Error(err))
		return err
	}
	dm.logger.Info(ctx, "Successfully updated task as failed", observability.Int64("task_id", taskID))
	return nil
}

// UpdateTaskError updates a task with error information
func (dm *DatabaseClient) UpdateTaskError(ctx context.Context, taskID int64, errorMsg string) error {
	// Check if the task already has a final status (completed or failed)
	var existingStatus string
	iter := dm.db.NewQuery(queries.GetTaskStatusByID, taskID).Iter()
	defer func() {
		if cerr := iter.Close(); cerr != nil {
			dm.logger.Error(ctx, "Error closing iterator", observability.Error(cerr))
		}
	}()
	if iter.Scan(&existingStatus) && (existingStatus == "completed" || existingStatus == "failed") {
		dm.logger.Info(ctx, "Task already has final status, not updating to failed.", observability.Int64("task_id", taskID), observability.String("status", existingStatus))
		return nil
	}
	// Allow updating from pending_confirmation to failed (timeout case)
	if err := dm.db.NewQuery(queries.UpdateTaskError, errorMsg, taskID).Exec(); err != nil {
		dm.logger.Error(ctx, "Error updating task error for task ID", observability.Int64("task_id", taskID), observability.Error(err))
		return err
	}
	dm.logger.Info(ctx, "Successfully updated task with error", observability.Int64("task_id", taskID), observability.String("error_msg", errorMsg))
	return nil
}

// UpdateTaskAggregatorFailed updates task when it failed (execution or aggregator submission)
// executionTxHash is the transaction hash from on-chain execution (may be empty if tx was never sent)
// proofCID contains all execution data if available
func (dm *DatabaseClient) UpdateTaskAggregatorFailed(ctx context.Context, taskID int64, errorMsg, executionTxHash, proofCID string) error {
	// Check if task already has a final status
	var existingStatus string
	iter := dm.db.NewQuery(queries.GetTaskStatusByID, taskID).Iter()
	defer func() {
		if cerr := iter.Close(); cerr != nil {
			dm.logger.Error(ctx, "Error closing iterator", observability.Error(cerr))
		}
	}()
	if iter.Scan(&existingStatus) && (existingStatus == "completed" || existingStatus == "failed") {
		dm.logger.Info(ctx, "Task already has final status, not updating to failed.", observability.Int64("task_id", taskID), observability.String("status", existingStatus))
		return nil
	}

	if err := dm.db.NewQuery(queries.UpdateTaskAggregatorFailed,
		errorMsg, executionTxHash, proofCID, taskID).Exec(); err != nil {
		dm.logger.Error(ctx, "Error updating task failed for task ID", observability.Int64("task_id", taskID), observability.Error(err))
		return err
	}
	dm.logger.Info(ctx, "Successfully updated task as failed", observability.Int64("task_id", taskID), observability.String("error_msg", errorMsg), observability.String("execution_tx_hash", executionTxHash))
	return nil
}

// UpdateTaskAggregatorSubmitted updates task when both execution and aggregator submission succeeded
// The task is now pending on-chain confirmation
// executionTxHash is the transaction hash from on-chain execution
// proofCID contains all execution data
func (dm *DatabaseClient) UpdateTaskAggregatorSubmitted(ctx context.Context, taskID int64, executionTxHash, proofCID string) error {
	// Check if task already has a final status
	var existingStatus string
	iter := dm.db.NewQuery(queries.GetTaskStatusByID, taskID).Iter()
	defer func() {
		if cerr := iter.Close(); cerr != nil {
			dm.logger.Error(ctx, "Error closing iterator", observability.Error(cerr))
		}
	}()
	if iter.Scan(&existingStatus) && (existingStatus == "completed" || existingStatus == "failed") {
		dm.logger.Info(ctx, "Task already has final status, not updating to pending_confirmation.", observability.Int64("task_id", taskID), observability.String("status", existingStatus))
		return nil
	}

	if err := dm.db.NewQuery(queries.UpdateTaskAggregatorSubmitted,
		executionTxHash, proofCID, taskID).Exec(); err != nil {
		dm.logger.Error(ctx, "Error updating task submitted for task ID", observability.Int64("task_id", taskID), observability.Error(err))
		return err
	}
	dm.logger.Info(ctx, "Successfully updated task as pending_confirmation", observability.Int64("task_id", taskID), observability.String("execution_tx_hash", executionTxHash))
	return nil
}

// GetUserEmailByJobID returns the user's email_id for a given job_id
func (dm *DatabaseClient) GetUserEmailByJobID(ctx context.Context, jobID string) (string, error) {
	var userAddress string
	iter := dm.db.NewQuery(queries.GetUserAddressByJobId, jobID).Iter()
	defer func() {
		if cerr := iter.Close(); cerr != nil {
			dm.logger.Error(ctx, "Error closing iterator", observability.Error(cerr))
		}
	}()

	if !iter.Scan(&userAddress) {
		return "", fmt.Errorf("user not found for job ID %s", jobID)
	}

	var email string
	iter = dm.db.NewQuery(queries.GetUserEmailByUserAddress, userAddress).Iter()
	defer func() {
		if cerr := iter.Close(); cerr != nil {
			dm.logger.Error(ctx, "Error closing iterator", observability.Error(cerr))
		}
	}()

	if !iter.Scan(&email) {
		return "", fmt.Errorf("email not found for user address %s", userAddress)
	}

	return email, nil
}

// GetUserEmailByTaskID returns the user's email_id for a given task_id
func (dm *DatabaseClient) GetUserEmailByTaskID(ctx context.Context, taskID int64) (string, error) {
	// Both task_opx_predicted_cost and job_id are stored as text in DB, so scan as strings first
	var predictedStr string
	var jobID string
	iter := dm.db.NewQuery(queries.GetTaskCostAndJobId, taskID).Iter()
	defer func() {
		if cerr := iter.Close(); cerr != nil {
			dm.logger.Error(ctx, "Error closing iterator", observability.Error(cerr))
		}
	}()
	if !iter.Scan(&predictedStr, &jobID) {
		return "", fmt.Errorf("job not found for task ID %d", taskID)
	}

	return dm.GetUserEmailByJobID(ctx, jobID)
}

// UpdatePointsInDatabase updates points for all involved parties in a task
func (dm *DatabaseClient) UpdateKeeperPointsInDatabase(ctx context.Context, data types.TaskSubmissionData) error {
	var jobID string
	var userAddress string
	var userTasks int64

	var keeperPointsStr string
	var rewardsBooster float64
	var noAttestedTasks int64
	var noExecutedTasks int64

	// Get task cost and job ID
	// Both task_opx_predicted_cost and job_id are stored as text in DB, so scan as strings first
	var taskPredictedOpxCostStr string
	iter := dm.db.NewQuery(queries.GetTaskCostAndJobId, data.TaskID).Iter()
	defer func() {
		if cerr := iter.Close(); cerr != nil {
			dm.logger.Error(ctx, "Error closing iterator", observability.Error(cerr))
		}
	}()

	if !iter.Scan(&taskPredictedOpxCostStr, &jobID) {
		dm.logger.Error(ctx, "Failed to get task fee and job ID for task ID", observability.Int64("task_id", data.TaskID))
		return fmt.Errorf("task not found for task ID %d", data.TaskID)
	}

	// TODO:
	// Alert if taskOpxCost is greater than taskPredictedOpxCost by a threshold
	// Use types.IsGreater(data.TaskOpxCost, taskPredictedOpxCostStr) for comparison
	_ = taskPredictedOpxCostStr // Keep for future alerting logic

	// Update the Attester Points
	for _, operatorID := range data.AttesterIds {
		// Get keeper_address from operator_id
		keeperAddress, err := dm.GetKeeperAddressByOperatorID(ctx, operatorID)
		if err != nil {
			dm.logger.Error(ctx, "Failed to get keeper address for operator_id", observability.Int64("operator_id", operatorID), observability.Error(err))
			return fmt.Errorf("keeper not found for operator_id %d: %w", operatorID, err)
		}

		// Use RetryableIter since the query needs parameters
		iter := dm.db.NewQuery(queries.GetAttesterPointsAndNoOfTasks, operatorID).Iter()
		defer func() {
			if cerr := iter.Close(); cerr != nil {
				dm.logger.Error(ctx, "Error closing iterator", observability.Error(cerr))
			}
		}()

		var keeperAddressFromQuery string
		if !iter.Scan(&keeperAddressFromQuery, &keeperPointsStr, &rewardsBooster, &noAttestedTasks) {
			dm.logger.Error(ctx, "Failed to get keeper points for operator_id", observability.Int64("operator_id", operatorID))
			return fmt.Errorf("keeper not found for operator_id %d", operatorID)
		}
		// Calculate new keeper points using string-based math
		// keeperPoints = keeperPoints + (rewardsBooster * TaskOpxCost)
		rewardedFee := types.MulByFloat(data.TaskOpxCost, rewardsBooster)
		keeperPointsStr = types.Add(keeperPointsStr, rewardedFee)
		noAttestedTasks = noAttestedTasks + 1

		if err := dm.db.NewQuery(queries.UpdateAttesterPointsAndNoOfTasks,
			keeperPointsStr, noAttestedTasks, keeperAddress).Exec(); err != nil {
			dm.logger.Error(ctx, "Failed to update keeper points", observability.Error(err))
			return err
		}
	}

	// Update the Performer Points
	// PerformerAddress is a string containing consensus_address
	if data.PerformerAddress == "" {
		dm.logger.Error(ctx, "Performer address is empty", observability.Int64("task_id", data.TaskID))
		return fmt.Errorf("performer address is empty for task %d", data.TaskID)
	}
	performerAddress, err := dm.GetKeeperAddressByConsensusAddress(ctx, data.PerformerAddress)
	if err != nil {
		dm.logger.Error(ctx, "Failed to get performer address by consensus address", observability.Error(err))
		return err
	}
	// Use RetryableIter since the query needs parameters
	iter = dm.db.NewQuery(queries.GetPerformerPointsAndNoOfTasks, performerAddress).Iter()
	defer func() {
		if cerr := iter.Close(); cerr != nil {
			dm.logger.Error(ctx, "Error closing iterator", observability.Error(cerr))
		}
	}()

	if !iter.Scan(&keeperPointsStr, &rewardsBooster, &noExecutedTasks) {
		dm.logger.Error(ctx, "Failed to get keeper points for performer_address", observability.String("performer_address", performerAddress))
		return fmt.Errorf("keeper not found for performer_address %s", performerAddress)
	}
	// Calculate new keeper points using string-based math
	if data.IsAccepted {
		// keeperPoints = keeperPoints + (rewardsBooster * TaskOpxCost)
		rewardedFee := types.MulByFloat(data.TaskOpxCost, rewardsBooster)
		keeperPointsStr = types.Add(keeperPointsStr, rewardedFee)
	} else {
		// keeperPoints = keeperPoints - (rewardsBooster * TaskOpxCost * 0.1)
		penaltyFee := types.MulByFloat(data.TaskOpxCost, rewardsBooster*0.1)
		keeperPointsStr = types.Sub(keeperPointsStr, penaltyFee)
	}
	noExecutedTasks = noExecutedTasks + 1

	if err := dm.db.NewQuery(queries.UpdatePerformerPointsAndNoOfTasks,
		keeperPointsStr, noExecutedTasks, performerAddress).Exec(); err != nil {
		dm.logger.Error(ctx, "Failed to update keeper points", observability.Error(err))
		return err
	}

	// Update the User Points
	iter = dm.db.NewQuery(queries.GetUserAddressByJobId, jobID).Iter()
	defer func() {
		if cerr := iter.Close(); cerr != nil {
			dm.logger.Error(ctx, "Error closing iterator", observability.Error(cerr))
		}
	}()

	if !iter.Scan(&userAddress) {
		dm.logger.Error(ctx, "Failed to get user address for job ID", observability.String("job_id", jobID))
		return fmt.Errorf("user not found for job ID %s", jobID)
	}

	var userPointsStr string
	iter = dm.db.NewQuery(queries.GetUserPoints, userAddress).Iter()
	defer func() {
		if cerr := iter.Close(); cerr != nil {
			dm.logger.Error(ctx, "Error closing iterator", observability.Error(cerr))
		}
	}()

	if !iter.Scan(&userPointsStr, &userTasks) {
		dm.logger.Error(ctx, "Failed to get user points for user address", observability.String("user_address", userAddress))
		return fmt.Errorf("user not found for user address %s", userAddress)
	}

	userTasks = userTasks + 1
	// Calculate new user points using string-based math
	userPointsStr = types.Add(userPointsStr, data.TaskOpxCost)
	lastUpdatedAt := time.Now().UTC()

	if err := dm.db.NewQuery(queries.UpdateUserPoints,
		userPointsStr, userTasks, lastUpdatedAt, userAddress).Exec(); err != nil {
		dm.logger.Error(ctx, "Failed to update user points for user address", observability.String("user_address", userAddress), observability.Error(err))
		return err
	}

	var jobCostActualStr string
	iter = dm.db.NewQuery(queries.GetJobCostActual, jobID).Iter()
	defer func() {
		if cerr := iter.Close(); cerr != nil {
			dm.logger.Error(ctx, "Error closing iterator", observability.Error(cerr))
		}
	}()

	if !iter.Scan(&jobCostActualStr) {
		dm.logger.Error(ctx, "Failed to get job cost actual for job ID", observability.String("job_id", jobID))
		return fmt.Errorf("job not found for job ID %s", jobID)
	}

	// Calculate new job cost actual using string-based math
	jobCostActualStr = types.Add(jobCostActualStr, data.TaskOpxCost)

	if err := dm.db.NewQuery(queries.UpdateJobCostActual, jobCostActualStr, jobID).Exec(); err != nil {
		dm.logger.Error(ctx, "Failed to update job cost actual for job ID", observability.String("job_id", jobID), observability.Error(err))
		return err
	}

	dm.logger.Info(ctx, "Successfully updated points for user address", observability.String("user_address", userAddress), observability.String("task_opx_cost", data.TaskOpxCost))
	return nil
}

// GetKeeperAddresses validates that keeper addresses exist (keeper_address is now the primary key)
func (dm *DatabaseClient) GetKeeperAddresses(ctx context.Context, keeperAddresses []string) ([]string, error) {
	var validAddresses []string
	for _, keeperAddress := range keeperAddresses {
		keeperAddress = strings.ToLower(keeperAddress)
		// Since keeper_address is the primary key, we can directly use it
		// Just validate that it exists by checking if we can get consensus address
		_, err := dm.GetConsensusAddressByKeeperAddress(ctx, keeperAddress)
		if err != nil {
			dm.logger.Error(ctx, "Failed to validate keeper address", observability.String("keeper_address", keeperAddress), observability.Error(err))
			return nil, fmt.Errorf("keeper not found for address %s: %w", keeperAddress, err)
		}
		validAddresses = append(validAddresses, keeperAddress)
	}
	return validAddresses, nil
}

// GetConsensusAddressByKeeperAddress gets the consensus address for a given keeper address
func (dm *DatabaseClient) GetConsensusAddressByKeeperAddress(ctx context.Context, keeperAddress string) (string, error) {
	keeperAddress = strings.ToLower(keeperAddress)
	var consensusAddress string

	iter := dm.db.NewQuery(queries.GetConsensusAddressByKeeperAddress, keeperAddress).Iter()
	defer func() {
		if cerr := iter.Close(); cerr != nil {
			dm.logger.Error(ctx, "Error closing iterator", observability.Error(cerr))
		}
	}()

	if !iter.Scan(&consensusAddress) {
		dm.logger.Error(ctx, "Failed to get consensus address for keeper", observability.String("keeper_address", keeperAddress))
		return "", fmt.Errorf("keeper not found for address %s", keeperAddress)
	}

	return consensusAddress, nil
}

// GetKeeperAddressByConsensusAddress gets the keeper address for a given consensus address
func (dm *DatabaseClient) GetKeeperAddressByConsensusAddress(ctx context.Context, consensusAddress string) (string, error) {
	consensusAddress = strings.ToLower(consensusAddress)
	var keeperAddress string

	iter := dm.db.NewQuery(queries.GetKeeperAddressByConsensusAddress, consensusAddress).Iter()
	defer func() {
		if cerr := iter.Close(); cerr != nil {
			dm.logger.Error(ctx, "Error closing iterator", observability.Error(cerr))
		}
	}()

	if !iter.Scan(&keeperAddress) {
		dm.logger.Error(ctx, "Failed to get keeper address for consensus address", observability.String("consensus_address", consensusAddress))
		return "", fmt.Errorf("keeper not found for consensus address %s", consensusAddress)
	}

	return keeperAddress, nil
}

// GetKeeperAddressByOperatorID gets the keeper address for a given operator ID
func (dm *DatabaseClient) GetKeeperAddressByOperatorID(ctx context.Context, operatorID int64) (string, error) {
	var keeperAddress string

	iter := dm.db.NewQuery(queries.GetKeeperAddressByOperatorID, operatorID).Iter()
	defer func() {
		if cerr := iter.Close(); cerr != nil {
			dm.logger.Error(ctx, "Error closing iterator", observability.Error(cerr))
		}
	}()

	if !iter.Scan(&keeperAddress) {
		dm.logger.Error(ctx, "Failed to get keeper address for operator_id", observability.Int64("operator_id", operatorID))
		return "", fmt.Errorf("keeper not found for operator_id %d", operatorID)
	}

	return keeperAddress, nil
}

// UpdateScriptStorage updates script storage for a custom job (TaskDefinitionID = 7)
// This is called after task execution to persist storage updates from the custom script
func (dm *DatabaseClient) UpdateScriptStorage(ctx context.Context, jobID string, storageUpdates map[string]string) error {
	if len(storageUpdates) == 0 {
		dm.logger.Debug(ctx, "No storage updates for job", observability.String("job_id", jobID))
		return nil
	}

	dm.logger.Info(ctx, "Updating storage keys for job", observability.Int("storage_count", len(storageUpdates)), observability.String("job_id", jobID))

	// Upsert each storage key-value pair
	for key, value := range storageUpdates {
		if err := dm.db.NewQuery(queries.UpsertScriptStorageQuery,
			jobID, key, value, time.Now().UTC()).Exec(); err != nil {
			dm.logger.Error(ctx, "Failed to update storage key for job", observability.String("key", key), observability.String("job_id", jobID), observability.Error(err))
			return fmt.Errorf("failed to update storage: %w", err)
		}
		dm.logger.Debug(ctx, "Updated storage", observability.String("job_id", jobID), observability.String("key", key))
	}

	dm.logger.Info(ctx, "Successfully updated storage keys for job", observability.Int("storage_count", len(storageUpdates)), observability.String("job_id", jobID))
	return nil
}

// GetJobIDByTaskID retrieves the job ID for a given task ID
func (dm *DatabaseClient) GetJobIDByTaskID(ctx context.Context, taskID int64) (string, error) {
	var jobID string
	iter := dm.db.NewQuery(queries.GetJobIDByTaskIDQuery, taskID).Iter()
	defer func() {
		if cerr := iter.Close(); cerr != nil {
			dm.logger.Error(ctx, "Error closing iterator", observability.Error(cerr))
		}
	}()

	if !iter.Scan(&jobID) {
		return "", fmt.Errorf("job not found for task ID %d", taskID)
	}

	return jobID, nil
}

// GetTaskDefinitionIDByTaskID retrieves the task definition ID for a given task ID
func (dm *DatabaseClient) GetTaskDefinitionIDByTaskID(ctx context.Context, taskID int64) (int, error) {
	var taskDefinitionID int
	iter := dm.db.NewQuery(queries.GetTaskDefinitionIDQuery, taskID).Iter()
	defer func() {
		if cerr := iter.Close(); cerr != nil {
			dm.logger.Error(ctx, "Error closing iterator", observability.Error(cerr))
		}
	}()

	if !iter.Scan(&taskDefinitionID) {
		return 0, fmt.Errorf("task definition ID not found for task ID %d", taskID)
	}

	return taskDefinitionID, nil
}
