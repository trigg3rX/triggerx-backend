package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"os"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/taskmanager/api/handler"
	"github.com/trigg3rX/triggerx-backend/internal/taskmanager/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/taskmanager/streams/jobs"
	"github.com/trigg3rX/triggerx-backend/internal/taskmanager/streams/tasks"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	gootel "go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
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

	// Initialize OpenTelemetry tracer
	_, err := InitTracer()
	if err != nil {
		deps.Logger.Error("Failed to initialize OpenTelemetry tracer", "error", err)
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

// InitTracer sets up OpenTelemetry tracing with OTLP exporter for Tempo
// Set TEMPO_OTLP_ENDPOINT env var to override the default (localhost:4318)
func InitTracer() (func(context.Context) error, error) {
	endpoint := os.Getenv("TEMPO_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "localhost:4318" // default to local Tempo
	}
	exporter, err := otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("triggerx-redis"),
		)),
	)
	gootel.SetTracerProvider(tp)
	return tp.Shutdown, nil
}
