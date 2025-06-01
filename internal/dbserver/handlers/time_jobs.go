package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
)

func (h *Handler) GetTimeBasedJobs(c *gin.Context) {
	var pollInterval int64
	if err := c.ShouldBindQuery(&pollInterval); err != nil {
		h.logger.Errorf("[GetTimeBasedJobs] Error getting poll interval: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	nextExecutionTimestamp := time.Now().Add(time.Duration(pollInterval) * time.Second)

	var jobs []types.TimeJobData

	jobs, err := h.timeJobRepository.GetTimeJobsByNextExecutionTimestamp(nextExecutionTimestamp)
	if err != nil {
		h.logger.Errorf("[GetTimeBasedJobs] Error retrieving time based jobs: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if jobs == nil {
		jobs = []types.TimeJobData{}
	}

	h.logger.Infof("[GetTimeBasedJobs] Successfully retrieved %d time based jobs", len(jobs))
	c.JSON(http.StatusOK, jobs)
}
