package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/errors"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// GetJobDataByJobID handles GET /jobs/:job_id
func (h *Handler) GetJobDataByJobID(c *gin.Context) {
	logger := h.getLogger(c)
	jobID := c.Param("job_id")
	if !types.IsValidJobID(jobID) {
		logger.Errorf("%s: %s", errors.ErrInvalidRequestBody, "Invalid job ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.ErrInvalidRequestBody})
		return
	}
	logger.Debugf("GET [GetJobDataByJobID] For job ID: %s", jobID)

	var jobData *types.CompleteJobDataDTO
	trackDBOp := metrics.TrackDBOperation("read", "job_data")
	baseJobData, err := h.jobRepository.GetByID(c.Request.Context(), jobID)
	trackDBOp(err)
	if err != nil || baseJobData == nil {
		logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}
	jobData = &types.CompleteJobDataDTO{JobDataDTO: *types.ConvertJobDataEntityToDTO(baseJobData)}

	switch baseJobData.TaskDefinitionID {
	case 1, 2:
		trackDBOp = metrics.TrackDBOperation("read", "time_job")
		timeJobData, err := h.timeJobRepository.GetByID(c.Request.Context(), jobID)
		trackDBOp(err)
		if err != nil {
			logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
			c.JSON(http.StatusNotFound, gin.H{"error": "time job not found"})
			return
		}
		jobData.TimeJobDataDTO = types.ConvertTimeJobDataEntityToDTO(timeJobData)
	case 3, 4:
		trackDBOp = metrics.TrackDBOperation("read", "event_job")
		eventJobData, err := h.eventJobRepository.GetByID(c.Request.Context(), jobID)
		trackDBOp(err)
		if err != nil {
			logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
			c.JSON(http.StatusNotFound, gin.H{"error": "event job not found"})
			return
		}
		jobData.EventJobDataDTO = types.ConvertEventJobDataEntityToDTO(eventJobData)
	case 5, 6:
		trackDBOp = metrics.TrackDBOperation("read", "condition_job")
		conditionJobData, err := h.conditionJobRepository.GetByID(c.Request.Context(), jobID)
		trackDBOp(err)
		if err != nil {
			logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
			c.JSON(http.StatusNotFound, gin.H{"error": "condition job not found"})
			return
		}
		jobData.ConditionJobDataDTO = types.ConvertConditionJobDataEntityToDTO(conditionJobData)
	}

	logger.Infof("GET [GetJobDataByJobID] Successful, job ID: %s", jobID)
	c.JSON(http.StatusOK, jobData)
}

func (h *Handler) GetJobsByUserAddress(c *gin.Context) {
	logger := h.getLogger(c)
	userAddress := strings.ToLower(c.Param("user_address"))
	if !types.IsValidEthAddress(userAddress) {
		logger.Errorf("%s: %s", errors.ErrInvalidRequestBody, "Invalid user address")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.ErrInvalidRequestBody,
		})
		return
	}

	// Get optional chain_id filter from query parameters
	chainIDFilter := c.Query("chain_id")
	if chainIDFilter != "" {
		logger.Debugf("GET [GetJobsByUserAddress] For user address: %s and chain_id: %s", userAddress, chainIDFilter)
	} else {
		logger.Debugf("GET [GetJobsByUserAddress] For user address: %s", userAddress)
	}

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

		// Apply chain_id filter if provided
		if chainIDFilter != "" && jobData.CreatedChainID != chainIDFilter {
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
			jobResponse.TimeJobDataDTO = types.ConvertTimeJobDataEntityToDTO(timeJobData)

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
			jobResponse.EventJobDataDTO = types.ConvertEventJobDataEntityToDTO(eventJobData)

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
			jobResponse.ConditionJobDataDTO = types.ConvertConditionJobDataEntityToDTO(conditionJobData)

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

	if chainIDFilter != "" {
		logger.Infof("GET [GetJobsByUserAddress] Successful, jobs: %d, user address: %s, chain_id: %s", len(jobs), userAddress, chainIDFilter)
	} else {
		logger.Infof("GET [GetJobsByUserAddress] Successful, jobs: %d, user address: %s", len(jobs), userAddress)
	}
	c.JSON(http.StatusOK, gin.H{
		"jobs": jobs,
	})
}
