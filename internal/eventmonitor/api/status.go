package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/config"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

// Server represents a simple HTTP server with only status endpoint
type Server struct {
	httpServer *http.Server
	logger     observability.Logger
}

// NewServer creates a new simple server with only /status endpoint
func NewServer(port string, logger observability.Logger) *Server {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	// Only status endpoint for Pulsate
	router.GET("/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "event-monitor",
			"version":   config.GetVersion(),
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	return &Server{
		httpServer: &http.Server{
			Addr:    fmt.Sprintf(":%s", port),
			Handler: router,
		},
		logger: logger,
	}
}

// Start starts the server
func (s *Server) Start(ctx context.Context) error {
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

// Stop gracefully stops the server
func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
