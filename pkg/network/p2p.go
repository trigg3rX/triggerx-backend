package network

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/multiformats/go-multiaddr"
)

type P2PConfig struct {
	Name    string
	Address string
}

type PeerIdentity struct {
	PrivKey []byte `json:"priv_key"`
}

type ServicePrivateKeys struct {
	ManagerPrivKey   string `json:"manager_p2p_private_key"`
	QuorumPrivKey    string `json:"quorum_p2p_private_key"`
	ValidatorPrivKey string `json:"validator_p2p_private_key"`
}

func LoadOrCreateIdentity(nodeType string) (crypto.PrivKey, error) {
	// For permanent nodes (manager, quorum, validator), load from services_privKeys.json
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
	}

	// For keeper nodes, use the existing identity file logic
	identityDir := filepath.Join(BaseDataDir, RegistryDir)
	if err := os.MkdirAll(identityDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create identity directory: %w", err)
	}

	identityFile := filepath.Join(identityDir, fmt.Sprintf("%s_identity.json", nodeType))

	if data, err := os.ReadFile(identityFile); err == nil {
		var identity PeerIdentity
		if err := json.Unmarshal(data, &identity); err == nil {
			if priv, err := crypto.UnmarshalPrivateKey(identity.PrivKey); err == nil {
				return priv, nil
			}
		}
	}

	// Generate new identity only for keeper nodes
	if nodeType == ServiceKeeper {
		priv, _, err := crypto.GenerateKeyPairWithReader(crypto.Ed25519, -1, rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("failed to generate key pair: %w", err)
		}

		privBytes, err := crypto.MarshalPrivateKey(priv)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal private key: %w", err)
		}

		identity := PeerIdentity{PrivKey: privBytes}
		identityJson, err := json.Marshal(identity)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal identity: %w", err)
		}

		if err := os.WriteFile(identityFile, identityJson, 0600); err != nil {
			return nil, fmt.Errorf("failed to save identity: %w", err)
		}

		return priv, nil
	}

	return nil, fmt.Errorf("invalid node type or missing private key")
}

func SetupP2PWithRegistry(ctx context.Context, config P2PConfig, registry *PeerRegistry) (host.Host, error) {
	// First, try to load existing identity
	priv, err := LoadOrCreateIdentity(config.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to setup peer identity: %w", err)
	}

	// Verify peer ID matches registry for permanent nodes (manager, validator, quorum)
	if config.Name != ServiceKeeper {
		if service, exists := registry.GetService(config.Name); exists {
			id, err := peer.IDFromPrivateKey(priv)
			if err != nil {
				return nil, fmt.Errorf("failed to get peer ID from private key: %w", err)
			}

			if service.PeerID == "" {
				// For first time setup
				if err := registry.UpdateService(config.Name, id, []string{config.Address}); err != nil {
					return nil, fmt.Errorf("failed to update registry with identity: %w", err)
				}
			} else if id.String() != service.PeerID {
				return nil, fmt.Errorf("loaded identity (%s) doesn't match registry (%s)", id.String(), service.PeerID)
			}
		}
	}

	maddr, err := multiaddr.NewMultiaddr(config.Address)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}

	connMgr, err := connmgr.NewConnManager(
		100, 400,
		connmgr.WithGracePeriod(time.Minute),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection manager: %w", err)
	}

	// Configure bootstrap peers based on node type
	var bootstrapPeers []peer.AddrInfo
	if config.Name != ServiceManager {
		// Non-manager nodes should bootstrap using manager's info
		if config.Name == ServiceKeeper {
			// Keepers bootstrap using quorum node
			quorumInfo, exists := registry.GetService(ServiceQuorum)
			if exists && quorumInfo.PeerID != "" {
				peerID, _ := peer.Decode(quorumInfo.PeerID)
				for _, addr := range quorumInfo.Addresses {
					maddr, _ := multiaddr.NewMultiaddr(addr)
					bootstrapPeers = append(bootstrapPeers, peer.AddrInfo{
						ID:    peerID,
						Addrs: []multiaddr.Multiaddr{maddr},
					})
				}
			}
		} else {
			// Validator and quorum bootstrap using manager
			managerInfo, exists := registry.GetService(ServiceManager)
			if exists && managerInfo.PeerID != "" {
				peerID, _ := peer.Decode(managerInfo.PeerID)
				for _, addr := range managerInfo.Addresses {
					maddr, _ := multiaddr.NewMultiaddr(addr)
					bootstrapPeers = append(bootstrapPeers, peer.AddrInfo{
						ID:    peerID,
						Addrs: []multiaddr.Multiaddr{maddr},
					})
				}
			}
		}
	}

	h, err := libp2p.New(
		libp2p.Identity(priv),
		libp2p.ListenAddrs(maddr),
		libp2p.ConnectionManager(connMgr),
		libp2p.EnableRelay(),
		libp2p.EnableAutoRelayWithStaticRelays(bootstrapPeers),
		libp2p.NATPortMap(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create host: %w", err)
	}

	return h, nil
}
