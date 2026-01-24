package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/api/handlers"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/api/middleware"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/config"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/events"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/redis"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/rpc/clients/conditionscheduler"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/websocket"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor"
	httpclientpkg "github.com/trigg3rX/triggerx-backend/pkg/http"
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
	Port           string
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	MaxHeaderBytes int
}

// Dependencies holds the server dependencies
type Dependencies struct {
	Logger                   observability.Logger
	Tracer                   observability.Tracer
	Metrics                  observability.Metrics
	DB                       *database.Connection
	RedisClient              *redis.Client
	RateLimiter              *middleware.RateLimiter
	ApiKeyAuth               *middleware.ApiKeyAuth
	Validator                *middleware.Validator
	Hub                      *websocket.Hub
	WSConnectionManager      *websocket.WebSocketConnectionManager
	ConditionSchedulerClient *conditionscheduler.Client
	DockerExecutor           dockerexecutor.DockerExecutorAPI
}

// NewServer creates a new API server
func NewServer(cfg Config, deps *Dependencies) *Server {
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
	router.Use(gin.Recovery())

	// Create server instance
	srv := &Server{
		router: router,
		logger: deps.Logger,
		httpServer: &http.Server{
			Addr:           fmt.Sprintf("0.0.0.0:%s", cfg.Port),
			Handler:        router,
			ReadTimeout:    cfg.ReadTimeout,
			WriteTimeout:   cfg.WriteTimeout,
			MaxHeaderBytes: cfg.MaxHeaderBytes,
		},
	}

	// Setup middleware
	srv.setupMiddleware(deps)

	// Setup routes
	srv.setupRoutes(deps)

	return srv
}

// Start starts the server
func (s *Server) Start() error {
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

// Stop gracefully stops the server
func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// GetRouter returns the gin router (for backward compatibility if needed)
func (s *Server) GetRouter() *gin.Engine {
	return s.router
}

// setupMiddleware sets up the middleware for the server
func (s *Server) setupMiddleware(deps *Dependencies) {
	ctx := context.Background()

	// Add tracing middleware before all others
	// This middleware requires X-Trace-ID header and rejects requests without it
	// /status endpoint bypasses this middleware
	traceMiddleware := middleware.TraceMiddleware(deps.Tracer)
	s.router.Use(func(c *gin.Context) {
		// Bypass trace middleware for /status endpoint
		if c.Request.URL.Path == "/status" {
			c.Next()
			return
		}
		traceMiddleware(c)
	})

	// Apply middleware in the correct order
	s.router.Use(middleware.RecoveryMiddleware(deps.Logger))      // First, to catch panics
	s.router.Use(middleware.TimeoutMiddleware(100 * time.Second)) // Set appropriate timeout
	s.router.Use(middleware.MetricsMiddleware())                  // Track HTTP metrics

	// Configure CORS
	s.router.Use(func(c *gin.Context) {
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

	// Apply retry middleware only to API routes
	apiGroup := s.router.Group("/api")
	apiGroup.Use(middleware.RetryMiddleware(ctx, retryConfig, deps.Logger))
}

// setupRoutes sets up the routes for the server
func (s *Server) setupRoutes(deps *Dependencies) {
	ctx := context.Background()

	// Create event publisher
	publisher := events.NewPublisher(deps.Hub, deps.Logger)

	// Initialize robust HTTP client
	httpClient, err := httpclientpkg.NewHTTPClient(httpclientpkg.DefaultHTTPRetryConfig())
	if err != nil {
		s.logger.Error(ctx, "Failed to create HTTP client", observability.Error(err))
		panic(fmt.Errorf("failed to initialize HTTP client: %w", err))
	}

	// Create handler w/ HTTP client, Redis client, and condition scheduler gRPC client
	notificationConfig := handlers.NotificationConfig{
		EmailFrom:     config.GetEmailUser(),
		EmailPassword: config.GetEmailPassword(),
		BotToken:      config.GetBotToken(),
	}

	handler := handlers.NewHandler(
		deps.DB,
		deps.Logger,
		deps.Tracer,
		notificationConfig,
		deps.DockerExecutor,
		deps.Hub,
		publisher,
		httpClient,
		deps.RedisClient,
		deps.ConditionSchedulerClient,
	)

	// Register status endpoint for Pulsate and nginx
	s.router.GET("/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "dbserver",
			"version":   config.GetVersion(),
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	api := s.router.Group("/api")
	// Code validation endpoint (raw source)
	api.POST("/code/validate", handler.ValidateCodeExecutable)

	protected := api.Group("")
	protected.Use(deps.ApiKeyAuth.GinMiddleware())

	// Public routes
	protected.GET("/users/:address", handler.GetUserDataByAddress)
	protected.POST("/users/email", handler.StoreUserEmail)

	// Apply validation middleware to routes that need it
	api.POST("/jobs", deps.Validator.GinMiddleware(), handler.CreateJobData)
	protected.GET("/jobs/by-apikey", handler.GetJobsByApiKey)
	api.PUT("/jobs/update/:id", handler.UpdateJobDataFromUser)
	protected.GET("/jobs/user/:user_address", handler.GetJobsByUserAddress) // unused for now, can be added in SDK
	protected.GET("/jobs/user/:user_address/chain/:created_chain_id", handler.GetJobsByUserAddressAndChainID)
	protected.PUT("/jobs/delete/:id", handler.DeleteJobData)
	protected.GET("/jobs/user/:user_address/:job_id", handler.GetJobDataByJobIDForUser)

	api.GET("/tasks/:id", handler.GetTaskDataByID)
	api.GET("/tasks/job/:job_id", handler.GetTasksByJobID)
	protected.GET("/tasks/recent", handler.GetRecentTasks)
	protected.GET("/tasks/user/:user_address", handler.GetTasksByUserAddress)
	protected.GET("/tasks/by-apikey/:api_key", handler.GetTasksByApiKey)
	protected.GET("/tasks/safe-address/:safe_address", handler.GetTasksBySafeAddress)

	protected.GET("/leaderboard/keepers", handler.GetKeeperLeaderboard)
	protected.GET("/leaderboard/users", handler.GetUserLeaderboard)

	api.GET("/fees", handler.GetTaskFees)

	api.POST("/claim-fund", handler.ClaimFund)

	// Admin routes
	admin := protected.Group("/admin")
	admin.POST("/api-keys", deps.Validator.GinMiddleware(), handler.CreateApiKey)
	admin.PUT("/api-keys/:key", handler.UpdateApiKey)
	admin.DELETE("/api-keys/:key", handler.DeleteApiKey)
	admin.GET("/api-keys/:owner", handler.GetApiKeysByOwner)

	// Keeper routes
	keeper := protected.Group("/keeper")
	keeper.Use(deps.ApiKeyAuth.KeeperMiddleware())
	// Keeper-specific routes will be added here later

	// WebSocket routes
	wsHandler := handlers.NewWebSocketHandler(deps.WSConnectionManager, deps.Logger)
	api.GET("/ws/tasks", wsHandler.HandleWebSocketConnection)
	api.GET("/ws/stats", wsHandler.GetWebSocketStats)   // unused for now, can be added in SDK after adding client_id variable
	api.GET("/ws/health", wsHandler.GetWebSocketHealth) // unused for now, can be added in SDK after adding client_id variable

	protected.GET("/users/safe-addresses/:user_address", handler.GetSafeAddressesByUser) // unused for now, can be added in SDK
	protected.GET("/jobs/safe-address/:safe_address", handler.GetJobsBySafeAddress)
}
