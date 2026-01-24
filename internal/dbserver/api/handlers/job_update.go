package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (h *Handler) DeleteJobData(c *gin.Context) {
	jobID := c.Param("id")
	if jobID == "" {
		h.logger.Error(c.Request.Context(), "[DeleteJobData] No job ID provided")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No job ID provided",
			"code":  "MISSING_JOB_ID",
		})
		return
	}

	// Track database operation
	trackDBOp := metrics.TrackDBOperation("read", "job_data")
	taskDefinitionID, err := h.jobRepository.GetTaskDefinitionIDByJobID(jobID)
	trackDBOp(err)
	if err != nil {
		h.logger.Error(c.Request.Context(), "[DeleteJobData] Error getting job data for jobID", observability.String("job_id", jobID), observability.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Job not found",
			"code":  "JOB_NOT_FOUND",
		})
		return
	}

	// Track job status update
	trackDBOp = metrics.TrackDBOperation("update", "job_data")
	err = h.jobRepository.UpdateJobStatus(jobID, "deleted")
	trackDBOp(err)
	if err != nil {
		h.logger.Error(c.Request.Context(), "[DeleteJobData] Error updating job status for jobID", observability.String("job_id", jobID), observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating job status: " + err.Error()})
		return
	}

	switch taskDefinitionID {
	case 1, 2, 7:
		// Time-based job (TDI 1, 2) or Agent job (TDI 7)
		trackDBOp = metrics.TrackDBOperation("update", "time_job")
		err = h.timeJobRepository.UpdateTimeJobStatus(jobID, false)
		trackDBOp(err)
		if err != nil {
			h.logger.Error(c.Request.Context(), "[DeleteJobData] Error updating time job status for jobID", observability.String("job_id", jobID), observability.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating time job status: " + err.Error()})
			return
		}
	case 3, 4, 8:
		// Event-based job (TDI 3, 4) or Agent job (TDI 8)
		trackDBOp = metrics.TrackDBOperation("update", "event_job")
		err = h.eventJobRepository.UpdateEventJobStatus(jobID, false)
		trackDBOp(err)
		if err != nil {
			h.logger.Error(c.Request.Context(), "[DeleteJobData] Error updating event job status for jobID", observability.String("job_id", jobID), observability.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating event job status: " + err.Error()})
			return
		}

		_, err = h.notifyPauseToConditionScheduler(c.Request.Context(), jobID)
		if err != nil {
			h.logger.Error(c.Request.Context(), "[DeleteJobData] Error sending pause to event scheduler for jobID", observability.String("job_id", jobID), observability.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error sending pause to event scheduler: " + err.Error()})
			return
		}
	case 5, 6, 9:
		// Condition-based job (TDI 5, 6) or Agent job (TDI 9)
		trackDBOp = metrics.TrackDBOperation("update", "condition_job")
		err = h.conditionJobRepository.UpdateConditionJobStatus(jobID, false)
		trackDBOp(err)
		if err != nil {
			h.logger.Error(c.Request.Context(), "[DeleteJobData] Error updating condition job status for jobID", observability.String("job_id", jobID), observability.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating condition job status: " + err.Error()})
			return
		}

		_, err = h.notifyPauseToConditionScheduler(c.Request.Context(), jobID)
		if err != nil {
			h.logger.Error(c.Request.Context(), "[DeleteJobData] Error sending pause to condition scheduler for jobID", observability.String("job_id", jobID), observability.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error sending pause to condition scheduler: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Job deleted successfully"})
	h.logger.Info(c.Request.Context(), "[DeleteJobData] Deleted job", observability.String("job_id", jobID))
}

func (h *Handler) UpdateJobDataFromUser(c *gin.Context) {
	var updateData types.UpdateJobDataFromUserRequest
	if err := c.ShouldBindJSON(&updateData); err != nil {
		h.logger.Error(c.Request.Context(), "[UpdateJobData] Error decoding request body", observability.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	jobID := updateData.JobID

	trackDBOp := metrics.TrackDBOperation("update", "job_data")
	err := h.jobRepository.UpdateJobFromUserInDB(jobID, &updateData)
	trackDBOp(err)
	if err != nil {
		h.logger.Error(c.Request.Context(), "[UpdateJobData] Error updating job data for jobID", observability.String("job_id", updateData.JobID), observability.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Job not found or update failed",
			"code":  "JOB_UPDATE_ERROR",
		})
		return
	}

	// Fetch the job to get its task_definition_id
	job, err := h.jobRepository.GetJobByID(jobID)
	if err == nil && (job.TaskDefinitionID == 1 || job.TaskDefinitionID == 2 || job.TaskDefinitionID == 7) {
		// For time-based jobs (including agent jobs TDI 7), update time_interval and next_execution_timestamp
		err = h.timeJobRepository.UpdateTimeJobInterval(jobID, updateData.TimeInterval)
		if err != nil {
			h.logger.Error(c.Request.Context(), "[UpdateJobData] Error updating time_interval for jobID", observability.String("job_id", updateData.JobID), observability.Error(err))
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Job updated successfully",
		"job_id":     updateData.JobID,
		"updated_at": time.Now().UTC(),
	})
	h.logger.Info(c.Request.Context(), "[UpdateJobDataFromUser] Updated job", observability.String("job_id", updateData.JobID))
}
