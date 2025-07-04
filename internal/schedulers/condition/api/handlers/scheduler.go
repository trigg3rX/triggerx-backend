package handlers

import (
	"net/http"
	"strconv"
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

// ScheduleJob schedules a new condition-based job
func (h *SchedulerHandler) ScheduleJob(c *gin.Context) {
	var req types.ScheduleConditionJobData
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request payload", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":    "error",
			"message":   "Invalid request payload",
			"error":     err.Error(),
			"timestamp": time.Now().UTC(),
		})
		return
	}

	// Convert request to ConditionJobData
	jobData := &types.ScheduleConditionJobData{
		JobID:                     req.JobID,
		TaskDefinitionID:          req.TaskDefinitionID,
		LastExecutedAt:            req.LastExecutedAt,
		ExpirationTime:            req.ExpirationTime,
		Recurring:                 req.Recurring,
		ConditionType:             req.ConditionType,
		UpperLimit:                req.UpperLimit,
		LowerLimit:                req.LowerLimit,
		ValueSourceType:           req.ValueSourceType,
		ValueSourceUrl:            req.ValueSourceUrl,
		TaskTargetData:            req.TaskTargetData,
	}

	// Schedule the job
	if err := h.scheduler.ScheduleJob(jobData); err != nil {
		h.logger.Error("Failed to schedule condition job", "job_id", req.JobID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":    "error",
			"message":   "Failed to schedule condition job",
			"error":     err.Error(),
			"timestamp": time.Now().UTC(),
		})
		return
	}

	h.logger.Info("Condition job scheduled successfully", "job_id", req.JobID)

	response := gin.H{
		"status":    "success",
		"message":   "Condition job scheduled successfully",
		"job_id":    req.JobID,
		"timestamp": time.Now().UTC(),
	}

	c.JSON(http.StatusOK, response)
}

// UnscheduleJob unschedules a condition-based job
func (h *SchedulerHandler) UnscheduleJob(c *gin.Context) {
	jobIDStr := c.Param("job_id")
	jobID, err := strconv.ParseInt(jobIDStr, 10, 64)
	if err != nil {
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
	jobIDStr := c.Param("job_id")
	jobID, err := strconv.ParseInt(jobIDStr, 10, 64)
	if err != nil {
		h.logger.Error("Invalid job ID", "job_id", jobIDStr, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":    "error",
			"message":   "Invalid job ID",
			"error":     err.Error(),
			"timestamp": time.Now().UTC(),
		})
		return
	}

	stats, err := h.scheduler.GetJobWorkerStats(jobID)
	if err != nil {
		h.logger.Error("Failed to get condition job stats", "job_id", jobID, "error", err)
		c.JSON(http.StatusNotFound, gin.H{
			"status":    "error",
			"message":   "Condition job not found",
			"error":     err.Error(),
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
	stats := h.scheduler.GetStats()

	response := gin.H{
		"status":    "success",
		"data":      stats,
		"timestamp": time.Now().UTC(),
	}

	c.JSON(http.StatusOK, response)
}

// Stop stops the scheduler
func (h *SchedulerHandler) Stop(c *gin.Context) {
	h.logger.Info("Received request to stop condition scheduler")

	// Stop the scheduler
	h.scheduler.Stop()

	response := gin.H{
		"status":    "success",
		"message":   "Condition scheduler stopped successfully",
		"timestamp": time.Now().UTC(),
	}

	c.JSON(http.StatusOK, response)
}

// Start starts the scheduler (placeholder for future implementation)
func (h *SchedulerHandler) Start(c *gin.Context) {
	h.logger.Info("Received request to start condition scheduler")

	// Note: Starting a stopped scheduler would require additional implementation
	// For now, we'll return a message indicating the current state

	response := gin.H{
		"status":    "info",
		"message":   "Condition scheduler start functionality not implemented - scheduler runs automatically on service start",
		"timestamp": time.Now().UTC(),
	}

	c.JSON(http.StatusOK, response)
}
