package db

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/config"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/events"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/handlers"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/middleware"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/redis"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/websocket"
	"github.com/trigg3rX/triggerx-backend/pkg/datastore"
	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"

	"github.com/gin-gonic/gin"
)

type Server struct {
	router      *gin.Engine
	datastore   datastore.DatastoreService
	logger      logging.Logger
	rateLimiter *middleware.RateLimiter
	apiKeyAuth  *middleware.ApiKeyAuth
	validator   *middleware.Validator
	redisClient *redis.Client

	// WebSocket components
	hub                 *websocket.Hub
	wsConnectionManager *websocket.WebSocketConnectionManager
}

func NewServer(datastore datastore.DatastoreService, logger logging.Logger) *Server {
	if !config.IsDevMode() {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize OpenTelemetry tracer
	_, err := middleware.InitTracer()
	if err != nil {
		logger.Errorf("Failed to initialize OpenTelemetry tracer: %v", err)
	}

	router := gin.New()
	router.Use(gin.Recovery())

	// Add tracing middleware before all others - creates traced logger for each request
	router.Use(middleware.TraceMiddleware(logger))

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
		datastore:   datastore,
		logger:      logger,
		rateLimiter: rateLimiter,
		redisClient: redisClient,
		validator:   middleware.NewValidator(logger),
	}

	apiKeyRepo := datastore.ApiKey()

	s.apiKeyAuth = middleware.NewApiKeyAuth(apiKeyRepo, rateLimiter, logger)

	// Initialize WebSocket components
	s.hub = websocket.NewHub(logger)

	// Create the task repository with publisher for WebSocket events
	taskRepo := datastore.Task()

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
	go s.hub.Run()
	logger.Info("WebSocket hub started successfully")

	return s
}

func (s *Server) RegisterRoutes(router *gin.Engine, dockerExecutor dockerexecutor.DockerExecutorAPI) {
	// Create event publisher
	publisher := events.NewPublisher(s.hub, s.logger)

	jobRepo := s.datastore.Job()
	timeJobRepo := s.datastore.TimeJob()
	eventJobRepo := s.datastore.EventJob()
	conditionJobRepo := s.datastore.ConditionJob()
	taskRepo := s.datastore.Task()
	userRepo := s.datastore.User()
	keeperRepo := s.datastore.Keeper()
	apiKeyRepo := s.datastore.ApiKey()

	// Create handler with WebSocket-enabled repository
	handler := handlers.NewHandler(
		s.logger,
		dockerExecutor,
		s.hub,
		publisher,
		s.datastore,
		jobRepo,
		timeJobRepo,
		eventJobRepo,
		conditionJobRepo,
		taskRepo,
		userRepo,
		keeperRepo,
		apiKeyRepo,
	)

	// Register metrics endpoint at root level without middleware
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.GET("/status", handler.HealthCheck)
	router.GET("/", handler.HealthCheck)

	api := router.Group("/api")

	protected := api.Group("")
	protected.Use(s.apiKeyAuth.GinMiddleware())

	// Returns UserDataDTO for the given address
	protected.GET("/users/:address", handler.GetUserDataByAddress)
	// Updates the email for the given address
	protected.PUT("/users/email", handler.UpdateUserEmail)

	// Create the Job (frontend and sdk)
	protected.POST("/jobs", s.validator.GinMiddleware(), handler.CreateJobData)
	// Returns jobs by user address, and optionally filters by chain_id if provided as a query param "?chain_id=1"
	protected.GET("/jobs/user/:user_address", handler.GetJobsByUserAddress)
	// Returns a single job data by job id
	protected.GET("/jobs/id/:job_id", handler.GetJobDataByJobID)
	// Updates an active job status to deleted by job id
	protected.PUT("/jobs/delete/:id", handler.DeleteJobData)
	// Updates a job data by job id
	protected.PUT("/jobs/update/:id", handler.UpdateJobDataFromUser)

	// Returns a single task data by task id
	protected.GET("/tasks/id/:id", handler.GetTaskDataByTaskID)
	// Returns tasks by job id
	protected.GET("/tasks/job/:job_id", handler.GetTasksByJobID)

	// Creates a new Keeper entry from Google Form
	api.POST("/keepers/form", s.validator.GinMiddleware(), handler.CreateKeeperDataFromGoogleForm)

	// Leaderboard routes
	protected.GET("/leaderboard/keepers", handler.GetKeeperLeaderboard)
	api.GET("/leaderboard/keepers/search", handler.GetKeeperByIdentifier)
	protected.GET("/leaderboard/users", handler.GetUserLeaderboard)
	protected.GET("/leaderboard/users/search", handler.GetUserLeaderboardByAddress)

	// Get the estimated fees for a task in a Job
	protected.GET("/jobs/fees", handler.CalculateFeesForJob)
	// Claim the fund from the faucet
	protected.POST("/claim-fund", handler.ClaimFund)	

	// Admin routes
	admin := protected.Group("/admin")
	admin.POST("/api-keys", s.validator.GinMiddleware(), handler.CreateApiKey)
	admin.PUT("/api-keys/:key", handler.DeleteApiKey)
	admin.GET("/api-keys/:owner", handler.GetApiKeysByOwner)

	// WebSocket routes
	wsHandler := handlers.NewWebSocketHandler(s.wsConnectionManager, s.logger)
	api.GET("/ws/tasks", wsHandler.HandleWebSocketConnection)
	api.GET("/ws/stats", wsHandler.GetWebSocketStats)
	api.GET("/ws/health", wsHandler.GetWebSocketHealth)
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

	// Graceful shutdown for WebSocket hub
	defer func() {
		if s.hub != nil {
			s.hub.Shutdown()
		}
	}()

	return s.router.Run(fmt.Sprintf(":%s", port))
}

func (s *Server) GetRouter() *gin.Engine {
	return s.router
}
