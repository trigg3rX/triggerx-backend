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
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

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

	ctx := c.Request.Context()
	var existingUser commonTypes.UserData
	var err error

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
	_, existingUser, err = h.userRepository.GetUserDataByAddress(strings.ToLower(tempJobs[0].UserAddress))
	trackDBOp(err)
	userSpan.End()

	if err != nil && err != gocql.ErrNotFound {
		h.logger.Error(ctx, "[CreateJobData] Error getting user ID for address", observability.String("user_address", tempJobs[0].UserAddress), observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if err == gocql.ErrNotFound {
		// Span: Create new user
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
		newUser.EtherBalance = commonTypes.NewBigInt(tempJobs[0].EtherBalance)
		newUser.TokenBalance = commonTypes.NewBigInt(tempJobs[0].TokenBalance)
		newUser.UserPoints = 0.0

		trackDBOp = metrics.TrackDBOperation("create", "users")
		existingUser, err = h.userRepository.CreateNewUser(&newUser)
		trackDBOp(err)
		createUserSpan.End()

		if err != nil {
			h.logger.Error(ctx, "[CreateJobData] Error creating new user for address", observability.String("user_address", tempJobs[0].UserAddress), observability.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
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
		jobID, err := h.jobRepository.CreateNewJob(jobData)
		trackDBOp(err)
		if err != nil {
			createJobSpan.RecordError(err)
			createJobSpan.SetStatus(codes.Error, "failed to create job")
		}
		createJobSpan.SetAttributes(attribute.String("job.id", jobID.String()))
		createJobSpan.End()

		if err != nil {
			h.logger.Error(jobCtx, "[CreateJobData] Error creating job", observability.String("job_id", jobID.String()), observability.Error(err))
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

			// Span: Create time job
			_, timeJobSpan := h.tracer.Start(jobCtx, "db.create_time_job",
				observability.WithSpanKind(trace.SpanKindClient),
				observability.WithAttributes(
					attribute.String("db.system", "cassandra"),
					attribute.String("db.operation", "insert"),
					attribute.String("db.collection", "time_jobs"),
					attribute.String("job.id", jobID.String()),
				),
			)
			trackDBOp = metrics.TrackDBOperation("create", "time_jobs")
			if err := h.timeJobRepository.CreateTimeJob(&timeJobData); err != nil {
				trackDBOp(err)
				timeJobSpan.RecordError(err)
				timeJobSpan.SetStatus(codes.Error, "failed to create time job")
				timeJobSpan.End()
				h.logger.Error(jobCtx, "[CreateJobData] Error inserting time job data for jobID", observability.Int64("job_id", jobID.Int64()), observability.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				return
			}
			trackDBOp(nil)
			timeJobSpan.End()

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

			// Span: Create event job
			_, eventJobSpan := h.tracer.Start(jobCtx, "db.create_event_job",
				observability.WithSpanKind(trace.SpanKindClient),
				observability.WithAttributes(
					attribute.String("db.system", "cassandra"),
					attribute.String("db.operation", "insert"),
					attribute.String("db.collection", "event_jobs"),
					attribute.String("job.id", jobID.String()),
				),
			)
			if err := h.eventJobRepository.CreateEventJob(&eventJobData); err != nil {
				eventJobSpan.RecordError(err)
				eventJobSpan.SetStatus(codes.Error, "failed to create event job")
				eventJobSpan.End()
				h.logger.Error(jobCtx, "[CreateJobData] Error inserting event job data for jobID", observability.Int64("job_id", jobID.Int64()), observability.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				return
			}
			eventJobSpan.End()
			scheduleConditionJobData.JobID = commonTypes.NewBigInt(jobID)
			scheduleConditionJobData.TaskDefinitionID = tempJobs[i].TaskDefinitionID
			scheduleConditionJobData.LastExecutedAt = time.Now()
			var eventJobOwnerAddress string
			if tempJobs[i].IsSafe {
				eventJobOwnerAddress = strings.ToLower(tempJobs[i].UserAddress)
			}
			scheduleConditionJobData.TaskTargetData = commonTypes.TaskTargetData{
				JobID:                     commonTypes.NewBigInt(jobID),
				JobOwnerAddress:           eventJobOwnerAddress,
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

			// Span: Create condition job
			_, conditionJobSpan := h.tracer.Start(jobCtx, "db.create_condition_job",
				observability.WithSpanKind(trace.SpanKindClient),
				observability.WithAttributes(
					attribute.String("db.system", "cassandra"),
					attribute.String("db.operation", "insert"),
					attribute.String("db.collection", "condition_jobs"),
					attribute.String("job.id", jobID.String()),
				),
			)
			if err := h.conditionJobRepository.CreateConditionJob(&conditionJobData); err != nil {
				conditionJobSpan.RecordError(err)
				conditionJobSpan.SetStatus(codes.Error, "failed to create condition job")
				conditionJobSpan.End()
				h.logger.Error(jobCtx, "[CreateJobData] Error inserting condition job data for jobID", observability.Int64("job_id", jobID.Int64()), observability.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				return
			}
			conditionJobSpan.End()
			scheduleConditionJobData.JobID = commonTypes.NewBigInt(jobID)
			scheduleConditionJobData.TaskDefinitionID = tempJobs[i].TaskDefinitionID
			scheduleConditionJobData.LastExecutedAt = time.Now()
			var condJobOwnerAddress string
			if tempJobs[i].IsSafe {
				condJobOwnerAddress = strings.ToLower(tempJobs[i].UserAddress)
			}
			scheduleConditionJobData.TaskTargetData = commonTypes.TaskTargetData{
				JobID:                     commonTypes.NewBigInt(jobID),
				JobOwnerAddress:           condJobOwnerAddress,
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

			// Span: Create custom job
			_, customJobSpan := h.tracer.Start(jobCtx, "db.create_custom_job",
				observability.WithSpanKind(trace.SpanKindClient),
				observability.WithAttributes(
					attribute.String("db.system", "cassandra"),
					attribute.String("db.operation", "insert"),
					attribute.String("db.collection", "custom_jobs"),
					attribute.String("job.id", jobID.String()),
				),
			)
			trackDBOp = metrics.TrackDBOperation("create", "custom_jobs")
			if err := h.customJobRepository.CreateCustomJob(&customJobData); err != nil {
				trackDBOp(err)
				customJobSpan.RecordError(err)
				customJobSpan.SetStatus(codes.Error, "failed to create custom job")
				customJobSpan.End()
				h.logger.Error(jobCtx, "[CreateJobData] Error inserting custom job data for jobID", observability.Int64("job_id", jobID.Int64()), observability.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				return
			}
			trackDBOp(nil)
			customJobSpan.End()

		default:
			h.logger.Error(c.Request.Context(), "[CreateJobData] Invalid task definition ID for job", observability.Int("task_definition_id", tempJobs[i].TaskDefinitionID), observability.Int("job_index", i))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task definition ID"})
			return
		}

		if tempJobs[i].TaskDefinitionID == 3 || tempJobs[i].TaskDefinitionID == 4 || tempJobs[i].TaskDefinitionID == 5 || tempJobs[i].TaskDefinitionID == 6 {
			success, err := h.notifyConditionScheduler(jobCtx, jobID, scheduleConditionJobData)
			if !success {
				h.logger.Error(jobCtx, "[CreateJobData] Error notifying condition scheduler for jobID", observability.Int64("job_id", jobID.Int64()), observability.Error(err))
			}
		}

		pointsToAdd := 10.0
		if tempJobs[i].Custom {
			pointsToAdd = 20.0
		}

		var currentPoints = existingUser.UserPoints
		newPoints := currentPoints + pointsToAdd
		// Span: Update user points
		_, updatePointsSpan := h.tracer.Start(jobCtx, "db.update_user_points",
			observability.WithSpanKind(trace.SpanKindClient),
			observability.WithAttributes(
				attribute.String("db.system", "cassandra"),
				attribute.String("db.operation", "update"),
				attribute.String("db.collection", "users"),
				attribute.Int64("user.id", existingUser.UserID),
			),
		)
		trackDBOp = metrics.TrackDBOperation("update", "users")
		if err := h.userRepository.UpdateUserTasksAndPoints(existingUser.UserID, 0, newPoints); err != nil {
			trackDBOp(err)
			updatePointsSpan.RecordError(err)
			updatePointsSpan.SetStatus(codes.Error, "failed to update user points")
			updatePointsSpan.End()
			h.logger.Error(jobCtx, "[CreateJobData] Error updating user points for userID", observability.Int64("user_id", existingUser.UserID), observability.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		trackDBOp(nil)
		updatePointsSpan.End()

		createdJobs.JobIDs[i] = commonTypes.NewBigInt(jobID)
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
			attribute.Int64("user.id", existingUser.UserID),
			attribute.Int("jobs.count", len(createdJobs.JobIDs)),
		),
	)
	allJobIDs := append(existingUser.JobIDs, createdJobs.JobIDs...)
	// Convert BigInt slice to big.Int slice for repository
	bigIntJobIDs := make([]*big.Int, len(allJobIDs))
	for i, jobID := range allJobIDs {
		bigIntJobIDs[i] = jobID.ToBigInt()
	}
	trackDBOp = metrics.TrackDBOperation("update", "users")
	if err := h.userRepository.UpdateUserJobIDs(existingUser.UserID, bigIntJobIDs); err != nil {
		trackDBOp(err)
		updateJobIDsSpan.RecordError(err)
		updateJobIDsSpan.SetStatus(codes.Error, "failed to update user job IDs")
		updateJobIDsSpan.End()
		h.logger.Error(ctx, "[CreateJobData] Error updating user job IDs for userID", observability.Int64("user_id", existingUser.UserID), observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	trackDBOp(nil)
	updateJobIDsSpan.End()

	c.JSON(http.StatusOK, createdJobs)
	h.logger.Info(ctx, "[CreateJobData] Successfully created jobs", observability.Int64("user_id", existingUser.UserID), observability.Int("jobs_count", len(tempJobs)))
}
