package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// WebSocketWorker monitors a user-specified WebSocket endpoint for events
// and notifies the scheduler when specific messages are received.
type WebSocketWorker struct {
	WebSocketConfig     *WebSocketConfig
	ConditionWorkerData *types.ConditionWorkerData // for JobID, limits, etc.
	Logger              logging.Logger
	Ctx                 context.Context
	Cancel              context.CancelFunc
	IsActive            bool
	Mutex               sync.RWMutex
	TriggerCallback     WorkerTriggerCallback
	CleanupCallback     WorkerCleanupCallback
	LastMessage         []byte
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
			w.Logger.Error("Failed to close websocket connection", "error", err)
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

	// Try numeric condition-style evaluation as in monitor_condition
	value, extractedErr := w.extractNumericValueFromMessage(message)
	if extractedErr != nil {
		w.Logger.Warn("Could not extract numeric value from message", "error", extractedErr, "job_id", w.ConditionWorkerData.JobID)
		return
	}

	satisfied, evalErr := w.evaluateCondition(value)
	if evalErr != nil {
		w.Logger.Error("Failed to evaluate websocket condition", "error", evalErr, "job_id", w.ConditionWorkerData.JobID)
		return
	}

	if !satisfied {
		w.Logger.Debug("WebSocket condition not satisfied", "job_id", w.ConditionWorkerData.JobID, "current_value", value, "condition_type", w.ConditionWorkerData.ConditionType)
		return
	}

	notification := &TriggerNotification{
		JobID:        w.ConditionWorkerData.JobID.ToBigInt(),
		TriggerValue: value,
		TriggeredAt:  time.Now(),
	}
	if w.TriggerCallback != nil {
		if err := w.TriggerCallback(notification); err != nil {
			w.Logger.Error("Failed to notify scheduler from WebSocket", "error", err)
		} else {
			w.Logger.Info("WebSocket condition satisfied and scheduler notified", "job_id", w.ConditionWorkerData.JobID, "trigger_value", value)
		}
	} else {
		w.Logger.Warn("No trigger callback for WebSocket worker", "job_id", w.ConditionWorkerData.JobID)
	}

	if !w.ConditionWorkerData.Recurring {
		w.Logger.Info("Non-recurring websocket job triggered, stopping worker", "job_id", w.ConditionWorkerData.JobID)
		go w.Stop()
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
			_ = w.CleanupCallback(w.ConditionWorkerData.JobID.ToBigInt())
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

// extractNumericValueFromMessage mirrors monitor_condition's extraction logic using SelectedKeyRoute and ValueResponse
func (w *WebSocketWorker) extractNumericValueFromMessage(body []byte) (float64, error) {
	// If key path is specified, use it to extract the value
	if w.ConditionWorkerData.SelectedKeyRoute != "" {
		if value, err := w.extractValueByKeyPath(body, w.ConditionWorkerData.SelectedKeyRoute); err == nil {
			return value, nil
		}
	}

	// Try to parse as ValueResponse
	var valueResp ValueResponse
	if err := json.Unmarshal(body, &valueResp); err == nil {
		if valueResp.SelectedKeyRoute != "" {
			if value, err := w.extractValueByKeyPath(body, w.ConditionWorkerData.SelectedKeyRoute); err == nil {
				return value, nil
			}
		}
		if valueResp.Value != 0 {
			return valueResp.Value, nil
		}
		if valueResp.Price != 0 {
			return valueResp.Price, nil
		}
		if valueResp.USD != 0 {
			return valueResp.USD, nil
		}
		if valueResp.Rate != 0 {
			return valueResp.Rate, nil
		}
		if valueResp.Result != 0 {
			return valueResp.Result, nil
		}
		if valueResp.Data != 0 {
			return valueResp.Data, nil
		}
	}

	// Fallback to direct parsing like monitor_condition.parseDirectValue
	var floatValue float64
	if err := json.Unmarshal(body, &floatValue); err == nil {
		return floatValue, nil
	}
	var stringValue string
	if err := json.Unmarshal(body, &stringValue); err == nil {
		if floatVal, parseErr := strconv.ParseFloat(stringValue, 64); parseErr == nil {
			return floatVal, nil
		}
	}

	return 0, fmt.Errorf("could not extract numeric value from websocket message: %s", string(body))
}

// extractValueByKeyPath parses JSON and navigates a dot-path (supports array indices)
func (w *WebSocketWorker) extractValueByKeyPath(body []byte, keyPath string) (float64, error) {
	var jsonData interface{}
	if err := json.Unmarshal(body, &jsonData); err != nil {
		return 0, fmt.Errorf("failed to parse JSON message: %w", err)
	}

	keys := strings.Split(keyPath, ".")
	var current = jsonData
	for _, k := range keys {
		switch v := current.(type) {
		case map[string]interface{}:
			if val, exists := v[k]; exists {
				current = val
			} else {
				return 0, fmt.Errorf("key '%s' not found in message", k)
			}
		case []interface{}:
			if idx, err := strconv.Atoi(k); err == nil && idx >= 0 && idx < len(v) {
				current = v[idx]
			} else {
				return 0, fmt.Errorf("invalid array index '%s'", k)
			}
		default:
			return 0, fmt.Errorf("cannot navigate to key '%s': intermediate is not object/array", k)
		}
	}
	return w.convertToFloat64(current)
}

// convertToFloat64 mirrors monitor_condition's conversion
func (w *WebSocketWorker) convertToFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		if floatVal, err := strconv.ParseFloat(v, 64); err == nil {
			return floatVal, nil
		}
		return 0, fmt.Errorf("cannot convert string '%s' to float64", v)
	default:
		return 0, fmt.Errorf("cannot convert value of type %s to float64", reflect.TypeOf(value))
	}
}

// evaluateCondition mirrors monitor_condition.evaluateCondition
func (w *WebSocketWorker) evaluateCondition(currentValue float64) (bool, error) {
	switch w.ConditionWorkerData.ConditionType {
	case ConditionGreaterThan:
		return currentValue > w.ConditionWorkerData.LowerLimit, nil
	case ConditionLessThan:
		return currentValue < w.ConditionWorkerData.UpperLimit, nil
	case ConditionBetween:
		return currentValue >= w.ConditionWorkerData.LowerLimit && currentValue <= w.ConditionWorkerData.UpperLimit, nil
	case ConditionEquals:
		return currentValue == w.ConditionWorkerData.LowerLimit, nil
	case ConditionNotEquals:
		return currentValue != w.ConditionWorkerData.LowerLimit, nil
	case ConditionGreaterEqual:
		return currentValue >= w.ConditionWorkerData.LowerLimit, nil
	case ConditionLessEqual:
		return currentValue <= w.ConditionWorkerData.UpperLimit, nil
	default:
		return false, fmt.Errorf("unsupported condition type: %s", w.ConditionWorkerData.ConditionType)
	}
}
