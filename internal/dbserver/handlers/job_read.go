package handlers

import (
	// "fmt"
	"math/big"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

// GetJobDataByJobIDForUser handles GET /jobs/user/:user_address/:job_id
func (h *Handler) GetJobDataByJobIDForUser(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Info(c.Request.Context(), "[GetJobDataByJobIDForUser] Retrieving job data by job ID with user ownership check", observability.String("trace_id", traceID))

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

	jobID := new(big.Int)
	_, ok := jobID.SetString(jobIDParam, 10)
	if !ok {
		h.logger.Error(c.Request.Context(), "[GetJobDataByJobIDForUser] invalid job_id", observability.String("job_id", jobIDParam))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job_id"})
		return
	}

	// Resolve user address to user ID
	userID, err := h.userRepository.GetUserIDByAddress(userAddress)
	if err != nil {
		h.logger.Error(c.Request.Context(), "[GetJobDataByJobIDForUser] failed to resolve user by address", observability.String("user_address", userAddress), observability.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Fetch the job and check ownership
	jobData, err := h.jobRepository.GetJobByID(jobID)
	if err != nil {
		h.logger.Error(c.Request.Context(), "[GetJobDataByJobIDForUser] failed to get job data", observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get job data"})
		return
	}

	if jobData.UserID != userID {
		h.logger.Warn(c.Request.Context(), "[GetJobDataByJobIDForUser] Access denied: user does not have access to this job", observability.String("user_address", userAddress), observability.Int64("user_id", userID), observability.String("job_id", jobIDParam), observability.Int64("job_user_id", jobData.UserID))
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden: job does not belong to user"})
		return
	}

	jobResponse := types.JobResponse{JobData: *jobData}

	// Check task_definition_id to determine job type
	switch jobData.TaskDefinitionID {
	case 1, 2:
		// Time-based job
		trackDBOp := metrics.TrackDBOperation("read", "time_job")
		timeJobData, err := h.timeJobRepository.GetTimeJobByJobID(jobID)
		trackDBOp(err)
		if err != nil {
			h.logger.Error(c.Request.Context(), "[GetJobDataByJobIDForUser] Error getting time job data for jobID", observability.String("job_id", jobIDParam), observability.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get time job data"})
			return
		}
		jobResponse.TimeJobData = &timeJobData

	case 3, 4:
		// Event-based job
		trackDBOp := metrics.TrackDBOperation("read", "event_job")
		eventJobData, err := h.eventJobRepository.GetEventJobByJobID(jobID)
		trackDBOp(err)
		if err != nil {
			h.logger.Error(c.Request.Context(), "[GetJobDataByJobIDForUser] Error getting event job data for jobID", observability.String("job_id", jobIDParam), observability.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get event job data"})
			return
		}
		jobResponse.EventJobData = &eventJobData

	case 5, 6:
		// Condition-based job
		trackDBOp := metrics.TrackDBOperation("read", "condition_job")
		conditionJobData, err := h.conditionJobRepository.GetConditionJobByJobID(jobID)
		trackDBOp(err)
		if err != nil {
			h.logger.Error(c.Request.Context(), "[GetJobDataByJobIDForUser] Error getting condition job data for jobID", observability.String("job_id", jobIDParam), observability.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get condition job data"})
			return
		}
		jobResponse.ConditionJobData = &conditionJobData

	default:
		h.logger.Error(c.Request.Context(), "[GetJobDataByJobIDForUser] Unknown task definition ID for jobID", observability.Int("task_definition_id", jobData.TaskDefinitionID), observability.String("job_id", jobIDParam))
	}

	c.JSON(http.StatusOK, types.ConvertJobResponseToAPI(jobResponse))
}

func (h *Handler) GetJobsByUserAddress(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Info(c.Request.Context(), "[GetJobsByUserAddress] Retrieving jobs by user address", observability.String("trace_id", traceID))
	userAddress := strings.ToLower(c.Param("user_address"))
	if userAddress == "" {
		h.logger.Error(c.Request.Context(), "[GetJobsByUserAddress] Invalid user address")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user address",
			"code":  "INVALID_ADDRESS",
		})
		return
	}

	h.logger.Info(c.Request.Context(), "[GetJobsByUserAddress] Retrieving jobs for user address", observability.String("user_address", userAddress))

	// First get user ID and job IDs
	trackDBOp := metrics.TrackDBOperation("read", "user_data")
	userID, jobIDs, err := h.userRepository.GetUserJobIDsByAddress(userAddress)
	trackDBOp(err)
	if err != nil {
		if err.Error() == "user address not found" {
			h.logger.Info(c.Request.Context(), "[GetJobsByUserAddress] No user found for address", observability.String("user_address", userAddress))
		} else {
			h.logger.Error(c.Request.Context(), "[GetJobsByUserAddress] Error getting user data for address", observability.String("user_address", userAddress), observability.Error(err))
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
			h.logger.Error(c.Request.Context(), "[GetJobsByUserAddress] Error getting job data for jobID", observability.Int64("job_id", jobID.Int64()), observability.Error(err))
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
				h.logger.Error(c.Request.Context(), "[GetJobsByUserAddress] Error getting time job data for jobID", observability.Int64("job_id", jobID.Int64()), observability.Error(err))
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
				h.logger.Error(c.Request.Context(), "[GetJobsByUserAddress] Error getting event job data for jobID", observability.Int64("job_id", jobID.Int64()), observability.Error(err))
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
				h.logger.Error(c.Request.Context(), "[GetJobsByUserAddress] Error getting condition job data for jobID", observability.Int64("job_id", jobID.Int64()), observability.Error(err))
				hasErrors = true
				continue
			}
			jobResponse.ConditionJobData = &conditionJobData

		default:
			h.logger.Error(c.Request.Context(), "[GetJobsByUserAddress] Unknown task definition ID for jobID", observability.Int("task_definition_id", jobData.TaskDefinitionID), observability.Int64("job_id", jobID.Int64()))
			hasErrors = true
			continue
		}

		jobs = append(jobs, jobResponse)
	}

	h.logger.Info(c.Request.Context(), "[GetJobsByUserAddress] Found jobs for user ID", observability.Int("jobs_count", len(jobs)), observability.Int64("user_id", userID))

	var jobsAPI []types.JobResponseAPI
	for _, job := range jobs {
		jobsAPI = append(jobsAPI, types.ConvertJobResponseToAPI(job))
	}

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
		"jobs": jobsAPI,
	})
}

// GetTaskFeesByJobID handles GET /jobs/:job_id/task-fees
func (h *Handler) GetTaskFeesByJobID(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Info(c.Request.Context(), "[GetTaskFeesByJobID] Getting task fees by job ID", observability.String("trace_id", traceID))

	jobIDParam := c.Param("job_id")
	if jobIDParam == "" {
		h.logger.Error(c.Request.Context(), "[GetTaskFeesByJobID] job_id param missing")
		c.JSON(http.StatusBadRequest, gin.H{"error": "job_id param missing"})
		return
	}

	jobID := new(big.Int)
	_, ok := jobID.SetString(jobIDParam, 10)
	if !ok {
		h.logger.Error(c.Request.Context(), "[GetTaskFeesByJobID] invalid job_id", observability.String("job_id", jobIDParam))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job_id"})
		return
	}

	taskFees, err := h.jobRepository.GetTaskFeesByJobID(jobID)
	if err != nil {
		h.logger.Error(c.Request.Context(), "[GetTaskFeesByJobID] failed to get task fees", observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get task fees"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"task_fees": taskFees})
}

// GetJobsByApiKey handles GET /jobs/by-apikey
func (h *Handler) GetJobsByApiKey(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Info(c.Request.Context(), "[GetJobsByApiKey] Retrieving jobs by API key", observability.String("trace_id", traceID))
	apiKey := c.GetHeader("X-Api-Key")
	if apiKey == "" {
		h.logger.Error(c.Request.Context(), "[GetJobsByApiKey] Missing X-Api-Key header")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing X-Api-Key header"})
		return
	}

	apiKeyData, err := h.apiKeysRepository.GetApiKeyDataByKey(apiKey)
	if err != nil {
		h.logger.Error(c.Request.Context(), "[GetJobsByApiKey] Invalid API key", observability.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key", "code": "INVALID_API_KEY"})
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
}

// GetJobDataByJobID handles GET /jobs/:job_id
func (h *Handler) GetJobDataByJobID(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Info(c.Request.Context(), "[GetJobDataByJobID] Retrieving job data by job ID", observability.String("trace_id", traceID))

	jobIDParam := c.Param("job_id")
	if jobIDParam == "" {
		h.logger.Error(c.Request.Context(), "[GetJobDataByJobID] job_id param missing")
		c.JSON(http.StatusBadRequest, gin.H{"error": "job_id param missing"})
		return
	}

	jobID := new(big.Int)
	_, ok := jobID.SetString(jobIDParam, 10)
	if !ok {
		h.logger.Error(c.Request.Context(), "[GetJobDataByJobID] invalid job_id", observability.String("job_id", jobIDParam))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job_id"})
		return
	}

	jobData, err := h.jobRepository.GetJobByID(jobID)
	if err != nil {
		h.logger.Error(c.Request.Context(), "[GetJobDataByJobID] failed to get job data", observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get job data"})
		return
	}

	c.JSON(http.StatusOK, jobData)
}

func (h *Handler) GetJobsByUserAddressAndChainID(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Info(c.Request.Context(), "[GetJobsByUserAddressAndChainID] Retrieving jobs by user address and created chain id", observability.String("trace_id", traceID))

	userAddress := strings.ToLower(c.Param("user_address"))
	createdChainIDParam := strings.ToLower(c.Param("created_chain_id"))

	if userAddress == "" {
		h.logger.Error(c.Request.Context(), "[GetJobsByUserAddressAndChainID] Invalid user address")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user address",
			"code":  "INVALID_ADDRESS",
		})
		return
	}
	if createdChainIDParam == "" {
		h.logger.Error(c.Request.Context(), "[GetJobsByUserAddressAndChainID] Invalid created_chain_id")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid created_chain_id",
			"code":  "INVALID_CREATED_CHAIN_ID",
		})
		return
	}

	h.logger.Info(c.Request.Context(), "[GetJobsByUserAddressAndChainID] Retrieving jobs for user address and created_chain_id", observability.String("user_address", userAddress), observability.String("created_chain_id", createdChainIDParam))

	// First get user ID and job IDs
	trackDBOp := metrics.TrackDBOperation("read", "user_data")
	_, jobIDs, err := h.userRepository.GetUserJobIDsByAddress(userAddress)
	trackDBOp(err)
	if err != nil {
		if err.Error() == "user address not found" {
			h.logger.Info(c.Request.Context(), "[GetJobsByUserAddressAndChainID] No user found for address", observability.String("user_address", userAddress))
		} else {
			h.logger.Error(c.Request.Context(), "[GetJobsByUserAddressAndChainID] Error getting user data for address", observability.String("user_address", userAddress), observability.Error(err))
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
			h.logger.Error(c.Request.Context(), "[GetJobsByUserAddressAndChainID] Error getting job data for jobID", observability.Int64("job_id", jobID.Int64()), observability.Error(err))
			hasErrors = true
			continue
		}

		// Filter by created_chain_id
		if strings.ToLower(jobData.CreatedChainID) != createdChainIDParam {
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
				h.logger.Error(c.Request.Context(), "[GetJobsByUserAddressAndChainID] Error getting time job data for jobID", observability.Int64("job_id", jobID.Int64()), observability.Error(err))
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
				h.logger.Error(c.Request.Context(), "[GetJobsByUserAddressAndChainID] Error getting event job data for jobID", observability.Int64("job_id", jobID.Int64()), observability.Error(err))
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
				h.logger.Error(c.Request.Context(), "[GetJobsByUserAddressAndChainID] Error getting condition job data for jobID", observability.Int64("job_id", jobID.Int64()), observability.Error(err))
				hasErrors = true
				continue
			}
			jobResponse.ConditionJobData = &conditionJobData

		default:
			h.logger.Error(c.Request.Context(), "[GetJobsByUserAddressAndChainID] Unknown task definition ID for jobID", observability.Int("task_definition_id", jobData.TaskDefinitionID), observability.Int64("job_id", jobID.Int64()))
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
}

// parseInt64 is a helper to parse int64 from string
// func parseInt64(s string) (int64, error) {
// 	var i int64
// 	_, err := fmt.Sscan(s, &i)
// 	return i, err
// }
