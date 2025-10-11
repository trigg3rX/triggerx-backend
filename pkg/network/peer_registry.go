package network

// import (
// 	"encoding/json"
// 	"fmt"
// 	"os"
// 	"path/filepath"
// 	"sync"

// 	"github.com/libp2p/go-libp2p/core/peer"
// )

// const (
// 	ServiceManager   = "manager"
// 	ServiceQuorum    = "quorum"
// 	ServiceValidator = "validator"
// 	ServiceKeeper    = "keeper"

// 	// Define base data directory
// 	BaseDataDir     = "data"
// 	RegistryDir     = "peer_registry"
// 	ServiceRegistry = "services.json"
// )

// type ServiceInfo struct {
// 	Name      string   `json:"name"`
// 	PeerID    string   `json:"peer_id"`
// 	Addresses []string `json:"addresses"`
// }

// type PeerRegistry struct {
// 	Services map[string]ServiceInfo `json:"services"`
// 	path     string
// 	mu       sync.RWMutex
// }

// func NewPeerRegistry() (*PeerRegistry, error) {
// 	registryDir := filepath.Join(BaseDataDir, RegistryDir)
// 	if err := os.MkdirAll(registryDir, 0755); err != nil {
// 		return nil, fmt.Errorf("failed to create registry directory: %w", err)
// 	}

// 	registry := &PeerRegistry{
// 		Services: make(map[string]ServiceInfo),
// 		path:     filepath.Join(registryDir, ServiceRegistry),
// 	}

// 	// Load existing registry if it exists
// 	if err := registry.load(); err != nil {
// 		if !os.IsNotExist(err) {
// 			return nil, fmt.Errorf("failed to load registry: %w", err)
// 		}
// 		// Initialize with default services with empty peer IDs
// 		registry.Services = map[string]ServiceInfo{
// 			ServiceManager:   {Name: ServiceManager, PeerID: "", Addresses: nil},
// 			ServiceQuorum:    {Name: ServiceQuorum, PeerID: "", Addresses: nil},
// 			ServiceValidator: {Name: ServiceValidator, PeerID: "", Addresses: nil},
// 		}
// 		// Save initial registry
// 		if err := registry.save(); err != nil {
// 			return nil, fmt.Errorf("failed to save initial registry: %w", err)
// 		}
// 	}

// 	return registry, nil
// }

// func (r *PeerRegistry) load() error {
// 	r.mu.Lock()
// 	defer r.mu.Unlock()

// 	data, err := os.ReadFile(r.path)
// 	if err != nil {
// 		return err
// 	}

// 	return json.Unmarshal(data, &r.Services)
// }

// func (r *PeerRegistry) save() error {
// 	data, err := json.MarshalIndent(r.Services, "", "  ")
// 	if err != nil {
// 		return fmt.Errorf("failed to marshal registry: %w", err)
// 	}

// 	if err := os.WriteFile(r.path, data, 0644); err != nil {
// 		return fmt.Errorf("failed to write registry file: %w", err)
// 	}

// 	return nil
// }

// func (r *PeerRegistry) UpdateService(serviceName string, peerID peer.ID, addrs []string) error {
// 	r.mu.Lock()
// 	defer r.mu.Unlock()

// 	r.Services[serviceName] = ServiceInfo{
// 		Name:      serviceName,
// 		PeerID:    peerID.String(),
// 		Addresses: addrs,
// 	}

// 	err := r.save()
// 	if err != nil {
// 		return fmt.Errorf("failed to save registry after update: %w", err)
// 	}

// 	return nil
// }

// func (r *PeerRegistry) GetService(serviceName string) (ServiceInfo, bool) {
// 	r.mu.RLock()
// 	defer r.mu.RUnlock()

// 	service, exists := r.Services[serviceName]
// 	return service, exists
// }

// func (r *PeerRegistry) GetAllServices() map[string]ServiceInfo {
// 	r.mu.RLock()
// 	defer r.mu.RUnlock()

// 	// Create a copy to avoid concurrent access issues
// 	services := make(map[string]ServiceInfo)
// 	for k, v := range r.Services {
// 		services[k] = v
// 	}
// 	return services
// }
