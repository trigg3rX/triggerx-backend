package handlers

import (
	"math/big"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
)

func (h *Handler) GetTaskDataByID(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Infof("[GetTaskDataByID] trace_id=%s - Retrieving task data", traceID)
	taskID := c.Param("id")
	if taskID == "" {
		h.logger.Error("[GetTaskDataByID] No task ID provided")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No task ID provided",
			"code":  "MISSING_TASK_ID",
		})
		return
	}

	taskIDInt, err := strconv.ParseInt(taskID, 10, 64)
	if err != nil {
		h.logger.Errorf("[GetTaskDataByID] Invalid task ID format: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid task ID format",
			"code":  "INVALID_TASK_ID",
		})
		return
	}

	h.logger.Infof("[GetTaskDataByID] Retrieving task data for task ID: %d", taskIDInt)

	trackDBOp := metrics.TrackDBOperation("read", "task_data")
	taskData, err := h.taskRepository.GetTaskDataByID(taskIDInt)
	trackDBOp(err)
	if err != nil {
		h.logger.Errorf("[GetTaskDataByID] Error retrieving task data for taskID %d: %v", taskIDInt, err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Task not found",
			"code":  "TASK_NOT_FOUND",
		})
		return
	}

	h.logger.Infof("[GetTaskDataByID] Successfully retrieved task data for task ID: %d", taskIDInt)
	c.JSON(http.StatusOK, taskData)
}

func (h *Handler) GetTasksByJobID(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Infof("[GetTasksByJobID] trace_id=%s - Retrieving tasks for job", traceID)
	jobID := c.Param("job_id")
	if jobID == "" {
		h.logger.Error("[GetTasksByJobID] No job ID provided")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No job ID provided",
			"code":  "MISSING_JOB_ID",
		})
		return
	}

	jobIDBig := new(big.Int)
	_, ok := jobIDBig.SetString(jobID, 10)
	if !ok {
		h.logger.Errorf("[GetTasksByJobID] Invalid job ID format: %v", jobID)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid job ID format",
			"code":  "INVALID_JOB_ID",
		})
		return
	}

	h.logger.Infof("[GetTasksByJobID] Retrieving tasks for job ID: %s", jobIDBig.String())

	trackDBOp := metrics.TrackDBOperation("read", "task_data")
	tasks, err := h.taskRepository.GetTasksByJobID(jobIDBig)
	trackDBOp(err)
	if err != nil {
		h.logger.Errorf("[GetTasksByJobID] Error retrieving tasks for jobID %s: %v", jobIDBig.String(), err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No tasks found for this job",
			"code":  "TASKS_NOT_FOUND",
		})
		return
	}

	h.logger.Infof("[GetTasksByJobID] Successfully retrieved %d tasks for job ID: %s", len(tasks), jobIDBig.String())
	c.JSON(http.StatusOK, tasks)
}
