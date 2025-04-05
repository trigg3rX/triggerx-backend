package dbserver

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/handlers"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/middleware"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/redis"
)

type Server struct {
	router        *mux.Router
	db            *database.Connection
	cors          *cors.Cors
	logger        logging.Logger
	metricsServer *metrics.MetricsServer
	rateLimiter   *middleware.RateLimiter
	apiKeyAuth    *middleware.ApiKeyAuth
	redisClient   *redis.Client
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

	// Initialize metrics server
	metricsServer := metrics.NewMetricsServer(db, logger)

	// Initialize Redis client
	redisClient, err := redis.NewClient(logger)
	if err != nil {
		logger.Errorf("Failed to initialize Redis client: %v", err)
		// Continue without Redis if unavailable, but this will affect rate limiting
	}
	
	// Initialize rate limiter with our Redis client
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
		cors:          corsHandler,
		logger:        logger,
		metricsServer: metricsServer,
		rateLimiter:   rateLimiter,
		redisClient:   redisClient,
	}
	
	// Initialize API key middleware
	s.apiKeyAuth = middleware.NewApiKeyAuth(db, rateLimiter, logger)

	s.routes()
	return s
}

func (s *Server) routes() {
	handler := handlers.NewHandler(s.db, s.logger)

	api := s.router.PathPrefix("/api").Subrouter()
	
	// Public routes (no API key required)
	// You can add routes here that don't need authentication
	
	// Protected routes (API key required)
	protected := api.PathPrefix("").Subrouter()
	protected.Use(s.apiKeyAuth.Middleware)
	
	// User routes
	api.HandleFunc("/users/{id}", handler.GetUserData).Methods("GET")
	api.HandleFunc("/wallet/points/{wallet_address}", handler.GetWalletPoints).Methods("GET")

	// Job routes
	protected.HandleFunc("/jobs", handler.CreateJobData).Methods("POST")
	protected.HandleFunc("/jobs/{id}", handler.GetJobData).Methods("GET")
	protected.HandleFunc("/jobs/{id}", handler.UpdateJobData).Methods("PUT")
	protected.HandleFunc("/jobs/user/{user_address}", handler.GetJobsByUserAddress).Methods("GET")
	protected.HandleFunc("/jobs/delete/{id}", handler.DeleteJobData).Methods("PUT")

	// // Task routes
	api.HandleFunc("/tasks", handler.CreateTaskData).Methods("POST")
	api.HandleFunc("/tasks/{id}", handler.GetTaskData).Methods("GET")

	// // Keeper routes
	api.HandleFunc("/keepers/all", handler.GetAllKeepers).Methods("GET")
	api.HandleFunc("/keepers/performers", handler.GetPerformers).Methods("GET")
	api.HandleFunc("/keepers", handler.CreateKeeperData).Methods("POST")
	api.HandleFunc("/keepers/form", handler.GoogleFormCreateKeeperData).Methods("POST")
	api.HandleFunc("/keepers/checkin", handler.KeeperHealthCheckIn).Methods("POST")
	api.HandleFunc("/keepers/{id}", handler.GetKeeperData).Methods("GET")
	api.HandleFunc("/keepers/{id}/increment-tasks", handler.IncrementKeeperTaskCount).Methods("POST")
	api.HandleFunc("/keepers/{id}/task-count", handler.GetKeeperTaskCount).Methods("GET")
	api.HandleFunc("/keepers/{id}/add-points", handler.AddTaskFeeToKeeperPoints).Methods("POST")
	api.HandleFunc("/keepers/{id}/points", handler.GetKeeperPoints).Methods("GET")

	api.HandleFunc("/leaderboard/keepers", handler.GetKeeperLeaderboard).Methods("GET")
	api.HandleFunc("/leaderboard/user", handler.GetUserLeaderboard).Methods("GET")

	// Fees routes
	api.HandleFunc("/fees", handler.GetTaskFees).Methods("GET")
	
	// API key management routes (these should be admin-only and properly secured)
	admin := api.PathPrefix("/admin").Subrouter()
	// Add authentication for admin routes here
	admin.HandleFunc("/api-keys", handler.CreateApiKey).Methods("POST")
	admin.HandleFunc("/api-keys/{key}", handler.UpdateApiKey).Methods("PUT")
	admin.HandleFunc("/api-keys/{key}", handler.DeleteApiKey).Methods("DELETE")
}

func (s *Server) Start(port string) error {
	s.logger.Infof("Starting server on port %s", port)

	// Start the metrics server
	s.metricsServer.Start()

	// Defer closing Redis client when server stops
	if s.redisClient != nil {
		defer s.redisClient.Close()
	}

	handler := s.cors.Handler(s.router)
	return http.ListenAndServe(fmt.Sprintf(":%s", port), handler)
}