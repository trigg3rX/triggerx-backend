package worker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
	"math/big"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

func TestWebSocketWorker_TriggersOnMessage(t *testing.T) {
	upgrader := websocket.Upgrader{}
	msgToSend := `{"event":"foo"}`

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
		JobID: types.NewBigInt(big.NewInt(123)),
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
		WebSocketConfig:    wsConfig,
		ConditionWorkerData: condData,
		Logger:             &noopLogger{},
		Ctx:                ctx,
		Cancel:             cancel,
		TriggerCallback:    cb,
	}
	worker.Start()
	triggerMu.Lock()
	wasTriggered := triggered
	triggerMu.Unlock()
	assert.True(t, wasTriggered, "callback should be triggered on websocket message")
}

type noopLogger struct{}
func (n *noopLogger) Info(msg string, tags ...interface{})   {}
func (n *noopLogger) Error(msg string, tags ...interface{})  {}
func (n *noopLogger) Warn(msg string, tags ...interface{})   {}
func (n *noopLogger) Debug(msg string, tags ...interface{})  {}
func (n *noopLogger) Errorf(format string, args ...interface{}) {}
func (n *noopLogger) Debugf(format string, args ...interface{}) {}
func (n *noopLogger) Fatal(msg string, tags ...interface{})         {}
func (n *noopLogger) Fatalf(format string, args ...interface{})     {}
func (n *noopLogger) Infof(format string, args ...interface{})      {}
func (n *noopLogger) Warnf(format string, args ...interface{})  {}
func (n *noopLogger) With(tags ...interface{}) logging.Logger { return n }
