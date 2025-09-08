package database

import (
	"fmt"
	"strings"

	"github.com/gocql/gocql"

	"github.com/trigg3rX/triggerx-backend/internal/registrar/clients/database/queries"
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
	var err error

	// Use RetryableIter since the query needs parameters
	iter := dm.db.NewQuery(queries.GetKeeperIDByAddress, data.OperatorAddress).Iter()
	defer func() {
		if cerr := iter.Close(); cerr != nil {
			dm.logger.Errorf("Error closing iterator: %v", cerr)
		}
	}()

	if iter.Scan(&keeperID) {
		err = nil
	} else {
		err = gocql.ErrNotFound
	}

	if err == gocql.ErrNotFound {
		var maxKeeperID int64
		if err := dm.db.NewQuery(queries.GetMaxKeeperID, &maxKeeperID).Scan(); err != nil {
			dm.logger.Debug("No keeper ID found, creating new keeper")
			maxKeeperID = 0
		}

		keeperID = maxKeeperID + 1

		if err := dm.db.NewQuery(queries.CreateKeeper,
			keeperID,
			data.OperatorAddress,
			data.RewardsReceiver,
			data.TxHash,
			data.OperatorID,
			data.VotingPower,
			data.Strategies,
			true, booster).Exec(); err != nil {
			dm.logger.Errorf("Error creating new keeper: %v", err)
			return 0, false, err
		}

		dm.logger.Infof("Keeper registered: %d | %s", keeperID, data.OperatorAddress)
		return keeperID, false, nil
	} else if err != nil {
		dm.logger.Errorf("Error getting keeper ID: %v", err)
		return 0, false, err
	} else {
		if err := dm.db.NewQuery(queries.UpdateKeeper,
			data.RewardsReceiver, data.TxHash, data.OperatorID, data.VotingPower,
			data.Strategies, true, booster, keeperID).Exec(); err != nil {
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

	// Use RetryableIter since the query needs parameters
	iter := dm.db.NewQuery(queries.GetKeeperIDByAddress, operatorAddress).Iter()
	defer func() {
		if cerr := iter.Close(); cerr != nil {
			dm.logger.Errorf("Error closing iterator: %v", cerr)
		}
	}()

	if !iter.Scan(&currentKeeperID) {
		dm.logger.Errorf("Error getting keeper ID: no results found")
		return fmt.Errorf("keeper not found for address %s", operatorAddress)
	}

	if err := dm.db.NewQuery(queries.UpdateKeeperRegistrationStatus,
		false, currentKeeperID).Exec(); err != nil {
		dm.logger.Errorf("Error updating keeper with ID %d: %v", currentKeeperID, err)
		return err
	}

	dm.logger.Infof("Keeper unregistered: %d | %s", currentKeeperID, operatorAddress)
	return nil
}
