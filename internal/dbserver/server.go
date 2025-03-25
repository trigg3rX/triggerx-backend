package dbserver

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/handlers"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/metrics"
)

type Server struct {
	router        *mux.Router
	db            *database.Connection
	cors          *cors.Cors
	logger        logging.Logger
	metricsServer *metrics.MetricsServer
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
		AllowedHeaders:   []string{"Content-Type", "Authorization", "Accept", "Content-Length", "Accept-Encoding", "Origin", "X-Requested-With", "X-CSRF-Token", "X-Auth-Token"},
		AllowCredentials: false,
		Debug:            true,
	})

	// Initialize metrics server
	metricsServer := metrics.NewMetricsServer(db, logger)

	s := &Server{
		router:        router,
		db:            db,
		cors:          corsHandler,
		logger:        logger,
		metricsServer: metricsServer,
	}

	s.routes()
	return s
}

func (s *Server) routes() {
	handler := handlers.NewHandler(s.db, s.logger)

	api := s.router.PathPrefix("/api").Subrouter()

	// User routes
	api.HandleFunc("/users/{id}", handler.GetUserData).Methods("GET")

	// // Job routes
	api.HandleFunc("/jobs", handler.CreateJobData).Methods("POST")
	api.HandleFunc("/jobs/{id}", handler.GetJobData).Methods("GET")
	api.HandleFunc("/jobs/{id}", handler.UpdateJobData).Methods("PUT")
	api.HandleFunc("/jobs/user/{user_address}", handler.GetJobsByUserAddress).Methods("GET")
	api.HandleFunc("/jobs/delete/{id}", handler.DeleteJobData).Methods("PUT")

	// // Task routes
	api.HandleFunc("/tasks", handler.CreateTaskData).Methods("POST")
	api.HandleFunc("/tasks/{id}", handler.GetTaskData).Methods("GET")

	// // Keeper routes
	api.HandleFunc("/keepers/all", handler.GetAllKeepers).Methods("GET")
	api.HandleFunc("/keepers/connection", handler.UpdateKeeperConnectionData).Methods("POST")
	api.HandleFunc("/keepers/performers", handler.GetPerformers).Methods("GET")
	api.HandleFunc("/keepers", handler.CreateKeeperData).Methods("POST")
	api.HandleFunc("/keepers/form", handler.GoogleFormCreateKeeperData).Methods("POST")
	api.HandleFunc("/keepers/{id}", handler.GetKeeperData).Methods("GET")
	api.HandleFunc("/keepers/{id}/increment-tasks", handler.IncrementKeeperTaskCount).Methods("POST")
	api.HandleFunc("/keepers/{id}/task-count", handler.GetKeeperTaskCount).Methods("GET")
	api.HandleFunc("/keepers/{id}/add-points", handler.AddTaskFeeToKeeperPoints).Methods("POST")
	api.HandleFunc("/keepers/{id}/points", handler.GetKeeperPoints).Methods("GET")

	api.HandleFunc("leaderboard/keepers", handler.GetKeeperLeaderboard).Methods("GET")
	

	// Fees routes
	api.HandleFunc("/fees", handler.GetTaskFees).Methods("GET")
}

func (s *Server) Start(port string) error {
	s.logger.Infof("Starting server on port %s", port)

	// Start the metrics server
	s.metricsServer.Start()

	handler := s.cors.Handler(s.router)
	return http.ListenAndServe(fmt.Sprintf(":%s", port), handler)
}
