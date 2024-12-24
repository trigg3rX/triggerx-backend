package eventbus

import (
	"log"
	"sync"

	"github.com/trigg3rX/triggerx-backend/pkg/events"
)

// EventHandler is a function that handles an event
type EventHandler func(event events.Event)

// EventBus manages event subscriptions and publications
type EventBus struct {
	handlers map[events.EventType][]EventHandler
	mu       sync.RWMutex
}

// New creates a new EventBus
func New() *EventBus {
	return &EventBus{
		handlers: make(map[events.EventType][]EventHandler),
	}
}

// Subscribe registers a handler for a specific event type
func (eb *EventBus) Subscribe(eventType events.EventType, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.handlers[eventType] = append(eb.handlers[eventType], handler)
	log.Printf("Subscribed to event type: %s", eventType)
}

// Publish sends an event to all subscribed handlers
func (eb *EventBus) Publish(event events.Event) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	if handlers, exists := eb.handlers[event.Type]; exists {
		for _, handler := range handlers {
			go func(h EventHandler) {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("Recovered from panic in event handler: %v", r)
					}
				}()
				h(event)
			}(handler)
		}
		log.Printf("Published event type: %s", event.Type)
	}
}

// Unsubscribe removes a handler for a specific event type
func (eb *EventBus) Unsubscribe(eventType events.EventType, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if handlers, exists := eb.handlers[eventType]; exists {
		for i, h := range handlers {
			if &h == &handler {
				eb.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
				break
			}
		}
	}
}
