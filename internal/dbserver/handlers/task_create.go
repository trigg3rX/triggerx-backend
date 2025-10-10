package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (h *Handler) CreateTaskData(c *gin.Context) {
	logger := h.getLogger(c)
	logger.Debugf("POST [CreateTaskData] Creating task")
	var taskData types.CreateTaskDataRequest
	if err := c.ShouldBindJSON(&taskData); err != nil {
		logger.Errorf("Error decoding request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}


	// Create new task entity
	newTask := &types.TaskDataEntity{
		TaskID:               0, // Will be auto-generated
		TaskNumber:           0,
		JobID:                taskData.JobID,
		TaskDefinitionID:     taskData.TaskDefinitionID,
		CreatedAt:            time.Now().UTC(),
		TaskOpxPredictedCost: "0",
		TaskOpxActualCost:    "0",
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
	err := h.taskRepository.Create(c.Request.Context(), newTask)
	trackDBOp(err)
	if err != nil {
		logger.Errorf("Error creating task: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create task",
			"code":  "TASK_CREATION_ERROR",
		})
		return
	}

	// Get the created task to retrieve the generated ID
	task, err := h.taskRepository.GetByID(c.Request.Context(), newTask.TaskID)
	if err != nil || task == nil {
		logger.Errorf("Error fetching created task: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch created task",
			"code":  "TASK_FETCH_ERROR",
		})
		return
	}

	taskID := task.TaskID

	// Add task ID to job's task_ids list
	trackDBOp = metrics.TrackDBOperation("update", "add_task_id")
	job, err := h.jobRepository.GetByID(c.Request.Context(), taskData.JobID)
	if err != nil || job == nil {
		logger.Errorf("Error getting job: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Job not found",
			"code":  "JOB_NOT_FOUND",
		})
		return
	}

	job.TaskIDs = append(job.TaskIDs, taskID)
	err = h.jobRepository.Update(c.Request.Context(), job)
	trackDBOp(err)
	if err != nil {
		logger.Errorf("Error adding task ID to job: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to add task ID to job",
			"code":  "TASK_ID_ADDITION_ERROR",
		})
		return
	}

	logger.Debugf("Successfully created task with ID: %d", taskID)
	c.JSON(http.StatusCreated, gin.H{"task_id": taskID})
}
