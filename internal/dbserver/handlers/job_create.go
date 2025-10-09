package handlers

import (
	"context"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/parser"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (h *Handler) CreateJobData(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Infof("[CreateJobData] trace_id=%s - Creating job data", traceID)
	var tempJobs []types.CreateJobData
	if err := c.ShouldBindJSON(&tempJobs); err != nil {
		h.logger.Errorf("[CreateJobData] Error decoding request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	if len(tempJobs) == 0 {
		h.logger.Error("[CreateJobData] No jobs provided in request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No jobs provided",
			"code":  "EMPTY_REQUEST",
		})
		return
	}

	ctx := context.Background()
	var existingUser *types.UserDataEntity
	var err error

	// Track user lookup
	trackDBOp := metrics.TrackDBOperation("read", "users")
	existingUser, err = h.userRepository.GetByNonID(ctx, "user_address", strings.ToLower(tempJobs[0].UserAddress))
	trackDBOp(err)

	if err != nil {
		h.logger.Errorf("[CreateJobData] Error getting user for address %s: %v", tempJobs[0].UserAddress, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if existingUser == nil {
		h.logger.Infof("[CreateJobData] User not found for address %s, creating new user", tempJobs[0].UserAddress)

		// Create new user entity
		newUser := &types.UserDataEntity{
			UserID:        0, // Will be auto-generated
			UserAddress:   strings.ToLower(tempJobs[0].UserAddress),
			EmailID:       "",
			JobIDs:        []big.Int{},
			OpxConsumed:   *big.NewInt(0),
			TotalJobs:     0,
			TotalTasks:    0,
			CreatedAt:     time.Now().UTC(),
			LastUpdatedAt: time.Now().UTC(),
		}

		// Track user creation
		trackDBOp = metrics.TrackDBOperation("create", "users")
		err = h.userRepository.Create(ctx, newUser)
		trackDBOp(err)

		if err != nil {
			h.logger.Errorf("[CreateJobData] Error creating new user for address %s: %v", tempJobs[0].UserAddress, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// Fetch the created user to get the generated ID
		existingUser, err = h.userRepository.GetByNonID(ctx, "user_address", strings.ToLower(tempJobs[0].UserAddress))
		if err != nil || existingUser == nil {
			h.logger.Errorf("[CreateJobData] Error fetching created user: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		h.logger.Infof("[CreateJobData] Created new user with userID %d | Address: %s", existingUser.UserID, existingUser.UserAddress)
	}

	h.logger.Infof("[CreateJobData] existingUserID: %d", existingUser.UserID)

	createdJobs := types.CreateJobResponse{
		UserID:            existingUser.UserID,
		AccountBalance:    &existingUser.OpxConsumed,
		TokenBalance:      &existingUser.OpxConsumed, // Note: Using OpxConsumed as placeholder, update schema if needed
		JobIDs:            make([]*big.Int, len(tempJobs)),
		TaskDefinitionIDs: make([]int, len(tempJobs)),
		TimeFrames:        make([]int64, len(tempJobs)),
	}

	for i := len(tempJobs) - 1; i >= 0; i-- {
		chainStatus := 1
		var linkJobID *big.Int = nil

		if i == 0 {
			chainStatus = 0
		}
		if i < len(tempJobs)-1 {
			linkJobID = createdJobs.JobIDs[i+1]
		}

		jobID := new(big.Int)
		if _, ok := jobID.SetString(tempJobs[i].JobID, 10); !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job_id format"})
			return
		}

		linkJobIDBigInt := big.NewInt(0)
		if linkJobID != nil {
			linkJobIDBigInt = linkJobID
		}

		jobCostPrediction := big.NewInt(int64(tempJobs[i].JobCostPrediction))

		jobData := &types.JobDataEntity{
			JobID:             *jobID,
			JobTitle:          tempJobs[i].JobTitle,
			TaskDefinitionID:  tempJobs[i].TaskDefinitionID,
			UserID:            existingUser.UserID,
			LinkJobID:         *linkJobIDBigInt,
			ChainStatus:       chainStatus,
			TimeFrame:         tempJobs[i].TimeFrame,
			Recurring:         tempJobs[i].Recurring,
			Status:            "pending",
			JobCostPrediction: *jobCostPrediction,
			JobCostActual:     *big.NewInt(0),
			Timezone:          tempJobs[i].Timezone,
			IsImua:            tempJobs[i].IsImua,
			CreatedChainID:    tempJobs[i].CreatedChainID,
			TaskIDs:           []int64{},
			CreatedAt:         time.Now().UTC(),
			UpdatedAt:         time.Now().UTC(),
			LastExecutedAt:    time.Time{},
		}

		// Track job creation
		trackDBOp = metrics.TrackDBOperation("create", "jobs")
		err = h.jobRepository.Create(ctx, jobData)
		trackDBOp(err)

		if err != nil {
			h.logger.Errorf("[CreateJobData] Error creating job: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		createdJobs.JobIDs[i] = jobID
		expirationTime := time.Now().Add(time.Duration(tempJobs[i].TimeFrame) * time.Second)
		var scheduleConditionJobData types.ScheduleConditionJobData

		switch tempJobs[i].TaskDefinitionID {
		case 1, 2:
			// Time-based job

			var nextExecutionTimestamp time.Time
			nextExecutionTimestamp, err := parser.CalculateNextExecutionTime(time.Now(), "interval", tempJobs[i].TimeInterval, tempJobs[i].CronExpression, tempJobs[i].SpecificSchedule)
			if err != nil {
				h.logger.Errorf("[getNextExecutionTimestamp] Error calculating next execution timestamp: %v", err)
				nextExecutionTimestamp = time.Now().Add(time.Duration(tempJobs[i].TimeInterval) * time.Second)
			}

			timeJobData := &types.TimeJobDataEntity{
				JobID:                     *jobID,
				TaskDefinitionID:          tempJobs[i].TaskDefinitionID,
				ExpirationTime:            expirationTime,
				TimeInterval:              tempJobs[i].TimeInterval,
				ScheduleType:              "interval",
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
			if err := h.timeJobRepository.Create(ctx, timeJobData); err != nil {
				trackDBOp(err)
				h.logger.Errorf("[CreateJobData] Error inserting time job data for jobID %v: %v", jobID, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				return
			}
			trackDBOp(nil)
			h.logger.Infof("[CreateJobData] Successfully created time-based job %v with interval %d seconds",
				jobID, timeJobData.TimeInterval)

		case 3, 4:
			// Event-based job
			eventJobData := &types.EventJobDataEntity{
				JobID:                      *jobID,
				TaskDefinitionID:           tempJobs[i].TaskDefinitionID,
				ExpirationTime:             expirationTime,
				Recurring:                  tempJobs[i].Recurring,
				TriggerChainID:             tempJobs[i].TriggerChainID,
				TriggerContractAddress:     tempJobs[i].TriggerContractAddress,
				TriggerEvent:               tempJobs[i].TriggerEvent,
				TriggerEventFilterParaName: tempJobs[i].EventFilterParaName,
				TriggerEventFilterValue:    tempJobs[i].EventFilterValue,
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

			if err := h.eventJobRepository.Create(ctx, eventJobData); err != nil {
				h.logger.Errorf("[CreateJobData] Error inserting event job data for jobID %v: %v", jobID, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				return
			}
			scheduleConditionJobData.JobID = types.NewBigInt(jobID)
			scheduleConditionJobData.TaskDefinitionID = tempJobs[i].TaskDefinitionID
			scheduleConditionJobData.LastExecutedAt = time.Now()
			scheduleConditionJobData.TaskTargetData = types.TaskTargetData{
				JobID:                     types.NewBigInt(jobID),
				TaskDefinitionID:          tempJobs[i].TaskDefinitionID,
				TargetChainID:             tempJobs[i].TargetChainID,
				TargetContractAddress:     tempJobs[i].TargetContractAddress,
				TargetFunction:            tempJobs[i].TargetFunction,
				ABI:                       tempJobs[i].ABI,
				ArgType:                   tempJobs[i].ArgType,
				Arguments:                 tempJobs[i].Arguments,
				DynamicArgumentsScriptUrl: tempJobs[i].DynamicArgumentsScriptUrl,
			}
			scheduleConditionJobData.EventWorkerData = types.EventWorkerData{
				JobID:                  types.NewBigInt(jobID),
				ExpirationTime:         expirationTime,
				Recurring:              tempJobs[i].Recurring,
				TriggerChainID:         tempJobs[i].TriggerChainID,
				TriggerContractAddress: tempJobs[i].TriggerContractAddress,
				TriggerEvent:           tempJobs[i].TriggerEvent,
				EventFilterParaName:    tempJobs[i].EventFilterParaName,
				EventFilterValue:       tempJobs[i].EventFilterValue,
			}
			filterEnabled := eventJobData.TriggerEventFilterParaName != "" && eventJobData.TriggerEventFilterValue != ""
			h.logger.Infof("[CreateJobData] Successfully created event-based job %v for event %s on contract %s (filter_enabled=%t)",
				jobID, eventJobData.TriggerEvent, eventJobData.TriggerContractAddress, filterEnabled)

		case 5, 6:
			// Condition-based job
			conditionJobData := &types.ConditionJobDataEntity{
				JobID:                     *jobID,
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

			if err := h.conditionJobRepository.Create(ctx, conditionJobData); err != nil {
				h.logger.Errorf("[CreateJobData] Error inserting condition job data for jobID %v: %v", jobID, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				return
			}
			scheduleConditionJobData.JobID = types.NewBigInt(jobID)
			scheduleConditionJobData.TaskDefinitionID = tempJobs[i].TaskDefinitionID
			scheduleConditionJobData.LastExecutedAt = time.Now()
			scheduleConditionJobData.TaskTargetData = types.TaskTargetData{
				JobID:                     types.NewBigInt(jobID),
				TaskDefinitionID:          tempJobs[i].TaskDefinitionID,
				TargetChainID:             tempJobs[i].TargetChainID,
				TargetContractAddress:     tempJobs[i].TargetContractAddress,
				TargetFunction:            tempJobs[i].TargetFunction,
				ABI:                       tempJobs[i].ABI,
				ArgType:                   tempJobs[i].ArgType,
				Arguments:                 tempJobs[i].Arguments,
				DynamicArgumentsScriptUrl: tempJobs[i].DynamicArgumentsScriptUrl,
			}
			scheduleConditionJobData.ConditionWorkerData = types.ConditionWorkerData{
				JobID:            types.NewBigInt(jobID),
				ExpirationTime:   expirationTime,
				Recurring:        tempJobs[i].Recurring,
				ConditionType:    tempJobs[i].ConditionType,
				UpperLimit:       tempJobs[i].UpperLimit,
				LowerLimit:       tempJobs[i].LowerLimit,
				ValueSourceType:  tempJobs[i].ValueSourceType,
				ValueSourceUrl:   tempJobs[i].ValueSourceUrl,
				SelectedKeyRoute: tempJobs[i].SelectedKeyRoute,
			}
			h.logger.Infof("[CreateJobData] Successfully created condition-based job %d with condition type %s (limits: %f-%f)",
				jobID, conditionJobData.ConditionType, conditionJobData.LowerLimit, conditionJobData.UpperLimit)
		default:
			h.logger.Errorf("[CreateJobData] Invalid task definition ID %d for job %d", tempJobs[i].TaskDefinitionID, i)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task definition ID"})
			return
		}

		if tempJobs[i].TaskDefinitionID == 3 || tempJobs[i].TaskDefinitionID == 4 || tempJobs[i].TaskDefinitionID == 5 || tempJobs[i].TaskDefinitionID == 6 {
			success, err := h.notifyConditionScheduler(jobID, scheduleConditionJobData)
			if !success {
				h.logger.Errorf("[CreateJobData] Error notifying condition scheduler for jobID %d: %v", jobID, err)
			} else {
				h.logger.Infof("[CreateJobData] Successfully notified condition scheduler for jobID %d", jobID)
			}
		}

		createdJobs.JobIDs[i] = jobID
		createdJobs.TaskDefinitionIDs[i] = tempJobs[i].TaskDefinitionID
		createdJobs.TimeFrames[i] = tempJobs[i].TimeFrame
	}

	// Update user's job_ids and total jobs
	allJobIDs := make([]big.Int, 0, len(existingUser.JobIDs)+len(createdJobs.JobIDs))
	allJobIDs = append(allJobIDs, existingUser.JobIDs...)
	for _, jobID := range createdJobs.JobIDs {
		allJobIDs = append(allJobIDs, *jobID)
	}

	existingUser.JobIDs = allJobIDs
	existingUser.TotalJobs = int64(len(allJobIDs))
	existingUser.LastUpdatedAt = time.Now().UTC()

	trackDBOp = metrics.TrackDBOperation("update", "users")
	if err := h.userRepository.Update(ctx, existingUser); err != nil {
		trackDBOp(err)
		h.logger.Errorf("[CreateJobData] Error updating user job IDs for userID %d: %v", existingUser.UserID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	trackDBOp(nil)
	h.logger.Infof("[CreateJobData] Successfully updated user %d with %d total jobs", existingUser.UserID, len(allJobIDs))

	// Track total operation duration
	trackDBOp = metrics.TrackDBOperation("create", "jobs")
	trackDBOp(nil)

	c.JSON(http.StatusOK, createdJobs)
	h.logger.Infof("[CreateJobData] Successfully completed job creation for user %d with %d new jobs",
		existingUser.UserID, len(tempJobs))
}
