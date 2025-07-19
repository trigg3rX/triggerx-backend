package dbserver

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/config"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/handlers"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/middleware"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/redis"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/docker"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	gootel "go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

const TraceIDHeader = "X-Trace-ID"
const TraceIDKey = "trace_id"

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
			semconv.ServiceNameKey.String("triggerx-backend"),
		)),
	)
	gootel.SetTracerProvider(tp)
	return tp.Shutdown, nil
}

// TraceMiddleware injects a trace ID into the Gin context and response headers
func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the global tracer
		tracer := gootel.Tracer("triggerx-backend")

		// Start a new span for this request
		ctx, span := tracer.Start(c.Request.Context(), c.Request.URL.Path)
		defer span.End()

		// Set span attributes
		span.SetAttributes(
			semconv.HTTPMethodKey.String(c.Request.Method),
			semconv.HTTPURLKey.String(c.Request.URL.String()),
			semconv.HTTPUserAgentKey.String(c.Request.UserAgent()),
		)

		// Get or generate trace ID
		traceID := c.GetHeader(TraceIDHeader)
		if traceID == "" {
			// Extract trace ID from span context
			spanContext := span.SpanContext()
			if spanContext.HasTraceID() {
				traceID = spanContext.TraceID().String()
			} else {
				traceID = uuid.New().String()
			}
		}

		// Store in context
		c.Set(TraceIDKey, traceID)
		c.Header(TraceIDHeader, traceID)

		// Update request context with span context
		c.Request = c.Request.WithContext(ctx)

		// Process request
		c.Next()

		// Set response status on span
		span.SetAttributes(semconv.HTTPStatusCodeKey.Int(c.Writer.Status()))
	}
}

type Server struct {
	router             *gin.Engine
	db                 *database.Connection
	logger             logging.Logger
	rateLimiter        *middleware.RateLimiter
	apiKeyAuth         *middleware.ApiKeyAuth
	validator          *middleware.Validator
	redisClient        *redis.Client
	notificationConfig handlers.NotificationConfig
	executor           docker.ExecutorConfig
}

func NewServer(db *database.Connection, logger logging.Logger) *Server {
	if !config.IsDevMode() {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize OpenTelemetry tracer
	_, err := InitTracer()
	if err != nil {
		logger.Errorf("Failed to initialize OpenTelemetry tracer: %v", err)
	}

	router := gin.New()
	router.Use(gin.Recovery())

	// Add tracing middleware before all others
	router.Use(TraceMiddleware())

	// Start metrics collection
	metrics.StartMetricsCollection()
	metrics.StartSystemMetricsCollection()
	metrics.TrackDBConnections()

	// Apply middleware in the correct order
	router.Use(middleware.RecoveryMiddleware(logger))           // First, to catch panics
	router.Use(middleware.TimeoutMiddleware(100 * time.Second)) // Set appropriate timeout
	router.Use(middleware.MetricsMiddleware())                  // Track HTTP metrics

	// Configure CORS
	router.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, Content-Length, Accept-Encoding, Origin, X-Requested-With, X-CSRF-Token, X-Auth-Token, X-Api-Key, ngrok-skip-browser-warning")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "false")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// Add retry middleware with custom configuration
	retryConfig := &middleware.RetryConfig{
		MaxRetries:      3,
		InitialDelay:    time.Second,
		MaxDelay:        10 * time.Second,
		BackoffFactor:   2.0,
		JitterFactor:    0.1,
		LogRetryAttempt: true,
		RetryStatusCodes: []int{
			http.StatusInternalServerError,
			http.StatusBadGateway,
			http.StatusServiceUnavailable,
			http.StatusGatewayTimeout,
			http.StatusTooManyRequests,
			http.StatusRequestTimeout,
			http.StatusConflict,
		},
	}

	// Initialize Redis client with enhanced features
	var redisClient *redis.Client
	client, err := redis.NewClient(logger)
	if err != nil {
		logger.Errorf("Failed to initialize Redis client: %v", err)
	} else {
		redisClient = client
		logger.Infof("Redis client initialized successfully")
	}

	// Initialize rate limiter
	var rateLimiter *middleware.RateLimiter
	if redisClient != nil {
		var err error
		rateLimiter, err = middleware.NewRateLimiterWithClient(redisClient, logger)
		if err != nil {
			logger.Errorf("Failed to initialize rate limiter: %v", err)
		} else {
			logger.Info("Rate limiter initialized successfully")
		}
	} else {
		logger.Warn("Rate limiter disabled - Redis client not available")
	}

	s := &Server{
		router:      router,
		db:          db,
		logger:      logger,
		rateLimiter: rateLimiter,
		redisClient: redisClient,
		validator:   middleware.NewValidator(logger),
		notificationConfig: handlers.NotificationConfig{
			EmailFrom:     config.GetEmailUser(),
			EmailPassword: config.GetEmailPassword(),
			BotToken:      config.GetBotToken(),
		},
		executor: docker.DefaultConfig(),
	}

	s.apiKeyAuth = middleware.NewApiKeyAuth(db, rateLimiter, logger)

	// Apply retry middleware only to API routes
	apiGroup := router.Group("/api")
	apiGroup.Use(middleware.RetryMiddleware(retryConfig, logger))

	return s
}

func (s *Server) RegisterRoutes(router *gin.Engine) {
	handler := handlers.NewHandler(s.db, s.logger, s.notificationConfig, s.executor)

	// Register metrics endpoint at root level without middleware
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	api := router.Group("/api")

	// Health check route - no authentication required
	api.GET("/health", handler.HealthCheck)

	protected := api.Group("")
	protected.Use(s.apiKeyAuth.GinMiddleware())

	// Public routes
	protected.GET("/users/:address", handler.GetUserDataByAddress)
	protected.POST("/users/email", handler.StoreUserEmail)

	// Apply validation middleware to routes that need it
	// api.POST("/jobs", s.validator.GinMiddleware(), handler.CreateJobData)
	protected.POST("/jobs", s.validator.GinMiddleware(), handler.CreateJobData)
	protected.GET("/jobs/by-apikey", handler.GetJobsByApiKey)
	api.GET("/jobs/time", handler.GetTimeBasedTasks)
	api.PUT("/jobs/update/:id", handler.UpdateJobDataFromUser)
	api.PUT("/jobs/:id/status/:status", handler.UpdateJobStatus)
	api.PUT("/jobs/:id/lastexecuted", handler.UpdateJobLastExecutedAt)
	protected.GET("/jobs/user/:user_address", handler.GetJobsByUserAddress)
	protected.PUT("/jobs/delete/:id", handler.DeleteJobData)
	api.GET("/jobs/:job_id/task-fees", handler.GetTaskFeesByJobID)

	api.POST("/tasks", s.validator.GinMiddleware(), handler.CreateTaskData)
	api.GET("/tasks/:id", handler.GetTaskDataByID)
	// api.PUT("/tasks/:id/fee", handler.UpdateTaskFee)
	// api.PUT("/tasks/:id/attestation", handler.UpdateTaskAttestationData)
	api.PUT("/tasks/execution/:id", handler.UpdateTaskExecutionData)
	api.GET("/tasks/job/:job_id", handler.GetTasksByJobID)

	api.POST("/keepers", s.validator.GinMiddleware(), handler.CreateKeeperData)
	api.POST("/keepers/form", s.validator.GinMiddleware(), handler.CreateKeeperDataGoogleForm)
	api.GET("/keepers/performers", handler.GetPerformers)
	api.GET("/keepers/:id", handler.GetKeeperData)
	api.POST("/keepers/:id/increment-tasks", handler.IncrementKeeperTaskCount)
	api.GET("/keepers/:id/task-count", handler.GetKeeperTaskCount)
	api.POST("/keepers/:id/add-points", handler.AddTaskFeeToKeeperPoints)
	api.GET("/keepers/:id/points", handler.GetKeeperPoints)

	protected.GET("/leaderboard/keepers", handler.GetKeeperLeaderboard)
	protected.GET("/leaderboard/users", handler.GetUserLeaderboard)
	protected.GET("/leaderboard/users/search", handler.GetUserLeaderboardByAddress)
	api.GET("/leaderboard/keepers/search", handler.GetKeeperByIdentifier)

	api.GET("/fees", handler.GetTaskFees)

	api.POST("/keepers/update-chat-id", handler.UpdateKeeperChatID)
	api.GET("/keepers/com-info/:id", handler.GetKeeperCommunicationInfo)
	api.POST("/claim-fund", handler.ClaimFund)

	// Admin routes
	admin := protected.Group("/admin")
	admin.POST("/api-keys", s.validator.GinMiddleware(), handler.CreateApiKey)
	admin.PUT("/api-keys/:key", handler.UpdateApiKey)
	admin.DELETE("/api-keys/:key", handler.DeleteApiKey)
	admin.GET("/api-keys/:owner", handler.GetApiKeysByOwner)

	// Keeper routes
	keeper := protected.Group("/keeper")
	keeper.Use(s.apiKeyAuth.KeeperMiddleware())
	// Keeper-specific routes will be added here later

}

func (s *Server) Start(port string) error {
	s.logger.Infof("Starting server on port %s", port)

	if s.redisClient != nil {
		defer func() {
			if err := s.redisClient.Close(); err != nil {
				s.logger.Errorf("Failed to close Redis client: %v", err)
			}
		}()
	}

	return s.router.Run(fmt.Sprintf(":%s", port))
}

func (s *Server) GetRouter() *gin.Engine {
	return s.router
}
