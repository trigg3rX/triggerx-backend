package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/parser"
	"github.com/trigg3rX/triggerx-backend/pkg/errors"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (h *Handler) CreateJobData(c *gin.Context) {
	logger := h.getLogger(c)
	var tempJobs []types.CreateJobDataRequest
	if err := c.ShouldBindJSON(&tempJobs); err != nil {
		logger.Errorf("%s: %v", errors.ErrInvalidRequestBody, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.ErrInvalidRequestBody})
		return
	}
	if len(tempJobs) == 0 {
		logger.Error("No jobs provided in request")
		logger.Errorf("%s: %s", errors.ErrInvalidRequestBody, "no jobs provided")
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.ErrInvalidRequestBody})
		return
	}
	logger.Debugf("POST [CreateJobData] Job: %v", tempJobs[0].JobID)

	var existingUser *types.UserDataEntity
	var err error

	// Track user lookup
	trackDBOp := metrics.TrackDBOperation("read", "users")
	existingUser, err = h.userRepository.GetByID(c.Request.Context(), strings.ToLower(tempJobs[0].UserAddress))
	trackDBOp(err)
	if err != nil {
		logger.Error("Error getting user", "address", tempJobs[0].UserAddress, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if existingUser == nil {
		logger.Info("User not found, creating new user", "address", tempJobs[0].UserAddress)

		// Create new user entity
		newUser := &types.UserDataEntity{
			UserAddress:   strings.ToLower(tempJobs[0].UserAddress),
			CreatedAt:     time.Now().UTC(),
			LastUpdatedAt: time.Now().UTC(),
		}

		// Track user creation
		trackDBOp = metrics.TrackDBOperation("create", "users")
		err = h.userRepository.Create(c.Request.Context(), newUser)
		trackDBOp(err)
		if err != nil {
			logger.Error("Error creating new user", "address", tempJobs[0].UserAddress, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		logger.Info("Created new user", "address", existingUser.UserAddress)
	}
	logger.Info("Processing job for existing user", "address", existingUser.UserAddress)

	createdJobs := types.CreateJobResponse{
		JobIDs:            make([]string, len(tempJobs)),
		TaskDefinitionIDs: make([]int, len(tempJobs)),
		TimeFrames:        make([]int64, len(tempJobs)),
	}

	for i := len(tempJobs) - 1; i >= 0; i-- {
		chainStatus := 1
		var linkJobID = ""

		if i == 0 {
			chainStatus = 0
		}
		if i < len(tempJobs)-1 {
			// Use the string JobID from the previous iteration
			linkJobID = createdJobs.JobIDs[i+1]
		}

		// Use JobID as string directly
		jobID := tempJobs[i].JobID

		jobData := &types.JobDataEntity{
			JobID:             jobID,
			JobTitle:          tempJobs[i].JobTitle,
			TaskDefinitionID:  tempJobs[i].TaskDefinitionID,
			UserAddress:       existingUser.UserAddress,
			LinkJobID:         linkJobID,
			ChainStatus:       chainStatus,
			TimeFrame:         tempJobs[i].TimeFrame,
			Recurring:         tempJobs[i].Recurring,
			Status:            "pending",
			JobCostPrediction: tempJobs[i].JobCostPrediction,
			JobCostActual:     "0",
			Timezone:          tempJobs[i].Timezone,
			IsImua:            tempJobs[i].IsImua,
			CreatedChainID:    tempJobs[i].CreatedChainID,
			TaskIDs:           []int64{},
			CreatedAt:         time.Now().UTC(),
			UpdatedAt:         time.Now().UTC(),
			LastExecutedAt:    time.Now().UTC(),
		}

		// Track job creation
		trackDBOp = metrics.TrackDBOperation("create", "jobs")
		err = h.jobRepository.Create(c.Request.Context(), jobData)
		trackDBOp(err)
		if err != nil {
			logger.Error("Error creating job", "jobID", jobID, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		expirationTime := time.Now().Add(time.Duration(tempJobs[i].TimeFrame) * time.Second)
		var scheduleConditionJobData types.ScheduleConditionJobData

		switch tempJobs[i].TaskDefinitionID {
		case 1, 2:
			// Time-based job
			var nextExecutionTimestamp time.Time
			nextExecutionTimestamp, err := parser.CalculateNextExecutionTime(time.Now(), tempJobs[i].ScheduleType, tempJobs[i].TimeInterval, tempJobs[i].CronExpression, tempJobs[i].SpecificSchedule)
			if err != nil {
				logger.Error("Error calculating next execution timestamp", "error", err)
				nextExecutionTimestamp = time.Now().Add(time.Duration(tempJobs[i].TimeInterval) * time.Second)
			}

			timeJobData := &types.TimeJobDataEntity{
				JobID:                     jobID,
				TaskDefinitionID:          tempJobs[i].TaskDefinitionID,
				ExpirationTime:            expirationTime,
				TimeInterval:              tempJobs[i].TimeInterval,
				ScheduleType:              tempJobs[i].ScheduleType,
				CronExpression:            tempJobs[i].CronExpression,
				SpecificSchedule:          tempJobs[i].SpecificSchedule,
				NextExecutionTimestamp:    nextExecutionTimestamp,
				TargetChainID:             tempJobs[i].TargetChainID,
				TargetContractAddress:     tempJobs[i].TargetContractAddress,
				TargetFunction:            tempJobs[i].TargetFunction,
				ABI:                       tempJobs[i].ABI,
				ArgType:                   tempJobs[i].ArgType,
				Arguments:                 tempJobs[i].Arguments,
				DynamicArgumentsScriptURL: tempJobs[i].DynamicArgumentsScriptUrl,
				IsCompleted:               false,
				LastExecutedAt:            time.Time{},
			}

			// Track time job creation
			trackDBOp = metrics.TrackDBOperation("create", "time_jobs")
			if err := h.timeJobRepository.Create(c.Request.Context(), timeJobData); err != nil {
				trackDBOp(err)
				logger.Error("Error inserting time job data", "jobID", jobID, "error", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				return
			}
			trackDBOp(nil)
			logger.Info("Successfully created time-based job",
				"jobID", jobID,
				"intervalSeconds", timeJobData.TimeInterval)

		case 3, 4:
			// Event-based job
			eventJobData := &types.EventJobDataEntity{
				JobID:                      jobID,
				TaskDefinitionID:           tempJobs[i].TaskDefinitionID,
				ExpirationTime:             expirationTime,
				Recurring:                  tempJobs[i].Recurring,
				TriggerChainID:             tempJobs[i].TriggerChainID,
				TriggerContractAddress:     tempJobs[i].TriggerContractAddress,
				TriggerEvent:               tempJobs[i].TriggerEvent,
				TriggerEventFilterParaName: tempJobs[i].TriggerEventFilterParaName,
				TriggerEventFilterValue:    tempJobs[i].TriggerEventFilterValue,
				TargetChainID:              tempJobs[i].TargetChainID,
				TargetContractAddress:      tempJobs[i].TargetContractAddress,
				TargetFunction:             tempJobs[i].TargetFunction,
				ABI:                        tempJobs[i].ABI,
				ArgType:                    tempJobs[i].ArgType,
				Arguments:                  tempJobs[i].Arguments,
				DynamicArgumentsScriptURL:  tempJobs[i].DynamicArgumentsScriptUrl,
				IsCompleted:                false,
				LastExecutedAt:             time.Time{},
			}

			if err := h.eventJobRepository.Create(c.Request.Context(), eventJobData); err != nil {
				logger.Error("Error inserting event job data", "jobID", jobID, "error", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				return
			}
			scheduleConditionJobData.JobID = jobID
			scheduleConditionJobData.TaskDefinitionID = tempJobs[i].TaskDefinitionID
			scheduleConditionJobData.LastExecutedAt = time.Time{}
			scheduleConditionJobData.TaskTargetData = types.TaskTargetData{
				JobID:                     jobID,
				TaskDefinitionID:          tempJobs[i].TaskDefinitionID,
				TargetChainID:             tempJobs[i].TargetChainID,
				TargetContractAddress:     tempJobs[i].TargetContractAddress,
				TargetFunction:            tempJobs[i].TargetFunction,
				ABI:                       tempJobs[i].ABI,
				ArgType:                   tempJobs[i].ArgType,
				Arguments:                 tempJobs[i].Arguments,
				DynamicArgumentsScriptUrl: tempJobs[i].DynamicArgumentsScriptUrl,
				IsImua:                    tempJobs[i].IsImua,
			}
			scheduleConditionJobData.EventWorkerData = types.EventWorkerData{
				JobID:                      jobID,
				ExpirationTime:             expirationTime,
				Recurring:                  tempJobs[i].Recurring,
				TriggerChainID:             tempJobs[i].TriggerChainID,
				TriggerContractAddress:     tempJobs[i].TriggerContractAddress,
				TriggerEvent:               tempJobs[i].TriggerEvent,
				TriggerEventFilterParaName: tempJobs[i].TriggerEventFilterParaName,
				TriggerEventFilterValue:    tempJobs[i].TriggerEventFilterValue,
			}
			logger.Info("Successfully created event-based job",
				"jobID", jobID,
				"event", eventJobData.TriggerEvent,
				"contract", eventJobData.TriggerContractAddress)

		case 5, 6:
			// Condition-based job
			conditionJobData := &types.ConditionJobDataEntity{
				JobID:                     jobID,
				TaskDefinitionID:          tempJobs[i].TaskDefinitionID,
				ExpirationTime:            expirationTime,
				Recurring:                 tempJobs[i].Recurring,
				ConditionType:             tempJobs[i].ConditionType,
				UpperLimit:                tempJobs[i].UpperLimit,
				LowerLimit:                tempJobs[i].LowerLimit,
				ValueSourceType:           tempJobs[i].ValueSourceType,
				ValueSourceURL:            tempJobs[i].ValueSourceUrl,
				TargetChainID:             tempJobs[i].TargetChainID,
				TargetContractAddress:     tempJobs[i].TargetContractAddress,
				TargetFunction:            tempJobs[i].TargetFunction,
				ABI:                       tempJobs[i].ABI,
				ArgType:                   tempJobs[i].ArgType,
				Arguments:                 tempJobs[i].Arguments,
				DynamicArgumentsScriptURL: tempJobs[i].DynamicArgumentsScriptUrl,
				IsCompleted:               false,
				SelectedKeyRoute:          tempJobs[i].SelectedKeyRoute,
				LastExecutedAt:            time.Time{},
			}

			if err := h.conditionJobRepository.Create(c.Request.Context(), conditionJobData); err != nil {
				logger.Error("Error inserting condition job data", "jobID", jobID, "error", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				return
			}
			scheduleConditionJobData.JobID = jobID
			scheduleConditionJobData.TaskDefinitionID = tempJobs[i].TaskDefinitionID
			scheduleConditionJobData.LastExecutedAt = time.Now()
			scheduleConditionJobData.TaskTargetData = types.TaskTargetData{
				JobID:                     jobID,
				TaskDefinitionID:          tempJobs[i].TaskDefinitionID,
				TargetChainID:             tempJobs[i].TargetChainID,
				TargetContractAddress:     tempJobs[i].TargetContractAddress,
				TargetFunction:            tempJobs[i].TargetFunction,
				ABI:                       tempJobs[i].ABI,
				ArgType:                   tempJobs[i].ArgType,
				Arguments:                 tempJobs[i].Arguments,
				DynamicArgumentsScriptUrl: tempJobs[i].DynamicArgumentsScriptUrl,
				IsImua:                    tempJobs[i].IsImua,
			}
			scheduleConditionJobData.ConditionWorkerData = types.ConditionWorkerData{
				JobID:            jobID,
				ExpirationTime:   expirationTime,
				Recurring:        tempJobs[i].Recurring,
				ConditionType:    tempJobs[i].ConditionType,
				UpperLimit:       tempJobs[i].UpperLimit,
				LowerLimit:       tempJobs[i].LowerLimit,
				ValueSourceType:  tempJobs[i].ValueSourceType,
				ValueSourceUrl:   tempJobs[i].ValueSourceUrl,
				SelectedKeyRoute: tempJobs[i].SelectedKeyRoute,
			}
			logger.Info("Successfully created condition-based job",
				"jobID", jobID,
				"conditionType", conditionJobData.ConditionType,
				"lowerLimit", conditionJobData.LowerLimit,
				"upperLimit", conditionJobData.UpperLimit)
		default:
			logger.Error("Invalid task definition ID", "taskDefID", tempJobs[i].TaskDefinitionID, "jobIndex", i)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task definition ID"})
			return
		}

		if tempJobs[i].TaskDefinitionID == 3 || tempJobs[i].TaskDefinitionID == 4 || tempJobs[i].TaskDefinitionID == 5 || tempJobs[i].TaskDefinitionID == 6 {
			success, err := h.notifyConditionScheduler(jobID, scheduleConditionJobData)
			if !success {
				logger.Error("Error notifying condition scheduler", "jobID", jobID, "error", err)
			} else {
				logger.Info("Successfully notified condition scheduler", "jobID", jobID)
			}
		}

		createdJobs.JobIDs[i] = jobID
		createdJobs.TaskDefinitionIDs[i] = tempJobs[i].TaskDefinitionID
		createdJobs.TimeFrames[i] = tempJobs[i].TimeFrame
	}

	// Update user's job_ids and total jobs
	allJobIDs := make([]string, 0, len(existingUser.JobIDs)+len(createdJobs.JobIDs))
	allJobIDs = append(allJobIDs, existingUser.JobIDs...) // Append existing JobIDs
	allJobIDs = append(allJobIDs, createdJobs.JobIDs...) // Append new JobIDs

	existingUser.JobIDs = allJobIDs
	existingUser.TotalJobs = int64(len(allJobIDs))
	existingUser.LastUpdatedAt = time.Now().UTC()

	trackDBOp = metrics.TrackDBOperation("update", "users")
	if err := h.userRepository.Update(c.Request.Context(), existingUser); err != nil {
		trackDBOp(err)
		logger.Error("Error updating user job IDs", "userAddress", existingUser.UserAddress, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	trackDBOp(nil)
	logger.Info("Successfully updated user with jobs", "userAddress", existingUser.UserAddress, "totalJobs", len(allJobIDs))

	// Track total operation duration
	trackDBOp = metrics.TrackDBOperation("create", "jobs")
	trackDBOp(nil)

	logger.Info("Successfully completed job creation",
		"userAddress", existingUser.UserAddress,
		"newJobsCount", len(tempJobs))
	c.JSON(http.StatusOK, createdJobs)
}
