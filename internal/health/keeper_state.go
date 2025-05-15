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
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

const (
	inactivityThreshold  = 70 * time.Second
	stateCleanupInterval = 5 * time.Second
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
	logger      logging.Logger
	initialized bool
}

var (
	stateManager     *KeeperStateManager
	stateManagerOnce sync.Once
)

// InitializeStateManager creates and initializes the state manager
func InitializeStateManager(logger logging.Logger) *KeeperStateManager {
	stateManagerOnce.Do(func() {
		stateManager = &KeeperStateManager{
			keepers:     make(map[string]*KeeperState),
			logger:      logger,
			initialized: true,
		}
		go stateManager.startCleanupRoutine()
	})
	return stateManager
}

// GetKeeperStateManager returns the singleton instance of KeeperStateManager
func GetKeeperStateManager() *KeeperStateManager {
	if stateManager == nil {
		panic("state manager not initialized")
	}
	return stateManager
}

func (ksm *KeeperStateManager) startCleanupRoutine() {
	ticker := time.NewTicker(stateCleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		ksm.checkInactiveKeepers()
	}
}

func (ksm *KeeperStateManager) checkInactiveKeepers() {
	now := time.Now().UTC()
	var inactiveKeepers []string

	ksm.mu.Lock()
	for address, state := range ksm.keepers {
		if state.IsActive && now.Sub(state.LastUpdated) > inactivityThreshold {
			ksm.logger.Info("Keeper became inactive",
				"keeper", address,
				"lastSeen", state.LastUpdated.Format(time.RFC3339),
			)
			state.IsActive = false
			inactiveKeepers = append(inactiveKeepers, address)
		}
	}
	ksm.mu.Unlock()

	for _, address := range inactiveKeepers {
		if err := ksm.updateKeeperStatusInDatabase(address, "", "", false); err != nil {
			ksm.logger.Error("Failed to update inactive status",
				"error", err,
				"keeper", address,
			)
		}
	}
}

// UpdateKeeperHealth updates the health status of a keeper
func (ksm *KeeperStateManager) UpdateKeeperHealth(health types.KeeperHealth) error {
	ksm.mu.Lock()
	defer ksm.mu.Unlock()

	address := health.KeeperAddress
	now := time.Now().UTC()

	existingState, exists := ksm.keepers[address]

	if !exists {
		ksm.keepers[address] = &KeeperState{
			Health:      health,
			IsActive:    true,
			LastUpdated: now,
		}

		if err := ksm.updateKeeperStatusInDatabase(health.KeeperAddress, health.Version, health.PeerID, true); err != nil {
			return fmt.Errorf("failed to update active status in database: %w", err)
		}

		ksm.logger.Info("New keeper added",
			"keeper", address,
			"version", health.Version,
		)
		return nil
	}

	wasActive := existingState.IsActive
	existingState.Health = health
	existingState.LastUpdated = now
	existingState.IsActive = true

	if !wasActive {
		if err := ksm.updateKeeperStatusInDatabase(health.KeeperAddress, health.Version, health.PeerID, true); err != nil {
			return fmt.Errorf("failed to update reactivated status in database: %w", err)
		}
		ksm.logger.Info("Keeper reactivated",
			"keeper", address,
			"version", health.Version,
		)
	}

	return nil
}

func (ksm *KeeperStateManager) updateKeeperStatusInDatabase(address, version, peerID string, isActive bool) error {
	payload := types.UpdateKeeperHealth{
		KeeperAddress: address,
		Active:        isActive,
		Timestamp:     time.Now().UTC(),
		Version:       version,
		PeerID:        peerID,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal database update payload: %w", err)
	}

	databaseURL := fmt.Sprintf("%s/api/keepers/checkin", config.GetDatabaseRPCAddress())

	response, err := http.Post(databaseURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to send status update to database: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		return fmt.Errorf("database returned non-OK status %d: %s", response.StatusCode, string(body))
	}

	ksm.logger.Debug("Updated keeper status in database",
		"keeper", address,
		"active", isActive,
		"version", version,
	)
	return nil
}

// IsKeeperActive checks if a keeper is currently active
func (ksm *KeeperStateManager) IsKeeperActive(keeperAddress string) bool {
	ksm.mu.RLock()
	defer ksm.mu.RUnlock()

	state, exists := ksm.keepers[keeperAddress]
	return exists && state.IsActive
}

// GetAllActiveKeepers returns a list of all active keeper addresses
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

// GetKeeperCount returns the total number of keepers and active keepers
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

// KeeperInfo represents detailed information about a keeper
type KeeperInfo struct {
	Address     string    `json:"address"`
	IsActive    bool      `json:"is_active"`
	Version     string    `json:"version,omitempty"`
	PeerID      string    `json:"peer_id,omitempty"`
	LastUpdated time.Time `json:"last_updated"`
}

// GetDetailedKeeperInfo returns detailed information about all keepers
func (ksm *KeeperStateManager) GetDetailedKeeperInfo() []KeeperInfo {
	ksm.mu.RLock()
	defer ksm.mu.RUnlock()

	var keeperInfoList []KeeperInfo

	for address, state := range ksm.keepers {
		info := KeeperInfo{
			Address:     address,
			IsActive:    state.IsActive,
			Version:     state.Health.Version,
			PeerID:      state.Health.PeerID,
			LastUpdated: state.LastUpdated,
		}
		keeperInfoList = append(keeperInfoList, info)
	}

	return keeperInfoList
}
