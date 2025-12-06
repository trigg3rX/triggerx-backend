package registry

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/types"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// RegistryManager manages the registry of monitoring requests
type RegistryManager struct {
	registry map[string]*types.RegistryEntry
	mu       sync.RWMutex
	logger   logging.Logger
}

// NewRegistryManager creates a new registry manager
func NewRegistryManager(logger logging.Logger) *RegistryManager {
	rm := &RegistryManager{
		registry: make(map[string]*types.RegistryEntry),
		logger:   logger,
	}

	// Start background cleanup goroutine
	go rm.cleanupExpired()

	return rm
}

// generateRegistryKey generates a composite key for the registry
func generateRegistryKey(chainID string, contractAddr string, eventSig string) string {
	return fmt.Sprintf("%s:%s:%s", chainID, strings.ToLower(contractAddr), eventSig)
}

// Register registers a new monitoring request
func (rm *RegistryManager) Register(req *types.MonitoringRequest) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Validate contract address
	contractAddr := common.HexToAddress(req.ContractAddr)
	if contractAddr == (common.Address{}) {
		return fmt.Errorf("invalid contract address: %s", req.ContractAddr)
	}

	// Compute event signature hash
	eventSigHash := crypto.Keccak256Hash([]byte(req.EventSig))

	// Generate registry key
	key := generateRegistryKey(req.ChainID, req.ContractAddr, req.EventSig)

	// Check if entry exists
	entry, exists := rm.registry[key]
	if !exists {
		// Create new entry
		ctx, cancel := context.WithCancel(context.Background())
		entry = &types.RegistryEntry{
			Key:          key,
			ChainID:      req.ChainID,
			ContractAddr: contractAddr,
			EventSig:     eventSigHash,
			Subscribers:  make(map[string]*types.Subscriber),
			LastBlock:    0,
			WorkerCtx:    ctx,
			WorkerCancel: cancel,
		}
		rm.registry[key] = entry
		rm.logger.Info("Created new registry entry", "key", key, "chain_id", req.ChainID)
	}

	// Add subscriber
	entry.Mu.Lock()
	entry.Subscribers[req.RequestID] = &types.Subscriber{
		RequestID:   req.RequestID,
		WebhookURL:  req.WebhookURL,
		ExpiresAt:   req.ExpiresAt,
		FilterParam: req.FilterParam,
		FilterValue: req.FilterValue,
	}
	entry.Mu.Unlock()

	rm.logger.Info("Registered monitoring request",
		"request_id", req.RequestID,
		"key", key,
		"subscribers", len(entry.Subscribers))

	return nil
}

// Unregister unregisters a monitoring request
func (rm *RegistryManager) Unregister(requestID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Find the entry containing this request ID
	var foundEntry *types.RegistryEntry
	var foundKey string

	for key, entry := range rm.registry {
		entry.Mu.RLock()
		if _, exists := entry.Subscribers[requestID]; exists {
			foundEntry = entry
			foundKey = key
			entry.Mu.RUnlock()
			break
		}
		entry.Mu.RUnlock()
	}

	if foundEntry == nil {
		return fmt.Errorf("request ID not found: %s", requestID)
	}

	// Remove subscriber
	foundEntry.Mu.Lock()
	delete(foundEntry.Subscribers, requestID)
	subscriberCount := len(foundEntry.Subscribers)
	foundEntry.Mu.Unlock()

	rm.logger.Info("Unregistered monitoring request",
		"request_id", requestID,
		"key", foundKey,
		"remaining_subscribers", subscriberCount)

	// If no subscribers remain, stop worker and remove entry
	if subscriberCount == 0 {
		foundEntry.WorkerCancel()
		delete(rm.registry, foundKey)
		rm.logger.Info("Removed registry entry (no subscribers)", "key", foundKey)
	}

	return nil
}

// GetEntry returns a registry entry by key
func (rm *RegistryManager) GetEntry(key string) (*types.RegistryEntry, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	entry, exists := rm.registry[key]
	return entry, exists
}

// GetEntryByRequestID returns a registry entry by request ID
func (rm *RegistryManager) GetEntryByRequestID(requestID string) (*types.RegistryEntry, string, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	for key, entry := range rm.registry {
		entry.Mu.RLock()
		if _, exists := entry.Subscribers[requestID]; exists {
			entry.Mu.RUnlock()
			return entry, key, true
		}
		entry.Mu.RUnlock()
	}

	return nil, "", false
}

// GetAllEntries returns all registry entries
func (rm *RegistryManager) GetAllEntries() map[string]*types.RegistryEntry {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// Return a copy to avoid external modification
	result := make(map[string]*types.RegistryEntry)
	for k, v := range rm.registry {
		result[k] = v
	}
	return result
}

// GetActiveMonitorCount returns the number of active monitors
func (rm *RegistryManager) GetActiveMonitorCount() int {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	return len(rm.registry)
}

// GetChainsSupported returns the list of supported chains
func (rm *RegistryManager) GetChainsSupported() []string {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	chainMap := make(map[string]bool)
	for _, entry := range rm.registry {
		chainMap[entry.ChainID] = true
	}

	chains := make([]string, 0, len(chainMap))
	for chainID := range chainMap {
		chains = append(chains, chainID)
	}

	return chains
}

// cleanupExpired periodically cleans up expired requests
func (rm *RegistryManager) cleanupExpired() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		rm.mu.Lock()
		now := time.Now()

		for key, entry := range rm.registry {
			entry.Mu.Lock()
			expiredRequestIDs := make([]string, 0)

			for requestID, subscriber := range entry.Subscribers {
				if subscriber.ExpiresAt.Before(now) {
					expiredRequestIDs = append(expiredRequestIDs, requestID)
				}
			}

			// Remove expired subscribers
			for _, requestID := range expiredRequestIDs {
				delete(entry.Subscribers, requestID)
				rm.logger.Info("Removed expired subscriber",
					"request_id", requestID,
					"key", key)
			}

			subscriberCount := len(entry.Subscribers)
			entry.Mu.Unlock()

			// If no subscribers remain, stop worker and remove entry
			if subscriberCount == 0 {
				entry.WorkerCancel()
				delete(rm.registry, key)
				rm.logger.Info("Removed registry entry (expired)", "key", key)
			}
		}

		rm.mu.Unlock()
	}
}
