package keeper

import (
	"context"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/health/config"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

func (sm *StateManager) startCleanupRoutine(ctx context.Context) {
	ticker := time.NewTicker(config.GetHealthCheckInterval())
	defer ticker.Stop()

	for range ticker.C {
		sm.checkInactiveKeepers(ctx)
	}
}

func (sm *StateManager) checkInactiveKeepers(ctx context.Context) {
	now := time.Now().UTC()
	var inactiveKeepers []string

	sm.mu.Lock()
	for address, state := range sm.keepers {
		if state.IsActive && now.Sub(state.LastCheckedIn) > config.GetHealthCheckKeeperTimeout() {
			sm.logger.Info(ctx, "Keeper became inactive",
				observability.String("keeper", address),
				observability.String("last_seen", state.LastCheckedIn.Format(time.RFC3339)),
			)
			state.IsActive = false
			inactiveKeepers = append(inactiveKeepers, address)
		}
	}
	sm.mu.Unlock()

	for _, address := range inactiveKeepers {
		keeperHealth := types.KeeperHealthCheckIn{
			KeeperAddress: address,
		}

		if err := sm.updateKeeperStatusInDatabase(ctx, keeperHealth, false); err != nil {
			sm.logger.Error(ctx, "Failed to update inactive status",
				observability.Error(err),
				observability.String("keeper", address),
			)
		}
	}
}
