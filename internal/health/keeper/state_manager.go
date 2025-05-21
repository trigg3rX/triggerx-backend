package keeper

import (
	"sync"

	"github.com/trigg3rX/triggerx-backend/internal/health/client"
	"github.com/trigg3rX/triggerx-backend/internal/health/types"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// StateManager manages the state of all keepers
type StateManager struct {
	keepers     map[string]*types.KeeperInfo
	mu          sync.RWMutex
	logger      logging.Logger
	initialized bool
	db          *client.DatabaseManager
}

var (
	stateManager     *StateManager
	stateManagerOnce sync.Once
)

// InitializeStateManager creates and initializes the state manager
func InitializeStateManager(logger logging.Logger) *StateManager {
	stateManagerOnce.Do(func() {
		// Create a new logger with component field and proper level
		stateLogger := logger.With("component", "state_manager")

		stateManager = &StateManager{
			keepers:     make(map[string]*types.KeeperInfo),
			logger:      stateLogger,
			initialized: true,
			db:          client.GetInstance(),
		}
		go stateManager.startCleanupRoutine()
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
func (sm *StateManager) IsKeeperActive(keeperAddress string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	state, exists := sm.keepers[keeperAddress]
	isActive := exists && state.IsActive

	sm.logger.Debug("Checked keeper active status",
		"keeper", keeperAddress,
		"exists", exists,
		"is_active", isActive,
	)

	return isActive
}

// GetAllActiveKeepers returns a list of all active keeper addresses
func (sm *StateManager) GetAllActiveKeepers() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var activeKeepers []string
	for address, state := range sm.keepers {
		if state.IsActive {
			activeKeepers = append(activeKeepers, address)
		}
	}

	sm.logger.Debug("Retrieved active keepers list",
		"total_active", len(activeKeepers),
	)

	return activeKeepers
}

// GetKeeperCount returns the total number of keepers and active keepers
func (sm *StateManager) GetKeeperCount() (total int, active int) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	total = len(sm.keepers)
	for _, state := range sm.keepers {
		if state.IsActive {
			active++
		}
	}

	sm.logger.Debug("Retrieved keeper counts",
		"total", total,
		"active", active,
	)

	return total, active
}

// GetDetailedKeeperInfo returns detailed information about all keepers
func (sm *StateManager) GetDetailedKeeperInfo() []types.KeeperInfo {
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
		}
		keeperInfoList = append(keeperInfoList, info)
	}

	sm.logger.Debug("Retrieved detailed keeper information",
		"total_keepers", len(keeperInfoList),
	)

	return keeperInfoList
}
