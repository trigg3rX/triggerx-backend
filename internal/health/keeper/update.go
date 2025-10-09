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
	address := keeperHealth.KeeperAddress
	now := time.Now().UTC()

	// Lock briefly to check and update in-memory state
	sm.mu.Lock()
	existingState, exists := sm.keepers[address]
	if !exists {
		sm.mu.Unlock()
		sm.logger.Warn("Received health check-in from unverified keeper",
			"keeper", address,
		)
		return ErrKeeperNotVerified
	}

	// Track if this is a state change (inactive -> active)
	wasActive := existingState.IsActive
	isNowActive := true

	// Update the state with new health check-in data
	existingState.Version = keeperHealth.Version
	existingState.LastCheckedIn = now
	existingState.IsActive = isNowActive
	existingState.IsImua = keeperHealth.IsImua

	// Only increment uptime if keeper was already active (true -> true)
	// Don't increment when keeper just came back online (false -> true)
	if wasActive {
		existingState.Uptime = existingState.Uptime + 60
	}

	// Create a copy of the state for DB update
	stateCopy := *existingState
	stateChanged := (wasActive != isNowActive)
	sm.mu.Unlock()

	// Only update database if state changed (inactive -> active)
	// Periodic dumps will handle persisting uptime for already-active keepers
	if stateChanged {
		if err := sm.retryWithBackoff(ctx, func() error {
			return sm.updateKeeperStatusInDatabase(ctx, &stateCopy, isNowActive)
		}, maxRetries); err != nil {
			return fmt.Errorf("failed to update keeper status in database: %w", err)
		}
		sm.logger.Info("Keeper state changed - updated in database",
			"keeper", address,
			"was_active", wasActive,
			"now_active", isNowActive,
		)
	}

	sm.logger.Info("Updated keeper health status",
		"keeper", address,
		"version", keeperHealth.Version,
		"is_imua", keeperHealth.IsImua,
		"uptime", stateCopy.Uptime,
		"state_changed", stateChanged,
	)
	return nil
}

// updateKeeperStatusInDatabase updates keeper status without acquiring lock
// Caller must ensure they have the keeper state and proper synchronization
func (sm *StateManager) updateKeeperStatusInDatabase(ctx context.Context, state *types.HealthKeeperInfo, isActive bool) error {
	// Use the proper UpdateKeeperStatus method with all required fields
	err := sm.db.UpdateKeeperStatus(
		ctx,
		state.KeeperAddress,
		state.ConsensusAddress,
		state.Version,
		state.Uptime,
		state.LastCheckedIn,
		"", // publicIP - not tracked in HealthKeeperInfo
		isActive,
	)
	if err != nil {
		return fmt.Errorf("failed to update keeper status in database: %w", err)
	}

	sm.logger.Debug("Updated keeper status in database",
		"keeper", state.KeeperAddress,
		"active", isActive,
		"version", state.Version,
		"uptime", state.Uptime,
	)
	return nil
}
