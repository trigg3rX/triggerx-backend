package keeper

import (
	"sync"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/health/interfaces"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// StateManager manages the state of all keepers
type StateManager struct {
	keepers           map[string]*types.HealthKeeperInfo
	mu                sync.RWMutex
	logger            logging.Logger
	db                interfaces.DatabaseManagerInterface
	notifier          interfaces.NotificationBotInterface
	notificationsSent map[string]time.Time // Tracks when notifications were last sent
}

var (
	stateManager     *StateManager
	stateManagerOnce sync.Once
)

// InitializeStateManager creates and initializes the state manager
func InitializeStateManager(
	logger logging.Logger,
	db interfaces.DatabaseManagerInterface,
	notifier interfaces.NotificationBotInterface,
) *StateManager {
	stateManagerOnce.Do(func() {
		// Create a new logger with component field and proper level
		stateLogger := logger.With("component", "state_manager")

		stateManager = &StateManager{
			keepers:           make(map[string]*types.HealthKeeperInfo),
			logger:            stateLogger,
			db:                db,
			notifier:          notifier,
			notificationsSent: make(map[string]time.Time),
		}
		go stateManager.startCleanupRoutine()
		go stateManager.startPeriodicDumpRoutine()
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
	isActive := exists && state != nil && state.IsActive

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
		if state != nil && state.IsActive {
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
		if state != nil && state.IsActive {
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
func (sm *StateManager) GetDetailedKeeperInfo() []types.HealthKeeperInfo {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var keeperInfoList []types.HealthKeeperInfo

	for address, state := range sm.keepers {
		if state != nil {
			info := types.HealthKeeperInfo{
				KeeperName:       state.KeeperName,
				KeeperAddress:    address,
				ConsensusAddress: state.ConsensusAddress,
				OperatorID:       state.OperatorID,
				Version:          state.Version,
				LastCheckedIn:    state.LastCheckedIn,
				IsActive:         state.IsActive,
				IsImua:           state.IsImua,
				Uptime:           state.Uptime,
			}
			keeperInfoList = append(keeperInfoList, info)
		}
	}

	sm.logger.Debug("Retrieved detailed keeper information",
		"total_keepers", len(keeperInfoList),
	)

	return keeperInfoList
}
