package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
)

func (h *Handler) CreateJobData(c *gin.Context) {
	var tempJobs []types.CreateJobData
	if err := c.ShouldBindJSON(&tempJobs); err != nil {
		h.logger.Errorf("[CreateJobData] Error decoding request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(tempJobs) == 0 {
		h.logger.Error("[CreateJobData] No jobs provided in request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "No jobs provided"})
		return
	}

	var existingUserID int64
	var existingUser types.UserData
	var err error

	existingUserID, existingUser, err = h.userRepository.GetUserDataByAddress(strings.ToLower(tempJobs[0].UserAddress))
	if err != nil && err != gocql.ErrNotFound {
		h.logger.Errorf("[CreateJobData] Error getting user ID for address %s: %v", tempJobs[0].UserAddress, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting user ID: " + err.Error()})
		return
	}

	h.logger.Infof("[CreateJobData] existingUserID: %d", existingUserID)

	if err == gocql.ErrNotFound {
		var newUser types.CreateUserDataRequest
		newUser.UserAddress = strings.ToLower(tempJobs[0].UserAddress)
		newUser.EtherBalance = tempJobs[0].EtherBalance
		newUser.TokenBalance = tempJobs[0].TokenBalance
		newUser.UserPoints = 0.0

		existingUser, err = h.userRepository.CreateNewUser(&newUser)
		if err != nil {
			h.logger.Errorf("[CreateJobData] Error creating new user for address %s: %v", tempJobs[0].UserAddress, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating new user: " + err.Error()})
			return
		}

		h.logger.Infof("[CreateJobData] Created new user with userID %d | Address: %s", existingUser.UserID, tempJobs[0].UserAddress)
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

		jobID, err := h.jobRepository.CreateNewJob(jobData)
		if err != nil {
			h.logger.Errorf("[CreateJobData] Error creating job: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating job: " + err.Error()})
			return
		}

		createdJobs.JobIDs[i] = jobID

		switch {
		case tempJobs[i].TaskDefinitionID == 1 || tempJobs[i].TaskDefinitionID == 2:
			// Time-based job
			timeJobData := types.TimeJobData{
				JobID:                     jobID,
				TimeFrame:                 tempJobs[i].TimeFrame,
				Recurring:                 tempJobs[i].Recurring,
				TimeInterval:              tempJobs[i].TimeInterval,
				ScheduleType:              tempJobs[i].ScheduleType,
				CronExpression:            tempJobs[i].CronExpression,
				SpecificSchedule:          tempJobs[i].SpecificSchedule,
				NextExecutionTimestamp:    time.Now().Add(time.Duration(tempJobs[i].TimeInterval) * time.Second),
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

			if err := h.timeJobRepository.CreateTimeJob(&timeJobData); err != nil {
				h.logger.Errorf("[CreateJobData] Error inserting time job data for jobID %d: %v", jobID, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting time job data: " + err.Error()})
				return
			}
			h.logger.Infof("[CreateJobData] Successfully created time-based job %d with interval %d seconds",
				jobID, timeJobData.TimeInterval)

		case tempJobs[i].TaskDefinitionID == 3 || tempJobs[i].TaskDefinitionID == 4:
			// Event-based job
			eventJobData := types.EventJobData{
				JobID:                     jobID,
				TimeFrame:                 tempJobs[i].TimeFrame,
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
			h.notifyEventScheduler(jobID, eventJobData)
			h.logger.Infof("[CreateJobData] Successfully created event-based job %d for event %s on contract %s",
				jobID, eventJobData.TriggerEvent, eventJobData.TriggerContractAddress)

		case tempJobs[i].TaskDefinitionID == 5 || tempJobs[i].TaskDefinitionID == 6:
			// Condition-based job
			conditionJobData := types.ConditionJobData{
				JobID:                     jobID,
				TimeFrame:                 tempJobs[i].TimeFrame,
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
			h.notifyConditionScheduler(jobID, conditionJobData)
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
		if err := h.userRepository.UpdateUserTasksAndPoints(existingUserID, 0, newPoints); err != nil {
			h.logger.Errorf("[CreateJobData] Error updating user points for userID %d: %v", existingUserID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating user points: " + err.Error()})
		}

		createdJobs.JobIDs[i] = jobID
		createdJobs.TaskDefinitionIDs[i] = tempJobs[i].TaskDefinitionID
		createdJobs.TimeFrames[i] = tempJobs[i].TimeFrame
	}

	// Update user's job_ids
	allJobIDs := append(existingUser.JobIDs, createdJobs.JobIDs...)
	if err := h.userRepository.UpdateUserJobIDs(existingUserID, allJobIDs); err != nil {
		h.logger.Errorf("[CreateJobData] Error updating user job IDs for userID %d: %v", existingUserID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating user job IDs: " + err.Error()})
		return
	}
	h.logger.Infof("[CreateJobData] Successfully updated user %d with %d total jobs", existingUserID, len(allJobIDs))

	c.JSON(http.StatusOK, createdJobs)
	h.logger.Infof("[CreateJobData] Successfully completed job creation for user %d with %d new jobs",
		existingUserID, len(tempJobs))
}

// notifyEventScheduler sends a notification to the event scheduler
func (h *Handler) notifyEventScheduler(jobID int64, job types.EventJobData) {
	success, err := h.SendDataToEventScheduler("/api/v1/job/schedule", job)
	if err != nil {
		h.logger.Errorf("[NotifyEventScheduler] Failed to notify event scheduler for job %d: %v", jobID, err)
	} else if success {
		h.logger.Infof("[NotifyEventScheduler] Successfully notified event scheduler for job %d", jobID)
	}
}

// notifyConditionScheduler sends a notification to the condition scheduler
func (h *Handler) notifyConditionScheduler(jobID int64, job types.ConditionJobData) {
	success, err := h.SendDataToConditionScheduler("/api/v1/job/schedule", job)
	if err != nil {
		h.logger.Errorf("[NotifyConditionScheduler] Failed to notify condition scheduler for job %d: %v", jobID, err)
	} else if success {
		h.logger.Infof("[NotifyConditionScheduler] Successfully notified condition scheduler for job %d", jobID)
	}
}

func (h *Handler) UpdateJobDataFromUser(c *gin.Context) {
	var updateData types.UpdateJobDataFromUserRequest
	if err := c.ShouldBindJSON(&updateData); err != nil {
		h.logger.Errorf("[UpdateJobData] Error decoding request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.jobRepository.UpdateJobFromUserInDB(&updateData)
	if err != nil {
		h.logger.Errorf("[UpdateJobData] Error updating job data for jobID %d: %v", updateData.JobID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating job data: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Job updated successfully",
		"job_id":     updateData.JobID,
		"updated_at": time.Now().UTC(),
	})
}

func (h *Handler) UpdateJobStatus(c *gin.Context) {
	jobID := c.Param("job_id")
	status := c.Param("status")

	// Validate status
	validStatuses := map[string]bool{
		"pending":  true,
		"in-queue": true,
		"running":  true,
	}

	if !validStatuses[status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status. Must be one of: pending, in-queue, running"})
		return
	}
	// Convert jobID string to int64
	jobIDInt, err := strconv.ParseInt(jobID, 10, 64)
	if err != nil {
		h.logger.Errorf("[UpdateJobStatus] Error converting job ID to int64: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	// Update the job status
	if err := h.jobRepository.UpdateJobStatus(jobIDInt, status); err != nil {
		h.logger.Errorf("[UpdateJobStatus] Error updating job status: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Job status updated successfully",
		"job_id":     jobID,
		"status":     status,
		"updated_at": time.Now().UTC(),
	})
}

func (h *Handler) UpdateJobLastExecutedAt(c *gin.Context) {
	var updateData types.UpdateJobLastExecutedAtRequest
	if err := c.ShouldBindJSON(&updateData); err != nil {
		h.logger.Errorf("[UpdateJobLastExecutedAt] Error decoding request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update main job_data table
	if err := h.jobRepository.UpdateJobLastExecutedAt(updateData.JobID, updateData.TaskIDs, updateData.JobCostActual, updateData.LastExecutedAt); err != nil {
		h.logger.Errorf("[UpdateJobLastExecutedAt] Error updating job data for jobID %d: %v", updateData.JobID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating job data: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":          "Last executed time updated successfully",
		"job_id":           updateData.JobID,
		"last_executed_at": updateData.LastExecutedAt,
		"updated_at":       time.Now().UTC(),
	})
}

func (h *Handler) GetJobsByUserAddress(c *gin.Context) {
	userAddress := c.Param("user_address")
	if userAddress == "" {
		h.logger.Error("[GetJobsByUserAddress] No user address provided")
		c.JSON(http.StatusBadRequest, gin.H{"error": "No user address provided"})
		return
	}

	userID, jobIDs, err := h.userRepository.GetUserJobIDsByAddress(strings.ToLower(userAddress))
	if err != nil {
		h.logger.Errorf("[GetJobsByUserAddress] Error getting user data for address %s: %v", userAddress, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting user data: " + err.Error()})
		return
	}

	if len(jobIDs) == 0 {
		c.JSON(http.StatusOK, []types.JobResponse{})
		return
	}

	var jobs []types.JobResponse
	for _, jobID := range jobIDs {
		// Get basic job data
		jobData, err := h.jobRepository.GetJobByID(jobID)
		if err != nil {
			h.logger.Errorf("[GetJobsByUserAddress] Error getting job data for jobID %d: %v", jobID, err)
			continue
		}

		jobResponse := types.JobResponse{JobData: *jobData}

		// Check task_definition_id to determine job type
		switch {
		case jobData.TaskDefinitionID == 1 || jobData.TaskDefinitionID == 2:
			// Time-based job
			timeJobData, err := h.timeJobRepository.GetTimeJobByJobID(jobID)
			if err != nil {
				h.logger.Errorf("[GetJobsByUserAddress] Error getting time job data for jobID %d: %v", jobID, err)
				continue
			}
			jobResponse.TimeJobData = &timeJobData

		case jobData.TaskDefinitionID == 3 || jobData.TaskDefinitionID == 4:
			// Event-based job
			eventJobData, err := h.eventJobRepository.GetEventJobByJobID(jobID)
			if err != nil {
				h.logger.Errorf("[GetJobsByUserAddress] Error getting event job data for jobID %d: %v", jobID, err)
				continue
			}
			jobResponse.EventJobData = &eventJobData

		case jobData.TaskDefinitionID == 5 || jobData.TaskDefinitionID == 6:
			// Condition-based job
			conditionJobData, err := h.conditionJobRepository.GetConditionJobByJobID(jobID)
			if err != nil {
				h.logger.Errorf("[GetJobsByUserAddress] Error getting condition job data for jobID %d: %v", jobID, err)
				continue
			}
			jobResponse.ConditionJobData = &conditionJobData

		default:
			// No specific job data if task_definition_id is not recognized
		}

		jobs = append(jobs, jobResponse)
	}

	h.logger.Infof("[GetJobsByUserAddress] Successfully retrieved %d jobs for user %d", len(jobs), userID)

	c.JSON(http.StatusOK, jobs)
}

func (h *Handler) DeleteJobData(c *gin.Context) {
	jobID := c.Param("id")
	if jobID == "" {
		h.logger.Error("[DeleteJobData] No job ID provided")
		c.JSON(http.StatusBadRequest, gin.H{"error": "No job ID provided"})
		return
	}

	jobIDInt, err := strconv.ParseInt(jobID, 10, 64)
	if err != nil {
		h.logger.Errorf("[DeleteJobData] Invalid job ID format: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID format"})
		return
	}

	taskDefinitionID, err := h.jobRepository.GetTaskDefinitionIDByJobID(jobIDInt)
	if err != nil {
		h.logger.Errorf("[DeleteJobData] Error getting job data for jobID %s: %v", jobID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting job data: " + err.Error()})
		return
	}

	err = h.jobRepository.UpdateJobStatus(jobIDInt, "deleted")
	if err != nil {
		h.logger.Errorf("[DeleteJobData] Error updating job status for jobID %s: %v", jobID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating job status: " + err.Error()})
		return
	}

	switch taskDefinitionID {
	case 1, 2:
		err = h.timeJobRepository.UpdateTimeJobStatus(jobIDInt, false)
		if err != nil {
			h.logger.Errorf("[DeleteJobData] Error updating time job status for jobID %s: %v", jobID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating time job status: " + err.Error()})
			return
		}
	case 3, 4:
		err = h.eventJobRepository.UpdateEventJobStatus(jobIDInt, false)
		if err != nil {
			h.logger.Errorf("[DeleteJobData] Error updating event job status for jobID %s: %v", jobID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating event job status: " + err.Error()})
			return
		}
		h.notifySchedulerForJobDeletion(jobID, taskDefinitionID)
	case 5, 6:
		err = h.conditionJobRepository.UpdateConditionJobStatus(jobIDInt, false)
		if err != nil {
			h.logger.Errorf("[DeleteJobData] Error updating condition job status for jobID %s: %v", jobID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating condition job status: " + err.Error()})
			return
		}
		h.notifySchedulerForJobDeletion(jobID, taskDefinitionID)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Job deleted successfully"})
}

// notifySchedulerForJobDeletion notifies the appropriate scheduler about job deletion
func (h *Handler) notifySchedulerForJobDeletion(jobIDStr string, taskDefinitionID int) {
	switch {
	case taskDefinitionID == 1 || taskDefinitionID == 2:
		// Time-based jobs - no notification needed, time scheduler polls the database
		h.logger.Infof("[NotifySchedulerDeletion] Time-based job %s deleted, no notification needed (polling-based)", jobIDStr)

	case taskDefinitionID == 3 || taskDefinitionID == 4:
		// Event-based jobs - notify event scheduler
		success, err := h.SendPauseToEventScheduler(fmt.Sprintf("/api/v1/job/%s", jobIDStr))
		if err != nil {
			h.logger.Errorf("[NotifySchedulerDeletion] Failed to notify event scheduler about job %s deletion: %v", jobIDStr, err)
		} else if success {
			h.logger.Infof("[NotifySchedulerDeletion] Successfully notified event scheduler about job %s deletion", jobIDStr)
		}

	case taskDefinitionID == 5 || taskDefinitionID == 6:
		// Condition-based jobs - notify condition scheduler
		success, err := h.SendPauseToConditionScheduler(fmt.Sprintf("/api/v1/job/%s", jobIDStr))
		if err != nil {
			h.logger.Errorf("[NotifySchedulerDeletion] Failed to notify condition scheduler about job %s deletion: %v", jobIDStr, err)
		} else if success {
			h.logger.Infof("[NotifySchedulerDeletion] Successfully notified condition scheduler about job %s deletion", jobIDStr)
		}

	default:
		if taskDefinitionID != 0 { // Only warn if we actually found a task definition ID
			h.logger.Warnf("[NotifySchedulerDeletion] Unknown task definition ID %d for deleted job %s", taskDefinitionID, jobIDStr)
		}
	}
}
