package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
	"github.com/trigg3rX/triggerx-backend/pkg/parser"
)

func (h *Handler) CreateJobData(c *gin.Context) {
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

	var existingUserID int64
	var existingUser types.UserData
	var err error

	// Track user lookup
	trackDBOp := metrics.TrackDBOperation("read", "users")
	existingUserID, existingUser, err = h.userRepository.GetUserDataByAddress(strings.ToLower(tempJobs[0].UserAddress))
	trackDBOp(err)

	if err != nil && err != gocql.ErrNotFound {
		h.logger.Errorf("[CreateJobData] Error getting user ID for address %s: %v", tempJobs[0].UserAddress, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database error while retrieving user",
			"code":  "DB_ERROR",
		})
		return
	}

	h.logger.Infof("[CreateJobData] existingUserID: %d", existingUserID)

	if err == gocql.ErrNotFound {
		var newUser types.CreateUserDataRequest
		newUser.UserAddress = strings.ToLower(tempJobs[0].UserAddress)
		newUser.EtherBalance = tempJobs[0].EtherBalance
		newUser.TokenBalance = tempJobs[0].TokenBalance
		newUser.UserPoints = 0.0

		// Track user creation
		trackDBOp = metrics.TrackDBOperation("create", "users")
		existingUser, err = h.userRepository.CreateNewUser(&newUser)
		trackDBOp(err)

		if err != nil {
			h.logger.Errorf("[CreateJobData] Error creating new user for address %s: %v", tempJobs[0].UserAddress, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create new user",
				"code":  "USER_CREATION_ERROR",
			})
			return
		}

		h.logger.Infof("[CreateJobData] Created new user with userID %d | Address: %s", existingUserID, tempJobs[0].UserAddress)
	}

	createdJobs := types.CreateJobResponse{
		UserID:            existingUserID,
		AccountBalance:    existingUser.EtherBalance,
		TokenBalance:      existingUser.TokenBalance,
		JobIDs:            make([]int64, len(tempJobs)),
		TaskDefinitionIDs: make([]int, len(tempJobs)),
		TimeFrames:        make([]int64, len(tempJobs)),
	}

	for i := len(tempJobs) - 1; i >= 0; i-- {
		chainStatus := 1
		var linkJobID int64 = -1

		if i == 0 {
			chainStatus = 0
		}
		if i < len(tempJobs)-1 {
			linkJobID = createdJobs.JobIDs[i+1]
		}

		jobData := &types.JobData{
			JobTitle:          tempJobs[i].JobTitle,
			TaskDefinitionID:  tempJobs[i].TaskDefinitionID,
			UserID:            existingUserID,
			LinkJobID:         linkJobID,
			ChainStatus:       chainStatus,
			Custom:            tempJobs[i].Custom,
			TimeFrame:         tempJobs[i].TimeFrame,
			Recurring:         tempJobs[i].Recurring,
			Status:            "pending",
			JobCostPrediction: tempJobs[i].JobCostPrediction,
			Timezone:          tempJobs[i].Timezone,
		}

		// Track job creation
		trackDBOp = metrics.TrackDBOperation("create", "jobs")
		jobID, err := h.jobRepository.CreateNewJob(jobData)
		trackDBOp(err)

		if err != nil {
			h.logger.Errorf("[CreateJobData] Error creating job: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create job",
				"code":  "JOB_CREATION_ERROR",
			})
			return
		}

		createdJobs.JobIDs[i] = jobID
		expirationTime := time.Now().Add(time.Duration(tempJobs[i].TimeFrame) * time.Second)

		switch tempJobs[i].TaskDefinitionID {
		case 1, 2:
			// Time-based job

			var nextExecutionTimestamp time.Time
			nextExecutionTimestamp, err := parser.CalculateNextExecutionTime(time.Now(), "interval", tempJobs[i].TimeInterval, tempJobs[i].CronExpression, tempJobs[i].SpecificSchedule)
			if err != nil {
				h.logger.Errorf("[getNextExecutionTimestamp] Error calculating next execution timestamp: %v", err)
				nextExecutionTimestamp = time.Now().Add(time.Duration(tempJobs[i].TimeInterval) * time.Second)
			}

			timeJobData := types.TimeJobData{
				JobID:                     jobID,
				ExpirationTime:            expirationTime,
				// Recurring:                 tempJobs[i].Recurring,
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
				DynamicArgumentsScriptUrl: tempJobs[i].DynamicArgumentsScriptUrl,
				IsCompleted:               false,
				IsActive:                  true,
			}

			// Track time job creation
			trackDBOp = metrics.TrackDBOperation("create", "time_jobs")
			if err := h.timeJobRepository.CreateTimeJob(&timeJobData); err != nil {
				trackDBOp(err)
				h.logger.Errorf("[CreateJobData] Error inserting time job data for jobID %d: %v", jobID, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting time job data: " + err.Error()})
				return
			}
			trackDBOp(nil)
			h.logger.Infof("[CreateJobData] Successfully created time-based job %d with interval %d seconds",
				jobID, timeJobData.TimeInterval)

		case 3, 4:
			// Event-based job
			eventJobData := types.EventJobData{
				JobID:                     jobID,
				ExpirationTime:            expirationTime,
				Recurring:                 tempJobs[i].Recurring,
				TriggerChainID:            tempJobs[i].TriggerChainID,
				TriggerContractAddress:    tempJobs[i].TriggerContractAddress,
				TriggerEvent:              tempJobs[i].TriggerEvent,
				TargetChainID:             tempJobs[i].TargetChainID,
				TargetContractAddress:     tempJobs[i].TargetContractAddress,
				TargetFunction:            tempJobs[i].TargetFunction,
				ABI:                       tempJobs[i].ABI,
				ArgType:                   tempJobs[i].ArgType,
				Arguments:                 tempJobs[i].Arguments,
				DynamicArgumentsScriptUrl: tempJobs[i].DynamicArgumentsScriptUrl,
				IsCompleted:               false,
				IsActive:                  true,
			}

			if err := h.eventJobRepository.CreateEventJob(&eventJobData); err != nil {
				h.logger.Errorf("[CreateJobData] Error inserting event job data for jobID %d: %v", jobID, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting event job data: " + err.Error()})
				return
			}
			success, err := h.notifyEventScheduler(jobID, eventJobData)
			if !success {
				h.logger.Errorf("[CreateJobData] Error notifying event scheduler for jobID %d: %v", jobID, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error notifying event scheduler: " + err.Error()})
				return
			}
			h.logger.Infof("[CreateJobData] Successfully created event-based job %d for event %s on contract %s",
				jobID, eventJobData.TriggerEvent, eventJobData.TriggerContractAddress)

		case 5, 6:
			// Condition-based job
			conditionJobData := types.ConditionJobData{
				JobID:                     jobID,
				ExpirationTime:            expirationTime,
				Recurring:                 tempJobs[i].Recurring,
				ConditionType:             tempJobs[i].ConditionType,
				UpperLimit:                tempJobs[i].UpperLimit,
				LowerLimit:                tempJobs[i].LowerLimit,
				ValueSourceType:           tempJobs[i].ValueSourceType,
				ValueSourceUrl:            tempJobs[i].ValueSourceUrl,
				TargetChainID:             tempJobs[i].TargetChainID,
				TargetContractAddress:     tempJobs[i].TargetContractAddress,
				TargetFunction:            tempJobs[i].TargetFunction,
				ABI:                       tempJobs[i].ABI,
				ArgType:                   tempJobs[i].ArgType,
				Arguments:                 tempJobs[i].Arguments,
				DynamicArgumentsScriptUrl: tempJobs[i].DynamicArgumentsScriptUrl,
				IsCompleted:               false,
				IsActive:                  true,
			}

			if err := h.conditionJobRepository.CreateConditionJob(&conditionJobData); err != nil {
				h.logger.Errorf("[CreateJobData] Error inserting condition job data for jobID %d: %v", jobID, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting condition job data: " + err.Error()})
				return
			}
			success, err := h.notifyConditionScheduler(jobID, conditionJobData)
			if !success {
				h.logger.Errorf("[CreateJobData] Error notifying condition scheduler for jobID %d: %v", jobID, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error notifying condition scheduler: " + err.Error()})
				return
			}
			h.logger.Infof("[CreateJobData] Successfully created condition-based job %d with condition type %s (limits: %f-%f)",
				jobID, conditionJobData.ConditionType, conditionJobData.LowerLimit, conditionJobData.UpperLimit)
		default:
			h.logger.Errorf("[CreateJobData] Invalid task definition ID %d for job %d", tempJobs[i].TaskDefinitionID, i)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task definition ID"})
			return
		}

		pointsToAdd := 10.0
		if tempJobs[i].Custom {
			pointsToAdd = 20.0
		}

		var currentPoints = existingUser.UserPoints
		newPoints := currentPoints + pointsToAdd
		trackDBOp = metrics.TrackDBOperation("update", "users")
		if err := h.userRepository.UpdateUserTasksAndPoints(existingUserID, 0, newPoints); err != nil {
			trackDBOp(err)
			h.logger.Errorf("[CreateJobData] Error updating user points for userID %d: %v", existingUserID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating user points: " + err.Error()})
			return
		}
		trackDBOp(nil)

		createdJobs.JobIDs[i] = jobID
		createdJobs.TaskDefinitionIDs[i] = tempJobs[i].TaskDefinitionID
		createdJobs.TimeFrames[i] = tempJobs[i].TimeFrame
	}

	// Update user's job_ids
	allJobIDs := append(existingUser.JobIDs, createdJobs.JobIDs...)
	trackDBOp = metrics.TrackDBOperation("update", "users")
	if err := h.userRepository.UpdateUserJobIDs(existingUserID, allJobIDs); err != nil {
		trackDBOp(err)
		h.logger.Errorf("[CreateJobData] Error updating user job IDs for userID %d: %v", existingUserID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating user job IDs: " + err.Error()})
		return
	}
	trackDBOp(nil)
	h.logger.Infof("[CreateJobData] Successfully updated user %d with %d total jobs", existingUserID, len(allJobIDs))

	// Track total operation duration
	trackDBOp = metrics.TrackDBOperation("create", "jobs")
	trackDBOp(nil)

	c.JSON(http.StatusOK, createdJobs)
	h.logger.Infof("[CreateJobData] Successfully completed job creation for user %d with %d new jobs",
		existingUserID, len(tempJobs))
}
