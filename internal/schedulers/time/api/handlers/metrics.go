package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/logging"
)

// MetricsHandler handles metrics endpoint requests
type MetricsHandler struct {
	logger logging.Logger
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler(logger logging.Logger) *MetricsHandler {
	return &MetricsHandler{
		logger: logger,
	}
}

// Metrics exposes Prometheus metrics
func (h *MetricsHandler) Metrics(c *gin.Context) {
	h.logger.Debug("Serving metrics endpoint")
	promhttp.Handler().ServeHTTP(c.Writer, c.Request)
}
