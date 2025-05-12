package database

import (
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

var (
	logger = logging.GetLogger(logging.Development, logging.RegistrarProcess)
	db     *database.Connection
)

func SetDatabaseConnection(connection *database.Connection) {
	db = connection
	logger.Info("Database connection set")
}

func KeeperRegistered(operatorAddress string, txHash string) error {
	logger.Infof("Updating keeper %s at database", operatorAddress)

	var booster float32 = 1
	var currentKeeperID int64
	if err := db.Session().Query(`
		SELECT keeper_id FROM triggerx.keeper_data WHERE keeper_address = ? ALLOW FILTERING`,
		operatorAddress).Scan(&currentKeeperID); err != nil {
		logger.Debugf("Keeper ID with address %s not found", operatorAddress)
		currentKeeperID = 0
	}

	if currentKeeperID == 0 {
		var maxKeeperID int64
		if err := db.Session().Query(`
			SELECT MAX(keeper_id) FROM triggerx.keeper_data`).Scan(&maxKeeperID); err != nil {
			logger.Debug("No keeper ID found, creating new keeper")
			maxKeeperID = 0
		}

		currentKeeperID = maxKeeperID + 1

		if err := db.Session().Query(`
			INSERT INTO triggerx.keeper_data (
				keeper_id, keeper_address, registered_tx, status, rewards_booster
			) VALUES (?, ?, ?, ?, ?)`,
			currentKeeperID, operatorAddress, txHash, true, booster).Exec(); err != nil {
			logger.Errorf(" Error creating new keeper: %v", err)
		}

		logger.Infof("Created new keeper with ID: %d", currentKeeperID)
		return nil
	} else {
		if err := db.Session().Query(`
			UPDATE triggerx.keeper_data SET 
				registered_tx = ?, status = ?
			WHERE keeper_id = ?`,
			txHash, true, currentKeeperID).Exec(); err != nil {
			logger.Errorf("Error updating keeper with ID %d: %v", currentKeeperID, err)
		}

		logger.Infof("Updated keeper with ID: %d", currentKeeperID)
	}
	return nil
}

func KeeperUnregistered(operatorAddress string) error {
	var currentKeeperID int64
	if err := db.Session().Query(`
		SELECT keeper_id FROM triggerx.keeper_data WHERE keeper_address = ? ALLOW FILTERING`,
		operatorAddress).Scan(&currentKeeperID); err != nil {
		logger.Errorf("Error getting keeper ID: %v", err)
		return err
	}

	if err := db.Session().Query(`
        UPDATE triggerx.keeper_data SET 
            status = ?
        WHERE keeper_id = ?`,
		false, currentKeeperID).Exec(); err != nil {
		logger.Errorf("Error updating keeper with ID %d: %v", currentKeeperID, err)
		return err
	}

	logger.Infof("Successfully updated keeper %s status to unregistered", operatorAddress)
	return nil
}

func UpdatePointsInDatabase(taskID int, performerAddress common.Address, attestersIds []string, isAccepted bool) error {
	var taskFee float64
	var jobID int64
	var userID int64

	// Get task fee and job ID
	if err := db.Session().Query(`
		SELECT task_fee, job_id 
		FROM triggerx.task_data 
		WHERE task_id = ?`,
		taskID).Scan(&taskFee, &jobID); err != nil {
		logger.Errorf("Failed to get task fee and job ID for task ID %d: %v", taskID, err)
		return err
	}

	logger.Infof("Task ID %d has a fee of %f and job ID %d", taskID, taskFee, jobID)

	// Get user ID from job ID
	if err := db.Session().Query(`
		SELECT user_id 
		FROM triggerx.job_data 
		WHERE job_id = ?`,
		jobID).Scan(&userID); err != nil {
		logger.Errorf("Failed to get user ID for job ID %d: %v", jobID, err)
		return err
	}

	// Update performer points
	err := UpdatePerformerPoints(performerAddress.Hex(), taskFee, isAccepted)
	if err != nil {
		return err
	}

	// Update attester points
	for _, attesterId := range attestersIds {
		if attesterId != "" {
			if err := UpdateAttesterPoints(attesterId, taskFee); err != nil {
				logger.Error(fmt.Sprintf("Attester points update failed: %v", err))
				continue
			}
		}
	}

	// Update user points
	if err := UpdateUserPoints(userID, taskFee); err != nil {
		logger.Errorf("Failed to update user points for user ID %d: %v", userID, err)
		return err
	}

	return nil
}

func UpdatePerformerPoints(performerAddress string, taskFee float64, isAccepted bool) error {
	var performerPoints float64
	var performerId int64
	var rewardsBooster float32

	performerAddress = strings.ToLower(performerAddress)

	if err := db.Session().Query(`
		SELECT keeper_id, keeper_points, rewards_booster FROM triggerx.keeper_data 
		WHERE keeper_address = ? ALLOW FILTERING`,
		performerAddress).Scan(&performerId, &performerPoints, &rewardsBooster); err != nil {
		logger.Error(fmt.Sprintf("Failed to get performer ID and points: %v", err))
		return err
	}

	newPerformerPoints := performerPoints + float64(rewardsBooster)*taskFee

	if err := db.Session().Query(`
		UPDATE triggerx.keeper_data 
		SET keeper_points = ? 
		WHERE keeper_id = ?`,
		newPerformerPoints, performerId).Exec(); err != nil {
		logger.Error(fmt.Sprintf("Failed to update performer points: %v", err))
		return err
	}

	logger.Infof("Added %f points to performer %s (ID: %d)", taskFee, performerAddress, performerId)
	return nil
}

func UpdateAttesterPoints(attesterId string, taskFee float64) error {
	var attesterPoints float64
	var keeperID int64
	var rewardsBooster float32

	if err := db.Session().Query(`
		SELECT keeper_id, rewards_booster FROM triggerx.keeper_data
		WHERE operator_id = ? ALLOW FILTERING`,
		attesterId).Scan(&keeperID, &rewardsBooster); err != nil {
		logger.Error(fmt.Sprintf("Failed to update attester points: %v", err))
		return err
	}

	if err := db.Session().Query(`
        SELECT keeper_points FROM triggerx.keeper_data 
        WHERE keeper_id = ? ALLOW FILTERING`,
		keeperID).Scan(&attesterPoints); err != nil {
		logger.Error(fmt.Sprintf("Failed to update attester points: %v", err))
		return err
	}

	newAttesterPoints := attesterPoints + float64(rewardsBooster)*taskFee

	if err := db.Session().Query(`
        UPDATE triggerx.keeper_data 
        SET keeper_points = ? 
        WHERE keeper_id = ?`,
		newAttesterPoints, keeperID).Exec(); err != nil {
		logger.Error(fmt.Sprintf("Failed to update attester points: %v", err))
		return err
	}

	logger.Infof("Added %f points to attester ID %s (total: %f)", taskFee, attesterId, newAttesterPoints)
	return nil
}

func UpdateUserPoints(userID int64, points float64) error {
	if err := db.Session().Query(`
		UPDATE triggerx.user_data 
		SET user_points = user_points + ?, last_updated_at = ?
		WHERE user_id = ?`,
		points, time.Now().UTC(), userID).Exec(); err != nil {
		logger.Errorf("Failed to update user points for user ID %d: %v", userID, err)
		return err
	}
	logger.Infof("Successfully updated points for user ID %d: added %.2f points", userID, points)
	return nil
}

func DailyRewardsPoints() error {
	var keeperID int64
	var rewardsBooster float32
	var keeperPoints float64
	var currentKeeperPoints []types.DailyRewardsPoints

	iter := db.Session().Query(`
		SELECT keeper_id, rewards_booster, keeper_points FROM triggerx.keeper_data
		WHERE status = true AND verified = true ALLOW FILTERING`).Iter()

	for iter.Scan(&keeperID, &rewardsBooster, &keeperPoints) {
		currentKeeperPoints = append(currentKeeperPoints, types.DailyRewardsPoints{
			KeeperID:       keeperID,
			RewardsBooster: rewardsBooster,
			KeeperPoints:   keeperPoints,
		})
	}
	if err := iter.Close(); err != nil {
		logger.Errorf("Failed to get daily rewards points: %v", err)
		return err
	}

	for _, currentKeeperPoint := range currentKeeperPoints {
		newPoints := currentKeeperPoint.KeeperPoints + float64(10*currentKeeperPoint.RewardsBooster)

		if err := db.Session().Query(`
			UPDATE triggerx.keeper_data 
			SET keeper_points = ? 
			WHERE keeper_id = ?`,
			newPoints, currentKeeperPoint.KeeperID).Exec(); err != nil {
			logger.Errorf("Failed to update daily rewards for keeper ID %d: %v", currentKeeperPoint.KeeperID, err)
			continue
		}

		logger.Infof("Added %d daily reward points to keeper ID %d (new total: %d)",
			10*currentKeeperPoint.RewardsBooster,
			currentKeeperPoint.KeeperID,
			newPoints)
	}

	return nil
}

// UpdateOperatorDetails updates the keeper_data table with details fetched from contracts
func UpdateOperatorDetails(operatorAddress string, operatorId string, votingPower string, rewardsReceiver string, strategies []string) error {
	logger.Infof("Updating operator details for %s in database", operatorAddress)

	var keeperId int64
	if err := db.Session().Query(`
		SELECT keeper_id FROM triggerx.keeper_data WHERE keeper_address = ? ALLOW FILTERING`,
		operatorAddress).Scan(&keeperId); err != nil {
		logger.Errorf("Could not find keeper with address %s: %v", operatorAddress, err)
		return err
	}

	if err := db.Session().Query(`
		UPDATE triggerx.keeper_data 
		SET operator_id = ?, rewards_address = ?, voting_power = ?, strategies = ?
		WHERE keeper_id = ?`,
		operatorId, rewardsReceiver, votingPower, strategies, keeperId).Exec(); err != nil {
		logger.Errorf("Failed to update operator_id for keeper ID %d: %v", keeperId, err)
		return err
	}

	logger.Infof("Successfully updated keeper %s details in database", operatorAddress)
	return nil
}
