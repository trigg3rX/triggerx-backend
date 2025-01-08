package network

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
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
	peers   map[peer.ID]PeerInfo
}

func NewDiscovery(ctx context.Context, h host.Host, name string) *Discovery {
	return &Discovery{
		host:    h,
		name:    name,
		context: ctx,
		peers:   make(map[peer.ID]PeerInfo),
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

	peerID, err := peer.Decode(service.PeerID)
	if err != nil {
		return "", fmt.Errorf("invalid peer ID in registry: %w", err)
	}

	// Try each address until one succeeds
	var lastErr error
	for _, addr := range service.Addresses {
		maddr, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			lastErr = err
			continue
		}

		peerInfo := peer.AddrInfo{
			ID:    peerID,
			Addrs: []multiaddr.Multiaddr{maddr},
		}

		if err = d.host.Connect(d.context, peerInfo); err != nil {
			lastErr = err
			continue
		}

		d.peers[peerID] = PeerInfo{
			Name:    serviceType,
			Address: strings.Join(service.Addresses, ","),
		}
		return peerID, nil
	}

	return "", fmt.Errorf("failed to connect to peer using any address: %w", lastErr)
}

// GetConnectedPeers returns all currently connected peers
func (d *Discovery) GetConnectedPeers() map[peer.ID]PeerInfo {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	// Create a copy to avoid concurrent access issues
	peers := make(map[peer.ID]PeerInfo)
	for k, v := range d.peers {
		peers[k] = v
	}
	return peers
}
