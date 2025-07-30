package database

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/registrar/clients/database/queries"
	"github.com/trigg3rX/triggerx-backend/internal/registrar/types"
)

// UpdateTaskSubmissionData updates task number, success status and execution details in database
func (dm *DatabaseClient) UpdateTaskSubmissionData(data types.TaskSubmissionData) error {
	dm.logger.Infof("Updating task %d with number %d and acceptance status %t", data.TaskID, data.TaskNumber, data.IsAccepted)

	performerId := data.KeeperIds[0]
	attesterIds := data.KeeperIds[1:]

	if err := dm.db.RetryableExec(queries.UpdateTaskSubmissionData,
		data.TaskNumber,
		data.IsAccepted,
		data.TaskSubmissionTxHash,
		performerId,
		attesterIds,
		data.AttesterSignatures,
		data.PerformerSignature,
		data.ExecutionTxHash,
		data.ExecutionTimestamp,
		data.TaskOpxCost,
		data.ProofOfTask,
		data.TaskID); err != nil {
		dm.logger.Errorf("Error updating task execution details for task ID %d: %v", data.TaskID, err)
		return err
	}

	dm.logger.Infof("Successfully updated task %d with submission details", data.TaskID)
	return nil
}

// UpdatePointsInDatabase updates points for all involved parties in a task
func (dm *DatabaseClient) UpdateKeeperPointsInDatabase(taskID int, keeperIds []string, taskOpxCost float64, isAccepted bool) error {
	var jobID *big.Int
	var userID int64
	var taskPredictedOpxCost float64

	// Get task cost and job ID
	if err := dm.db.RetryableScan(queries.GetTaskCostAndJobId,
		taskID, &taskPredictedOpxCost, &jobID); err != nil {
		dm.logger.Errorf("Failed to get task fee and job ID for task ID %d: %v", taskID, err)
		return err
	}

	dm.logger.Debugf("Details: taskID: %d, taskPredictedOpxCost: %f, taskOpxCost: %f, jobID: %d", taskID, taskPredictedOpxCost, taskOpxCost, jobID)

	// TODO:
	// Alert if taskOpxCost is greater than taskPredictedOpxCost by a threshold

	// Get user ID from job ID
	if err := dm.db.RetryableScan(queries.GetUserIdByJobId,
		jobID, &userID); err != nil {
		dm.logger.Errorf("Failed to get user ID for job ID %d: %v", jobID, err)
		return err
	}

	// Update the Keeper Points, neglect Performer if isAccepted is false
	for i, keeperId := range keeperIds {
		var keeperPoints float64
		var rewardsBooster float32
		var noExecutedTasks int64
		var noAttestedTasks int64

		if err := dm.db.RetryableScan(queries.GetKeeperPointsAndNoOfTasks,
			keeperId, &keeperPoints, &rewardsBooster, &noExecutedTasks, &noAttestedTasks); err != nil {
			dm.logger.Error(fmt.Sprintf("Failed to get keeper points: %v", err))
			return err
		}

		// Performer if task is accepted, add points and increment no_executed_tasks
		if i == 0 && isAccepted {
			keeperPoints = keeperPoints + float64(rewardsBooster)*taskOpxCost
			noExecutedTasks = noExecutedTasks + 1
		// Performer if task is not accepted, deduct points (10% of taskOpxCost)
		} else if i == 0 && !isAccepted {
			keeperPoints = keeperPoints - float64(rewardsBooster)*taskOpxCost*0.1
		} else if i != 0 {
			keeperPoints = keeperPoints + float64(rewardsBooster)*taskOpxCost
			noAttestedTasks = noAttestedTasks + 1
		}

		if err := dm.db.RetryableExec(queries.UpdateKeeperPointsAndNoOfTasks,
			keeperPoints, noExecutedTasks, noAttestedTasks, keeperId); err != nil {
			dm.logger.Error(fmt.Sprintf("Failed to update keeper points: %v", err))
			return err
		}
	}

	// Update the User Points
	var userPoints float64
	if err := dm.db.RetryableScan(queries.GetUserPoints,
		userID, &userPoints); err != nil {
		dm.logger.Errorf("Failed to get user points: %v", err)
		return err
	}

	userPoints = userPoints + taskOpxCost
	lastUpdatedAt := time.Now().UTC()

	if err := dm.db.RetryableExec(queries.UpdateUserPoints,
		userPoints, lastUpdatedAt, userID); err != nil {
		dm.logger.Errorf("Failed to update user points for user ID %d: %v", userID, err)
		return err
	}
	dm.logger.Infof("Successfully updated points for user ID %d: added %.2f points", userID, taskOpxCost)
	return nil
}

// GetKeeperIds gets keeper IDs from keeper addresses
func (dm *DatabaseClient) GetKeeperIds(keeperAddresses []string) ([]int64, error) {
	var keeperIds []int64
	for _, keeperAddress := range keeperAddresses {
		var keeperID int64
		keeperAddress = strings.ToLower(keeperAddress)
		if err := dm.db.RetryableScan(queries.GetKeeperIDByAddress,
			keeperAddress, &keeperID); err != nil {
			dm.logger.Errorf("Failed to get keeper ID for address %s: %v", keeperAddress, err)
			return nil, err
		}
		keeperIds = append(keeperIds, keeperID)
	}
	return keeperIds, nil
}
