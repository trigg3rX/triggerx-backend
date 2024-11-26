// github.com/trigg3rX/go-backend/pkg/network/discovery.go
package network

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "os"
    // "strings"
    "time"

    "github.com/libp2p/go-libp2p/core/host"
    "github.com/libp2p/go-libp2p/core/peer"
    "github.com/multiformats/go-multiaddr"
)

const (
    PeerInfoFilePath         = "peer_info.json"
    PeerConnectionTimeout    = 30 * time.Second
    DiscoveryLogPrefix       = "[DISCOVERY] "
)

type PeerInfo struct {
    Name    string `json:"name"`
    Address string `json:"address"`
}

type Discovery struct {
    host           host.Host
    name           string
    context        context.Context
    peerStore      map[peer.ID]peer.AddrInfo
    staticPeerList map[string]string
}

func NewDiscovery(ctx context.Context, h host.Host, name string) *Discovery {
    return &Discovery{
        host:           h,
        name:           name,
        context:        ctx,
        peerStore:      make(map[peer.ID]peer.AddrInfo),
        staticPeerList: make(map[string]string),
    }
}

// AddStaticPeer allows manually adding a peer for connection
func (d *Discovery) AddStaticPeer(name, address string) {
    d.staticPeerList[name] = address
    log.Printf("%sAdded static peer: %s - %s", DiscoveryLogPrefix, name, address)
}

func (d *Discovery) SavePeerInfo() error {
    peerInfos := make(map[string]PeerInfo)

    // Read existing peer info
    if file, err := os.Open(PeerInfoFilePath); err == nil {
        decoder := json.NewDecoder(file)
        decoder.Decode(&peerInfos)
        file.Close()
    }

    // Ensure we have at least one address
    if len(d.host.Addrs()) == 0 {
        log.Printf("%sWarning: No addresses available for this host", DiscoveryLogPrefix)
        return nil
    }

    fullAddr := fmt.Sprintf("%s/p2p/%s", d.host.Addrs()[0], d.host.ID().String())
    peerInfos[d.name] = PeerInfo{
        Name:    d.name,
        Address: fullAddr,
    }

    // Add static peers to file
    for name, addr := range d.staticPeerList {
        peerInfos[name] = PeerInfo{
            Name:    name,
            Address: addr,
        }
    }

    file, err := os.Create(PeerInfoFilePath)
    if err != nil {
        return fmt.Errorf("%sunable to create peer info file: %v", DiscoveryLogPrefix, err)
    }
    defer file.Close()

    encoder := json.NewEncoder(file)
    return encoder.Encode(peerInfos)
}

func (d *Discovery) ConnectToPeer(info PeerInfo) (*peer.ID, error) {
    log.Printf("%sTrying to connect to peer: %s - %s", DiscoveryLogPrefix, info.Name, info.Address)
    
    maddr, err := multiaddr.NewMultiaddr(info.Address)
    if err != nil {
        return nil, fmt.Errorf("%sinvalid peer address: %v", DiscoveryLogPrefix, err)
    }

    peerInfo, err := peer.AddrInfoFromP2pAddr(maddr)
    if err != nil {
        return nil, fmt.Errorf("%sinvalid peer info: %v", DiscoveryLogPrefix, err)
    }

    ctx, cancel := context.WithTimeout(d.context, PeerConnectionTimeout)
    defer cancel()

    err = d.host.Connect(ctx, *peerInfo)
    if err != nil {
        log.Printf("%sFailed to connect to peer %s: %v", DiscoveryLogPrefix, info.Name, err)
        return nil, err
    }

    log.Printf("%sSuccessfully connected to peer %s with ID: %s", 
        DiscoveryLogPrefix, info.Name, peerInfo.ID)
    
    d.peerStore[peerInfo.ID] = *peerInfo
    return &peerInfo.ID, nil
}

func (d *Discovery) FindPeers() error {
    peerInfos := make(map[string]PeerInfo)
    
    // Try to read from file
    if file, err := os.Open(PeerInfoFilePath); err == nil {
        decoder := json.NewDecoder(file)
        decoder.Decode(&peerInfos)
        file.Close()
    }

    // Add static peers if defined
    for name, addr := range d.staticPeerList {
        peerInfos[name] = PeerInfo{Name: name, Address: addr}
    }

    log.Printf("%sDiscovering peers. Total potential peers: %d", 
        DiscoveryLogPrefix, len(peerInfos))

    var successfulConnections int
    for name, info := range peerInfos {
        if _, err := d.ConnectToPeer(info); err == nil {
            successfulConnections++
        } else {
            log.Printf("%sFailed to connect to %s: %v", 
                DiscoveryLogPrefix, name, err)
        }
    }

    log.Printf("%sCompleted peer discovery. Successful connections: %d", 
        DiscoveryLogPrefix, successfulConnections)

    return nil
}

func (d *Discovery) ConnectToPeerByName(name string) (*peer.ID, error) {
    peerInfos := make(map[string]PeerInfo)
    
    // Read from file
    if file, err := os.Open(PeerInfoFilePath); err == nil {
        decoder := json.NewDecoder(file)
        decoder.Decode(&peerInfos)
        file.Close()
    }

    // Check static peer list
    if staticAddr, exists := d.staticPeerList[name]; exists {
        peerInfos[name] = PeerInfo{Name: name, Address: staticAddr}
    }

    // Try connection
    if info, exists := peerInfos[name]; exists {
        return d.ConnectToPeer(info)
    }

    return nil, fmt.Errorf("%speer %s not found in peer info or static list", 
        DiscoveryLogPrefix, name)
}

// Getter for Host to allow manual connection if needed
func (d *Discovery) Host() host.Host {
    return d.host
}