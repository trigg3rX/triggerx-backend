package websocket

import (
	"context"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

// Client represents a WebSocket client connection
type Client struct {
	ID        string
	Conn      *websocket.Conn
	Hub       *Hub
	Send      chan *Message
	Rooms     map[string]bool // Track which rooms this client is subscribed to
	UserID    string          // Associated user ID for authentication
	APIKey    string          // API key for authentication
	LastPing  time.Time
	mu        sync.RWMutex
	closeOnce sync.Once
	logger    observability.Logger
	ctx       context.Context
	cancel    context.CancelFunc
	// OnClose is called once when the client disconnects (read loop exits)
	OnClose func()
}

// NewClient creates a new WebSocket client
func NewClient(id string, conn *websocket.Conn, hub *Hub, logger observability.Logger) *Client {
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
func (c *Client) ReadPump(ctx context.Context) {
	defer func() {
		c.Hub.unregister <- c
		if err := c.Conn.Close(); err != nil {
			c.logger.Warn(ctx, "Error closing WebSocket for client", observability.String("client_id", c.ID), observability.Error(err))
		}
		if c.OnClose != nil {
			c.OnClose()
		}
	}()

	// Set read limits and timeouts
	c.Conn.SetReadLimit(512)
	if err := c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
		c.logger.Warn(ctx, "Failed to set read deadline for client", observability.String("client_id", c.ID), observability.Error(err))
	}
	c.Conn.SetPongHandler(func(string) error {
		if err := c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
			c.logger.Warn(ctx, "Failed to refresh read deadline on pong for client", observability.String("client_id", c.ID), observability.Error(err))
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
						// c.logger.Info(ctx, "WebSocket closed", observability.String("client_id", c.ID), observability.Int("code", ce.Code), observability.String("name", name), observability.String("text", ce.Text))
					default:
						c.logger.Error(ctx, "WebSocket closed unexpectedly", observability.String("client_id", c.ID), observability.Int("code", ce.Code), observability.String("name", name), observability.String("text", ce.Text))
					}
				} else if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					c.logger.Error(ctx, "WebSocket error", observability.String("client_id", c.ID), observability.Error(err))
				} else {
					c.logger.Info(ctx, "WebSocket read ended", observability.String("client_id", c.ID), observability.Error(err))
				}
				return
			}

			c.handleMessage(ctx, &msg)
		}
	}
}

// WritePump handles writing messages to the WebSocket connection
func (c *Client) WritePump(ctx context.Context) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		if err := c.Conn.Close(); err != nil {
			c.logger.Warn(ctx, "Error closing WebSocket", observability.String("client_id", c.ID), observability.Error(err))
		}
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if err := c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
				c.logger.Warn(ctx, "Failed to set write deadline", observability.String("client_id", c.ID), observability.Error(err))
			}
			if !ok {
				if err := c.Conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					c.logger.Warn(ctx, "Failed to write close message", observability.String("client_id", c.ID), observability.Error(err))
				}
				return
			}

			if err := c.Conn.WriteJSON(message); err != nil {
				c.logger.Error(ctx, "Error writing message", observability.String("client_id", c.ID), observability.Error(err))
				return
			}

		case <-ticker.C:
			if err := c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
				c.logger.Warn(ctx, "Failed to set write deadline (ping)", observability.String("client_id", c.ID), observability.Error(err))
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
func (c *Client) handleMessage(ctx context.Context, msg *Message) {
	switch msg.Type {
	case MessageTypeSubscribe:
		c.handleSubscribe(ctx, msg)
	case MessageTypeUnsubscribe:
		c.handleUnsubscribe(ctx, msg)
	case MessageTypePing:
		c.handlePing(ctx, msg)
	default:
		c.sendMessage(ctx, NewErrorMessage("INVALID_MESSAGE_TYPE", "Unknown message type"))
	}
}

// handleSubscribe processes subscription requests
func (c *Client) handleSubscribe(ctx context.Context, msg *Message) {
	data, ok := msg.Data.(map[string]interface{})
	if !ok {
		c.sendMessage(ctx, NewErrorMessage("INVALID_SUBSCRIPTION_DATA", "Invalid subscription data format"))
		return
	}

	room, ok := data["room"].(string)
	if !ok || room == "" {
		c.sendMessage(ctx, NewErrorMessage("INVALID_ROOM", "Room is required for subscription"))
		return
	}

	// Validate room format and permissions
	if !c.validateRoomAccess(room, data) {
		c.sendMessage(ctx, NewErrorMessage("ACCESS_DENIED", "Access denied to room"))
		return
	}

	c.mu.Lock()
	c.Rooms[room] = true
	c.mu.Unlock()

	c.Hub.subscribe <- &Subscription{
		Client: c,
		Room:   room,
	}

	c.sendMessage(ctx, NewSuccessMessage("Subscribed to room", map[string]string{"room": room}))
	c.logger.Info(ctx, "Client subscribed to room", observability.String("client_id", c.ID), observability.String("room", room))
}

// handleUnsubscribe processes unsubscription requests
func (c *Client) handleUnsubscribe(ctx context.Context, msg *Message) {
	data, ok := msg.Data.(map[string]interface{})
	if !ok {
		c.sendMessage(ctx, NewErrorMessage("INVALID_UNSUBSCRIPTION_DATA", "Invalid unsubscription data format"))
		return
	}

	room, ok := data["room"].(string)
	if !ok || room == "" {
		c.sendMessage(ctx, NewErrorMessage("INVALID_ROOM", "Room is required for unsubscription"))
		return
	}

	c.mu.Lock()
	delete(c.Rooms, room)
	c.mu.Unlock()

	c.Hub.unsubscribe <- &Subscription{
		Client: c,
		Room:   room,
	}

	c.sendMessage(ctx, NewSuccessMessage("Unsubscribed from room", map[string]string{"room": room}))
	c.logger.Info(ctx, "Client unsubscribed from room", observability.String("client_id", c.ID), observability.String("room", room))
}

// handlePing processes ping messages
func (c *Client) handlePing(ctx context.Context, msg *Message) {
	c.sendMessage(ctx, NewMessage(MessageTypePong, nil))
}

// validateRoomAccess validates if the client has access to the requested room
func (c *Client) validateRoomAccess(room string, data map[string]interface{}) bool {
	// Basic room format validation
	if len(room) < 3 {
		return false
	}

	// Check if it's a user-specific room
	if len(room) >= 5 && room[:5] == "user:" {
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
func (c *Client) sendMessage(ctx context.Context, msg *Message) {
	select {
	case c.Send <- msg:
	default:
		// Channel is full, close connection
		c.logger.Warn(ctx, "Client send channel is full, closing connection", observability.String("client_id", c.ID))
		c.Close()
	}
}

// Close closes the client connection safely (idempotent)
func (c *Client) Close() {
	c.closeOnce.Do(func() {
		c.cancel()
		close(c.Send)
	})
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
