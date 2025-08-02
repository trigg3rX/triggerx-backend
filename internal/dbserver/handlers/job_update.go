package handlers

import (
	"math/big"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
)

func (h *Handler) DeleteJobData(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Infof("[DeleteJobData] trace_id=%s - Deleting job data", traceID)

	jobID := c.Param("id")
	if jobID == "" {
		h.logger.Error("[DeleteJobData] No job ID provided")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No job ID provided",
			"code":  "MISSING_JOB_ID",
		})
		return
	}

	jobIDBig := new(big.Int)
	_, ok := jobIDBig.SetString(jobID, 10)
	if !ok {
		h.logger.Errorf("[DeleteJobData] Invalid job ID format: %v", jobID)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid job ID format",
			"code":  "INVALID_JOB_ID",
		})
		return
	}

	// Track database operation
	trackDBOp := metrics.TrackDBOperation("read", "job_data")
	taskDefinitionID, err := h.jobRepository.GetTaskDefinitionIDByJobID(jobIDBig)
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
	err = h.jobRepository.UpdateJobStatus(jobIDBig, "deleted")
	trackDBOp(err)
	if err != nil {
		h.logger.Errorf("[DeleteJobData] Error updating job status for jobID %s: %v", jobID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating job status: " + err.Error()})
		return
	}

	switch taskDefinitionID {
	case 1, 2:
		trackDBOp = metrics.TrackDBOperation("update", "time_job")
		err = h.timeJobRepository.UpdateTimeJobStatus(jobIDBig, false)
		trackDBOp(err)
		if err != nil {
			h.logger.Errorf("[DeleteJobData] Error updating time job status for jobID %s: %v", jobID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating time job status: " + err.Error()})
			return
		}
	case 3, 4:
		trackDBOp = metrics.TrackDBOperation("update", "event_job")
		err = h.eventJobRepository.UpdateEventJobStatus(jobIDBig, false)
		trackDBOp(err)
		if err != nil {
			h.logger.Errorf("[DeleteJobData] Error updating event job status for jobID %s: %v", jobID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating event job status: " + err.Error()})
			return
		}

		_, err = h.notifyPauseToConditionScheduler(jobIDBig)
		if err != nil {
			h.logger.Errorf("[DeleteJobData] Error sending pause to event scheduler for jobID %s: %v", jobID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error sending pause to event scheduler: " + err.Error()})
			return
		}
	case 5, 6:
		trackDBOp = metrics.TrackDBOperation("update", "condition_job")
		err = h.conditionJobRepository.UpdateConditionJobStatus(jobIDBig, false)
		trackDBOp(err)
		if err != nil {
			h.logger.Errorf("[DeleteJobData] Error updating condition job status for jobID %s: %v", jobID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating condition job status: " + err.Error()})
			return
		}

		_, err = h.notifyPauseToConditionScheduler(jobIDBig)
		if err != nil {
			h.logger.Errorf("[DeleteJobData] Error sending pause to condition scheduler for jobID %s: %v", jobID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error sending pause to condition scheduler: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Job deleted successfully"})
}

func (h *Handler) UpdateJobDataFromUser(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Infof("[UpdateJobData] trace_id=%s - Updating job data from user", traceID)

	var updateData types.UpdateJobDataFromUserRequest
	if err := c.ShouldBindJSON(&updateData); err != nil {
		h.logger.Errorf("[UpdateJobData] Error decoding request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	jobID := new(big.Int)
	if _, ok := jobID.SetString(updateData.JobID, 10); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job_id format"})
		return
	}

	trackDBOp := metrics.TrackDBOperation("update", "job_data")
	err := h.jobRepository.UpdateJobFromUserInDB(jobID, &updateData)
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
	job, err := h.jobRepository.GetJobByID(jobID)
	if err == nil && (job.TaskDefinitionID == 1 || job.TaskDefinitionID == 2) {
		// For time-based jobs, update time_interval and next_execution_timestamp
		err = h.timeJobRepository.UpdateTimeJobInterval(jobID, updateData.TimeInterval)
		if err != nil {
			h.logger.Errorf("[UpdateJobData] Error updating time_interval for jobID %d: %v", updateData.JobID, err)
		}
		updatedAt := job.UpdatedAt
		nextExecution := updatedAt.Add(time.Duration(updateData.TimeInterval) * time.Second)
		err = h.timeJobRepository.UpdateTimeJobNextExecutionTimestamp(jobID, nextExecution)
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
	traceID := h.getTraceID(c)
	h.logger.Infof("[UpdateJobStatus] trace_id=%s - Updating job status", traceID)

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

	// Convert jobID string to *big.Int
	jobIDBig := new(big.Int)
	_, ok := jobIDBig.SetString(jobID, 10)
	if !ok {
		h.logger.Errorf("[UpdateJobStatus] Error converting job ID to *big.Int: %v", jobID)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid job ID format",
			"code":  "INVALID_JOB_ID",
		})
		return
	}

	// Update the job status
	trackDBOp := metrics.TrackDBOperation("update", "job_data")
	if err := h.jobRepository.UpdateJobStatus(jobIDBig, status); err != nil {
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
	traceID := h.getTraceID(c)
	h.logger.Infof("[UpdateJobLastExecutedAt] trace_id=%s - Updating job last executed at", traceID)

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
