package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

// MetricsHandler handles metrics endpoint requests
type MetricsHandler struct {
	logger    observability.Logger
	collector *metrics.Collector
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler(logger observability.Logger, collector *metrics.Collector) *MetricsHandler {
	return &MetricsHandler{
		logger:    logger,
		collector: collector,
	}
}

// Metrics exposes Prometheus metrics
func (h *MetricsHandler) Metrics(c *gin.Context) {
	h.collector.Handler().ServeHTTP(c.Writer, c.Request)
}
