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

func loadOrCreateIdentity(nodeType string) (crypto.PrivKey, error) {
	identityDir := "data/peer_registry"
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

func SetupP2PWithRegistry(ctx context.Context, config P2PConfig, registry *PeerRegistry) (host.Host, error) {
	// First, try to load existing identity
	priv, err := loadOrCreateIdentity(config.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to setup peer identity: %w", err)
	}

	// id, _ := peer.IDFromPrivateKey(priv)
	// fmt.Printf("Loaded identity with peer ID: %s\n", id.String())

	if service, exists := registry.GetService(config.Name); exists {

		if service.PeerID != "" {
			if id, err := peer.IDFromPrivateKey(priv); err == nil {
				if id.String() != service.PeerID {
					return nil, fmt.Errorf("loaded identity (%s) doesn't match registry (%s)", id.String(), service.PeerID)
				}
			}
		} else {
			if id, err := peer.IDFromPrivateKey(priv); err == nil {
				if err := registry.UpdateService(config.Name, id, []string{config.Address}); err != nil {
					return nil, fmt.Errorf("failed to update registry with identity: %w", err)
				}
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

	h, err := libp2p.New(
		libp2p.Identity(priv),
		libp2p.ListenAddrs(maddr),
		libp2p.ConnectionManager(connMgr),
		libp2p.EnableRelay(),
		libp2p.EnableAutoRelayWithStaticRelays([]peer.AddrInfo{}),
		libp2p.NATPortMap(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create host: %w", err)
	}

	return h, nil
}
