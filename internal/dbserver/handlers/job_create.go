package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
	"github.com/trigg3rX/triggerx-backend/pkg/observability/tracing"
	"github.com/trigg3rX/triggerx-backend/pkg/parser"
	commonTypes "github.com/trigg3rX/triggerx-backend/pkg/types"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

func (h *Handler) CreateJobData(c *gin.Context) {
	// Start main business operation span
	tracer := otel.Tracer("triggerx-dbserver")
	ctx := c.Request.Context()
	ctx, span := tracer.Start(ctx, "job.create_batch")
	defer span.End()

	var tempJobs []types.CreateJobData
	if err := c.ShouldBindJSON(&tempJobs); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error.type", "invalid_request"))
		h.logger.Errorf("[CreateJobData] Error decoding request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	if len(tempJobs) == 0 {
		span.SetAttributes(attribute.String("error.type", "empty_request"))
		h.logger.Error("[CreateJobData] No jobs provided in request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No jobs provided",
			"code":  "EMPTY_REQUEST",
		})
		return
	}

	// Add business context to main span
	span.SetAttributes(
		attribute.String(tracing.TriggerXAttributes.UserAddress, strings.ToLower(tempJobs[0].UserAddress)),
		attribute.Int("job.batch_size", len(tempJobs)),
		attribute.String("operation.type", "create_job_batch"),
	)

	// Use trace-aware logging
	h.logger.InfoWithTrace(ctx, "Starting job creation batch",
		"user_address", strings.ToLower(tempJobs[0].UserAddress),
		"batch_size", len(tempJobs),
		"operation", "create_job_batch",
	)

	var existingUserID int64
	var existingUser types.UserData
	var err error

	// Track user lookup with both metrics and tracing
	trackDBOp := metrics.TrackDBOperation("read", "users")
	dbTracer := tracing.NewDatabaseTracer("triggerx-dbserver")
	userAddress := strings.ToLower(tempJobs[0].UserAddress)
	traceDBOp := dbTracer.TraceDBOperation(ctx, "SELECT", "users", "GetUserDataByAddress")
	existingUserID, existingUser, err = h.userRepository.GetUserDataByAddress(userAddress)
	trackDBOp(err)
	traceDBOp(err)

	if err != nil && err != gocql.ErrNotFound {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error.type", "user_lookup_failed"))
		h.logger.ErrorWithTrace(ctx, "Error getting user ID for address",
			"user_address", tempJobs[0].UserAddress,
			"error", err.Error(),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Add user context to span
	if err != gocql.ErrNotFound {
		span.SetAttributes(
			attribute.Int64(tracing.TriggerXAttributes.UserID, existingUserID),
			attribute.String("user.status", "existing"),
		)
		h.logger.InfoWithTrace(ctx, "Found existing user",
			"user_id", existingUserID,
			"user_address", userAddress,
		)
	}

	h.logger.Infof("[CreateJobData] existingUserID: %d", existingUserID)

	if err == gocql.ErrNotFound {
		// Create new user span
		_, userSpan := tracer.Start(ctx, "user.create")
		userSpan.SetAttributes(
			attribute.String(tracing.TriggerXAttributes.UserAddress, userAddress),
			attribute.String("user.status", "new"),
		)

		var newUser types.CreateUserDataRequest
		newUser.UserAddress = userAddress
		newUser.EtherBalance = tempJobs[0].EtherBalance
		newUser.TokenBalance = tempJobs[0].TokenBalance
		newUser.UserPoints = 0.0

		// Track user creation
		trackDBOp = metrics.TrackDBOperation("create", "users")
		traceDBOp = dbTracer.TraceDBOperation(ctx, "INSERT", "users", "CreateNewUser")
		existingUser, err = h.userRepository.CreateNewUser(&newUser)
		trackDBOp(err)
		traceDBOp(err)

		if err != nil {
			userSpan.RecordError(err)
			userSpan.End()
			span.RecordError(err)
			span.SetAttributes(attribute.String("error.type", "user_creation_failed"))
			h.logger.Errorf("[CreateJobData] Error creating new user for address %s: %v", tempJobs[0].UserAddress, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		userSpan.SetAttributes(
			attribute.Int64(tracing.TriggerXAttributes.UserID, existingUser.UserID),
			attribute.String("user.ether_balance", existingUser.EtherBalance.String()),
			attribute.String("user.token_balance", existingUser.TokenBalance.String()),
		)
		userSpan.End()

		span.SetAttributes(
			attribute.Int64(tracing.TriggerXAttributes.UserID, existingUser.UserID),
			attribute.String("user.status", "created"),
		)

		h.logger.InfoWithTrace(ctx, "Created new user successfully",
			"user_id", existingUser.UserID,
			"user_address", existingUser.UserAddress,
			"ether_balance", existingUser.EtherBalance.String(),
			"token_balance", existingUser.TokenBalance.String(),
		)
	}

	createdJobs := types.CreateJobResponse{
		UserID:            existingUser.UserID,
		AccountBalance:    existingUser.EtherBalance,
		TokenBalance:      existingUser.TokenBalance,
		JobIDs:            make([]int64, len(tempJobs)),
		TaskDefinitionIDs: make([]int, len(tempJobs)),
		TimeFrames:        make([]int64, len(tempJobs)),
	}

	// Add final user context to main span
	span.SetAttributes(
		attribute.Int64(tracing.TriggerXAttributes.UserID, existingUser.UserID),
		attribute.String("user.account_balance", existingUser.EtherBalance.String()),
		attribute.String("user.token_balance", existingUser.TokenBalance.String()),
	)

	for i := len(tempJobs) - 1; i >= 0; i-- {
		// Create individual job span
		_, jobSpan := tracer.Start(ctx, "job.create_single")
		jobSpan.SetAttributes(
			attribute.Int64(tracing.TriggerXAttributes.UserID, existingUser.UserID),
			attribute.String(tracing.TriggerXAttributes.UserAddress, existingUser.UserAddress),
			attribute.Int(tracing.TriggerXAttributes.TaskDefID, tempJobs[i].TaskDefinitionID),
			attribute.String("job.title", tempJobs[i].JobTitle),
			attribute.Int("job.batch_index", i),
		)

		chainStatus := 1
		var linkJobID int64 = -1

		if i == 0 {
			chainStatus = 0
		}
		if i < len(tempJobs)-1 {
			linkJobID = createdJobs.JobIDs[i+1]
			jobSpan.SetAttributes(attribute.Int64("job.link_job_id", linkJobID))
		}

		jobData := &types.JobData{
			JobTitle:          tempJobs[i].JobTitle,
			TaskDefinitionID:  tempJobs[i].TaskDefinitionID,
			UserID:            existingUser.UserID,
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
		traceDBOp = dbTracer.TraceDBOperation(ctx, "INSERT", "jobs", "CreateNewJob")
		jobID, err := h.jobRepository.CreateNewJob(jobData)
		trackDBOp(err)
		traceDBOp(err)

		if err != nil {
			jobSpan.RecordError(err)
			jobSpan.End()
			span.RecordError(err)
			h.logger.Errorf("[CreateJobData] Error creating job: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// Add job ID to span now that it's created
		jobSpan.SetAttributes(
			attribute.Int64(tracing.TriggerXAttributes.JobID, jobID),
			attribute.String("job.status", "pending"),
			attribute.Int("job.chain_status", chainStatus),
		)

		createdJobs.JobIDs[i] = jobID
		expirationTime := time.Now().Add(time.Duration(tempJobs[i].TimeFrame) * time.Second)
		var scheduleConditionJobData commonTypes.ScheduleConditionJobData

		// Determine job type and add specific attributes
		switch tempJobs[i].TaskDefinitionID {
		case 1, 2:
			// Time-based job
			jobSpan.SetAttributes(
				attribute.String(tracing.TriggerXAttributes.JobType, "time_based"),
				attribute.Int64("time.interval", tempJobs[i].TimeInterval),
				attribute.String("time.schedule_type", "interval"),
			)

			var nextExecutionTimestamp time.Time
			nextExecutionTimestamp, err := parser.CalculateNextExecutionTime(time.Now(), "interval", tempJobs[i].TimeInterval, tempJobs[i].CronExpression, tempJobs[i].SpecificSchedule)
			if err != nil {
				h.logger.Errorf("[getNextExecutionTimestamp] Error calculating next execution timestamp: %v", err)
				nextExecutionTimestamp = time.Now().Add(time.Duration(tempJobs[i].TimeInterval) * time.Second)
			}

			timeJobData := types.TimeJobData{
				JobID:            jobID,
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
				h.logger.Errorf("[CreateJobData] Error inserting time job data for jobID %d: %v", jobID, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				return
			}
			trackDBOp(nil)
			h.logger.Infof("[CreateJobData] Successfully created time-based job %d with interval %d seconds",
				jobID, timeJobData.TimeInterval)

		case 3, 4:
			// Event-based job
			jobSpan.SetAttributes(
				attribute.String(tracing.TriggerXAttributes.JobType, "event_based"),
				attribute.String(tracing.TriggerXAttributes.BlockchainNetwork, tempJobs[i].TriggerChainID),
				attribute.String("trigger.contract_address", tempJobs[i].TriggerContractAddress),
				attribute.String("trigger.event", tempJobs[i].TriggerEvent),
			)

			eventJobData := types.EventJobData{
				JobID:                     jobID,
				TaskDefinitionID:          tempJobs[i].TaskDefinitionID,
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
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				return
			}
			scheduleConditionJobData.JobID = jobID
			scheduleConditionJobData.TaskDefinitionID = tempJobs[i].TaskDefinitionID
			scheduleConditionJobData.LastExecutedAt = time.Now()
			scheduleConditionJobData.TaskTargetData = commonTypes.TaskTargetData{
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
			scheduleConditionJobData.EventWorkerData = commonTypes.EventWorkerData{
				JobID:                  jobID,
				ExpirationTime:         expirationTime,
				Recurring:              tempJobs[i].Recurring,
				TriggerChainID:         tempJobs[i].TriggerChainID,
				TriggerContractAddress: tempJobs[i].TriggerContractAddress,
				TriggerEvent:           tempJobs[i].TriggerEvent,
			}
			h.logger.Infof("[CreateJobData] Successfully created event-based job %d for event %s on contract %s",
				jobID, eventJobData.TriggerEvent, eventJobData.TriggerContractAddress)

		case 5, 6:
			// Condition-based job
			jobSpan.SetAttributes(
				attribute.String(tracing.TriggerXAttributes.JobType, "condition_based"),
			)

			conditionJobData := types.ConditionJobData{
				JobID:                     jobID,
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
			}

			if err := h.conditionJobRepository.CreateConditionJob(&conditionJobData); err != nil {
				h.logger.Errorf("[CreateJobData] Error inserting condition job data for jobID %d: %v", jobID, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				return
			}
			scheduleConditionJobData.JobID = jobID
			scheduleConditionJobData.TaskDefinitionID = tempJobs[i].TaskDefinitionID
			scheduleConditionJobData.LastExecutedAt = time.Now()
			scheduleConditionJobData.TaskTargetData = commonTypes.TaskTargetData{
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
			scheduleConditionJobData.ConditionWorkerData = commonTypes.ConditionWorkerData{
				JobID:           jobID,
				ExpirationTime:  expirationTime,
				Recurring:       tempJobs[i].Recurring,
				ConditionType:   tempJobs[i].ConditionType,
				UpperLimit:      tempJobs[i].UpperLimit,
				LowerLimit:      tempJobs[i].LowerLimit,
				ValueSourceType: tempJobs[i].ValueSourceType,
				ValueSourceUrl:  tempJobs[i].ValueSourceUrl,
			}
			h.logger.Infof("[CreateJobData] Successfully created condition-based job %d with condition type %s (limits: %f-%f)",
				jobID, conditionJobData.ConditionType, conditionJobData.LowerLimit, conditionJobData.UpperLimit)
		default:
			h.logger.Errorf("[CreateJobData] Invalid task definition ID %d for job %d", tempJobs[i].TaskDefinitionID, i)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task definition ID"})
			return
		}

		if tempJobs[i].TaskDefinitionID == 3 || tempJobs[i].TaskDefinitionID == 4 || tempJobs[i].TaskDefinitionID == 5 || tempJobs[i].TaskDefinitionID == 6 {
			go func() {
				success, err := h.notifyConditionScheduler(jobID, scheduleConditionJobData)
				if !success {
					h.logger.Errorf("[CreateJobData] Error notifying condition scheduler for jobID %d: %v", jobID, err)
				} else {
					h.logger.Infof("[CreateJobData] Successfully notified condition scheduler for jobID %d", jobID)
				}
			}()
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
			h.logger.Errorf("[CreateJobData] Error updating user points for userID %d: %v", existingUser.UserID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		trackDBOp(nil)

		createdJobs.JobIDs[i] = jobID
		createdJobs.TaskDefinitionIDs[i] = tempJobs[i].TaskDefinitionID
		createdJobs.TimeFrames[i] = tempJobs[i].TimeFrame

		jobSpan.End()
	}

	// Update user's job_ids
	allJobIDs := append(existingUser.JobIDs, createdJobs.JobIDs...)
	trackDBOp = metrics.TrackDBOperation("update", "users")
	if err := h.userRepository.UpdateUserJobIDs(existingUser.UserID, allJobIDs); err != nil {
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
