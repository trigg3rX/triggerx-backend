package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/registrar/api/handlers"
	"github.com/trigg3rX/triggerx-backend/internal/registrar/config"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type Server struct {
	router     *gin.Engine
	httpServer *http.Server
	logger     logging.Logger
}

func NewServer(logger logging.Logger) *Server {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	srv := &Server{
		router: router,
		logger: logger,
		httpServer: &http.Server{
			Addr:    fmt.Sprintf(":%s", config.GetRegistrarPort()),
			Handler: router,
		},
	}

	srv.setupRoutes()
	return srv
}

func (s *Server) setupRoutes() {
	handler := handlers.NewHandler(s.logger)

	s.router.Use(gin.Recovery())
	s.router.Use(LoggingMiddleware(s.logger))

	s.router.GET("/metrics", handler.HandleMetrics)
	s.router.GET("/status", handler.HandleStatus)

	// status := s.router.Group("/status")
	// {
	//     status.GET("/", s.handleStatus)
	// }
}
