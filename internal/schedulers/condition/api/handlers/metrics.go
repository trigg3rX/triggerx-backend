package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend-imua/internal/schedulers/condition/metrics"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/logging"
)

type MetricsHandler struct {
	logger    logging.Logger
	collector *metrics.Collector
}

func NewMetricsHandler(logger logging.Logger) *MetricsHandler {
	return &MetricsHandler{
		logger:    logger,
		collector: metrics.NewCollector(),
	}
}

// Metrics serves Prometheus metrics
func (h *MetricsHandler) Metrics(c *gin.Context) {
	traceID := getTraceID(c)
	h.logger.Info("[Metrics] trace_id=" + traceID + " - Serving metrics")
	h.collector.Handler().ServeHTTP(c.Writer, c.Request)
}
