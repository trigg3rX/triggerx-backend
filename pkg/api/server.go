package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/eventbus"
)

type Server struct {
	router   *mux.Router
	db       *database.Connection
	cors     *cors.Cors
	eventBus *eventbus.EventBus
}

func NewServer(db *database.Connection) *Server {
	router := mux.NewRouter()

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

	// Create new event bus
	eb := eventbus.New()

	// Start WebSocket server on port 8081
	eb.StartWebSocketServer("8081")

	s := &Server{
		router:   router,
		db:       db,
		cors:     corsHandler,
		eventBus: eb,
	}

	s.routes()
	return s
}

func (s *Server) routes() {
	// Create handler with both database and event bus
	handler := NewHandler(s.db, s.eventBus)

	// Add the base /api prefix to all routes
	api := s.router.PathPrefix("/api").Subrouter()
	api.Use(mux.CORSMethodMiddleware(api)) // For preflight requests

	// User routes
	api.HandleFunc("/users", handler.CreateUserData).Methods("POST")
	api.HandleFunc("/users/{id}", handler.GetUserData).Methods("GET")
	api.HandleFunc("/users/{id}", handler.UpdateUserData).Methods("PUT")
	// api.HandleFunc("/users/{id}", handler.DeleteUserData).Methods("DELETE")

	// Job routes
	api.HandleFunc("/jobs/latest-id", handler.GetLatestJobID).Methods("GET")
	api.HandleFunc("/jobs", handler.CreateJobData).Methods("POST")
	api.HandleFunc("/jobs/{id}", handler.GetJobData).Methods("GET")
	api.HandleFunc("/jobs/{id}", handler.UpdateJobData).Methods("PUT")
	// api.HandleFunc("/jobs/{id}", handler.DeleteJobData).Methods("DELETE")
	api.HandleFunc("/jobs/user/{user_address}", handler.GetJobsByUserAddress).Methods("GET")

	// Task routes
	api.HandleFunc("/tasks", handler.CreateTaskData).Methods("POST")
	api.HandleFunc("/tasks/{id}", handler.GetTaskData).Methods("GET")
	// api.HandleFunc("/tasks/{id}", handler.UpdateTaskData).Methods("PUT")
	// api.HandleFunc("/tasks/{id}", handler.DeleteTaskData).Methods("DELETE")

	// Quorum routes
	api.HandleFunc("/quorums", handler.CreateQuorumData).Methods("POST")
	api.HandleFunc("/quorums/{id}", handler.GetQuorumData).Methods("GET")
	api.HandleFunc("/quorums/{id}", handler.UpdateQuorumData).Methods("PUT")
	api.HandleFunc("/quorums/free", handler.GetFreeQuorum).Methods("GET")
	// api.HandleFunc("/quorums/{id}", handler.DeleteQuorumData).Methods("DELETE")

	// Keeper routes
	api.HandleFunc("/keepers", handler.CreateKeeperData).Methods("POST")
	api.HandleFunc("/keepers/{id}", handler.GetKeeperData).Methods("GET")
	api.HandleFunc("/keepers/{id}", handler.UpdateKeeperData).Methods("PUT")
	// api.HandleFunc("/keepers/{id}", handler.DeleteKeeperData).Methods("DELETE")

	// Task History routes
	api.HandleFunc("/task_history", handler.CreateTaskHistory).Methods("POST")
	api.HandleFunc("/task_history/{id}", handler.GetTaskHistory).Methods("GET")
	// api.HandleFunc("/task_history/{id}", handler.UpdateTaskHistory).Methods("PUT")
	// api.HandleFunc("/task_history/{id}", handler.DeleteTaskHistory).Methods("DELETE")
}

func (s *Server) Start(port string) error {
	log.Printf("Starting server on port %s", port)

	// Wrap the router with the CORS handler
	handler := s.cors.Handler(s.router)

	return http.ListenAndServe(fmt.Sprintf(":%s", port), handler)
}

// GetEventBus returns the event bus instance
// This can be useful for other parts of the application that need to subscribe to events
func (s *Server) GetEventBus() *eventbus.EventBus {
	return s.eventBus
}
