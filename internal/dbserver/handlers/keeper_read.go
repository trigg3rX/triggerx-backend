package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
)

func (h *Handler) GetKeeperData(c *gin.Context) {
	logger := h.getLogger(c)
	keeperID := c.Param("id")
	logger.Debugf("GET [GetKeeperData] Retrieving keeper with ID: %s", keeperID)

	keeperIDInt, err := strconv.ParseInt(keeperID, 10, 64)
	if err != nil {
		logger.Errorf("Error parsing keeper ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid keeper ID format",
			"code":  "INVALID_KEEPER_ID",
		})
		return
	}

	trackDBOp := metrics.TrackDBOperation("read", "keeper_data")
	keeperData, err := h.keeperRepository.GetByNonID(c.Request.Context(), "operator_id", keeperIDInt)
	trackDBOp(err)
	if err != nil || keeperData == nil {
		logger.Errorf("Error retrieving keeper data: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Keeper not found",
			"code":  "KEEPER_NOT_FOUND",
		})
		return
	}

	logger.Debugf("Successfully retrieved keeper with ID: %s", keeperID)
	c.JSON(http.StatusOK, keeperData)
}

func (h *Handler) GetKeeperTaskCount(c *gin.Context) {
	logger := h.getLogger(c)
	keeperID := c.Param("id")
	logger.Debugf("GET [GetKeeperTaskCount] For keeper with ID: %s", keeperID)

	keeperIDInt, err := strconv.ParseInt(keeperID, 10, 64)
	if err != nil {
		logger.Errorf("Error parsing keeper ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid keeper ID format",
			"code":  "INVALID_KEEPER_ID",
		})
		return
	}

	trackDBOp := metrics.TrackDBOperation("read", "keeper_tasks")
	keeper, err := h.keeperRepository.GetByNonID(c.Request.Context(), "operator_id", keeperIDInt)
	trackDBOp(err)
	if err != nil || keeper == nil {
		logger.Errorf("Error retrieving keeper: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Keeper not found",
			"code":  "KEEPER_NOT_FOUND",
		})
		return
	}

	taskCount := int64(0)
	if keeper.NoExecutedTasks != 0 {
		taskCount = keeper.NoExecutedTasks
	}

	logger.Debugf("Successfully retrieved task count %d for keeper ID: %s", taskCount, keeperID)
	c.JSON(http.StatusOK, gin.H{"no_executed_tasks": taskCount})
}

func (h *Handler) GetKeeperPoints(c *gin.Context) {
	logger := h.getLogger(c)
	keeperID := c.Param("id")
	logger.Debugf("GET [GetKeeperPoints] For keeper with ID: %s", keeperID)

	keeperIDInt, err := strconv.ParseInt(keeperID, 10, 64)
	if err != nil {
		logger.Errorf("Error parsing keeper ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trackDBOp := metrics.TrackDBOperation("read", "keeper_points")
	keeper, err := h.keeperRepository.GetByNonID(c.Request.Context(), "operator_id", keeperIDInt)
	trackDBOp(err)
	if err != nil || keeper == nil {
		logger.Errorf("Error retrieving keeper: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Keeper not found",
			"code":  "KEEPER_NOT_FOUND",
		})
		return
	}

	logger.Debugf("Successfully retrieved points for keeper ID: %s", keeperID)
	c.JSON(http.StatusOK, gin.H{"keeper_points": keeper.KeeperPoints})
}

func (h *Handler) GetKeeperCommunicationInfo(c *gin.Context) {
	logger := h.getLogger(c)
	keeperID := c.Param("id")
	logger.Debugf("GET [GetKeeperCommunicationInfo] For keeper with ID: %s", keeperID)

	keeperIDInt, err := strconv.ParseInt(keeperID, 10, 64)
	if err != nil {
		logger.Errorf("Error parsing keeper ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trackDBOp := metrics.TrackDBOperation("read", "keeper_communication")
	keeper, err := h.keeperRepository.GetByNonID(c.Request.Context(), "operator_id", keeperIDInt)
	trackDBOp(err)
	if err != nil || keeper == nil {
		logger.Errorf("Error retrieving keeper for ID %s: %v", keeperID, err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Keeper not found",
			"code":  "KEEPER_NOT_FOUND",
		})
		return
	}

	// Return communication info
	chatID := int64(0)
	if keeper.ChatID != 0 {
		chatID = keeper.ChatID
	}

	keeperData := map[string]interface{}{
		"chat_id":     chatID,
		"keeper_name": keeper.KeeperName,
		"email_id":    keeper.EmailID,
	}

	logger.Debugf("Successfully retrieved chat ID, keeper name, and email for ID: %s", keeperID)
	c.JSON(http.StatusOK, keeperData)
}
