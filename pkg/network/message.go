package network

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"


	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

const MessageProtocol = "/triggerx/message/1.0.0"

type Message struct {
	From       string      `json:"from"`
	To         string      `json:"to"`
	Content    interface{} `json:"content"`
	Timestamp  string      `json:"timestamp"`
	ID         string      `json:"id"`
	RetryCount int         `json:"retryCount"`
	ACK        bool        `json:"ack"`
}

type Messaging struct {
	host   host.Host
	name   string
	peers  map[string]peer.ID
	logger logging.Logger
}

func NewMessaging(h host.Host, name string) *Messaging {
	// Initialize logger for messaging
	logger := logging.GetLogger(logging.Development, logging.ProcessName(name))

	return &Messaging{
		host:   h,
		name:   name,
		peers:  make(map[string]peer.ID),
		logger: logger,
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

		m.logger.Infof("MSG RCV - From: %s, Content: %v ID: %s", msg.From, msg.Content, msg.ID)

		onMessage(msg)
	}
}

func (m *Messaging) SendMessage(to string, peerID peer.ID, content interface{}, ack bool) error {
	msg := Message{
		From:       m.name,
		To:         to,
		Content:    content,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		ID:         strings.ReplaceAll(uuid.New().String()[:13], "-", ""),
		RetryCount: 3,
		ACK:        ack,
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("error marshaling message: %w", err)
	}
	msgBytes = append(msgBytes, '\n')

	var lastErr error
	for i := 0; i < msg.RetryCount; i++ {
		stream, err := m.host.NewStream(context.Background(), peerID, protocol.ID(MessageProtocol))
		if err != nil {
			lastErr = fmt.Errorf("error opening stream: %w", err)
			m.logger.Warnf("Retry %d/%d: Failed to open stream: %v", i+1, msg.RetryCount, err)
			time.Sleep(time.Second * time.Duration(1<<uint(i))) // Exponential backoff
			continue
		}
		defer stream.Close()

		if _, err = stream.Write(msgBytes); err != nil {
			lastErr = fmt.Errorf("error sending message: %w", err)
			m.logger.Warnf("Retry %d/%d: Failed to write message: %v", i+1, msg.RetryCount, err)
			time.Sleep(time.Second * time.Duration(1<<uint(i))) // Exponential backoff
			continue
		}

		// Log successful sent message
		m.logger.Infof("MSG SNT - To: %s, Content: %v, ID: %s",	to,	msg.Content, msg.ID)

		return nil // Success
	}

	return fmt.Errorf("failed after %d retries: %w", msg.RetryCount, lastErr)
}

func (m *Messaging) BroadcastMessage(content interface{}) error {

	connectedPeers := m.host.Network().Peers()
	presentTime := time.Now().UTC().Format(time.RFC3339)
	broadcastID := strings.ReplaceAll(uuid.New().String()[:13], "-", "")

	for _, peerID := range connectedPeers {
		msg := Message{
			From:       m.name,
			Content:    content,
			Timestamp:  presentTime,
			ID:         broadcastID,
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
	}
	m.logger.Infof("MSG BRC - %d Peers, Content: %v, ID: %s", len(connectedPeers), content, broadcastID)

	return nil
}