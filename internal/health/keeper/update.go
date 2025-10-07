package keeper

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

const (
	maxRetries = 3
)

// Custom error types
var (
	ErrKeeperNotVerified = errors.New("keeper not verified")
)

// UpdateKeeperHealth updates the health status of a keeper
func (sm *StateManager) UpdateKeeperStatus(ctx context.Context, keeperHealth types.HealthKeeperInfo) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	address := keeperHealth.KeeperAddress
	now := time.Now().UTC()

	existingState, exists := sm.keepers[address]
	if !exists {
		sm.logger.Warn("Received health check-in from unverified keeper",
			"keeper", address,
		)
		return ErrKeeperNotVerified
	}

	// Update the state with new health check-in data
	existingState.Version = keeperHealth.Version
	existingState.LastCheckedIn = now
	existingState.IsActive = true
	existingState.IsImua = keeperHealth.IsImua

	// Update database
	if err := sm.retryWithBackoff(ctx, func() error {
		return sm.updateKeeperStatusInDatabase(ctx, keeperHealth, true)
	}, maxRetries); err != nil {
		return fmt.Errorf("failed to update keeper status in database: %w", err)
	}

	sm.logger.Info("Updated keeper health status",
		"keeper", address,
		"version", keeperHealth.Version,
		"is_imua", keeperHealth.IsImua,
	)
	return nil
}

func (sm *StateManager) updateKeeperStatusInDatabase(ctx context.Context, keeperHealth types.HealthKeeperInfo, isActive bool) error {
	if err := sm.db.UpdateAllKeepersStatus(ctx, []types.HealthKeeperInfo{keeperHealth}); err != nil {
		return fmt.Errorf("failed to update keeper status in database: %w", err)
	}

	sm.logger.Debug("Updated keeper status in database",
		"keeper", keeperHealth.KeeperAddress,
		"active", isActive,
		"version", keeperHealth.Version,
	)
	return nil
}
