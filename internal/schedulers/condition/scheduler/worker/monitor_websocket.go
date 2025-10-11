package worker

import (
	"context"
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
	"sync"
	"time"
)

// WebSocketWorker monitors a user-specified WebSocket endpoint for events
// and notifies the scheduler when specific messages are received.
type WebSocketWorker struct {
	WebSocketConfig    *WebSocketConfig
	ConditionWorkerData *types.ConditionWorkerData // for JobID, limits, etc.
	Logger             logging.Logger
	Ctx                context.Context
	Cancel             context.CancelFunc
	IsActive           bool
	Mutex              sync.RWMutex
	TriggerCallback    WorkerTriggerCallback
	CleanupCallback    WorkerCleanupCallback
	LastMessage        []byte
}

// Start begins the WebSocketWorker's monitoring loop
func (w *WebSocketWorker) Start() {
	w.Mutex.Lock()
	w.IsActive = true
	w.Mutex.Unlock()

	w.Logger.Info("Starting WebSocket worker", "url", w.WebSocketConfig.URL, "job_id", w.ConditionWorkerData.JobID)
	conn, _, err := websocket.DefaultDialer.Dial(w.WebSocketConfig.URL, nil)
	if err != nil {
		w.Logger.Error("Failed to dial websocket", "error", err, "url", w.WebSocketConfig.URL)
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			w.Logger.Error("Failed to close websocket", "error", err)
		}
	}()

	for {
		select {
		case <-w.Ctx.Done():
			w.Logger.Info("WebSocket worker context canceled, stopping", "job_id", w.ConditionWorkerData.JobID)
			return
		default:
			_, message, err := conn.ReadMessage()
			if err != nil {
				w.Logger.Error("Error reading from websocket", "error", err)
				// You may want to trigger a reconnect/backoff here.
				return
			}
			w.LastMessage = message
			w.handleWebSocketMessage(message)
		}
	}
}

// handleWebSocketMessage processes incoming messages and triggers callback if an event is detected.
func (w *WebSocketWorker) handleWebSocketMessage(message []byte) {
	w.Logger.Info("WebSocket message received", "message", string(message), "job_id", w.ConditionWorkerData.JobID)

	// Implement logic to parse the message and decide if trigger is needed.
	var jsonMsg map[string]interface{}
	if err := json.Unmarshal(message, &jsonMsg); err != nil {
		w.Logger.Warn("WebSocket message not JSON, using raw string", "job_id", w.ConditionWorkerData.JobID)
	}

	// If your business requires filtering/matching, add it here. Otherwise, default: always trigger.
	notification := &TriggerNotification{
		JobID:       w.ConditionWorkerData.JobID,
		TriggeredAt: time.Now(),
	}
	if w.TriggerCallback != nil {
		if err := w.TriggerCallback(notification); err != nil {
			w.Logger.Error("Failed to notify scheduler from WebSocket", "error", err)
		}
	} else {
		w.Logger.Warn("No trigger callback for WebSocket worker", "job_id", w.ConditionWorkerData.JobID)
	}
}

// Stop gracefully stops the WebSocket worker
func (w *WebSocketWorker) Stop() {
	w.Mutex.Lock()
	defer w.Mutex.Unlock()
	if w.IsActive {
		w.Cancel()
		w.IsActive = false
		if w.CleanupCallback != nil {
			_ = w.CleanupCallback(w.ConditionWorkerData.JobID)
		}
		w.Logger.Info("WebSocket worker stopped", "job_id", w.ConditionWorkerData.JobID)
	}
}

// IsRunning returns whether the worker is currently running
func (w *WebSocketWorker) IsRunning() bool {
	w.Mutex.RLock()
	defer w.Mutex.RUnlock()
	return w.IsActive
}
