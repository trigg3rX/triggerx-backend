package handlers

import (
	"net/http"
	"sort"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
)

func (h *Handler) CreateKeeperData(c *gin.Context) {
	var req types.CreateKeeperData
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error decoding request body"})
		return
	}
	if req.KeeperAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid keeper address"})
		return
	}

	existingKeeperID, err := h.keeperRepository.CheckKeeperExists(req.KeeperAddress)
	if err != nil {
		h.logger.Errorf("Error checking if keeper exists: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if existingKeeperID != -1 {
		h.logger.Infof(" Keeper already exists with ID: %d", existingKeeperID)
		c.JSON(http.StatusOK, gin.H{"message": "Keeper already exists"})
		return
	}

	currentKeeperID, err := h.keeperRepository.CreateKeeper(req)
	if err != nil {
		h.logger.Errorf("Error creating keeper: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// TODO: Send email to keeper
	// subject := "Welcome to TriggerX: Operator Whitelisting Confirmed"
	// emailBody := fmt.Sprintf(`
	// 	Hey %s,
	// 	<br><br>
	// 	Thanks for filling out the TriggerX whitelisting form — you're all set!
	// 	<br><br>
	// 	To stay updated on what's coming next (and not miss anything important), hop into our Telegram group:
	// 	<br><br>
	// 	<a href="https://t.me/+1I4euCfrchMxZDhl">https://t.me/+1I4euCfrchMxZDhl</a>
	// 	<br><br>
	// 	This is where we'll be sharing everything you need to know as an operator — from technical updates to rewards info and more.
	// 	<br><br>
	// 	See you there!
	// 	<br><br>
	// 	– Team TriggerX
	// `, req.KeeperName)

	// if err := h.sendEmailNotification(req.EmailID, subject, emailBody); err != nil {
	// 	h.logger.Errorf(" Error sending welcome email to keeper %s: %v", req.KeeperName, err)
	// 	// Note: We don't return here as the keeper creation was successful
	// } else {
	// 	h.logger.Infof(" Welcome email sent successfully to keeper %s at %s", req.KeeperName, req.EmailID)
	// }

	h.logger.Infof(" Successfully created keeper with ID: %d", currentKeeperID)

	c.JSON(http.StatusCreated, gin.H{"message": "Keeper created successfully"})
}

func (h *Handler) GetPerformers(c *gin.Context) {
	performers, err := h.keeperRepository.GetKeeperAsPerformer()
	if err != nil {
		h.logger.Errorf("[GetPerformers] Error retrieving performers: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	sort.Slice(performers, func(i, j int) bool {
		return performers[i].KeeperID < performers[j].KeeperID
	})

	h.logger.Infof("[GetPerformers] Successfully retrieved %d performers", len(performers))
	c.JSON(http.StatusOK, performers)
}

func (h *Handler) GetKeeperData(c *gin.Context) {
	keeperID := c.Param("id")
	h.logger.Infof("[GetKeeperData] Retrieving keeper with ID: %s", keeperID)

	keeperIDInt, err := strconv.ParseInt(keeperID, 10, 64)
	if err != nil {
		h.logger.Errorf("[GetKeeperData] Error parsing keeper ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid keeper ID format"})
		return
	}

	keeperData, err := h.keeperRepository.GetKeeperDataByID(keeperIDInt)
	if err != nil {
		h.logger.Errorf("[GetKeeperData] Error retrieving keeper data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("[GetKeeperData] Successfully retrieved keeper with ID: %s", keeperID)
	c.JSON(http.StatusOK, keeperData)
}

func (h *Handler) IncrementKeeperTaskCount(c *gin.Context) {
	keeperID := c.Param("id")
	h.logger.Infof("[IncrementKeeperTaskCount] Incrementing task count for keeper with ID: %s", keeperID)

	keeperIDInt, err := strconv.ParseInt(keeperID, 10, 64)
	if err != nil {
		h.logger.Errorf("[IncrementKeeperTaskCount] Error parsing keeper ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid keeper ID format"})
		return
	}

	newCount, err := h.keeperRepository.IncrementKeeperTaskCount(keeperIDInt)
	if err != nil {
		h.logger.Errorf("[IncrementKeeperTaskCount] Error incrementing task count: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("[IncrementKeeperTaskCount] Successfully incremented task count to %d for keeper ID: %s", newCount, keeperID)
	c.JSON(http.StatusOK, gin.H{"no_executed_tasks": newCount})
}

func (h *Handler) GetKeeperTaskCount(c *gin.Context) {
	keeperID := c.Param("id")
	h.logger.Infof("[GetKeeperTaskCount] Retrieving task count for keeper with ID: %s", keeperID)

	keeperIDInt, err := strconv.ParseInt(keeperID, 10, 64)
	if err != nil {
		h.logger.Errorf("[GetKeeperTaskCount] Error parsing keeper ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	taskCount, err := h.keeperRepository.GetKeeperTaskCount(keeperIDInt)
	if err != nil {
		h.logger.Errorf("[GetKeeperTaskCount] Error retrieving task count: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("[GetKeeperTaskCount] Successfully retrieved task count %d for keeper ID: %s", taskCount, keeperID)
	c.JSON(http.StatusOK, gin.H{"no_executed_tasks": taskCount})
}

func (h *Handler) AddTaskFeeToKeeperPoints(c *gin.Context) {
	keeperID := c.Param("id")
	keeperIDInt, err := strconv.ParseInt(keeperID, 10, 64)
	if err != nil {
		h.logger.Errorf("[AddTaskFeeToKeeperPoints] Error parsing keeper ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid keeper ID format"})
		return
	}

	var requestBody struct {
		TaskID int64 `json:"task_id"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		h.logger.Errorf("[AddTaskFeeToKeeperPoints] Error decoding request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	taskID := requestBody.TaskID
	h.logger.Infof("[AddTaskFeeToKeeperPoints] Processing task fee for task ID %d to keeper with ID: %s", taskID, keeperID)

	taskFee, err := h.taskRepository.GetTaskFee(taskID)
	if err != nil {
		h.logger.Errorf("[AddTaskFeeToKeeperPoints] Error retrieving task fee for task ID %d: %v", taskID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	newPoints, err := h.keeperRepository.UpdateKeeperPoints(keeperIDInt, taskFee)
	if err != nil {
		h.logger.Errorf("[AddTaskFeeToKeeperPoints] Error retrieving current points: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("[AddTaskFeeToKeeperPoints] Successfully added task fee %d from task ID %d to keeper ID: %s, new points: %d",
		taskFee, taskID, keeperID, newPoints)
	c.JSON(http.StatusOK, gin.H{
		"task_id":       taskID,
		"task_fee":      taskFee,
		"keeper_points": newPoints,
	})
}

func (h *Handler) GetKeeperPoints(c *gin.Context) {
	keeperID := c.Param("id")
	h.logger.Infof("[GetKeeperPoints] Retrieving points for keeper with ID: %s", keeperID)

	keeperIDInt, err := strconv.ParseInt(keeperID, 10, 64)
	if err != nil {
		h.logger.Errorf("[GetKeeperPoints] Error parsing keeper ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	points, err := h.keeperRepository.GetKeeperPointsByIDInDB(keeperIDInt)
	if err != nil {
		h.logger.Errorf("[GetKeeperPoints] Error retrieving keeper points: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("[GetKeeperPoints] Successfully retrieved points %d for keeper ID: %s", points, keeperID)
	c.JSON(http.StatusOK, gin.H{"keeper_points": points})
}

func (h *Handler) UpdateKeeperChatID(c *gin.Context) {
	var requestData types.UpdateKeeperChatIDRequest
	if err := c.ShouldBindJSON(&requestData); err != nil {
		h.logger.Errorf("[UpdateKeeperChatID] Error decoding request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error decoding request body"})
		return
	}

	if requestData.KeeperAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid keeper address"})
		return
	}

	err := h.keeperRepository.UpdateKeeperChatID(requestData.KeeperAddress, requestData.ChatID)
	if err != nil {
		h.logger.Errorf("[UpdateKeeperChatID] Error updating chat ID for keeper %s: %v", requestData.KeeperAddress, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("[UpdateKeeperChatID] Successfully updated chat ID for keeper: %s", requestData.KeeperAddress)
	c.JSON(http.StatusOK, gin.H{"message": "Chat ID updated successfully"})
}

func (h *Handler) GetKeeperCommunicationInfo(c *gin.Context) {
	keeperID := c.Param("id")
	h.logger.Infof("[GetKeeperChatInfo] Retrieving chat ID, keeper name, and email for keeper with ID: %s", keeperID)

	keeperIDInt, err := strconv.ParseInt(keeperID, 10, 64)
	if err != nil {
		h.logger.Errorf("[GetKeeperChatInfo] Error parsing keeper ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid keeper ID format"})
		return
	}

	keeperData, err := h.keeperRepository.GetKeeperCommunicationInfo(keeperIDInt)
	if err != nil {
		h.logger.Errorf("[GetKeeperChatInfo] Error retrieving chat ID, keeper name, and email for ID %s: %v", keeperID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("[GetKeeperChatInfo] Successfully retrieved chat ID, keeper name, and email for ID: %s", keeperID)
	c.JSON(http.StatusOK, keeperData)
}
