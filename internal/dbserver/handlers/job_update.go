package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend-imua/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend-imua/internal/dbserver/types"
)

func (h *Handler) DeleteJobData(c *gin.Context) {
	jobID := c.Param("id")
	if jobID == "" {
		h.logger.Error("[DeleteJobData] No job ID provided")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No job ID provided",
			"code":  "MISSING_JOB_ID",
		})
		return
	}

	jobIDInt, err := strconv.ParseInt(jobID, 10, 64)
	if err != nil {
		h.logger.Errorf("[DeleteJobData] Invalid job ID format: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid job ID format",
			"code":  "INVALID_JOB_ID",
		})
		return
	}

	// Track database operation
	trackDBOp := metrics.TrackDBOperation("read", "job_data")
	taskDefinitionID, err := h.jobRepository.GetTaskDefinitionIDByJobID(jobIDInt)
	trackDBOp(err)
	if err != nil {
		h.logger.Errorf("[DeleteJobData] Error getting job data for jobID %s: %v", jobID, err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Job not found",
			"code":  "JOB_NOT_FOUND",
		})
		return
	}

	// Track job status update
	trackDBOp = metrics.TrackDBOperation("update", "job_data")
	err = h.jobRepository.UpdateJobStatus(jobIDInt, "deleted")
	trackDBOp(err)
	if err != nil {
		h.logger.Errorf("[DeleteJobData] Error updating job status for jobID %s: %v", jobID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating job status: " + err.Error()})
		return
	}

	switch taskDefinitionID {
	case 1, 2:
		trackDBOp = metrics.TrackDBOperation("update", "time_job")
		err = h.timeJobRepository.UpdateTimeJobStatus(jobIDInt, false)
		trackDBOp(err)
		if err != nil {
			h.logger.Errorf("[DeleteJobData] Error updating time job status for jobID %s: %v", jobID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating time job status: " + err.Error()})
			return
		}
	case 3, 4:
		trackDBOp = metrics.TrackDBOperation("update", "event_job")
		err = h.eventJobRepository.UpdateEventJobStatus(jobIDInt, false)
		trackDBOp(err)
		if err != nil {
			h.logger.Errorf("[DeleteJobData] Error updating event job status for jobID %s: %v", jobID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating event job status: " + err.Error()})
			return
		}

		_, err = h.notifyPauseToConditionScheduler(jobIDInt)
		if err != nil {
			h.logger.Errorf("[DeleteJobData] Error sending pause to event scheduler for jobID %s: %v", jobID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error sending pause to event scheduler: " + err.Error()})
			return
		}
	case 5, 6:
		trackDBOp = metrics.TrackDBOperation("update", "condition_job")
		err = h.conditionJobRepository.UpdateConditionJobStatus(jobIDInt, false)
		trackDBOp(err)
		if err != nil {
			h.logger.Errorf("[DeleteJobData] Error updating condition job status for jobID %s: %v", jobID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating condition job status: " + err.Error()})
			return
		}

		_, err = h.notifyPauseToConditionScheduler(jobIDInt)
		if err != nil {
			h.logger.Errorf("[DeleteJobData] Error sending pause to condition scheduler for jobID %s: %v", jobID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error sending pause to condition scheduler: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Job deleted successfully"})
}

func (h *Handler) UpdateJobDataFromUser(c *gin.Context) {
	var updateData types.UpdateJobDataFromUserRequest
	if err := c.ShouldBindJSON(&updateData); err != nil {
		h.logger.Errorf("[UpdateJobData] Error decoding request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	trackDBOp := metrics.TrackDBOperation("update", "job_data")
	err := h.jobRepository.UpdateJobFromUserInDB(&updateData)
	trackDBOp(err)
	if err != nil {
		h.logger.Errorf("[UpdateJobData] Error updating job data for jobID %d: %v", updateData.JobID, err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Job not found or update failed",
			"code":  "JOB_UPDATE_ERROR",
		})
		return
	}

	// Fetch the job to get its task_definition_id
	job, err := h.jobRepository.GetJobByID(updateData.JobID)
	if err == nil && (job.TaskDefinitionID == 1 || job.TaskDefinitionID == 2) {
		// For time-based jobs, update time_interval and next_execution_timestamp
		err = h.timeJobRepository.UpdateTimeJobInterval(updateData.JobID, updateData.TimeInterval)
		if err != nil {
			h.logger.Errorf("[UpdateJobData] Error updating time_interval for jobID %d: %v", updateData.JobID, err)
		}
		updatedAt := job.UpdatedAt
		nextExecution := updatedAt.Add(time.Duration(updateData.TimeInterval) * time.Second)
		err = h.timeJobRepository.UpdateTimeJobNextExecutionTimestamp(updateData.JobID, nextExecution)
		if err != nil {
			h.logger.Errorf("[UpdateJobData] Error updating next_execution_timestamp for jobID %d: %v", updateData.JobID, err)
			// Not returning error to client, just logging
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Job updated successfully",
		"job_id":     updateData.JobID,
		"updated_at": time.Now().UTC(),
	})
}

func (h *Handler) UpdateJobStatus(c *gin.Context) {
	jobID := c.Param("job_id")
	status := c.Param("status")

	// Validate status
	validStatuses := map[string]bool{
		"pending":  true,
		"in-queue": true,
		"running":  true,
	}

	if !validStatuses[status] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid status. Must be one of: pending, in-queue, running",
			"code":  "INVALID_STATUS",
		})
		return
	}

	// Convert jobID string to int64
	jobIDInt, err := strconv.ParseInt(jobID, 10, 64)
	if err != nil {
		h.logger.Errorf("[UpdateJobStatus] Error converting job ID to int64: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid job ID format",
			"code":  "INVALID_JOB_ID",
		})
		return
	}

	// Update the job status
	trackDBOp := metrics.TrackDBOperation("update", "job_data")
	if err := h.jobRepository.UpdateJobStatus(jobIDInt, status); err != nil {
		trackDBOp(err)
		h.logger.Errorf("[UpdateJobStatus] Error updating job status: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	trackDBOp(nil)

	c.JSON(http.StatusOK, gin.H{
		"message":    "Job status updated successfully",
		"job_id":     jobID,
		"status":     status,
		"updated_at": time.Now().UTC(),
	})
}

func (h *Handler) UpdateJobLastExecutedAt(c *gin.Context) {
	var updateData types.UpdateJobLastExecutedAtRequest
	if err := c.ShouldBindJSON(&updateData); err != nil {
		h.logger.Errorf("[UpdateJobLastExecutedAt] Error decoding request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update main job_data table
	trackDBOp := metrics.TrackDBOperation("update", "job_data")
	if err := h.jobRepository.UpdateJobLastExecutedAt(updateData.JobID, updateData.TaskIDs, updateData.JobCostActual, updateData.LastExecutedAt); err != nil {
		trackDBOp(err)
		h.logger.Errorf("[UpdateJobLastExecutedAt] Error updating job data for jobID %d: %v", updateData.JobID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating job data: " + err.Error()})
		return
	}
	trackDBOp(nil)

	c.JSON(http.StatusOK, gin.H{
		"message":          "Last executed time updated successfully",
		"job_id":           updateData.JobID,
		"last_executed_at": updateData.LastExecutedAt,
		"updated_at":       time.Now().UTC(),
	})
}
