package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
)

func (h *Handler) DeleteJobData(c *gin.Context) {
	jobID := c.Param("id")
	if jobID == "" {
		h.logger.Error("[DeleteJobData] No job ID provided")
		c.JSON(http.StatusBadRequest, gin.H{"error": "No job ID provided"})
		return
	}

	jobIDInt, err := strconv.ParseInt(jobID, 10, 64)
	if err != nil {
		h.logger.Errorf("[DeleteJobData] Invalid job ID format: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID format"})
		return
	}

	taskDefinitionID, err := h.jobRepository.GetTaskDefinitionIDByJobID(jobIDInt)
	if err != nil {
		h.logger.Errorf("[DeleteJobData] Error getting job data for jobID %s: %v", jobID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting job data: " + err.Error()})
		return
	}

	err = h.jobRepository.UpdateJobStatus(jobIDInt, "deleted")
	if err != nil {
		h.logger.Errorf("[DeleteJobData] Error updating job status for jobID %s: %v", jobID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating job status: " + err.Error()})
		return
	}

	switch taskDefinitionID {
	case 1, 2:
		err = h.timeJobRepository.UpdateTimeJobStatus(jobIDInt, false)
		if err != nil {
			h.logger.Errorf("[DeleteJobData] Error updating time job status for jobID %s: %v", jobID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating time job status: " + err.Error()})
			return
		}
	case 3, 4:
		err = h.eventJobRepository.UpdateEventJobStatus(jobIDInt, false)
		if err != nil {
			h.logger.Errorf("[DeleteJobData] Error updating event job status for jobID %s: %v", jobID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating event job status: " + err.Error()})
			return
		}
		_, err = h.SendPauseToEventScheduler("/pause")
		if err != nil {
			h.logger.Errorf("[DeleteJobData] Error sending pause to event scheduler for jobID %s: %v", jobID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error sending pause to event scheduler: " + err.Error()})
			return
		}
	case 5, 6:
		err = h.conditionJobRepository.UpdateConditionJobStatus(jobIDInt, false)
		if err != nil {
			h.logger.Errorf("[DeleteJobData] Error updating condition job status for jobID %s: %v", jobID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating condition job status: " + err.Error()})
			return
		}
		_, err = h.SendPauseToConditionScheduler("/pause")
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.jobRepository.UpdateJobFromUserInDB(&updateData)
	if err != nil {
		h.logger.Errorf("[UpdateJobData] Error updating job data for jobID %d: %v", updateData.JobID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating job data: " + err.Error()})
		return
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status. Must be one of: pending, in-queue, running"})
		return
	}
	// Convert jobID string to int64
	jobIDInt, err := strconv.ParseInt(jobID, 10, 64)
	if err != nil {
		h.logger.Errorf("[UpdateJobStatus] Error converting job ID to int64: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	// Update the job status
	if err := h.jobRepository.UpdateJobStatus(jobIDInt, status); err != nil {
		h.logger.Errorf("[UpdateJobStatus] Error updating job status: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

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
	if err := h.jobRepository.UpdateJobLastExecutedAt(updateData.JobID, updateData.TaskIDs, updateData.JobCostActual, updateData.LastExecutedAt); err != nil {
		h.logger.Errorf("[UpdateJobLastExecutedAt] Error updating job data for jobID %d: %v", updateData.JobID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating job data: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":          "Last executed time updated successfully",
		"job_id":           updateData.JobID,
		"last_executed_at": updateData.LastExecutedAt,
		"updated_at":       time.Now().UTC(),
	})
}