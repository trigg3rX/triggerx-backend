package cmd

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net"
	"os"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/trigg3rX/triggerx-backend/pkg/network"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func SetupCommand() *cli.Command {
	return &cli.Command{
		Name:  "setup",
		Usage: "Setup the keeper config file",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config",
				Usage:       "Path to the config file (<NAME>.yaml)",
				Value: "./triggerx_keeper.yaml",
			},
		},
		Action: setupKeeper,
	}
}

func setupKeeper(c *cli.Context) error {
	yamlFile, err := os.ReadFile(c.String("config"))
	if err != nil {
		fmt.Printf("Error reading YAML file at %s: %v\n", c.String("config"), err)
		os.Exit(1)
	}

	// Parse YAML using NodeConfig struct
	var config types.NodeConfig
	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		fmt.Printf("Error parsing YAML at %s: %v\n", c.String("config"), err)
		os.Exit(1)
	}

	// Get local IP address
	ip, err := getOutboundIP()
	if err != nil {
		fmt.Printf("Error getting IP address: %v\n", err)
		os.Exit(1)
	}

	privKey, _, err := crypto.GenerateKeyPairWithReader(crypto.Ed25519, -1, rand.Reader)
	if err != nil {
		fmt.Printf("failed to generate key pair: %v\n", err)
	}

	privBytes, err := crypto.MarshalPrivateKey(privKey)
	if err != nil {
		fmt.Printf("failed to marshal private key: %v\n", err)
	}

	identity := network.PeerIdentity{PrivKey: privBytes}
	identityJson, err := json.Marshal(identity)
	if err != nil {
		fmt.Printf("failed to marshal identity: %v\n", err)
	}

	if err := os.WriteFile("config-files/keeper_identity.json", identityJson, 0600); err != nil {
		fmt.Printf("failed to save identity: %v\n", err)
	}

	// Convert private key to peer ID
	peerID, err := peer.IDFromPrivateKey(privKey)
	if err != nil {
		fmt.Printf("Error converting private key to peer ID: %v\n", err)
		os.Exit(1)
	}

	// Update connection_address and p2p_peer_id fields
	config.ConnectionAddress = ip
	config.P2pPeerId = peerID.String()

	// Write back to file
	yamlData, err := yaml.Marshal(config)
	if err != nil {
		fmt.Printf("Error marshaling YAML: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile("config-files/"+c.String("config"), yamlData, 0644); err != nil {
		fmt.Printf("Error writing YAML file at %s: %v\n", c.String("config"), err)
		os.Exit(1)
	}

	fmt.Printf("Successfully updated connection_address to %s and peer ID to %s\n",
		config.ConnectionAddress, config.P2pPeerId)
	return nil
}

func getOutboundIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}
