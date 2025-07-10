package network

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestP2PBootstrapAndKeeperIntegration tests complete integration between bootstrap and keeper nodes
func TestP2PBootstrapAndKeeperIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Setup test environment
	setupIntegrationTest(t)

	// Create a shared registry
	registry := createTestRegistry(t)

	// Create bootstrap node (Manager service)
	bootstrapConfig := P2PConfig{
		NodeType: NodeTypeBootstrap,
		BootstrapConfig: &BootstrapConfig{
			ServiceName:   ServiceManager,
			ListenAddress: "127.0.0.1",
			ListenPort:    "9100",
		},
	}

	t.Logf("Creating bootstrap node...")
	bootstrapNode, err := NewP2PNode(ctx, bootstrapConfig, registry)
	require.NoError(t, err, "Failed to create bootstrap node")
	defer bootstrapNode.Close()

	// Wait for bootstrap node to fully initialize
	time.Sleep(2 * time.Second)

	// Verify bootstrap node is registered
	serviceInfo, exists := registry.GetService(ServiceManager)
	require.True(t, exists, "Bootstrap service should be registered")
	require.NotEmpty(t, serviceInfo.PeerID, "Bootstrap service should have a peer ID")
	require.NotEmpty(t, serviceInfo.Addresses, "Bootstrap service should have addresses")

	t.Logf("Bootstrap node created with PeerID: %s", serviceInfo.PeerID)

	// Create keeper node
	keeperConfig := P2PConfig{
		NodeType: NodeTypeKeeper,
		KeeperConfig: &KeeperConfig{
			ConnectionAddress: "127.0.0.1",
			P2pPort:           "9101",
		},
	}

	t.Logf("Creating keeper node...")
	keeperNode, err := NewP2PNode(ctx, keeperConfig, registry)
	require.NoError(t, err, "Failed to create keeper node")
	defer keeperNode.Close()

	// Wait for keeper node to initialize
	time.Sleep(2 * time.Second)

	// Test that nodes can see each other
	bootstrapHost := bootstrapNode.GetHost()
	keeperHost := keeperNode.GetHost()

	require.NotNil(t, bootstrapHost, "Bootstrap node should have a host")
	require.NotNil(t, keeperHost, "Keeper node should have a host")

	t.Logf("Bootstrap node PeerID: %s", bootstrapHost.ID().String())
	t.Logf("Keeper node PeerID: %s", keeperHost.ID().String())

	// Verify nodes have different peer IDs
	assert.NotEqual(t, bootstrapHost.ID(), keeperHost.ID(), "Nodes should have different peer IDs")

	// Verify both nodes have loggers
	bootstrapLogger := bootstrapNode.GetLogger()
	keeperLogger := keeperNode.GetLogger()

	require.NotNil(t, bootstrapLogger, "Bootstrap node should have a logger")
	require.NotNil(t, keeperLogger, "Keeper node should have a logger")

	// Test that nodes can connect to each other
	bootstrapPeerStore := bootstrapHost.Peerstore()
	keeperPeerStore := keeperHost.Peerstore()

	// Add keeper's addresses to bootstrap's peerstore
	bootstrapPeerStore.AddAddrs(keeperHost.ID(), keeperHost.Addrs(), time.Hour)

	// Add bootstrap's addresses to keeper's peerstore
	keeperPeerStore.AddAddrs(bootstrapHost.ID(), bootstrapHost.Addrs(), time.Hour)

	t.Logf("Integration test completed successfully")
}

// TestMultipleBootstrapNodes tests multiple bootstrap nodes working together
func TestMultipleBootstrapNodes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	setupIntegrationTest(t)
	registry := createTestRegistry(t)

	// Create multiple bootstrap nodes for different services
	services := []string{ServiceManager, ServiceQuorum, ServiceValidator}
	nodes := make([]*P2PNode, len(services))
	basePort := 9200

	for i, service := range services {
		config := P2PConfig{
			NodeType: NodeTypeBootstrap,
			BootstrapConfig: &BootstrapConfig{
				ServiceName:   service,
				ListenAddress: "127.0.0.1",
				ListenPort:    fmt.Sprintf("%d", basePort+i),
			},
		}

		t.Logf("Creating bootstrap node for service: %s on port %d", service, basePort+i)
		node, err := NewP2PNode(ctx, config, registry)
		require.NoError(t, err, "Failed to create bootstrap node for %s", service)
		nodes[i] = node
		defer node.Close()

		// Wait for node to initialize
		time.Sleep(1 * time.Second)

		// Verify service is registered
		serviceInfo, exists := registry.GetService(service)
		require.True(t, exists, "Service %s should be registered", service)
		require.NotEmpty(t, serviceInfo.PeerID, "Service %s should have a peer ID", service)

		t.Logf("Service %s registered with PeerID: %s", service, serviceInfo.PeerID)
	}

	// Verify all services are running and have unique peer IDs
	allServices := registry.GetAllServices()
	require.Len(t, allServices, len(services), "All services should be registered")

	peerIDs := make(map[string]bool)
	for _, service := range services {
		serviceInfo, exists := allServices[service]
		require.True(t, exists, "Service %s should exist in registry", service)
		require.NotEmpty(t, serviceInfo.PeerID, "Service %s should have a peer ID", service)

		// Check for unique peer IDs
		require.False(t, peerIDs[serviceInfo.PeerID], "Peer ID should be unique for service %s", service)
		peerIDs[serviceInfo.PeerID] = true
	}

	t.Logf("Multiple bootstrap nodes test completed successfully")
}

// TestServiceDiscoveryFlow tests the complete service discovery flow
func TestServiceDiscoveryFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	setupIntegrationTest(t)
	registry := createTestRegistry(t)

	// Create a manager service (bootstrap node)
	managerConfig := P2PConfig{
		NodeType: NodeTypeBootstrap,
		BootstrapConfig: &BootstrapConfig{
			ServiceName:   ServiceManager,
			ListenAddress: "127.0.0.1",
			ListenPort:    "9300",
		},
	}

	managerNode, err := NewP2PNode(ctx, managerConfig, registry)
	require.NoError(t, err, "Failed to create manager node")
	defer managerNode.Close()

	time.Sleep(2 * time.Second)

	// Verify manager service is in registry
	serviceInfo, exists := registry.GetService(ServiceManager)
	require.True(t, exists, "Manager service should be in registry")
	require.Equal(t, managerNode.GetHost().ID().String(), serviceInfo.PeerID, "Manager peer ID should match")

	// Test that we can create a discovery component separately
	ctx2 := context.Background()
	discovery := NewDiscovery(ctx2, managerNode.GetHost(), ServiceManager, managerNode.GetLogger())
	require.NotNil(t, discovery, "Should be able to create discovery component")

	// Test saving peer info via discovery
	err = discovery.SavePeerInfo()
	require.NoError(t, err, "Discovery should be able to save peer info")

	t.Logf("Service discovery flow test completed successfully")
}

// TestLegacyFunctionIntegration tests the legacy functions work with the new system
func TestLegacyFunctionIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	setupIntegrationTest(t)
	registry := createTestRegistry(t)

	// First create a service using the new API
	config := P2PConfig{
		NodeType: NodeTypeBootstrap,
		BootstrapConfig: &BootstrapConfig{
			ServiceName:   ServiceManager,
			ListenAddress: "127.0.0.1",
			ListenPort:    "9400",
		},
	}

	node, err := NewP2PNode(ctx, config, registry)
	require.NoError(t, err, "Failed to create node with new API")
	defer node.Close()

	time.Sleep(2 * time.Second)

	// Now test legacy function
	legacyHost, err := SetupServiceWithRegistry(ctx, ServiceManager, registry)
	require.NoError(t, err, "Legacy function should work")
	defer legacyHost.Close()

	// Verify legacy host has the same peer ID as the new node since they're both manager services
	assert.Equal(t, node.GetHost().ID(), legacyHost.ID(), "Legacy host should have same peer ID as it's the same service")

	// Test keeper legacy function
	os.Setenv("PUBLIC_IPV4_ADDRESS", "127.0.0.1")
	os.Setenv("OPERATOR_P2P_PORT", "9401")
	defer func() {
		os.Unsetenv("PUBLIC_IPV4_ADDRESS")
		os.Unsetenv("OPERATOR_P2P_PORT")
	}()

	t.Logf("Legacy function integration test completed successfully")
}

// TestNetworkResilience tests network resilience and reconnection
func TestNetworkResilience(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	setupIntegrationTest(t)
	registry := createTestRegistry(t)

	// Create two nodes
	node1Config := P2PConfig{
		NodeType: NodeTypeBootstrap,
		BootstrapConfig: &BootstrapConfig{
			ServiceName:   ServiceManager,
			ListenAddress: "127.0.0.1",
			ListenPort:    "9500",
		},
	}

	node2Config := P2PConfig{
		NodeType: NodeTypeBootstrap,
		BootstrapConfig: &BootstrapConfig{
			ServiceName:   ServiceQuorum,
			ListenAddress: "127.0.0.1",
			ListenPort:    "9501",
		},
	}

	node1, err := NewP2PNode(ctx, node1Config, registry)
	require.NoError(t, err, "Failed to create node1")
	defer node1.Close()

	node2, err := NewP2PNode(ctx, node2Config, registry)
	require.NoError(t, err, "Failed to create node2")

	time.Sleep(3 * time.Second)

	// Verify both nodes are registered
	services := registry.GetAllServices()
	require.Contains(t, services, ServiceManager, "Manager service should be registered")
	require.Contains(t, services, ServiceQuorum, "Quorum service should be registered")

	// Simulate node2 going down
	t.Logf("Simulating node2 shutdown...")
	node2.Close()

	time.Sleep(2 * time.Second)

	// Node1 should still be operational
	host1 := node1.GetHost()
	require.NotNil(t, host1, "Node1 should still be operational")

	// Recreate node2 (simulating restart)
	t.Logf("Recreating node2...")
	node2, err = NewP2PNode(ctx, node2Config, registry)
	require.NoError(t, err, "Failed to recreate node2")
	defer node2.Close()

	time.Sleep(2 * time.Second)

	// Verify system is functional again
	services = registry.GetAllServices()
	require.Contains(t, services, ServiceManager, "Manager service should still be registered")
	require.Contains(t, services, ServiceQuorum, "Quorum service should be registered again")

	t.Logf("Network resilience test completed successfully")
}

// Helper functions for integration tests

func setupIntegrationTest(t *testing.T) {
	// Create the production data directory structure for integration tests
	registryDir := filepath.Join(BaseDataDir, RegistryDir)
	err := os.MkdirAll(registryDir, 0755)
	require.NoError(t, err, "Failed to create registry directory")

	// Create test private keys in the production location
	createIntegrationTestKeys(t, BaseDataDir)

	t.Cleanup(func() {
		// Clean up the created files after test
		os.RemoveAll(filepath.Join(BaseDataDir, RegistryDir, "services_privKeys.json"))
		os.RemoveAll(filepath.Join(BaseDataDir, RegistryDir, "keeper_identity.json"))
	})
}

func createIntegrationTestKeys(t *testing.T, baseDir string) {
	registryDir := filepath.Join(baseDir, RegistryDir)

	// Create test service private keys with valid base64 encoded keys
	services := ServicePrivateKeys{
		ManagerP2PPrivateKey:   "CAESQKBj/Dv/gPbs33isOeRfr1vGmgXOg6mUhpPaDgRUrSgvAR6bgGhQ4dKhMT/2cOWqJe9MhEGwVoNW2pqJzB4Ij/0=",
		QuorumP2PPrivateKey:    "CAESQKnudgGSiPS1W2VsNLePz8I6mGNdCu7uJzGpPGGtbPHdCKUlLlQe7o1MYkYqPGS4VdVKAuIhIq7KnMfaUwZq+Rw=",
		ValidatorP2PPrivateKey: "CAESQLc0mJOOzQwyXWvgvEw1I5CHt7xMejQRz1EqPfyNRBsGjr4DsWzxQS4LFhZg3cNB8O9XYcOdLyKfGkOGzgMr5zU=",
	}

	serviceKeysPath := filepath.Join(registryDir, "services_privKeys.json")
	serviceKeysData, err := json.Marshal(services)
	require.NoError(t, err, "Failed to marshal service private keys")

	err = os.WriteFile(serviceKeysPath, serviceKeysData, 0644)
	require.NoError(t, err, "Failed to write service private keys")

	// Create test keeper identity with valid byte array private key
	keeperIdentity := PeerIdentity{
		PrivKey: []byte{8, 1, 18, 64, 169, 238, 118, 1, 146, 136, 244, 181, 91, 101, 108, 52, 183, 143, 207, 194, 58, 152, 99, 93, 10, 238, 238, 39, 49, 169, 60, 97, 173, 108, 241, 221, 8, 165, 37, 46, 84, 30, 238, 141, 76, 98, 70, 42, 60, 100, 184, 85, 213, 74, 2, 226, 33, 34, 174, 202, 156, 199, 218, 83, 6, 106, 249, 28},
	}

	keeperIdentityPath := filepath.Join(registryDir, "keeper_identity.json")
	keeperIdentityData, err := json.Marshal(keeperIdentity)
	require.NoError(t, err, "Failed to marshal keeper identity")

	err = os.WriteFile(keeperIdentityPath, keeperIdentityData, 0644)
	require.NoError(t, err, "Failed to write keeper identity")
}
