package handlers

import (
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/scheduler"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type SchedulerHandler struct {
	logger    observability.Logger
	scheduler *scheduler.ConditionBasedScheduler
}

func NewSchedulerHandler(logger observability.Logger, scheduler *scheduler.ConditionBasedScheduler) *SchedulerHandler {
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
	h.logger.Info(c.Request.Context(), "[ScheduleJob] trace_id="+traceID+" - Scheduling job")

	var jobData types.ScheduleConditionJobData
	if err := c.ShouldBindJSON(&jobData); err != nil {
		h.logger.Error(c.Request.Context(), "Invalid request payload", observability.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"status":    "error",
			"message":   "Invalid request payload",
			"error":     err.Error(),
			"timestamp": time.Now().UTC(),
		})
		return
	}

	// Schedule the job
	if err := h.scheduler.ScheduleJob(c.Request.Context(), &jobData); err != nil {
		h.logger.Error(c.Request.Context(), "Failed to schedule condition job", observability.String("job_id", jobData.JobID.String()), observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":    "error",
			"message":   "Failed to schedule condition job",
			"error":     err.Error(),
			"timestamp": time.Now().UTC(),
		})
		return
	}

	h.logger.Info(c.Request.Context(), "Condition job scheduled successfully", observability.String("job_id", jobData.JobID.String()))

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
	h.logger.Info(c.Request.Context(), "[UnscheduleJob] trace_id="+traceID+" - Unscheduling job")

	var jobData types.ScheduleConditionJobData
	if err := c.ShouldBindJSON(&jobData); err != nil {
		h.logger.Error(c.Request.Context(), "Invalid request payload", observability.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"status":    "error",
			"message":   "Invalid request payload",
			"error":     err.Error(),
			"timestamp": time.Now().UTC(),
		})
		return
	}

	// Extract jobID from the request body
	if jobData.JobID == nil {
		err := fmt.Errorf("job ID is required")
		h.logger.Error(c.Request.Context(), "Missing job ID in request", observability.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"status":    "error",
			"message":   "Job ID is required",
			"error":     err.Error(),
			"timestamp": time.Now().UTC(),
		})
		return
	}

	jobID := jobData.JobID.ToBigInt()

	// Unschedule the job
	if err := h.scheduler.UnscheduleJob(c.Request.Context(), jobID); err != nil {
		h.logger.Error(c.Request.Context(), "Failed to unschedule condition job", observability.String("job_id", jobID.String()), observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":    "error",
			"message":   "Failed to unschedule condition job",
			"error":     err.Error(),
			"timestamp": time.Now().UTC(),
		})
		return
	}

	h.logger.Info(c.Request.Context(), "Condition job unscheduled successfully", observability.String("job_id", jobID.String()))

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
	h.logger.Info(c.Request.Context(), "[GetJobStats] trace_id="+traceID+" - Getting job stats")

	jobIDStr := c.Param("job_id")
	jobID := new(big.Int)
	_, ok := jobID.SetString(jobIDStr, 10)
	if !ok {
		err := fmt.Errorf("invalid job ID: %s", jobIDStr)
		h.logger.Error(c.Request.Context(), "Invalid job ID", observability.String("job_id", jobIDStr), observability.Error(err))
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
		h.logger.Error(c.Request.Context(), "Failed to get condition job stats", observability.String("job_id", jobID.String()))
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
	h.logger.Info(c.Request.Context(), "[GetStats] trace_id="+traceID+" - Getting scheduler stats")

	stats := h.scheduler.GetStats()

	response := gin.H{
		"status":    "success",
		"data":      stats,
		"timestamp": time.Now().UTC(),
	}

	c.JSON(http.StatusOK, response)
}
