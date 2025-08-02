package handlers

import (
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/scheduler"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type SchedulerHandler struct {
	logger    logging.Logger
	scheduler *scheduler.ConditionBasedScheduler
}

func NewSchedulerHandler(logger logging.Logger, scheduler *scheduler.ConditionBasedScheduler) *SchedulerHandler {
	return &SchedulerHandler{
		logger:    logger,
		scheduler: scheduler,
	}
}

// getTraceID retrieves the trace ID from the Gin context
func getTraceID(c *gin.Context) string {
	traceID, exists := c.Get("trace_id")
	if !exists {
		return ""
	}
	return traceID.(string)
}

// ScheduleJob schedules a new condition-based job
func (h *SchedulerHandler) ScheduleJob(c *gin.Context) {
	traceID := getTraceID(c)
	h.logger.Info("[ScheduleJob] trace_id=" + traceID + " - Scheduling job")

	var jobData types.ScheduleConditionJobData
	if err := c.ShouldBindJSON(&jobData); err != nil {
		h.logger.Error("Invalid request payload", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":    "error",
			"message":   "Invalid request payload",
			"error":     err.Error(),
			"timestamp": time.Now().UTC(),
		})
		return
	}

	// Schedule the job
	if err := h.scheduler.ScheduleJob(&jobData); err != nil {
		h.logger.Error("Failed to schedule condition job", "job_id", jobData.JobID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":    "error",
			"message":   "Failed to schedule condition job",
			"error":     err.Error(),
			"timestamp": time.Now().UTC(),
		})
		return
	}

	h.logger.Info("Condition job scheduled successfully", "job_id", jobData.JobID)

	response := gin.H{
		"status":    "success",
		"message":   "Condition job scheduled successfully",
		"job_id":    jobData.JobID,
		"timestamp": time.Now().UTC(),
	}

	c.JSON(http.StatusOK, response)
}

// UnscheduleJob unschedules a condition-based job
func (h *SchedulerHandler) UnscheduleJob(c *gin.Context) {
	traceID := getTraceID(c)
	h.logger.Info("[UnscheduleJob] trace_id=" + traceID + " - Unscheduling job")

	jobIDStr := c.Param("job_id")
	jobID := new(big.Int)
	_, ok := jobID.SetString(jobIDStr, 10)
	if !ok {
		err := fmt.Errorf("invalid job ID: %s", jobIDStr)
		h.logger.Error("Invalid job ID", "job_id", jobIDStr, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":    "error",
			"message":   "Invalid job ID",
			"error":     err.Error(),
			"timestamp": time.Now().UTC(),
		})
		return
	}

	// Unschedule the job
	if err := h.scheduler.UnscheduleJob(jobID); err != nil {
		h.logger.Error("Failed to unschedule condition job", "job_id", jobID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":    "error",
			"message":   "Failed to unschedule condition job",
			"error":     err.Error(),
			"timestamp": time.Now().UTC(),
		})
		return
	}

	h.logger.Info("Condition job unscheduled successfully", "job_id", jobID)

	response := gin.H{
		"status":    "success",
		"message":   "Condition job unscheduled successfully",
		"job_id":    jobID,
		"timestamp": time.Now().UTC(),
	}

	c.JSON(http.StatusOK, response)
}

// GetJobStats returns statistics for a specific condition job
func (h *SchedulerHandler) GetJobStats(c *gin.Context) {
	traceID := getTraceID(c)
	h.logger.Info("[GetJobStats] trace_id=" + traceID + " - Getting job stats")

	jobIDStr := c.Param("job_id")
	jobID := new(big.Int)
	_, ok := jobID.SetString(jobIDStr, 10)
	if !ok {
		err := fmt.Errorf("invalid job ID: %s", jobIDStr)
		h.logger.Error("Invalid job ID", "job_id", jobIDStr, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":    "error",
			"message":   "Invalid job ID",
			"error":     err.Error(),
			"timestamp": time.Now().UTC(),
		})
		return
	}

	stats := h.scheduler.GetStats()
	if stats == nil {
		h.logger.Error("Failed to get condition job stats", "job_id", jobID)
		c.JSON(http.StatusNotFound, gin.H{
			"status":    "error",
			"message":   "Condition job not found",
			"error":     "condition job not found",
			"timestamp": time.Now().UTC(),
		})
		return
	}

	response := gin.H{
		"status":    "success",
		"data":      stats,
		"timestamp": time.Now().UTC(),
	}

	c.JSON(http.StatusOK, response)
}

// GetStats returns current scheduler statistics
func (h *SchedulerHandler) GetStats(c *gin.Context) {
	traceID := getTraceID(c)
	h.logger.Info("[GetStats] trace_id=" + traceID + " - Getting scheduler stats")

	stats := h.scheduler.GetStats()

	response := gin.H{
		"status":    "success",
		"data":      stats,
		"timestamp": time.Now().UTC(),
	}

	c.JSON(http.StatusOK, response)
}
