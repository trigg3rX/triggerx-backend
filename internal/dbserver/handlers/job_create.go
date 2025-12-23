package handlers

import (
	"io"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/config"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/parser"
	commonTypes "github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (h *Handler) CreateJobData(c *gin.Context) {
	var tempJobs []types.CreateJobData
	if err := c.ShouldBindJSON(&tempJobs); err != nil {
		h.logger.Error(c.Request.Context(), "[CreateJobData] Error decoding request body", observability.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	if len(tempJobs) == 0 {
		h.logger.Error(c.Request.Context(), "[CreateJobData] No jobs provided in request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No jobs provided",
			"code":  "EMPTY_REQUEST",
		})
		return
	}

	var existingUserID int64
	var existingUser commonTypes.UserData
	var err error

	// Track user lookup
	trackDBOp := metrics.TrackDBOperation("read", "users")
	existingUserID, existingUser, err = h.userRepository.GetUserDataByAddress(strings.ToLower(tempJobs[0].UserAddress))
	trackDBOp(err)

	if err != nil && err != gocql.ErrNotFound {
		h.logger.Error(c.Request.Context(), "[CreateJobData] Error getting user ID for address", observability.String("user_address", tempJobs[0].UserAddress), observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	h.logger.Debug(c.Request.Context(), "[CreateJobData] existingUserID", observability.Int64("existing_user_id", existingUserID))

	if err == gocql.ErrNotFound {
		var newUser types.CreateUserDataRequest
		newUser.UserAddress = strings.ToLower(tempJobs[0].UserAddress)
		newUser.EtherBalance = commonTypes.NewBigInt(tempJobs[0].EtherBalance)
		newUser.TokenBalance = commonTypes.NewBigInt(tempJobs[0].TokenBalance)
		newUser.UserPoints = 0.0

		// Track user creation
		trackDBOp = metrics.TrackDBOperation("create", "users")
		existingUser, err = h.userRepository.CreateNewUser(&newUser)
		trackDBOp(err)

		if err != nil {
			h.logger.Error(c.Request.Context(), "[CreateJobData] Error creating new user for address", observability.String("user_address", tempJobs[0].UserAddress), observability.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		h.logger.Info(c.Request.Context(), "[CreateJobData] Created new user with userID", observability.Int64("user_id", existingUser.UserID), observability.String("user_address", existingUser.UserAddress))
	}

	createdJobs := types.CreateJobResponse{
		UserID:            existingUser.UserID,
		AccountBalance:    existingUser.EtherBalance.ToBigInt(),
		TokenBalance:      existingUser.TokenBalance.ToBigInt(),
		JobIDs:            make([]*commonTypes.BigInt, len(tempJobs)),
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
			linkJobID = createdJobs.JobIDs[i+1].ToBigInt()
		}

		jobID := new(big.Int)
		if _, ok := jobID.SetString(tempJobs[i].JobID, 10); !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job_id format"})
			return
		}

		jobData := &commonTypes.JobData{
			JobID:             commonTypes.FromBigInt(jobID),
			JobTitle:          tempJobs[i].JobTitle,
			TaskDefinitionID:  tempJobs[i].TaskDefinitionID,
			UserID:            existingUser.UserID,
			LinkJobID:         commonTypes.FromBigInt(linkJobID),
			ChainStatus:       chainStatus,
			Custom:            tempJobs[i].Custom,
			TimeFrame:         tempJobs[i].TimeFrame,
			Recurring:         tempJobs[i].Recurring,
			Status:            "pending",
			JobCostPrediction: tempJobs[i].JobCostPrediction,
			Timezone:          tempJobs[i].Timezone,
			IsImua:            tempJobs[i].IsImua,
			CreatedChainID:    tempJobs[i].CreatedChainID,
			SafeAddress:       "",
		}

		// Before creating job, validate IPFS code for dynamic jobs (TaskDefinitionID==2,4,6,7 & DynamicArgumentsScriptUrl)
		if (tempJobs[i].TaskDefinitionID == 2 || tempJobs[i].TaskDefinitionID == 4 || tempJobs[i].TaskDefinitionID == 6 || tempJobs[i].TaskDefinitionID == 7) && tempJobs[i].DynamicArgumentsScriptUrl != "" {
			// Parse CID or gateway URL if needed
			ipfsUrl := tempJobs[i].DynamicArgumentsScriptUrl
			ctx := c.Request.Context()
			resp, err := h.httpClient.Get(ctx, ipfsUrl)
			if err != nil {
				h.logger.Error(c.Request.Context(), "[CreateJobData] Failed to download file", observability.Error(err))
				c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to download file: " + err.Error()})
				return
			}
			defer func() {
				if err := resp.Body.Close(); err != nil {
					h.logger.Error(c.Request.Context(), "[CreateJobData] Error closing response body", observability.Error(err))
				}
			}()

			if resp.StatusCode != http.StatusOK {
				h.logger.Error(c.Request.Context(), "[CreateJobData] Unexpected status code", observability.Int("status_code", resp.StatusCode))
				c.JSON(http.StatusBadRequest, gin.H{"error": "Unexpected status code: " + strconv.Itoa(resp.StatusCode)})
				return
			}

			content, err := io.ReadAll(resp.Body)
			if err != nil {
				h.logger.Error(c.Request.Context(), "[CreateJobData] Failed to read response body", observability.Error(err))
				c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read response body: " + err.Error()})
				return
			}
			ipfsCode := string(content)

			valReq := ValidateCodeRequest{
				Code:             ipfsCode,
				Language:         tempJobs[i].Language,
				SelectedSafe:     tempJobs[i].SafeAddress,
				TargetFunction:   tempJobs[i].TargetFunction,
				TaskDefinitionID: tempJobs[i].TaskDefinitionID,
				IsSafe:           tempJobs[i].IsSafe,
			}
			valResp, _ := h.ValidateCodeInternal(ctx, valReq, ipfsUrl, config.GetAlchemyAPIKey())
			if !valResp.Executable || !valResp.SafeMatch {
				errMsg := "IPFS code validation failed: "
				if valResp.Error != "" {
					errMsg += valResp.Error
				} else if !valResp.SafeMatch {
					errMsg += "SafeAddress does not match code output."
				} else {
					errMsg += "Code not executable."
				}
				// Emit a log for observability of why the request is not proceeding
				// outputPreview := valResp.Output
				// if len(outputPreview) > 200 {
				// 	outputPreview = outputPreview[:200] + "..."
				// }
				// h.logger.Info(c.Request.Context(), "[CreateJobData] IPFS code validation failed", observability.String("user_address", tempJobs[i].UserAddress), observability.String("job_id", tempJobs[i].JobID), observability.Int("task_definition_id", tempJobs[i].TaskDefinitionID), observability.Bool("is_safe", valReq.IsSafe), observability.String("selected_safe", valReq.SelectedSafe), observability.Bool("executable", valResp.Executable), observability.Bool("safe_match", valResp.SafeMatch), observability.String("error", valResp.Error), observability.String("output_preview", outputPreview))
				// Return 200 with structured validation failure so clients can display message without treating as transport error
				c.JSON(http.StatusOK, gin.H{
					"status":                "validation_failed",
					"message":               errMsg,
					"validation_executable": valResp.Executable,
					"validation_safe_match": valResp.SafeMatch,
					"validation_output":     valResp.Output,
				})
				return
			}
		}

		// Handle safe address if IsSafe is true
		if tempJobs[i].IsSafe {
			if tempJobs[i].SafeAddress == "" {
				h.logger.Error(c.Request.Context(), "[CreateJobData] IsSafe is true but SafeAddress is empty for job", observability.String("job_id", tempJobs[i].JobID))
				c.JSON(http.StatusBadRequest, gin.H{"error": "SafeAddress is required when IsSafe is true"})
				return
			}

			// Lowercase the safe address for consistency
			safeAddr := strings.ToLower(tempJobs[i].SafeAddress)

			// Check if safe address already exists for this user
			exists, err := h.safeAddressRepository.CheckSafeAddressExists(strings.ToLower(tempJobs[i].UserAddress), safeAddr)
			if err != nil {
				h.logger.Error(c.Request.Context(), "[CreateJobData] Error checking safe address existence", observability.Error(err))
			} else if !exists {
				// Create safe address entry if it doesn't exist
				if err := h.safeAddressRepository.CreateSafeAddress(strings.ToLower(tempJobs[i].UserAddress), safeAddr, tempJobs[i].SafeName); err != nil {
					h.logger.Error(c.Request.Context(), "[CreateJobData] Error creating safe address", observability.Error(err))
				} else {
					h.logger.Info(c.Request.Context(), "[CreateJobData] Created safe address for user", observability.String("safe_address", safeAddr), observability.String("user_address", tempJobs[i].UserAddress))
				}
			}

			// Set the safe address in job data
			jobData.SafeAddress = safeAddr
		}

		// Track job creation
		trackDBOp = metrics.TrackDBOperation("create", "jobs")
		jobID, err := h.jobRepository.CreateNewJob(jobData)
		trackDBOp(err)

		if err != nil {
			h.logger.Error(c.Request.Context(), "[CreateJobData] Error creating job", observability.String("job_id", jobID.String()), observability.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		createdJobs.JobIDs[i] = commonTypes.NewBigInt(jobID)
		expirationTime := time.Now().Add(time.Duration(tempJobs[i].TimeFrame) * time.Second)
		var scheduleConditionJobData commonTypes.ScheduleConditionJobData

		switch tempJobs[i].TaskDefinitionID {
		case 1, 2:
			// Time-based job

			var nextExecutionTimestamp time.Time
			nextExecutionTimestamp, err := parser.CalculateNextExecutionTime(time.Now(), "interval", tempJobs[i].TimeInterval, tempJobs[i].CronExpression, tempJobs[i].SpecificSchedule)
			if err != nil {
				h.logger.Error(c.Request.Context(), "[getNextExecutionTimestamp] Error calculating next execution timestamp", observability.Error(err))
				nextExecutionTimestamp = time.Now().Add(time.Duration(tempJobs[i].TimeInterval) * time.Second)
			}

			timeJobData := commonTypes.TimeJobData{
				JobID:            commonTypes.NewBigInt(jobID),
				TaskDefinitionID: tempJobs[i].TaskDefinitionID,
				ExpirationTime:   expirationTime,
				// Recurring:                 tempJobs[i].Recurring,
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
				DynamicArgumentsScriptUrl: tempJobs[i].DynamicArgumentsScriptUrl,
				IsCompleted:               false,
				IsActive:                  true,
			}

			// Track time job creation
			trackDBOp = metrics.TrackDBOperation("create", "time_jobs")
			if err := h.timeJobRepository.CreateTimeJob(&timeJobData); err != nil {
				trackDBOp(err)
				h.logger.Error(c.Request.Context(), "[CreateJobData] Error inserting time job data for jobID", observability.Int64("job_id", jobID.Int64()), observability.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				return
			}
			trackDBOp(nil)
			h.logger.Info(c.Request.Context(), "[CreateJobData] Successfully created time-based job with interval", observability.Int64("job_id", jobID.Int64()), observability.Int64("interval", timeJobData.TimeInterval))

		case 3, 4:
			// Event-based job
			eventJobData := commonTypes.EventJobData{
				JobID:                     commonTypes.NewBigInt(jobID),
				TaskDefinitionID:          tempJobs[i].TaskDefinitionID,
				ExpirationTime:            expirationTime,
				Recurring:                 tempJobs[i].Recurring,
				TriggerChainID:            tempJobs[i].TriggerChainID,
				TriggerContractAddress:    tempJobs[i].TriggerContractAddress,
				TriggerEvent:              tempJobs[i].TriggerEvent,
				EventFilterParaName:       tempJobs[i].EventFilterParaName,
				EventFilterValue:          tempJobs[i].EventFilterValue,
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
				h.logger.Error(c.Request.Context(), "[CreateJobData] Error inserting event job data for jobID", observability.Int64("job_id", jobID.Int64()), observability.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				return
			}
			scheduleConditionJobData.JobID = commonTypes.NewBigInt(jobID)
			scheduleConditionJobData.TaskDefinitionID = tempJobs[i].TaskDefinitionID
			scheduleConditionJobData.LastExecutedAt = time.Now()
			scheduleConditionJobData.TaskTargetData = commonTypes.TaskTargetData{
				JobID:                     commonTypes.NewBigInt(jobID),
				TaskDefinitionID:          tempJobs[i].TaskDefinitionID,
				TargetChainID:             tempJobs[i].TargetChainID,
				TargetContractAddress:     tempJobs[i].TargetContractAddress,
				TargetFunction:            tempJobs[i].TargetFunction,
				ABI:                       tempJobs[i].ABI,
				ArgType:                   tempJobs[i].ArgType,
				Arguments:                 tempJobs[i].Arguments,
				DynamicArgumentsScriptUrl: tempJobs[i].DynamicArgumentsScriptUrl,
			}
			scheduleConditionJobData.EventWorkerData = commonTypes.EventWorkerData{
				JobID:                  commonTypes.NewBigInt(jobID),
				ExpirationTime:         expirationTime,
				Recurring:              tempJobs[i].Recurring,
				TriggerChainID:         tempJobs[i].TriggerChainID,
				TriggerContractAddress: tempJobs[i].TriggerContractAddress,
				TriggerEvent:           tempJobs[i].TriggerEvent,
				EventFilterParaName:    tempJobs[i].EventFilterParaName,
				EventFilterValue:       tempJobs[i].EventFilterValue,
			}
			filterEnabled := eventJobData.EventFilterParaName != "" && eventJobData.EventFilterValue != ""
			h.logger.Info(c.Request.Context(), "[CreateJobData] Successfully created event-based job for event on contract (filter_enabled)", observability.Int64("job_id", jobID.Int64()), observability.String("trigger_event", eventJobData.TriggerEvent), observability.String("trigger_contract_address", eventJobData.TriggerContractAddress), observability.Bool("filter_enabled", filterEnabled))

		case 5, 6:
			// Condition-based job
			conditionJobData := commonTypes.ConditionJobData{
				JobID:                     commonTypes.NewBigInt(jobID),
				TaskDefinitionID:          tempJobs[i].TaskDefinitionID,
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
				SelectedKeyRoute:          tempJobs[i].SelectedKeyRoute,
			}

			if err := h.conditionJobRepository.CreateConditionJob(&conditionJobData); err != nil {
				h.logger.Error(c.Request.Context(), "[CreateJobData] Error inserting condition job data for jobID", observability.Int64("job_id", jobID.Int64()), observability.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				return
			}
			scheduleConditionJobData.JobID = commonTypes.NewBigInt(jobID)
			scheduleConditionJobData.TaskDefinitionID = tempJobs[i].TaskDefinitionID
			scheduleConditionJobData.LastExecutedAt = time.Now()
			scheduleConditionJobData.TaskTargetData = commonTypes.TaskTargetData{
				JobID:                     commonTypes.NewBigInt(jobID),
				TaskDefinitionID:          tempJobs[i].TaskDefinitionID,
				TargetChainID:             tempJobs[i].TargetChainID,
				TargetContractAddress:     tempJobs[i].TargetContractAddress,
				TargetFunction:            tempJobs[i].TargetFunction,
				ABI:                       tempJobs[i].ABI,
				ArgType:                   tempJobs[i].ArgType,
				Arguments:                 tempJobs[i].Arguments,
				DynamicArgumentsScriptUrl: tempJobs[i].DynamicArgumentsScriptUrl,
			}
			scheduleConditionJobData.ConditionWorkerData = commonTypes.ConditionWorkerData{
				JobID:            commonTypes.NewBigInt(jobID),
				ExpirationTime:   expirationTime,
				Recurring:        tempJobs[i].Recurring,
				ConditionType:    tempJobs[i].ConditionType,
				UpperLimit:       tempJobs[i].UpperLimit,
				LowerLimit:       tempJobs[i].LowerLimit,
				ValueSourceType:  tempJobs[i].ValueSourceType,
				ValueSourceUrl:   tempJobs[i].ValueSourceUrl,
				SelectedKeyRoute: tempJobs[i].SelectedKeyRoute,
			}
			h.logger.Info(c.Request.Context(), "[CreateJobData] Successfully created condition-based job with condition type (limits)", observability.Int64("job_id", jobID.Int64()), observability.String("condition_type", conditionJobData.ConditionType), observability.Float64("lower_limit", conditionJobData.LowerLimit), observability.Float64("upper_limit", conditionJobData.UpperLimit))

		case 7:
			// Custom script job (TaskDefinitionID = 7)
			var nextExecutionTime time.Time
			nextExecutionTime, err := parser.CalculateNextExecutionTime(time.Now(), "interval", tempJobs[i].TimeInterval, "", "")
			if err != nil {
				h.logger.Error(c.Request.Context(), "[CreateJobData] Error calculating next execution time for custom job", observability.Error(err))
				nextExecutionTime = time.Now().Add(time.Duration(tempJobs[i].TimeInterval) * time.Second)
			}

			customJobData := commonTypes.CustomJobData{
				JobID:             commonTypes.NewBigInt(jobID),
				TargetChainID:     tempJobs[i].TargetChainID,
				TaskDefinitionID:  7,
				ExpirationTime:    expirationTime,
				Recurring:         tempJobs[i].Recurring,
				CustomScriptUrl:   tempJobs[i].DynamicArgumentsScriptUrl,
				TimeInterval:      tempJobs[i].TimeInterval,
				ScriptLanguage:    tempJobs[i].Language,
				NextExecutionTime: nextExecutionTime,
				LastExecutedAt:    time.Now(),
				IsCompleted:       false,
				IsActive:          true,
			}

			// Track custom job creation
			trackDBOp = metrics.TrackDBOperation("create", "custom_jobs")
			if err := h.customJobRepository.CreateCustomJob(&customJobData); err != nil {
				trackDBOp(err)
				h.logger.Error(c.Request.Context(), "[CreateJobData] Error inserting custom job data for jobID", observability.Int64("job_id", jobID.Int64()), observability.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				return
			}
			trackDBOp(nil)
			h.logger.Info(c.Request.Context(), "[CreateJobData] Successfully created custom script job with interval, language", observability.Int64("job_id", jobID.Int64()), observability.Int64("interval", customJobData.TimeInterval), observability.String("language", customJobData.ScriptLanguage))

		default:
			h.logger.Error(c.Request.Context(), "[CreateJobData] Invalid task definition ID for job", observability.Int("task_definition_id", tempJobs[i].TaskDefinitionID), observability.Int("job_index", i))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task definition ID"})
			return
		}

		if tempJobs[i].TaskDefinitionID == 3 || tempJobs[i].TaskDefinitionID == 4 || tempJobs[i].TaskDefinitionID == 5 || tempJobs[i].TaskDefinitionID == 6 {
			success, err := h.notifyConditionScheduler(c.Request.Context(), jobID, scheduleConditionJobData)
			if !success {
				h.logger.Error(c.Request.Context(), "[CreateJobData] Error notifying condition scheduler for jobID", observability.Int64("job_id", jobID.Int64()), observability.Error(err))
			} else {
				h.logger.Debug(c.Request.Context(), "[CreateJobData] Successfully notified condition scheduler for jobID", observability.Int64("job_id", jobID.Int64()))
			}
		}

		pointsToAdd := 10.0
		if tempJobs[i].Custom {
			pointsToAdd = 20.0
		}

		var currentPoints = existingUser.UserPoints
		newPoints := currentPoints + pointsToAdd
		trackDBOp = metrics.TrackDBOperation("update", "users")
		if err := h.userRepository.UpdateUserTasksAndPoints(existingUser.UserID, 0, newPoints); err != nil {
			trackDBOp(err)
			h.logger.Error(c.Request.Context(), "[CreateJobData] Error updating user points for userID", observability.Int64("user_id", existingUser.UserID), observability.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		trackDBOp(nil)

		createdJobs.JobIDs[i] = commonTypes.NewBigInt(jobID)
		createdJobs.TaskDefinitionIDs[i] = tempJobs[i].TaskDefinitionID
		createdJobs.TimeFrames[i] = tempJobs[i].TimeFrame
	}

	// Update user's job_ids
	allJobIDs := append(existingUser.JobIDs, createdJobs.JobIDs...)
	// Convert BigInt slice to big.Int slice for repository
	bigIntJobIDs := make([]*big.Int, len(allJobIDs))
	for i, jobID := range allJobIDs {
		bigIntJobIDs[i] = jobID.ToBigInt()
	}
	trackDBOp = metrics.TrackDBOperation("update", "users")
	if err := h.userRepository.UpdateUserJobIDs(existingUser.UserID, bigIntJobIDs); err != nil {
		trackDBOp(err)
		h.logger.Error(c.Request.Context(), "[CreateJobData] Error updating user job IDs for userID", observability.Int64("user_id", existingUser.UserID), observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	trackDBOp(nil)
	h.logger.Debug(c.Request.Context(), "[CreateJobData] Successfully updated user with total jobs", observability.Int64("user_id", existingUser.UserID), observability.Int("total_jobs", len(allJobIDs)))

	// Track total operation duration
	trackDBOp = metrics.TrackDBOperation("create", "jobs")
	trackDBOp(nil)

	c.JSON(http.StatusOK, createdJobs)
	h.logger.Debug(c.Request.Context(), "[CreateJobData] Successfully completed job creation for user with new jobs", observability.Int64("user_id", existingUser.UserID), observability.Int("new_jobs", len(tempJobs)))
}
