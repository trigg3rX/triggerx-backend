package handlers

import (
	"math/big"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
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

	jobIDBig := new(big.Int)
	_, ok := jobIDBig.SetString(jobID, 10)
	if !ok {
		h.logger.Error(c.Request.Context(), "[DeleteJobData] Invalid job ID format", observability.String("job_id", jobID))
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
		h.logger.Error(c.Request.Context(), "[DeleteJobData] Error getting job data for jobID", observability.String("job_id", jobID), observability.Error(err))
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
		h.logger.Error(c.Request.Context(), "[DeleteJobData] Error updating job status for jobID", observability.String("job_id", jobID), observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating job status: " + err.Error()})
		return
	}

	switch taskDefinitionID {
	case 1, 2:
		trackDBOp = metrics.TrackDBOperation("update", "time_job")
		err = h.timeJobRepository.UpdateTimeJobStatus(jobIDBig, false)
		trackDBOp(err)
		if err != nil {
			h.logger.Error(c.Request.Context(), "[DeleteJobData] Error updating time job status for jobID", observability.String("job_id", jobID), observability.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating time job status: " + err.Error()})
			return
		}
	case 3, 4:
		trackDBOp = metrics.TrackDBOperation("update", "event_job")
		err = h.eventJobRepository.UpdateEventJobStatus(jobIDBig, false)
		trackDBOp(err)
		if err != nil {
			h.logger.Error(c.Request.Context(), "[DeleteJobData] Error updating event job status for jobID", observability.String("job_id", jobID), observability.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating event job status: " + err.Error()})
			return
		}

		_, err = h.notifyPauseToConditionScheduler(c.Request.Context(), jobIDBig)
		if err != nil {
			h.logger.Error(c.Request.Context(), "[DeleteJobData] Error sending pause to event scheduler for jobID", observability.String("job_id", jobID), observability.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error sending pause to event scheduler: " + err.Error()})
			return
		}
	case 5, 6:
		trackDBOp = metrics.TrackDBOperation("update", "condition_job")
		err = h.conditionJobRepository.UpdateConditionJobStatus(jobIDBig, false)
		trackDBOp(err)
		if err != nil {
			h.logger.Error(c.Request.Context(), "[DeleteJobData] Error updating condition job status for jobID", observability.String("job_id", jobID), observability.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating condition job status: " + err.Error()})
			return
		}

		_, err = h.notifyPauseToConditionScheduler(c.Request.Context(), jobIDBig)
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

	jobID := new(big.Int)
	if _, ok := jobID.SetString(updateData.JobID, 10); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job_id format"})
		return
	}

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
	if err == nil && (job.TaskDefinitionID == 1 || job.TaskDefinitionID == 2) {
		// For time-based jobs, update time_interval and next_execution_timestamp
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
