package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

type MetricsHandler struct {
	logger    observability.Logger
	collector *metrics.Collector
}

func NewMetricsHandler(logger observability.Logger) *MetricsHandler {
	return &MetricsHandler{
		logger:    logger,
		collector: metrics.NewCollector(),
	}
}

// Metrics serves Prometheus metrics
func (h *MetricsHandler) Metrics(c *gin.Context) {
	traceID := getTraceID(c)
	h.logger.Info(c.Request.Context(), "[Metrics] trace_id=" + traceID + " - Serving metrics")
	h.collector.Handler().ServeHTTP(c.Writer, c.Request)
}
