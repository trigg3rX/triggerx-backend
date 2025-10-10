package websocket

import (
	"context"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// Client represents a WebSocket client connection
type Client struct {
	ID       string
	Conn     *websocket.Conn
	Hub      *Hub
	Send     chan *Message
	Rooms    map[string]bool // Track which rooms this client is subscribed to
	UserAddress   string          // Associated user address for authentication
	APIKey   string          // API key for authentication
	LastPing time.Time
	mu       sync.RWMutex
	logger   logging.Logger
	ctx      context.Context
	cancel   context.CancelFunc
	// OnClose is called once when the client disconnects (read loop exits)
	OnClose func()
}

// NewClient creates a new WebSocket client
func NewClient(id string, conn *websocket.Conn, hub *Hub, logger logging.Logger) *Client {
	ctx, cancel := context.WithCancel(context.Background())

	return &Client{
		ID:       id,
		Conn:     conn,
		Hub:      hub,
		Send:     make(chan *Message, 256),
		Rooms:    make(map[string]bool),
		LastPing: time.Now(),
		logger:   logger,
		ctx:      ctx,
		cancel:   cancel,
	}
}

// closeCodeName returns a human-readable name for a WebSocket close code
func closeCodeName(code int) string {
	switch code {
	case websocket.CloseNormalClosure:
		return "Normal Closure"
	case websocket.CloseGoingAway:
		return "Going Away"
	case websocket.CloseProtocolError:
		return "Protocol Error"
	case websocket.CloseUnsupportedData:
		return "Unsupported Data"
	case websocket.CloseNoStatusReceived:
		return "No Status Received"
	case websocket.CloseAbnormalClosure:
		return "Abnormal Closure"
	case websocket.CloseInvalidFramePayloadData:
		return "Invalid Frame Payload Data"
	case websocket.ClosePolicyViolation:
		return "Policy Violation"
	case websocket.CloseMessageTooBig:
		return "Message Too Big"
	case websocket.CloseMandatoryExtension:
		return "Mandatory Extension"
	case websocket.CloseInternalServerErr:
		return "Internal Server Error"
	case websocket.CloseServiceRestart:
		return "Service Restart"
	case websocket.CloseTryAgainLater:
		return "Try Again Later"
	case websocket.CloseTLSHandshake:
		return "TLS Handshake"
	default:
		return "Unknown"
	}
}

// ReadPump handles reading messages from the WebSocket connection
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		if err := c.Conn.Close(); err != nil {
			c.logger.Warnf("Error closing WebSocket for client %s: %v", c.ID, err)
		}
		if c.OnClose != nil {
			c.OnClose()
		}
	}()

	// Set read limits and timeouts
	c.Conn.SetReadLimit(512)
	if err := c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
		c.logger.Warnf("Failed to set read deadline for client %s: %v", c.ID, err)
	}
	c.Conn.SetPongHandler(func(string) error {
		if err := c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
			c.logger.Warnf("Failed to refresh read deadline on pong for client %s: %v", c.ID, err)
		}
		c.mu.Lock()
		c.LastPing = time.Now()
		c.mu.Unlock()
		return nil
	})

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			var msg Message
			err := c.Conn.ReadJSON(&msg)
			if err != nil {
				if ce, ok := err.(*websocket.CloseError); ok {
					name := closeCodeName(ce.Code)
					// Downgrade logging for normal/benign closes
					switch ce.Code {
					case websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived:
						c.logger.Infof("WebSocket closed for client %s: code=%d (%s), text=%s", c.ID, ce.Code, name, ce.Text)
					default:
						c.logger.Errorf("WebSocket closed unexpectedly for client %s: code=%d (%s), text=%s", c.ID, ce.Code, name, ce.Text)
					}
				} else if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					c.logger.Errorf("WebSocket error for client %s: %v", c.ID, err)
				} else {
					c.logger.Infof("WebSocket read ended for client %s: %v", c.ID, err)
				}
				return
			}

			c.handleMessage(&msg)
		}
	}
}

// WritePump handles writing messages to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		if err := c.Conn.Close(); err != nil {
			c.logger.Warnf("Error closing WebSocket for client %s: %v", c.ID, err)
		}
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if err := c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
				c.logger.Warnf("Failed to set write deadline for client %s: %v", c.ID, err)
			}
			if !ok {
				if err := c.Conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					c.logger.Warnf("Failed to write close message for client %s: %v", c.ID, err)
				}
				return
			}

			if err := c.Conn.WriteJSON(message); err != nil {
				c.logger.Errorf("Error writing message to client %s: %v", c.ID, err)
				return
			}

		case <-ticker.C:
			if err := c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
				c.logger.Warnf("Failed to set write deadline (ping) for client %s: %v", c.ID, err)
			}
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-c.ctx.Done():
			return
		}
	}
}

// handleMessage processes incoming messages from the client
func (c *Client) handleMessage(msg *Message) {
	switch msg.Type {
	case MessageTypeSubscribe:
		c.handleSubscribe(msg)
	case MessageTypeUnsubscribe:
		c.handleUnsubscribe(msg)
	case MessageTypePing:
		c.handlePing(msg)
	default:
		c.sendMessage(NewErrorMessage("INVALID_MESSAGE_TYPE", "Unknown message type"))
	}
}

// handleSubscribe processes subscription requests
func (c *Client) handleSubscribe(msg *Message) {
	data, ok := msg.Data.(map[string]interface{})
	if !ok {
		c.sendMessage(NewErrorMessage("INVALID_SUBSCRIPTION_DATA", "Invalid subscription data format"))
		return
	}

	room, ok := data["room"].(string)
	if !ok || room == "" {
		c.sendMessage(NewErrorMessage("INVALID_ROOM", "Room is required for subscription"))
		return
	}

	// Validate room format and permissions
	if !c.validateRoomAccess(room, data) {
		c.sendMessage(NewErrorMessage("ACCESS_DENIED", "Access denied to room"))
		return
	}

	c.mu.Lock()
	c.Rooms[room] = true
	c.mu.Unlock()

	c.Hub.subscribe <- &Subscription{
		Client: c,
		Room:   room,
	}

	c.sendMessage(NewSuccessMessage("Subscribed to room", map[string]string{"room": room}))
	c.logger.Infof("Client %s subscribed to room %s", c.ID, room)
}

// handleUnsubscribe processes unsubscription requests
func (c *Client) handleUnsubscribe(msg *Message) {
	data, ok := msg.Data.(map[string]interface{})
	if !ok {
		c.sendMessage(NewErrorMessage("INVALID_UNSUBSCRIPTION_DATA", "Invalid unsubscription data format"))
		return
	}

	room, ok := data["room"].(string)
	if !ok || room == "" {
		c.sendMessage(NewErrorMessage("INVALID_ROOM", "Room is required for unsubscription"))
		return
	}

	c.mu.Lock()
	delete(c.Rooms, room)
	c.mu.Unlock()

	c.Hub.unsubscribe <- &Subscription{
		Client: c,
		Room:   room,
	}

	c.sendMessage(NewSuccessMessage("Unsubscribed from room", map[string]string{"room": room}))
	c.logger.Infof("Client %s unsubscribed from room %s", c.ID, room)
}

// handlePing processes ping messages
func (c *Client) handlePing(msg *Message) {
	c.sendMessage(NewMessage(MessageTypePong, nil))
}

// validateRoomAccess validates if the client has access to the requested room
func (c *Client) validateRoomAccess(room string, data map[string]interface{}) bool {
	// Basic room format validation
	if len(room) < 3 {
		return false
	}

	// Check if it's a user-specific room
	if room[:5] == "user:" {
		userAddress, ok := data["user_address"].(string)
		if !ok || userAddress == "" {
			return false
		}
		// For now, allow access if user_address matches (in production, validate against API key)
		return c.UserAddress == userAddress
	}

	// For job and task rooms, allow access (in production, validate against API key permissions)
	return true
}

// sendMessage sends a message to the client
func (c *Client) sendMessage(msg *Message) {
	select {
	case c.Send <- msg:
	default:
		// Channel is full, close connection
		c.logger.Warnf("Client %s send channel is full, closing connection", c.ID)
		c.Close()
	}
}

// Close closes the client connection
func (c *Client) Close() {
	c.cancel()
	close(c.Send)
}

// IsInRoom checks if the client is subscribed to a specific room
func (c *Client) IsInRoom(room string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Rooms[room]
}

// GetRooms returns a copy of the client's subscribed rooms
func (c *Client) GetRooms() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	rooms := make([]string, 0, len(c.Rooms))
	for room := range c.Rooms {
		rooms = append(rooms, room)
	}
	return rooms
}
