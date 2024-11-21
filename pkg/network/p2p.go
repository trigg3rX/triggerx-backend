package network

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/multiformats/go-multiaddr"
)

type P2PConfig struct {
	Name    string
	Address string
}

var KeeperConfigs = map[string]string{
	"Frodo":  "/ip4/127.0.0.1/tcp/3000",
	"Sam":    "/ip4/127.0.0.1/tcp/3001",
	"Merry":  "/ip4/127.0.0.1/tcp/3002",
	"Pippin": "/ip4/127.0.0.1/tcp/3003",
}

// SetupP2P creates and configures a libp2p host
func SetupP2P(ctx context.Context, config P2PConfig) (host.Host, error) {
	maddr, err := multiaddr.NewMultiaddr(config.Address)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %v", err)
	}

	h, err := libp2p.New(
		libp2p.ListenAddrs(maddr),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create host: %v", err)
	}

	return h, nil
}
