package keeper

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

    "github.com/libp2p/go-libp2p/core/peer"
	"github.com/trigg3rX/go-backend/pkg/network"
)

type Node struct {
	name      string
	messaging *network.Messaging
	discovery *network.Discovery
	peers     map[string]string // name -> peer ID
}

func NewNode(ctx context.Context, name string) (*Node, error) {
	addr, exists := network.KeeperConfigs[name]
	if !exists {
		return nil, fmt.Errorf("invalid keeper name")
	}

	config := network.P2PConfig{
		Name:    name,
		Address: addr,
	}

	host, err := network.SetupP2P(ctx, config)
	if err != nil {
		return nil, err
	}

	messaging := network.NewMessaging(host, name)
	discovery := network.NewDiscovery(ctx, host, name)

	node := &Node{
		name:      name,
		messaging: messaging,
		discovery: discovery,
		peers:     make(map[string]string),
	}

	messaging.InitMessageHandling(node.handleMessage)

	return node, nil
}

func (n *Node) handleMessage(msg network.Message) {
	prettyJSON, _ := json.MarshalIndent(msg, "", "  ")
	fmt.Printf("\nReceived message from %s:\n%s\n", msg.From, string(prettyJSON))
}

func (n *Node) Start() error {
	if err := n.discovery.SavePeerInfo(); err != nil {
		return err
	}

	go n.autoConnectToPeers()
	n.startMessageLoop()

	return nil
}

func (n *Node) autoConnectToPeers() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			file, err := os.Open(network.PeerInfoFilePath)
			if err != nil {
				continue
			}

			var peerInfos map[string]network.PeerInfo
			decoder := json.NewDecoder(file)
			err = decoder.Decode(&peerInfos)
			file.Close()
			if err != nil {
				continue
			}

			for name, info := range peerInfos {
				if name == n.name {
					continue
				}

				if _, exists := n.peers[name]; exists {
					continue
				}

				peerID, err := n.discovery.ConnectToPeer(info)
				if err != nil {
					log.Printf("Failed to connect to %s: %v", name, err)
					continue
				}

				n.peers[name] = peerID.String()
			}
		}
	}
}

func (n *Node) startMessageLoop() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Println("\nConnected peers:")
		for peerName := range n.peers {
			fmt.Printf("- %s\n", peerName)
		}

		fmt.Print("Enter keeper name to send message: ")
		if !scanner.Scan() {
			break
		}
		recipient := scanner.Text()

		testMessage := map[string]interface{}{
			"action":    "update",
			"data":      "test data",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"messageId": "123456789",
		}

		if peerIDStr, ok := n.peers[recipient]; ok {
			peerID, err := peer.Decode(peerIDStr)
			if err != nil {
				log.Printf("Error decoding peer ID: %v", err)
				continue
			}

			err = n.messaging.SendMessage(recipient, peerID, testMessage)
			if err != nil {
				log.Printf("Error sending message: %v", err)
			}
		} else {
			log.Printf("Peer %s not connected", recipient)
		}
	}
}