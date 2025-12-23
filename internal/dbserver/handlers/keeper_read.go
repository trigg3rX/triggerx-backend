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
		h.logger.Error(c.Request.Context(), "[GetPerformers] Error retrieving performers", observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	sort.Slice(performers, func(i, j int) bool {
		return performers[i].KeeperID < performers[j].KeeperID
	})

	h.logger.Info(c.Request.Context(), "[GetPerformers] Successfully retrieved performers", observability.Int("performers_count", len(performers)))
	c.JSON(http.StatusOK, performers)
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
		h.logger.Error(c.Request.Context(), "[GetKeeperData] Error retrieving keeper data", observability.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Keeper not found",
			"code":  "KEEPER_NOT_FOUND",
		})
		return
	}

	h.logger.Info(c.Request.Context(), "[GetKeeperData] Successfully retrieved keeper with ID", observability.String("keeper_id", keeperID))
	c.JSON(http.StatusOK, keeperData)
}

func (h *Handler) GetKeeperTaskCount(c *gin.Context) {
	keeperID := c.Param("id")
	h.logger.Info(c.Request.Context(), "[GetKeeperTaskCount] Retrieving task count for keeper with ID", observability.String("keeper_id", keeperID))

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
		h.logger.Error(c.Request.Context(), "[GetKeeperTaskCount] Error retrieving task count", observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info(c.Request.Context(), "[GetKeeperTaskCount] Successfully retrieved task count for keeper ID", observability.Int64("task_count", taskCount), observability.String("keeper_id", keeperID))
	c.JSON(http.StatusOK, gin.H{"no_executed_tasks": taskCount})
}

func (h *Handler) GetKeeperPoints(c *gin.Context) {
	keeperID := c.Param("id")
	h.logger.Info(c.Request.Context(), "[GetKeeperPoints] Retrieving points for keeper with ID", observability.String("keeper_id", keeperID))

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
		h.logger.Error(c.Request.Context(), "[GetKeeperPoints] Error retrieving keeper points", observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info(c.Request.Context(), "[GetKeeperPoints] Successfully retrieved points for keeper ID", observability.Float64("points", points), observability.String("keeper_id", keeperID))
	c.JSON(http.StatusOK, gin.H{"keeper_points": points})
}

func (h *Handler) GetKeeperCommunicationInfo(c *gin.Context) {
	keeperID := c.Param("id")
	h.logger.Info(c.Request.Context(), "[GetKeeperChatInfo] Retrieving chat ID, keeper name, and email for keeper with ID", observability.String("keeper_id", keeperID))

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
		h.logger.Error(c.Request.Context(), "[GetKeeperChatInfo] Error retrieving chat ID, keeper name, and email for ID", observability.String("keeper_id", keeperID), observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info(c.Request.Context(), "[GetKeeperChatInfo] Successfully retrieved chat ID, keeper name, and email for ID", observability.String("keeper_id", keeperID))
	c.JSON(http.StatusOK, keeperData)
}
