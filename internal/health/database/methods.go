package database

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type KeeperHealthCheckIn struct {
	KeeperAddress    string    `json:"keeper_address" validate:"required,eth_addr"`
	ConsensusPubKey  string    `json:"consensus_pub_key" validate:"required"`
	ConsensusAddress string    `json:"consensus_address" validate:"required,eth_addr"`
	Version          string    `json:"version" validate:"required"`
	Timestamp        time.Time `json:"timestamp" validate:"required"`
	Signature        string    `json:"signature" validate:"required"`
	IsImua           bool      `json:"is_imua" validate:"required"`
}

// UpdateKeeperStatus updates the status of a keeper.
// Only called when the isActive state in KeeperState is changed.
// If false -> true, update the keeper "online" status to true, and the last checked in timestamp.
// If true -> false, update the keeper "online" status to false, the uptime and the last checked in timestamp.
func (dm *DatabaseManager) UpdateKeeperStatus(
	ctx context.Context,
	keeperAddress string,
	consensusAddress string,
	version string,
	uptime int64,
	timestamp time.Time,
	publicIP string,
	isActive bool,
) error {
	// Use pointers for optional fields to avoid updating unset fields
	online := isActive
	uptimePtr := uptime

	keeperData := &types.KeeperDataEntity{
		KeeperAddress:    strings.ToLower(keeperAddress),
		Online:           online,
		Uptime:           uptimePtr,
		LastCheckedIn:    timestamp,
		ConsensusAddress: strings.ToLower(consensusAddress),
		Version:          version,
		PublicIP:         publicIP,
	}

	// Keeper has gone offline, set online = false, uptime and timestamp
	if err := dm.keeperRepo.Update(ctx, keeperData); err != nil {
		dm.logger.Error("Failed to update keeper inactive status",
			"error", err,
			"keeper", keeperData.KeeperAddress,
		)
		return err
	}

	return nil
}

// GetVerifiedKeepers retrieves only verified keepers from the database
func (dm *DatabaseManager) GetVerifiedKeepers(ctx context.Context) ([]types.HealthKeeperInfo, error) {
	// Get all keepers and filter for verified ones
	// Pass boolean values directly for query filtering (not pointers)
	allKeepers, err := dm.keeperRepo.GetByFields(ctx, map[string]interface{}{
		"registered":  true,
		"whitelisted": true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get verified keepers: %w", err)
	}

	var keepers []types.HealthKeeperInfo
	for _, keeper := range allKeepers {
		// Helper functions to safely dereference pointers
		online := false
		if keeper.Online {
			online = keeper.Online
		}

		uptime := int64(0)
		if keeper.Uptime != 0 {
			uptime = keeper.Uptime
		}

		operatorID := int64(0)
		if keeper.OperatorID != 0 {
			operatorID = keeper.OperatorID
		}

		onImua := false
		if keeper.OnImua {
			onImua = keeper.OnImua
		}

		keepers = append(keepers, types.HealthKeeperInfo{
			KeeperName:       keeper.KeeperName,
			KeeperAddress:    keeper.KeeperAddress,
			ConsensusAddress: keeper.ConsensusAddress,
			OperatorID:       fmt.Sprintf("%d", operatorID),
			Version:          keeper.Version,
			IsActive:         online,
			Uptime:           uptime,
			LastCheckedIn:    keeper.LastCheckedIn,
			IsImua:           onImua,
		})
	}

	dm.logger.Debug("Retrieved verified keepers from database",
		"count", len(keepers),
	)
	return keepers, nil
}

// UpdateAllKeepersStatus updates the status of all keepers upon service shutdown
// Will update the last checked in timestamp and uptime values for keepers
// OnlineKeepers is a list of keepers which are currently online
func (dm *DatabaseManager) UpdateAllKeepersStatus(ctx context.Context, onlineKeepers []types.HealthKeeperInfo) error {
	now := time.Now().UTC()

	// Update each online keeper's status
	for _, keeper := range onlineKeepers {
		uptime := keeper.Uptime
		// Prepare partial update with only the fields we want to modify
		update := &types.KeeperDataEntity{
			KeeperAddress: keeper.KeeperAddress,
			Uptime:        uptime,
			LastCheckedIn: now,
		}

		// Update the keeper status
		if err := dm.keeperRepo.Update(ctx, update); err != nil {
			dm.logger.Error("Failed to update keeper status on shutdown",
				"error", err,
				"keeper_address", keeper.KeeperAddress,
			)
			continue
		}
	}

	dm.logger.Info("Completed status update for all keepers on shutdown")
	return nil
}

// UpdateKeeperChatID updates the chat ID for a keeper
func (dm *DatabaseManager) UpdateKeeperChatID(ctx context.Context, keeperAddress string, chatID int64) error {
	keeper, err := dm.keeperRepo.GetByNonID(ctx, "keeper_address", keeperAddress)
	if err != nil {
		return err
	}
	keeper.ChatID = chatID
	if err := dm.keeperRepo.Update(ctx, keeper); err != nil {
		return err
	}
	return nil
}

// GetKeeperChatInfo gets the chat ID for a keeper
func (dm *DatabaseManager) GetKeeperChatInfo(ctx context.Context, keeperAddress string) (int64, string, error) {
	keeper, err := dm.keeperRepo.GetByNonID(ctx, "keeper_address", keeperAddress)
	if err != nil {
		return 0, "", err
	}

	chatID := int64(0)
	if keeper.ChatID != 0 {
		chatID = keeper.ChatID
	}

	return chatID, keeper.EmailID, nil
}

// AddKeeperPoints adds points to a keeper's rewards balance
func (dm *DatabaseManager) AddKeeperPoints(ctx context.Context, keeperAddress string, points int64) error {
	// Get current keeper data
	keeper, err := dm.keeperRepo.GetByNonID(ctx, "keeper_address", strings.ToLower(keeperAddress))
	if err != nil {
		dm.logger.Error("Failed to get keeper for points update",
			"error", err,
			"keeper", keeperAddress,
		)
		return fmt.Errorf("failed to get keeper: %w", err)
	}

	// Parse current points
	currentPoints := int64(0)
	if keeper.KeeperPoints != "" {
		var ok bool
		parsedPoints := new(big.Int)
		parsedPoints, ok = parsedPoints.SetString(keeper.KeeperPoints, 10)
		if ok {
			currentPoints = parsedPoints.Int64()
		}
	}

	// Add new points
	newTotal := currentPoints + points
	newPointsStr := new(big.Int).SetInt64(newTotal).String()

	// Update keeper points
	update := &types.KeeperDataEntity{
		KeeperAddress: strings.ToLower(keeperAddress),
		KeeperPoints:  newPointsStr,
	}

	if err := dm.keeperRepo.Update(ctx, update); err != nil {
		dm.logger.Error("Failed to update keeper points",
			"error", err,
			"keeper", keeperAddress,
			"points_added", points,
		)
		return fmt.Errorf("failed to update keeper points: %w", err)
	}

	dm.logger.Info("Updated keeper points",
		"keeper", keeperAddress,
		"points_added", points,
		"new_total", newTotal,
	)

	return nil
}
