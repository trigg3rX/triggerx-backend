package registrar

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// Function to add a keeper to the database via the API when an operator is registered
func AddKeeperToDatabase(operatorAddress string, blsKeysArray [4]*big.Int, txHash string) error {
	logger.Info(fmt.Sprintf("Adding operator %s to database as keeper", operatorAddress))

	// Convert BLS keys to string array for database
	blsKeys := make([]string, 4)
	for i, key := range blsKeysArray {
		blsKeys[i] = key.String()
	}

	var currentKeeperID int64
	if err := db.Session().Query(`
		SELECT keeper_id FROM triggerx.keeper_data WHERE keeper_address = ? ALLOW FILTERING`,
		operatorAddress).Scan(&currentKeeperID); err != nil {
		dbLogger.Errorf("[CreateKeeperData] Error getting max keeper ID: %v", err)
		return err
	}

	dbLogger.Infof("[CreateKeeperData] Updating keeper with ID: %d", currentKeeperID)
	if err := db.Session().Query(`
        UPDATE triggerx.keeper_data SET 
            registered_tx = ?, consensus_keys = ?, status = ?
        WHERE keeper_id = ?`,
		txHash, blsKeys, true, currentKeeperID).Exec(); err != nil {
		dbLogger.Errorf("[CreateKeeperData] Error creating keeper with ID %d: %v", currentKeeperID, err)
		return err
	}

	logger.Info(fmt.Sprintf("Successfully added keeper %s to database", operatorAddress))
	return nil
}

// New function to handle database updates
func UpdatePointsInDatabase(taskID int, performerAddress common.Address, attestersIds []string) error {
	if db == nil {
		logger.Error("Database connection is not initialized")
		return fmt.Errorf("database connection not initialized, please restart the service")
	}

	// Get task fee
	var taskFee int64
	err := db.Session().Query(`SELECT task_fee FROM triggerx.task_data WHERE task_id = ?`,
		taskID).Scan(&taskFee)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get task fee for task ID %d: %v", taskID, err))
		return fmt.Errorf("failed to get task fee for task ID %d: %v", taskID, err)
	}

	logger.Info(fmt.Sprintf("Task ID %d has a fee of %d", taskID, taskFee))

	// Skip if performer address is empty
	if performerAddress != (common.Address{}) {
		if err := UpdatePerformerPoints(performerAddress.Hex(), taskFee); err != nil {
			return err
		}
	}

	// Process attesters
	for _, attesterId := range attestersIds {
		if attesterId != "" {
			if err := UpdateAttesterPoints(attesterId, taskFee); err != nil {
				logger.Error(fmt.Sprintf("Attester points update failed: %v", err))
				continue
			}
		}
	}

	return nil
}

// Helper function to update performer points
func UpdatePerformerPoints(performerAddress string, taskFee int64) error {
	var performerPoints int64
	var performerId int64

	// First, get the keeper_id using the keeper_address (requires a scan with ALLOW FILTERING)
	if err := db.Session().Query(`
		SELECT keeper_id, keeper_points FROM triggerx.keeper_data 
		WHERE keeper_address = ? ALLOW FILTERING`,
		performerAddress).Scan(&performerId, &performerPoints); err != nil {
		logger.Error(fmt.Sprintf("Failed to get performer ID and points: %v", err))
		return err
	}

	//multiplyer 2 till 7th April
	newPerformerPoints := performerPoints + 2*taskFee

	// Now update using the primary key (keeper_id)
	if err := db.Session().Query(`
		UPDATE triggerx.keeper_data 
		SET keeper_points = ? 
		WHERE keeper_id = ?`,
		newPerformerPoints, performerId).Exec(); err != nil {
		logger.Error(fmt.Sprintf("Failed to update performer points: %v", err))
		return err
	}

	logger.Info(fmt.Sprintf("Added %d points to performer %s (ID: %d)", taskFee, performerAddress, performerId))
	return nil
}

// Helper function to update attester points
func UpdateAttesterPoints(attesterId string, taskFee int64) error {
	var attesterPoints int64
	var keeperID int

	if err := db.Session().Query(`
		SELECT keeper_id FROM triggerx.keeper_data
		WHERE attester_id = ? ALLOW FILTERING`,
	attesterId).Scan(&keeperID); err != nil {
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

	// Calculate new points (2x multiplier until April 7th)
	newAttesterPoints := attesterPoints + 2 * taskFee

	// Try updating with integer ID first
	if err := db.Session().Query(`
        UPDATE triggerx.keeper_data 
        SET keeper_points = ? 
        WHERE keeper_id = ?`,
	newAttesterPoints, keeperID).Exec(); err != nil {
		logger.Error(fmt.Sprintf("Failed to update attester points: %v", err))
		return err
	}	

	logger.Info(fmt.Sprintf("Added %d points to attester ID %s (total: %d)", taskFee, attesterId, newAttesterPoints))
	return nil
}
