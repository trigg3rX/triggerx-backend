package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/trigg3rX/triggerx-backend/internal/health/api/handlers"
	"github.com/trigg3rX/triggerx-backend/internal/health/core/keeper"
	"github.com/trigg3rX/triggerx-backend/internal/health/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

// Server represents the HTTP API server
type Server struct {
	router     *gin.Engine
	httpServer *http.Server
	logger     observability.Logger
}

// Config holds the server configuration
type Config struct {
	Port           string
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	MaxHeaderBytes int
}

// Dependencies holds all dependencies needed by the server
type Dependencies struct {
	Logger       observability.Logger
	Tracer       observability.Tracer
	Metrics      observability.Metrics
	StateManager *keeper.StateManager
}

// NewServer creates a new HTTP API server
func NewServer(cfg Config, deps *Dependencies) *Server {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Recovery())

	// Add tracing middleware before logging middleware to ensure trace context is available
	router.Use(TraceMiddleware(deps.Tracer))
	router.Use(LoggerMiddleware(deps.Logger))

	// Register routes
	registerRoutes(router, deps.Logger, deps.StateManager)

	httpServer := &http.Server{
		Addr:           fmt.Sprintf("0.0.0.0:%s", cfg.Port),
		Handler:        router,
		ReadTimeout:    cfg.ReadTimeout,
		WriteTimeout:   cfg.WriteTimeout,
		MaxHeaderBytes: cfg.MaxHeaderBytes,
	}

	return &Server{
		router:     router,
		httpServer: httpServer,
		logger:     deps.Logger,
	}
}

// registerRoutes registers all HTTP routes for the health service
func registerRoutes(router *gin.Engine, logger observability.Logger, stateManager *keeper.StateManager) {
	handler := handlers.NewHandler(logger, stateManager)

	// Start metrics collection (metrics should already be initialized via InitializeMetrics)
	metrics.StartMetricsCollection()

	// Service status endpoint for Pulsate and nginx
	router.GET("/status", handlers.HandleStatus)

	router.POST("/health", handler.HandleCheckInEvent)
	router.GET("/operators", handler.GetDetailedKeeperStatus)
}

// Start starts the HTTP server
func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

// Stop gracefully stops the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	if err := s.httpServer.Shutdown(ctx); err != nil {
		if closeErr := s.httpServer.Close(); closeErr != nil {
			s.logger.Error(ctx, "Forced HTTP server close error", observability.Error(closeErr))
		}
		return err
	}
	return nil
}
