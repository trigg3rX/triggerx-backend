package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
)

// HealthCheck provides a health check endpoint for the database server
func (h *Handler) HealthCheck(c *gin.Context) {
	startTime := time.Now()

	// Check database connection by executing a simple query
	dbStatus := "healthy"
	dbError := ""

	// Track database health check operation
	trackDBOp := metrics.TrackDBOperation("read", "system_health")

	// Use a simple system query to test the connection
	query := h.db.Session().Query("SELECT now() FROM system.local")
	var timestamp time.Time
	if err := query.Scan(&timestamp); err != nil {
		dbStatus = "unhealthy"
		dbError = err.Error()
		h.logger.Errorf("Database health check failed: %v", err)
		trackDBOp(err)
		metrics.HealthChecksTotal.WithLabelValues("unhealthy").Inc()
	} else {
		trackDBOp(nil)
		metrics.HealthChecksTotal.WithLabelValues("healthy").Inc()
	}

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
	h.logger.Infof("Health check completed: status=%s, db_status=%s, duration=%v",
		response["status"], dbStatus, duration)

	c.JSON(httpStatus, response)
}
