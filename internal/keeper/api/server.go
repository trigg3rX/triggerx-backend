package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/api/handlers"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/core/execution"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/core/validation"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// Server represents the API server
type Server struct {
	router     *gin.Engine
	httpServer *http.Server
	logger     logging.Logger
}

// Config holds the server configuration
type Config struct {
	Port           string
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	MaxHeaderBytes int
}

// Dependencies holds the server dependencies
type Dependencies struct {
	Logger    logging.Logger
	Executor  execution.TaskExecutor
	Validator validation.TaskValidator
}

// NewServer creates a new API server
func NewServer(cfg Config, deps Dependencies) *Server {
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = 10 * time.Second
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = 10 * time.Second
	}
	if cfg.MaxHeaderBytes == 0 {
		cfg.MaxHeaderBytes = 1 << 20 // 1MB
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Create server instance
	srv := &Server{
		router: router,
		logger: deps.Logger,
		httpServer: &http.Server{
			Addr:           fmt.Sprintf(":%s", cfg.Port),
			Handler:        router,
			ReadTimeout:    cfg.ReadTimeout,
			WriteTimeout:   cfg.WriteTimeout,
			MaxHeaderBytes: cfg.MaxHeaderBytes,
		},
	}

	// Setup middleware
	srv.setupMiddleware()

	// Setup routes
	srv.setupRoutes(deps)

	return srv
}

// Start starts the server
func (s *Server) Start() error {
	s.logger.Info("Starting API server", "addr", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

// Stop gracefully stops the server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping API server")
	return s.httpServer.Shutdown(ctx)
}

// setupMiddleware sets up the middleware for the server
func (s *Server) setupMiddleware() {
	s.router.Use(gin.Recovery())
	s.router.Use(LoggerMiddleware(s.logger))
}

// setupRoutes sets up the routes for the server
func (s *Server) setupRoutes(deps Dependencies) {
	// Create handlers
	taskHandler := handlers.NewTaskHandler(deps.Logger, deps.Executor, deps.Validator)
	metricsHandler := handlers.NewMetricsHandler(deps.Logger)

	// Task routes
	s.router.POST("/p2p/message", taskHandler.ExecuteTask)
	// s.router.POST("/task/validate", taskHandler.ValidateTask)

	// Health and metrics routes
	s.router.GET("/metrics", metricsHandler.Metrics)
}
