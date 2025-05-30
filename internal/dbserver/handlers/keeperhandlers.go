package handlers

import (
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (h *Handler) CreateKeeperDataGoogleForm(c *gin.Context) {
	var keeperData types.GoogleFormCreateKeeperData
	if err := c.ShouldBindJSON(&keeperData); err != nil {
		h.logger.Errorf("Error decoding request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	keeperData.KeeperAddress = strings.ToLower(keeperData.KeeperAddress)

	var existingKeeperID int64
	if err := h.db.Session().Query(`
		SELECT keeper_id FROM triggerx.keeper_data WHERE keeper_address = ? ALLOW FILTERING`,
		keeperData.KeeperAddress).Scan(&existingKeeperID); err == nil {
		h.logger.Infof(" Keeper already exists with ID: %d", existingKeeperID)
	}

	if existingKeeperID != 0 {
		if err := h.db.Session().Query(`
			UPDATE triggerx.keeper_data SET 
			keeper_name = ?, keeper_address = ?, rewards_address = ?, email_id = ?
			WHERE keeper_id = ?`,
			keeperData.KeeperName, keeperData.KeeperAddress, keeperData.RewardsAddress,
			keeperData.EmailID, existingKeeperID).Exec(); err != nil {
			h.logger.Errorf(" Error updating keeper with ID %d: %v", existingKeeperID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		var maxKeeperID int64
		if err := h.db.Session().Query(`
			SELECT MAX(keeper_id) FROM triggerx.keeper_data`).Scan(&maxKeeperID); err != nil {
			h.logger.Errorf(" Error getting max keeper ID : %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		currentKeeperID := maxKeeperID + 1
		var booster float32 = 1
		var rewards float64 = 0.0

		h.logger.Infof(" Creating keeper with ID: %d", currentKeeperID)
		if err := h.db.Session().Query(`
			INSERT INTO triggerx.keeper_data (
				keeper_id, keeper_name, keeper_address, rewards_booster,
				rewards_address, keeper_points, verified, email_id
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			currentKeeperID, keeperData.KeeperName, keeperData.KeeperAddress, booster,
			keeperData.RewardsAddress, rewards, true, keeperData.EmailID).Exec(); err != nil {
			h.logger.Errorf(" Error creating keeper with ID %d: %v", currentKeeperID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

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
		// `, keeperData.KeeperName)

		// if err := h.sendEmailNotification(keeperData.EmailID, subject, emailBody); err != nil {
		// 	h.logger.Errorf(" Error sending welcome email to keeper %s: %v", keeperData.KeeperName, err)
		// 	// Note: We don't return here as the keeper creation was successful
		// } else {
		// 	h.logger.Infof(" Welcome email sent successfully to keeper %s at %s", keeperData.KeeperName, keeperData.EmailID)
		// }

		h.logger.Infof(" Successfully created keeper with ID: %d", currentKeeperID)
	}

	c.JSON(http.StatusCreated, keeperData)
}

func (h *Handler) GetKeeperData(c *gin.Context) {
	keeperID := c.Param("id")
	h.logger.Infof("[GetKeeperData] Retrieving keeper with ID: %s", keeperID)

	var keeperData types.KeeperData
	if err := h.db.Session().Query(`
        SELECT keeper_id, keeper_name, keeper_address, registered_tx, operator_id,
			rewards_address, rewards_booster, voting_power, keeper_points, connection_address,
			strategies, verified, status, online, version, no_executed_tasks, chat_id, email_id
		FROM triggerx.keeper_data 
        WHERE keeper_id = ?`, keeperID).Scan(
		&keeperData.KeeperID, &keeperData.KeeperName, &keeperData.KeeperAddress,
		&keeperData.RegisteredTx, &keeperData.OperatorID,
		&keeperData.RewardsAddress, &keeperData.RewardsBooster, &keeperData.VotingPower,
		&keeperData.KeeperPoints, &keeperData.ConnectionAddress,
		&keeperData.Strategies, &keeperData.Verified, &keeperData.Status,
		&keeperData.Online, &keeperData.Version, &keeperData.NoExcTask,
		&keeperData.ChatID, &keeperData.EmailID); err != nil {
		h.logger.Errorf("[GetKeeperData] Error retrieving keeper with ID %s: %v", keeperID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("[GetKeeperData] Successfully retrieved keeper with ID: %s", keeperID)
	c.JSON(http.StatusOK, keeperData)
}

func (h *Handler) GetPerformers(c *gin.Context) {
	var performers []types.GetPerformerData
	iter := h.db.Session().Query(`SELECT keeper_id, keeper_address 
			FROM triggerx.keeper_data 
			WHERE keeper_id = 2
			ALLOW FILTERING`).Iter()

	var performer types.GetPerformerData
	for iter.Scan(
		&performer.KeeperID, &performer.KeeperAddress) {
		performers = append(performers, performer)
	}

	if err := iter.Close(); err != nil {
		h.logger.Errorf("[GetPerformers] Error retrieving performers: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if performers == nil {
		performers = []types.GetPerformerData{}
	}

	sort.Slice(performers, func(i, j int) bool {
		return performers[i].KeeperID < performers[j].KeeperID
	})

	h.logger.Infof("[GetPerformers] Successfully retrieved %d performers", len(performers))
	c.JSON(http.StatusOK, performers)
}

func (h *Handler) GetAllKeepers(c *gin.Context) {
	h.logger.Infof("[GetAllKeepers] Retrieving all keepers")
	var keepers []types.KeeperData

	iter := h.db.Session().Query(`
		SELECT keeper_id, keeper_name, keeper_address, registered_tx, operator_id,
		       rewards_address, rewards_booster, voting_power, keeper_points, connection_address,
		       strategies, verified, status, online, version, no_executed_tasks, chat_id, email_id
		FROM triggerx.keeper_data`).Iter()

	var keeper types.KeeperData
	var tmpStrategies []string

	for iter.Scan(
		&keeper.KeeperID, &keeper.KeeperName, &keeper.KeeperAddress, &keeper.RegisteredTx,
		&keeper.OperatorID, &keeper.RewardsAddress, &keeper.RewardsBooster, &keeper.VotingPower,
		&keeper.KeeperPoints, &keeper.ConnectionAddress, &tmpStrategies,
		&keeper.Verified, &keeper.Status, &keeper.Online, &keeper.Version, &keeper.NoExcTask,
		&keeper.ChatID, &keeper.EmailID) {

		keeper.Strategies = make([]string, len(tmpStrategies))
		copy(keeper.Strategies, tmpStrategies)

		keepers = append(keepers, keeper)
	}

	if err := iter.Close(); err != nil {
		h.logger.Errorf("[GetAllKeepers] Error retrieving keepers: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if keepers == nil {
		keepers = []types.KeeperData{}
	}

	h.logger.Infof("[GetAllKeepers] Successfully retrieved %d keepers", len(keepers))
	c.JSON(http.StatusOK, keepers)
}

func (h *Handler) IncrementKeeperTaskCount(c *gin.Context) {
	keeperID := c.Param("id")
	h.logger.Infof("[IncrementKeeperTaskCount] Incrementing task count for keeper with ID: %s", keeperID)

	var currentCount int
	if err := h.db.Session().Query(`
		SELECT no_executed_tasks FROM triggerx.keeper_data WHERE keeper_id = ?`,
		keeperID).Scan(&currentCount); err != nil {
		h.logger.Errorf("[IncrementKeeperTaskCount] Error retrieving current task count: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	newCount := currentCount + 1

	if err := h.db.Session().Query(`
		UPDATE triggerx.keeper_data SET no_executed_tasks = ? WHERE keeper_id = ?`,
		newCount, keeperID).Exec(); err != nil {
		h.logger.Errorf("[IncrementKeeperTaskCount] Error updating task count: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("[IncrementKeeperTaskCount] Successfully incremented task count to %d for keeper ID: %s", newCount, keeperID)
	c.JSON(http.StatusOK, gin.H{"no_executed_tasks": newCount})
}

func (h *Handler) GetKeeperTaskCount(c *gin.Context) {
	keeperID := c.Param("id")
	h.logger.Infof("[GetKeeperTaskCount] Retrieving task count for keeper with ID: %s", keeperID)

	var taskCount int
	if err := h.db.Session().Query(`
		SELECT no_executed_tasks FROM triggerx.keeper_data WHERE keeper_id = ?`,
		keeperID).Scan(&taskCount); err != nil {
		h.logger.Errorf("[GetKeeperTaskCount] Error retrieving task count: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("[GetKeeperTaskCount] Successfully retrieved task count %d for keeper ID: %s", taskCount, keeperID)
	c.JSON(http.StatusOK, gin.H{"no_executed_tasks": taskCount})
}

func (h *Handler) AddTaskFeeToKeeperPoints(c *gin.Context) {
	keeperID := c.Param("id")

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

	var taskFee int64
	if err := h.db.Session().Query(`
		SELECT task_fee FROM triggerx.task_data WHERE task_id = ?`,
		taskID).Scan(&taskFee); err != nil {
		h.logger.Errorf("[AddTaskFeeToKeeperPoints] Error retrieving task fee for task ID %d: %v", taskID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var currentPoints int64
	if err := h.db.Session().Query(`
		SELECT keeper_points FROM triggerx.keeper_data WHERE keeper_id = ?`,
		keeperID).Scan(&currentPoints); err != nil {
		h.logger.Errorf("[AddTaskFeeToKeeperPoints] Error retrieving current points: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	newPoints := currentPoints + taskFee

	if err := h.db.Session().Query(`
		UPDATE triggerx.keeper_data SET keeper_points = ? WHERE keeper_id = ?`,
		newPoints, keeperID).Exec(); err != nil {
		h.logger.Errorf("[AddTaskFeeToKeeperPoints] Error updating keeper points: %v", err)
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

	var points int64
	if err := h.db.Session().Query(`
		SELECT keeper_points FROM triggerx.keeper_data WHERE keeper_id = ?`,
		keeperID).Scan(&points); err != nil {
		h.logger.Errorf("[GetKeeperPoints] Error retrieving keeper points: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("[GetKeeperPoints] Successfully retrieved points %d for keeper ID: %s", points, keeperID)
	c.JSON(http.StatusOK, gin.H{"keeper_points": points})
}

func (h *Handler) UpdateKeeperChatID(c *gin.Context) {
	var requestData struct {
		KeeperName string `json:"keeper_name"`
		ChatID     int64  `json:"chat_id"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		h.logger.Errorf("[UpdateKeeperChatID] Error decoding request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("[UpdateKeeperChatID] Finding keeper ID for keeper: %s", requestData.KeeperName)

	var keeperID string
	if err := h.db.Session().Query(`
		SELECT keeper_id FROM triggerx.keeper_data 
		WHERE keeper_name = ?  ALLOW FILTERING`, requestData.KeeperName).Consistency(gocql.One).Scan(&keeperID); err != nil {
		h.logger.Errorf("[UpdateKeeperChatID] Error finding keeper ID for keeper %s: %v", requestData.KeeperName, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("[UpdateKeeperChatID] Updating chat ID for keeper ID: %s", keeperID)

	if err := h.db.Session().Query(`
		UPDATE triggerx.keeper_data 
		SET chat_id = ? 
		WHERE keeper_id = ?`,
		requestData.ChatID, keeperID).Exec(); err != nil {
		h.logger.Errorf("[UpdateKeeperChatID] Error updating chat ID for keeper ID %s: %v", keeperID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("[UpdateKeeperChatID] Successfully updated chat ID for keeper: %s", requestData.KeeperName)
	c.JSON(http.StatusOK, gin.H{"message": "Chat ID updated successfully"})
}

func (h *Handler) GetKeeperCommunicationInfo(c *gin.Context) {
	keeperID := c.Param("id")
	h.logger.Infof("[GetKeeperChatInfo] Retrieving chat ID, keeper name, and email for keeper with ID: %s", keeperID)

	var keeperData struct {
		ChatID     int64  `json:"chat_id"`
		KeeperName string `json:"keeper_name"`
		EmailID    string `json:"email_id"`
	}

	if err := h.db.Session().Query(`
        SELECT chat_id, keeper_name, email_id 
        FROM triggerx.keeper_data 
        WHERE keeper_id = ? ALLOW FILTERING`, keeperID).Scan(&keeperData.ChatID, &keeperData.KeeperName, &keeperData.EmailID); err != nil {
		h.logger.Errorf("[GetKeeperChatInfo] Error retrieving chat ID, keeper name, and email for ID %s: %v", keeperID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("[GetKeeperChatInfo] Successfully retrieved chat ID, keeper name, and email for ID: %s", keeperID)
	c.JSON(http.StatusOK, keeperData)
}
