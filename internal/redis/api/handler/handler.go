package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/trigg3rX/triggerx-backend-imua/internal/redis/metrics"
	"github.com/trigg3rX/triggerx-backend-imua/internal/redis/streams/jobs"
	"github.com/trigg3rX/triggerx-backend-imua/internal/redis/streams/tasks"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/logging"
)

// Handler encapsulates the dependencies for Redis handlers
type handler struct {
	logger           logging.Logger
	taskStreamMgr    *tasks.TaskStreamManager
	jobStreamMgr     *jobs.JobStreamManager
	metricsCollector *metrics.Collector
}

// NewHandler creates a new instance of Handler
func NewHandler(logger logging.Logger, taskStreamMgr *tasks.TaskStreamManager, jobStreamMgr *jobs.JobStreamManager, metricsCollector *metrics.Collector) *handler {
	return &handler{
		logger:           logger,
		taskStreamMgr:    taskStreamMgr,
		jobStreamMgr:     jobStreamMgr,
		metricsCollector: metricsCollector,
	}
}

// HandleRoot provides basic service information
func (h *handler) HandleRoot(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service":   "TriggerX Redis Service",
		"status":    "running",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   "1.0.0",
		"endpoints": []string{
			"GET /health - Service health check",
			"GET /metrics - Prometheus metrics",
			"POST /scheduler/submit-task - Submit task from scheduler",
			"POST /task/validate - Task validation",
			"POST /p2p/message - P2P message handling",
			"GET /streams/info - Stream information",
		},
	})
}

// HandleHealth provides detailed health information
func (h *handler) HandleHealth(c *gin.Context) {

	// Get stream info
	taskStreamInfo := h.taskStreamMgr.GetStreamInfo()
	jobStreamInfo := h.jobStreamMgr.GetJobStreamInfo()

	healthData := gin.H{
		"service":        "TriggerX Redis Service",
		"status":         "healthy",
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
		"task_streams":   taskStreamInfo,
		"job_streams":    jobStreamInfo,
		"uptime_seconds": time.Since(time.Now().Add(-time.Hour)).Seconds(), // This would be calculated from actual start time
	}

	c.JSON(http.StatusOK, healthData)
}

// HandleMetrics exposes Prometheus metrics
func (h *handler) HandleMetrics(c *gin.Context) {
	h.metricsCollector.Handler().ServeHTTP(c.Writer, c.Request)
}

// GetStreamsInfo provides detailed stream information
func (h *handler) GetStreamsInfo(c *gin.Context) {
	taskStreamInfo := h.taskStreamMgr.GetStreamInfo()
	jobStreamInfo := h.jobStreamMgr.GetJobStreamInfo()

	streamInfo := gin.H{
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
		"task_streams": taskStreamInfo,
		"job_streams":  jobStreamInfo,
	}

	c.JSON(http.StatusOK, streamInfo)
}

// SubmitTaskFromScheduler handles task submissions from schedulers
func (h *handler) SubmitTaskFromScheduler(c *gin.Context) {
	var request tasks.SchedulerTaskRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("[SubmitTaskFromScheduler] Failed to bind scheduler task request", "trace_id", getTraceID(c), "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	traceID := getTraceID(c)
	h.logger.Info("[SubmitTaskFromScheduler] Received task submission from scheduler", "trace_id", traceID,
		"task_id", request.SendTaskDataToKeeper.TaskID,
		"scheduler_id", request.SchedulerID,
		"source", request.Source)

	// Submit task to Redis orchestrator
	performerData, err := h.taskStreamMgr.ReceiveTaskFromScheduler(&request)
	if err != nil {
		h.logger.Error("[SubmitTaskFromScheduler] Failed to process scheduler task submission", "trace_id", traceID,
			"task_id", request.SendTaskDataToKeeper.TaskID,
			"scheduler_id", request.SchedulerID,
			"error", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to process task",
			"task_id": request.SendTaskDataToKeeper.TaskID,
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("[SubmitTaskFromScheduler] Task submitted successfully", "trace_id", traceID,
		"task_id", request.SendTaskDataToKeeper.TaskID,
		"performer_id", performerData.KeeperID,
		"performer_address", performerData.KeeperAddress)

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"task_id":   request.SendTaskDataToKeeper.TaskID,
		"message":   "Task submitted successfully",
		"performer": performerData,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// getTraceID retrieves the trace ID from the Gin context
func getTraceID(c *gin.Context) string {
	traceID, exists := c.Get("trace_id")
	if !exists {
		return ""
	}
	return traceID.(string)
}
