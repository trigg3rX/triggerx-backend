package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/errors"
)

// HealthCheck provides a health check endpoint for the database server
func (h *Handler) HealthCheck(c *gin.Context) {
	logger := h.getLogger(c)
	logger.Debugf("GET [HealthCheck] Health check requested")
	startTime := time.Now()

	// Check database connection
	trackDBOp := metrics.TrackDBOperation("read", "system_health")
	err := h.datastore.HealthCheck(c.Request.Context())
	trackDBOp(err)
	if err != nil {
		logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "error": errors.ErrDBOperationFailed})
		return
	}
	metrics.HealthChecksTotal.WithLabelValues("healthy").Inc()

	// Prepare response
	response := gin.H{
		"status":    "ok",
		"timestamp": startTime.Unix(),
		"service":   "dbserver",
		"version":   "1.0.0",
		"uptime":    time.Since(startTime).String(),
		"database": gin.H{
			"status": "healthy",
			"error":  "",
		},
		"checks": gin.H{
			"database_connection": err == nil,
		},
	}

	logger.Debugf("Health check completed")
	c.JSON(http.StatusOK, response)
}
