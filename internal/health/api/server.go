package api

import (
	"github.com/gin-gonic/gin"

	"github.com/trigg3rX/triggerx-backend/internal/health/api/handler"
	"github.com/trigg3rX/triggerx-backend/internal/health/api/middleware"
	"github.com/trigg3rX/triggerx-backend/internal/health/keeper"
	healthmetrics "github.com/trigg3rX/triggerx-backend/internal/health/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/health/rewards"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/metrics"
)

// RegisterRoutes registers all HTTP routes for the health service
func RegisterRoutes(router *gin.Engine, logger logging.Logger, rewardsService *rewards.Service) {
	// Initialize metrics collector
	metricsCollector := metrics.NewCollector("health")
	metricsCollector.Start()

	// Initialize health-specific metrics
	hm := healthmetrics.NewHealthMetrics(metricsCollector)

	// Apply logger middleware with metrics
	router.Use(middleware.Logger(logger, hm))

	// Create handler with metrics and rewards service
	h := handler.NewHandler(logger, keeper.GetStateManager(), hm, rewardsService)

	// Register routes
	router.GET("/", h.HandleRoot)
	router.POST("/health", h.HandleCheckInEvent)
	router.GET("/status", h.GetKeeperStatus)
	router.GET("/operators", h.GetDetailedKeeperStatus)
	router.GET("/rewards/health", h.HandleGetRewardsHealth)
	router.GET("/rewards/uptime", h.HandleGetKeeperDailyUptime)
	router.GET("/metrics", gin.WrapH(metricsCollector.Handler()))
}
