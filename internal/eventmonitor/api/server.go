package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/api/handlers"
	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/config"
	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/registry"
	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/service"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

// Server represents the API server
type Server struct {
	router     *gin.Engine
	httpServer *http.Server
	logger     observability.Logger
}

// Config holds the server configuration
type Config struct {
	Port string
}

// Dependencies holds the server dependencies
type Dependencies struct {
	Logger          observability.Logger
	RegistryManager *registry.RegistryManager
	Service         *service.Service
}

// NewServer creates a new API server
func NewServer(cfg Config, deps Dependencies) *Server {
	if config.IsDevMode() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Create server instance
	srv := &Server{
		router: router,
		logger: deps.Logger,
		httpServer: &http.Server{
			Addr:    fmt.Sprintf("%s:%s", config.GetHost(), cfg.Port),
			Handler: router,
		},
	}

	// Setup middleware
	srv.setupMiddleware()

	// Setup routes
	srv.setupRoutes(deps)

	return srv
}

// Start starts the server
func (s *Server) Start(ctx context.Context) error {
	s.logger.Info(ctx, "Starting API server", observability.String("addr", s.httpServer.Addr))
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

// Stop gracefully stops the server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info(ctx, "Stopping API server")
	return s.httpServer.Shutdown(ctx)
}

// setupMiddleware sets up the middleware for the server
func (s *Server) setupMiddleware() {
	// Recovery middleware
	s.router.Use(gin.Recovery())

	// Logging and Metrics middleware
	s.router.Use(LoggerMiddleware(s.logger))
}

// setupRoutes sets up the routes for the server
func (s *Server) setupRoutes(deps Dependencies) {
	// Health check
	s.router.GET("/health", handlers.HandleHealth(deps.Logger, deps.RegistryManager))

	// API v1 routes
	v1 := s.router.Group("/api/v1/monitor")
	{
		v1.POST("/register", handlers.HandleRegister(deps.Logger, deps.Service))
		v1.POST("/unregister", handlers.HandleUnregister(deps.Logger, deps.Service))
		v1.GET("/status/:request_id", handlers.HandleStatus(deps.Logger, deps.RegistryManager))
	}
}
