// github.com/trigg3rX/go-backend/pkg/network/discovery.go
package network

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "os"
    "time"

    "github.com/libp2p/go-libp2p/core/host"
    "github.com/libp2p/go-libp2p/core/peer"
    "github.com/multiformats/go-multiaddr"
)

const PeerInfoFilePath = "peer_info.json"
const PeerConnectionTimeout = 30 * time.Second

type PeerInfo struct {
    Name    string `json:"name"`
    Address string `json:"address"`
}

type Discovery struct {
    host    host.Host
    name    string
    context context.Context
}

func NewDiscovery(ctx context.Context, h host.Host, name string) *Discovery {
    return &Discovery{
        host:    h,
        name:    name,
        context: ctx,
    }
}

func (d *Discovery) SavePeerInfo() error {
    peerInfos := make(map[string]PeerInfo)

    if file, err := os.Open(PeerInfoFilePath); err == nil {
        decoder := json.NewDecoder(file)
        decoder.Decode(&peerInfos)
        file.Close()
    }

    fullAddr := fmt.Sprintf("%s/p2p/%s", d.host.Addrs()[0], d.host.ID().String())
    peerInfos[d.name] = PeerInfo{
        Name:    d.name,
        Address: fullAddr,
    }

    file, err := os.Create(PeerInfoFilePath)
    if err != nil {
        return fmt.Errorf("unable to create peer info file: %v", err)
    }
    defer file.Close()

    encoder := json.NewEncoder(file)
    return encoder.Encode(peerInfos)
}

func (d Discovery) ConnectToPeer(info PeerInfo) (peer.ID, error) {
    maddr, err := multiaddr.NewMultiaddr(info.Address)
    if err != nil {
        return "", fmt.Errorf("invalid peer address: %v", err)
    }

    peerInfo, err := peer.AddrInfoFromP2pAddr(maddr)
    if err != nil {
        return "", fmt.Errorf("invalid peer info: %v", err)
    }

    ctx, cancel := context.WithTimeout(d.context, PeerConnectionTimeout)
    defer cancel()

    err = d.host.Connect(ctx, *peerInfo)
    if err != nil {
        return "", fmt.Errorf("failed to connect to peer: %v", err)
    }

    log.Printf("Connected to peer %s with ID: %s", info.Name, peerInfo.ID)
    return peerInfo.ID, nil
}
