package websocket

import (
	"context"
	"sync"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
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

	logger logging.Logger
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
type InitialDataCallback func(room string, client *Client) error

// NewHub creates a new WebSocket hub
func NewHub(logger logging.Logger) *Hub {
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
func (h *Hub) Run() {
	h.logger.Info("Starting WebSocket hub")

	// Start task event processor
	go h.processTaskEvents()

	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case subscription := <-h.subscribe:
			h.subscribeToRoom(subscription)

		case subscription := <-h.unsubscribe:
			h.unsubscribeFromRoom(subscription)

		case broadcastMsg := <-h.broadcast:
			h.broadcastToRooms(broadcastMsg)

		case <-h.ctx.Done():
			h.logger.Info("WebSocket hub shutting down")
			return
		}
	}
}

// registerClient registers a new client
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client] = true
	h.logger.Infof("Client %s registered. Total clients: %d", client.ID, len(h.clients))
}

// unregisterClient unregisters a client
func (h *Hub) unregisterClient(client *Client) {
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

		h.logger.Infof("Client %s unregistered. Total clients: %d", client.ID, len(h.clients))
	}
}

// subscribeToRoom subscribes a client to a room
func (h *Hub) subscribeToRoom(subscription *Subscription) {
	h.mu.Lock()
	defer h.mu.Unlock()

	client := subscription.Client
	room := subscription.Room

	if h.rooms[room] == nil {
		h.rooms[room] = make(map[*Client]bool)
	}

	h.rooms[room][client] = true
	h.logger.Infof("Client %s subscribed to room %s", client.ID, room)

	// Call initial data callback if set
	if h.initialDataCallback != nil {
		go func() {
			if err := h.initialDataCallback(room, client); err != nil {
				h.logger.Errorf("Error fetching initial data for room %s: %v", room, err)
			}
		}()
	}
}

// unsubscribeFromRoom unsubscribes a client from a room
func (h *Hub) unsubscribeFromRoom(subscription *Subscription) {
	h.mu.Lock()
	defer h.mu.Unlock()

	client := subscription.Client
	room := subscription.Room

	if clients, exists := h.rooms[room]; exists {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.rooms, room)
		}
		h.logger.Infof("Client %s unsubscribed from room %s", client.ID, room)
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
func (h *Hub) processTaskEvents() {
	for {
		select {
		case taskEvent := <-h.taskEvents:
			h.broadcastTaskEvent(taskEvent)
		case <-h.ctx.Done():
			return
		}
	}
}

// broadcastTaskEvent broadcasts a task event to relevant rooms
func (h *Hub) broadcastTaskEvent(taskEvent *TaskEventData) {
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

		if taskEvent.UserAddress != "" {
			rooms = append(rooms, "user:"+taskEvent.UserAddress)
		}
	}

	message := NewTaskEventMessage(messageType, taskEvent)

	h.broadcast <- &BroadcastMessage{
		Message: message,
		Rooms:   rooms,
	}

	h.logger.Infof("Broadcasted task event %s for task %d to %d rooms", messageType, taskEvent.TaskID, len(rooms))
}

// BroadcastTaskCreated broadcasts a task created event
func (h *Hub) BroadcastTaskCreated(taskData *TaskEventData) {
	taskData.Timestamp = time.Now()
	select {
	case h.taskEvents <- taskData:
	default:
		h.logger.Warn("Task events channel is full, dropping task created event")
	}
}

// BroadcastTaskUpdated broadcasts a task updated event
func (h *Hub) BroadcastTaskUpdated(taskData *TaskEventData) {
	taskData.Timestamp = time.Now()
	select {
	case h.taskEvents <- taskData:
	default:
		h.logger.Warn("Task events channel is full, dropping task updated event")
	}
}

// BroadcastTaskStatusChanged broadcasts a task status changed event
func (h *Hub) BroadcastTaskStatusChanged(taskData *TaskEventData) {
	taskData.Timestamp = time.Now()
	select {
	case h.taskEvents <- taskData:
	default:
		h.logger.Warn("Task events channel is full, dropping task status changed event")
	}
}

// BroadcastTaskFeeUpdated broadcasts a task fee updated event
func (h *Hub) BroadcastTaskFeeUpdated(taskData *TaskEventData) {
	taskData.Timestamp = time.Now()
	select {
	case h.taskEvents <- taskData:
	default:
		h.logger.Warn("Task events channel is full, dropping task fee updated event")
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
func (h *Hub) Shutdown() {
	h.logger.Info("Shutting down WebSocket hub")
	h.cancel()

	// Close all client connections
	h.mu.Lock()
	for client := range h.clients {
		client.Close()
	}
	h.mu.Unlock()
}
