package keeper

import (
	"context"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

const (
	inactivityThreshold  = 70 * time.Second
	stateCleanupInterval = 5 * time.Second
	periodicDumpInterval = 5 * time.Minute // Persist uptime every 5 minutes
)

func (sm *StateManager) startCleanupRoutine() {
	ticker := time.NewTicker(stateCleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		sm.checkInactiveKeepers()
	}
}

func (sm *StateManager) checkInactiveKeepers() {
	now := time.Now().UTC()
	var inactiveKeepers []*types.HealthKeeperInfo

	sm.mu.Lock()
	for address, state := range sm.keepers {
		if state.IsActive && now.Sub(state.LastCheckedIn) > inactivityThreshold {
			sm.logger.Info("Keeper became inactive",
				"keeper", address,
				"lastSeen", state.LastCheckedIn.Format(time.RFC3339),
			)
			state.IsActive = false
			// Store a copy of the complete state
			stateCopy := *state
			inactiveKeepers = append(inactiveKeepers, &stateCopy)
		}
	}
	sm.mu.Unlock()

	// Update database for all inactive keepers
	for _, keeperState := range inactiveKeepers {
		if err := sm.updateKeeperStatusInDatabase(context.Background(), keeperState, false); err != nil {
			sm.logger.Error("Failed to update inactive status",
				"error", err,
				"keeper", keeperState.KeeperAddress,
			)
		}
	}
}

// startPeriodicDumpRoutine periodically persists keeper state to database
func (sm *StateManager) startPeriodicDumpRoutine() {
	ticker := time.NewTicker(periodicDumpInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := sm.PeriodicDump(context.Background()); err != nil {
			sm.logger.Error("Periodic dump failed", "error", err)
		}
	}
}
