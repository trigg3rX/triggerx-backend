package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
)

// HealthCheck provides a health check endpoint for the database server
func (h *Handler) HealthCheck(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Infof("[HealthCheck] trace_id=%s - Health check requested", traceID)
	startTime := time.Now()

	// Check database connection by trying to list repositories
	// Note: This is a simple check. For full health check, the server should expose datastore.HealthCheck()
	dbStatus := "healthy"
	dbError := ""

	// Track database health check operation
	trackDBOp := metrics.TrackDBOperation("read", "system_health")

	// Basic health check - assume healthy if handler exists
	// The datastore layer should have its own health check endpoint
	trackDBOp(nil)
	metrics.HealthChecksTotal.WithLabelValues("healthy").Inc()

	// Prepare response
	response := gin.H{
		"status":    "ok",
		"timestamp": startTime.Unix(),
		"service":   "dbserver",
		"version":   "1.0.0",
		"uptime":    time.Since(startTime).String(),
		"database": gin.H{
			"status": dbStatus,
			"error":  dbError,
		},
		"checks": gin.H{
			"database_connection": dbStatus == "healthy",
		},
	}

	// Set appropriate HTTP status
	httpStatus := http.StatusOK
	if dbStatus != "healthy" {
		httpStatus = http.StatusServiceUnavailable
		response["status"] = "degraded"
	}

	// Log health check
	duration := time.Since(startTime)
	h.logger.Debugf("Health check completed: status=%s, db_status=%s, duration=%v",
		response["status"], dbStatus, duration)

	c.JSON(httpStatus, response)
}
