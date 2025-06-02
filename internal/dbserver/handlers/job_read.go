package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
)

func (h *Handler) GetJobsByUserAddress(c *gin.Context) {
	userAddress := c.Param("user_address")
	if userAddress == "" {
		h.logger.Error("[GetJobsByUserAddress] No user address provided")
		c.JSON(http.StatusBadRequest, gin.H{"error": "No user address provided"})
		return
	}

	jobIDs, err := h.userRepository.GetUserJobIDsByAddress(strings.ToLower(userAddress))
	if err != nil {
		h.logger.Errorf("[GetJobsByUserAddress] Error getting user data for address %s: %v", userAddress, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting user data: " + err.Error()})
		return
	}

	if len(jobIDs) == 0 {
		c.JSON(http.StatusOK, []types.JobResponse{})
		return
	}

	var jobs []types.JobResponse
	for _, jobID := range jobIDs {
		// Get basic job data
		jobData, err := h.jobRepository.GetJobByID(jobID)
		if err != nil {
			h.logger.Errorf("[GetJobsByUserAddress] Error getting job data for jobID %d: %v", jobID, err)
			continue
		}

		jobResponse := types.JobResponse{JobData: *jobData}

		// Check task_definition_id to determine job type
		switch {
		case jobData.TaskDefinitionID == 1 || jobData.TaskDefinitionID == 2:
			// Time-based job
			timeJobData, err := h.timeJobRepository.GetTimeJobByJobID(jobID)
			if err != nil {
				h.logger.Errorf("[GetJobsByUserAddress] Error getting time job data for jobID %d: %v", jobID, err)
				continue
			}
			jobResponse.TimeJobData = &timeJobData

		case jobData.TaskDefinitionID == 3 || jobData.TaskDefinitionID == 4:
			// Event-based job
			eventJobData, err := h.eventJobRepository.GetEventJobByJobID(jobID)
			if err != nil {
				h.logger.Errorf("[GetJobsByUserAddress] Error getting event job data for jobID %d: %v", jobID, err)
				continue
			}
			jobResponse.EventJobData = &eventJobData

		case jobData.TaskDefinitionID == 5 || jobData.TaskDefinitionID == 6:
			// Condition-based job
			conditionJobData, err := h.conditionJobRepository.GetConditionJobByJobID(jobID)
			if err != nil {
				h.logger.Errorf("[GetJobsByUserAddress] Error getting condition job data for jobID %d: %v", jobID, err)
				continue
			}
			jobResponse.ConditionJobData = &conditionJobData

		default:
			// No specific job data if task_definition_id is not recognized
		}

		jobs = append(jobs, jobResponse)
	}

	c.JSON(http.StatusOK, jobs)
}