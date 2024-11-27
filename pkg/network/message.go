// github.com/trigg3rX/go-backend/pkg/network/message.go
package network

import (
    "bufio"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "time"

    "github.com/libp2p/go-libp2p/core/host"
    "github.com/libp2p/go-libp2p/core/network"
    "github.com/libp2p/go-libp2p/core/peer"
    "github.com/libp2p/go-libp2p/core/protocol"
)

const MessageProtocol = "/keeper/message/1.0.0"

type Message struct {
    From      string      `json:"from"`
    To        string      `json:"to"`
    Content   interface{} `json:"content"`
    Type      string      `json:"type"`
    Timestamp string      `json:"timestamp"`
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

func (m *Messaging) handleStream(stream network.Stream, onMessage func(Message)) {
    reader := bufio.NewReader(stream)
    defer stream.Close()

    remotePeerID := stream.Conn().RemotePeer()

    for {
        str, err := reader.ReadString('\n')
        if err != nil {
            if err != io.EOF {
                log.Printf("Error reading from stream: %v", err)
            }
            return
        }

        var msg Message
        if err := json.Unmarshal([]byte(str), &msg); err != nil {
            log.Printf("Error decoding message: %v", err)
            continue
        }

        m.peers[msg.From] = remotePeerID
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
        return fmt.Errorf("error marshaling message: %v", err)
    }
    msgBytes = append(msgBytes, '\n')

    stream, err := m.host.NewStream(context.Background(), peerID, protocol.ID(MessageProtocol))
    if err != nil {
        return fmt.Errorf("error opening stream: %v", err)
    }
    defer stream.Close()

    _, err = stream.Write(msgBytes)
    if err != nil {
        return fmt.Errorf("error sending message: %v", err)
    }

    prettyJSON, err := json.MarshalIndent(msg, "", "  ")
    if err != nil {
        log.Printf("Error formatting message: %v", err)
    } else {
        fmt.Printf("\nSent message to %s:\n%s\n", to, string(prettyJSON))
    }

    return nil
}