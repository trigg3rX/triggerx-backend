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
	UserID   string          // Associated user ID for authentication
	APIKey   string          // API key for authentication
	LastPing time.Time
	mu       sync.RWMutex
	logger   logging.Logger
	ctx      context.Context
	cancel   context.CancelFunc
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

// ReadPump handles reading messages from the WebSocket connection
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	// Set read limits and timeouts
	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
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
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					c.logger.Errorf("WebSocket error for client %s: %v", c.ID, err)
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
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteJSON(message); err != nil {
				c.logger.Errorf("Error writing message to client %s: %v", c.ID, err)
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
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
		userID, ok := data["user_id"].(string)
		if !ok || userID == "" {
			return false
		}
		// For now, allow access if user_id matches (in production, validate against API key)
		return c.UserID == userID
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
