package network

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/host/peerstore/pstoremem"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/multiformats/go-multiaddr"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type P2PConfig struct {
	Name    string
	Address peer.AddrInfo
}

type PeerIdentity struct {
	PrivKey []byte `json:"priv_key"`
}

type ServicePrivateKeys struct {
	ManagerPrivKey   string `json:"manager_p2p_private_key"`
	QuorumPrivKey    string `json:"quorum_p2p_private_key"`
	ValidatorPrivKey string `json:"validator_p2p_private_key"`
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
			privKeyStr = keys.ManagerPrivKey
		case ServiceQuorum:
			privKeyStr = keys.QuorumPrivKey
		case ServiceValidator:
			privKeyStr = keys.ValidatorPrivKey
		}

		privKeyBytes, err := crypto.ConfigDecodeKey(privKeyStr)
		if err != nil {
			return nil, fmt.Errorf("failed to decode private key: %w", err)
		}

		return crypto.UnmarshalPrivateKey(privKeyBytes)
	} else {
		identityFile := filepath.Join(BaseDataDir, RegistryDir, "keeper_identity.json")

		if data, err := os.ReadFile(identityFile); err == nil {
			var identity PeerIdentity
			if err := json.Unmarshal(data, &identity); err == nil {
				if priv, err := crypto.UnmarshalPrivateKey(identity.PrivKey); err == nil {
					return priv, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("invalid node type or missing private key")
}

func SetupServiceWithRegistry(ctx context.Context, serviceName string, registry *PeerRegistry) (host.Host, error) {
	priv, err := LoadIdentity(serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to setup peer identity: %w", err)
	}

	serviceInfo, exists := registry.GetService(serviceName)
	if !exists {
		return nil, fmt.Errorf("service %s not found in registry", serviceName)
	}

	listenAddr, err := multiaddr.NewMultiaddr(serviceInfo.Addresses[0])

	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}

	connMgr, err := connmgr.NewConnManager(
		200, 400,
		connmgr.WithGracePeriod(time.Minute),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection manager: %w", err)
	}

	promReg := prometheus.NewRegistry()
	ps, err := pstoremem.NewPeerstore()
	if err != nil {
		return nil, fmt.Errorf("failed to create peerstore: %w", err)
	}

	// Configure bootstrap peers based on node type
	var bootstrapPeers []peer.AddrInfo
	if serviceName != ServiceManager {
		managerInfo, exists := registry.GetService(ServiceManager)
		if exists && managerInfo.PeerID != "" {
			peerID, _ := peer.Decode(managerInfo.PeerID)
			for _, addr := range managerInfo.Addresses {
				maddr, _ := multiaddr.NewMultiaddr(addr)
				bootstrapPeers = append(bootstrapPeers, peer.AddrInfo{
					ID:    peerID,
					Addrs: []multiaddr.Multiaddr{maddr},
				})
				ps.AddAddrs(peerID, []multiaddr.Multiaddr{maddr}, time.Hour*168)
			}
		}
	}

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
		return nil, fmt.Errorf("failed to create host: %w", err)
	}

	return h, nil
}

func SetupKeeperWithRegistry(ctx context.Context, keeperConfig types.NodeConfig, registry *PeerRegistry) (host.Host, error) {
	priv, err := LoadIdentity(ServiceKeeper)
	if err != nil {
		return nil, fmt.Errorf("failed to setup peer identity: %w", err)
	}

	addr := fmt.Sprintf("/ip4/%s/tcp/%s", keeperConfig.ConnectionAddress, keeperConfig.P2pPort)
	listenAddr, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}

	connMgr, err := connmgr.NewConnManager(
		200, 400,
		connmgr.WithGracePeriod(time.Minute),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection manager: %w", err)
	}

	promReg := prometheus.NewRegistry()
	ps, err := pstoremem.NewPeerstore()
	if err != nil {
		return nil, fmt.Errorf("failed to create peerstore: %w", err)
	}

	services := registry.GetAllServices()

	var bootstrapPeers []peer.AddrInfo
	for _, service := range services {
		peerID, _ := peer.Decode(service.PeerID)
		maddr, _ := multiaddr.NewMultiaddr(service.Addresses[0])
		bootstrapPeers = append(bootstrapPeers, peer.AddrInfo{
			ID:    peerID,
			Addrs: []multiaddr.Multiaddr{maddr},
		})
		ps.AddAddrs(peerID, []multiaddr.Multiaddr{maddr}, time.Hour*168)
	}

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
		return nil, fmt.Errorf("failed to create host: %w", err)
	}

	return h, nil
}
