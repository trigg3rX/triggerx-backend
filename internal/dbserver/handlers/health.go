package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/observability/tracing"
)

// HealthCheck provides a health check endpoint for the database server
func (h *Handler) HealthCheck(c *gin.Context) {
	startTime := time.Now()

	// Check database connection by executing a simple query
	dbStatus := "healthy"
	dbError := ""

	// Track database health check operation
	trackDBOp := metrics.TrackDBOperation("read", "system_health")

	// Add OpenTelemetry database tracing
	dbTracer := tracing.NewDatabaseTracer("triggerx-dbserver")
	query := "SELECT now() FROM system.local"
	traceDBOp := dbTracer.TraceDBOperation(c.Request.Context(), "SELECT", "system_health", query)

	// Use a simple system query to test the connection
	queryObj := h.db.Session().Query(query)
	var timestamp time.Time
	if err := queryObj.Scan(&timestamp); err != nil {
		dbStatus = "unhealthy"
		dbError = err.Error()
		h.logger.Errorf("Database health check failed: %v", err)
		trackDBOp(err)
		traceDBOp(err)
		metrics.HealthChecksTotal.WithLabelValues("unhealthy").Inc()
	} else {
		trackDBOp(nil)
		traceDBOp(nil)
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
	h.logger.Debugf("Health check completed: status=%s, db_status=%s, duration=%v",
		response["status"], dbStatus, duration)

	c.JSON(httpStatus, response)
}
