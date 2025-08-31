package websocket

import (
	"encoding/json"
	"time"
)

// MessageType represents the type of WebSocket message
type MessageType string

const (
	// Task-related message types
	MessageTypeTaskCreated       MessageType = "TASK_CREATED"
	MessageTypeTaskUpdated       MessageType = "TASK_UPDATED"
	MessageTypeTaskStatusChanged MessageType = "TASK_STATUS_CHANGED"
	MessageTypeTaskFeeUpdated    MessageType = "TASK_FEE_UPDATED"

	// System message types
	MessageTypeSubscribe   MessageType = "SUBSCRIBE"
	MessageTypeUnsubscribe MessageType = "UNSUBSCRIBE"
	MessageTypePing        MessageType = "PING"
	MessageTypePong        MessageType = "PONG"
	MessageTypeError       MessageType = "ERROR"
	MessageTypeSuccess     MessageType = "SUCCESS"
)

// Message represents a WebSocket message
type Message struct {
	Type      MessageType `json:"type"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
}

// TaskEventData represents task-related event data
type TaskEventData struct {
	TaskID    int64       `json:"task_id"`
	JobID     string      `json:"job_id"`
	UserID    string      `json:"user_id,omitempty"`
	Changes   interface{} `json:"changes,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// SubscriptionData represents subscription request data
type SubscriptionData struct {
	Room   string `json:"room"`
	JobID  string `json:"job_id,omitempty"`
	TaskID string `json:"task_id,omitempty"`
	UserID string `json:"user_id,omitempty"`
}

// ErrorData represents error message data
type ErrorData struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// SuccessData represents success message data
type SuccessData struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// NewMessage creates a new WebSocket message
func NewMessage(msgType MessageType, data interface{}) *Message {
	return &Message{
		Type:      msgType,
		Data:      data,
		Timestamp: time.Now(),
	}
}

// NewTaskEventMessage creates a new task event message
func NewTaskEventMessage(msgType MessageType, taskData *TaskEventData) *Message {
	return &Message{
		Type:      msgType,
		Data:      taskData,
		Timestamp: time.Now(),
	}
}

// NewErrorMessage creates a new error message
func NewErrorMessage(code, message string) *Message {
	return &Message{
		Type: MessageTypeError,
		Data: &ErrorData{
			Code:    code,
			Message: message,
		},
		Timestamp: time.Now(),
	}
}

// NewSuccessMessage creates a new success message
func NewSuccessMessage(message string, data interface{}) *Message {
	return &Message{
		Type: MessageTypeSuccess,
		Data: &SuccessData{
			Message: message,
			Data:    data,
		},
		Timestamp: time.Now(),
	}
}

// ToJSON converts message to JSON bytes
func (m *Message) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// FromJSON creates message from JSON bytes
func FromJSON(data []byte) (*Message, error) {
	var msg Message
	err := json.Unmarshal(data, &msg)
	return &msg, err
}
