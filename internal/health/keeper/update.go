package keeper

import (
	"fmt"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/health/types"
)

const (
	maxRetries = 3
)

// UpdateKeeperHealth updates the health status of a keeper
func (sm *StateManager) UpdateKeeperHealth(keeperHealth types.KeeperHealthCheckIn) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	address := keeperHealth.KeeperAddress
	now := time.Now().UTC()

	existingState, exists := sm.keepers[address]
	if !exists {
		sm.logger.Warn("Received health check-in from unverified keeper",
			"keeper", address,
		)
		return fmt.Errorf("keeper not verified")
	}

	// Update the state with new health check-in data
	existingState.Version = keeperHealth.Version
	existingState.PeerID = keeperHealth.PeerID
	existingState.LastCheckedIn = now
	existingState.IsActive = true

	// Update database
	if err := sm.retryWithBackoff(func() error {
		return sm.updateKeeperStatusInDatabase(keeperHealth, true)
	}, maxRetries); err != nil {
		return fmt.Errorf("failed to update keeper status in database: %w", err)
	}

	sm.logger.Info("Updated keeper health status",
		"keeper", address,
		"version", keeperHealth.Version,
	)
	return nil
}

func (sm *StateManager) updateKeeperStatusInDatabase(keeperHealth types.KeeperHealthCheckIn, isActive bool) error {
	if err := sm.db.UpdateKeeperHealth(keeperHealth, isActive); err != nil {
		return fmt.Errorf("failed to update keeper status in database: %w", err)
	}

	sm.logger.Debug("Updated keeper status in database",
		"keeper", keeperHealth.KeeperAddress,
		"active", isActive,
		"version", keeperHealth.Version,
	)
	return nil
}
