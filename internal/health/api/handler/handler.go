package handler

import (
	"github.com/trigg3rX/triggerx-backend/internal/health/keeper"
	healthmetrics "github.com/trigg3rX/triggerx-backend/internal/health/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// Handler encapsulates the dependencies for health handlers
type Handler struct {
	logger        logging.Logger
	stateManager  *keeper.StateManager
	healthMetrics *healthmetrics.HealthMetrics
}

// NewHandler creates a new instance of Handler
func NewHandler(logger logging.Logger, stateManager *keeper.StateManager, healthMetrics *healthmetrics.HealthMetrics) *Handler {
	return &Handler{
		logger:        logger,
		stateManager:  stateManager,
		healthMetrics: healthMetrics,
	}
}
