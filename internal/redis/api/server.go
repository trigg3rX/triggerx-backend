package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend-imua/internal/redis/api/handler"
	"github.com/trigg3rX/triggerx-backend-imua/internal/redis/metrics"
	"github.com/trigg3rX/triggerx-backend-imua/internal/redis/streams/jobs"
	"github.com/trigg3rX/triggerx-backend-imua/internal/redis/streams/tasks"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/logging"
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
	Logger           logging.Logger
	TaskStreamMgr    *tasks.TaskStreamManager
	JobStreamMgr     *jobs.JobStreamManager
	MetricsCollector *metrics.Collector
}

// NewServer creates a new API server
func NewServer(cfg Config, deps Dependencies) *Server {
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = 30 * time.Second
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = 30 * time.Second
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
	s.logger.Info("Starting Redis API server", "addr", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start Redis server: %w", err)
	}
	return nil
}

// Stop gracefully stops the server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping Redis API server")
	return s.httpServer.Shutdown(ctx)
}

// setupMiddleware sets up the middleware for the server
func (s *Server) setupMiddleware() {
	s.router.Use(gin.Recovery())
	s.router.Use(TraceMiddleware())
	s.router.Use(StreamMetricsMiddleware())
	s.router.Use(LoggerMiddleware(s.logger))
	s.router.Use(ErrorMiddleware(s.logger))
}

// setupRoutes sets up the routes for the server
func (s *Server) setupRoutes(deps Dependencies) {
	// Create handlers
	redisHandler := handler.NewHandler(deps.Logger, deps.TaskStreamMgr, deps.JobStreamMgr, deps.MetricsCollector)

	// Redis service routes
	s.router.GET("/", redisHandler.HandleRoot)
	s.router.GET("/health", redisHandler.HandleHealth)
	s.router.GET("/metrics", redisHandler.HandleMetrics)

	// Task stream routes
	s.router.GET("/streams/info", redisHandler.GetStreamsInfo)

	// Scheduler routes
	s.router.POST("/scheduler/submit-task", redisHandler.SubmitTaskFromScheduler)

	// P2P message handling (similar to keeper)
	s.router.POST("/task/validate", redisHandler.HandleValidateRequest)
	s.router.POST("/p2p/message", redisHandler.HandleP2PMessage)
}
