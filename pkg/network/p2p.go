package network

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"encoding/base64"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/host/peerstore/pstoremem"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/multiformats/go-multiaddr"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/logging"
)

// P2PNodeType represents the type of P2P node
type P2PNodeType string

const (
	NodeTypeBootstrap P2PNodeType = "bootstrap"
	NodeTypeKeeper    P2PNodeType = "keeper"
)

// BootstrapConfig contains configurable environment settings for bootstrap nodes
type BootstrapConfig struct {
	ServiceName    string   // manager, quorum, validator
	ListenAddress  string   // IP address to listen on
	ListenPort     string   // Port to listen on
	BootstrapPeers []string // List of bootstrap peer addresses
}

// KeeperConfig contains fixed environment settings for keepers
type KeeperConfig struct {
	ConnectionAddress string // Fixed connection address for keepers
	P2pPort           string // Fixed P2P port for keepers
}

// P2PConfig represents the unified configuration for P2P nodes
type P2PConfig struct {
	NodeType        P2PNodeType
	BootstrapConfig *BootstrapConfig // Used for bootstrap nodes
	KeeperConfig    *KeeperConfig    // Used for keeper nodes
}

type PeerIdentity struct {
	PrivKey []byte `json:"priv_key"`
}

type ServicePrivateKeys struct {
	ManagerP2PPrivateKey   string `json:"manager_p2p_private_key"`
	QuorumP2PPrivateKey    string `json:"quorum_p2p_private_key"`
	ValidatorP2PPrivateKey string `json:"validator_p2p_private_key"`
}

// P2PNode represents a P2P node instance
type P2PNode struct {
	host     host.Host
	nodeType P2PNodeType
	config   P2PConfig
	registry *PeerRegistry
	logger   logging.Logger
}

// NewP2PNode creates a new P2P node with the specified configuration
func NewP2PNode(ctx context.Context, config P2PConfig, registry *PeerRegistry) (*P2PNode, error) {
	// Initialize logger based on node type
	var processName logging.ProcessName
	var serviceName string

	switch config.NodeType {
	case NodeTypeBootstrap:
		if config.BootstrapConfig == nil {
			return nil, fmt.Errorf("bootstrap config is required for bootstrap node")
		}
		serviceName = config.BootstrapConfig.ServiceName
		processName = logging.ProcessName(fmt.Sprintf("p2p-%s", serviceName))
	case NodeTypeKeeper:
		if config.KeeperConfig == nil {
			return nil, fmt.Errorf("keeper config is required for keeper node")
		}
		serviceName = "keeper"
		processName = logging.KeeperProcess
	default:
		return nil, fmt.Errorf("unsupported node type: %s", config.NodeType)
	}

	// Initialize logger
	logConfig := logging.LoggerConfig{
		ProcessName:   processName,
		IsDevelopment: true, // You can make this configurable if needed
	}

	logger, err := logging.NewZapLogger(logConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger for %s: %w", serviceName, err)
	}

	logger.Infof("Initializing P2P node - Type: %s, Service: %s", config.NodeType, serviceName)

	var h host.Host
	switch config.NodeType {
	case NodeTypeBootstrap:
		h, err = setupBootstrapNode(ctx, *config.BootstrapConfig, registry, logger)
	case NodeTypeKeeper:
		h, err = setupKeeperNode(ctx, *config.KeeperConfig, registry, logger)
	}

	if err != nil {
		logger.Errorf("Failed to setup P2P node: %v", err)
		return nil, fmt.Errorf("failed to setup P2P node: %w", err)
	}

	logger.Infof("P2P node successfully initialized - PeerID: %s", h.ID().String())

	return &P2PNode{
		host:     h,
		nodeType: config.NodeType,
		config:   config,
		registry: registry,
		logger:   logger,
	}, nil
}

// GetHost returns the libp2p host instance
func (p *P2PNode) GetHost() host.Host {
	return p.host
}

// GetNodeType returns the node type
func (p *P2PNode) GetNodeType() P2PNodeType {
	return p.nodeType
}

// GetLogger returns the logger instance
func (p *P2PNode) GetLogger() logging.Logger {
	return p.logger
}

// Close closes the P2P node
func (p *P2PNode) Close() error {
	if p.host != nil {
		p.logger.Infof("Closing P2P node - PeerID: %s", p.host.ID().String())
		return p.host.Close()
	}
	return nil
}

// setupBootstrapNode sets up a bootstrap node (manager, quorum, validator) with configurable environment
func setupBootstrapNode(ctx context.Context, config BootstrapConfig, registry *PeerRegistry, logger logging.Logger) (host.Host, error) {
	logger.Infof("Setting up bootstrap node for service: %s", config.ServiceName)

	priv, err := LoadIdentity(config.ServiceName)
	if err != nil {
		logger.Errorf("Failed to load identity for service %s: %v", config.ServiceName, err)
		return nil, fmt.Errorf("failed to setup peer identity for service %s: %w", config.ServiceName, err)
	}

	// Use configurable address from BootstrapConfig
	addr := fmt.Sprintf("/ip4/%s/tcp/%s", config.ListenAddress, config.ListenPort)
	logger.Debugf("Bootstrap node listen address: %s", addr)

	listenAddr, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		logger.Errorf("Invalid listen address %s: %v", addr, err)
		return nil, fmt.Errorf("invalid listen address %s: %w", addr, err)
	}

	connMgr, err := connmgr.NewConnManager(
		200, 400,
		connmgr.WithGracePeriod(time.Minute),
	)
	if err != nil {
		logger.Errorf("Failed to create connection manager: %v", err)
		return nil, fmt.Errorf("failed to create connection manager: %w", err)
	}

	promReg := prometheus.NewRegistry()
	ps, err := pstoremem.NewPeerstore()
	if err != nil {
		logger.Errorf("Failed to create peerstore: %v", err)
		return nil, fmt.Errorf("failed to create peerstore: %w", err)
	}

	// Configure bootstrap peers based on service type
	var bootstrapPeers []peer.AddrInfo
	if config.ServiceName != ServiceManager {
		logger.Debugf("Looking for manager service in registry for bootstrap")
		managerInfo, exists := registry.GetService(ServiceManager)
		if exists && managerInfo.PeerID != "" {
			peerID, err := peer.Decode(managerInfo.PeerID)
			if err == nil {
				logger.Infof("Found manager service, adding as bootstrap peer: %s", peerID.String())
				for _, addr := range managerInfo.Addresses {
					maddr, err := multiaddr.NewMultiaddr(addr)
					if err == nil {
						bootstrapPeers = append(bootstrapPeers, peer.AddrInfo{
							ID:    peerID,
							Addrs: []multiaddr.Multiaddr{maddr},
						})
						ps.AddAddrs(peerID, []multiaddr.Multiaddr{maddr}, time.Hour*168)
					}
				}
			}
		} else {
			logger.Warnf("Manager service not found in registry or missing peer ID")
		}
	}

	// Add additional bootstrap peers from config
	logger.Debugf("Adding %d additional bootstrap peers from config", len(config.BootstrapPeers))
	for _, bootstrapAddr := range config.BootstrapPeers {
		maddr, err := multiaddr.NewMultiaddr(bootstrapAddr)
		if err != nil {
			logger.Warnf("Skipping invalid bootstrap address: %s", bootstrapAddr)
			continue // Skip invalid addresses
		}
		addrInfo, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			logger.Warnf("Skipping invalid peer address: %s", bootstrapAddr)
			continue // Skip invalid peer addresses
		}
		logger.Debugf("Added bootstrap peer: %s", addrInfo.ID.String())
		bootstrapPeers = append(bootstrapPeers, *addrInfo)
		ps.AddAddrs(addrInfo.ID, addrInfo.Addrs, time.Hour*168)
	}

	logger.Infof("Creating libp2p host with %d bootstrap peers", len(bootstrapPeers))
	h, err := libp2p.New(
		libp2p.Identity(priv),
		libp2p.ListenAddrs(listenAddr),
		libp2p.ConnectionManager(connMgr),
		libp2p.Peerstore(ps),
		libp2p.PrometheusRegisterer(promReg),
		libp2p.EnableRelay(),
		libp2p.EnableAutoRelayWithStaticRelays(bootstrapPeers),
		libp2p.NATPortMap(),
	)
	if err != nil {
		logger.Errorf("Failed to create bootstrap host: %v", err)
		return nil, fmt.Errorf("failed to create bootstrap host: %w", err)
	}

	// Update registry with this service's information
	addrs := make([]string, 0)
	for _, addr := range h.Addrs() {
		addrs = append(addrs, addr.String())
	}
	logger.Infof("Updating registry for service %s with %d addresses", config.ServiceName, len(addrs))
	if err := registry.UpdateService(config.ServiceName, h.ID(), addrs); err != nil {
		logger.Errorf("Failed to update service registry: %v", err)
		return nil, fmt.Errorf("failed to update service registry: %w", err)
	}

	logger.Infof("Bootstrap node setup completed successfully - PeerID: %s", h.ID().String())
	return h, nil
}

// setupKeeperNode sets up a keeper node with fixed environment
func setupKeeperNode(ctx context.Context, config KeeperConfig, registry *PeerRegistry, logger logging.Logger) (host.Host, error) {
	logger.Infof("Setting up keeper node")

	priv, err := LoadIdentity(ServiceKeeper)
	if err != nil {
		logger.Errorf("Failed to load keeper identity: %v", err)
		return nil, fmt.Errorf("failed to setup peer identity for keeper: %w", err)
	}

	// Use fixed address from KeeperConfig
	addr := fmt.Sprintf("/ip4/%s/tcp/%s", config.ConnectionAddress, config.P2pPort)
	logger.Debugf("Keeper node listen address: %s", addr)

	listenAddr, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		logger.Errorf("Invalid keeper address %s: %v", addr, err)
		return nil, fmt.Errorf("invalid keeper address %s: %w", addr, err)
	}

	connMgr, err := connmgr.NewConnManager(
		200, 400,
		connmgr.WithGracePeriod(time.Minute),
	)
	if err != nil {
		logger.Errorf("Failed to create connection manager: %v", err)
		return nil, fmt.Errorf("failed to create connection manager: %w", err)
	}

	promReg := prometheus.NewRegistry()
	ps, err := pstoremem.NewPeerstore()
	if err != nil {
		logger.Errorf("Failed to create peerstore: %v", err)
		return nil, fmt.Errorf("failed to create peerstore: %w", err)
	}

	// Connect to all bootstrap services
	services := registry.GetAllServices()
	var bootstrapPeers []peer.AddrInfo
	logger.Debugf("Found %d services in registry", len(services))

	for serviceName, service := range services {
		if service.PeerID == "" || len(service.Addresses) == 0 {
			logger.Debugf("Skipping service %s - missing peer info", serviceName)
			continue // Skip services without peer info
		}

		peerID, err := peer.Decode(service.PeerID)
		if err != nil {
			logger.Warnf("Skipping service %s - invalid peer ID: %v", serviceName, err)
			continue // Skip invalid peer IDs
		}

		logger.Infof("Adding bootstrap service: %s (PeerID: %s)", serviceName, peerID.String())
		for _, addrStr := range service.Addresses {
			maddr, err := multiaddr.NewMultiaddr(addrStr)
			if err != nil {
				logger.Warnf("Skipping invalid address for service %s: %s", serviceName, addrStr)
				continue // Skip invalid addresses
			}
			bootstrapPeers = append(bootstrapPeers, peer.AddrInfo{
				ID:    peerID,
				Addrs: []multiaddr.Multiaddr{maddr},
			})
			ps.AddAddrs(peerID, []multiaddr.Multiaddr{maddr}, time.Hour*168)
		}
	}

	logger.Infof("Creating keeper host with %d bootstrap peers", len(bootstrapPeers))
	h, err := libp2p.New(
		libp2p.Identity(priv),
		libp2p.ListenAddrs(listenAddr),
		libp2p.ConnectionManager(connMgr),
		libp2p.Peerstore(ps),
		libp2p.PrometheusRegisterer(promReg),
		libp2p.EnableRelay(),
		libp2p.EnableAutoRelayWithStaticRelays(bootstrapPeers),
		libp2p.NATPortMap(),
	)
	if err != nil {
		logger.Errorf("Failed to create keeper host: %v", err)
		return nil, fmt.Errorf("failed to create keeper host: %w", err)
	}

	logger.Infof("Keeper node setup completed successfully - PeerID: %s", h.ID().String())
	return h, nil
}

func LoadIdentity(nodeType string) (crypto.PrivKey, error) {
	if nodeType == ServiceManager || nodeType == ServiceQuorum || nodeType == ServiceValidator {
		privKeysPath := filepath.Join(BaseDataDir, RegistryDir, "services_privKeys.json")
		data, err := os.ReadFile(privKeysPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read services private keys: %w", err)
		}

		var keys ServicePrivateKeys
		if err := json.Unmarshal(data, &keys); err != nil {
			return nil, fmt.Errorf("failed to unmarshal private keys: %w", err)
		}

		var privKeyStr string
		switch nodeType {
		case ServiceManager:
			privKeyStr = keys.ManagerP2PPrivateKey
		case ServiceQuorum:
			privKeyStr = keys.QuorumP2PPrivateKey
		case ServiceValidator:
			privKeyStr = keys.ValidatorP2PPrivateKey
		}

		privKeyBytes, err := base64.StdEncoding.DecodeString(privKeyStr)
		if err != nil {
			return nil, fmt.Errorf("failed to decode private key: %w", err)
		}

		return crypto.UnmarshalPrivateKey(privKeyBytes)
	} else if nodeType == ServiceKeeper {
		identityFile := filepath.Join(BaseDataDir, RegistryDir, "keeper_identity.json")

		data, err := os.ReadFile(identityFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read keeper identity: %w", err)
		}

		var identity PeerIdentity
		if err := json.Unmarshal(data, &identity); err != nil {
			return nil, fmt.Errorf("failed to unmarshal keeper identity: %w", err)
		}

		priv, err := crypto.UnmarshalPrivateKey(identity.PrivKey)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal keeper private key: %w", err)
		}

		return priv, nil
	}

	return nil, fmt.Errorf("invalid node type: %s", nodeType)
}

// Legacy functions for backward compatibility

// SetupServiceWithRegistry sets up a bootstrap service node with registry
func SetupServiceWithRegistry(ctx context.Context, serviceName string, registry *PeerRegistry) (host.Host, error) {
	// Initialize logger for legacy function
	logConfig := logging.LoggerConfig{
		ProcessName:   logging.ProcessName(fmt.Sprintf("p2p-%s", serviceName)),
		IsDevelopment: true,
	}

	logger, err := logging.NewZapLogger(logConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger for legacy service setup: %w", err)
	}

	logger.Infof("Setting up service using legacy function - Service: %s", serviceName)

	serviceInfo, exists := registry.GetService(serviceName)
	if !exists {
		logger.Errorf("Service %s not found in registry", serviceName)
		return nil, fmt.Errorf("service %s not found in registry", serviceName)
	}

	if len(serviceInfo.Addresses) == 0 {
		logger.Errorf("No addresses found for service %s", serviceName)
		return nil, fmt.Errorf("no addresses found for service %s", serviceName)
	}

	// Parse the first address to extract IP and port
	addr, err := multiaddr.NewMultiaddr(serviceInfo.Addresses[0])
	if err != nil {
		logger.Errorf("Invalid address in registry: %v", err)
		return nil, fmt.Errorf("invalid address in registry: %w", err)
	}

	// Extract IP and port from multiaddr
	var ip, port string
	multiaddr.ForEach(addr, func(c multiaddr.Component) bool {
		if c.Protocol().Code == multiaddr.P_IP4 {
			ip = c.Value()
		} else if c.Protocol().Code == multiaddr.P_TCP {
			port = c.Value()
		}
		return true
	})

	logger.Debugf("Extracted address - IP: %s, Port: %s", ip, port)

	config := P2PConfig{
		NodeType: NodeTypeBootstrap,
		BootstrapConfig: &BootstrapConfig{
			ServiceName:   serviceName,
			ListenAddress: ip,
			ListenPort:    port,
		},
	}

	node, err := NewP2PNode(ctx, config, registry)
	if err != nil {
		logger.Errorf("Failed to create P2P node: %v", err)
		return nil, err
	}

	logger.Infof("Legacy service setup completed successfully")
	return node.GetHost(), nil
}
