package dbserver

import (
	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/api"
	"github.com/trigg3rX/triggerx-backend/pkg/datastore"
	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// Server is the orchestration layer for the database server service
// It wraps the API server and provides a clean interface for the main entry point
type Server struct {
	apiServer *api.Server
}

// NewServer creates a new database server instance
// This is the main entry point for initializing the database server service
func NewServer(datastore datastore.DatastoreService, logger logging.Logger) *Server {
	apiServer := api.NewServer(datastore, logger)

	return &Server{
		apiServer: apiServer,
	}
}

// RegisterRoutes registers all API routes with the given router
// This includes protected and public endpoints, WebSocket routes, and admin routes
func (s *Server) RegisterRoutes(router *gin.Engine, dockerExecutor dockerexecutor.DockerExecutorAPI) {
	s.apiServer.RegisterRoutes(router, dockerExecutor)
}

// GetRouter returns the underlying Gin router
// This is used for custom server configuration (e.g., in tests or custom deployments)
func (s *Server) GetRouter() *gin.Engine {
	return s.apiServer.GetRouter()
}

// Start starts the HTTP server on the specified port
// This is a convenience method for simple deployments
// For production use with graceful shutdown, use GetRouter() with a custom http.Server
func (s *Server) Start(port string) error {
	return s.apiServer.Start(port)
}
