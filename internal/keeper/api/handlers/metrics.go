package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend-imua/internal/keeper/metrics"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/logging"
)

// MetricsHandler handles metrics endpoint requests
type MetricsHandler struct {
	logger    logging.Logger
	collector *metrics.Collector
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler(logger logging.Logger) *MetricsHandler {
	collector := metrics.NewCollector()
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
