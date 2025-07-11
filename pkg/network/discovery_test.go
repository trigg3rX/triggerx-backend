package network

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/logging"
)

var logger logging.Logger

func init() {
	var err error
	logger, err = logging.NewZapLogger(logging.LoggerConfig{
		ProcessName:   logging.ProcessName(fmt.Sprintf("discovery-%s", ServiceManager)),
		IsDevelopment: true,
	})
	if err != nil {
		panic(err)
	}
}

// TestNewDiscovery tests the creation of a new Discovery instance
func TestNewDiscovery(t *testing.T) {
	ctx := context.Background()

	// Create a test host
	host, err := libp2p.New()
	require.NoError(t, err)
	defer host.Close()

	tests := []struct {
		name        string
		serviceName string
		shouldPanic bool
	}{
		{
			name:        "valid discovery creation",
			serviceName: ServiceManager,
			shouldPanic: false,
		},
		{
			name:        "empty service name",
			serviceName: "",
			shouldPanic: false, // Should still work with empty service name
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			discovery := NewDiscovery(ctx, host, tt.serviceName, logger)

			assert.NotNil(t, discovery)
			assert.NotNil(t, discovery.logger)
			assert.Equal(t, host, discovery.host)
		})
	}
}

// TestSavePeerInfo tests saving peer information
func TestSavePeerInfo(t *testing.T) {
	ctx := context.Background()

	// Create test host
	host1, err := libp2p.New()
	require.NoError(t, err)
	defer host1.Close()

	discovery := NewDiscovery(ctx, host1, ServiceManager, logger)

	// Test saving peer info
	err = discovery.SavePeerInfo()
	assert.NoError(t, err)
}

// TestConnectToPeer tests peer connection functionality
func TestConnectToPeer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create test hosts
	host1, err := libp2p.New()
	require.NoError(t, err)
	defer host1.Close()

	host2, err := libp2p.New()
	require.NoError(t, err)
	defer host2.Close()

	discovery1 := NewDiscovery(ctx, host1, ServiceManager, logger)
	discovery2 := NewDiscovery(ctx, host2, ServiceQuorum, logger)

	// Save host2 info in registry first
	err = discovery2.SavePeerInfo()
	require.NoError(t, err)

	// Test connecting to peer
	peerID, err := discovery1.ConnectToPeer(ServiceQuorum)
	assert.NoError(t, err)
	assert.Equal(t, host2.ID(), peerID)

	// Verify connection
	assert.True(t, discovery1.IsConnected(host2.ID()))
	assert.True(t, discovery2.IsConnected(host1.ID()))

	// Test connecting to non-existent service
	_, err = discovery1.ConnectToPeer("non-existent-service")
	assert.Error(t, err)
}

// TestGetConnectedPeers tests retrieving connected peers
func TestGetConnectedPeers(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create test hosts
	host1, err := libp2p.New()
	require.NoError(t, err)
	defer host1.Close()

	host2, err := libp2p.New()
	require.NoError(t, err)
	defer host2.Close()

	host3, err := libp2p.New()
	require.NoError(t, err)
	defer host3.Close()

	discovery1 := NewDiscovery(ctx, host1, ServiceManager, logger)
	discovery2 := NewDiscovery(ctx, host2, ServiceQuorum, logger)
	discovery3 := NewDiscovery(ctx, host3, ServiceValidator, logger)

	// Check initial state - may have existing peers from previous tests
	initialPeers := discovery1.GetConnectedPeers()
	initialCount := len(initialPeers)

	// Save peer info and connect to peer2
	err = discovery2.SavePeerInfo()
	require.NoError(t, err)
	_, err = discovery1.ConnectToPeer(ServiceQuorum)
	require.NoError(t, err)

	// Save peer info and connect to peer3
	err = discovery3.SavePeerInfo()
	require.NoError(t, err)
	_, err = discovery1.ConnectToPeer(ServiceValidator)
	require.NoError(t, err)

	// Should have initial + 2 connected peers
	connectedPeers := discovery1.GetConnectedPeers()
	assert.GreaterOrEqual(t, len(connectedPeers), initialCount+2)
}

// TestIsConnected tests the connection status check
func TestIsConnected(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create test hosts
	host1, err := libp2p.New()
	require.NoError(t, err)
	defer host1.Close()

	host2, err := libp2p.New()
	require.NoError(t, err)
	defer host2.Close()

	discovery1 := NewDiscovery(ctx, host1, ServiceManager, logger)
	discovery2 := NewDiscovery(ctx, host2, ServiceQuorum, logger)

	// Initially not connected
	assert.False(t, discovery1.IsConnected(host2.ID()))

	// Save peer info and connect
	err = discovery2.SavePeerInfo()
	require.NoError(t, err)
	_, err = discovery1.ConnectToPeer(ServiceQuorum)
	require.NoError(t, err)

	// Should be connected now
	assert.True(t, discovery1.IsConnected(host2.ID()))

	// Test with non-existent peer
	nonExistentPeerID, err := peer.Decode("12D3KooWGzxzKZYveHXtpG6AsrUJBcWxHBFS2HsEoGTxrMLvKXtf")
	require.NoError(t, err)
	assert.False(t, discovery1.IsConnected(nonExistentPeerID))
}

// TestDiscoveryWithServiceRegistry tests discovery with service registry
func TestDiscoveryWithServiceRegistry(t *testing.T) {
	ctx := context.Background()

	host1, err := libp2p.New()
	require.NoError(t, err)
	defer host1.Close()

	discovery := NewDiscovery(ctx, host1, ServiceManager, logger)

	// Test saving peer info
	err = discovery.SavePeerInfo()
	assert.NoError(t, err)

	// Test connecting to non-existent service
	_, err = discovery.ConnectToPeer("non-existent")
	assert.Error(t, err)
}

// TestDiscoveryLogger tests that the logger is properly initialized
func TestDiscoveryLogger(t *testing.T) {
	ctx := context.Background()

	host, err := libp2p.New()
	require.NoError(t, err)
	defer host.Close()

	discovery := NewDiscovery(ctx, host, ServiceManager, logger)

	// Verify logger is not nil
	assert.NotNil(t, discovery.logger)

	// Test that we can call logger methods without panicking
	assert.NotPanics(t, func() {
		discovery.logger.Info("Test log message")
		discovery.logger.Debug("Test debug message")
		discovery.logger.Warn("Test warning message")
	})
}

// TestDiscoveryMultipleConnections tests handling multiple connections
func TestDiscoveryMultipleConnections(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Create multiple test hosts
	const numHosts = 5
	hosts := make([]host.Host, numHosts)
	discoveries := make([]*Discovery, numHosts)

	for i := 0; i < numHosts; i++ {
		host, err := libp2p.New()
		require.NoError(t, err)
		defer host.Close()
		hosts[i] = host

		serviceName := fmt.Sprintf("service-%d", i)
		discovery := NewDiscovery(ctx, host, serviceName, logger)
		discoveries[i] = discovery

		// Save each service to registry
		err = discovery.SavePeerInfo()
		require.NoError(t, err)
	}

	// Connect first host to all others
	mainDiscovery := discoveries[0]

	// Check initial connected peers count
	initialPeers := mainDiscovery.GetConnectedPeers()
	initialCount := len(initialPeers)

	connectionsAdded := 0
	for i := 1; i < numHosts; i++ {
		serviceName := fmt.Sprintf("service-%d", i)
		_, err := mainDiscovery.ConnectToPeer(serviceName)
		require.NoError(t, err)
		connectionsAdded++

		// Verify connection
		assert.True(t, mainDiscovery.IsConnected(hosts[i].ID()))
	}

	// Verify the main host has at least the new connections
	connectedPeers := mainDiscovery.GetConnectedPeers()
	assert.GreaterOrEqual(t, len(connectedPeers), initialCount+connectionsAdded)
}

// Benchmark tests for Discovery
func BenchmarkNewDiscovery(b *testing.B) {
	ctx := context.Background()
	host, err := libp2p.New()
	require.NoError(b, err)
	defer host.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		discovery := NewDiscovery(ctx, host, ServiceManager, logger)
		_ = discovery
	}
}

func BenchmarkSavePeerInfo(b *testing.B) {
	ctx := context.Background()
	host1, err := libp2p.New()
	require.NoError(b, err)
	defer host1.Close()

	discovery := NewDiscovery(ctx, host1, ServiceManager, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := discovery.SavePeerInfo()
		if err != nil {
			b.Fatalf("Failed to save peer info: %v", err)
		}
	}
}

func BenchmarkConnectToPeer(b *testing.B) {
	ctx := context.Background()
	host1, err := libp2p.New()
	require.NoError(b, err)
	defer host1.Close()

	host2, err := libp2p.New()
	require.NoError(b, err)
	defer host2.Close()

	discovery1 := NewDiscovery(ctx, host1, ServiceManager, logger)
	discovery2 := NewDiscovery(ctx, host2, ServiceQuorum, logger)

	// Save peer info once
	err = discovery2.SavePeerInfo()
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Connect and disconnect for each iteration
		_, err := discovery1.ConnectToPeer(ServiceQuorum)
		if err != nil {
			b.Fatalf("Failed to connect to peer: %v", err)
		}

		// Disconnect to reset for next iteration
		host1.Network().ClosePeer(host2.ID())
	}
}
