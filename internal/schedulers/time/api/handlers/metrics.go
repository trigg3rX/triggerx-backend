package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

// MetricsHandler handles metrics endpoint requests
type MetricsHandler struct {
	logger observability.Logger
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler(logger observability.Logger) *MetricsHandler {
	return &MetricsHandler{
		logger: logger,
	}
}

// Metrics exposes Prometheus metrics
func (h *MetricsHandler) Metrics(c *gin.Context) {
	traceID := getTraceID(c)
	h.logger.Info(c.Request.Context(), "[Metrics] trace_id=" + traceID + " - Serving metrics")
	promhttp.Handler().ServeHTTP(c.Writer, c.Request)
}
