package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

// MetricsHandler handles metrics endpoint requests
type MetricsHandler struct {
	logger    observability.Logger
	collector *metrics.Collector
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler(logger observability.Logger, metricsInstance observability.Metrics) *MetricsHandler {
	collector := metrics.NewCollector(metricsInstance)
	collector.Start()

	return &MetricsHandler{
		logger:    logger,
		collector: collector,
	}
}

// Metrics handles metrics endpoint requests
func (h *MetricsHandler) Metrics(c *gin.Context) {
	h.collector.Handler().ServeHTTP(c.Writer, c.Request)
}
