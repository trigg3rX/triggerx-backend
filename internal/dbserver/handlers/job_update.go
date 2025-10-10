package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/errors"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (h *Handler) DeleteJobData(c *gin.Context) {
	logger := h.getLogger(c)
	jobID := c.Param("id")
	if types.IsValidJobID(jobID) {
		logger.Errorf("%s: %s", errors.ErrInvalidRequestBody, "Invalid job ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.ErrInvalidRequestBody})
		return
	}
	logger.Debugf("PUT [DeleteJobData] Job ID: %s", jobID)

	// Track database operation
	trackDBOp := metrics.TrackDBOperation("read", "job_data")
	job, err := h.jobRepository.GetByID(c.Request.Context(), jobID)
	trackDBOp(err)
	if err != nil || job == nil {
		logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
		c.JSON(http.StatusNotFound, gin.H{"error": errors.ErrDBRecordNotFound})
		return
	}

	// Track job status update
	job.Status = string(types.JobStatusDeleted)
	job.UpdatedAt = time.Now().UTC()

	trackDBOp = metrics.TrackDBOperation("update", "job_data")
	err = h.jobRepository.Update(c.Request.Context(), job)
	trackDBOp(err)
	if err != nil {
		logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating job status: " + err.Error()})
		return
	}

	switch job.TaskDefinitionID {
	case 1, 2:
		trackDBOp = metrics.TrackDBOperation("update", "time_job")
		timeJob, err := h.timeJobRepository.GetByNonID(c.Request.Context(), "job_id", jobID)
		if err == nil && timeJob != nil {
			timeJob.IsCompleted = true
			err = h.timeJobRepository.Update(c.Request.Context(), timeJob)
		}
		trackDBOp(err)
		if err != nil {
			logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating time job status: " + err.Error()})
			return
		}
	case 3, 4:
		trackDBOp = metrics.TrackDBOperation("update", "event_job")
		eventJob, err := h.eventJobRepository.GetByNonID(c.Request.Context(), "job_id", jobID)
		if err == nil && eventJob != nil {
			eventJob.IsCompleted = true
			err = h.eventJobRepository.Update(c.Request.Context(), eventJob)
		}
		trackDBOp(err)
		if err != nil {
			logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating event job status: " + err.Error()})
			return
		}

		_, err = h.notifyPauseToConditionScheduler(jobID)
		if err != nil {
			logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error sending pause to event scheduler: " + err.Error()})
			return
		}
	case 5, 6:
		trackDBOp = metrics.TrackDBOperation("update", "condition_job")
		conditionJob, err := h.conditionJobRepository.GetByNonID(c.Request.Context(), "job_id", jobID)
		if err == nil && conditionJob != nil {
			conditionJob.IsCompleted = true
			err = h.conditionJobRepository.Update(c.Request.Context(), conditionJob)
		}
		trackDBOp(err)
		if err != nil {
			logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating condition job status: " + err.Error()})
			return
		}

		_, err = h.notifyPauseToConditionScheduler(jobID)
		if err != nil {
			logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error sending pause to condition scheduler: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Job deleted successfully"})
}

func (h *Handler) UpdateJobDataFromUser(c *gin.Context) {
	logger := h.getLogger(c)
	logger.Debugf("PUT [UpdateJobDataFromUser] Updating job data from user")

	var updateData types.UpdateJobDataFromUserRequest
	if err := c.ShouldBindJSON(&updateData); err != nil {
		logger.Errorf("Error decoding request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	jobID := updateData.JobID

	// Get the job first
	trackDBOp := metrics.TrackDBOperation("read", "job_data")
	job, err := h.jobRepository.GetByID(c.Request.Context(), jobID)
	trackDBOp(err)
	if err != nil || job == nil {
		logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Job not found",
			"code":  "JOB_NOT_FOUND",
		})
		return
	}

	// Update job fields from request
	job.JobTitle = updateData.JobTitle
	job.UpdatedAt = time.Now().UTC()

	trackDBOp = metrics.TrackDBOperation("update", "job_data")
	err = h.jobRepository.Update(c.Request.Context(), job)
	trackDBOp(err)
	if err != nil {
		logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Job update failed",
			"code":  "JOB_UPDATE_ERROR",
		})
		return
	}

	// For time-based jobs, update time_interval and next_execution_timestamp
	if job.TaskDefinitionID == 1 || job.TaskDefinitionID == 2 {
		timeJob, err := h.timeJobRepository.GetByNonID(c.Request.Context(), "job_id", jobID)
		if err == nil && timeJob != nil {
			timeJob.TimeInterval = updateData.TimeInterval
			nextExecution := job.UpdatedAt.Add(time.Duration(updateData.TimeInterval) * time.Second)
			timeJob.NextExecutionTimestamp = nextExecution

			err = h.timeJobRepository.Update(c.Request.Context(), timeJob)
			if err != nil {
				logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
				// Not returning error to client, just logging
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Job updated successfully",
		"job_id":     updateData.JobID,
		"updated_at": time.Now().UTC(),
	})
}

func (h *Handler) UpdateJobLastExecutedAt(c *gin.Context) {
	logger := h.getLogger(c)
	logger.Debugf("PUT [UpdateJobLastExecutedAt] Updating job last executed at")

	var req types.UpdateJobLastExecutedAtRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Errorf("Error decoding request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get the job first
	trackDBOp := metrics.TrackDBOperation("read", "job_data")
	job, err := h.jobRepository.GetByID(c.Request.Context(), req.JobID)
	trackDBOp(err)
	if err != nil || job == nil {
		logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
		c.JSON(http.StatusNotFound, gin.H{"error": errors.ErrDBRecordNotFound})
		return
	}

	// Update job fields
	job.LastExecutedAt = req.LastExecutedAt
	job.TaskIDs = append(job.TaskIDs, req.TaskIDs)
	job.JobCostActual = types.Add(job.JobCostActual, req.JobCostActual)
	job.UpdatedAt = time.Now().UTC()

	// Update main job_data table
	trackDBOp = metrics.TrackDBOperation("update", "job_data")
	if err := h.jobRepository.Update(c.Request.Context(), job); err != nil {
		trackDBOp(err)
		logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating job data: " + err.Error()})
		return
	}
	trackDBOp(nil)

	c.JSON(http.StatusOK, gin.H{
		"message":          "Last executed time updated successfully",
		"job_id":           req.JobID,
		"last_executed_at": req.LastExecutedAt,
		"updated_at":       time.Now().UTC(),
	})
}
