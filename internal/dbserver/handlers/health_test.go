package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestHealthCheck_Healthy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &Handler{logger: &MockLogger{}}
	h.scanNowQuery = func(ts *time.Time) error { *ts = time.Now(); return nil }
	r := gin.New()
	r.GET("/health", h.HealthCheck)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHealthCheck_Unhealthy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &Handler{logger: &MockLogger{}}
	h.scanNowQuery = func(ts *time.Time) error { return errors.New("db down") }
	r := gin.New()
	r.GET("/health", h.HealthCheck)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when DB unhealthy, got %d", w.Code)
	}
}
