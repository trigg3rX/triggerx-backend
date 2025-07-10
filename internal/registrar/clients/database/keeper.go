package database

import (
	"fmt"
	"strings"
	"time"

	"github.com/gocql/gocql"

	"github.com/trigg3rX/triggerx-backend/internal/registrar/types"
)

// KeeperRegistered registers a new keeper or updates an existing one (status = true)
func (dm *DatabaseClient) UpdateKeeperRegistrationData(data types.KeeperRegistrationData) (int64, bool, error) {
	data.OperatorAddress = strings.ToLower(data.OperatorAddress)
	data.TxHash = strings.ToLower(data.TxHash)
	data.RewardsReceiver = strings.ToLower(data.RewardsReceiver)

	dm.logger.Infof("Updating keeper %s at database", data.OperatorAddress)

	var booster float32 = 1
	var keeperID int64

	err := dm.db.Session().Query(`
		SELECT keeper_id FROM triggerx.keeper_data WHERE keeper_address = ? ALLOW FILTERING`,
		data.OperatorAddress).Scan(&keeperID)

	if err == gocql.ErrNotFound {
		var maxKeeperID int64
		if err := dm.db.Session().Query(`
			SELECT MAX(keeper_id) FROM triggerx.keeper_data`).Scan(&maxKeeperID); err != nil {
			dm.logger.Debug("No keeper ID found, creating new keeper")
			maxKeeperID = 0
		}

		keeperID = maxKeeperID + 1

		if err := dm.db.Session().Query(`
			INSERT INTO triggerx.keeper_data (
				keeper_id, keeper_address, rewards_address, registered_tx, operator_id, voting_power, strategies, registered, rewards_booster
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			keeperID, data.OperatorAddress, data.RewardsReceiver, data.TxHash, data.OperatorID, data.VotingPower, data.Strategies, true, booster).Exec(); err != nil {
			dm.logger.Errorf("Error creating new keeper: %v", err)
			return 0, false, err
		}

		dm.logger.Infof("Keeper registered: %d | %s", keeperID, data.OperatorAddress)
		return keeperID, false, nil
	} else if err != nil {
		dm.logger.Errorf("Error getting keeper ID: %v", err)
		return 0, false, err
	} else {
		if err := dm.db.Session().Query(`
			UPDATE triggerx.keeper_data SET 
				rewards_address = ?, registered_tx = ?, operator_id = ?, voting_power = ?, strategies = ?, registered = ?, rewards_booster = ?
			WHERE keeper_id = ?`,
			data.RewardsReceiver, data.TxHash, data.OperatorID, data.VotingPower, data.Strategies, true, booster, keeperID).Exec(); err != nil {
			dm.logger.Errorf("Error updating keeper with ID %d: %v", keeperID, err)
			return 0, false, err
		}
		dm.logger.Infof("Keeper registered: %d | %s", keeperID, data.OperatorAddress)
		return keeperID, true, nil
	}
}

// KeeperUnregistered marks a keeper as unregistered (status = false)
func (dm *DatabaseClient) KeeperUnregistered(operatorAddress string) error {
	var currentKeeperID int64
	operatorAddress = strings.ToLower(operatorAddress)

	if err := dm.db.Session().Query(`
		SELECT keeper_id FROM triggerx.keeper_data WHERE keeper_address = ? ALLOW FILTERING`,
		operatorAddress).Scan(&currentKeeperID); err != nil {
		dm.logger.Errorf("Error getting keeper ID: %v", err)
		return err
	}

	if err := dm.db.Session().Query(`
		UPDATE triggerx.keeper_data SET 
			registered = ?
		WHERE keeper_id = ?`,
		false, currentKeeperID).Exec(); err != nil {
		dm.logger.Errorf("Error updating keeper with ID %d: %v", currentKeeperID, err)
		return err
	}

	dm.logger.Infof("Keeper unregistered: %d | %s", currentKeeperID, operatorAddress)
	return nil
}

// UpdatePointsInDatabase updates points for all involved parties in a task
func (dm *DatabaseClient) UpdateKeeperPointsInDatabase(taskID int, keeperIds []string, isAccepted bool) error {
	var taskOpxCost float64
	var jobID int64
	var userID int64

	// Get task cost and job ID
	if err := dm.db.Session().Query(`
		SELECT task_opx_cost, job_id 
		FROM triggerx.task_data 
		WHERE task_id = ?`,
		taskID).Scan(&taskOpxCost, &jobID); err != nil {
		dm.logger.Errorf("Failed to get task fee and job ID for task ID %d: %v", taskID, err)
		return err
	}

	dm.logger.Infof("Task ID %d has a cost of %f and job ID %d", taskID, taskOpxCost, jobID)

	// Get user ID from job ID
	if err := dm.db.Session().Query(`
		SELECT user_id 
		FROM triggerx.job_data 
		WHERE job_id = ?`,
		jobID).Scan(&userID); err != nil {
		dm.logger.Errorf("Failed to get user ID for job ID %d: %v", jobID, err)
		return err
	}

	// Update the Keeper Points, neglect Performer if isAccepted is false
	for _, keeperId := range keeperIds {
		if !isAccepted {
			continue
		}

		var keeperPoints float64
		var rewardsBooster float32

		if err := dm.db.Session().Query(`
			SELECT keeper_points, rewards_booster FROM triggerx.keeper_data
			WHERE keeper_id = ? ALLOW FILTERING`,
			keeperId).Scan(&keeperPoints, &rewardsBooster); err != nil {
			dm.logger.Error(fmt.Sprintf("Failed to get keeper points: %v", err))
			return err
		}

		keeperPoints = keeperPoints + float64(rewardsBooster)*taskOpxCost

		if err := dm.db.Session().Query(`
			UPDATE triggerx.keeper_data 
			SET keeper_points = ?
			WHERE keeper_id = ?`,
			keeperPoints, keeperId).Exec(); err != nil {
			dm.logger.Error(fmt.Sprintf("Failed to update keeper points: %v", err))
			return err
		}
	}

	// Update the User Points
	var userPoints float64
	if err := dm.db.Session().Query(`
		SELECT user_points FROM triggerx.user_data
		WHERE user_id = ?`,
		userID).Scan(&userPoints); err != nil {
		dm.logger.Errorf("Failed to get user points: %v", err)
		return err
	}

	userPoints = userPoints + taskOpxCost
	lastUpdatedAt := time.Now().UTC()

	if err := dm.db.Session().Query(`
		UPDATE triggerx.user_data 
		SET user_points = ?, last_updated_at = ?
		WHERE user_id = ?`,
		userPoints, lastUpdatedAt, userID).Exec(); err != nil {
		dm.logger.Errorf("Failed to update user points for user ID %d: %v", userID, err)
		return err
	}
	dm.logger.Infof("Successfully updated points for user ID %d: added %.2f points", userID, taskOpxCost)
	return nil
}
