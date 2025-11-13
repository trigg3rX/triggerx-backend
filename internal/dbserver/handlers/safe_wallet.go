package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
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
		h.logger.Errorf("Error fetching safe addresses for user %s: %v", userAddress, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch safe addresses"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"safe_addresses": safeAddresses})
}

// GetJobsBySafeAddress handles GET /jobs/safe-address/:safe_address
func (h *Handler) GetJobsBySafeAddress(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Infof("[GetJobsBySafeAddress] trace_id=%s - Retrieving jobs by safe address", traceID)

	safeAddress := strings.ToLower(c.Param("safe_address"))
	if safeAddress == "" {
		h.logger.Error("[GetJobsBySafeAddress] Invalid safe address")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid safe address",
			"code":  "INVALID_SAFE_ADDRESS",
		})
		return
	}

	h.logger.Infof("[GetJobsBySafeAddress] Retrieving jobs for safe address: %s", safeAddress)

	// Get jobs by safe address
	trackDBOp := metrics.TrackDBOperation("read", "job_data")
	jobDataList, err := h.jobRepository.GetJobsBySafeAddress(safeAddress)
	trackDBOp(err)
	if err != nil {
		h.logger.Errorf("[GetJobsBySafeAddress] Error getting jobs for safe address %s: %v", safeAddress, err)
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
		case 1, 2:
			// Time-based job
			trackDBOp = metrics.TrackDBOperation("read", "time_job")
			timeJobData, err := h.timeJobRepository.GetTimeJobByJobID(jobData.JobID.ToBigInt())
			trackDBOp(err)
			if err != nil {
				h.logger.Errorf("[GetJobsBySafeAddress] Error getting time job data for jobID %s: %v", jobData.JobID.String(), err)
				hasErrors = true
				continue
			}
			jobResponse.TimeJobData = &timeJobData

		case 3, 4:
			// Event-based job
			trackDBOp = metrics.TrackDBOperation("read", "event_job")
			eventJobData, err := h.eventJobRepository.GetEventJobByJobID(jobData.JobID.ToBigInt())
			trackDBOp(err)
			if err != nil {
				h.logger.Errorf("[GetJobsBySafeAddress] Error getting event job data for jobID %s: %v", jobData.JobID.String(), err)
				hasErrors = true
				continue
			}
			jobResponse.EventJobData = &eventJobData

		case 5, 6:
			// Condition-based job
			trackDBOp = metrics.TrackDBOperation("read", "condition_job")
			conditionJobData, err := h.conditionJobRepository.GetConditionJobByJobID(jobData.JobID.ToBigInt())
			trackDBOp(err)
			if err != nil {
				h.logger.Errorf("[GetJobsBySafeAddress] Error getting condition job data for jobID %s: %v", jobData.JobID.String(), err)
				hasErrors = true
				continue
			}
			jobResponse.ConditionJobData = &conditionJobData

		default:
			h.logger.Errorf("[GetJobsBySafeAddress] Unknown task definition ID %d for jobID %s", jobData.TaskDefinitionID, jobData.JobID.String())
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

	h.logger.Infof("[GetJobsBySafeAddress] Successfully retrieved jobs: %+v", jobsAPI)

	// If we have only jobs and no errors, return success
	c.JSON(http.StatusOK, gin.H{
		"jobs": jobsAPI,
	})

}
