package database

import (
	"github.com/trigg3rX/triggerx-backend/internal/registrar/types"
)

// DailyRewardsPoints processes daily rewards for all eligible keepers
func (dm *DatabaseClient) DailyRewardsPoints() error {
	var keeperID int64
	var rewardsBooster float32
	var keeperPoints float64
	var currentKeeperPoints []types.DailyRewardsPoints

	iter := dm.db.Session().Query(`
		SELECT keeper_id, rewards_booster, keeper_points FROM triggerx.keeper_data
		WHERE registered = true AND whitelisted = true ALLOW FILTERING`).Iter()

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

		_, err := retryWithBackoff(func() (interface{}, error) {
			return nil, dm.db.Session().Query(`
				UPDATE triggerx.keeper_data 
				SET keeper_points = ? 
				WHERE keeper_id = ?`,
				newPoints, currentKeeperPoint.KeeperID).Exec()
		}, dm.logger)
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
