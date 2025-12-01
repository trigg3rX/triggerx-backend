package websocket

import (
	"context"
	"time"
)

// WebSocketClientInterface defines the interface for WebSocket operations
type WebSocketClientInterface interface {
	Connect(ctx context.Context) error
	ReadMessage(ctx context.Context) ([]byte, error)
	WriteMessage(ctx context.Context, messageType int, data []byte) error
	WriteTextMessage(ctx context.Context, data []byte) error
	WriteBinaryMessage(ctx context.Context, data []byte) error
	IsConnected() bool
	GetReconnectCount() int
	GetLastMessageTime() time.Time
	MessageChannel() <-chan []byte
	ErrorChannel() <-chan error
	Close() error
	GetConn() interface{} // Returns *websocket.Conn but using interface{} to avoid exposing dependency
}
