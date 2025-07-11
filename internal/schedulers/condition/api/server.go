package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend-imua/internal/schedulers/condition/api/handlers"
	"github.com/trigg3rX/triggerx-backend-imua/internal/schedulers/condition/scheduler"
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
	Port string
}

// Dependencies holds the server dependencies
type Dependencies struct {
	Logger    logging.Logger
	Scheduler *scheduler.ConditionBasedScheduler
}

// NewServer creates a new API server
func NewServer(cfg Config, deps Dependencies) *Server {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Create server instance
	srv := &Server{
		router: router,
		logger: deps.Logger,
		httpServer: &http.Server{
			Addr:    fmt.Sprintf(":%s", cfg.Port),
			Handler: router,
		},
	}

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

// setupRoutes sets up the routes for the server
func (s *Server) setupRoutes(deps Dependencies) {
	// Apply metrics middleware to all routes
	s.router.Use(MetricsMiddleware())

	// Create handlers
	statusHandler := handlers.NewStatusHandler(deps.Logger)
	metricsHandler := handlers.NewMetricsHandler(deps.Logger)
	schedulerHandler := handlers.NewSchedulerHandler(deps.Logger, deps.Scheduler)

	// Health and monitoring endpoints
	s.router.GET("/status", statusHandler.Status)
	s.router.GET("/metrics", metricsHandler.Metrics)

	// Scheduler management endpoints
	api := s.router.Group("/api/v1")
	{
		api.GET("/scheduler/stats", schedulerHandler.GetStats)

		// Job management endpoints
		api.POST("/job/schedule", schedulerHandler.ScheduleJob)
		api.POST("/job/pause", schedulerHandler.UnscheduleJob)
		api.GET("/job/stats/:job_id", schedulerHandler.GetJobStats)

		api.PUT("/job/task/:job_id/:task_id", schedulerHandler.UpdateJobsTask)
	}
}
