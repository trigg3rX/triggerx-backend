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
    api.HandleFunc("/users/{id}", handler.UpdateUserData).Methods("PUT")
    api.HandleFunc("/users/{id}", handler.DeleteUserData).Methods("DELETE")
    
    // Job routes
    api.HandleFunc("/jobs", handler.CreateJobData).Methods("POST")
    api.HandleFunc("/jobs/{id}", handler.GetJobData).Methods("GET")
    api.HandleFunc("/jobs/{id}", handler.UpdateJobData).Methods("PUT")
    api.HandleFunc("/jobs/{id}", handler.DeleteJobData).Methods("DELETE")
    
    // Task routes
    api.HandleFunc("/tasks", handler.CreateTaskData).Methods("POST")
    api.HandleFunc("/tasks/{id}", handler.GetTaskData).Methods("GET")
    api.HandleFunc("/tasks/{id}", handler.UpdateTaskData).Methods("PUT")
    api.HandleFunc("/tasks/{id}", handler.DeleteTaskData).Methods("DELETE")

    // Quorum routes
    api.HandleFunc("/quorums", handler.CreateQuorumData).Methods("POST")
    api.HandleFunc("/quorums/{id}", handler.GetQuorumData).Methods("GET")
    api.HandleFunc("/quorums/{id}", handler.UpdateQuorumData).Methods("PUT")
    api.HandleFunc("/quorums/{id}", handler.DeleteQuorumData).Methods("DELETE")

    // Keeper routes
    api.HandleFunc("/keepers", handler.CreateKeeperData).Methods("POST")
    api.HandleFunc("/keepers/{id}", handler.GetKeeperData).Methods("GET")
    api.HandleFunc("/keepers/{id}", handler.UpdateKeeperData).Methods("PUT")
    api.HandleFunc("/keepers/{id}", handler.DeleteKeeperData).Methods("DELETE")

    // Task History routes
    api.HandleFunc("/task_history", handler.CreateTaskHistory).Methods("POST")
    api.HandleFunc("/task_history/{id}", handler.GetTaskHistory).Methods("GET")
    api.HandleFunc("/task_history/{id}", handler.UpdateTaskHistory).Methods("PUT")
    api.HandleFunc("/task_history/{id}", handler.DeleteTaskHistory).Methods("DELETE")
}

func (s *Server) Start(port string) error {
    log.Printf("Starting server on port %s", port)
    return http.ListenAndServe(fmt.Sprintf(":%s", port), s.router)
} 