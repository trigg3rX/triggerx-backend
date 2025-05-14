package dbserver

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/config"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/handlers"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/middleware"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/telegram"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/redis"
)

type Server struct {
	router             *mux.Router
	db                 *database.Connection
	cors               *cors.Cors
	logger             logging.Logger
	metricsServer      *metrics.MetricsServer
	rateLimiter        *middleware.RateLimiter
	apiKeyAuth         *middleware.ApiKeyAuth
	redisClient        *redis.Client
	notificationConfig handlers.NotificationConfig
	telegramBot        *telegram.Bot
}

func NewServer(db *database.Connection, processName logging.ProcessName) *Server {
	router := mux.NewRouter()

	logger := logging.GetLogger(logging.Development, processName)

	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{"*",
			"https://app.triggerx.network",
			"https://www.triggerx.network",
			"http://localhost:3000",
			"http://localhost:3001",
			"https://data.triggerx.network",
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "Accept", "Content-Length", "Accept-Encoding", "Origin", "X-Requested-With", "X-CSRF-Token", "X-Auth-Token", "X-Api-Key"},
		AllowCredentials: false,
		Debug:            true,
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

	bot, err := telegram.NewBot(os.Getenv("BOT_TOKEN"), logger, db)
	if err != nil {
		logger.Errorf("Failed to initialize Telegram bot: %v", err)
	}

	s := &Server{
		router:        router,
		db:            db,
		cors:          corsHandler,
		logger:        logger,
		metricsServer: metricsServer,
		rateLimiter:   rateLimiter,
		redisClient:   redisClient,
		telegramBot:   bot,
		notificationConfig: handlers.NotificationConfig{
			EmailFrom:     config.EmailUser,
			EmailPassword: config.EmailPassword,
			BotToken:      config.BotToken,
		},
	}

	s.apiKeyAuth = middleware.NewApiKeyAuth(db, rateLimiter, logger)

	s.routes()
	return s
}

func (s *Server) routes() {
	handler := handlers.NewHandler(s.db, s.logger, s.notificationConfig)

	api := s.router.PathPrefix("/api").Subrouter()

	protected := api.PathPrefix("").Subrouter()
	protected.Use(s.apiKeyAuth.Middleware)

	api.HandleFunc("/users/{id}", handler.GetUserData).Methods("GET")
	api.HandleFunc("/wallet/points/{wallet_address}", handler.GetWalletPoints).Methods("GET")

	api.HandleFunc("/jobs", handler.CreateJobData).Methods("POST")
	api.HandleFunc("/jobs/{id}", handler.GetJobData).Methods("GET")
	api.HandleFunc("/jobs/{id}", handler.UpdateJobData).Methods("PUT")
	api.HandleFunc("/jobs/{id}/lastexecuted", handler.UpdateJobLastExecutedAt).Methods("PUT")
	api.HandleFunc("/jobs/user/{user_address}", handler.GetJobsByUserAddress).Methods("GET")
	api.HandleFunc("/jobs/delete/{id}", handler.DeleteJobData).Methods("PUT")

	api.HandleFunc("/tasks", handler.CreateTaskData).Methods("POST")
	api.HandleFunc("/tasks/{id}", handler.GetTaskData).Methods("GET")
	api.HandleFunc("/tasks/{id}/fee", handler.UpdateTaskFee).Methods("PUT")

	api.HandleFunc("/keepers/all", handler.GetAllKeepers).Methods("GET")
	api.HandleFunc("/keepers/performers", handler.GetPerformers).Methods("GET")
	api.HandleFunc("/keepers/form", handler.CreateKeeperDataGoogleForm).Methods("POST")
	api.HandleFunc("/keepers/checkin", handler.KeeperHealthCheckIn).Methods("POST")
	api.HandleFunc("/keepers/{id}", handler.GetKeeperData).Methods("GET")
	api.HandleFunc("/keepers/{id}/increment-tasks", handler.IncrementKeeperTaskCount).Methods("POST")
	api.HandleFunc("/keepers/{id}/task-count", handler.GetKeeperTaskCount).Methods("GET")
	api.HandleFunc("/keepers/{id}/add-points", handler.AddTaskFeeToKeeperPoints).Methods("POST")
	api.HandleFunc("/keepers/{id}/points", handler.GetKeeperPoints).Methods("GET")

	api.HandleFunc("/leaderboard/keepers", handler.GetKeeperLeaderboard).Methods("GET")
	api.HandleFunc("/leaderboard/users", handler.GetUserLeaderboard).Methods("GET")
	api.HandleFunc("/leaderboard/users/search", handler.GetUserByAddress).Methods("GET")
	api.HandleFunc("/leaderboard/keepers/search", handler.GetKeeperByIdentifier).Methods("GET")

	api.HandleFunc("/fees", handler.GetTaskFees).Methods("GET")

	api.HandleFunc("/keepers/update-chat-id", handler.UpdateKeeperChatID).Methods("POST")

	api.HandleFunc("/keepers/com-info/{id}", handler.GetKeeperCommunicationInfo).Methods("GET")

	api.HandleFunc("/claim-fund", handler.ClaimFund).Methods("POST")

	admin := api.PathPrefix("/admin").Subrouter()
	admin.HandleFunc("/api-keys", handler.CreateApiKey).Methods("POST")
	admin.HandleFunc("/api-keys/{key}", handler.UpdateApiKey).Methods("PUT")
	admin.HandleFunc("/api-keys/{key}", handler.DeleteApiKey).Methods("DELETE")
}

func (s *Server) Start(port string) error {
	s.logger.Infof("Starting server on port %s", port)

	s.metricsServer.Start()

	if s.redisClient != nil {
		defer s.redisClient.Close()
	}

	if s.telegramBot != nil {
		go s.telegramBot.Start()
	}

	handler := s.cors.Handler(s.router)
	return http.ListenAndServe(fmt.Sprintf(":%s", port), handler)
}
