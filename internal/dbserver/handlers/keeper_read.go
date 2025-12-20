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
	traceID := h.getTraceID(c)
	h.logger.Info(c.Request.Context(), "[GetPerformers] trace_id=%s - Retrieving performers", observability.String("trace_id", traceID))
	trackDBOp := metrics.TrackDBOperation("read", "keepers")
	performers, err := h.keeperRepository.GetKeeperAsPerformer()
	trackDBOp(err)
	if err != nil {
		h.logger.Error(c.Request.Context(), "[GetPerformers] Error retrieving performers: %v", observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	sort.Slice(performers, func(i, j int) bool {
		return performers[i].KeeperID < performers[j].KeeperID
	})

	h.logger.Info(c.Request.Context(), "[GetPerformers] Successfully retrieved %d performers", observability.Int("performers_count", len(performers)))
	c.JSON(http.StatusOK, performers)
}

func (h *Handler) GetKeeperData(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Info(c.Request.Context(), "[GetKeeperData] trace_id=%s - Retrieving keeper data", observability.String("trace_id", traceID))
	keeperID := c.Param("id")
	h.logger.Info(c.Request.Context(), "[GetKeeperData] Retrieving keeper with ID: %s", observability.String("keeper_id", keeperID))

	keeperIDInt, err := strconv.ParseInt(keeperID, 10, 64)
	if err != nil {
		h.logger.Error(c.Request.Context(), "[GetKeeperData] Error parsing keeper ID: %v", observability.Error(err))
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
		h.logger.Error(c.Request.Context(), "[GetKeeperData] Error retrieving keeper data: %v", observability.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Keeper not found",
			"code":  "KEEPER_NOT_FOUND",
		})
		return
	}

	h.logger.Info(c.Request.Context(), "[GetKeeperData] Successfully retrieved keeper with ID: %s", observability.String("keeper_id", keeperID))
	c.JSON(http.StatusOK, keeperData)
}

func (h *Handler) GetKeeperTaskCount(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Info(c.Request.Context(), "[GetKeeperTaskCount] trace_id=%s - Retrieving keeper task count", observability.String("trace_id", traceID))
	keeperID := c.Param("id")
	h.logger.Info(c.Request.Context(), "[GetKeeperTaskCount] Retrieving task count for keeper with ID: %s", observability.String("keeper_id", keeperID))

	keeperIDInt, err := strconv.ParseInt(keeperID, 10, 64)
	if err != nil {
		h.logger.Error(c.Request.Context(), "[GetKeeperTaskCount] Error parsing keeper ID: %v", observability.Error(err))
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
		h.logger.Error(c.Request.Context(), "[GetKeeperTaskCount] Error retrieving task count: %v", observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info(c.Request.Context(), "[GetKeeperTaskCount] Successfully retrieved task count %d for keeper ID: %s", observability.Int64("task_count", taskCount), observability.String("keeper_id", keeperID))
	c.JSON(http.StatusOK, gin.H{"no_executed_tasks": taskCount})
}

func (h *Handler) GetKeeperPoints(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Info(c.Request.Context(), "[GetKeeperPoints] trace_id=%s - Retrieving points for keeper", observability.String("trace_id", traceID))
	keeperID := c.Param("id")
	h.logger.Info(c.Request.Context(), "[GetKeeperPoints] Retrieving points for keeper with ID: %s", observability.String("keeper_id", keeperID))

	keeperIDInt, err := strconv.ParseInt(keeperID, 10, 64)
	if err != nil {
		h.logger.Error(c.Request.Context(), "[GetKeeperPoints] Error parsing keeper ID: %v", observability.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trackDBOp := metrics.TrackDBOperation("read", "keeper_points")
	points, err := h.keeperRepository.GetKeeperPointsByIDInDB(keeperIDInt)
	trackDBOp(err)
	if err != nil {
		h.logger.Error(c.Request.Context(), "[GetKeeperPoints] Error retrieving keeper points: %v", observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info(c.Request.Context(), "[GetKeeperPoints] Successfully retrieved points %f for keeper ID: %s", observability.Float64("points", points), observability.String("keeper_id", keeperID))
	c.JSON(http.StatusOK, gin.H{"keeper_points": points})
}

func (h *Handler) GetKeeperCommunicationInfo(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Info(c.Request.Context(), "[GetKeeperCommunicationInfo] trace_id=%s - Retrieving communication info for keeper", observability.String("trace_id", traceID))
	keeperID := c.Param("id")
	h.logger.Info(c.Request.Context(), "[GetKeeperChatInfo] Retrieving chat ID, keeper name, and email for keeper with ID: %s", observability.String("keeper_id", keeperID))

	keeperIDInt, err := strconv.ParseInt(keeperID, 10, 64)
	if err != nil {
		h.logger.Error(c.Request.Context(), "[GetKeeperChatInfo] Error parsing keeper ID: %v", observability.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trackDBOp := metrics.TrackDBOperation("read", "keeper_communication")
	keeperData, err := h.keeperRepository.GetKeeperCommunicationInfo(keeperIDInt)
	trackDBOp(err)
	if err != nil {
		h.logger.Error(c.Request.Context(), "[GetKeeperChatInfo] Error retrieving chat ID, keeper name, and email for ID %s: %v", observability.String("keeper_id", keeperID), observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info(c.Request.Context(), "[GetKeeperChatInfo] Successfully retrieved chat ID, keeper name, and email for ID: %s", observability.String("keeper_id", keeperID))
	c.JSON(http.StatusOK, keeperData)
}
