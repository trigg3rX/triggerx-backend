package database

import (
	"strings"
)
// GetKeeperIds gets keeper IDs from keeper addresses
func (dm *DatabaseClient) GetKeeperIds(keeperAddresses []string) ([]int64, error) {
	var keeperIds []int64
	for _, keeperAddress := range keeperAddresses {
		var keeperID int64
		keeperAddress = strings.ToLower(keeperAddress)
		if err := dm.db.Session().Query(`
			SELECT keeper_id FROM triggerx.keeper_data WHERE keeper_address = ? ALLOW FILTERING`,
			keeperAddress).Scan(&keeperID); err != nil {
			dm.logger.Errorf("Failed to get keeper ID for address %s: %v", keeperAddress, err)
			return nil, err
		}
		keeperIds = append(keeperIds, keeperID)
	}
	return keeperIds, nil
}
