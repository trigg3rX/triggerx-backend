package api

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all the manager service routes
func RegisterRoutes(router *gin.Engine, h *Handlers) {
	// Job management routes
	jobGroup := router.Group("/job")
	{
		jobGroup.POST("/create", h.HandleCreateJobEvent)
		jobGroup.POST("/update", h.HandleUpdateJobEvent)
		jobGroup.POST("/pause", h.HandlePauseJobEvent)
		jobGroup.POST("/resume", h.HandleResumeJobEvent)
		jobGroup.POST("/state/update", h.HandleJobStateUpdate)
	}

	// Task management routes
	taskGroup := router.Group("/task")
	{
		taskGroup.POST("/validate", h.ValidateTask)
	}

	// P2P communication routes
	p2pGroup := router.Group("/p2p")
	{
		p2pGroup.POST("/message", h.ExecuteTask)
	}
}
