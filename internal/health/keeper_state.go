package health

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/health/config"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

const (
	// InactivityThreshold defines how long a keeper can be without a check-in before being marked inactive
	InactivityThreshold = 70 * time.Second

	// StateCleanupInterval defines how often to check for inactive keepers
	StateCleanupInterval = 5 * time.Second
)

// KeeperState represents the current state of a keeper
type KeeperState struct {
	Health      types.KeeperHealth
	IsActive    bool
	LastUpdated time.Time
}

// KeeperStateManager manages the state of all keepers
type KeeperStateManager struct {
	keepers     map[string]*KeeperState
	mu          sync.RWMutex
	initialized bool
}

// Global instance of the state manager
var stateManager *KeeperStateManager
var stateManagerOnce sync.Once

// GetKeeperStateManager returns the singleton instance of KeeperStateManager
func GetKeeperStateManager() *KeeperStateManager {
	stateManagerOnce.Do(func() {
		stateManager = &KeeperStateManager{
			keepers:     make(map[string]*KeeperState),
			initialized: true,
		}
		// Start background cleanup routine
		go stateManager.startCleanupRoutine()
	})
	return stateManager
}

// startCleanupRoutine periodically checks for inactive keepers
func (ksm *KeeperStateManager) startCleanupRoutine() {
	ticker := time.NewTicker(StateCleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		ksm.checkInactiveKeepers()
	}
}

// checkInactiveKeepers identifies and processes inactive keepers
func (ksm *KeeperStateManager) checkInactiveKeepers() {
	now := time.Now()
	var inactiveKeepers []string

	ksm.mu.Lock()
	for address, state := range ksm.keepers {
		if state.IsActive && now.Sub(state.LastUpdated) > InactivityThreshold {
			logger.Infof("Keeper %s is now inactive (last seen: %s)", address, state.LastUpdated.Format(time.RFC3339))
			state.IsActive = false
			inactiveKeepers = append(inactiveKeepers, address)
		}
	}
	ksm.mu.Unlock()

	// Update database for inactive keepers
	for _, address := range inactiveKeepers {
		if err := ksm.updateKeeperStatusInDatabase(address, "", false); err != nil {
			logger.Errorf("Failed to update inactive status for keeper %s: %v", address, err)
		}
	}
}

// UpdateKeeperHealth updates the state for a keeper based on a health check-in
func (ksm *KeeperStateManager) UpdateKeeperHealth(health types.KeeperHealth) error {
	ksm.mu.Lock()
	defer ksm.mu.Unlock()

	address := health.KeeperAddress
	now := time.Now()

	// Check if keeper exists in our state
	existingState, exists := ksm.keepers[address]

	if !exists {
		// New keeper or previously inactive keeper
		ksm.keepers[address] = &KeeperState{
			Health:      health,
			IsActive:    true,
			LastUpdated: now,
		}

		// Update database to set keeper as active
		if err := ksm.updateKeeperStatusInDatabase(health.KeeperAddress, health.Version,true); err != nil {
			return fmt.Errorf("failed to update active status in database: %w", err)
		}

		logger.Infof("New keeper %s added to active state", address)
	} else {
		// Existing keeper - update health info
		wasActive := existingState.IsActive

		existingState.Health = health
		existingState.LastUpdated = now
		existingState.IsActive = true

		// If the keeper was inactive and is now active, update database
		if !wasActive {
			if err := ksm.updateKeeperStatusInDatabase(health.KeeperAddress, health.Version, true); err != nil {
				return fmt.Errorf("failed to update reactivated status in database: %w", err)
			}
			logger.Infof("Keeper %s reactivated", address)
		}
	}

	return nil
}

// updateKeeperStatusInDatabase calls the database API to update a keeper's active status
func (ksm *KeeperStateManager) updateKeeperStatusInDatabase(address string, version string, isActive bool) error {
	payload := types.UpdateKeeperHealth{
		KeeperAddress: address,
		Active:        isActive,
		Timestamp:     time.Now(),
		Version:       version,
	}
	
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal database update payload: %w", err)
	}

	// Construct database URL for status update
	databaseURL := fmt.Sprintf("%s/api/keepers/checkin", config.DatabaseIPAddress)

	// Send update to database
	response, err := http.Post(databaseURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to send status update to database: %w", err)
	}
	defer response.Body.Close()

	// Check response
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		return fmt.Errorf("database returned non-OK status %d: %s", response.StatusCode, string(body))
	}

	logger.Infof("Updated keeper %s status to %v in database", address, isActive)
	return nil
}

// IsKeeperActive checks if a keeper is currently active
func (ksm *KeeperStateManager) IsKeeperActive(keeperAddress string) bool {
	ksm.mu.RLock()
	defer ksm.mu.RUnlock()

	state, exists := ksm.keepers[keeperAddress]
	return exists && state.IsActive
}

// GetAllActiveKeepers returns a list of all active keepers
func (ksm *KeeperStateManager) GetAllActiveKeepers() []string {
	ksm.mu.RLock()
	defer ksm.mu.RUnlock()

	var activeKeepers []string
	for address, state := range ksm.keepers {
		if state.IsActive {
			activeKeepers = append(activeKeepers, address)
		}
	}

	return activeKeepers
}

// GetKeeperCount returns the count of all keepers and active keepers
func (ksm *KeeperStateManager) GetKeeperCount() (total int, active int) {
	ksm.mu.RLock()
	defer ksm.mu.RUnlock()

	total = len(ksm.keepers)
	for _, state := range ksm.keepers {
		if state.IsActive {
			active++
		}
	}

	return total, active
}
