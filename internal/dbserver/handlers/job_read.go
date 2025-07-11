package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend-imua/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend-imua/internal/dbserver/types"
)

func (h *Handler) GetJobsByUserAddress(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Infof("[GetJobsByUserAddress] trace_id=%s - Retrieving jobs by user address", traceID)
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
		if err.Error() == "user address not found" {
			h.logger.Infof("[GetJobsByUserAddress] No user found for address %s", userAddress)
		} else {
			h.logger.Errorf("[GetJobsByUserAddress] Error getting user data for address %s: %v", userAddress, err)
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "No jobs found for this user",
			"jobs":    []types.JobResponse{},
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

// GetTaskFeesByJobID handles GET /jobs/:job_id/task-fees
func (h *Handler) GetTaskFeesByJobID(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Infof("[GetTaskFeesByJobID] trace_id=%s - Getting task fees by job ID", traceID)

	jobIDParam := c.Param("job_id")
	if jobIDParam == "" {
		h.logger.Error("[GetTaskFeesByJobID] job_id param missing")
		c.JSON(http.StatusBadRequest, gin.H{"error": "job_id param missing"})
		return
	}

	jobID, err := parseInt64(jobIDParam)
	if err != nil {
		h.logger.Errorf("[GetTaskFeesByJobID] invalid job_id: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job_id"})
		return
	}

	taskFees, err := h.jobRepository.GetTaskFeesByJobID(jobID)
	if err != nil {
		h.logger.Errorf("[GetTaskFeesByJobID] failed to get task fees: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get task fees"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"task_fees": taskFees})
}

// GetJobsByApiKey handles GET /jobs/by-apikey
func (h *Handler) GetJobsByApiKey(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Infof("[GetJobsByApiKey] trace_id=%s - Retrieving jobs by API key", traceID)
	apiKey := c.GetHeader("X-Api-Key")
	if apiKey == "" {
		h.logger.Error("[GetJobsByApiKey] Missing X-Api-Key header")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing X-Api-Key header"})
		return
	}

	apiKeyData, err := h.apiKeysRepository.GetApiKeyDataByKey(apiKey)
	if err != nil {
		h.logger.Errorf("[GetJobsByApiKey] Invalid API key: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
		return
	}

	userAddress := apiKeyData.Owner
	if userAddress == "" {
		h.logger.Error("[GetJobsByApiKey] No owner found for API key")
		c.JSON(http.StatusNotFound, gin.H{"error": "No owner found for API key"})
		return
	}

	// Reuse the logic from GetJobsByUserAddress
	c.Params = append(c.Params, gin.Param{Key: "user_address", Value: userAddress})
	h.GetJobsByUserAddress(c)
}

// parseInt64 is a helper to parse int64 from string
func parseInt64(s string) (int64, error) {
	var i int64
	_, err := fmt.Sscan(s, &i)
	return i, err
}
