package database

import (
	"github.com/trigg3rX/triggerx-backend/internal/registrar/clients/database/queries"
	"github.com/trigg3rX/triggerx-backend/internal/registrar/types"
)

// DailyRewardsPoints processes daily rewards for all eligible keepers
func (dm *DatabaseClient) DailyRewardsPoints() error {
	var keeperID int64
	var rewardsBooster float64
	var keeperPoints float64
	var currentKeeperPoints []types.DailyRewardsPoints

	iter := dm.db.RetryableIter(queries.GetDailyRewardsPoints)

	for iter.Scan(&keeperID, &rewardsBooster, &keeperPoints) {
		if keeperID == 1 || keeperID == 2 || keeperID == 3 || keeperID == 4 {
			continue
		}

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

		err := dm.db.RetryableExec(queries.UpdateKeeperPoints,
			newPoints, currentKeeperPoint.KeeperID)
		if err != nil {
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
