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
			IsActive:         false,
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

// DumpState updates all keepers to inactive in the database
func (sm *StateManager) DumpState(ctx context.Context) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.logger.Info("Dumping keeper state to database...")

	for address, state := range sm.keepers {
		if state.IsActive {
			// Create a minimal health check-in with just the address
			health := types.HealthKeeperInfo{
				KeeperAddress: address,
			}

			if err := sm.retryWithBackoff(ctx, func() error {
				return sm.updateKeeperStatusInDatabase(ctx, health, false)
			}, maxRetries); err != nil {
				sm.logger.Error("Failed to update keeper status during state dump",
					"error", err,
					"keeper", address,
				)
				continue
			}
		}
	}

	sm.logger.Info("Successfully dumped keeper state")
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
