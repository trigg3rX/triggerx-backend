package dbserver

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/config"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/handlers"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/middleware"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/redis"
)

type Server struct {
	router             *gin.Engine
	db                 *database.Connection
	logger             logging.Logger
	metricsServer      *metrics.MetricsServer
	rateLimiter        *middleware.RateLimiter
	apiKeyAuth         *middleware.ApiKeyAuth
	redisClient        *redis.Client
	notificationConfig handlers.NotificationConfig
}

func NewServer(db *database.Connection, processName logging.ProcessName) *Server {
	if !config.IsDevMode() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())

	logger := logging.GetServiceLogger()

	// Configure CORS
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, Content-Length, Accept-Encoding, Origin, X-Requested-With, X-CSRF-Token, X-Auth-Token, X-Api-Key")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "false")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	metricsServer := metrics.NewMetricsServer(db, logger)

	redisClient, err := redis.NewClient(logger)
	if err != nil {
		logger.Errorf("Failed to initialize Redis client: %v", err)
	}

	var rateLimiter *middleware.RateLimiter
	if redisClient != nil {
		rateLimiter, err = middleware.NewRateLimiterWithClient(redisClient, logger)
		if err != nil {
			logger.Errorf("Failed to initialize rate limiter: %v", err)
		}
	}

	s := &Server{
		router:        router,
		db:            db,
		logger:        logger,
		metricsServer: metricsServer,
		rateLimiter:   rateLimiter,
		redisClient:   redisClient,
		notificationConfig: handlers.NotificationConfig{
			EmailFrom:     config.GetEmailUser(),
			EmailPassword: config.GetEmailPassword(),
			BotToken:      config.GetBotToken(),
		},
	}

	s.apiKeyAuth = middleware.NewApiKeyAuth(db, rateLimiter, logger)

	return s
}

func (s *Server) RegisterRoutes(router *gin.Engine) {
	handler := handlers.NewHandler(s.db, s.logger, s.notificationConfig)

	api := router.Group("/api")

	// Public routes
	api.GET("/users/:id", handler.GetUserData)
	api.GET("/wallet/points/:wallet_address", handler.GetWalletPoints)

	protected := api.Group("")
	protected.Use(s.apiKeyAuth.GinMiddleware())

	api.POST("/jobs", handler.CreateJobData)
	api.GET("/jobs/:id", handler.GetJobData)
	api.PUT("/jobs/:id", handler.UpdateJobData)
	api.PUT("/jobs/:id/lastexecuted", handler.UpdateJobLastExecutedAt)
	api.GET("/jobs/user/:user_address", handler.GetJobsByUserAddress)
	api.PUT("/jobs/delete/:id", handler.DeleteJobData)

	api.POST("/tasks", handler.CreateTaskData)
	api.GET("/tasks/:id", handler.GetTaskData)
	api.PUT("/tasks/:id/fee", handler.UpdateTaskFee)

	api.GET("/keepers/all", handler.GetAllKeepers)
	api.GET("/keepers/performers", handler.GetPerformers)
	api.POST("/keepers/form", handler.CreateKeeperDataGoogleForm)
	api.GET("/keepers/:id", handler.GetKeeperData)
	api.POST("/keepers/:id/increment-tasks", handler.IncrementKeeperTaskCount)
	api.GET("/keepers/:id/task-count", handler.GetKeeperTaskCount)
	api.POST("/keepers/:id/add-points", handler.AddTaskFeeToKeeperPoints)
	api.GET("/keepers/:id/points", handler.GetKeeperPoints)

	api.GET("/leaderboard/keepers", handler.GetKeeperLeaderboard)
	api.GET("/leaderboard/users", handler.GetUserLeaderboard)
	api.GET("/leaderboard/users/search", handler.GetUserByAddress)
	api.GET("/leaderboard/keepers/search", handler.GetKeeperByIdentifier)

	api.GET("/fees", handler.GetTaskFees)

	api.POST("/keepers/update-chat-id", handler.UpdateKeeperChatID)
	api.GET("/keepers/com-info/:id", handler.GetKeeperCommunicationInfo)
	api.POST("/claim-fund", handler.ClaimFund)

	// Admin routes
	admin := protected.Group("/admin")
	admin.POST("/api-keys", handler.CreateApiKey)
	admin.PUT("/api-keys/:key", handler.UpdateApiKey)
	admin.DELETE("/api-keys/:key", handler.DeleteApiKey)
}

func (s *Server) Start(port string) error {
	s.logger.Infof("Starting server on port %s", port)

	s.metricsServer.Start()

	if s.redisClient != nil {
		defer s.redisClient.Close()
	}

	return s.router.Run(fmt.Sprintf(":%s", port))
}
