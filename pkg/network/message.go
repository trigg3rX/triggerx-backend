package network

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

const MessageProtocol = "/keeper/message/1.0.0"

type Message struct {
	From       string      `json:"from"`
	To         string      `json:"to"`
	Content    interface{} `json:"content"`
	Type       string      `json:"type"`
	Timestamp  string      `json:"timestamp"`
	ID         string      `json:"id"`
	RetryCount int         `json:"retryCount"`
	ACK        bool        `json:"ack"`
}

type Messaging struct {
	host  host.Host
	name  string
	peers map[string]peer.ID
}

func NewMessaging(h host.Host, name string) *Messaging {
	return &Messaging{
		host:  h,
		name:  name,
		peers: make(map[string]peer.ID),
	}
}

func (m *Messaging) InitMessageHandling(onMessage func(Message)) {
	m.host.SetStreamHandler(protocol.ID(MessageProtocol), func(stream network.Stream) {
		m.handleStream(stream, onMessage)
	})
}

func (m *Messaging) GetHost() host.Host {
    return m.host
}

func (m *Messaging) handleStream(stream network.Stream, onMessage func(Message)) {
	reader := bufio.NewReader(stream)
	defer stream.Close()

	for {
		str, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				return
			}
			return
		}

		var msg Message
		if err := json.Unmarshal([]byte(str), &msg); err != nil {
			continue
		}

		m.peers[msg.From] = stream.Conn().RemotePeer()
		onMessage(msg)
	}
}

func (m *Messaging) SendMessage(to string, peerID peer.ID, content interface{}) error {
	msg := Message{
		From:      m.name,
		To:        to,
		Content:   content,
		Type:      "JSON_MESSAGE",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("error marshaling message: %w", err)
	}
	msgBytes = append(msgBytes, '\n')

	stream, err := m.host.NewStream(context.Background(), peerID, protocol.ID(MessageProtocol))
	if err != nil {
		return fmt.Errorf("error opening stream: %w", err)
	}
	defer stream.Close()

	if _, err = stream.Write(msgBytes); err != nil {
		return fmt.Errorf("error sending message: %w", err)
	}

	return nil
}

func (m *Messaging) SendMessageWithRetry(to string, peerID peer.ID, content interface{}, maxRetries int) error {
	// ... implement retry logic with exponential backoff ...
	return nil
}
