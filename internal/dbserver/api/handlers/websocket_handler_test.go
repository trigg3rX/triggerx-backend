package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	ws "github.com/trigg3rX/triggerx-backend/internal/dbserver/websocket"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

func TestWebSocketStats(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Build a minimal real manager with a hub so types match
	hub := ws.NewHub(&logging.MockLogger{})
	mgr := ws.NewWebSocketConnectionManager(nil, nil, nil, hub, &logging.MockLogger{})
	wsh := NewWebSocketHandler(mgr, &logging.MockLogger{})

	r := gin.New()
	r.GET("/ws/stats", wsh.GetWebSocketStats)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ws/stats", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestWebSocketHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	hub := ws.NewHub(&logging.MockLogger{})
	mgr := ws.NewWebSocketConnectionManager(nil, nil, nil, hub, &logging.MockLogger{})
	wsh := NewWebSocketHandler(mgr, &logging.MockLogger{})

	r := gin.New()
	r.GET("/ws/health", wsh.GetWebSocketHealth)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ws/health", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
