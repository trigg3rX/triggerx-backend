package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetTaskDataByID(c *gin.Context) {
	taskID := c.Param("id")
	h.logger.Infof("[GetTaskDataByID] Fetching task with ID: %s", taskID)

	taskIDInt, err := strconv.ParseInt(taskID, 10, 64)
	if err != nil {
		h.logger.Errorf("[GetTaskDataByID] Error parsing task ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}
	taskData, err := h.taskRepository.GetTaskDataByID(taskIDInt)
	if err != nil {
		h.logger.Errorf("[GetTaskDataByID] Error retrieving task with ID %s: %v", taskID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("[GetTaskDataByID] Successfully retrieved task with ID: %s", taskID)
	c.JSON(http.StatusOK, taskData)
}

func (h *Handler) GetTasksByJobID(c *gin.Context) {
	jobID := c.Param("id")
	h.logger.Infof("[GetTasksByJobID] Fetching tasks for job with ID: %s", jobID)

	jobIDInt, err := strconv.ParseInt(jobID, 10, 64)
	if err != nil {
		h.logger.Errorf("[GetTasksByJobID] Error parsing job ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}
	taskData, err := h.taskRepository.GetTasksByJobID(jobIDInt)
	if err != nil {
		h.logger.Errorf("[GetTasksByJobID] Error retrieving tasks for job with ID %s: %v", jobID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("[GetTasksByJobID] Successfully retrieved tasks for job with ID: %s", jobID)
	c.JSON(http.StatusOK, taskData)
}
