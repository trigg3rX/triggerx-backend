package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/config"
	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/registry"
	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/types"
	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/webhook"
	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/worker"
	nodeclient "github.com/trigg3rX/triggerx-backend/pkg/client/nodeclient"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// Service manages the event monitor service
type Service struct {
	registryManager *registry.RegistryManager
	nodeClients     map[string]*nodeclient.NodeClient // chainID -> NodeClient
	workers         map[string]*worker.Worker         // registry key -> Worker
	webhookClient   *webhook.Client
	logger          logging.Logger
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
}

// NewService creates a new event monitor service
func NewService(logger logging.Logger) (*Service, error) {
	ctx, cancel := context.WithCancel(context.Background())

	rm := registry.NewRegistryManager(logger)
	wc := webhook.NewClient(logger)

	// Initialize node clients for supported chains
	nodeClients := make(map[string]*nodeclient.NodeClient)
	chainRPCUrls := config.GetChainRPCUrls()

	for chainID, rpcURL := range chainRPCUrls {
		// Determine network from chain ID
		var network nodeclient.Network
		switch chainID {
		case "11155111":
			network = nodeclient.NetworkEthereumSepolia
		case "84532":
			network = nodeclient.NetworkBaseSepolia
		case "11155420":
			network = nodeclient.NetworkOptimismSepolia
		case "421614":
			network = nodeclient.NetworkArbitrumSepolia
		default:
			logger.Warn("Unknown chain ID, using custom URL", "chain_id", chainID)
			// Create custom config with base URL
			cfg := nodeclient.DefaultConfig(config.GetAlchemyAPIKey(), "", logger)
			cfg.BaseURL = rpcURL
			client, err := nodeclient.NewNodeClient(cfg)
			if err != nil {
				logger.Error("Failed to create node client", "chain_id", chainID, "error", err)
				continue
			}
			nodeClients[chainID] = client
			continue
		}

		// Create node client with Alchemy config
		cfg := nodeclient.DefaultConfig(config.GetAlchemyAPIKey(), network, logger)
		if config.GetAlchemyAPIKey() == "" {
			// Use custom URL if no API key
			cfg.BaseURL = rpcURL
		}
		client, err := nodeclient.NewNodeClient(cfg)
		if err != nil {
			logger.Error("Failed to create node client", "chain_id", chainID, "error", err)
			continue
		}
		nodeClients[chainID] = client
		logger.Info("Initialized node client", "chain_id", chainID)
	}

	return &Service{
		registryManager: rm,
		nodeClients:     nodeClients,
		workers:         make(map[string]*worker.Worker),
		webhookClient:   wc,
		logger:          logger,
		ctx:             ctx,
		cancel:          cancel,
	}, nil
}

// Start starts the service
func (s *Service) Start() error {
	s.logger.Info("Starting event monitor service")

	// Start monitoring registry changes
	go s.monitorRegistry()

	return nil
}

// Stop stops the service
func (s *Service) Stop() {
	s.logger.Info("Stopping event monitor service")

	// Cancel context
	s.cancel()

	// Stop all workers
	s.mu.Lock()
	for key, w := range s.workers {
		w.Stop()
		delete(s.workers, key)
	}
	s.mu.Unlock()

	// Wait for all goroutines to finish
	s.wg.Wait()

	// Close node clients
	for chainID, client := range s.nodeClients {
		client.Close()
		s.logger.Info("Closed node client", "chain_id", chainID)
	}

	s.logger.Info("Event monitor service stopped")
}

// Register registers a monitoring request and starts a worker if needed
func (s *Service) Register(req *types.MonitoringRequest) error {
	// Register in registry
	if err := s.registryManager.Register(req); err != nil {
		return err
	}

	// Generate registry key
	key := s.generateRegistryKey(req.ChainID, req.ContractAddr, req.EventSig)

	// Check if worker already exists
	s.mu.Lock()
	_, exists := s.workers[key]
	s.mu.Unlock()

	if !exists {
		// Start new worker
		if err := s.startWorker(key); err != nil {
			// Rollback registration
			_ = s.registryManager.Unregister(req.RequestID)
			return fmt.Errorf("failed to start worker: %w", err)
		}
	}

	return nil
}

// Unregister unregisters a monitoring request and stops worker if needed
func (s *Service) Unregister(requestID string) error {
	// Get entry to check if we need to stop worker
	entry, key, exists := s.registryManager.GetEntryByRequestID(requestID)
	if !exists {
		return fmt.Errorf("request ID not found: %s", requestID)
	}

	// Unregister from registry
	if err := s.registryManager.Unregister(requestID); err != nil {
		return err
	}

	// Check if worker should be stopped (no more subscribers)
	entry.Mu.RLock()
	subscriberCount := len(entry.Subscribers)
	entry.Mu.RUnlock()

	if subscriberCount == 0 {
		// Stop worker
		s.mu.Lock()
		if w, exists := s.workers[key]; exists {
			w.Stop()
			delete(s.workers, key)
			s.logger.Info("Stopped worker", "key", key)
		}
		s.mu.Unlock()
	}

	return nil
}

// GetRegistryManager returns the registry manager
func (s *Service) GetRegistryManager() *registry.RegistryManager {
	return s.registryManager
}

// startWorker starts a worker for a registry entry
func (s *Service) startWorker(key string) error {
	entry, exists := s.registryManager.GetEntry(key)
	if !exists {
		return fmt.Errorf("registry entry not found: %s", key)
	}

	// Get node client for chain
	nodeClient, exists := s.nodeClients[entry.ChainID]
	if !exists {
		return fmt.Errorf("node client not found for chain: %s", entry.ChainID)
	}

	// Create worker
	w := worker.NewWorker(entry, nodeClient, s.webhookClient, s.logger)

	s.mu.Lock()
	s.workers[key] = w
	s.mu.Unlock()

	// Start worker in goroutine
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		w.Start()
	}()

	s.logger.Info("Started worker", "key", key, "chain_id", entry.ChainID)
	return nil
}

// monitorRegistry monitors registry changes and starts/stops workers as needed
func (s *Service) monitorRegistry() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.syncWorkers()
		}
	}
}

// syncWorkers syncs workers with registry entries
func (s *Service) syncWorkers() {
	entries := s.registryManager.GetAllEntries()

	s.mu.Lock()
	defer s.mu.Unlock()

	// Start workers for entries that don't have workers
	for key, entry := range entries {
		if _, exists := s.workers[key]; !exists {
			// Check if node client exists
			if _, exists := s.nodeClients[entry.ChainID]; exists {
				// Start worker
				w := worker.NewWorker(entry, s.nodeClients[entry.ChainID], s.webhookClient, s.logger)
				s.workers[key] = w
				s.wg.Add(1)
				go func(workerKey string, worker *worker.Worker) {
					defer s.wg.Done()
					worker.Start()
				}(key, w)
				s.logger.Info("Started worker (sync)", "key", key)
			}
		}
	}

	// Stop workers for entries that no longer exist
	for key, w := range s.workers {
		if _, exists := entries[key]; !exists {
			w.Stop()
			delete(s.workers, key)
			s.logger.Info("Stopped worker (sync)", "key", key)
		}
	}
}

// generateRegistryKey generates a registry key
func (s *Service) generateRegistryKey(chainID string, contractAddr string, eventSig string) string {
	return fmt.Sprintf("%s:%s:%s", chainID, strings.ToLower(contractAddr), eventSig)
}
