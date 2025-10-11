package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/errors"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// HealthCheck provides a health check endpoint for the database server
func (h *Handler) HealthCheck(c *gin.Context) {
	logger := h.getLogger(c)
	logger.Debugf("GET [HealthCheck] Requested")
	startTime := time.Now().UTC()

	// Check database connection
	trackDBOp := metrics.TrackDBOperation("read", "system_health")
	err := h.datastore.HealthCheck(c.Request.Context())
	trackDBOp(err)
	if err != nil {
		logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
		c.JSON(http.StatusServiceUnavailable, types.HealthCheckResponse{Status: "unhealthy", Error: errors.ErrDBOperationFailed})
		metrics.HealthChecksTotal.WithLabelValues("unhealthy").Inc()
		return
	}
	metrics.HealthChecksTotal.WithLabelValues("healthy").Inc()

	// Prepare response
	response := types.HealthCheckResponse{
		Status:    "healthy",
		Timestamp: startTime,
		Service:   "dbserver",
		Version:   "1.0.0",
	}

	logger.Info("GET [HealthCheck] Completed")
	c.JSON(http.StatusOK, response)
}
