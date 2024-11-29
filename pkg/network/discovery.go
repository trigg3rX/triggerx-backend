package network

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "sync"
    "time"

    "github.com/libp2p/go-libp2p/core/host"
    "github.com/libp2p/go-libp2p/core/peer"
    "github.com/multiformats/go-multiaddr"
)

const (
    PeerInfoFilePath = "/home/prapti_shah/peer_info.json"
    PeerConnectionTimeout = 30 * time.Second
)

type PeerInfo struct {
    Name    string `json:"name"`
    Address string `json:"address"`
}

type Discovery struct {
    host    host.Host
    name    string
    context context.Context
    mutex   sync.RWMutex
}

func NewDiscovery(ctx context.Context, h host.Host, name string) *Discovery {
    return &Discovery{
        host:    h,
        name:    name,
        context: ctx,
    }
}

// expandHomePath expands the ~ to the user's home directory
func expandHomePath(path string) (string, error) {
    if len(path) > 1 && path[:2] == "~/" {
        homeDir, err := os.UserHomeDir()
        if err != nil {
            return "", err
        }
        path = filepath.Join(homeDir, path[2:])
    }
    return path, nil
}

// SavePeerInfo saves peer information to a JSON file
func (d *Discovery) SavePeerInfo() error {
    d.mutex.Lock()
    defer d.mutex.Unlock()

    filePath, err := expandHomePath(PeerInfoFilePath)
    if err != nil {
        return fmt.Errorf("error expanding file path: %v", err)
    }

    // Ensure directory exists
    if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
        return fmt.Errorf("unable to create directory: %v", err)
    }

    

    // Read existing peer infos
    peerInfos := make(map[string]PeerInfo)
    
    // Try to read existing file
    if existingFile, err := os.Open(filePath); err == nil {
        defer existingFile.Close()
        decoder := json.NewDecoder(existingFile)
        if err := decoder.Decode(&peerInfos); err != nil {
            log.Printf("Warning: existing peer info file could not be decoded: %v", err)
        }
    }

    // Create full address for current host
    var fullAddrs []string
    for _, addr := range d.host.Addrs() {
        fullAddr := fmt.Sprintf("%s/p2p/%s", addr, d.host.ID().String())
        fullAddrs = append(fullAddrs, fullAddr)
    }
    
    // Add or update peer info
    peerInfos[d.name] = PeerInfo{
        Name:    d.name,
        Address: strings.Join(fullAddrs, ","), // Store multiple addresses
    }
    // Write updated peer infos
    file, err := os.Create(filePath)
    if err != nil {
        return fmt.Errorf("unable to create peer info file: %v", err)
    }
    defer file.Close()

    encoder := json.NewEncoder(file)
    if err := encoder.Encode(peerInfos); err != nil {
        return fmt.Errorf("error writing peer info: %v", err)
    }

    log.Printf("Peer info saved to %s", filePath)
    return nil
}

// LoadPeerInfo loads peer information from the JSON file
func LoadPeerInfo() (map[string]PeerInfo, error) {
    filePath, err := expandHomePath(PeerInfoFilePath)
    if err != nil {
        return nil, fmt.Errorf("error expanding file path: %v", err)
    }

    // If file doesn't exist, return an empty map
    if _, err := os.Stat(filePath); os.IsNotExist(err) {
        return make(map[string]PeerInfo), nil
    }

    file, err := os.Open(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to open peer info file: %v", err)
    }
    defer file.Close()

    var peerInfos map[string]PeerInfo
    decoder := json.NewDecoder(file)
    if err := decoder.Decode(&peerInfos); err != nil {
        return nil, fmt.Errorf("failed to decode peer info: %v", err)
    }

    return peerInfos, nil
}

// ConnectToPeer connects to a specific peer
func (d *Discovery) ConnectToPeer(info PeerInfo) (peer.ID, error) {
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