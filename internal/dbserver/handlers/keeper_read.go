package handlers

import (
	"net/http"
	"sort"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

func (h *Handler) GetPerformers(c *gin.Context) {
	trackDBOp := metrics.TrackDBOperation("read", "keepers")
	performers, err := h.keeperRepository.GetKeeperAsPerformer()
	trackDBOp(err)
	if err != nil {
		h.logger.Warn(c.Request.Context(), "[GetPerformers] Failed to retrieve performers", observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	sort.Slice(performers, func(i, j int) bool {
		return performers[i].KeeperID < performers[j].KeeperID
	})

	c.JSON(http.StatusOK, performers)
	h.logger.Debug(c.Request.Context(), "[GetPerformers] Retrieved performers", observability.Int("performers_count", len(performers)))
}

func (h *Handler) GetKeeperData(c *gin.Context) {
	keeperID := c.Param("id")

	keeperIDInt, err := strconv.ParseInt(keeperID, 10, 64)
	if err != nil {
		h.logger.Error(c.Request.Context(), "[GetKeeperData] Error parsing keeper ID", observability.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid keeper ID format",
			"code":  "INVALID_KEEPER_ID",
		})
		return
	}

	trackDBOp := metrics.TrackDBOperation("read", "keeper_data")
	keeperData, err := h.keeperRepository.GetKeeperDataByID(keeperIDInt)
	trackDBOp(err)
	if err != nil {
		h.logger.Warn(c.Request.Context(), "[GetKeeperData] Failed to retrieve keeper data", observability.String("keeper_id", keeperID), observability.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Keeper not found",
			"code":  "KEEPER_NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, keeperData)
	h.logger.Debug(c.Request.Context(), "[GetKeeperData] Retrieved keeper data", observability.String("keeper_id", keeperID))
}

func (h *Handler) GetKeeperTaskCount(c *gin.Context) {
	keeperID := c.Param("id")

	keeperIDInt, err := strconv.ParseInt(keeperID, 10, 64)
	if err != nil {
		h.logger.Error(c.Request.Context(), "[GetKeeperTaskCount] Error parsing keeper ID", observability.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid keeper ID format",
			"code":  "INVALID_KEEPER_ID",
		})
		return
	}

	trackDBOp := metrics.TrackDBOperation("read", "keeper_tasks")
	taskCount, err := h.keeperRepository.GetKeeperTaskCount(keeperIDInt)
	trackDBOp(err)
	if err != nil {
		h.logger.Warn(c.Request.Context(), "[GetKeeperTaskCount] Failed to retrieve task count", observability.String("keeper_id", keeperID), observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"no_executed_tasks": taskCount})
	h.logger.Debug(c.Request.Context(), "[GetKeeperTaskCount] Retrieved task count", observability.String("keeper_id", keeperID), observability.Int64("task_count", taskCount))
}

func (h *Handler) GetKeeperPoints(c *gin.Context) {
	keeperID := c.Param("id")

	keeperIDInt, err := strconv.ParseInt(keeperID, 10, 64)
	if err != nil {
		h.logger.Error(c.Request.Context(), "[GetKeeperPoints] Error parsing keeper ID", observability.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trackDBOp := metrics.TrackDBOperation("read", "keeper_points")
	points, err := h.keeperRepository.GetKeeperPointsByIDInDB(keeperIDInt)
	trackDBOp(err)
	if err != nil {
		h.logger.Warn(c.Request.Context(), "[GetKeeperPoints] Failed to retrieve keeper points", observability.String("keeper_id", keeperID), observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"keeper_points": points})
	h.logger.Debug(c.Request.Context(), "[GetKeeperPoints] Retrieved keeper points", observability.String("keeper_id", keeperID), observability.Float64("points", points))
}

func (h *Handler) GetKeeperCommunicationInfo(c *gin.Context) {
	keeperID := c.Param("id")

	keeperIDInt, err := strconv.ParseInt(keeperID, 10, 64)
	if err != nil {
		h.logger.Error(c.Request.Context(), "[GetKeeperChatInfo] Error parsing keeper ID", observability.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trackDBOp := metrics.TrackDBOperation("read", "keeper_communication")
	keeperData, err := h.keeperRepository.GetKeeperCommunicationInfo(keeperIDInt)
	trackDBOp(err)
	if err != nil {
		h.logger.Warn(c.Request.Context(), "[GetKeeperCommunicationInfo] Failed to retrieve communication info", observability.String("keeper_id", keeperID), observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, keeperData)
	h.logger.Debug(c.Request.Context(), "[GetKeeperCommunicationInfo] Retrieved communication info", observability.String("keeper_id", keeperID))
}
