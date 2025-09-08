package database

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/clients/database/queries"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/types"
)

// UpdateTaskSubmissionData updates task number, success status and execution details in database
func (dm *DatabaseClient) UpdateTaskSubmissionData(data types.TaskSubmissionData) error {
	// dm.logger.Infof("Updating task %d with task number %d and acceptance status %t", data.TaskID, data.TaskNumber, data.IsAccepted)

	performerId, err := dm.GetKeeperIds([]string{data.PerformerAddress})
	if err != nil {
		dm.logger.Errorf("Failed to get performer ID: %v", err)
		return err
	}
	attesterIds := data.AttesterIds

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
		data.TaskID).Exec(); err != nil {
		dm.logger.Errorf("Error updating task execution details for task ID %d: %v", data.TaskID, err)
		return err
	}

	dm.logger.Infof("Successfully updated task %d with submission details", data.TaskID)
	return nil
}

func (dm *DatabaseClient) UpdateTaskFailed(taskID int64) error {
	if err := dm.db.NewQuery(queries.UpdateTaskFailed, taskID).Exec(); err != nil {
		dm.logger.Errorf("Error updating task failed for task ID %d: %v", taskID, err)
		return err
	}
	dm.logger.Infof("Successfully updated task %d as failed", taskID)
	return nil
}

// UpdatePointsInDatabase updates points for all involved parties in a task
func (dm *DatabaseClient) UpdateKeeperPointsInDatabase(data types.TaskSubmissionData) error {
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
			dm.logger.Errorf("Error closing iterator: %v", cerr)
		}
	}()

	if !iter.Scan(&taskPredictedOpxCost, &jobID) {
		dm.logger.Errorf("Failed to get task fee and job ID for task ID %d: no results found", data.TaskID)
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
				dm.logger.Errorf("Error closing iterator: %v", cerr)
			}
		}()

		if !iter.Scan(&keeperId, &keeperPoints, &rewardsBooster, &noAttestedTasks) {
			dm.logger.Error(fmt.Sprintf("Failed to get keeper points for operator_id %d: no results found", operator_id))
			return fmt.Errorf("keeper not found for operator_id %d", operator_id)
		}
		keeperPoints = keeperPoints + float64(rewardsBooster)*data.TaskOpxCost
		noAttestedTasks = noAttestedTasks + 1

		// dm.logger.Infof("Keeper points: %f, Rewards booster: %f, No attested tasks: %d", keeperPoints, rewardsBooster, noAttestedTasks)

		if err := dm.db.NewQuery(queries.UpdateAttesterPointsAndNoOfTasks,
			keeperPoints, noAttestedTasks, keeperId).Exec(); err != nil {
			dm.logger.Error(fmt.Sprintf("Failed to update keeper points: %v", err))
			return err
		}
	}

	// Update the Performer Points
	performerId, err := dm.GetKeeperIds([]string{data.PerformerAddress})
	if err != nil {
		dm.logger.Errorf("Failed to get performer ID: %v", err)
		return err
	}
	// Use RetryableIter since the query needs parameters
	iter = dm.db.NewQuery(queries.GetPerformerPointsAndNoOfTasks, performerId[0]).Iter()
	defer func() {
		if cerr := iter.Close(); cerr != nil {
			dm.logger.Errorf("Error closing iterator: %v", cerr)
		}
	}()

	if !iter.Scan(&keeperPoints, &rewardsBooster, &noExecutedTasks) {
		dm.logger.Error(fmt.Sprintf("Failed to get keeper points for performer_id %d: no results found", performerId[0]))
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
		dm.logger.Error(fmt.Sprintf("Failed to update keeper points: %v", err))
		return err
	}

	// Update the User Points
	iter = dm.db.NewQuery(queries.GetUserIdByJobId, jobID).Iter()
	defer func() {
		if cerr := iter.Close(); cerr != nil {
			dm.logger.Errorf("Error closing iterator: %v", cerr)
		}
	}()

	if !iter.Scan(&userID) {
		dm.logger.Errorf("Failed to get user ID for job ID %d: no results found", jobID)
		return fmt.Errorf("user not found for job ID %d", jobID)
	}

	var userPoints float64
	iter = dm.db.NewQuery(queries.GetUserPoints, userID).Iter()
	defer func() {
		if cerr := iter.Close(); cerr != nil {
			dm.logger.Errorf("Error closing iterator: %v", cerr)
		}
	}()

	if !iter.Scan(&userPoints, &userTasks) {
		dm.logger.Errorf("Failed to get user points for user ID %d: no results found", userID)
		return fmt.Errorf("user not found for user ID %d", userID)
	}

	userTasks = userTasks + 1
	userPoints = userPoints + data.TaskOpxCost
	lastUpdatedAt := time.Now().UTC()

	if err := dm.db.NewQuery(queries.UpdateUserPoints,
		userPoints, userTasks, lastUpdatedAt, userID).Exec(); err != nil {
		dm.logger.Errorf("Failed to update user points for user ID %d: %v", userID, err)
		return err
	}

	var jobCostActual float64
	iter = dm.db.NewQuery(queries.GetJobCostActual, jobID).Iter()
	defer func() {
		if cerr := iter.Close(); cerr != nil {
			dm.logger.Errorf("Error closing iterator: %v", cerr)
		}
	}()

	if !iter.Scan(&jobCostActual) {
		dm.logger.Errorf("Failed to get job cost actual for job ID %d: no results found", jobID)
		return fmt.Errorf("job not found for job ID %d", jobID)
	}

	jobCostActual = jobCostActual + data.TaskOpxCost

	if err := dm.db.NewQuery(queries.UpdateJobCostActual, jobCostActual, jobID).Exec(); err != nil {
		dm.logger.Errorf("Failed to update job cost actual for job ID %d: %v", jobID, err)
		return err
	}

	dm.logger.Infof("Successfully updated points for user ID %d: added %.2f points", userID, data.TaskOpxCost)
	return nil
}

// GetKeeperIds gets keeper IDs from keeper addresses
func (dm *DatabaseClient) GetKeeperIds(keeperAddresses []string) ([]int64, error) {
	var keeperIds []int64
	for _, keeperAddress := range keeperAddresses {
		var keeperID int64
		keeperAddress = strings.ToLower(keeperAddress)

		// Use RetryableIter since the query needs parameters
		iter := dm.db.NewQuery(queries.GetKeeperIDByAddress, keeperAddress).Iter()
		defer func() {
			if cerr := iter.Close(); cerr != nil {
				dm.logger.Errorf("Error closing iterator: %v", cerr)
			}
		}()

		if iter.Scan(&keeperID) {
			dm.logger.Infof("Keeper ID for address %s: %d", keeperAddress, keeperID)
			keeperIds = append(keeperIds, keeperID)
		} else {
			dm.logger.Errorf("Failed to get keeper ID for address %s: no results found", keeperAddress)
			return nil, fmt.Errorf("keeper not found for address %s", keeperAddress)
		}
	}
	return keeperIds, nil
}
