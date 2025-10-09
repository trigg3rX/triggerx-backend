package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (h *Handler) GetPerformers(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Infof("[GetPerformers] trace_id=%s - Retrieving performers", traceID)

	ctx := context.Background()

	trackDBOp := metrics.TrackDBOperation("read", "keepers")
	performers, err := h.keeperRepository.List(ctx)
	trackDBOp(err)
	if err != nil {
		h.logger.Errorf("[GetPerformers] Error retrieving performers: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Filter for registered and whitelisted keepers
	var activePerformers []*types.KeeperDataEntity
	for _, keeper := range performers {
		if keeper.Registered != nil && *keeper.Registered && keeper.Whitelisted != nil && *keeper.Whitelisted {
			activePerformers = append(activePerformers, keeper)
		}
	}

	h.logger.Infof("[GetPerformers] Successfully retrieved %d performers", len(activePerformers))
	c.JSON(http.StatusOK, activePerformers)
}

func (h *Handler) GetKeeperData(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Infof("[GetKeeperData] trace_id=%s - Retrieving keeper data", traceID)
	keeperID := c.Param("id")
	h.logger.Infof("[GetKeeperData] Retrieving keeper with ID: %s", keeperID)

	keeperIDInt, err := strconv.ParseInt(keeperID, 10, 64)
	if err != nil {
		h.logger.Errorf("[GetKeeperData] Error parsing keeper ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid keeper ID format",
			"code":  "INVALID_KEEPER_ID",
		})
		return
	}

	ctx := context.Background()

	trackDBOp := metrics.TrackDBOperation("read", "keeper_data")
	keeperData, err := h.keeperRepository.GetByNonID(ctx, "operator_id", keeperIDInt)
	trackDBOp(err)
	if err != nil || keeperData == nil {
		h.logger.Errorf("[GetKeeperData] Error retrieving keeper data: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Keeper not found",
			"code":  "KEEPER_NOT_FOUND",
		})
		return
	}

	h.logger.Infof("[GetKeeperData] Successfully retrieved keeper with ID: %s", keeperID)
	c.JSON(http.StatusOK, keeperData)
}

func (h *Handler) GetKeeperTaskCount(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Infof("[GetKeeperTaskCount] trace_id=%s - Retrieving keeper task count", traceID)
	keeperID := c.Param("id")
	h.logger.Infof("[GetKeeperTaskCount] Retrieving task count for keeper with ID: %s", keeperID)

	keeperIDInt, err := strconv.ParseInt(keeperID, 10, 64)
	if err != nil {
		h.logger.Errorf("[GetKeeperTaskCount] Error parsing keeper ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid keeper ID format",
			"code":  "INVALID_KEEPER_ID",
		})
		return
	}

	ctx := context.Background()

	trackDBOp := metrics.TrackDBOperation("read", "keeper_tasks")
	keeper, err := h.keeperRepository.GetByNonID(ctx, "operator_id", keeperIDInt)
	trackDBOp(err)
	if err != nil || keeper == nil {
		h.logger.Errorf("[GetKeeperTaskCount] Error retrieving keeper: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Keeper not found",
			"code":  "KEEPER_NOT_FOUND",
		})
		return
	}

	taskCount := int64(0)
	if keeper.NoExecutedTasks != nil {
		taskCount = *keeper.NoExecutedTasks
	}

	h.logger.Infof("[GetKeeperTaskCount] Successfully retrieved task count %d for keeper ID: %s", taskCount, keeperID)
	c.JSON(http.StatusOK, gin.H{"no_executed_tasks": taskCount})
}

func (h *Handler) GetKeeperPoints(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Infof("[GetKeeperPoints] trace_id=%s - Retrieving points for keeper", traceID)
	keeperID := c.Param("id")
	h.logger.Infof("[GetKeeperPoints] Retrieving points for keeper with ID: %s", keeperID)

	keeperIDInt, err := strconv.ParseInt(keeperID, 10, 64)
	if err != nil {
		h.logger.Errorf("[GetKeeperPoints] Error parsing keeper ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()

	trackDBOp := metrics.TrackDBOperation("read", "keeper_points")
	keeper, err := h.keeperRepository.GetByNonID(ctx, "operator_id", keeperIDInt)
	trackDBOp(err)
	if err != nil || keeper == nil {
		h.logger.Errorf("[GetKeeperPoints] Error retrieving keeper: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Keeper not found",
			"code":  "KEEPER_NOT_FOUND",
		})
		return
	}

	h.logger.Infof("[GetKeeperPoints] Successfully retrieved points for keeper ID: %s", keeperID)
	c.JSON(http.StatusOK, gin.H{"keeper_points": keeper.KeeperPoints})
}

func (h *Handler) GetKeeperCommunicationInfo(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Infof("[GetKeeperCommunicationInfo] trace_id=%s - Retrieving communication info for keeper", traceID)
	keeperID := c.Param("id")
	h.logger.Infof("[GetKeeperChatInfo] Retrieving chat ID, keeper name, and email for keeper with ID: %s", keeperID)

	keeperIDInt, err := strconv.ParseInt(keeperID, 10, 64)
	if err != nil {
		h.logger.Errorf("[GetKeeperChatInfo] Error parsing keeper ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()

	trackDBOp := metrics.TrackDBOperation("read", "keeper_communication")
	keeper, err := h.keeperRepository.GetByNonID(ctx, "operator_id", keeperIDInt)
	trackDBOp(err)
	if err != nil || keeper == nil {
		h.logger.Errorf("[GetKeeperChatInfo] Error retrieving keeper for ID %s: %v", keeperID, err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Keeper not found",
			"code":  "KEEPER_NOT_FOUND",
		})
		return
	}

	// Return communication info
	chatID := int64(0)
	if keeper.ChatID != nil {
		chatID = *keeper.ChatID
	}

	keeperData := map[string]interface{}{
		"chat_id":     chatID,
		"keeper_name": keeper.KeeperName,
		"email_id":    keeper.EmailID,
	}

	h.logger.Infof("[GetKeeperChatInfo] Successfully retrieved chat ID, keeper name, and email for ID: %s", keeperID)
	c.JSON(http.StatusOK, keeperData)
}
