package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// GetJobDataByJobIDForUser handles GET /jobs/user/:user_address/:job_id
func (h *Handler) GetJobDataByJobIDForUser(c *gin.Context) {
	userAddress := strings.ToLower(c.Param("user_address"))
	if userAddress == "" {
		h.logger.Error(c.Request.Context(), "[GetJobDataByJobIDForUser] user_address param missing")
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_address param missing"})
		return
	}

	jobIDParam := c.Param("job_id")
	if jobIDParam == "" {
		h.logger.Error(c.Request.Context(), "[GetJobDataByJobIDForUser] job_id param missing")
		c.JSON(http.StatusBadRequest, gin.H{"error": "job_id param missing"})
		return
	}

	jobID := jobIDParam

	// Fetch the job and check ownership
	jobData, err := h.jobRepository.GetJobByID(jobID)
	if err != nil {
		h.logger.Error(c.Request.Context(), "[GetJobDataByJobIDForUser] failed to get job data", observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get job data"})
		return
	}

	if strings.ToLower(jobData.UserAddress) != userAddress {
		h.logger.Warn(c.Request.Context(), "[GetJobDataByJobIDForUser] Access denied", observability.String("user_address", userAddress), observability.String("job_id", jobIDParam))
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden: job does not belong to user"})
		return
	}

	jobResponse := types.JobResponse{JobData: *jobData}

	// Check task_definition_id to determine job type
	switch jobData.TaskDefinitionID {
	case 1, 2, 7:
		// Time-based job (TDI 1, 2) or Agent job (TDI 7)
		trackDBOp := metrics.TrackDBOperation("read", "time_job")
		timeJobData, err := h.timeJobRepository.GetTimeJobByJobID(jobID)
		trackDBOp(err)
		if err != nil {
			h.logger.Warn(c.Request.Context(), "[GetJobDataByJobIDForUser] Failed to get time job data", observability.String("job_id", jobIDParam), observability.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get time job data"})
			return
		}
		jobResponse.TimeJobData = timeJobData

	case 3, 4, 8:
		// Event-based job (TDI 3, 4) or Agent job (TDI 8)
		trackDBOp := metrics.TrackDBOperation("read", "event_job")
		eventJobData, err := h.eventJobRepository.GetEventJobByJobID(jobID)
		trackDBOp(err)
		if err != nil {
			h.logger.Warn(c.Request.Context(), "[GetJobDataByJobIDForUser] Failed to get event job data", observability.String("job_id", jobIDParam), observability.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get event job data"})
			return
		}
		jobResponse.EventJobData = eventJobData

	case 5, 6, 9:
		// Condition-based job (TDI 5, 6) or Agent job (TDI 9)
		trackDBOp := metrics.TrackDBOperation("read", "condition_job")
		conditionJobData, err := h.conditionJobRepository.GetConditionJobByJobID(jobID)
		trackDBOp(err)
		if err != nil {
			h.logger.Warn(c.Request.Context(), "[GetJobDataByJobIDForUser] Failed to get condition job data", observability.String("job_id", jobIDParam), observability.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get condition job data"})
			return
		}
		jobResponse.ConditionJobData = conditionJobData

	default:
		h.logger.Warn(c.Request.Context(), "[GetJobDataByJobIDForUser] Unknown task definition ID", observability.Int("task_definition_id", jobData.TaskDefinitionID), observability.String("job_id", jobIDParam))
	}

	c.JSON(http.StatusOK, types.ConvertJobResponseToAPI(jobResponse))
	h.logger.Debug(c.Request.Context(), "[GetJobDataByJobIDForUser] Retrieved job data", observability.String("user_address", userAddress), observability.String("job_id", jobIDParam))
}

func (h *Handler) GetJobsByUserAddress(c *gin.Context) {
	userAddress := strings.ToLower(c.Param("user_address"))
	if userAddress == "" {
		h.logger.Warn(c.Request.Context(), "[GetJobsByUserAddress] Invalid user address")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user address",
			"code":  "INVALID_ADDRESS",
		})
		return
	}

	// Get user job IDs
	trackDBOp := metrics.TrackDBOperation("read", "user_data")
	jobIDs, err := h.userRepository.GetUserJobIDsByAddress(userAddress)
	trackDBOp(err)
	if err != nil {
		h.logger.Warn(c.Request.Context(), "[GetJobsByUserAddress] Failed to get user data", observability.String("user_address", userAddress), observability.Error(err))
		c.JSON(http.StatusOK, gin.H{
			"message": "No jobs found for this user",
			"jobs":    []types.JobResponseAPI{},
		})
		return
	}

	if len(jobIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "No jobs found for this user",
			"jobs":    []types.JobResponseAPI{},
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
			h.logger.Warn(c.Request.Context(), "[GetJobsByUserAddress] Failed to get job data", observability.String("job_id", jobID), observability.Error(err))
			hasErrors = true
			continue
		}

		jobResponse := types.JobResponse{JobData: *jobData}

		// Check task_definition_id to determine job type
		switch jobData.TaskDefinitionID {
		case 1, 2, 7:
			// Time-based job (TDI 1, 2) or Agent job (TDI 7)
			trackDBOp = metrics.TrackDBOperation("read", "time_job")
			timeJobData, err := h.timeJobRepository.GetTimeJobByJobID(jobID)
			trackDBOp(err)
			if err != nil {
				h.logger.Warn(c.Request.Context(), "[GetJobsByUserAddress] Failed to get time job data", observability.String("job_id", jobID), observability.Error(err))
				hasErrors = true
				continue
			}
			jobResponse.TimeJobData = timeJobData

		case 3, 4, 8:
			// Event-based job (TDI 3, 4) or Agent job (TDI 8)
			trackDBOp = metrics.TrackDBOperation("read", "event_job")
			eventJobData, err := h.eventJobRepository.GetEventJobByJobID(jobID)
			trackDBOp(err)
			if err != nil {
				h.logger.Warn(c.Request.Context(), "[GetJobsByUserAddress] Failed to get event job data", observability.String("job_id", jobID), observability.Error(err))
				hasErrors = true
				continue
			}
			jobResponse.EventJobData = eventJobData

		case 5, 6, 9:
			// Condition-based job (TDI 5, 6) or Agent job (TDI 9)
			trackDBOp = metrics.TrackDBOperation("read", "condition_job")
			conditionJobData, err := h.conditionJobRepository.GetConditionJobByJobID(jobID)
			trackDBOp(err)
			if err != nil {
				h.logger.Warn(c.Request.Context(), "[GetJobsByUserAddress] Failed to get condition job data", observability.String("job_id", jobID), observability.Error(err))
				hasErrors = true
				continue
			}
			jobResponse.ConditionJobData = conditionJobData

		default:
			h.logger.Warn(c.Request.Context(), "[GetJobsByUserAddress] Unknown task definition ID", observability.Int("task_definition_id", jobData.TaskDefinitionID), observability.String("job_id", jobID))
			hasErrors = true
			continue
		}

		jobs = append(jobs, jobResponse)
	}

	var jobsAPI []types.JobResponseAPI
	for _, job := range jobs {
		jobsAPI = append(jobsAPI, types.ConvertJobResponseToAPI(job))
	}

	// If we have both jobs and errors, return a partial success response
	if len(jobs) > 0 && hasErrors {
		c.JSON(http.StatusPartialContent, gin.H{
			"message": "Some jobs were retrieved successfully, but there were errors with others",
			"jobs":    jobsAPI,
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
		"jobs": jobsAPI,
	})
	h.logger.Debug(c.Request.Context(), "[GetJobsByUserAddress] Retrieved jobs", observability.String("user_address", userAddress), observability.Int("jobs_count", len(jobs)))
}

// GetJobsByApiKey handles GET /jobs/by-apikey
func (h *Handler) GetJobsByApiKey(c *gin.Context) {
	apiKey := c.GetHeader("X-Api-Key")
	if apiKey == "" {
		h.logger.Error(c.Request.Context(), "[GetJobsByApiKey] Missing X-Api-Key header")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing X-Api-Key header"})
		return
	}

	apiKeyData, err := h.apiKeysRepository.GetApiKeyDataByKey(apiKey)
	if err != nil {
		h.logger.Warn(c.Request.Context(), "[GetJobsByApiKey] Invalid API key", observability.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key", "code": "INVALID_API_KEY"})
		return
	}

	userAddress := apiKeyData.Owner
	if userAddress == "" {
		h.logger.Error(c.Request.Context(), "[GetJobsByApiKey] No owner found for API key")
		c.JSON(http.StatusNotFound, gin.H{"error": "No owner found for API key", "code": "NO_OWNER_FOUND"})
		return
	}

	// Reuse the logic from GetJobsByUserAddress
	c.Params = append(c.Params, gin.Param{Key: "user_address", Value: userAddress})
	h.GetJobsByUserAddress(c)
	h.logger.Debug(c.Request.Context(), "[GetJobsByApiKey] Retrieved jobs", observability.String("user_address", userAddress))
}

// GetJobDataByJobID handles GET /jobs/:job_id
func (h *Handler) GetJobDataByJobID(c *gin.Context) {
	jobIDParam := c.Param("job_id")
	if jobIDParam == "" {
		h.logger.Error(c.Request.Context(), "[GetJobDataByJobID] job_id param missing")
		c.JSON(http.StatusBadRequest, gin.H{"error": "job_id param missing"})
		return
	}

	jobID := jobIDParam

	jobData, err := h.jobRepository.GetJobByID(jobID)
	if err != nil {
		h.logger.Warn(c.Request.Context(), "[GetJobDataByJobID] Failed to get job data", observability.String("job_id", jobIDParam), observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get job data"})
		return
	}

	c.JSON(http.StatusOK, jobData)
	h.logger.Debug(c.Request.Context(), "[GetJobDataByJobID] Retrieved job data", observability.String("job_id", jobIDParam))
}

func (h *Handler) GetJobsByUserAddressAndChainID(c *gin.Context) {
	userAddress := strings.ToLower(c.Param("user_address"))
	createdChainIDParam := c.Param("created_chain_id")

	if userAddress == "" {
		h.logger.Warn(c.Request.Context(), "[GetJobsByUserAddressAndChainID] Invalid user address")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user address",
			"code":  "INVALID_ADDRESS",
		})
		return
	}
	if createdChainIDParam == "" {
		h.logger.Warn(c.Request.Context(), "[GetJobsByUserAddressAndChainID] Invalid created_chain_id")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid created_chain_id",
			"code":  "INVALID_CREATED_CHAIN_ID",
		})
		return
	}

	createdChainID := createdChainIDParam

	// Get user job IDs
	trackDBOp := metrics.TrackDBOperation("read", "user_data")
	jobIDs, err := h.userRepository.GetUserJobIDsByAddress(userAddress)
	trackDBOp(err)
	if err != nil {
		h.logger.Warn(c.Request.Context(), "[GetJobsByUserAddressAndChainID] Failed to get user data", observability.String("user_address", userAddress), observability.Error(err))
		c.JSON(http.StatusOK, gin.H{
			"message": "No jobs found for this user",
			"jobs":    []types.JobResponseAPI{},
		})
		return
	}

	if len(jobIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "No jobs found for this user",
			"jobs":    []types.JobResponseAPI{},
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
			h.logger.Warn(c.Request.Context(), "[GetJobsByUserAddressAndChainID] Failed to get job data", observability.String("job_id", jobID), observability.Error(err))
			hasErrors = true
			continue
		}

		// Filter by created_chain_id
		if jobData.CreatedChainID != createdChainID {
			continue
		}

		jobResponse := types.JobResponse{JobData: *jobData}

		// Check task_definition_id to determine job type
		switch jobData.TaskDefinitionID {
		case 1, 2, 7:
			// Time-based job (TDI 1, 2) or Agent job (TDI 7)
			trackDBOp = metrics.TrackDBOperation("read", "time_job")
			timeJobData, err := h.timeJobRepository.GetTimeJobByJobID(jobID)
			trackDBOp(err)
			if err != nil {
				h.logger.Warn(c.Request.Context(), "[GetJobsByUserAddressAndChainID] Failed to get time job data", observability.String("job_id", jobID), observability.Error(err))
				hasErrors = true
				continue
			}
			jobResponse.TimeJobData = timeJobData

		case 3, 4, 8:
			// Event-based job (TDI 3, 4) or Agent job (TDI 8)
			trackDBOp = metrics.TrackDBOperation("read", "event_job")
			eventJobData, err := h.eventJobRepository.GetEventJobByJobID(jobID)
			trackDBOp(err)
			if err != nil {
				h.logger.Warn(c.Request.Context(), "[GetJobsByUserAddressAndChainID] Failed to get event job data", observability.String("job_id", jobID), observability.Error(err))
				hasErrors = true
				continue
			}
			jobResponse.EventJobData = eventJobData

		case 5, 6, 9:
			// Condition-based job (TDI 5, 6) or Agent job (TDI 9)
			trackDBOp = metrics.TrackDBOperation("read", "condition_job")
			conditionJobData, err := h.conditionJobRepository.GetConditionJobByJobID(jobID)
			trackDBOp(err)
			if err != nil {
				h.logger.Warn(c.Request.Context(), "[GetJobsByUserAddressAndChainID] Failed to get condition job data", observability.String("job_id", jobID), observability.Error(err))
				hasErrors = true
				continue
			}
			jobResponse.ConditionJobData = conditionJobData

		default:
			h.logger.Warn(c.Request.Context(), "[GetJobsByUserAddressAndChainID] Unknown task definition ID", observability.Int("task_definition_id", jobData.TaskDefinitionID), observability.String("job_id", jobID))
			hasErrors = true
			continue
		}

		jobs = append(jobs, jobResponse)
	}

	var jobsAPI []types.JobResponseAPI
	for _, job := range jobs {
		jobsAPI = append(jobsAPI, types.ConvertJobResponseToAPI(job))
	}

	// If we have both jobs and errors, return a partial success response
	if len(jobs) > 0 && hasErrors {
		c.JSON(http.StatusPartialContent, gin.H{
			"message": "Some jobs were retrieved successfully, but there were errors with others",
			"jobs":    jobsAPI,
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
		"jobs": jobsAPI,
	})
	if len(jobs) > 0 {
		h.logger.Debug(c.Request.Context(), "[GetJobsByUserAddressAndChainID] Retrieved jobs", observability.String("user_address", userAddress), observability.String("created_chain_id", createdChainIDParam), observability.Int("jobs_count", len(jobs)))
	}
}
