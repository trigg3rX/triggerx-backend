package main

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"

	"github.com/trigg3rX/triggerx-backend/pkg/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func getOutboundIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

func findFreePort(startPort int) (int, error) {
	for port := startPort; port < 65535; port++ {
		addr := fmt.Sprintf(":%d", port)
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			continue
		}
		listener.Close()
		return port, nil
	}
	return 0, fmt.Errorf("no free ports found")
}

func main() {
	// Get local IP address
	ip, err := getOutboundIP()
	if err != nil {
		fmt.Printf("Error getting IP address: %v\n", err)
		os.Exit(1)
	}

	// Find free port starting from 3000
	port, err := findFreePort(9003)
	if err != nil {
		fmt.Printf("Error finding free port: %v\n", err)
		os.Exit(1)
	}

	// Generate peer ID
	privKey, err := network.LoadOrCreateIdentity("keeper")
	if err != nil {
		fmt.Printf("Error generating peer ID: %v\n", err)
		os.Exit(1)
	}

	// Convert private key to peer ID
	peerID, err := peer.IDFromPrivateKey(privKey)
	if err != nil {
		fmt.Printf("Error converting private key to peer ID: %v\n", err)
		os.Exit(1)
	}

	// Read existing YAML
	yamlFile, err := os.ReadFile("triggerx_keeper.yaml")
	if err != nil {
		fmt.Printf("Error reading YAML file: %v\n", err)
		os.Exit(1)
	}

	// Parse YAML using NodeConfig struct
	var config types.NodeConfig
	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		fmt.Printf("Error parsing YAML: %v\n", err)
		os.Exit(1)
	}

	// Update connection_address and p2p_peer_id fields
	config.ConnectionAddress = ip
	config.P2pPeerId = peerID.String()
	config.P2pPort = strconv.Itoa(port)
	
	// Write back to file
	yamlData, err := yaml.Marshal(config)
	if err != nil {
		fmt.Printf("Error marshaling YAML: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile("config-files/triggerx_keeper.yaml", yamlData, 0644); err != nil {
		fmt.Printf("Error writing YAML file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully updated connection_address to %s and peer ID to %s\n",
		config.ConnectionAddress, config.P2pPeerId)
}
