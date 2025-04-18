package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/gocql/gocql"
	"github.com/gorilla/mux"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
	"gopkg.in/gomail.v2"
)

// Add these new types for notification configuration
type NotificationConfig struct {
	EmailFrom     string
	EmailPassword string
	BotToken      string
}

func (h *Handler) HandleUpdateKeeperStatus(w http.ResponseWriter, r *http.Request) {
	// Extract keeper address from URL parameters
	vars := mux.Vars(r)
	keeperAddress := vars["address"]

	if keeperAddress == "" {
		http.Error(w, "Keeper address is required", http.StatusBadRequest)
		return
	}

	// Parse request body
	var statusUpdate types.KeeperStatusUpdate
	err := json.NewDecoder(r.Body).Decode(&statusUpdate)
	if err != nil {
		h.logger.Error("Failed to decode request body: " + err.Error())
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// First get the keeper_id using the keeper_address
	var keeperID int64
	err = h.db.Session().Query(`
    SELECT keeper_id FROM triggerx.keeper_data 
    WHERE keeper_address = ? ALLOW FILTERING`,
		keeperAddress).Scan(&keeperID)

	if err != nil {
		h.logger.Error("Failed to get keeper ID: " + err.Error())
		http.Error(w, "Database error while getting keeper ID", http.StatusInternalServerError)
		return
	}

	// Update keeper status in database using the keeper_id (primary key)
	err = h.db.Session().Query(`
    UPDATE triggerx.keeper_data 
    SET status = ? 
    WHERE keeper_id = ?`,
		statusUpdate.Status, keeperID).Exec()

	if err != nil {
		h.logger.Error("Failed to update keeper status: " + err.Error())
		http.Error(w, "Failed to update keeper status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Keeper status updated successfully",
	})
}

func (h *Handler) CreateKeeperData(w http.ResponseWriter, r *http.Request) {
	var keeperData types.CreateKeeperData
	if err := json.NewDecoder(r.Body).Decode(&keeperData); err != nil {
		h.logger.Errorf("[CreateKeeperData] Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the maximum keeper ID from the database
	var currentKeeperID int64
	if err := h.db.Session().Query(`
		SELECT keeper_id FROM triggerx.keeper_data WHERE keeper_address = ? ALLOW FILTERING`,
		keeperData.KeeperAddress).Scan(&currentKeeperID); err != nil {
		h.logger.Errorf("[CreateKeeperData] Error getting max keeper ID: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[CreateKeeperData] Updating keeper with ID: %d", currentKeeperID)
	if err := h.db.Session().Query(`
        UPDATE triggerx.keeper_data SET 
            registered_tx = ?, consensus_keys = ?, status = ?, chat_id = ?
        WHERE keeper_id = ?`,
		keeperData.RegisteredTx, keeperData.ConsensusKeys, true, keeperData.ChatID, currentKeeperID).Exec(); err != nil {
		h.logger.Errorf("[CreateKeeperData] Error creating keeper with ID %d: %v", currentKeeperID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[CreateKeeperData] Successfully updated keeper with ID: %d", currentKeeperID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(keeperData)
}

func (h *Handler) GoogleFormCreateKeeperData(w http.ResponseWriter, r *http.Request) {
	var keeperData types.GoogleFormCreateKeeperData
	if err := json.NewDecoder(r.Body).Decode(&keeperData); err != nil {
		h.logger.Errorf("[GoogleFormCreateKeeperData] Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the keeper_address already exists
	var existingKeeperID int64
	if err := h.db.Session().Query(`
		SELECT keeper_id FROM triggerx.keeper_data WHERE keeper_address = ? ALLOW FILTERING`,
		keeperData.KeeperAddress).Scan(&existingKeeperID); err == nil {
		h.logger.Warnf("[GoogleFormCreateKeeperData] Keeper with address %s already exists with ID: %d", keeperData.KeeperAddress, existingKeeperID)
		http.Error(w, "Keeper with this address already exists", http.StatusConflict)
		return
	}

	// Get the maximum keeper ID from the database
	var maxKeeperID int64
	if err := h.db.Session().Query(`
		SELECT MAX(keeper_id) FROM triggerx.keeper_data`).Scan(&maxKeeperID); err != nil {
		h.logger.Errorf("[GoogleFormCreateKeeperData] Error getting max keeper ID : %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	currentKeeperID := maxKeeperID + 1

	h.logger.Infof("[GoogleFormCreateKeeperData] Creating keeper with ID: %d", currentKeeperID)
	if err := h.db.Session().Query(`
        INSERT INTO triggerx.keeper_data (
            keeper_id, keeper_name, keeper_address, 
            rewards_address, no_exctask, keeper_points, verified, email_id
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		currentKeeperID, keeperData.KeeperName, keeperData.KeeperAddress,
		keeperData.RewardsAddress, 0, 0, true, keeperData.EmailID).Exec(); err != nil {
		h.logger.Errorf("[GoogleFormCreateKeeperData] Error creating keeper with ID %d: %v", currentKeeperID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[GoogleFormCreateKeeperData] Successfully created keeper with ID: %d", currentKeeperID)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(keeperData)
}

func (h *Handler) GetKeeperData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keeperID := vars["id"]
	h.logger.Infof("[GetKeeperData] Retrieving keeper with ID: %s", keeperID)

	var keeperData types.KeeperData
	if err := h.db.Session().Query(`
        SELECT keeper_id, keeper_address, rewards_address, stakes, strategies, 
               verified, registered_tx, status, consensus_keys, connection_address, 
               no_exctask, keeper_points, keeper_name, chat_id
        FROM triggerx.keeper_data 
        WHERE keeper_id = ?`, keeperID).Scan(
		&keeperData.KeeperID, &keeperData.KeeperAddress,
		&keeperData.RewardsAddress, &keeperData.Stakes, &keeperData.Strategies,
		&keeperData.Verified, &keeperData.RegisteredTx, &keeperData.Status,
		&keeperData.ConsensusKeys, &keeperData.ConnectionAddress,
		&keeperData.NoExcTask, &keeperData.KeeperPoints, &keeperData.KeeperName,
		&keeperData.ChatID); err != nil {
		h.logger.Errorf("[GetKeeperData] Error retrieving keeper with ID %s: %v", keeperID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[GetKeeperData] Successfully retrieved keeper with ID: %s", keeperID)
	json.NewEncoder(w).Encode(keeperData)
}

func (h *Handler) GetPerformers(w http.ResponseWriter, r *http.Request) {
	var performers []types.GetPerformerData
	iter := h.db.Session().Query(`SELECT keeper_id, keeper_address 
			FROM triggerx.keeper_data 
			WHERE verified = true AND status = true AND online = true
			ALLOW FILTERING`).Iter()

	var performer types.GetPerformerData
	for iter.Scan(
		&performer.KeeperID, &performer.KeeperAddress) {
		performers = append(performers, performer)
	}

	if err := iter.Close(); err != nil {
		h.logger.Errorf("[GetPerformers] Error retrieving performers: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if performers == nil {
		performers = []types.GetPerformerData{}
	}

	// Sort the results in memory after fetching them
	sort.Slice(performers, func(i, j int) bool {
		return performers[i].KeeperID < performers[j].KeeperID
	})

	w.Header().Set("Content-Type", "application/json")

	h.logger.Infof("[GetPerformers] Successfully retrieved %d performers", len(performers))

	jsonData, err := json.Marshal(performers)
	if err != nil {
		h.logger.Errorf("[GetPerformers] Error marshaling performers: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func (h *Handler) GetAllKeepers(w http.ResponseWriter, r *http.Request) {
	h.logger.Infof("[GetAllKeepers] Retrieving all keepers")
	var keepers []types.KeeperData

	// Explicitly name columns instead of using SELECT *
	iter := h.db.Session().Query(`
		SELECT keeper_id, keeper_address, registered_tx, 
		       rewards_address, stakes, strategies,
		       verified, status, consensus_keys,
		       connection_address, no_exctask, keeper_points, chat_id
		FROM triggerx.keeper_data`).Iter()

	var keeper types.KeeperData
	var tmpConsensusKeys, tmpStrategies []string
	var tmpStakes []float64

	for iter.Scan(
		&keeper.KeeperID, &keeper.KeeperAddress, &keeper.RegisteredTx,
		&keeper.RewardsAddress, &tmpStakes, &tmpStrategies,
		&keeper.Verified, &keeper.Status, &tmpConsensusKeys,
		&keeper.ConnectionAddress, &keeper.NoExcTask, &keeper.KeeperPoints,
		&keeper.ChatID) {

		// Make deep copies of the slices to avoid reference issues
		keeper.Stakes = make([]float64, len(tmpStakes))
		copy(keeper.Stakes, tmpStakes)

		keeper.Strategies = make([]string, len(tmpStrategies))
		copy(keeper.Strategies, tmpStrategies)

		keeper.ConsensusKeys = make([]string, len(tmpConsensusKeys))
		copy(keeper.ConsensusKeys, tmpConsensusKeys)

		keepers = append(keepers, keeper)
	}

	if err := iter.Close(); err != nil {
		h.logger.Errorf("[GetAllKeepers] Error retrieving keepers: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if keepers == nil {
		keepers = []types.KeeperData{}
	}

	jsonData, err := json.Marshal(keepers)
	if err != nil {
		h.logger.Errorf("[GetAllKeepers] Error marshaling keepers: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[GetAllKeepers] Successfully retrieved %d keepers", len(keepers))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func (h *Handler) UpdateKeeperConnectionData(w http.ResponseWriter, r *http.Request) {
	var keeperData types.UpdateKeeperConnectionData
	var response types.UpdateKeeperConnectionDataResponse
	if err := json.NewDecoder(r.Body).Decode(&keeperData); err != nil {
		h.logger.Errorf("[UpdateKeeperData] Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.logger.Infof("[UpdateKeeperData] Processing update for keeper with address: %s", keeperData.KeeperAddress)

	var keeperID int64
	if err := h.db.Session().Query(`
		SELECT keeper_id FROM triggerx.keeper_data WHERE keeper_address = ? ALLOW FILTERING`,
		keeperData.KeeperAddress).Scan(&keeperID); err != nil {
		h.logger.Errorf("[UpdateKeeperData] Error retrieving keeper ID for address %s: %v", keeperData.KeeperAddress, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[UpdateKeeperData] Found keeper with ID: %d, updating connection data", keeperID)

	if err := h.db.Session().Query(`
        UPDATE triggerx.keeper_data 
        SET connection_address = ?, verified = ?
        WHERE keeper_id = ?`,
		keeperData.ConnectionAddress, true, keeperID).Exec(); err != nil {
		h.logger.Errorf("[UpdateKeeperData] Error updating keeper with ID %d: %v", keeperID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response.KeeperID = keeperID
	response.KeeperAddress = keeperData.KeeperAddress
	response.Verified = true

	h.logger.Infof("[UpdateKeeperData] Successfully updated keeper with ID: %d", keeperID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// IncrementKeeperTaskCount increments the no_exctask counter for a keeper
func (h *Handler) IncrementKeeperTaskCount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keeperID := vars["id"]
	h.logger.Infof("[IncrementKeeperTaskCount] Incrementing task count for keeper with ID: %s", keeperID)

	// First get the current count
	var currentCount int
	if err := h.db.Session().Query(`
		SELECT no_exctask FROM triggerx.keeper_data WHERE keeper_id = ?`,
		keeperID).Scan(&currentCount); err != nil {
		h.logger.Errorf("[IncrementKeeperTaskCount] Error retrieving current task count: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Increment the count
	newCount := currentCount + 1

	// Update the database
	if err := h.db.Session().Query(`
		UPDATE triggerx.keeper_data SET no_exctask = ? WHERE keeper_id = ?`,
		newCount, keeperID).Exec(); err != nil {
		h.logger.Errorf("[IncrementKeeperTaskCount] Error updating task count: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[IncrementKeeperTaskCount] Successfully incremented task count to %d for keeper ID: %s", newCount, keeperID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]int{"no_exctask": newCount})
}

// GetKeeperTaskCount retrieves the no_exctask counter for a keeper
func (h *Handler) GetKeeperTaskCount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keeperID := vars["id"]
	h.logger.Infof("[GetKeeperTaskCount] Retrieving task count for keeper with ID: %s", keeperID)

	// Get the current count
	var taskCount int
	if err := h.db.Session().Query(`
		SELECT no_exctask FROM triggerx.keeper_data WHERE keeper_id = ?`,
		keeperID).Scan(&taskCount); err != nil {
		h.logger.Errorf("[GetKeeperTaskCount] Error retrieving task count: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[GetKeeperTaskCount] Successfully retrieved task count %d for keeper ID: %s", taskCount, keeperID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]int{"no_exctask": taskCount})
}

// AddTaskFeeToKeeperPoints adds the task fee from a specific task to the keeper's points
func (h *Handler) AddTaskFeeToKeeperPoints(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keeperID := vars["id"]

	// Parse the task ID from the request body
	var requestBody struct {
		TaskID int64 `json:"task_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		h.logger.Errorf("[AddTaskFeeToKeeperPoints] Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	taskID := requestBody.TaskID
	h.logger.Infof("[AddTaskFeeToKeeperPoints] Processing task fee for task ID %d to keeper with ID: %s", taskID, keeperID)

	// First get the task fee from the task_data table
	var taskFee int64
	if err := h.db.Session().Query(`
		SELECT task_fee FROM triggerx.task_data WHERE task_id = ?`,
		taskID).Scan(&taskFee); err != nil {
		h.logger.Errorf("[AddTaskFeeToKeeperPoints] Error retrieving task fee for task ID %d: %v", taskID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Then get the current keeper points
	var currentPoints int64
	if err := h.db.Session().Query(`
		SELECT keeper_points FROM triggerx.keeper_data WHERE keeper_id = ?`,
		keeperID).Scan(&currentPoints); err != nil {
		h.logger.Errorf("[AddTaskFeeToKeeperPoints] Error retrieving current points: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Add the task fee to the points
	newPoints := currentPoints + taskFee

	// Update the database
	if err := h.db.Session().Query(`
		UPDATE triggerx.keeper_data SET keeper_points = ? WHERE keeper_id = ?`,
		newPoints, keeperID).Exec(); err != nil {
		h.logger.Errorf("[AddTaskFeeToKeeperPoints] Error updating keeper points: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[AddTaskFeeToKeeperPoints] Successfully added task fee %d from task ID %d to keeper ID: %s, new points: %d",
		taskFee, taskID, keeperID, newPoints)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]int64{
		"task_id":       taskID,
		"task_fee":      taskFee,
		"keeper_points": newPoints,
	})
}

// GetKeeperPoints retrieves the keeper_points for a keeper
func (h *Handler) GetKeeperPoints(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keeperID := vars["id"]
	h.logger.Infof("[GetKeeperPoints] Retrieving points for keeper with ID: %s", keeperID)

	// Get the current points
	var points int64
	if err := h.db.Session().Query(`
		SELECT keeper_points FROM triggerx.keeper_data WHERE keeper_id = ?`,
		keeperID).Scan(&points); err != nil {
		h.logger.Errorf("[GetKeeperPoints] Error retrieving keeper points: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[GetKeeperPoints] Successfully retrieved points %d for keeper ID: %s", points, keeperID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]int64{"keeper_points": points})
}

// Add these new functions for notifications
func (h *Handler) sendTelegramNotification(chatID int64, message string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", h.config.BotToken)
	payload := map[string]interface{}{
		"chat_id": chatID,
		"text":    message,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		h.logger.Errorf("[Notification] Failed to marshal Telegram payload: %v", err)
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		h.logger.Errorf("[Notification] Failed to send Telegram message: %v", err)
		return err
	}
	defer resp.Body.Close()

	h.logger.Infof("[Notification] Telegram message sent successfully to chat ID: %d (Status: %d)", chatID, resp.StatusCode)
	return nil
}

func (h *Handler) sendEmailNotification(to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", h.config.EmailFrom)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer("smtp.zoho.in", 587, h.config.EmailFrom, h.config.EmailPassword)
	if err := d.DialAndSend(m); err != nil {
		h.logger.Errorf("[Notification] Failed to send email to %s: %v", to, err)
		return err
	}

	h.logger.Infof("[Notification] Email sent successfully to: %s", to)
	return nil
}

func (h *Handler) checkAndNotifyOfflineKeeper(keeperID string) {
	// Wait for 10 minutes
	time.Sleep(10 * time.Minute)

	h.logger.Infof("[OfflineCheck] Checking current status for keeper ID: %s", keeperID)

	// Check if keeper is still offline
	var online bool
	err := h.db.Session().Query(`
		SELECT online FROM triggerx.keeper_data WHERE keeper_id = ?`,
		keeperID).Scan(&online)

	if err != nil {
		h.logger.Errorf("[OfflineCheck] Error checking keeper online status: %v", err)
		return
	}

	if !online {
		// Fetch keeper communication info
		var chatID int64
		var keeperName, emailID string
		err := h.db.Session().Query(`
			SELECT chat_id, keeper_name, email_id 
			FROM triggerx.keeper_data 
			WHERE keeper_id = ?`,
			keeperID).Scan(&chatID, &keeperName, &emailID)

		if err != nil {
			h.logger.Errorf("[OfflineCheck] Error fetching keeper communication info: %v", err)
			return
		}

		// Send Telegram notification
		if chatID != 0 {
			telegramMsg := fmt.Sprintf("Keeper %s is down for more than 10 minutes. Please check and start it.", keeperName)
			if err := h.sendTelegramNotification(chatID, telegramMsg); err != nil {
				h.logger.Errorf("[OfflineCheck] Failed to send Telegram notification to keeper %s: %v", keeperName, err)
			}
		} else {
			h.logger.Warn("[OfflineCheck] No Telegram chat ID found for keeper %s", keeperName)
		}

		// Send email notification
		if emailID != "" {
			subject := fmt.Sprintf("TriggerX Keeper Down Alert - %s", keeperName)
			emailBody := fmt.Sprintf(`
				<h2>Keeper Update</h2>
				<p>This is a critical information from TriggerX. Your keeper <strong>%s</strong> has been down for more than 10 minutes. Please take action immediately.</p>
				<p>Regards,<br>TriggerX Team</p>
			`, keeperName)

			if err := h.sendEmailNotification(emailID, subject, emailBody); err != nil {
				h.logger.Errorf("[OfflineCheck] Failed to send email notification to keeper %s: %v", keeperName, err)
			}
		} else {
			h.logger.Warn("[OfflineCheck] No email address found for keeper %s", keeperName)
		}

		h.logger.Infof("[OfflineCheck] Completed notification process for offline keeper %s", keeperName)
	} else {
		h.logger.Infof("[OfflineCheck] Keeper %s is back online, no notifications needed", keeperID)
	}
}

// Modify the KeeperHealthCheckIn function
func (h *Handler) KeeperHealthCheckIn(w http.ResponseWriter, r *http.Request) {
	var keeperHealth types.UpdateKeeperHealth
	if err := json.NewDecoder(r.Body).Decode(&keeperHealth); err != nil {
		h.logger.Errorf("[KeeperHealthCheckIn] Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(keeperHealth.KeeperAddress) > 0 && !bytes.HasPrefix([]byte(keeperHealth.KeeperAddress), []byte("0x")) {
		h.logger.Infof("[KeeperHealthCheckIn] Adding 0x prefix to keeper address: %s", keeperHealth.KeeperAddress)
		keeperHealth.KeeperAddress = "0x" + keeperHealth.KeeperAddress
	}

	var keeperID string
	if err := h.db.Session().Query(`
		SELECT keeper_id FROM triggerx.keeper_data WHERE keeper_address = ? ALLOW FILTERING`,
		keeperHealth.KeeperAddress).Scan(&keeperID); err != nil {
		h.logger.Errorf("[KeeperHealthCheckIn] Error retrieving keeper_id for address %s: %v",
			keeperHealth.KeeperAddress, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if keeperID == "" {
		h.logger.Errorf("[KeeperHealthCheckIn] No keeper found with address: %s", keeperHealth.KeeperAddress)
		http.Error(w, "Keeper not found", http.StatusNotFound)
		return
	}
	if keeperHealth.PeerID == "" {
		keeperHealth.PeerID = "no-peer-id"
	}

	h.logger.Infof("[KeeperHealthCheckIn] Keeper ID: %s | Online: %t", keeperID, keeperHealth.Active)

	if keeperHealth.Version == "" {
		if err := h.db.Session().Query(`
			UPDATE triggerx.keeper_data SET online = ?, peer_id = ? WHERE keeper_id = ?`,
			keeperHealth.Active, keeperHealth.PeerID, keeperID).Exec(); err != nil {
			h.logger.Errorf("[KeeperHealthCheckIn] Error updating keeper status: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		if err := h.db.Session().Query(`
			UPDATE triggerx.keeper_data SET online = ?, version = ?, peer_id = ? WHERE keeper_id = ?`,
			keeperHealth.Active, keeperHealth.Version, keeperHealth.PeerID, keeperID).Exec(); err != nil {
			h.logger.Errorf("[KeeperHealthCheckIn] Error updating keeper status: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if !keeperHealth.Active {
		// Start a goroutine to check status after 10 minutes
		go h.checkAndNotifyOfflineKeeper(keeperID)
	}

	h.logger.Infof("[KeeperHealthCheckIn] Updated Keeper status for ID: %s", keeperID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(keeperHealth)
}

func (h *Handler) UpdateKeeperChatID(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		KeeperName string `json:"keeper_name"`
		ChatID     int64  `json:"chat_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		h.logger.Errorf("[UpdateKeeperChatID] Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.logger.Infof("[UpdateKeeperChatID] Finding keeper ID for keeper: %s", requestData.KeeperName)

	// Step 1: Find the keeper_id using keeper_name
	var keeperID string // Adjust the type based on your schema
	if err := h.db.Session().Query(`
		SELECT keeper_id FROM triggerx.keeper_data 
		WHERE keeper_name = ?  ALLOW FILTERING`, requestData.KeeperName).Consistency(gocql.One).Scan(&keeperID); err != nil {
		h.logger.Errorf("[UpdateKeeperChatID] Error finding keeper ID for keeper %s: %v", requestData.KeeperName, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[UpdateKeeperChatID] Updating chat ID for keeper ID: %s", keeperID)

	// Step 2: Update the chat_id for the specified keeper_id
	if err := h.db.Session().Query(`
		UPDATE triggerx.keeper_data 
		SET chat_id = ? 
		WHERE keeper_id = ?`,
		requestData.ChatID, keeperID).Exec(); err != nil {
		h.logger.Errorf("[UpdateKeeperChatID] Error updating chat ID for keeper ID %s: %v", keeperID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[UpdateKeeperChatID] Successfully updated chat ID for keeper: %s", requestData.KeeperName)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Chat ID updated successfully "})
}

func (h *Handler) GetKeeperCommunicationInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keeperID := vars["id"]
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[GetKeeperChatInfo] Successfully retrieved chat ID, keeper name, and email for ID: %s", keeperID)
	json.NewEncoder(w).Encode(keeperData)
}
