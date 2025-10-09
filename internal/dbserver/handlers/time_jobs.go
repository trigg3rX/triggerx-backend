package handlers

import (
	"context"
	"math/big"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/config"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (h *Handler) GetTimeBasedTasks(c *gin.Context) {
	pollLookAhead := config.GetPollingLookAhead()
	lookAheadTime := time.Now().Add(time.Duration(pollLookAhead) * time.Second)

	ctx := context.Background()

	// Get all time jobs
	trackDBOp := metrics.TrackDBOperation("read", "time_jobs")
	allTimeJobs, err := h.timeJobRepository.List(ctx)
	trackDBOp(err)
	if err != nil {
		h.logger.Errorf("[GetTimeBasedTasks] Error retrieving time jobs: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve time based tasks",
			"code":  "TIME_TASKS_FETCH_ERROR",
		})
		return
	}

	// Filter jobs by next execution timestamp
	var filteredJobs []*types.TimeJobDataEntity
	for _, job := range allTimeJobs {
		if !job.IsCompleted && job.NextExecutionTimestamp.Before(lookAheadTime) {
			filteredJobs = append(filteredJobs, job)
		}
	}

	// Convert to ScheduleTimeTaskData format
	var tasks []types.ScheduleTimeTaskData
	for _, timeJob := range filteredJobs {
		// Create task for this job
		newTask := &types.TaskDataEntity{
			TaskID:               0, // Will be auto-generated
			TaskNumber:           0,
			JobID:                timeJob.JobID,
			TaskDefinitionID:     timeJob.TaskDefinitionID,
			CreatedAt:            time.Now().UTC(),
			TaskOpxPredictedCost: *big.NewInt(0),
			TaskOpxActualCost:    *big.NewInt(0),
			IsSuccessful:         false,
			IsAccepted:           false,
		}

		trackDBOp = metrics.TrackDBOperation("create", "task_data")
		err := h.taskRepository.Create(ctx, newTask)
		trackDBOp(err)
		if err != nil {
			h.logger.Errorf("[GetTimeBasedJobs] Error creating task data: %v", err)
			continue
		}

		// Get created task to get ID
		task, err := h.taskRepository.GetByID(ctx, newTask.TaskID)
		if err != nil || task == nil {
			continue
		}

		// Update job with new task ID
		job, err := h.jobRepository.GetByID(ctx, &timeJob.JobID)
		if err == nil && job != nil {
			job.TaskIDs = append(job.TaskIDs, task.TaskID)
			err = h.jobRepository.Update(ctx, job)
			if err != nil {
				h.logger.Errorf("[GetTimeBasedJobs] Error updating job with new task ID: %v", err)
				continue
			}
		}

		// Build response
		scheduleData := types.ScheduleTimeTaskData{
			TaskID:                 task.TaskID,
			TaskDefinitionID:       timeJob.TaskDefinitionID,
			LastExecutedAt:         timeJob.LastExecutedAt,
			ExpirationTime:         timeJob.ExpirationTime,
			NextExecutionTimestamp: timeJob.NextExecutionTimestamp,
			ScheduleType:           timeJob.ScheduleType,
			TimeInterval:           timeJob.TimeInterval,
			CronExpression:         timeJob.CronExpression,
			SpecificSchedule:       timeJob.SpecificSchedule,
			TaskTargetData: types.TaskTargetData{
				JobID:                     types.NewBigInt(&timeJob.JobID),
				TaskDefinitionID:          timeJob.TaskDefinitionID,
				TargetChainID:             timeJob.TargetChainID,
				TargetContractAddress:     timeJob.TargetContractAddress,
				TargetFunction:            timeJob.TargetFunction,
				ABI:                       timeJob.ABI,
				ArgType:                   timeJob.ArgType,
				Arguments:                 timeJob.Arguments,
				DynamicArgumentsScriptUrl: timeJob.DynamicArgumentsScriptURL,
			},
		}
		tasks = append(tasks, scheduleData)
	}

	h.logger.Infof("[GetTimeBasedJobs] Successfully retrieved %d time based jobs", len(tasks))
	c.JSON(http.StatusOK, tasks)
}
