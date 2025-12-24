package websocket

import (
	"context"
	"sync"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Inbound messages from the clients
	broadcast chan *BroadcastMessage

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Room subscriptions
	rooms map[string]map[*Client]bool

	// Subscription requests
	subscribe chan *Subscription

	// Unsubscription requests
	unsubscribe chan *Subscription

	// Task event channel
	taskEvents chan *TaskEventData

	// Initial data callback
	initialDataCallback InitialDataCallback

	// Context for graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc

	// Mutex for thread safety
	mu sync.RWMutex

	logger observability.Logger
}

// BroadcastMessage represents a message to be broadcasted to specific rooms
type BroadcastMessage struct {
	Message *Message
	Rooms   []string
}

// Subscription represents a client subscription to a room
type Subscription struct {
	Client *Client
	Room   string
}

// InitialDataCallback is a function type for fetching initial data when subscribing to a room
type InitialDataCallback func(ctx context.Context, room string, client *Client) error

// NewHub creates a new WebSocket hub
func NewHub(logger observability.Logger) *Hub {
	ctx, cancel := context.WithCancel(context.Background())

	return &Hub{
		clients:     make(map[*Client]bool),
		broadcast:   make(chan *BroadcastMessage, 256),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		rooms:       make(map[string]map[*Client]bool),
		subscribe:   make(chan *Subscription),
		unsubscribe: make(chan *Subscription),
		taskEvents:  make(chan *TaskEventData, 256),
		ctx:         ctx,
		cancel:      cancel,
		logger:      logger,
	}
}

// SetInitialDataCallback sets the callback function for fetching initial data
func (h *Hub) SetInitialDataCallback(callback InitialDataCallback) {
	h.initialDataCallback = callback
}

// Run starts the hub's main loop
func (h *Hub) Run(ctx context.Context) {
	h.logger.Info(ctx, "Starting WebSocket hub")

	// Start task event processor
	go h.processTaskEvents(ctx)

	for {
		select {
		case client := <-h.register:
			h.registerClient(ctx, client)

		case client := <-h.unregister:
			h.unregisterClient(ctx, client)

		case subscription := <-h.subscribe:
			h.subscribeToRoom(ctx, subscription)

		case subscription := <-h.unsubscribe:
			h.unsubscribeFromRoom(ctx, subscription)

		case broadcastMsg := <-h.broadcast:
			h.broadcastToRooms(broadcastMsg)

		case <-h.ctx.Done():
			// h.logger.Info(ctx, "WebSocket hub shutting down")
			return
		}
	}
}

// registerClient registers a new client
func (h *Hub) registerClient(ctx context.Context, client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client] = true
	// h.logger.Info(ctx, "Client registered", observability.String("client_id", client.ID), observability.Int("total_clients", len(h.clients)))
}

// unregisterClient unregisters a client
func (h *Hub) unregisterClient(ctx context.Context, client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.Send)

		// Remove client from all rooms
		for room, clients := range h.rooms {
			if clients[client] {
				delete(clients, client)
				if len(clients) == 0 {
					delete(h.rooms, room)
				}
			}
		}

		// h.logger.Info(ctx, "Client unregistered", observability.String("client_id", client.ID), observability.Int("total_clients", len(h.clients)))
	}
}

// subscribeToRoom subscribes a client to a room
func (h *Hub) subscribeToRoom(ctx context.Context, subscription *Subscription) {
	h.mu.Lock()
	defer h.mu.Unlock()

	client := subscription.Client
	room := subscription.Room

	if h.rooms[room] == nil {
		h.rooms[room] = make(map[*Client]bool)
	}

	h.rooms[room][client] = true
	// h.logger.Info(ctx, "Client subscribed to room", observability.String("client_id", client.ID), observability.String("room", room))

	// Call initial data callback if set
	if h.initialDataCallback != nil {
		go func() {
			if err := h.initialDataCallback(ctx, room, client); err != nil {
				h.logger.Error(ctx, "Error fetching initial data for room", observability.String("room", room), observability.Error(err))
			}
		}()
	}
}

// unsubscribeFromRoom unsubscribes a client from a room
func (h *Hub) unsubscribeFromRoom(ctx context.Context, subscription *Subscription) {
	h.mu.Lock()
	defer h.mu.Unlock()

	client := subscription.Client
	room := subscription.Room

	if clients, exists := h.rooms[room]; exists {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.rooms, room)
		}
		// h.logger.Info(ctx, "Client unsubscribed from room", observability.String("client_id", client.ID), observability.String("room", room))
	}
}

// broadcastToRooms broadcasts a message to specific rooms
func (h *Hub) broadcastToRooms(broadcastMsg *BroadcastMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	message := broadcastMsg.Message
	rooms := broadcastMsg.Rooms

	// If no rooms specified, broadcast to all clients
	if len(rooms) == 0 {
		for client := range h.clients {
			select {
			case client.Send <- message:
			default:
				close(client.Send)
				delete(h.clients, client)
			}
		}
		return
	}

	// Broadcast to specific rooms
	for _, room := range rooms {
		if clients, exists := h.rooms[room]; exists {
			for client := range clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
					delete(clients, client)
				}
			}
		}
	}
}

// processTaskEvents processes task events and broadcasts them to appropriate rooms
func (h *Hub) processTaskEvents(ctx context.Context) {
	for {
		select {
		case taskEvent := <-h.taskEvents:
			h.broadcastTaskEvent(ctx, taskEvent)
		case <-h.ctx.Done():
			return
		}
	}
}

// broadcastTaskEvent broadcasts a task event to relevant rooms
func (h *Hub) broadcastTaskEvent(ctx context.Context, taskEvent *TaskEventData) {
	var messageType MessageType
	var rooms []string

	// Determine message type and target rooms based on event
	switch {
	case taskEvent.TaskID != 0:
		messageType = MessageTypeTaskUpdated
		rooms = []string{
			"task:" + string(rune(taskEvent.TaskID)),
			"job:" + taskEvent.JobID,
		}

		if taskEvent.UserID != "" {
			rooms = append(rooms, "user:"+taskEvent.UserID)
		}
	}

	message := NewTaskEventMessage(messageType, taskEvent)

	h.broadcast <- &BroadcastMessage{
		Message: message,
		Rooms:   rooms,
	}

	h.logger.Info(ctx, "Broadcasted task event", observability.String("message_type", string(messageType)), observability.Int64("task_id", taskEvent.TaskID), observability.Int("total_rooms", len(rooms)))
}

// BroadcastTaskCreated broadcasts a task created event
func (h *Hub) BroadcastTaskCreated(ctx context.Context, taskData *TaskEventData) {
	taskData.Timestamp = time.Now()
	select {
	case h.taskEvents <- taskData:
	default:
		h.logger.Warn(ctx, "Task events channel is full, dropping task created event")
	}
}

// BroadcastTaskUpdated broadcasts a task updated event
func (h *Hub) BroadcastTaskUpdated(ctx context.Context, taskData *TaskEventData) {
	taskData.Timestamp = time.Now()
	select {
	case h.taskEvents <- taskData:
	default:
		h.logger.Warn(ctx, "Task events channel is full, dropping task updated event")
	}
}

// BroadcastTaskStatusChanged broadcasts a task status changed event
func (h *Hub) BroadcastTaskStatusChanged(ctx context.Context, taskData *TaskEventData) {
	taskData.Timestamp = time.Now()
	select {
	case h.taskEvents <- taskData:
	default:
		h.logger.Warn(ctx, "Task events channel is full, dropping task status changed event")
	}
}

// BroadcastTaskFeeUpdated broadcasts a task fee updated event
func (h *Hub) BroadcastTaskFeeUpdated(ctx context.Context, taskData *TaskEventData) {
	taskData.Timestamp = time.Now()
	select {
	case h.taskEvents <- taskData:
	default:
		h.logger.Warn(ctx, "Task events channel is full, dropping task fee updated event")
	}
}

// GetStats returns hub statistics
func (h *Hub) GetStats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return map[string]interface{}{
		"total_clients": len(h.clients),
		"total_rooms":   len(h.rooms),
		"rooms":         h.getRoomStats(),
	}
}

// getRoomStats returns statistics for each room
func (h *Hub) getRoomStats() map[string]int {
	roomStats := make(map[string]int)
	for room, clients := range h.rooms {
		roomStats[room] = len(clients)
	}
	return roomStats
}

// Shutdown gracefully shuts down the hub
func (h *Hub) Shutdown(ctx context.Context) {
	h.logger.Info(ctx, "Shutting down WebSocket hub")
	h.cancel()

	// Close all client connections
	h.mu.Lock()
	for client := range h.clients {
		client.Close()
	}
	h.mu.Unlock()
}
