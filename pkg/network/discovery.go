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
	"github.com/trigg3rX/triggerx-backend-imua/pkg/logging"
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
	logger  logging.Logger
}

func NewDiscovery(ctx context.Context, h host.Host, name string, logger logging.Logger) *Discovery {
	logger.Infof("Initializing discovery for service: %s", name)

	return &Discovery{
		host:    h,
		name:    name,
		context: ctx,
		logger:  logger,
	}
}

// SavePeerInfo saves this peer's info to the registry
func (d *Discovery) SavePeerInfo() error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.logger.Debugf("Saving peer info for service: %s", d.name)

	registry, err := NewPeerRegistry()
	if err != nil {
		d.logger.Errorf("Failed to create peer registry: %v", err)
		return fmt.Errorf("failed to create peer registry: %w", err)
	}

	addrs := make([]string, 0)
	for _, addr := range d.host.Addrs() {
		addrs = append(addrs, addr.String())
	}

	d.logger.Infof("Updating registry with %d addresses for service %s", len(addrs), d.name)
	err = registry.UpdateService(d.name, d.host.ID(), addrs)
	if err != nil {
		d.logger.Errorf("Failed to update service in registry: %v", err)
		return err
	}

	d.logger.Debugf("Successfully saved peer info for service: %s", d.name)
	return nil
}

// ConnectToPeer connects to a specific peer using stored registry info
func (d *Discovery) ConnectToPeer(serviceType string) (peer.ID, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.logger.Infof("Attempting to connect to service: %s", serviceType)

	registry, err := NewPeerRegistry()
	if err != nil {
		d.logger.Errorf("Failed to load peer registry: %v", err)
		return "", fmt.Errorf("failed to load peer registry: %w", err)
	}

	service, exists := registry.GetService(serviceType)
	if !exists {
		d.logger.Errorf("Service %s not found in registry", serviceType)
		return "", fmt.Errorf("service %s not found in registry", serviceType)
	}

	if service.PeerID == "" {
		d.logger.Errorf("Service %s has no peer ID", serviceType)
		return "", fmt.Errorf("service %s has no peer ID", serviceType)
	}

	targetPeerID, err := peer.Decode(service.PeerID)
	if err != nil {
		d.logger.Errorf("Invalid peer ID in registry for service %s: %v", serviceType, err)
		return "", fmt.Errorf("invalid peer ID in registry: %w", err)
	}

	d.logger.Debugf("Target peer ID: %s", targetPeerID.String())

	// Convert addresses to multiaddr
	var addrs []multiaddr.Multiaddr
	for i, addrStr := range service.Addresses {
		addr, err := multiaddr.NewMultiaddr(addrStr)
		if err != nil {
			d.logger.Warnf("Skipping invalid address %d for service %s: %s", i, serviceType, addrStr)
			continue
		}
		addrs = append(addrs, addr)
	}

	if len(addrs) == 0 {
		d.logger.Errorf("No valid addresses found for service %s", serviceType)
		return "", fmt.Errorf("no valid addresses found for service %s", serviceType)
	}

	d.logger.Debugf("Found %d valid addresses for service %s", len(addrs), serviceType)

	// Add target's addresses to our peerstore with permanent TTL
	d.host.Peerstore().AddAddrs(targetPeerID, addrs, peerstore.PermanentAddrTTL)
	d.logger.Debugf("Added target addresses to peerstore")

	// Add our addresses to target's info in registry
	myAddrs := make([]string, 0)
	for _, addr := range d.host.Addrs() {
		myAddrs = append(myAddrs, addr.String())
	}
	if err := registry.UpdateService(d.name, d.host.ID(), myAddrs); err != nil {
		d.logger.Warnf("Failed to update our service info in registry: %v", err)
		return "", fmt.Errorf("failed to update our service info: %w", err)
	}

	// Establish connection from our side
	d.logger.Debugf("Establishing connection to peer %s", targetPeerID.String())
	if err := d.host.Connect(d.context, peer.AddrInfo{
		ID:    targetPeerID,
		Addrs: addrs,
	}); err != nil {
		d.logger.Errorf("Failed to connect to peer %s: %v", targetPeerID, err)
		return "", fmt.Errorf("failed to connect to peer %s: %w", targetPeerID, err)
	}

	// Wait a short time for the connection to stabilize
	time.Sleep(time.Millisecond * 100)

	// Verify connection status
	conns := d.host.Network().ConnsToPeer(targetPeerID)
	if len(conns) == 0 {
		d.logger.Errorf("Failed to establish stable connection with peer %s", targetPeerID)
		return "", fmt.Errorf("failed to establish stable connection with peer %s", targetPeerID)
	}

	d.logger.Infof("Successfully connected to service %s (PeerID: %s) with %d connection(s)",
		serviceType, targetPeerID.String(), len(conns))

	return targetPeerID, nil
}

// GetConnectedPeers returns all currently connected peers
func (d *Discovery) GetConnectedPeers() map[peer.ID]PeerInfo {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	d.logger.Debugf("Getting connected peers")

	peers := make(map[peer.ID]PeerInfo)

	// Get all peers from peerstore that have addresses
	peersWithAddrs := d.host.Peerstore().PeersWithAddrs()
	d.logger.Debugf("Found %d peers with addresses in peerstore", len(peersWithAddrs))

	for _, peerID := range peersWithAddrs {
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

	d.logger.Debugf("Returning %d connected peers", len(peers))
	return peers
}

func (d *Discovery) IsConnected(peerID peer.ID) bool {
	connected := len(d.host.Network().ConnsToPeer(peerID)) > 0
	d.logger.Debugf("Checking connection to peer %s: %v", peerID.String(), connected)
	return connected
}

// GetLogger returns the logger instance
func (d *Discovery) GetLogger() logging.Logger {
	return d.logger
}
