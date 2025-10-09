package keeper

import (
	"context"
	"fmt"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// LoadVerifiedKeepers loads only verified keepers from the database
func (sm *StateManager) LoadVerifiedKeepers(ctx context.Context) error {
	sm.logger.Info("Loading verified keepers from database...")

	// Get only verified keepers from database
	keepers, err := sm.db.GetVerifiedKeepers(ctx)
	if err != nil {
		return fmt.Errorf("failed to load verified keepers from database: %w", err)
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Clear existing state
	sm.keepers = make(map[string]*types.HealthKeeperInfo)

	// Load each keeper's state (initially marked as inactive)
	for _, keeper := range keepers {
		state := &types.HealthKeeperInfo{
			KeeperName:       keeper.KeeperName,
			KeeperAddress:    keeper.KeeperAddress,
			ConsensusAddress: keeper.ConsensusAddress,
			OperatorID:       keeper.OperatorID,
			Version:          keeper.Version,
			IsActive:         keeper.IsActive,
			Uptime:           keeper.Uptime,
			LastCheckedIn:    keeper.LastCheckedIn,
			IsImua:           keeper.IsImua,
		}
		sm.keepers[keeper.KeeperAddress] = state
	}

	sm.logger.Info("Successfully loaded verified keepers",
		"count", len(sm.keepers),
	)
	return nil
}

// DumpState persists all keeper states to database preserving their actual status
func (sm *StateManager) DumpState(ctx context.Context) error {
	sm.logger.Info("Dumping keeper state to database...")

	// Collect all keepers (both active and inactive) to update while holding lock
	sm.mu.Lock()
	var keepersToUpdate []*types.HealthKeeperInfo
	for _, state := range sm.keepers {
		// Create a copy to avoid holding reference to map value
		stateCopy := *state
		keepersToUpdate = append(keepersToUpdate, &stateCopy)
	}
	sm.mu.Unlock()

	// Update database without holding lock, preserving actual isActive state
	for _, state := range keepersToUpdate {
		if err := sm.retryWithBackoff(ctx, func() error {
			// Preserve the actual isActive state instead of forcing to false
			return sm.updateKeeperStatusInDatabase(ctx, state, state.IsActive)
		}, maxRetries); err != nil {
			sm.logger.Error("Failed to update keeper status during state dump",
				"error", err,
				"keeper", state.KeeperAddress,
			)
			continue
		}
	}

	sm.logger.Info("Successfully dumped keeper state",
		"total_keepers", len(keepersToUpdate),
	)
	return nil
}

// PeriodicDump periodically persists uptime for active keepers to database
// This ensures uptime data is not lost if service crashes between check-ins
func (sm *StateManager) PeriodicDump(ctx context.Context) error {
	sm.logger.Debug("Starting periodic keeper state dump...")

	// Collect only active keepers while holding lock
	sm.mu.RLock()
	var activeKeepers []*types.HealthKeeperInfo
	for _, state := range sm.keepers {
		if state.IsActive {
			stateCopy := *state
			activeKeepers = append(activeKeepers, &stateCopy)
		}
	}
	sm.mu.RUnlock()

	if len(activeKeepers) == 0 {
		sm.logger.Debug("No active keepers to dump")
		return nil
	}

	// Update database for all active keepers without holding lock
	successCount := 0
	for _, state := range activeKeepers {
		if err := sm.updateKeeperStatusInDatabase(ctx, state, true); err != nil {
			sm.logger.Warn("Failed to update keeper in periodic dump",
				"error", err,
				"keeper", state.KeeperAddress,
			)
			continue
		}
		successCount++
	}

	sm.logger.Info("Completed periodic keeper state dump",
		"active_keepers", len(activeKeepers),
		"successful_updates", successCount,
	)
	return nil
}

// RetryWithBackoff retries a database operation with exponential backoff
func (sm *StateManager) retryWithBackoff(ctx context.Context, operation func() error, maxRetries int) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err = operation()
		if err == nil {
			return nil
		}

		// Calculate backoff duration (exponential backoff with jitter)
		backoff := time.Duration(i) * time.Second
		sm.logger.Warn("Database operation failed, retrying...",
			"error", err,
			"attempt", i+1,
			"maxRetries", maxRetries,
			"backoff", backoff,
		)

		time.Sleep(backoff)
	}

	return fmt.Errorf("operation failed after %d retries: %w", maxRetries, err)
}
