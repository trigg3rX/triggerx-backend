package handlers

import (
	"context"
	"math/big"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// Helper functions to convert Entity types to API types
func convertJobDataEntityToAPI(entity *types.JobDataEntity) types.JobDataAPI {
	return types.JobDataAPI{
		JobID:             &entity.JobID,
		JobTitle:          entity.JobTitle,
		TaskDefinitionID:  entity.TaskDefinitionID,
		UserID:            entity.UserID,
		CreatedChainID:    entity.CreatedChainID,
		LinkJobID:         &entity.LinkJobID,
		ChainStatus:       entity.ChainStatus,
		TimeFrame:         entity.TimeFrame,
		Recurring:         entity.Recurring,
		Status:            entity.Status,
		JobCostPrediction: 0, // Convert from big.Int if needed
		JobCostActual:     0, // Convert from big.Int if needed
		TaskIDs:           entity.TaskIDs,
		CreatedAt:         entity.CreatedAt,
		UpdatedAt:         entity.UpdatedAt,
		LastExecutedAt:    entity.LastExecutedAt,
	}
}

func convertTimeJobDataEntityToAPI(entity *types.TimeJobDataEntity) *types.TimeJobDataAPI {
	if entity == nil {
		return nil
	}
	return &types.TimeJobDataAPI{
		JobID:                     &entity.JobID,
		TaskDefinitionID:          entity.TaskDefinitionID,
		ScheduleType:              entity.ScheduleType,
		TimeInterval:              entity.TimeInterval,
		CronExpression:            entity.CronExpression,
		SpecificSchedule:          entity.SpecificSchedule,
		NextExecutionTimestamp:    entity.NextExecutionTimestamp,
		TargetChainID:             entity.TargetChainID,
		TargetContractAddress:     entity.TargetContractAddress,
		TargetFunction:            entity.TargetFunction,
		ABI:                       entity.ABI,
		ArgType:                   entity.ArgType,
		Arguments:                 entity.Arguments,
		DynamicArgumentsScriptUrl: entity.DynamicArgumentsScriptURL,
		IsCompleted:               entity.IsCompleted,
		LastExecutedAt:            entity.LastExecutedAt,
		ExpirationTime:            entity.ExpirationTime,
	}
}

func convertEventJobDataEntityToAPI(entity *types.EventJobDataEntity) *types.EventJobDataAPI {
	if entity == nil {
		return nil
	}
	return &types.EventJobDataAPI{
		JobID:                     &entity.JobID,
		TaskDefinitionID:          entity.TaskDefinitionID,
		Recurring:                 entity.Recurring,
		TriggerChainID:            entity.TriggerChainID,
		TriggerContractAddress:    entity.TriggerContractAddress,
		TriggerEvent:              entity.TriggerEvent,
		EventFilterParaName:       entity.TriggerEventFilterParaName,
		EventFilterValue:          entity.TriggerEventFilterValue,
		TargetChainID:             entity.TargetChainID,
		TargetContractAddress:     entity.TargetContractAddress,
		TargetFunction:            entity.TargetFunction,
		ABI:                       entity.ABI,
		ArgType:                   entity.ArgType,
		Arguments:                 entity.Arguments,
		DynamicArgumentsScriptUrl: entity.DynamicArgumentsScriptURL,
		IsCompleted:               entity.IsCompleted,
		LastExecutedAt:            entity.LastExecutedAt,
		ExpirationTime:            entity.ExpirationTime,
	}
}

func convertConditionJobDataEntityToAPI(entity *types.ConditionJobDataEntity) *types.ConditionJobDataAPI {
	if entity == nil {
		return nil
	}
	return &types.ConditionJobDataAPI{
		JobID:                     &entity.JobID,
		TaskDefinitionID:          entity.TaskDefinitionID,
		Recurring:                 entity.Recurring,
		ConditionType:             entity.ConditionType,
		UpperLimit:                entity.UpperLimit,
		LowerLimit:                entity.LowerLimit,
		ValueSourceType:           entity.ValueSourceType,
		ValueSourceUrl:            entity.ValueSourceURL,
		SelectedKeyRoute:          entity.SelectedKeyRoute,
		TargetChainID:             entity.TargetChainID,
		TargetContractAddress:     entity.TargetContractAddress,
		TargetFunction:            entity.TargetFunction,
		ABI:                       entity.ABI,
		ArgType:                   entity.ArgType,
		Arguments:                 entity.Arguments,
		DynamicArgumentsScriptUrl: entity.DynamicArgumentsScriptURL,
		IsCompleted:               entity.IsCompleted,
		LastExecutedAt:            entity.LastExecutedAt,
		ExpirationTime:            entity.ExpirationTime,
	}
}

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

	ctx := context.Background()

	// First get user ID and job IDs
	trackDBOp := metrics.TrackDBOperation("read", "user_data")
	user, err := h.userRepository.GetByNonID(ctx, "user_address", userAddress)
	trackDBOp(err)
	if err != nil {
		h.logger.Errorf("[GetJobsByUserAddress] Error getting user data for address %s: %v", userAddress, err)
		c.JSON(http.StatusOK, gin.H{
			"message": "No jobs found for this user",
			"jobs":    []types.JobResponse{},
		})
		return
	}

	if user == nil || len(user.JobIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "No jobs found for this user",
			"jobs":    []types.JobResponse{},
		})
		return
	}

	userID := user.UserID
	jobIDs := user.JobIDs

	var jobs []types.JobResponse
	var hasErrors bool

	for _, jobID := range jobIDs {
		// Get basic job data
		trackDBOp = metrics.TrackDBOperation("read", "job_data")
		jobData, err := h.jobRepository.GetByID(ctx, &jobID)
		trackDBOp(err)
		if err != nil {
			h.logger.Errorf("[GetJobsByUserAddress] Error getting job data for jobID %v: %v", &jobID, err)
			hasErrors = true
			continue
		}

		if jobData == nil {
			h.logger.Errorf("[GetJobsByUserAddress] Job data not found for jobID %v", &jobID)
			hasErrors = true
			continue
		}

		jobResponse := types.JobResponse{JobData: convertJobDataEntityToAPI(jobData)}

		// Check task_definition_id to determine job type
		switch jobData.TaskDefinitionID {
		case 1, 2:
			// Time-based job
			trackDBOp = metrics.TrackDBOperation("read", "time_job")
			timeJobData, err := h.timeJobRepository.GetByNonID(ctx, "job_id", &jobID)
			trackDBOp(err)
			if err != nil {
				h.logger.Errorf("[GetJobsByUserAddress] Error getting time job data for jobID %v: %v", &jobID, err)
				hasErrors = true
				continue
			}
			jobResponse.TimeJobData = convertTimeJobDataEntityToAPI(timeJobData)

		case 3, 4:
			// Event-based job
			trackDBOp = metrics.TrackDBOperation("read", "event_job")
			eventJobData, err := h.eventJobRepository.GetByNonID(ctx, "job_id", &jobID)
			trackDBOp(err)
			if err != nil {
				h.logger.Errorf("[GetJobsByUserAddress] Error getting event job data for jobID %v: %v", &jobID, err)
				hasErrors = true
				continue
			}
			jobResponse.EventJobData = convertEventJobDataEntityToAPI(eventJobData)

		case 5, 6:
			// Condition-based job
			trackDBOp = metrics.TrackDBOperation("read", "condition_job")
			conditionJobData, err := h.conditionJobRepository.GetByNonID(ctx, "job_id", &jobID)
			trackDBOp(err)
			if err != nil {
				h.logger.Errorf("[GetJobsByUserAddress] Error getting condition job data for jobID %v: %v", &jobID, err)
				hasErrors = true
				continue
			}
			jobResponse.ConditionJobData = convertConditionJobDataEntityToAPI(conditionJobData)

		default:
			h.logger.Errorf("[GetJobsByUserAddress] Unknown task definition ID %d for jobID %v", jobData.TaskDefinitionID, &jobID)
			hasErrors = true
			continue
		}

		jobs = append(jobs, jobResponse)
	}

	h.logger.Infof("[GetJobsByUserAddress] Found %d jobs for user ID %d", len(jobs), userID)

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
	h.logger.Infof("[GetTaskFeesByJobID] trace_id=%s - Getting task fees by job ID", traceID)

	jobIDParam := c.Param("job_id")
	if jobIDParam == "" {
		h.logger.Error("[GetTaskFeesByJobID] job_id param missing")
		c.JSON(http.StatusBadRequest, gin.H{"error": "job_id param missing"})
		return
	}

	jobID := new(big.Int)
	_, ok := jobID.SetString(jobIDParam, 10)
	if !ok {
		h.logger.Errorf("[GetTaskFeesByJobID] invalid job_id: %v", jobIDParam)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job_id"})
		return
	}

	ctx := context.Background()

	// Get job data
	job, err := h.jobRepository.GetByID(ctx, jobID)
	if err != nil || job == nil {
		h.logger.Errorf("[GetTaskFeesByJobID] failed to get job: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}

	// Get task fees from task_ids
	var taskFees []map[string]interface{}
	for _, taskID := range job.TaskIDs {
		task, err := h.taskRepository.GetByID(ctx, taskID)
		if err != nil || task == nil {
			continue
		}
		taskFees = append(taskFees, map[string]interface{}{
			"task_id":                 task.TaskID,
			"task_opx_predicted_cost": task.TaskOpxPredictedCost,
			"task_opx_actual_cost":    task.TaskOpxActualCost,
		})
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

	ctx := context.Background()

	apiKeyData, err := h.apiKeysRepository.GetByNonID(ctx, "key", apiKey)
	if err != nil || apiKeyData == nil {
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

// GetJobDataByJobID handles GET /jobs/:job_id
func (h *Handler) GetJobDataByJobID(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Infof("[GetJobDataByJobID] trace_id=%s - Retrieving job data by job ID", traceID)

	jobIDParam := c.Param("job_id")
	if jobIDParam == "" {
		h.logger.Error("[GetJobDataByJobID] job_id param missing")
		c.JSON(http.StatusBadRequest, gin.H{"error": "job_id param missing"})
		return
	}

	jobID := new(big.Int)
	_, ok := jobID.SetString(jobIDParam, 10)
	if !ok {
		h.logger.Errorf("[GetJobDataByJobID] invalid job_id: %v", jobIDParam)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job_id"})
		return
	}

	ctx := context.Background()
	jobData, err := h.jobRepository.GetByID(ctx, jobID)
	if err != nil || jobData == nil {
		h.logger.Errorf("[GetJobDataByJobID] failed to get job data: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}

	c.JSON(http.StatusOK, jobData)
}

func (h *Handler) GetJobsByUserAddressAndChainID(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Infof("[GetJobsByUserAddressAndChainID] trace_id=%s - Retrieving jobs by user address and created chain id", traceID)

	userAddress := strings.ToLower(c.Param("user_address"))
	createdChainIDParam := strings.ToLower(c.Param("created_chain_id"))

	if userAddress == "" {
		h.logger.Error("[GetJobsByUserAddressAndChainID] Invalid user address")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user address",
			"code":  "INVALID_ADDRESS",
		})
		return
	}
	if createdChainIDParam == "" {
		h.logger.Error("[GetJobsByUserAddressAndChainID] Invalid created_chain_id")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid created_chain_id",
			"code":  "INVALID_CREATED_CHAIN_ID",
		})
		return
	}

	h.logger.Infof("[GetJobsByUserAddressAndChainID] Retrieving jobs for user address: %s and created_chain_id: %s", userAddress, createdChainIDParam)

	ctx := context.Background()

	// First get user ID and job IDs
	trackDBOp := metrics.TrackDBOperation("read", "user_data")
	user, err := h.userRepository.GetByNonID(ctx, "user_address", userAddress)
	trackDBOp(err)
	if err != nil {
		h.logger.Errorf("[GetJobsByUserAddressAndChainID] Error getting user data for address %s: %v", userAddress, err)
		c.JSON(http.StatusOK, gin.H{
			"message": "No jobs found for this user",
			"jobs":    []types.JobResponse{},
		})
		return
	}

	if user == nil || len(user.JobIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "No jobs found for this user",
			"jobs":    []types.JobResponse{},
		})
		return
	}

	jobIDs := user.JobIDs

	var jobs []types.JobResponse
	var hasErrors bool

	for _, jobID := range jobIDs {
		// Get basic job data
		trackDBOp = metrics.TrackDBOperation("read", "job_data")
		jobData, err := h.jobRepository.GetByID(ctx, &jobID)
		trackDBOp(err)
		if err != nil {
			h.logger.Errorf("[GetJobsByUserAddressAndChainID] Error getting job data for jobID %v: %v", &jobID, err)
			hasErrors = true
			continue
		}

		if jobData == nil {
			h.logger.Errorf("[GetJobsByUserAddressAndChainID] Job data not found for jobID %v", &jobID)
			hasErrors = true
			continue
		}

		// Filter by created_chain_id
		if strings.ToLower(jobData.CreatedChainID) != createdChainIDParam {
			continue
		}

		jobResponse := types.JobResponse{JobData: convertJobDataEntityToAPI(jobData)}

		// Check task_definition_id to determine job type
		switch jobData.TaskDefinitionID {
		case 1, 2:
			// Time-based job
			trackDBOp = metrics.TrackDBOperation("read", "time_job")
			timeJobData, err := h.timeJobRepository.GetByNonID(ctx, "job_id", &jobID)
			trackDBOp(err)
			if err != nil {
				h.logger.Errorf("[GetJobsByUserAddressAndChainID] Error getting time job data for jobID %v: %v", &jobID, err)
				hasErrors = true
				continue
			}
			jobResponse.TimeJobData = convertTimeJobDataEntityToAPI(timeJobData)

		case 3, 4:
			// Event-based job
			trackDBOp = metrics.TrackDBOperation("read", "event_job")
			eventJobData, err := h.eventJobRepository.GetByNonID(ctx, "job_id", &jobID)
			trackDBOp(err)
			if err != nil {
				h.logger.Errorf("[GetJobsByUserAddressAndChainID] Error getting event job data for jobID %v: %v", &jobID, err)
				hasErrors = true
				continue
			}
			jobResponse.EventJobData = convertEventJobDataEntityToAPI(eventJobData)

		case 5, 6:
			// Condition-based job
			trackDBOp = metrics.TrackDBOperation("read", "condition_job")
			conditionJobData, err := h.conditionJobRepository.GetByNonID(ctx, "job_id", &jobID)
			trackDBOp(err)
			if err != nil {
				h.logger.Errorf("[GetJobsByUserAddressAndChainID] Error getting condition job data for jobID %v: %v", &jobID, err)
				hasErrors = true
				continue
			}
			jobResponse.ConditionJobData = convertConditionJobDataEntityToAPI(conditionJobData)

		default:
			h.logger.Errorf("[GetJobsByUserAddressAndChainID] Unknown task definition ID %d for jobID %v", jobData.TaskDefinitionID, &jobID)
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
