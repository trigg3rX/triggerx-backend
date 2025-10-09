package handlers

import (
	"context"
	"math/big"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (h *Handler) CreateTaskData(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Infof("[CreateTaskData] trace_id=%s - Creating task", traceID)
	var taskData types.CreateTaskDataRequest
	if err := c.ShouldBindJSON(&taskData); err != nil {
		h.logger.Errorf("[CreateTaskData] Error decoding request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	ctx := context.Background()

	// Create new task entity
	newTask := &types.TaskDataEntity{
		TaskID:               0, // Will be auto-generated
		TaskNumber:           0,
		JobID:                *taskData.JobID,
		TaskDefinitionID:     taskData.TaskDefinitionID,
		CreatedAt:            time.Now().UTC(),
		TaskOpxPredictedCost: *big.NewInt(0),
		TaskOpxActualCost:    *big.NewInt(0),
		ExecutionTimestamp:   time.Time{},
		ExecutionTxHash:      "",
		TaskPerformerID:      0,
		TaskAttesterIDs:      []int64{},
		ConvertedArguments:   "",
		ProofOfTask:          "",
		SubmissionTxHash:     "",
		IsSuccessful:         false,
		IsAccepted:           false,
		IsImua:               taskData.IsImua,
	}

	trackDBOp := metrics.TrackDBOperation("create", "task_data")
	err := h.taskRepository.Create(ctx, newTask)
	trackDBOp(err)
	if err != nil {
		h.logger.Errorf("[CreateTaskData] Error creating task: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create task",
			"code":  "TASK_CREATION_ERROR",
		})
		return
	}

	// Get the created task to retrieve the generated ID
	task, err := h.taskRepository.GetByID(ctx, newTask.TaskID)
	if err != nil || task == nil {
		h.logger.Errorf("[CreateTaskData] Error fetching created task: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch created task",
			"code":  "TASK_FETCH_ERROR",
		})
		return
	}

	taskID := task.TaskID

	// Add task ID to job's task_ids list
	trackDBOp = metrics.TrackDBOperation("update", "add_task_id")
	job, err := h.jobRepository.GetByID(ctx, taskData.JobID)
	if err != nil || job == nil {
		h.logger.Errorf("[CreateTaskData] Error getting job: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Job not found",
			"code":  "JOB_NOT_FOUND",
		})
		return
	}

	job.TaskIDs = append(job.TaskIDs, taskID)
	err = h.jobRepository.Update(ctx, job)
	trackDBOp(err)
	if err != nil {
		h.logger.Errorf("[CreateTaskData] Error adding task ID to job: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to add task ID to job",
			"code":  "TASK_ID_ADDITION_ERROR",
		})
		return
	}

	h.logger.Infof("[CreateTaskData] Successfully created task with ID: %d", taskID)
	c.JSON(http.StatusCreated, gin.H{"task_id": taskID})
}
