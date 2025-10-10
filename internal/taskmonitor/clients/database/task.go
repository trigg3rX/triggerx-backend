package database

import (
	"context"
	"fmt"

	"strings"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// UpdateTaskSubmissionData updates task number, success status and execution details in database
func (dm *DatabaseClient) UpdateTaskSubmissionData(data types.TaskSubmissionData) error {
	// dm.logger.Infof("Updating task %d with task number %d and acceptance status %t", data.TaskID, data.TaskNumber, data.IsAccepted)

	ctx := context.Background()

	performerId, err := dm.GetKeeperIds([]string{data.PerformerAddress})
	if err != nil {
		dm.logger.Errorf("Failed to get performer ID: %v", err)
		return err
	}
	attesterIds := data.AttesterIds

	// Convert []interface{} to []string for Cassandra
	convertedArgsStrings := make([]string, len(data.ConvertedArguments))
	for i, arg := range data.ConvertedArguments {
		convertedArgsStrings[i] = fmt.Sprintf("%v", arg)
	}

	// Get the task entity
	task, err := dm.taskRepo.GetByID(ctx, data.TaskID)
	if err != nil {
		dm.logger.Errorf("Failed to get task %d: %v", data.TaskID, err)
		return err
	}

	if task == nil {
		dm.logger.Errorf("Task %d not found", data.TaskID)
		return fmt.Errorf("task %d not found", data.TaskID)
	}

	// Update task fields
	task.TaskNumber = data.TaskNumber
	task.IsAccepted = data.IsAccepted
	task.IsSuccessful = true
	task.SubmissionTxHash = data.TaskSubmissionTxHash
	task.TaskPerformerID = performerId[0]
	task.TaskAttesterIDs = attesterIds
	task.ExecutionTxHash = data.ExecutionTxHash
	task.ExecutionTimestamp = data.ExecutionTimestamp
	task.TaskOpxActualCost = data.TaskOpxCost
	task.ProofOfTask = data.ProofOfTask

	if err := dm.taskRepo.Update(ctx, task); err != nil {
		dm.logger.Errorf("Error updating task execution details for task ID %d: %v", data.TaskID, err)
		return err
	}

	dm.logger.Infof("Successfully updated task %d with submission details", data.TaskID)
	return nil
}

func (dm *DatabaseClient) UpdateTaskFailed(taskID int64) error {
	ctx := context.Background()

	// Get the task entity
	task, err := dm.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		dm.logger.Errorf("Failed to get task %d: %v", taskID, err)
		return err
	}

	if task == nil {
		dm.logger.Errorf("Task %d not found", taskID)
		return fmt.Errorf("task %d not found", taskID)
	}

	// Update task fields
	task.IsSuccessful = false

	if err := dm.taskRepo.Update(ctx, task); err != nil {
		dm.logger.Errorf("Error updating task failed for task ID %d: %v", taskID, err)
		return err
	}
	dm.logger.Infof("Successfully updated task %d as failed", taskID)
	return nil
}

// GetUserEmailByJobID returns the user's email_id for a given job_id
func (dm *DatabaseClient) GetUserEmailByJobID(jobID string) (string, error) {
	ctx := context.Background()

	// Get job to find user_address
	job, err := dm.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		dm.logger.Errorf("Failed to get job %v: %v", jobID, err)
		return "", err
	}

	if job == nil {
		return "", fmt.Errorf("user not found for job ID %v", jobID)
	}

	// Get user to find email
	user, err := dm.userRepo.GetByID(ctx, job.UserAddress)
	if err != nil {
		dm.logger.Errorf("Failed to get user %s: %v", job.UserAddress, err)
		return "", err
	}

	if user == nil {
		return "", fmt.Errorf("email not found for user address %s", job.UserAddress)
	}

	return user.EmailID, nil
}

// GetUserEmailByTaskID returns the user's email for a given task_id
func (dm *DatabaseClient) GetUserEmailByTaskID(taskID int64) (string, error) {
	ctx := context.Background()

	// Get task to find job_id
	task, err := dm.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		dm.logger.Errorf("Failed to get task %d: %v", taskID, err)
		return "", err
	}

	if task == nil {
		return "", fmt.Errorf("job not found for task ID %d", taskID)
	}

	return dm.GetUserEmailByJobID(task.JobID)
}

// UpdatePointsInDatabase updates points for all involved parties in a task
func (dm *DatabaseClient) UpdateKeeperPointsInDatabase(data types.TaskSubmissionData) error {
	ctx := context.Background()

	// Get task cost and job ID
	task, err := dm.taskRepo.GetByID(ctx, data.TaskID)
	if err != nil {
		dm.logger.Errorf("Failed to get task %d: %v", data.TaskID, err)
		return err
	}

	if task == nil {
		dm.logger.Errorf("Task %d not found", data.TaskID)
		return fmt.Errorf("task not found for task ID %d", data.TaskID)
	}

	_ = task.TaskOpxPredictedCost // taskPredictedOpxCost - for future use
	jobID := task.JobID

	// dm.logger.Debugf("Details: taskID: %d, taskPredictedOpxCost: %f, taskOpxCost: %f, jobID: %d", data.TaskID, taskPredictedOpxCost, data.TaskOpxCost, jobID)

	// TODO:
	// Alert if taskOpxCost is greater than taskPredictedOpxCost by a threshold

	// Update the Attester Points
	for _, operator_id := range data.AttesterIds {
		// Get keeper by operator_id
		keeper, err := dm.keeperRepo.GetByNonID(ctx, "operator_id", operator_id)
		if err != nil {
			dm.logger.Errorf("Failed to get keeper for operator_id %d: %v", operator_id, err)
			return err
		}

		if keeper == nil {
			dm.logger.Errorf("Keeper not found for operator_id %d", operator_id)
			return fmt.Errorf("keeper not found for operator_id %d", operator_id)
		}

		// Update keeper points and attested tasks
		rewardsBoosterFloat := keeper.RewardsBooster
		pointsToAdd := types.Add(keeper.KeeperPoints, types.Mul(rewardsBoosterFloat, data.TaskOpxCost))
		keeper.KeeperPoints = pointsToAdd

		// Update attested tasks count (handle nil pointer)
		if keeper.NoAttestedTasks == 0 {
			noAttestedTasks := int64(1)
			keeper.NoAttestedTasks = noAttestedTasks
		} else {
			keeper.NoAttestedTasks = keeper.NoAttestedTasks + 1
		}

		// dm.logger.Infof("Keeper points: %f, Rewards booster: %f, No attested tasks: %d", keeper.KeeperPoints, keeper.RewardsBooster, keeper.NoAttestedTasks)

		if err := dm.keeperRepo.Update(ctx, keeper); err != nil {
			dm.logger.Errorf("Failed to update keeper points: %v", err)
			return err
		}
	}

	// Update the Performer Points
	performerId, err := dm.GetKeeperIds([]string{data.PerformerAddress})
	if err != nil {
		dm.logger.Errorf("Failed to get performer ID: %v", err)
		return err
	}

	// Get performer keeper
	performerKeeper, err := dm.keeperRepo.GetByID(ctx, performerId[0])
	if err != nil {
		dm.logger.Errorf("Failed to get performer keeper %d: %v", performerId[0], err)
		return err
	}

	if performerKeeper == nil {
		dm.logger.Errorf("Performer keeper %d not found", performerId[0])
		return fmt.Errorf("keeper not found for performer_id %d", performerId[0])
	}

	rewardsBoosterFloat := performerKeeper.RewardsBooster
	if data.IsAccepted {
		pointsToAdd := types.Add(performerKeeper.KeeperPoints, types.Mul(rewardsBoosterFloat, data.TaskOpxCost))
		performerKeeper.KeeperPoints = pointsToAdd
	} else {
		pointsToSubtract := types.Add(performerKeeper.KeeperPoints, types.Mul(rewardsBoosterFloat, data.TaskOpxCost))
		performerKeeper.KeeperPoints = pointsToSubtract
	}

	// Update executed tasks count (handle nil pointer)
	if performerKeeper.NoExecutedTasks == 0 {
		noExecutedTasks := int64(1)
		performerKeeper.NoExecutedTasks = noExecutedTasks
	} else {
		performerKeeper.NoExecutedTasks = performerKeeper.NoExecutedTasks + 1
	}

	if err := dm.keeperRepo.Update(ctx, performerKeeper); err != nil {
		dm.logger.Errorf("Failed to update keeper points: %v", err)
		return err
	}

	// Update the User Points
	// Get job to find user_address
	job, err := dm.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		dm.logger.Errorf("Failed to get job %d: %v", jobID, err)
		return err
	}

	if job == nil {
		dm.logger.Errorf("Job %d not found", jobID)
		return fmt.Errorf("user not found for job ID %v", jobID)
	}

	// Get user
	user, err := dm.userRepo.GetByID(ctx, job.UserAddress)
	if err != nil {
		dm.logger.Errorf("Failed to get user %s: %v", job.UserAddress, err)
		return err
	}

	if user == nil {
		dm.logger.Errorf("User %s not found", job.UserAddress)
		return fmt.Errorf("user not found for user address %s", job.UserAddress)
	}

	user.TotalTasks = user.TotalTasks + 1
	pointsToAdd := types.Add(user.OpxConsumed, data.TaskOpxCost)
	user.OpxConsumed = pointsToAdd
	user.LastUpdatedAt = time.Now().UTC()

	if err := dm.userRepo.Update(ctx, user); err != nil {
		dm.logger.Errorf("Failed to update user points for user address %s: %v", job.UserAddress, err)
		return err
	}

	// Update job cost actual (we already have the job entity from above)
	costToAdd := types.Add(job.JobCostActual, data.TaskOpxCost)
	job.JobCostActual = costToAdd

	if err := dm.jobRepo.Update(ctx, job); err != nil {
		dm.logger.Errorf("Failed to update job cost actual for job ID %d: %v", jobID, err)
		return err
	}

	dm.logger.Infof("Successfully updated points for user address %s: added %.2f points", job.UserAddress, data.TaskOpxCost)
	return nil
}

// GetKeeperIds gets keeper operator IDs from keeper addresses
func (dm *DatabaseClient) GetKeeperIds(keeperAddresses []string) ([]int64, error) {
	ctx := context.Background()
	var keeperIds []int64

	for _, keeperAddress := range keeperAddresses {
		keeperAddress = strings.ToLower(keeperAddress)

		// Get keeper by address using repository
		keeper, err := dm.keeperRepo.GetByNonID(ctx, "keeper_address", keeperAddress)
		if err != nil {
			dm.logger.Errorf("Failed to get keeper for address %s: %v", keeperAddress, err)
			return nil, err
		}

		if keeper == nil {
			dm.logger.Errorf("Keeper not found for address %s", keeperAddress)
			return nil, fmt.Errorf("keeper not found for address %s", keeperAddress)
		}

		if keeper.OperatorID == 0 {
			dm.logger.Errorf("Keeper operator ID is nil for address %s", keeperAddress)
			return nil, fmt.Errorf("keeper operator ID is nil for address %s", keeperAddress)
		}

		dm.logger.Infof("Keeper operator ID for address %s: %d", keeperAddress, keeper.OperatorID)
		keeperIds = append(keeperIds, keeper.OperatorID)
	}
	return keeperIds, nil
}
