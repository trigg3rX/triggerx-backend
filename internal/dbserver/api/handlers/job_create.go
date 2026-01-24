package handlers

import (
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/config"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/parser"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
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

	ctx := c.Request.Context()
	var existingUser *types.UserDataDTO
	var err error
	userWasNewlyCreated := false

	// Calculate total points from all task definition IDs upfront
	var totalPointsToAdd float64
	for _, job := range tempJobs {
		pointsToAdd := 10.0
		if job.TaskDefinitionID == 7 || job.TaskDefinitionID == 8 || job.TaskDefinitionID == 9 {
			pointsToAdd = 20.0 // Agent jobs get more points
		}
		totalPointsToAdd += pointsToAdd
	}
	totalPointsStr := strconv.FormatFloat(totalPointsToAdd, 'f', -1, 64)

	// Span: Get or create user
	ctx, userSpan := h.tracer.Start(ctx, "db.get_user",
		observability.WithSpanKind(trace.SpanKindClient),
		observability.WithAttributes(
			attribute.String("db.system", "cassandra"),
			attribute.String("db.operation", "select"),
			attribute.String("db.collection", "users"),
			attribute.String("user.address", tempJobs[0].UserAddress),
		),
	)
	trackDBOp := metrics.TrackDBOperation("read", "users")
	existingUser, err = h.userRepository.GetUserDataByAddress(strings.ToLower(tempJobs[0].UserAddress))
	trackDBOp(err)
	userSpan.End()

	if err != nil && err != gocql.ErrNotFound {
		h.logger.Error(ctx, "[CreateJobData] Error getting user for address", observability.String("user_address", tempJobs[0].UserAddress), observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if err == gocql.ErrNotFound {
		userWasNewlyCreated = true
		// Span: Create new user with calculated points
		ctx, createUserSpan := h.tracer.Start(ctx, "db.create_user",
			observability.WithSpanKind(trace.SpanKindClient),
			observability.WithAttributes(
				attribute.String("db.system", "cassandra"),
				attribute.String("db.operation", "insert"),
				attribute.String("db.collection", "users"),
				attribute.String("user.address", tempJobs[0].UserAddress),
			),
		)

		var newUser types.CreateUserDataRequest
		newUser.UserAddress = strings.ToLower(tempJobs[0].UserAddress)
		newUser.EmailID = ""
		newUser.UserPoints = totalPointsStr // Set calculated points when creating user

		trackDBOp = metrics.TrackDBOperation("create", "users")
		err = h.userRepository.CreateNewUser(&newUser)
		if err != nil {
			trackDBOp(err)
			createUserSpan.RecordError(err)
			createUserSpan.SetStatus(codes.Error, "failed to create user")
			createUserSpan.End()
			h.logger.Error(ctx, "[CreateJobData] Error creating new user for address", observability.String("user_address", tempJobs[0].UserAddress), observability.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		trackDBOp(nil)
		createUserSpan.End()

		// Fetch the newly created user
		existingUser, err = h.userRepository.GetUserDataByAddress(strings.ToLower(tempJobs[0].UserAddress))
		if err != nil {
			h.logger.Error(ctx, "[CreateJobData] Error fetching newly created user", observability.String("user_address", tempJobs[0].UserAddress), observability.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
	}

	createdJobs := types.CreateJobResponse{
		JobIDs:            make([]string, len(tempJobs)),
		TaskDefinitionIDs: make([]int, len(tempJobs)),
		TimeFrames:        make([]int64, len(tempJobs)),
	}

	// Calculate final user points (existing + new points)
	var finalUserPoints string
	var currentPointsFloat float64
	if existingUser.UserPoints != "" {
		if parsed, err := strconv.ParseFloat(existingUser.UserPoints, 64); err == nil {
			currentPointsFloat = parsed
		}
	}
	finalPointsFloat := currentPointsFloat + totalPointsToAdd
	finalUserPoints = strconv.FormatFloat(finalPointsFloat, 'f', -1, 64)

	for i := len(tempJobs) - 1; i >= 0; i-- {
		chainStatus := 1
		var linkJobID = ""

		if i == 0 {
			chainStatus = 0
		}
		if i < len(tempJobs)-1 {
			linkJobID = createdJobs.JobIDs[i+1]
		}

		jobID := tempJobs[i].JobID
		if jobID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job_id format"})
			return
		}

		// Convert JobCostPrediction from float64 to string (Wei-based)
		jobCostPredictionStr := strconv.FormatFloat(tempJobs[i].JobCostPrediction, 'f', -1, 64)

		jobData := &types.JobDataEntity{
			JobID:             jobID,
			JobTitle:          tempJobs[i].JobTitle,
			TaskDefinitionID:  tempJobs[i].TaskDefinitionID,
			UserAddress:       strings.ToLower(tempJobs[0].UserAddress),
			LinkJobID:         linkJobID,
			ChainStatus:       chainStatus,
			JobType:           "frontend", // Default job type
			TimeFrame:         tempJobs[i].TimeFrame,
			Recurring:         tempJobs[i].Recurring,
			Status:            "pending",
			JobCostPrediction: jobCostPredictionStr,
			Timezone:          tempJobs[i].Timezone,
			IsImua:            tempJobs[i].IsImua,
			CreatedChainID:    tempJobs[i].CreatedChainID,
			SafeAddress:       "",
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		// Before creating job, validate IPFS code for dynamic jobs (TaskDefinitionID==2,4,6,7 & DynamicArgumentsScriptUrl or AgentScriptURL)
		ipfsUrl := ""
		if tempJobs[i].TaskDefinitionID == 7 {
			ipfsUrl = tempJobs[i].AgentScriptURL
		} else {
			ipfsUrl = tempJobs[i].DynamicArgumentsScriptUrl
		}

		if (tempJobs[i].TaskDefinitionID == 2 || tempJobs[i].TaskDefinitionID == 4 || tempJobs[i].TaskDefinitionID == 6 || tempJobs[i].TaskDefinitionID == 7) && ipfsUrl != "" {
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
				Language:         tempJobs[i].AgentScriptLanguage,
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
				}
			}

			// Set the safe address in job data
			jobData.SafeAddress = safeAddr
		}

		// Span: Create job
		jobCtx, createJobSpan := h.tracer.Start(ctx, "db.create_job",
			observability.WithSpanKind(trace.SpanKindClient),
			observability.WithAttributes(
				attribute.String("db.system", "cassandra"),
				attribute.String("db.operation", "insert"),
				attribute.String("db.collection", "jobs"),
				attribute.Int("task_definition_id", tempJobs[i].TaskDefinitionID),
			),
		)
		trackDBOp = metrics.TrackDBOperation("create", "jobs")
		err = h.jobRepository.CreateNewJob(jobData)
		trackDBOp(err)
		if err != nil {
			createJobSpan.RecordError(err)
			createJobSpan.SetStatus(codes.Error, "failed to create job")
			createJobSpan.End()
			h.logger.Error(jobCtx, "[CreateJobData] Error creating job", observability.String("job_id", jobID), observability.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		createJobSpan.SetAttributes(attribute.String("job.id", jobID))
		createJobSpan.End()

		createdJobs.JobIDs[i] = jobID
		expirationTime := time.Now().Add(time.Duration(tempJobs[i].TimeFrame) * time.Second)
		var scheduleConditionJobData types.ScheduleConditionJobData

		switch tempJobs[i].TaskDefinitionID {
		case 1, 2, 7:
			// Time-based job (TDI 1, 2) or Agent job (TDI 7)
			var nextExecutionTimestamp time.Time
			nextExecutionTimestamp, err := parser.CalculateNextExecutionTime(time.Now(), "interval", tempJobs[i].TimeInterval, tempJobs[i].CronExpression, tempJobs[i].SpecificSchedule)
			if err != nil {
				h.logger.Error(c.Request.Context(), "[getNextExecutionTimestamp] Error calculating next execution timestamp", observability.Error(err))
				nextExecutionTimestamp = time.Now().Add(time.Duration(tempJobs[i].TimeInterval) * time.Second)
			}

			timeJobData := &types.TimeJobDataEntity{
				JobID:                     jobID,
				TaskDefinitionID:          tempJobs[i].TaskDefinitionID,
				ScheduleType:              tempJobs[i].ScheduleType,
				TimeInterval:              tempJobs[i].TimeInterval,
				CronExpression:            tempJobs[i].CronExpression,
				SpecificSchedule:          tempJobs[i].SpecificSchedule,
				Timezone:                  tempJobs[i].Timezone,
				NextExecutionTimestamp:    nextExecutionTimestamp,
				TargetChainID:             tempJobs[i].TargetChainID,
				TargetContractAddress:     tempJobs[i].TargetContractAddress,
				TargetFunction:            tempJobs[i].TargetFunction,
				ABI:                       tempJobs[i].ABI,
				ArgType:                   tempJobs[i].ArgType,
				Arguments:                 tempJobs[i].Arguments,
				DynamicArgumentsScriptURL: tempJobs[i].DynamicArgumentsScriptUrl,
				// Agent job fields (TDI 7)
				AgentScriptURL:      tempJobs[i].AgentScriptURL,
				AgentScriptLanguage: tempJobs[i].AgentScriptLanguage,
				AgentScriptHash:     "", // Will be calculated if needed
				AgentTargetChainID:  tempJobs[i].AgentTargetChainID,
				MaxExecutionTime:    tempJobs[i].MaxExecutionTime,
				ChallengePeriod:     tempJobs[i].ChallengePeriod,
				IsActive:            true,
				LastExecutedAt:      time.Time{},
				ExpirationTime:      expirationTime,
			}

			// Span: Create time job
			_, timeJobSpan := h.tracer.Start(jobCtx, "db.create_time_job",
				observability.WithSpanKind(trace.SpanKindClient),
				observability.WithAttributes(
					attribute.String("db.system", "cassandra"),
					attribute.String("db.operation", "insert"),
					attribute.String("db.collection", "time_jobs"),
					attribute.String("job.id", jobID),
				),
			)
			trackDBOp = metrics.TrackDBOperation("create", "time_jobs")
			if err := h.timeJobRepository.CreateTimeJob(timeJobData); err != nil {
				trackDBOp(err)
				timeJobSpan.RecordError(err)
				timeJobSpan.SetStatus(codes.Error, "failed to create time job")
				timeJobSpan.End()
				h.logger.Error(jobCtx, "[CreateJobData] Error inserting time job data for jobID", observability.String("job_id", jobID), observability.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				return
			}
			trackDBOp(nil)
			timeJobSpan.End()

		case 3, 4, 8:
			// Event-based job (TDI 3, 4) or Agent job (TDI 8)
			eventJobData := &types.EventJobDataEntity{
				JobID:                     jobID,
				TaskDefinitionID:          tempJobs[i].TaskDefinitionID,
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
				DynamicArgumentsScriptURL: tempJobs[i].DynamicArgumentsScriptUrl,
				// Agent job fields (TDI 8)
				AgentScriptURL:      tempJobs[i].AgentScriptURL,
				AgentScriptLanguage: tempJobs[i].AgentScriptLanguage,
				AgentScriptHash:     "",
				AgentTargetChainID:  tempJobs[i].AgentTargetChainID,
				MaxExecutionTime:    tempJobs[i].MaxExecutionTime,
				ChallengePeriod:     tempJobs[i].ChallengePeriod,
				IsActive:            true,
				LastExecutedAt:      time.Time{},
				ExpirationTime:      expirationTime,
			}

			// Span: Create event job
			_, eventJobSpan := h.tracer.Start(jobCtx, "db.create_event_job",
				observability.WithSpanKind(trace.SpanKindClient),
				observability.WithAttributes(
					attribute.String("db.system", "cassandra"),
					attribute.String("db.operation", "insert"),
					attribute.String("db.collection", "event_jobs"),
					attribute.String("job.id", jobID),
				),
			)
			if err := h.eventJobRepository.CreateEventJob(eventJobData); err != nil {
				eventJobSpan.RecordError(err)
				eventJobSpan.SetStatus(codes.Error, "failed to create event job")
				eventJobSpan.End()
				h.logger.Error(jobCtx, "[CreateJobData] Error inserting event job data for jobID", observability.String("job_id", jobID), observability.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				return
			}
			eventJobSpan.End()
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
			}
			scheduleConditionJobData.EventWorkerData = types.EventWorkerData{
				JobID:                  jobID,
				ExpirationTime:         expirationTime,
				Recurring:              tempJobs[i].Recurring,
				TriggerChainID:         tempJobs[i].TriggerChainID,
				TriggerContractAddress: tempJobs[i].TriggerContractAddress,
				TriggerEvent:           tempJobs[i].TriggerEvent,
				EventFilterParaName:    tempJobs[i].EventFilterParaName,
				EventFilterValue:       tempJobs[i].EventFilterValue,
			}

		case 5, 6, 9:
			// Condition-based job (TDI 5, 6) or Agent job (TDI 9)
			conditionJobData := &types.ConditionJobDataEntity{
				JobID:                     jobID,
				TaskDefinitionID:          tempJobs[i].TaskDefinitionID,
				Recurring:                 tempJobs[i].Recurring,
				ConditionType:             tempJobs[i].ConditionType,
				UpperLimit:                tempJobs[i].UpperLimit,
				LowerLimit:                tempJobs[i].LowerLimit,
				ValueSourceType:           tempJobs[i].ValueSourceType,
				ValueSourceURL:            tempJobs[i].ValueSourceUrl,
				SelectedKeyRoute:          tempJobs[i].SelectedKeyRoute,
				TargetChainID:             tempJobs[i].TargetChainID,
				TargetContractAddress:     tempJobs[i].TargetContractAddress,
				TargetFunction:            tempJobs[i].TargetFunction,
				ABI:                       tempJobs[i].ABI,
				ArgType:                   tempJobs[i].ArgType,
				Arguments:                 tempJobs[i].Arguments,
				DynamicArgumentsScriptURL: tempJobs[i].DynamicArgumentsScriptUrl,
				// Agent job fields (TDI 9)
				AgentScriptURL:      tempJobs[i].AgentScriptURL,
				AgentScriptLanguage: tempJobs[i].AgentScriptLanguage,
				AgentScriptHash:     "",
				AgentTargetChainID:  tempJobs[i].AgentTargetChainID,
				MaxExecutionTime:    tempJobs[i].MaxExecutionTime,
				ChallengePeriod:     tempJobs[i].ChallengePeriod,
				IsActive:            true,
				LastExecutedAt:      time.Time{},
				ExpirationTime:      expirationTime,
			}

			// Span: Create condition job
			_, conditionJobSpan := h.tracer.Start(jobCtx, "db.create_condition_job",
				observability.WithSpanKind(trace.SpanKindClient),
				observability.WithAttributes(
					attribute.String("db.system", "cassandra"),
					attribute.String("db.operation", "insert"),
					attribute.String("db.collection", "condition_jobs"),
					attribute.String("job.id", jobID),
				),
			)
			if err := h.conditionJobRepository.CreateConditionJob(conditionJobData); err != nil {
				conditionJobSpan.RecordError(err)
				conditionJobSpan.SetStatus(codes.Error, "failed to create condition job")
				conditionJobSpan.End()
				h.logger.Error(jobCtx, "[CreateJobData] Error inserting condition job data for jobID", observability.String("job_id", jobID), observability.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				return
			}
			conditionJobSpan.End()
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

		default:
			h.logger.Error(c.Request.Context(), "[CreateJobData] Invalid task definition ID for job", observability.Int("task_definition_id", tempJobs[i].TaskDefinitionID), observability.Int("job_index", i))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task definition ID"})
			return
		}

		if tempJobs[i].TaskDefinitionID == 3 || tempJobs[i].TaskDefinitionID == 4 || tempJobs[i].TaskDefinitionID == 5 || tempJobs[i].TaskDefinitionID == 6 {
			success, err := h.notifyConditionScheduler(jobCtx, jobID, scheduleConditionJobData)
			if !success {
				h.logger.Error(jobCtx, "[CreateJobData] Error notifying condition scheduler for jobID", observability.String("job_id", jobID), observability.Error(err))
			}
		}

		createdJobs.JobIDs[i] = jobID
		createdJobs.TaskDefinitionIDs[i] = tempJobs[i].TaskDefinitionID
		createdJobs.TimeFrames[i] = tempJobs[i].TimeFrame
	}

	// Span: Update user's job_ids
	_, updateJobIDsSpan := h.tracer.Start(ctx, "db.update_user_job_ids",
		observability.WithSpanKind(trace.SpanKindClient),
		observability.WithAttributes(
			attribute.String("db.system", "cassandra"),
			attribute.String("db.operation", "update"),
			attribute.String("db.collection", "users"),
			attribute.String("user.address", existingUser.UserAddress),
			attribute.Int("jobs.count", len(createdJobs.JobIDs)),
		),
	)
	allJobIDs := append(existingUser.JobIDs, createdJobs.JobIDs...)
	trackDBOp = metrics.TrackDBOperation("update", "users")
	if err := h.userRepository.UpdateUserJobIDs(existingUser.UserAddress, allJobIDs); err != nil {
		trackDBOp(err)
		updateJobIDsSpan.RecordError(err)
		updateJobIDsSpan.SetStatus(codes.Error, "failed to update user job IDs")
		updateJobIDsSpan.End()
		h.logger.Error(ctx, "[CreateJobData] Error updating user job IDs for userAddress", observability.String("user_address", existingUser.UserAddress), observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	trackDBOp(nil)
	updateJobIDsSpan.End()

	// Update user points once after all jobs are created (only if user already existed before job creation)
	if !userWasNewlyCreated {
		// Span: Update user points
		_, updatePointsSpan := h.tracer.Start(ctx, "db.update_user_points",
			observability.WithSpanKind(trace.SpanKindClient),
			observability.WithAttributes(
				attribute.String("db.system", "cassandra"),
				attribute.String("db.operation", "update"),
				attribute.String("db.collection", "users"),
				attribute.String("user.address", existingUser.UserAddress),
			),
		)
		trackDBOp = metrics.TrackDBOperation("update", "users")
		if err := h.userRepository.UpdateUserPoints(existingUser.UserAddress, finalUserPoints); err != nil {
			trackDBOp(err)
			updatePointsSpan.RecordError(err)
			updatePointsSpan.SetStatus(codes.Error, "failed to update user points")
			updatePointsSpan.End()
			h.logger.Error(ctx, "[CreateJobData] Error updating user points for userAddress", observability.String("user_address", existingUser.UserAddress), observability.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		trackDBOp(nil)
		updatePointsSpan.End()
	}

	// Set the final user points in the response
	createdJobs.UserPoints = finalUserPoints

	c.JSON(http.StatusOK, createdJobs)
	h.logger.Info(ctx, "[CreateJobData] Successfully created jobs", observability.String("user_address", existingUser.UserAddress), observability.Int("jobs_count", len(tempJobs)))
}
