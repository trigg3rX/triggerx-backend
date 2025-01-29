package network

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/multiformats/go-multiaddr"
)

type PeerInfo struct {
	Name    string
	Address string
}

type Discovery struct {
	host    host.Host
	name    string
	context context.Context
	mutex   sync.RWMutex
}

func NewDiscovery(ctx context.Context, h host.Host, name string) *Discovery {
	return &Discovery{
		host:    h,
		name:    name,
		context: ctx,
	}
}

// SavePeerInfo saves this peer's info to the registry
func (d *Discovery) SavePeerInfo() error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	registry, err := NewPeerRegistry()
	if err != nil {
		return fmt.Errorf("failed to create peer registry: %w", err)
	}

	addrs := make([]string, 0)
	for _, addr := range d.host.Addrs() {
		addrs = append(addrs, addr.String())
	}

	return registry.UpdateService(d.name, d.host.ID(), addrs)
}

// ConnectToPeer connects to a specific peer using stored registry info
func (d *Discovery) ConnectToPeer(serviceType string) (peer.ID, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	registry, err := NewPeerRegistry()
	if err != nil {
		return "", fmt.Errorf("failed to load peer registry: %w", err)
	}

	service, exists := registry.GetService(serviceType)
	if !exists {
		return "", fmt.Errorf("service %s not found in registry", serviceType)
	}

	targetPeerID, err := peer.Decode(service.PeerID)
	if err != nil {
		return "", fmt.Errorf("invalid peer ID in registry: %w", err)
	}

	// Convert addresses to multiaddr
	var addrs []multiaddr.Multiaddr
	for _, addrStr := range service.Addresses {
		addr, err := multiaddr.NewMultiaddr(addrStr)
		if err != nil {
			continue
		}
		addrs = append(addrs, addr)
	}

	// Add target's addresses to our peerstore with permanent TTL
	d.host.Peerstore().AddAddrs(targetPeerID, addrs, peerstore.PermanentAddrTTL)

	// Add our addresses to target's info in registry
	myAddrs := make([]string, 0)
	for _, addr := range d.host.Addrs() {
		myAddrs = append(myAddrs, addr.String())
	}
	if err := registry.UpdateService(d.name, d.host.ID(), myAddrs); err != nil {
		return "", fmt.Errorf("failed to update our service info: %w", err)
	}

	// Establish connection from our side
	if err := d.host.Connect(d.context, peer.AddrInfo{
		ID:    targetPeerID,
		Addrs: addrs,
	}); err != nil {
		return "", fmt.Errorf("failed to connect to peer %s: %w", targetPeerID, err)
	}

	// Wait a short time for the connection to stabilize
	time.Sleep(time.Millisecond * 100)

	// Verify connection status
	conns := d.host.Network().ConnsToPeer(targetPeerID)
	if len(conns) == 0 {
		return "", fmt.Errorf("failed to establish stable connection with peer %s", targetPeerID)
	}

	return targetPeerID, nil
}

// GetConnectedPeers returns all currently connected peers
func (d *Discovery) GetConnectedPeers() map[peer.ID]PeerInfo {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	peers := make(map[peer.ID]PeerInfo)

	// Get all peers from peerstore that have addresses
	for _, peerID := range d.host.Peerstore().PeersWithAddrs() {
		addrs := d.host.Peerstore().Addrs(peerID)
		if len(addrs) > 0 {
			addrStrings := make([]string, len(addrs))
			for i, addr := range addrs {
				addrStrings[i] = addr.String()
			}

			peers[peerID] = PeerInfo{
				Address: strings.Join(addrStrings, ","),
			}
		}
	}

	return peers
}

func (d *Discovery) IsConnected(peerID peer.ID) bool {
	return len(d.host.Network().ConnsToPeer(peerID)) > 0
}
