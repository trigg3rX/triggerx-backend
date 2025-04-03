package p2p

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"

	"github.com/trigg3rX/triggerx-backend/internal/manager/config"
)

const (
	ServiceManager   = "manager"
	ServicePerformer = "performer"
	ServiceValidator = "validator"

	BaseDataDir = "data"
	RegistryDir = "peer_registry"
)

type P2PConfig struct {
	Name    string
	Address peer.AddrInfo
}

// Add this new type to store host globally
type P2PHost struct {
	Host host.Host
}

var globalHost *P2PHost

func LoadIdentity() (crypto.PrivKey, error) {
	privKeyStr := config.P2PPrivateKey

	privKeyBytes, err := crypto.ConfigDecodeKey(privKeyStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	return crypto.UnmarshalPrivateKey(privKeyBytes)
}

func ConnectToAggregator() error {
	// If we already have a host, return
	if globalHost != nil {
		return nil
	}

	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		return fmt.Errorf("failed to load environment variables: %w", err)
	}

	// Initialize P2P host for manager
	priv, err := LoadIdentity()
	if err != nil {
		return fmt.Errorf("failed to load manager identity: %w", err)
	}

	// Create libp2p host
	h, err := libp2p.New(
		libp2p.Identity(priv),
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/9000"),
	)
	if err != nil {
		return fmt.Errorf("failed to create libp2p host: %w", err)
	}

	// Store host globally
	globalHost = &P2PHost{Host: h}

	targetIP := os.Getenv("AGGREGATOR_IP_ADDRESS")
	targetPort := os.Getenv("AGGREGATOR_P2P_PORT")

	// Parse aggregator multiaddr
	targetAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%s/p2p/12D3KooWBNFG1QjuF3UKAKvqhdXcxh9iBmj88cM5eU2EK5Pa91KB", targetIP, targetPort))
	if err != nil {
		return fmt.Errorf("failed to parse aggregator address: %w", err)
	}

	// Extract peer info
	info, err := peer.AddrInfoFromP2pAddr(targetAddr)
	if err != nil {
		return fmt.Errorf("failed to get peer info: %w", err)
	}

	// Connect to the aggregator
	if err := h.Connect(context.Background(), *info); err != nil {
		return fmt.Errorf("failed to connect to aggregator: %w", err)
	}

	return nil
}

func GetP2PHost() *P2PHost {
	return globalHost
}

// Add this new function to clean up the host
func CloseP2PHost() error {
	if globalHost != nil && globalHost.Host != nil {
		if err := globalHost.Host.Close(); err != nil {
			return fmt.Errorf("failed to close p2p host: %w", err)
		}
		globalHost = nil
	}
	return nil
}

func SendTaskToPerformer(jobID int64) error {
	return nil
}
