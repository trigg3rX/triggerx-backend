package client

import (
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// DatabaseManager handles database operations
type DatabaseManager struct {
	logger logging.Logger
	db     *database.Connection
}

var instance *DatabaseManager

// InitDatabaseManager initializes the database manager with a logger
func InitDatabaseManager(logger logging.Logger, connection *database.Connection) {
	if logger == nil {
		panic("logger cannot be nil")
	}
	if connection == nil {
		panic("database connection cannot be nil")
	}

	instance = &DatabaseManager{
		logger: logger.With("component", "database"),
		db:     connection,
	}
}

// GetInstance returns the database manager instance
func GetInstance() *DatabaseManager {
	if instance == nil {
		panic("database manager not initialized")
	}
	return instance
}

// KeeperRegistered registers a new keeper or updates an existing one (status = true)
func (dm *DatabaseManager) KeeperRegistered(operatorAddress string, txHash string) error {
	dm.logger.Infof("Updating keeper %s at database", operatorAddress)

	var booster float32 = 1
	var currentKeeperID int64
	operatorAddress = strings.ToLower(operatorAddress)
	txHash = strings.ToLower(txHash)

	if err := dm.db.Session().Query(`
		SELECT keeper_id FROM triggerx.keeper_data WHERE keeper_address = ? ALLOW FILTERING`,
		operatorAddress).Scan(&currentKeeperID); err != nil {
		dm.logger.Debugf("Keeper ID with address %s not found", operatorAddress)
		currentKeeperID = 0
	}

	if currentKeeperID == 0 {
		var maxKeeperID int64
		if err := dm.db.Session().Query(`
			SELECT MAX(keeper_id) FROM triggerx.keeper_data`).Scan(&maxKeeperID); err != nil {
			dm.logger.Debug("No keeper ID found, creating new keeper")
			maxKeeperID = 0
		}

		currentKeeperID = maxKeeperID + 1

		if err := dm.db.Session().Query(`
			INSERT INTO triggerx.keeper_data (
				keeper_id, keeper_address, registered_tx, registered, rewards_booster
			) VALUES (?, ?, ?, ?, ?)`,
			currentKeeperID, operatorAddress, txHash, true, booster).Exec(); err != nil {
			dm.logger.Errorf("Error creating new keeper: %v", err)
			return err
		}

		dm.logger.Infof("Keeper registered: %d | %s", currentKeeperID, operatorAddress)
		return nil
	} else {
		if err := dm.db.Session().Query(`
			UPDATE triggerx.keeper_data SET 
				registered_tx = ?, registered = ?
			WHERE keeper_id = ?`,
			txHash, true, currentKeeperID).Exec(); err != nil {
			dm.logger.Errorf("Error updating keeper with ID %d: %v", currentKeeperID, err)
			return err
		}
		dm.logger.Infof("Keeper registered: %d | %s", currentKeeperID, operatorAddress)
		return nil
	}
}

// KeeperUnregistered marks a keeper as unregistered (status = false)
func (dm *DatabaseManager) KeeperUnregistered(operatorAddress string) error {
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
func (dm *DatabaseManager) UpdatePointsInDatabase(taskID int, performerAddress common.Address, attestersIds []string, isAccepted bool) error {
	var taskFee float64
	var jobID int64
	var userID int64

	if err := dm.db.Session().Query(`
		SELECT task_opx_cost, job_id 
		FROM triggerx.task_data 
		WHERE task_id = ?`,
		taskID).Scan(&taskFee, &jobID); err != nil {
		dm.logger.Errorf("Failed to get task fee and job ID for task ID %d: %v", taskID, err)
		return err
	}

	dm.logger.Infof("Task ID %d has a cost of %f and job ID %d", taskID, taskFee, jobID)

	if err := dm.db.Session().Query(`
		SELECT user_id 
		FROM triggerx.job_data 
		WHERE job_id = ?`,
		jobID).Scan(&userID); err != nil {
		dm.logger.Errorf("Failed to get user ID for job ID %d: %v", jobID, err)
		return err
	}

	if err := dm.UpdatePerformerPoints(performerAddress.Hex(), taskFee, isAccepted); err != nil {
		return err
	}

	for _, attesterId := range attestersIds {
		if attesterId != "" {
			if err := dm.UpdateAttesterPoints(attesterId, taskFee); err != nil {
				dm.logger.Error(fmt.Sprintf("Attester points update failed: %v", err))
				continue
			}
		}
	}

	if err := dm.UpdateUserPoints(userID, taskFee); err != nil {
		dm.logger.Errorf("Failed to update user points for user ID %d: %v", userID, err)
		return err
	}

	return nil
}

// UpdatePerformerPoints updates points for a task performer
func (dm *DatabaseManager) UpdatePerformerPoints(performerAddress string, taskFee float64, isAccepted bool) error {
	var performerPoints float64
	var performerId int64
	var rewardsBooster float32

	performerAddress = strings.ToLower(performerAddress)

	if err := dm.db.Session().Query(`
		SELECT keeper_id, keeper_points, rewards_booster FROM triggerx.keeper_data 
		WHERE keeper_address = ? ALLOW FILTERING`,
		performerAddress).Scan(&performerId, &performerPoints, &rewardsBooster); err != nil {
		dm.logger.Error(fmt.Sprintf("Failed to get performer ID and points: %v", err))
		return err
	}

	newPerformerPoints := performerPoints + float64(rewardsBooster)*taskFee

	if isAccepted {
		if err := dm.db.Session().Query(`
				UPDATE triggerx.keeper_data 
				SET keeper_points = ? 
				WHERE keeper_id = ?`,
			newPerformerPoints, performerId).Exec(); err != nil {
			dm.logger.Error(fmt.Sprintf("Failed to update performer points: %v", err))
			return err
		}
	} else {
		if err := dm.db.Session().Query(`
			UPDATE triggerx.keeper_data 
			SET keeper_points = ? 
			WHERE keeper_id = ?`,
			performerPoints, performerId).Exec(); err != nil {
			dm.logger.Error(fmt.Sprintf("Failed to update performer points: %v", err))
			return err
		}
	}

	dm.logger.Infof("Added %f points to performer %s (ID: %d)", taskFee, performerAddress, performerId)
	return nil
}

// UpdateAttesterPoints updates points for a task attester
func (dm *DatabaseManager) UpdateAttesterPoints(attesterId string, taskFee float64) error {
	var attesterPoints float64
	var keeperID int64
	var rewardsBooster float32

	if err := dm.db.Session().Query(`
		SELECT keeper_id, rewards_booster FROM triggerx.keeper_data
		WHERE operator_id = ? ALLOW FILTERING`,
		attesterId).Scan(&keeperID, &rewardsBooster); err != nil {
		dm.logger.Error(fmt.Sprintf("Failed to update attester points: %v", err))
		return err
	}

	if err := dm.db.Session().Query(`
		SELECT keeper_points FROM triggerx.keeper_data 
		WHERE keeper_id = ? ALLOW FILTERING`,
		keeperID).Scan(&attesterPoints); err != nil {
		dm.logger.Error(fmt.Sprintf("Failed to update attester points: %v", err))
		return err
	}

	newAttesterPoints := attesterPoints + float64(rewardsBooster)*taskFee

	if err := dm.db.Session().Query(`
		UPDATE triggerx.keeper_data 
		SET keeper_points = ? 
		WHERE keeper_id = ?`,
		newAttesterPoints, keeperID).Exec(); err != nil {
		dm.logger.Error(fmt.Sprintf("Failed to update attester points: %v", err))
		return err
	}

	dm.logger.Infof("Added %f points to attester ID %s (total: %f)", taskFee, attesterId, newAttesterPoints)
	return nil
}

// UpdateUserPoints updates points for a user
func (dm *DatabaseManager) UpdateUserPoints(userID int64, points float64) error {
	var userPoints float64
	if err := dm.db.Session().Query(`
		SELECT user_points FROM triggerx.user_data
		WHERE user_id = ?`,
		userID).Scan(&userPoints); err != nil {
		dm.logger.Errorf("Failed to get user points: %v", err)
		return err
	}

	userPoints = userPoints + points
	lastUpdatedAt := time.Now().UTC()

	if err := dm.db.Session().Query(`
		UPDATE triggerx.user_data 
		SET user_points = ?, last_updated_at = ?
		WHERE user_id = ?`,
		userPoints, lastUpdatedAt, userID).Exec(); err != nil {
		dm.logger.Errorf("Failed to update user points for user ID %d: %v", userID, err)
		return err
	}
	dm.logger.Infof("Successfully updated points for user ID %d: added %.2f points", userID, points)
	return nil
}

// DailyRewardsPoints processes daily rewards for all eligible keepers
func (dm *DatabaseManager) DailyRewardsPoints() error {
	var keeperID int64
	var rewardsBooster float32
	var keeperPoints float64
	var currentKeeperPoints []types.DailyRewardsPoints

	iter := dm.db.Session().Query(`
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
		dm.logger.Errorf("Failed to get daily rewards points: %v", err)
		return err
	}

	for _, currentKeeperPoint := range currentKeeperPoints {
		newPoints := currentKeeperPoint.KeeperPoints + float64(10*currentKeeperPoint.RewardsBooster)

		if err := dm.db.Session().Query(`
			UPDATE triggerx.keeper_data 
			SET keeper_points = ? 
			WHERE keeper_id = ?`,
			newPoints, currentKeeperPoint.KeeperID).Exec(); err != nil {
			dm.logger.Errorf("Failed to update daily rewards for keeper ID %d: %v", currentKeeperPoint.KeeperID, err)
			continue
		}

		dm.logger.Infof("Added %d daily reward points to keeper ID %d (new total: %d)",
			10*currentKeeperPoint.RewardsBooster,
			currentKeeperPoint.KeeperID,
			newPoints)
	}

	return nil
}

const (
	maxRetries = 3
	retryDelay = 2 * time.Second
)

// retryWithBackoff executes the given function with exponential backoff retry logic
func retryWithBackoff[T any](operation func() (T, error), logger logging.Logger) (T, error) {
	var result T
	var err error
	delay := retryDelay

	for attempt := 1; attempt <= maxRetries; attempt++ {
		result, err = operation()
		if err == nil {
			return result, nil
		}

		if attempt < maxRetries {
			logger.Warnf("Attempt %d failed: %v. Retrying in %v...", attempt, err, delay)
			time.Sleep(delay)
			delay *= 2 // Exponential backoff
		}
	}

	return result, fmt.Errorf("failed after %d attempts: %v", maxRetries, err)
}

// UpdateOperatorDetails updates the details of an operator
func (dm *DatabaseManager) UpdateOperatorDetails(operatorAddress string, operatorId string, votingPower string, rewardsReceiver string, strategies []string) error {
	operatorAddress = strings.ToLower(operatorAddress)
	dm.logger.Infof("Updating operator details for %s in database", operatorAddress)

	// Retry getting keeper ID
	keeperId, err := retryWithBackoff(func() (int64, error) {
		var id int64
		err := dm.db.Session().Query(`
			SELECT keeper_id FROM triggerx.keeper_data WHERE keeper_address = ? ALLOW FILTERING`,
			operatorAddress).Scan(&id)
		return id, err
	}, dm.logger)
	if err != nil {
		dm.logger.Errorf("Could not find keeper with address %s after retries: %v", operatorAddress, err)
		return err
	}

	// Retry updating operator details
	_, err = retryWithBackoff(func() (interface{}, error) {
		return nil, dm.db.Session().Query(`
			UPDATE triggerx.keeper_data 
			SET operator_id = ?, rewards_address = ?, voting_power = ?, strategies = ?
			WHERE keeper_id = ?`,
			operatorId, rewardsReceiver, votingPower, strategies, keeperId).Exec()
	}, dm.logger)
	if err != nil {
		dm.logger.Errorf("Failed to update operator_id for keeper ID %d after retries: %v", keeperId, err)
		return err
	}

	dm.logger.Infof("Successfully updated keeper %s details in database", operatorAddress)
	return nil
}

// Public wrapper functions
func KeeperRegistered(operatorAddress string, txHash string) error {
	return GetInstance().KeeperRegistered(operatorAddress, txHash)
}

func KeeperUnregistered(operatorAddress string) error {
	return GetInstance().KeeperUnregistered(operatorAddress)
}

func UpdatePointsInDatabase(taskID int, performerAddress common.Address, attestersIds []string, isAccepted bool) error {
	return GetInstance().UpdatePointsInDatabase(taskID, performerAddress, attestersIds, isAccepted)
}

func DailyRewardsPoints() error {
	return GetInstance().DailyRewardsPoints()
}

func UpdateOperatorDetails(operatorAddress string, operatorId string, votingPower string, rewardsReceiver string, strategies []string) error {
	return GetInstance().UpdateOperatorDetails(operatorAddress, operatorId, votingPower, rewardsReceiver, strategies)
}
