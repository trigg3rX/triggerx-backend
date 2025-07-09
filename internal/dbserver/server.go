package dbserver

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/config"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/handlers"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/middleware"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/redis"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/docker"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type Server struct {
	router             *gin.Engine
	db                 *database.Connection
	logger             logging.Logger
	rateLimiter        *middleware.RateLimiter
	apiKeyAuth         *middleware.ApiKeyAuth
	validator          *middleware.Validator
	redisClient        *redis.Client
	notificationConfig handlers.NotificationConfig
	docker             docker.DockerConfig
}

func NewServer(db *database.Connection, logger logging.Logger) *Server {
	if !config.IsDevMode() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())

	// Start metrics collection
	metrics.StartMetricsCollection()
	metrics.StartSystemMetricsCollection()
	metrics.TrackDBConnections()

	// Apply middleware in the correct order
	router.Use(middleware.RecoveryMiddleware(logger))          // First, to catch panics
	router.Use(middleware.TimeoutMiddleware(30 * time.Second)) // Set appropriate timeout
	router.Use(middleware.MetricsMiddleware())                 // Track HTTP metrics

	// Configure CORS
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, Content-Length, Accept-Encoding, Origin, X-Requested-With, X-CSRF-Token, X-Auth-Token, X-Api-Key")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "false")

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
		docker: docker.DockerConfig{
			Image:          "golang:latest",
			TimeoutSeconds: 600,
			AutoCleanup:    true,
			MemoryLimit:    "1024m",
			CPULimit:       1.0,
		},
	}

	s.apiKeyAuth = middleware.NewApiKeyAuth(db, rateLimiter, logger)

	// Apply retry middleware only to API routes
	apiGroup := router.Group("/api")
	apiGroup.Use(middleware.RetryMiddleware(retryConfig, logger))

	return s
}

func (s *Server) RegisterRoutes(router *gin.Engine) {
	handler := handlers.NewHandler(s.db, s.logger, s.notificationConfig, s.docker)

	// Register metrics endpoint at root level without middleware
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	api := router.Group("/api")

	// Health check route - no authentication required
	api.GET("/health", handler.HealthCheck)

	// Public routes
	api.GET("/users/:address", handler.GetUserDataByAddress)
	api.GET("/wallet/points/:address", handler.GetWalletPoints)

	protected := api.Group("")
	protected.Use(s.apiKeyAuth.GinMiddleware())

	// Apply validation middleware to routes that need it
	api.POST("/jobs", s.validator.GinMiddleware(), handler.CreateJobData)
	api.GET("/jobs/time", handler.GetTimeBasedTasks)
	api.PUT("/jobs/:id", handler.UpdateJobDataFromUser)
	api.PUT("/jobs/:id/status/:status", handler.UpdateJobStatus)
	api.PUT("/jobs/:id/lastexecuted", handler.UpdateJobLastExecutedAt)
	api.GET("/jobs/user/:user_address", handler.GetJobsByUserAddress)
	api.PUT("/jobs/delete/:id", handler.DeleteJobData)

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

	api.GET("/leaderboard/keepers", handler.GetKeeperLeaderboard)
	api.GET("/leaderboard/users", handler.GetUserLeaderboard)
	api.GET("/leaderboard/users/search", handler.GetUserLeaderboardByAddress)
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
