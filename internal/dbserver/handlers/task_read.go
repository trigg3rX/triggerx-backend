package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
)

func (h *Handler) GetTaskDataByID(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		h.logger.Error("[GetTaskDataByID] No task ID provided")
		c.JSON(http.StatusBadRequest, gin.H{"error": "No task ID provided"})
		return
	}

	taskIDInt, err := strconv.ParseInt(taskID, 10, 64)
	if err != nil {
		h.logger.Errorf("[GetTaskDataByID] Invalid task ID format: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID format"})
		return
	}

	h.logger.Infof("[GetTaskDataByID] Retrieving task data for task ID: %d", taskIDInt)

	trackDBOp := metrics.TrackDBOperation("read", "task_data")
	taskData, err := h.taskRepository.GetTaskDataByID(taskIDInt)
	trackDBOp(err)
	if err != nil {
		h.logger.Errorf("[GetTaskDataByID] Error retrieving task data for taskID %d: %v", taskIDInt, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("[GetTaskDataByID] Successfully retrieved task data for task ID: %d", taskIDInt)
	c.JSON(http.StatusOK, taskData)
}

func (h *Handler) GetTasksByJobID(c *gin.Context) {
	jobID := c.Param("job_id")
	if jobID == "" {
		h.logger.Error("[GetTasksByJobID] No job ID provided")
		c.JSON(http.StatusBadRequest, gin.H{"error": "No job ID provided"})
		return
	}

	jobIDInt, err := strconv.ParseInt(jobID, 10, 64)
	if err != nil {
		h.logger.Errorf("[GetTasksByJobID] Invalid job ID format: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID format"})
		return
	}

	h.logger.Infof("[GetTasksByJobID] Retrieving tasks for job ID: %d", jobIDInt)

	trackDBOp := metrics.TrackDBOperation("read", "task_data")
	tasks, err := h.taskRepository.GetTasksByJobID(jobIDInt)
	trackDBOp(err)
	if err != nil {
		h.logger.Errorf("[GetTasksByJobID] Error retrieving tasks for jobID %d: %v", jobIDInt, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("[GetTasksByJobID] Successfully retrieved %d tasks for job ID: %d", len(tasks), jobIDInt)
	c.JSON(http.StatusOK, tasks)
}
