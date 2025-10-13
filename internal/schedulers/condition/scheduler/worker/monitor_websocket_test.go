package worker

import (
	"context"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func TestWebSocketWorker_TriggersOnMessage(t *testing.T) {
	upgrader := websocket.Upgrader{}
	// Send a numeric payload compatible with SelectedKeyRoute
	msgToSend := `{"price":{"usd":4025.75}}`

	// Setup in-memory websocket server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("failed to upgrade: %v", err)
		}
		defer conn.Close()
		conn.WriteMessage(websocket.TextMessage, []byte(msgToSend))
		// Wait so client can read
		time.Sleep(200 * time.Millisecond)
	}))
	defer ts.Close()

	wsURL := "ws" + ts.URL[len("http"):]
	condData := &types.ConditionWorkerData{
		JobID:            types.NewBigInt(big.NewInt(123)),
		ConditionType:    "greater_than",
		SelectedKeyRoute: "price.usd",
		LowerLimit:       3500,
		UpperLimit:       0,
		Recurring:        true,
	}
	wsConfig := &WebSocketConfig{URL: wsURL}
	var triggered bool
	var triggerMu sync.Mutex
	cb := func(notification *TriggerNotification) error {
		triggerMu.Lock()
		defer triggerMu.Unlock()
		triggered = true
		assert.Equal(t, condData.JobID.Int64(), notification.JobID.Int64())
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	worker := &WebSocketWorker{
		WebSocketConfig:     wsConfig,
		ConditionWorkerData: condData,
		Logger:              &noopLogger{},
		Ctx:                 ctx,
		Cancel:              cancel,
		TriggerCallback:     cb,
	}
	worker.Start()
	triggerMu.Lock()
	wasTriggered := triggered
	triggerMu.Unlock()
	assert.True(t, wasTriggered, "callback should be triggered on websocket message")
}

func TestWebSocketWorker_ConditionCases(t *testing.T) {
	upgrader := websocket.Upgrader{}

	type testCase struct {
		name          string
		msgToSend     string
		selectedRoute string
		conditionType string
		lower         float64
		upper         float64
		recurring     bool
		expectTrigger bool
	}

	cases := []testCase{
		{name: "greater_than true", msgToSend: `{"price":{"usd":4025.75}}`, selectedRoute: "price.usd", conditionType: "greater_than", lower: 3500, expectTrigger: true},
		{name: "greater_than false", msgToSend: `{"price":{"usd":3200}}`, selectedRoute: "price.usd", conditionType: "greater_than", lower: 3500, expectTrigger: false},
		{name: "less_than true", msgToSend: `{"price":{"usd":100}}`, selectedRoute: "price.usd", conditionType: "less_than", upper: 200, expectTrigger: true},
		{name: "less_than false", msgToSend: `{"price":{"usd":250}}`, selectedRoute: "price.usd", conditionType: "less_than", upper: 200, expectTrigger: false},
		{name: "between true", msgToSend: `{"metric":{"latency":15}}`, selectedRoute: "metric.latency", conditionType: "between", lower: 10, upper: 20, expectTrigger: true},
		{name: "between false", msgToSend: `{"metric":{"latency":25}}`, selectedRoute: "metric.latency", conditionType: "between", lower: 10, upper: 20, expectTrigger: false},
		{name: "equals true", msgToSend: `{"value":{"x":100}}`, selectedRoute: "value.x", conditionType: "equals", lower: 100, expectTrigger: true},
		{name: "equals false", msgToSend: `{"value":{"x":99}}`, selectedRoute: "value.x", conditionType: "equals", lower: 100, expectTrigger: false},
		{name: "not_equals true", msgToSend: `{"value":{"x":99}}`, selectedRoute: "value.x", conditionType: "not_equals", lower: 100, expectTrigger: true},
		{name: "not_equals false", msgToSend: `{"value":{"x":100}}`, selectedRoute: "value.x", conditionType: "not_equals", lower: 100, expectTrigger: false},
		{name: "greater_equal true", msgToSend: `{"v":100}`, selectedRoute: "v", conditionType: "greater_equal", lower: 100, expectTrigger: true},
		{name: "less_equal true", msgToSend: `{"v":100}`, selectedRoute: "v", conditionType: "less_equal", upper: 100, expectTrigger: true},
		{name: "string numeric true", msgToSend: `{"price":{"usd":"4025.75"}}`, selectedRoute: "price.usd", conditionType: "greater_than", lower: 3500, expectTrigger: true},
		{name: "direct numeric string true (no route)", msgToSend: `"4025.75"`, selectedRoute: "", conditionType: "greater_than", lower: 3500, expectTrigger: true},
		{name: "missing key path -> no trigger", msgToSend: `{"foo":1}`, selectedRoute: "price.usd", conditionType: "greater_than", lower: 0, expectTrigger: false},
		{name: "non-json -> no trigger", msgToSend: `hello`, selectedRoute: "price.usd", conditionType: "greater_than", lower: 0, expectTrigger: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// In-memory websocket server that sends one message per connection
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				conn, err := upgrader.Upgrade(w, r, nil)
				if err != nil {
					t.Fatalf("failed to upgrade: %v", err)
				}
				defer conn.Close()
				conn.WriteMessage(websocket.TextMessage, []byte(tc.msgToSend))
				time.Sleep(100 * time.Millisecond)
			}))
			defer ts.Close()

			wsURL := "ws" + ts.URL[len("http"):]
			condData := &types.ConditionWorkerData{
				JobID:            types.NewBigInt(big.NewInt(999)),
				Recurring:        true,
				ConditionType:    tc.conditionType,
				SelectedKeyRoute: tc.selectedRoute,
				LowerLimit:       tc.lower,
				UpperLimit:       tc.upper,
			}
			wsConfig := &WebSocketConfig{URL: wsURL}

			var triggered bool
			var mu sync.Mutex
			cb := func(notification *TriggerNotification) error {
				mu.Lock()
				triggered = true
				mu.Unlock()
				return nil
			}

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			worker := &WebSocketWorker{
				WebSocketConfig:     wsConfig,
				ConditionWorkerData: condData,
				Logger:              &noopLogger{},
				Ctx:                 ctx,
				Cancel:              cancel,
				TriggerCallback:     cb,
			}

			worker.Start()

			mu.Lock()
			got := triggered
			mu.Unlock()

			if tc.expectTrigger {
				assert.True(t, got, "expected trigger to be true")
			} else {
				assert.False(t, got, "expected trigger to be false")
			}
		})
	}
}

type noopLogger struct{}

func (n *noopLogger) Info(msg string, tags ...interface{})      {}
func (n *noopLogger) Error(msg string, tags ...interface{})     {}
func (n *noopLogger) Warn(msg string, tags ...interface{})      {}
func (n *noopLogger) Debug(msg string, tags ...interface{})     {}
func (n *noopLogger) Errorf(format string, args ...interface{}) {}
func (n *noopLogger) Debugf(format string, args ...interface{}) {}
func (n *noopLogger) Fatal(msg string, tags ...interface{})     {}
func (n *noopLogger) Fatalf(format string, args ...interface{}) {}
func (n *noopLogger) Infof(format string, args ...interface{})  {}
func (n *noopLogger) Warnf(format string, args ...interface{})  {}
func (n *noopLogger) With(tags ...interface{}) logging.Logger   { return n }
