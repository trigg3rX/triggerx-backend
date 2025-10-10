package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/errors"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (h *Handler) GetJobsByUserAddress(c *gin.Context) {
	logger := h.getLogger(c)
	userAddress := strings.ToLower(c.Param("user_address"))
	if types.IsValidEthAddress(userAddress) {
		logger.Errorf("%s: %s", errors.ErrInvalidRequestBody, "Invalid user address")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.ErrInvalidRequestBody,
		})
		return
	}
	logger.Debugf("GET [GetJobsByUserAddress] For user address: %s", userAddress)

	// First get user Address and job IDs
	trackDBOp := metrics.TrackDBOperation("read", "user_data")
	userData, err := h.userRepository.GetByID(c.Request.Context(), userAddress)
	trackDBOp(err)
	if err != nil {
		logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
		c.JSON(http.StatusOK, gin.H{
			"message": errors.ErrDBOperationFailed,
		})
		return
	}

	if userData == nil || len(userData.JobIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": errors.ErrDBRecordNotFound,
			"jobs":    []types.CompleteJobDataDTO{},
		})
		return
	}

	var jobs []types.CompleteJobDataDTO
	var hasErrors bool

	for _, jobID := range userData.JobIDs {
		// Get basic job data
		trackDBOp = metrics.TrackDBOperation("read", "job_data")
		jobData, err := h.jobRepository.GetByID(c.Request.Context(), jobID)
		trackDBOp(err)
		if err != nil {
			logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
			hasErrors = true
			continue
		}

		if jobData == nil {
			logger.Errorf("%s: %v", errors.ErrDBRecordNotFound, err)
			hasErrors = true
			continue
		}
		jobResponse := types.CompleteJobDataDTO{JobDataDTO: *types.ConvertJobDataEntityToDTO(jobData)}

		// Check task_definition_id to determine job type
		switch jobData.TaskDefinitionID {
		case 1, 2:
			// Time-based job
			trackDBOp = metrics.TrackDBOperation("read", "time_job")
			timeJobData, err := h.timeJobRepository.GetByID(c.Request.Context(), jobID)
			trackDBOp(err)
			if err != nil {
				logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
				hasErrors = true
				continue
			}
			jobResponse.TimeJobDataDTO = *types.ConvertTimeJobDataEntityToDTO(timeJobData)

		case 3, 4:
			// Event-based job
			trackDBOp = metrics.TrackDBOperation("read", "event_job")
			eventJobData, err := h.eventJobRepository.GetByID(c.Request.Context(), jobID)
			trackDBOp(err)
			if err != nil {
				logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
				hasErrors = true
				continue
			}
			jobResponse.EventJobDataDTO = *types.ConvertEventJobDataEntityToDTO(eventJobData)

		case 5, 6:
			// Condition-based job
			trackDBOp = metrics.TrackDBOperation("read", "condition_job")
			conditionJobData, err := h.conditionJobRepository.GetByID(c.Request.Context(), jobID)
			trackDBOp(err)
			if err != nil {
				logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
				hasErrors = true
				continue
			}
			jobResponse.ConditionJobDataDTO = *types.ConvertConditionJobDataEntityToDTO(conditionJobData)

		default:
			logger.Errorf("Unknown task definition ID %d for jobID %v", jobData.TaskDefinitionID, jobID)
			hasErrors = true
			continue
		}

		jobs = append(jobs, jobResponse)
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
			"error": errors.ErrDBOperationFailed,
		})
		return
	}

	logger.Infof("Successfully retrieved %d jobs for user address %s", len(jobs), userAddress)
	c.JSON(http.StatusOK, gin.H{
		"jobs": jobs,
	})
}

// GetTaskFeesByJobID handles GET /jobs/:job_id/task-fees
func (h *Handler) GetTaskFeesByJobID(c *gin.Context) {
	logger := h.getLogger(c)
	jobID := c.Param("job_id")
	if types.IsValidJobID(jobID) {
		logger.Errorf("%s: %s", errors.ErrInvalidRequestBody, "Invalid job ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.ErrInvalidRequestBody})
		return
	}
	logger.Debugf("GET [GetTaskFeesByJobID] For job ID: %s", jobID)

	// Get job data
	trackDBOp := metrics.TrackDBOperation("read", "job_data")
	job, err := h.jobRepository.GetByID(c.Request.Context(), jobID)
	trackDBOp(err)
	if err != nil || job == nil {
		logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}

	// Get task fees from task_ids
	var taskFees []types.GetTaskFeesByJobIDResponse
	for _, taskID := range job.TaskIDs {
		trackDBOp = metrics.TrackDBOperation("read", "task_data")
		task, err := h.taskRepository.GetByID(c.Request.Context(), taskID)
		trackDBOp(err)
		if err != nil || task == nil {
			trackDBOp(err)
			continue
		}
		taskFees = append(taskFees, types.GetTaskFeesByJobIDResponse{
			TaskID:               task.TaskID,
			TaskOpxPredictedCost: task.TaskOpxPredictedCost,
			TaskOpxActualCost:    task.TaskOpxActualCost,
		})
	}

	c.JSON(http.StatusOK, gin.H{"task_fees": taskFees})
}

// GetJobsByApiKey handles GET /jobs/by-apikey
func (h *Handler) GetJobsByApiKey(c *gin.Context) {
	logger := h.getLogger(c)
	logger.Debugf("GET [GetJobsByApiKey] Retrieving jobs by API key")
	apiKey := c.GetHeader("X-Api-Key")
	if apiKey == "" {
		logger.Errorf("Missing X-Api-Key header")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing X-Api-Key header"})
		return
	}

	trackDBOp := metrics.TrackDBOperation("read", "api_key_data")
	apiKeyData, err := h.apiKeysRepository.GetByID(c.Request.Context(), apiKey)
	trackDBOp(err)
	if err != nil || apiKeyData == nil {
		logger.Errorf("Invalid API key: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
		return
	}

	userAddress := apiKeyData.Owner
	if userAddress == "" {
		logger.Errorf("No owner found for API key")
		c.JSON(http.StatusNotFound, gin.H{"error": "No owner found for API key"})
		return
	}

	// Reuse the logic from GetJobsByUserAddress
	c.Params = append(c.Params, gin.Param{Key: "user_address", Value: userAddress})
	h.GetJobsByUserAddress(c)
}

// GetJobDataByJobID handles GET /jobs/:job_id
func (h *Handler) GetJobDataByJobID(c *gin.Context) {
	logger := h.getLogger(c)
	jobID := c.Param("job_id")
	if types.IsValidJobID(jobID) {
		logger.Errorf("%s: %s", errors.ErrInvalidRequestBody, "Invalid job ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.ErrInvalidRequestBody})
		return
	}
	logger.Debugf("GET [GetJobDataByJobID] For job ID: %s", c.Param("job_id"))

	trackDBOp := metrics.TrackDBOperation("read", "job_data")
	jobData, err := h.jobRepository.GetByID(c.Request.Context(), jobID)
	trackDBOp(err)
	if err != nil || jobData == nil {
		logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}

	c.JSON(http.StatusOK, jobData)
}

func (h *Handler) GetJobsByUserAddressAndChainID(c *gin.Context) {
	logger := h.getLogger(c)
	userAddress := strings.ToLower(c.Param("user_address"))
	createdChainID := strings.ToLower(c.Param("created_chain_id"))
	if types.IsValidEthAddress(userAddress) {	
		logger.Errorf("Invalid user address")
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.ErrInvalidRequestBody})
		return
	}
	if types.IsValidChainID(createdChainID) {
		logger.Errorf("Invalid created_chain_id")
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.ErrInvalidRequestBody})
		return
	}
	logger.Debugf("GET [GetJobsByUserAddressAndChainID] For user address: %s and created chain ID: %s", userAddress, createdChainID)

	// First get user address and job IDs
	trackDBOp := metrics.TrackDBOperation("read", "user_data")
	user, err := h.userRepository.GetByNonID(c.Request.Context(), "user_address", userAddress)
	trackDBOp(err)
	if err != nil {
		logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
		c.JSON(http.StatusOK, gin.H{
			"message": errors.ErrDBRecordNotFound,
		})
		return
	}

	if user == nil || len(user.JobIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": errors.ErrDBRecordNotFound,
			"jobs":    []types.CompleteJobDataDTO{},
		})
		return
	}

	var jobs []types.CompleteJobDataDTO
	var hasErrors bool

	for _, jobID := range user.JobIDs {
		// Get basic job data
		trackDBOp = metrics.TrackDBOperation("read", "job_data")
		jobData, err := h.jobRepository.GetByID(c.Request.Context(), jobID)
		trackDBOp(err)
		if err != nil {
			logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
			hasErrors = true
			continue
		}

		if jobData == nil {
			logger.Errorf("%s: %v", errors.ErrDBRecordNotFound, err)
			hasErrors = true
			continue
		}

		// Filter by created_chain_id
		if strings.ToLower(jobData.CreatedChainID) != createdChainID {
			continue
		}

		jobResponse := types.CompleteJobDataDTO{JobDataDTO: *types.ConvertJobDataEntityToDTO(jobData)}

		// Check task_definition_id to determine job type
		switch jobData.TaskDefinitionID {
		case 1, 2:
			// Time-based job
			trackDBOp = metrics.TrackDBOperation("read", "time_job")
			timeJobData, err := h.timeJobRepository.GetByNonID(c.Request.Context(), "job_id", jobID)
			trackDBOp(err)
			if err != nil {
				logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
				hasErrors = true
				continue
			}
			jobResponse.TimeJobDataDTO = *types.ConvertTimeJobDataEntityToDTO(timeJobData)

		case 3, 4:
			// Event-based job
			trackDBOp = metrics.TrackDBOperation("read", "event_job")
			eventJobData, err := h.eventJobRepository.GetByNonID(c.Request.Context(), "job_id", jobID)
			trackDBOp(err)
			if err != nil {
				logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
				hasErrors = true
				continue
			}
			jobResponse.EventJobDataDTO = *types.ConvertEventJobDataEntityToDTO(eventJobData)

		case 5, 6:
			// Condition-based job
			trackDBOp = metrics.TrackDBOperation("read", "condition_job")
			conditionJobData, err := h.conditionJobRepository.GetByNonID(c.Request.Context(), "job_id", jobID)
			trackDBOp(err)
			if err != nil {
				logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
				hasErrors = true
				continue
			}
			jobResponse.ConditionJobDataDTO = *types.ConvertConditionJobDataEntityToDTO(conditionJobData)

		default:
			logger.Errorf("Unknown task definition ID %d for jobID %v", jobData.TaskDefinitionID, jobID)
			hasErrors = true
			continue
		}

		jobs = append(jobs, jobResponse)
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
			"error": errors.ErrDBOperationFailed,
			"code":  "JOB_RETRIEVAL_ERROR",
		})
		return
	}

	// If we have only jobs and no errors, return success
	c.JSON(http.StatusOK, gin.H{
		"jobs": jobs,
	})
}
