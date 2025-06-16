package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/trigg3rX/triggerx-backend/internal/redis"
	"github.com/trigg3rX/triggerx-backend/internal/redis/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// Handler encapsulates the dependencies for Redis handlers
type Handler struct {
	logger           logging.Logger
	taskStreamMgr    *redis.TaskStreamManager
	jobStreamMgr     *redis.JobStreamManager
	metricsCollector *metrics.Collector
}

// NewHandler creates a new instance of Handler
func NewHandler(logger logging.Logger, taskStreamMgr *redis.TaskStreamManager, jobStreamMgr *redis.JobStreamManager, metricsCollector *metrics.Collector) *Handler {
	return &Handler{
		logger:           logger,
		taskStreamMgr:    taskStreamMgr,
		jobStreamMgr:     jobStreamMgr,
		metricsCollector: metricsCollector,
	}
}

// HandleRoot provides basic service information
func (h *Handler) HandleRoot(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service":   "TriggerX Redis Service",
		"status":    "running",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   "1.0.0",
		"endpoints": []string{
			"GET /health - Service health check",
			"GET /metrics - Prometheus metrics",
			"POST /task/validate - Task validation",
			"POST /p2p/message - P2P message handling",
			"GET /streams/info - Stream information",
		},
	})
}

// HandleHealth provides detailed health information
func (h *Handler) HandleHealth(c *gin.Context) {
	// Get Redis info
	redisInfo := redis.GetRedisInfo()

	// Get stream info
	taskStreamInfo := h.taskStreamMgr.GetStreamInfo()
	jobStreamInfo := h.jobStreamMgr.GetStreamInfo()

	healthData := gin.H{
		"service":        "TriggerX Redis Service",
		"status":         "healthy",
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
		"redis":          redisInfo,
		"task_streams":   taskStreamInfo,
		"job_streams":    jobStreamInfo,
		"uptime_seconds": time.Since(time.Now().Add(-time.Hour)).Seconds(), // This would be calculated from actual start time
	}

	c.JSON(http.StatusOK, healthData)
}

// HandleMetrics exposes Prometheus metrics
func (h *Handler) HandleMetrics(c *gin.Context) {
	h.metricsCollector.Handler().ServeHTTP(c.Writer, c.Request)
}

// GetStreamsInfo provides detailed stream information
func (h *Handler) GetStreamsInfo(c *gin.Context) {
	taskStreamInfo := h.taskStreamMgr.GetStreamInfo()
	jobStreamInfo := h.jobStreamMgr.GetStreamInfo()

	streamInfo := gin.H{
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
		"task_streams": taskStreamInfo,
		"job_streams":  jobStreamInfo,
		"redis_info":   redis.GetRedisInfo(),
	}

	c.JSON(http.StatusOK, streamInfo)
}
