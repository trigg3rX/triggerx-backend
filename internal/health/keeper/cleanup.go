package keeper

import (
	"time"

	commonTypes "github.com/trigg3rX/triggerx-backend-imua/pkg/types"
)

const (
	inactivityThreshold  = 70 * time.Second
	stateCleanupInterval = 5 * time.Second
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
	var inactiveKeepers []string

	sm.mu.Lock()
	for address, state := range sm.keepers {
		if state.IsActive && now.Sub(state.LastCheckedIn) > inactivityThreshold {
			sm.logger.Info("Keeper became inactive",
				"keeper", address,
				"lastSeen", state.LastCheckedIn.Format(time.RFC3339),
			)
			state.IsActive = false
			inactiveKeepers = append(inactiveKeepers, address)
		}
	}
	sm.mu.Unlock()

	for _, address := range inactiveKeepers {
		keeperHealth := commonTypes.KeeperHealthCheckIn{
			KeeperAddress: address,
		}

		if err := sm.updateKeeperStatusInDatabase(keeperHealth, false); err != nil {
			sm.logger.Error("Failed to update inactive status",
				"error", err,
				"keeper", address,
			)
		}
	}
}
