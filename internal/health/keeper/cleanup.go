package keeper

import (
	"context"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/health/config"
	"github.com/trigg3rX/triggerx-backend/internal/health/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
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
		keeperHealth := types.KeeperHealthCheckInRequest{
			KeeperAddress: address,
		}

		if err := sm.updateKeeperStatusInDatabase(ctx, keeperHealth, false); err != nil {
			sm.logger.Error(ctx, "Failed to update inactive status",
				observability.Error(err),
				observability.String("keeper", address),
			)
		}
	}

	// Update keeper counts metrics after marking keepers as inactive
	if len(inactiveKeepers) > 0 {
		total, active := sm.GetKeeperCount(ctx)
		metrics.UpdateKeeperCounts(ctx, total, active)

		// Update keepers online by version metric
		keepersByVersion := sm.GetKeepersByVersion(ctx)
		metrics.UpdateKeepersOnlineByVersion(ctx, keepersByVersion)
	}
}
