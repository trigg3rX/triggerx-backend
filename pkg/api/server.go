package api

import (
    "fmt"
    "log"
    "net/http"
    "github.com/gorilla/mux"
    "github.com/trigg3rX/go-backend/pkg/database"
)

type Server struct {
    router *mux.Router
    db     *database.Connection
}

func NewServer(db *database.Connection) *Server {
    s := &Server{
        router: mux.NewRouter(),
        db:     db,
    }
    s.routes()
    return s
}

func (s *Server) routes() {
    handler := NewHandler(s.db)
    
    // Add the base /api prefix to all routes
    api := s.router.PathPrefix("/api").Subrouter()
    
    // User routes
    api.HandleFunc("/users", handler.CreateUserData).Methods("POST")
    api.HandleFunc("/users/{id}", handler.GetUserData).Methods("GET")
    
    // Job routes
    api.HandleFunc("/jobs", handler.CreateJobData).Methods("POST")
    api.HandleFunc("/jobs/{id}", handler.GetJobData).Methods("GET")
    
    // Task routes
    api.HandleFunc("/tasks", handler.CreateTaskData).Methods("POST")
    api.HandleFunc("/tasks/{id}", handler.GetTaskData).Methods("GET")
}

func (s *Server) Start(port string) error {
    log.Printf("Starting server on port %s", port)
    return http.ListenAndServe(fmt.Sprintf(":%s", port), s.router)
} 