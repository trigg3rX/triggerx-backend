package db

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/client/conditionscheduler"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/config"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/events"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/handlers"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/middleware"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/redis"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/repository"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/websocket"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor"
	httpclientpkg "github.com/trigg3rX/triggerx-backend/pkg/http"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"

	"github.com/gin-gonic/gin"
)

type Server struct {
	router             *gin.Engine
	db                 *database.Connection
	logger             observability.Logger
	tracer             observability.Tracer
	rateLimiter        *middleware.RateLimiter
	apiKeyAuth         *middleware.ApiKeyAuth
	validator          *middleware.Validator
	redisClient        *redis.Client
	notificationConfig handlers.NotificationConfig
	obsMetrics         observability.Metrics

	// WebSocket components
	hub                 *websocket.Hub
	wsConnectionManager *websocket.WebSocketConnectionManager
}

func NewServer(ctx context.Context, db *database.Connection, logger observability.Logger, tracer observability.Tracer, obsMetrics observability.Metrics) *Server {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Recovery())

	// Add tracing middleware before all others
	// This middleware requires X-Trace-ID header and rejects requests without it
	router.Use(middleware.TraceMiddleware(tracer))

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
		logger.Error(ctx, "Failed to initialize Redis client", observability.Error(err))
	} else {
		redisClient = client
		logger.Info(ctx, "Redis client initialized successfully")
	}

	// Initialize rate limiter
	var rateLimiter *middleware.RateLimiter
	if redisClient != nil {
		var err error
		rateLimiter, err = middleware.NewRateLimiterWithClient(redisClient, logger)
		if err != nil {
			logger.Error(ctx, "Failed to initialize rate limiter", observability.Error(err))
		} else {
			logger.Info(ctx, "Rate limiter initialized successfully")
		}
	} else {
		logger.Warn(ctx, "Rate limiter disabled - Redis client not available")
	}

	s := &Server{
		router:      router,
		db:          db,
		logger:      logger,
		tracer:      tracer,
		rateLimiter: rateLimiter,
		redisClient: redisClient,
		validator:   middleware.NewValidator(ctx, logger),
		obsMetrics:  obsMetrics,
		notificationConfig: handlers.NotificationConfig{
			EmailFrom:     config.GetEmailUser(),
			EmailPassword: config.GetEmailPassword(),
			BotToken:      config.GetBotToken(),
		},
	}

	s.apiKeyAuth = middleware.NewApiKeyAuth(db, rateLimiter, logger)

	// Initialize WebSocket components
	s.hub = websocket.NewHub(logger)

	// Create the task repository with publisher for WebSocket events
	taskRepo := repository.NewTaskRepositoryWithPublisher(db, nil) // publisher will be set later if needed

	// Create and set the initial data handler for the hub
	initialDataHandler := handlers.NewInitialDataHandler(taskRepo, logger)
	s.hub.SetInitialDataCallback(initialDataHandler.HandleInitialData)
	s.wsConnectionManager = websocket.NewWebSocketConnectionManager(
		websocket.NewWebSocketUpgrader(logger),
		websocket.NewWebSocketAuthMiddleware(s.apiKeyAuth, logger),
		websocket.NewWebSocketRateLimiter(s.rateLimiter, 100, logger), // Max 100 connections per IP
		s.hub,
		logger,
	)

	// Start WebSocket hub
	go s.hub.Run(ctx)
	logger.Info(ctx, "WebSocket hub started successfully")

	// Apply retry middleware only to API routes
	apiGroup := router.Group("/api")
	apiGroup.Use(middleware.RetryMiddleware(ctx, retryConfig, logger))
	return s
}

func (s *Server) RegisterRoutes(ctx context.Context, router *gin.Engine, dockerExecutor dockerexecutor.DockerExecutorAPI) error {
	// Create event publisher
	publisher := events.NewPublisher(s.hub, s.logger)

	// Initialize robust HTTP client
	httpClient, err := httpclientpkg.NewHTTPClient(httpclientpkg.DefaultHTTPRetryConfig())
	if err != nil {
		s.logger.Error(ctx, "Failed to create HTTP client", observability.Error(err))
		return fmt.Errorf("failed to initialize HTTP client: %w", err)
	}

	// Initialize condition scheduler gRPC client
	conditionSchedulerClient, err := conditionscheduler.NewClient(
		config.GetConditionSchedulerRPCUrl(),
		s.logger,
		s.tracer,
	)
	if err != nil {
		s.logger.Error(ctx, "Failed to create condition scheduler gRPC client", observability.Error(err))
		panic(err)
	}

	// Create handler w/ HTTP client, Redis client, and condition scheduler gRPC client
	handler := handlers.NewHandler(s.db, s.logger, s.notificationConfig, dockerExecutor, s.hub, publisher, httpClient, s.redisClient, conditionSchedulerClient)

	// Register metrics endpoint at root level without middleware
	router.GET("/metrics", gin.WrapH(metrics.NewCollector(s.obsMetrics, s.logger).Handler()))

	// Register status endpoint for Pulsate and nginx
	router.GET("/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "dbserver",
			"version":   config.GetVersion(),
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	api := router.Group("/api")
	// Code validation endpoint (raw source)
	api.POST("/code/validate", handler.ValidateCodeExecutable)

	// Health check route - no authentication required
	api.GET("/health", handler.HealthCheck)

	protected := api.Group("")
	protected.Use(s.apiKeyAuth.GinMiddleware())

	// Public routes
	protected.GET("/users/:address", handler.GetUserDataByAddress)
	protected.POST("/users/email", handler.StoreUserEmail)

	// Apply validation middleware to routes that need it
	// api.POST("/jobs", s.validator.GinMiddleware(), handler.CreateJobData)
	api.POST("/jobs", s.validator.GinMiddleware(), handler.CreateJobData)
	protected.GET("/jobs/by-apikey", handler.GetJobsByApiKey)
	api.PUT("/jobs/update/:id", handler.UpdateJobDataFromUser)
	api.PUT("/jobs/:id/status/:status", handler.UpdateJobStatus)
	api.PUT("/jobs/:id/lastexecuted", handler.UpdateJobLastExecutedAt)
	protected.GET("/jobs/user/:user_address", handler.GetJobsByUserAddress)
	protected.GET("/jobs/user/:user_address/chain/:created_chain_id", handler.GetJobsByUserAddressAndChainID)
	protected.PUT("/jobs/delete/:id", handler.DeleteJobData)
	protected.GET("/jobs/user/:user_address/:job_id", handler.GetJobDataByJobIDForUser)
	api.GET("/jobs/:job_id/task-fees", handler.GetTaskFeesByJobID)

	api.GET("/tasks/:id", handler.GetTaskDataByID)
	api.GET("/tasks/job/:job_id", handler.GetTasksByJobID)
	protected.GET("/tasks/recent", handler.GetRecentTasks)
	protected.GET("/tasks/user/:user_address", handler.GetTasksByUserAddress)
	protected.GET("/tasks/by-apikey/:api_key", handler.GetTasksByApiKey)
	protected.GET("/tasks/safe-address/:safe_address", handler.GetTasksBySafeAddress)

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

	// WebSocket routes
	wsHandler := handlers.NewWebSocketHandler(s.wsConnectionManager, s.logger)
	api.GET("/ws/tasks", wsHandler.HandleWebSocketConnection)
	api.GET("/ws/stats", wsHandler.GetWebSocketStats)
	api.GET("/ws/health", wsHandler.GetWebSocketHealth)

	protected.GET("/users/safe-addresses/:user_address", handler.GetSafeAddressesByUser)
	protected.GET("/jobs/safe-address/:safe_address", handler.GetJobsBySafeAddress)

	return nil
}

func (s *Server) Start(ctx context.Context, port string) error {
	s.logger.Info(ctx, "Starting server on port", observability.String("port", port))

	if s.redisClient != nil {
		defer func() {
			if err := s.redisClient.Close(); err != nil {
				s.logger.Error(ctx, "Failed to close Redis client", observability.Error(err))
			}
		}()
	}

	// Graceful shutdown for WebSocket hub
	defer func() {
		if s.hub != nil {
			s.hub.Shutdown(ctx)
		}
	}()

	return s.router.Run(fmt.Sprintf("0.0.0.0:%s", port))
}

func (s *Server) GetRouter() *gin.Engine {
	return s.router
}
