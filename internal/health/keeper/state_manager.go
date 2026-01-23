package keeper

import (
	"context"
	"sync"

	"github.com/trigg3rX/triggerx-backend/internal/health/client"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// StateManager manages the state of all keepers
type StateManager struct {
	keepers     map[string]*types.KeeperInfo
	mu          sync.RWMutex
	logger      observability.Logger
	tracer      observability.Tracer
	initialized bool
	db          *client.DatabaseManager
}

var (
	stateManager     *StateManager
	stateManagerOnce sync.Once
)

// InitializeStateManager creates and initializes the state manager
func InitializeStateManager(ctx context.Context, logger observability.Logger, tracer observability.Tracer) *StateManager {
	stateManagerOnce.Do(func() {
		// Create a new logger with component field and proper level
		stateLogger := logger.With(observability.String("component", "state_manager"))

		stateManager = &StateManager{
			keepers:     make(map[string]*types.KeeperInfo),
			logger:      stateLogger,
			tracer:      tracer,
			initialized: true,
			db:          client.GetInstance(),
		}
		go stateManager.startCleanupRoutine(ctx)
	})
	return stateManager
}

// GetStateManager returns the singleton instance of StateManager
func GetStateManager() *StateManager {
	if stateManager == nil {
		panic("state manager not initialized")
	}
	return stateManager
}

// IsKeeperActive checks if a keeper is currently active
func (sm *StateManager) IsKeeperActive(ctx context.Context, keeperAddress string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	state, exists := sm.keepers[keeperAddress]
	isActive := exists && state.IsActive

	// sm.logger.Debug(ctx, "Checked keeper active status",
	// 	observability.String("keeper", keeperAddress),
	// 	observability.Bool("exists", exists),
	// 	observability.Bool("is_active", isActive),
	// )

	return isActive
}

// GetAllActiveKeepers returns a list of all active keeper addresses
func (sm *StateManager) GetAllActiveKeepers(ctx context.Context) []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var activeKeepers []string
	for address, state := range sm.keepers {
		if state.IsActive {
			activeKeepers = append(activeKeepers, address)
		}
	}

	// sm.logger.Debug(ctx, "Retrieved active keepers list",
	// 	observability.Int("total_active", len(activeKeepers)),
	// )

	return activeKeepers
}

// GetKeeperCount returns the total number of keepers and active keepers
func (sm *StateManager) GetKeeperCount(ctx context.Context) (total int, active int) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	total = len(sm.keepers)
	for _, state := range sm.keepers {
		if state.IsActive {
			active++
		}
	}

	// sm.logger.Debug(ctx, "Retrieved keeper counts",
	// 	observability.Int("total", total),
	// 	observability.Int("active", active),
	// )

	return total, active
}

// GetKeepersByVersion returns a map of version to count of active keepers
func (sm *StateManager) GetKeepersByVersion(ctx context.Context) map[string]int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	keepersByVersion := make(map[string]int)
	for _, state := range sm.keepers {
		if state.IsActive && state.Version != "" {
			keepersByVersion[state.Version]++
		}
	}

	return keepersByVersion
}

// GetDetailedKeeperInfo returns detailed information about all keepers
func (sm *StateManager) GetDetailedKeeperInfo(ctx context.Context) []types.KeeperInfo {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var keeperInfoList []types.KeeperInfo

	for address, state := range sm.keepers {
		info := types.KeeperInfo{
			KeeperName:       state.KeeperName,
			KeeperAddress:    address,
			ConsensusAddress: state.ConsensusAddress,
			OperatorID:       state.OperatorID,
			Version:          state.Version,
			PeerID:           state.PeerID,
			LastCheckedIn:    state.LastCheckedIn,
			IsActive:         state.IsActive,
			Network:          state.Network,
		}
		keeperInfoList = append(keeperInfoList, info)
	}

	// sm.logger.Debug(ctx, "Retrieved detailed keeper information",
	// 	observability.Int("total_keepers", len(keeperInfoList)),
	// )

	return keeperInfoList
}

// GetKeeperUptimes retrieves uptime for all keepers from the database
func (sm *StateManager) GetKeeperUptimes(ctx context.Context) (map[string]int64, error) {
	return sm.db.GetKeeperUptimes(ctx)
}
