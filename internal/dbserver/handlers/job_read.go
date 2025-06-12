package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
)

func (h *Handler) GetJobsByUserAddress(c *gin.Context) {
	userAddress := strings.ToLower(c.Param("user_address"))
	if userAddress == "" {
		h.logger.Error("[GetJobsByUserAddress] Invalid user address")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user address",
			"code":  "INVALID_ADDRESS",
		})
		return
	}

	h.logger.Infof("[GetJobsByUserAddress] Retrieving jobs for user address: %s", userAddress)

	// First get user ID and job IDs
	trackDBOp := metrics.TrackDBOperation("read", "user_data")
	userID, jobIDs, err := h.userRepository.GetUserJobIDsByAddress(userAddress)
	trackDBOp(err)
	if err != nil {
		h.logger.Errorf("[GetJobsByUserAddress] Error getting user data for address %s: %v", userAddress, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve user data",
			"code":  "USER_DATA_ERROR",
		})
		return
	}

	if len(jobIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "No jobs found for this user",
			"jobs":    []types.JobResponse{},
		})
		return
	}

	var jobs []types.JobResponse
	var hasErrors bool

	for _, jobID := range jobIDs {
		// Get basic job data
		trackDBOp = metrics.TrackDBOperation("read", "job_data")
		jobData, err := h.jobRepository.GetJobByID(jobID)
		trackDBOp(err)
		if err != nil {
			h.logger.Errorf("[GetJobsByUserAddress] Error getting job data for jobID %d: %v", jobID, err)
			hasErrors = true
			continue
		}

		jobResponse := types.JobResponse{JobData: *jobData}

		// Check task_definition_id to determine job type
		switch jobData.TaskDefinitionID {
		case 1, 2:
			// Time-based job
			trackDBOp = metrics.TrackDBOperation("read", "time_job")
			timeJobData, err := h.timeJobRepository.GetTimeJobByJobID(jobID)
			trackDBOp(err)
			if err != nil {
				h.logger.Errorf("[GetJobsByUserAddress] Error getting time job data for jobID %d: %v", jobID, err)
				hasErrors = true
				continue
			}
			jobResponse.TimeJobData = &timeJobData

		case 3, 4:
			// Event-based job
			trackDBOp = metrics.TrackDBOperation("read", "event_job")
			eventJobData, err := h.eventJobRepository.GetEventJobByJobID(jobID)
			trackDBOp(err)
			if err != nil {
				h.logger.Errorf("[GetJobsByUserAddress] Error getting event job data for jobID %d: %v", jobID, err)
				hasErrors = true
				continue
			}
			jobResponse.EventJobData = &eventJobData

		case 5, 6:
			// Condition-based job
			trackDBOp = metrics.TrackDBOperation("read", "condition_job")
			conditionJobData, err := h.conditionJobRepository.GetConditionJobByJobID(jobID)
			trackDBOp(err)
			if err != nil {
				h.logger.Errorf("[GetJobsByUserAddress] Error getting condition job data for jobID %d: %v", jobID, err)
				hasErrors = true
				continue
			}
			jobResponse.ConditionJobData = &conditionJobData

		default:
			h.logger.Errorf("[GetJobsByUserAddress] Unknown task definition ID %d for jobID %d", jobData.TaskDefinitionID, jobID)
			hasErrors = true
			continue
		}

		jobs = append(jobs, jobResponse)
	}

	h.logger.Infof("[GetJobsByUserAddress] Found %d jobs for user ID %d", len(jobs), userID)

	// If we have both jobs and errors, return a partial success response
	if len(jobs) > 0 && hasErrors {
		c.JSON(http.StatusPartialContent, gin.H{
			"message": "Some jobs were retrieved successfully, but there were errors with others",
			"jobs":    jobs,
		})
		return
	}

	// If we have only errors and no jobs, return an error response
	if len(jobs) == 0 && hasErrors {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve jobs",
			"code":  "JOB_RETRIEVAL_ERROR",
		})
		return
	}

	// If we have only jobs and no errors, return success
	c.JSON(http.StatusOK, gin.H{
		"jobs": jobs,
	})
}
