package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// GetSafeAddressesByUser handles GET /users/safe-addresses/:user_address
func (h *Handler) GetSafeAddressesByUser(c *gin.Context) {
	userAddress := c.Param("user_address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_address param required"})
		return
	}

	safeAddresses, err := h.safeAddressRepository.GetSafeAddressesByUser(userAddress)
	if err != nil {
		h.logger.Error(c.Request.Context(), "Error fetching safe addresses for user", observability.String("user_address", userAddress), observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch safe addresses"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"safe_addresses": safeAddresses})
	h.logger.Debug(c.Request.Context(), "[GetSafeAddressesByUser] Retrieved safe addresses", observability.String("user_address", userAddress), observability.Int("safe_addresses_count", len(safeAddresses)))
}

// GetJobsBySafeAddress handles GET /jobs/safe-address/:safe_address
func (h *Handler) GetJobsBySafeAddress(c *gin.Context) {
	safeAddress := strings.ToLower(c.Param("safe_address"))
	if safeAddress == "" {
		h.logger.Error(c.Request.Context(), "[GetJobsBySafeAddress] Invalid safe address")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid safe address",
			"code":  "INVALID_SAFE_ADDRESS",
		})
		return
	}

	// Get jobs by safe address
	trackDBOp := metrics.TrackDBOperation("read", "job_data")
	jobDataList, err := h.jobRepository.GetJobsBySafeAddress(safeAddress)
	trackDBOp(err)
	if err != nil {
		h.logger.Error(c.Request.Context(), "[GetJobsBySafeAddress] Error getting jobs for safe address", observability.String("safe_address", safeAddress), observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve jobs",
			"code":  "JOB_RETRIEVAL_ERROR",
		})
		return
	}

	if len(jobDataList) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "No jobs found for this safe address",
			"jobs":    []types.JobResponseAPI{},
		})
		return
	}

	var jobs []types.JobResponse
	var hasErrors bool

	for _, jobData := range jobDataList {
		jobResponse := types.JobResponse{JobData: jobData}

		// Check task_definition_id to determine job type
		switch jobData.TaskDefinitionID {
		case 1, 2, 7:
			// Time-based job (TDI 1, 2) or Agent job (TDI 7)
			trackDBOp = metrics.TrackDBOperation("read", "time_job")
			timeJobData, err := h.timeJobRepository.GetTimeJobByJobID(jobData.JobID)
			trackDBOp(err)
			if err != nil {
				h.logger.Error(c.Request.Context(), "[GetJobsBySafeAddress] Error getting time job data for jobID", observability.String("job_id", jobData.JobID), observability.Error(err))
				hasErrors = true
				continue
			}
			jobResponse.TimeJobData = timeJobData

		case 3, 4, 8:
			// Event-based job (TDI 3, 4) or Agent job (TDI 8)
			trackDBOp = metrics.TrackDBOperation("read", "event_job")
			eventJobData, err := h.eventJobRepository.GetEventJobByJobID(jobData.JobID)
			trackDBOp(err)
			if err != nil {
				h.logger.Error(c.Request.Context(), "[GetJobsBySafeAddress] Error getting event job data for jobID", observability.String("job_id", jobData.JobID), observability.Error(err))
				hasErrors = true
				continue
			}
			jobResponse.EventJobData = eventJobData

		case 5, 6, 9:
			// Condition-based job (TDI 5, 6) or Agent job (TDI 9)
			trackDBOp = metrics.TrackDBOperation("read", "condition_job")
			conditionJobData, err := h.conditionJobRepository.GetConditionJobByJobID(jobData.JobID)
			trackDBOp(err)
			if err != nil {
				h.logger.Error(c.Request.Context(), "[GetJobsBySafeAddress] Error getting condition job data for jobID", observability.String("job_id", jobData.JobID), observability.Error(err))
				hasErrors = true
				continue
			}
			jobResponse.ConditionJobData = conditionJobData

		default:
			h.logger.Error(c.Request.Context(), "[GetJobsBySafeAddress] Unknown task definition ID for jobID", observability.Int("task_definition_id", jobData.TaskDefinitionID), observability.String("job_id", jobData.JobID))
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
	h.logger.Debug(c.Request.Context(), "[GetJobsBySafeAddress] Retrieved jobs", observability.String("safe_address", safeAddress), observability.Int("jobs_count", len(jobs)))
}
