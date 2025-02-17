package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/events"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type Server struct {
	router *mux.Router
	db     *database.Connection
	cors   *cors.Cors
	logger logging.Logger
}

func NewServer(db *database.Connection, processName logging.ProcessName) *Server {
	router := mux.NewRouter()

	// Get the logger instance
	logger := logging.GetLogger(logging.Development, processName)

	// Initialize event bus for the API service
	if err := events.InitEventBus("localhost:6379"); err != nil {
		logger.Fatalf("Failed to initialize event bus: %v", err)
	}

	// Create a new CORS handler
	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{
			"https://www.triggerx.network",
			"http://localhost:3000",
		},
		// AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "Accept", "Content-Length", "Accept-Encoding", "Origin", "X-Requested-With", "X-CSRF-Token", "X-Auth-Token"},
		AllowCredentials: true,
		Debug:            true,
	})

	s := &Server{
		router: router,
		db:     db,
		cors:   corsHandler,
		logger: logger,
	}

	s.routes()
	return s
}

func (s *Server) routes() {
	handler := NewHandler(s.db, s.logger)

	// Add the base /api prefix to all routes
	api := s.router.PathPrefix("/api").Subrouter()
	api.Use(mux.CORSMethodMiddleware(api))

	// User routes
	// api.HandleFunc("/users", handler.CreateUserData).Methods("POST")
	api.HandleFunc("/users/{id}", handler.GetUserData).Methods("GET")
	// api.HandleFunc("/users/{id}", handler.UpdateUserData).Methods("PUT")

	// // Job routes
	// // Job routes
	api.HandleFunc("/jobs/latest-id", handler.GetLatestJobID).Methods("GET")
	api.HandleFunc("/jobs", handler.CreateJobData).Methods("POST")
	api.HandleFunc("/jobs/{id}", handler.GetJobData).Methods("GET")
	api.HandleFunc("/jobs/{id}", handler.UpdateJobData).Methods("PUT")
	api.HandleFunc("/jobs/user/{user_address}", handler.GetJobsByUserAddress).Methods("GET")

	// // Task routes
	api.HandleFunc("/tasks", handler.CreateTaskData).Methods("POST")
	api.HandleFunc("/tasks/{id}", handler.GetTaskData).Methods("GET")

	// // Quorum routes
	api.HandleFunc("/quorums/all", handler.GetAllQuorums).Methods("GET")
	api.HandleFunc("/quorums/registration", handler.GetQuorumNoForRegistration).Methods("GET")
	api.HandleFunc("/quorums", handler.CreateQuorumData).Methods("POST")
	api.HandleFunc("/quorums/{id}", handler.GetQuorumData).Methods("GET")
	api.HandleFunc("/quorums/{id}", handler.UpdateQuorumData).Methods("PUT")

	// // Keeper routes
	api.HandleFunc("/get_peer_info/{id}", handler.GetKeeperPeerInfo).Methods("GET")
	api.HandleFunc("/keepers", handler.CreateKeeperData).Methods("POST")
	api.HandleFunc("/keepers/{id}", handler.GetKeeperData).Methods("GET")
	api.HandleFunc("/keepers/{id}", handler.UpdateKeeperData).Methods("PUT")
	api.HandleFunc("/keepers/address/{address}", handler.CheckKeeperRegistration).Methods("GET")

	// // Task History routes
	api.HandleFunc("/task_history", handler.CreateTaskHistory).Methods("POST")
	api.HandleFunc("/task_history/{id}", handler.GetTaskHistory).Methods("GET")
	// api.HandleFunc("/task_history/{id}", handler.UpdateTaskHistory).Methods("PUT")
	// api.HandleFunc("/task_history/{id}", handler.DeleteTaskHistory).Methods("DELETE")

	// Fees routes
	api.HandleFunc("/fees", handler.GetTaskFees).Methods("GET")
}

func (s *Server) Start(port string) error {
	s.logger.Infof("Starting server on port %s", port)

	// Wrap the router with the CORS handler
	handler := s.cors.Handler(s.router)

	return http.ListenAndServe(fmt.Sprintf(":%s", port), handler)
}
