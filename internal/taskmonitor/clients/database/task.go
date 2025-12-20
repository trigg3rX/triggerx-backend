package database

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/clients/database/queries"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/types"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

// UpdateTaskSubmissionData updates task number, success status and execution details in database
func (dm *DatabaseClient) UpdateTaskSubmissionData(ctx context.Context, data types.TaskSubmissionData) error {
	// dm.logger.Infof("Updating task %d with task number %d and acceptance status %t", data.TaskID, data.TaskNumber, data.IsAccepted)

	performerId, err := dm.GetKeeperIds(ctx, []string{data.PerformerAddress})
	if err != nil {
		dm.logger.Error(ctx, "Failed to get performer ID", observability.Error(err))
		return err
	}
	attesterIds := data.AttesterIds

	// Convert []interface{} to []string for Cassandra
	convertedArgsStrings := make([]string, len(data.ConvertedArguments))
	for i, arg := range data.ConvertedArguments {
		convertedArgsStrings[i] = fmt.Sprintf("%v", arg)
	}

	if err := dm.db.NewQuery(queries.UpdateTaskSubmissionData,
		data.TaskNumber,
		data.IsAccepted,
		data.TaskSubmissionTxHash,
		performerId[0],
		attesterIds,
		data.ExecutionTxHash,
		data.ExecutionTimestamp,
		data.TaskOpxCost,
		data.ProofOfTask,
		convertedArgsStrings,
		data.TaskID).Exec(); err != nil {
		dm.logger.Error(ctx, "Error updating task execution details for task ID %d", observability.Int64("task_id", data.TaskID), observability.Error(err))
		return err
	}

	dm.logger.Info(ctx, "Successfully updated task %d with submission details", observability.Int64("task_id", data.TaskID))
	return nil
}

func (dm *DatabaseClient) UpdateTaskFailed(ctx context.Context, taskID int64) error {
	// First, check if the task already has a status (i.e., is already failed or completed)
	var existingStatus string
	iter := dm.db.NewQuery(queries.GetTaskStatusByID, taskID).Iter()
	defer func() {
		if cerr := iter.Close(); cerr != nil {
			dm.logger.Error(ctx, "Error closing iterator", observability.Error(cerr))
		}
	}()
	if iter.Scan(&existingStatus) && existingStatus != "" {
		dm.logger.Info(ctx, "Task %d already has a status '%s', not updating to failed.", observability.Int64("task_id", taskID), observability.String("status", existingStatus))
		return nil
	}
	// Proceed with marking as failed only if status is absent or empty
	if err := dm.db.NewQuery(queries.UpdateTaskFailed, taskID).Exec(); err != nil {
		dm.logger.Error(ctx, "Error updating task failed for task ID %d", observability.Int64("task_id", taskID), observability.Error(err))
		return err
	}
	dm.logger.Info(ctx, "Successfully updated task %d as failed", observability.Int64("task_id", taskID))
	return nil
}

// UpdateTaskError updates a task with error information
func (dm *DatabaseClient) UpdateTaskError(ctx context.Context, taskID int64, errorMsg string) error {
	// First, check if the task already has a status (i.e., is already failed or completed)
	var existingStatus string
	iter := dm.db.NewQuery(queries.GetTaskStatusByID, taskID).Iter()
	defer func() {
		if cerr := iter.Close(); cerr != nil {
			dm.logger.Error(ctx, "Error closing iterator", observability.Error(cerr))
		}
	}()
	if iter.Scan(&existingStatus) && existingStatus != "" {
		dm.logger.Info(ctx, "Task %d already has a status '%s', not updating to failed.", observability.Int64("task_id", taskID), observability.String("status", existingStatus))
		return nil
	}
	// Proceed with marking as failed only if status is absent or empty
	if err := dm.db.NewQuery(queries.UpdateTaskError, errorMsg, taskID).Exec(); err != nil {
		dm.logger.Error(ctx, "Error updating task error for task ID %d", observability.Int64("task_id", taskID), observability.Error(err))
		return err
	}
	dm.logger.Info(ctx, "Successfully updated task %d with error: %s", observability.Int64("task_id", taskID), observability.String("error_msg", errorMsg))
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
		dm.logger.Info(ctx, "Task %d already has final status '%s', not updating to failed.", observability.Int64("task_id", taskID), observability.String("status", existingStatus))
		return nil
	}

	if err := dm.db.NewQuery(queries.UpdateTaskAggregatorFailed,
		errorMsg, executionTxHash, proofCID, taskID).Exec(); err != nil {
		dm.logger.Error(ctx, "Error updating task failed for task ID %d", observability.Int64("task_id", taskID), observability.Error(err))
		return err
	}
	dm.logger.Info(ctx, "Successfully updated task %d as failed: %s (tx_hash: %s)", observability.Int64("task_id", taskID), observability.String("error_msg", errorMsg), observability.String("execution_tx_hash", executionTxHash))
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
		dm.logger.Info(ctx, "Task %d already has final status '%s', not updating to pending_confirmation.", observability.Int64("task_id", taskID), observability.String("status", existingStatus))
		return nil
	}

	if err := dm.db.NewQuery(queries.UpdateTaskAggregatorSubmitted,
		executionTxHash, proofCID, taskID).Exec(); err != nil {
		dm.logger.Error(ctx, "Error updating task submitted for task ID %d", observability.Int64("task_id", taskID), observability.Error(err))
		return err
	}
	dm.logger.Info(ctx, "Successfully updated task %d as pending_confirmation (tx_hash: %s)", observability.Int64("task_id", taskID), observability.String("execution_tx_hash", executionTxHash))
	return nil
}

// GetUserEmailByJobID returns the user's email_id for a given job_id
func (dm *DatabaseClient) GetUserEmailByJobID(ctx context.Context, jobID *big.Int) (string, error) {
	var userID int64
	iter := dm.db.NewQuery(queries.GetUserIdByJobId, jobID).Iter()
	defer func() {
		if cerr := iter.Close(); cerr != nil {
			dm.logger.Error(ctx, "Error closing iterator", observability.Error(cerr))
		}
	}()

	if !iter.Scan(&userID) {
		return "", fmt.Errorf("user not found for job ID %d", jobID)
	}

	var email string
	iter = dm.db.NewQuery(queries.GetUserEmailByUserID, userID).Iter()
	defer func() {
		if cerr := iter.Close(); cerr != nil {
			dm.logger.Error(ctx, "Error closing iterator", observability.Error(cerr))
		}
	}()

	if !iter.Scan(&email) {
		return "", fmt.Errorf("email not found for user ID %d", userID)
	}

	return email, nil
}

// GetUserEmailByTaskID returns the user's email_id for a given task_id
func (dm *DatabaseClient) GetUserEmailByTaskID(ctx context.Context, taskID int64) (string, error) {
	var predicted float64
	var jobID *big.Int
	iter := dm.db.NewQuery(queries.GetTaskCostAndJobId, taskID).Iter()
	defer func() {
		if cerr := iter.Close(); cerr != nil {
			dm.logger.Error(ctx, "Error closing iterator", observability.Error(cerr))
		}
	}()
	if !iter.Scan(&predicted, &jobID) {
		return "", fmt.Errorf("job not found for task ID %d", taskID)
	}
	return dm.GetUserEmailByJobID(ctx, jobID)
}

// UpdatePointsInDatabase updates points for all involved parties in a task
func (dm *DatabaseClient) UpdateKeeperPointsInDatabase(ctx context.Context, data types.TaskSubmissionData) error {
	var jobID *big.Int
	var userID int64
	var userTasks int64
	var taskPredictedOpxCost float64

	var keeperId int64
	var keeperPoints float64
	var rewardsBooster float64
	var noAttestedTasks int64
	var noExecutedTasks int64

	// Get task cost and job ID
	iter := dm.db.NewQuery(queries.GetTaskCostAndJobId, data.TaskID).Iter()
	defer func() {
		if cerr := iter.Close(); cerr != nil {
			dm.logger.Error(ctx, "Error closing iterator", observability.Error(cerr))
		}
	}()

	if !iter.Scan(&taskPredictedOpxCost, &jobID) {
		dm.logger.Error(ctx, "Failed to get task fee and job ID for task ID %d: no results found", observability.Int64("task_id", data.TaskID))
		return fmt.Errorf("task not found for task ID %d", data.TaskID)
	}

	// dm.logger.Debugf("Details: taskID: %d, taskPredictedOpxCost: %f, taskOpxCost: %f, jobID: %d", data.TaskID, taskPredictedOpxCost, data.TaskOpxCost, jobID)

	// TODO:
	// Alert if taskOpxCost is greater than taskPredictedOpxCost by a threshold

	// Update the Attester Points
	for _, operator_id := range data.AttesterIds {
		// Use RetryableIter since the query needs parameters
		iter := dm.db.NewQuery(queries.GetAttesterPointsAndNoOfTasks, operator_id).Iter()
		defer func() {
			if cerr := iter.Close(); cerr != nil {
				dm.logger.Error(ctx, "Error closing iterator", observability.Error(cerr))
			}
		}()

		if !iter.Scan(&keeperId, &keeperPoints, &rewardsBooster, &noAttestedTasks) {
			dm.logger.Error(ctx, "Failed to get keeper points for operator_id %d: no results found", observability.Int64("operator_id", operator_id))
			return fmt.Errorf("keeper not found for operator_id %d", operator_id)
		}
		keeperPoints = keeperPoints + float64(rewardsBooster)*data.TaskOpxCost
		noAttestedTasks = noAttestedTasks + 1

		// dm.logger.Infof("Keeper points: %f, Rewards booster: %f, No attested tasks: %d", keeperPoints, rewardsBooster, noAttestedTasks)

		if err := dm.db.NewQuery(queries.UpdateAttesterPointsAndNoOfTasks,
			keeperPoints, noAttestedTasks, keeperId).Exec(); err != nil {
			dm.logger.Error(ctx, "Failed to update keeper points", observability.Error(err))
			return err
		}
	}

	// Update the Performer Points
	performerId, err := dm.GetKeeperIds(ctx, []string{data.PerformerAddress})
	if err != nil {
		dm.logger.Error(ctx, "Failed to get performer ID", observability.Error(err))
		return err
	}
	// Use RetryableIter since the query needs parameters
	iter = dm.db.NewQuery(queries.GetPerformerPointsAndNoOfTasks, performerId[0]).Iter()
	defer func() {
		if cerr := iter.Close(); cerr != nil {
			dm.logger.Error(ctx, "Error closing iterator", observability.Error(cerr))
		}
	}()

	if !iter.Scan(&keeperPoints, &rewardsBooster, &noExecutedTasks) {
		dm.logger.Error(ctx, "Failed to get keeper points for performer_id %d: no results found", observability.Int64("performer_id", performerId[0]))
		return fmt.Errorf("keeper not found for performer_id %d", performerId[0])
	}
	if data.IsAccepted {
		keeperPoints = keeperPoints + float64(rewardsBooster)*data.TaskOpxCost
	} else {
		keeperPoints = keeperPoints - float64(rewardsBooster)*data.TaskOpxCost*0.1
	}
	noExecutedTasks = noExecutedTasks + 1

	if err := dm.db.NewQuery(queries.UpdatePerformerPointsAndNoOfTasks,
		keeperPoints, noExecutedTasks, performerId[0]).Exec(); err != nil {
		dm.logger.Error(ctx, "Failed to update keeper points", observability.Error(err))
		return err
	}

	// Update the User Points
	iter = dm.db.NewQuery(queries.GetUserIdByJobId, jobID).Iter()
	defer func() {
		if cerr := iter.Close(); cerr != nil {
			dm.logger.Error(ctx, "Error closing iterator", observability.Error(cerr))
		}
	}()

	if !iter.Scan(&userID) {
		dm.logger.Error(ctx, "Failed to get user ID for job ID %d: no results found", observability.Int64("job_id", jobID.Int64()))
		return fmt.Errorf("user not found for job ID %d", jobID.Int64())
	}

	var userPoints float64
	iter = dm.db.NewQuery(queries.GetUserPoints, userID).Iter()
	defer func() {
		if cerr := iter.Close(); cerr != nil {
			dm.logger.Error(ctx, "Error closing iterator", observability.Error(cerr))
		}
	}()

	if !iter.Scan(&userPoints, &userTasks) {
		dm.logger.Error(ctx, "Failed to get user points for user ID %d: no results found", observability.Int64("user_id", userID))
		return fmt.Errorf("user not found for user ID %d", userID)
	}

	userTasks = userTasks + 1
	userPoints = userPoints + data.TaskOpxCost
	lastUpdatedAt := time.Now().UTC()

	if err := dm.db.NewQuery(queries.UpdateUserPoints,
		userPoints, userTasks, lastUpdatedAt, userID).Exec(); err != nil {
		dm.logger.Error(ctx, "Failed to update user points for user ID %d", observability.Int64("user_id", userID), observability.Error(err))
		return err
	}

	var jobCostActual float64
	iter = dm.db.NewQuery(queries.GetJobCostActual, jobID).Iter()
	defer func() {
		if cerr := iter.Close(); cerr != nil {
			dm.logger.Error(ctx, "Error closing iterator", observability.Error(cerr))
		}
	}()

	if !iter.Scan(&jobCostActual) {
		dm.logger.Error(ctx, "Failed to get job cost actual for job ID %d: no results found", observability.Int64("job_id", jobID.Int64()))
		return fmt.Errorf("job not found for job ID %d", jobID)
	}

	jobCostActual = jobCostActual + data.TaskOpxCost

	if err := dm.db.NewQuery(queries.UpdateJobCostActual, jobCostActual, jobID).Exec(); err != nil {
		dm.logger.Error(ctx, "Failed to update job cost actual for job ID %d", observability.Int64("job_id", jobID.Int64()), observability.Error(err))
		return err
	}

	dm.logger.Info(ctx, "Successfully updated points for user ID %d: added %.2f points", observability.Int64("user_id", userID), observability.Float64("task_opx_cost", data.TaskOpxCost))
	return nil
}

// GetKeeperIds gets keeper IDs from keeper addresses
func (dm *DatabaseClient) GetKeeperIds(ctx context.Context, keeperAddresses []string) ([]int64, error) {
	var keeperIds []int64
	for _, keeperAddress := range keeperAddresses {
		var keeperID int64
		keeperAddress = strings.ToLower(keeperAddress)

		// Use RetryableIter since the query needs parameters
		iter := dm.db.NewQuery(queries.GetKeeperIDByAddress, keeperAddress).Iter()
		defer func() {
			if cerr := iter.Close(); cerr != nil {
				dm.logger.Error(ctx, "Error closing iterator", observability.Error(cerr))
			}
		}()

		if iter.Scan(&keeperID) {
			dm.logger.Info(ctx, "Keeper ID for address %s: %d", observability.String("keeper_address", keeperAddress), observability.Int64("keeper_id", keeperID))
			keeperIds = append(keeperIds, keeperID)
		} else {
			dm.logger.Error(ctx, "Failed to get keeper ID for address %s: no results found", observability.String("keeper_address", keeperAddress))
			return nil, fmt.Errorf("keeper not found for address %s", keeperAddress)
		}
	}
	return keeperIds, nil
}

// UpdateScriptStorage updates script storage for a custom job (TaskDefinitionID = 7)
// This is called after task execution to persist storage updates from the custom script
func (dm *DatabaseClient) UpdateScriptStorage(ctx context.Context, jobID *big.Int, storageUpdates map[string]string) error {
	if len(storageUpdates) == 0 {
		dm.logger.Debug(ctx, "No storage updates for job %s", observability.String("job_id", jobID.String()))
		return nil
	}

	dm.logger.Info(ctx, "Updating %d storage keys for job %s", observability.Int("storage_count", len(storageUpdates)), observability.String("job_id", jobID.String()))

	// Upsert each storage key-value pair
	for key, value := range storageUpdates {
		if err := dm.db.NewQuery(queries.UpsertScriptStorageQuery,
			jobID, key, value, time.Now().UTC()).Exec(); err != nil {
			dm.logger.Error(ctx, "Failed to update storage key '%s' for job %s: %v", observability.String("key", key), observability.String("job_id", jobID.String()), observability.Error(err))
			return fmt.Errorf("failed to update storage: %w", err)
		}
		dm.logger.Debug(ctx, "Updated storage: job=%s, key=%s", observability.String("job_id", jobID.String()), observability.String("key", key))
	}

	dm.logger.Info(ctx, "Successfully updated %d storage keys for job %s", observability.Int("storage_count", len(storageUpdates)), observability.String("job_id", jobID.String()))
	return nil
}

// GetJobIDByTaskID retrieves the job ID for a given task ID
func (dm *DatabaseClient) GetJobIDByTaskID(ctx context.Context, taskID int64) (*big.Int, error) {
	var jobID *big.Int
	iter := dm.db.NewQuery(queries.GetJobIDByTaskIDQuery, taskID).Iter()
	defer func() {
		if cerr := iter.Close(); cerr != nil {
			dm.logger.Error(ctx, "Error closing iterator", observability.Error(cerr))
		}
	}()

	if !iter.Scan(&jobID) {
		return nil, fmt.Errorf("job not found for task ID %d", taskID)
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
