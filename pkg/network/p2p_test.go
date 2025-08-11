package network

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestP2PConfig tests the P2P configuration validation
func TestP2PConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      P2PConfig
		shouldError bool
		errorMsg    string
	}{
		{
			name: "valid bootstrap config",
			config: P2PConfig{
				NodeType: NodeTypeBootstrap,
				BootstrapConfig: &BootstrapConfig{
					ServiceName:   ServiceManager,
					ListenAddress: "127.0.0.1",
					ListenPort:    "9000",
				},
			},
			shouldError: false,
		},
		{
			name: "valid keeper config",
			config: P2PConfig{
				NodeType: NodeTypeKeeper,
				KeeperConfig: &KeeperConfig{
					ConnectionAddress: "127.0.0.1",
					P2pPort:           "9012",
				},
			},
			shouldError: false,
		},
		{
			name: "bootstrap config missing",
			config: P2PConfig{
				NodeType:        NodeTypeBootstrap,
				BootstrapConfig: nil,
			},
			shouldError: true,
			errorMsg:    "bootstrap config is required for bootstrap node",
		},
		{
			name: "keeper config missing",
			config: P2PConfig{
				NodeType:     NodeTypeKeeper,
				KeeperConfig: nil,
			},
			shouldError: true,
			errorMsg:    "keeper config is required for keeper node",
		},
		{
			name: "invalid node type",
			config: P2PConfig{
				NodeType: P2PNodeType("invalid"),
			},
			shouldError: true,
			errorMsg:    "unsupported node type: invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			registry := createTestRegistry(t)
			defer cleanupTestRegistry(t, registry)

			// Create test private keys if needed
			if !tt.shouldError {
				createTestPrivateKeys(t)
			}

			node, err := NewP2PNode(ctx, tt.config, registry)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if node == nil {
				t.Error("Expected node to be created")
				return
			}

			// Verify node properties
			if node.GetNodeType() != tt.config.NodeType {
				t.Errorf("Expected node type %s, got %s", tt.config.NodeType, node.GetNodeType())
			}

			if node.GetHost() == nil {
				t.Error("Expected host to be initialized")
			}

			if node.GetLogger() == nil {
				t.Error("Expected logger to be initialized")
			}

			// Clean up
			if err := node.Close(); err != nil {
				t.Errorf("Error closing node: %v", err)
			}
		})
	}
}

// TestP2PNodeLifecycle tests the complete lifecycle of a P2P node
func TestP2PNodeLifecycle(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	registry := createTestRegistry(t)
	defer cleanupTestRegistry(t, registry)
	createTestPrivateKeys(t)

	// Test bootstrap node lifecycle
	t.Run("bootstrap_node_lifecycle", func(t *testing.T) {
		config := P2PConfig{
			NodeType: NodeTypeBootstrap,
			BootstrapConfig: &BootstrapConfig{
				ServiceName:   ServiceManager,
				ListenAddress: "127.0.0.1",
				ListenPort:    "9001",
			},
		}

		// Create node
		node, err := NewP2PNode(ctx, config, registry)
		if err != nil {
			t.Fatalf("Failed to create bootstrap node: %v", err)
		}

		// Verify node is running
		host := node.GetHost()
		if host == nil {
			t.Fatal("Host should not be nil")
		}

		peerID := host.ID()
		if peerID == "" {
			t.Fatal("Peer ID should not be empty")
		}

		// Verify logger
		logger := node.GetLogger()
		if logger == nil {
			t.Fatal("Logger should not be nil")
		}

		// Test registry update
		serviceInfo, exists := registry.GetService(ServiceManager)
		if !exists {
			t.Fatal("Service should be registered")
		}

		if serviceInfo.PeerID != peerID.String() {
			t.Errorf("Expected peer ID %s, got %s", peerID.String(), serviceInfo.PeerID)
		}

		// Close node
		if err := node.Close(); err != nil {
			t.Errorf("Error closing node: %v", err)
		}
	})

	// Test keeper node lifecycle
	t.Run("keeper_node_lifecycle", func(t *testing.T) {
		config := P2PConfig{
			NodeType: NodeTypeKeeper,
			KeeperConfig: &KeeperConfig{
				ConnectionAddress: "127.0.0.1",
				P2pPort:           "9013",
			},
		}

		// Create node
		node, err := NewP2PNode(ctx, config, registry)
		if err != nil {
			t.Fatalf("Failed to create keeper node: %v", err)
		}

		// Verify node is running
		host := node.GetHost()
		if host == nil {
			t.Fatal("Host should not be nil")
		}

		// Close node
		if err := node.Close(); err != nil {
			t.Errorf("Error closing node: %v", err)
		}
	})
}

// TestLoadIdentity tests the identity loading functionality
func TestLoadIdentity(t *testing.T) {
	createTestPrivateKeys(t)

	tests := []struct {
		name     string
		nodeType string
		wantErr  bool
	}{
		{
			name:     "manager identity",
			nodeType: ServiceManager,
			wantErr:  false,
		},
		{
			name:     "quorum identity",
			nodeType: ServiceQuorum,
			wantErr:  false,
		},
		{
			name:     "validator identity",
			nodeType: ServiceValidator,
			wantErr:  false,
		},
		{
			name:     "keeper identity",
			nodeType: ServiceKeeper,
			wantErr:  false,
		},
		{
			name:     "invalid identity",
			nodeType: "invalid",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			priv, err := LoadIdentity(tt.nodeType)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if priv == nil {
				t.Error("Expected private key to be loaded")
			}
		})
	}
}

// TestLegacyFunctions tests backward compatibility functions
func TestLegacyFunctions(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	registry := createTestRegistry(t)
	defer cleanupTestRegistry(t, registry)
	createTestPrivateKeys(t)

	t.Run("SetupServiceWithRegistry", func(t *testing.T) {
		// First add a service to registry
		serviceName := ServiceManager

		// Create a temporary node to get addresses
		tempConfig := P2PConfig{
			NodeType: NodeTypeBootstrap,
			BootstrapConfig: &BootstrapConfig{
				ServiceName:   serviceName,
				ListenAddress: "127.0.0.1",
				ListenPort:    "9002",
			},
		}

		tempNode, err := NewP2PNode(ctx, tempConfig, registry)
		if err != nil {
			t.Fatalf("Failed to create temp node: %v", err)
		}
		err = tempNode.Close()
		if err != nil {
			t.Fatalf("Failed to close temp node: %v", err)
		}

		// Now test legacy function
		host, err := SetupServiceWithRegistry(ctx, serviceName, registry)
		if err != nil {
			t.Fatalf("SetupServiceWithRegistry failed: %v", err)
		}

		if host == nil {
			t.Fatal("Expected host to be created")
		}

		err = host.Close()
		if err != nil {
			t.Fatalf("Failed to close host: %v", err)
		}
	})
}

// Helper functions for testing

func createTestRegistry(t *testing.T) *PeerRegistry {
	// Create test data directory
	testDir := filepath.Join("testdata", "test_registry")
	err := os.RemoveAll(testDir) // Clean up any existing test data
	if err != nil {
		t.Fatalf("Failed to remove test registry directory: %v", err)
	}

	// Create a temporary registry with custom data directory
	registry := &PeerRegistry{
		Services: make(map[string]ServiceInfo),
		path:     filepath.Join(testDir, RegistryDir, ServiceRegistry),
	}

	// Create the directory structure
	if err := os.MkdirAll(filepath.Join(testDir, RegistryDir), 0755); err != nil {
		t.Fatalf("Failed to create test registry directory: %v", err)
	}

	// Initialize with default services
	registry.Services = map[string]ServiceInfo{
		ServiceManager:   {Name: ServiceManager, PeerID: "", Addresses: nil},
		ServiceQuorum:    {Name: ServiceQuorum, PeerID: "", Addresses: nil},
		ServiceValidator: {Name: ServiceValidator, PeerID: "", Addresses: nil},
	}

	// Save initial registry
	if err := registry.save(); err != nil {
		t.Fatalf("Failed to save initial test registry: %v", err)
	}

	t.Cleanup(func() {
		err := os.RemoveAll(testDir)
		if err != nil {
			t.Fatalf("Failed to remove test registry directory: %v", err)
		}
	})

	return registry
}

func cleanupTestRegistry(t *testing.T, registry *PeerRegistry) {
	// Registry cleanup is handled by the t.Cleanup in createTestRegistry
}

func createTestPrivateKeys(t *testing.T) {
	// Create test data directory structure in the actual data directory
	// that LoadIdentity expects (BaseDataDir/RegistryDir)
	registryDir := filepath.Join(BaseDataDir, RegistryDir)
	if err := os.MkdirAll(registryDir, 0755); err != nil {
		t.Fatalf("Failed to create registry directory: %v", err)
	}

	// Create test service private keys with unique keys for each service
	services := ServicePrivateKeys{
		ManagerP2PPrivateKey:   "CAESQKBj/Dv/gPbs33isOeRfr1vGmgXOg6mUhpPaDgRUrSgvAR6bgGhQ4dKhMT/2cOWqJe9MhEGwVoNW2pqJzB4Ij/0=",
		QuorumP2PPrivateKey:    "CAESQKnudgGSiPS1W2VsNLePz8I6mGNdCu7uJzGpPGGtbPHdCKUlLlQe7o1MYkYqPGS4VdVKAuIhIq7KnMfaUwZq+Rw=",
		ValidatorP2PPrivateKey: "CAESQLc0mJOOzQwyXWvgvEw1I5CHt7xMejQRz1EqPfyNRBsGjr4DsWzxQS4LFhZg3cNB8O9XYcOdLyKfGkOGzgMr5zU=",
	}

	serviceKeysPath := filepath.Join(registryDir, "services_privKeys.json")
	serviceKeysData, err := json.Marshal(services)
	if err != nil {
		t.Fatalf("Failed to marshal service private keys: %v", err)
	}

	if err := os.WriteFile(serviceKeysPath, serviceKeysData, 0644); err != nil {
		t.Fatalf("Failed to write service private keys: %v", err)
	}

	// Create test keeper identity
	keeperIdentity := PeerIdentity{
		PrivKey: []byte{8, 1, 18, 64, 169, 238, 118, 1, 146, 136, 244, 181, 91, 101, 108, 52, 183, 143, 207, 194, 58, 152, 99, 93, 10, 238, 238, 39, 49, 169, 60, 97, 173, 108, 241, 221, 8, 165, 37, 46, 84, 30, 238, 141, 76, 98, 70, 42, 60, 100, 184, 85, 213, 74, 2, 226, 33, 34, 174, 202, 156, 199, 218, 83, 6, 106, 249, 28},
	}

	keeperIdentityPath := filepath.Join(registryDir, "keeper_identity.json")
	keeperIdentityData, err := json.Marshal(keeperIdentity)
	if err != nil {
		t.Fatalf("Failed to marshal keeper identity: %v", err)
	}

	if err := os.WriteFile(keeperIdentityPath, keeperIdentityData, 0644); err != nil {
		t.Fatalf("Failed to write keeper identity: %v", err)
	}

	// Schedule cleanup
	t.Cleanup(func() {
		err := os.RemoveAll(serviceKeysPath)
		if err != nil {
			t.Fatalf("Failed to remove service keys: %v", err)
		}
		err = os.RemoveAll(keeperIdentityPath)
		if err != nil {
			t.Fatalf("Failed to remove keeper identity: %v", err)
		}
	})
}

// Benchmark tests
func BenchmarkNewP2PNode(b *testing.B) {
	ctx := context.Background()
	registry := createTestRegistryForBench(b)
	createTestPrivateKeysForBench(b)

	config := P2PConfig{
		NodeType: NodeTypeBootstrap,
		BootstrapConfig: &BootstrapConfig{
			ServiceName:   ServiceManager,
			ListenAddress: "127.0.0.1",
			ListenPort:    "9000",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		node, err := NewP2PNode(ctx, config, registry)
		if err != nil {
			b.Fatalf("Failed to create P2P node: %v", err)
		}
		err = node.Close()
		if err != nil {
			b.Fatalf("Failed to close P2P node: %v", err)
		}
	}
}

func createTestRegistryForBench(b *testing.B) *PeerRegistry {
	testDir := filepath.Join("testdata", "bench_registry")
	err := os.RemoveAll(testDir)
	if err != nil {
		b.Fatalf("Failed to remove test registry directory: %v", err)
	}

	// Create a temporary registry with custom data directory
	registry := &PeerRegistry{
		Services: make(map[string]ServiceInfo),
		path:     filepath.Join(testDir, RegistryDir, ServiceRegistry),
	}

	// Create the directory structure
	if err := os.MkdirAll(filepath.Join(testDir, RegistryDir), 0755); err != nil {
		b.Fatalf("Failed to create test registry directory: %v", err)
	}

	// Initialize with default services
	registry.Services = map[string]ServiceInfo{
		ServiceManager:   {Name: ServiceManager, PeerID: "", Addresses: nil},
		ServiceQuorum:    {Name: ServiceQuorum, PeerID: "", Addresses: nil},
		ServiceValidator: {Name: ServiceValidator, PeerID: "", Addresses: nil},
	}

	// Save initial registry
	if err := registry.save(); err != nil {
		b.Fatalf("Failed to save initial test registry: %v", err)
	}

	b.Cleanup(func() {
		err := os.RemoveAll(testDir)
		if err != nil {
			b.Fatalf("Failed to remove test registry directory: %v", err)
		}
	})

	return registry
}

func createTestPrivateKeysForBench(b *testing.B) {
	testDir := filepath.Join("testdata", "bench_registry")
	registryDir := filepath.Join(testDir, RegistryDir)
	if err := os.MkdirAll(registryDir, 0755); err != nil {
		b.Fatalf("Failed to create registry directory: %v", err)
	}

	services := ServicePrivateKeys{
		ManagerP2PPrivateKey:   "CAESQKBj/Dv/gPbs33isOeRfr1vGmgXOg6mUhpPaDgRUrSgvAR6bgGhQ4dKhMT/2cOWqJe9MhEGwVoNW2pqJzB4Ij/0=",
		QuorumP2PPrivateKey:    "CAESQKnudgGSiPS1W2VsNLePz8I6mGNdCu7uJzGpPGGtbPHdCKUlLlQe7o1MYkYqPGS4VdVKAuIhIq7KnMfaUwZq+Rw=",
		ValidatorP2PPrivateKey: "CAESQLc0mJOOzQwyXWvgvEw1I5CHt7xMejQRz1EqPfyNRBsGjr4DsWzxQS4LFhZg3cNB8O9XYcOdLyKfGkOGzgMr5zU=",
	}

	serviceKeysPath := filepath.Join(registryDir, "services_privKeys.json")
	serviceKeysData, _ := json.Marshal(services)
	err := os.WriteFile(serviceKeysPath, serviceKeysData, 0644)
	if err != nil {
		b.Fatalf("Failed to write service keys: %v", err)
	}

	keeperIdentity := PeerIdentity{
		PrivKey: []byte{8, 1, 18, 64, 169, 238, 118, 1, 146, 136, 244, 181, 91, 101, 108, 52, 183, 143, 207, 194, 58, 152, 99, 93, 10, 238, 238, 39, 49, 169, 60, 97, 173, 108, 241, 221, 8, 165, 37, 46, 84, 30, 238, 141, 76, 98, 70, 42, 60, 100, 184, 85, 213, 74, 2, 226, 33, 34, 174, 202, 156, 199, 218, 83, 6, 106, 249, 28},
	}

	keeperIdentityPath := filepath.Join(registryDir, "keeper_identity.json")
	keeperIdentityData, _ := json.Marshal(keeperIdentity)
	err = os.WriteFile(keeperIdentityPath, keeperIdentityData, 0644)
	if err != nil {
		b.Fatalf("Failed to write keeper identity: %v", err)
	}
}
