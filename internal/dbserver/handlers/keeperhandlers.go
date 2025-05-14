package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gocql/gocql"
	"github.com/gorilla/mux"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
	"gopkg.in/gomail.v2"
)

type NotificationConfig struct {
	EmailFrom     string
	EmailPassword string
	BotToken      string
}

func (h *Handler) CreateKeeperDataGoogleForm(w http.ResponseWriter, r *http.Request) {
	var keeperData types.GoogleFormCreateKeeperData
	if err := json.NewDecoder(r.Body).Decode(&keeperData); err != nil {
		h.logger.Errorf("Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
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
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		var maxKeeperID int64
		if err := h.db.Session().Query(`
			SELECT MAX(keeper_id) FROM triggerx.keeper_data`).Scan(&maxKeeperID); err != nil {
			h.logger.Errorf(" Error getting max keeper ID : %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
			http.Error(w, err.Error(), http.StatusInternalServerError)
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

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(keeperData)
}

func (h *Handler) GetKeeperData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keeperID := vars["id"]
	h.logger.Infof("[GetKeeperData] Retrieving keeper with ID: %s", keeperID)

	var keeperData types.KeeperData
	if err := h.db.Session().Query(`
        SELECT keeper_id, keeper_name, keeper_address, registered_tx, operator_id,
			rewards_address, rewards_booster, voting_power, keeper_points, connection_address,
			strategies, verified, status, online, version, no_exctask, chat_id, email_id
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
			WHERE keeper_id = 2
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

	iter := h.db.Session().Query(`
		SELECT keeper_id, keeper_name, keeper_address, registered_tx, operator_id,
		       rewards_address, rewards_booster, voting_power, keeper_points, connection_address,
		       strategies, verified, status, online, version, no_exctask, chat_id, email_id
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

func (h *Handler) IncrementKeeperTaskCount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keeperID := vars["id"]
	h.logger.Infof("[IncrementKeeperTaskCount] Incrementing task count for keeper with ID: %s", keeperID)

	var currentCount int
	if err := h.db.Session().Query(`
		SELECT no_exctask FROM triggerx.keeper_data WHERE keeper_id = ?`,
		keeperID).Scan(&currentCount); err != nil {
		h.logger.Errorf("[IncrementKeeperTaskCount] Error retrieving current task count: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	newCount := currentCount + 1

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

func (h *Handler) GetKeeperTaskCount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keeperID := vars["id"]
	h.logger.Infof("[GetKeeperTaskCount] Retrieving task count for keeper with ID: %s", keeperID)

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

func (h *Handler) AddTaskFeeToKeeperPoints(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keeperID := vars["id"]

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

	var taskFee int64
	if err := h.db.Session().Query(`
		SELECT task_fee FROM triggerx.task_data WHERE task_id = ?`,
		taskID).Scan(&taskFee); err != nil {
		h.logger.Errorf("[AddTaskFeeToKeeperPoints] Error retrieving task fee for task ID %d: %v", taskID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var currentPoints int64
	if err := h.db.Session().Query(`
		SELECT keeper_points FROM triggerx.keeper_data WHERE keeper_id = ?`,
		keeperID).Scan(&currentPoints); err != nil {
		h.logger.Errorf("[AddTaskFeeToKeeperPoints] Error retrieving current points: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	newPoints := currentPoints + taskFee

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

func (h *Handler) GetKeeperPoints(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keeperID := vars["id"]
	h.logger.Infof("[GetKeeperPoints] Retrieving points for keeper with ID: %s", keeperID)

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

func (h *Handler) checkAndNotifyOfflineKeeper(keeperID int64) {
	time.Sleep(10 * time.Minute)

	h.logger.Infof("[OfflineCheck] Checking current status for keeper ID: %s", keeperID)

	var online bool
	err := h.db.Session().Query(`
		SELECT online FROM triggerx.keeper_data WHERE keeper_id = ?`,
		keeperID).Scan(&online)

	if err != nil {
		h.logger.Errorf("[OfflineCheck] Error checking keeper online status: %v", err)
		return
	}

	if !online {
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

		if chatID != 0 {
			telegramMsg := fmt.Sprintf("Keeper %s is down for more than 10 minutes. Please check and start it.", keeperName)
			if err := h.sendTelegramNotification(chatID, telegramMsg); err != nil {
				h.logger.Errorf("[OfflineCheck] Failed to send Telegram notification to keeper %s: %v", keeperName, err)
			}
		} else {
			h.logger.Warn("[OfflineCheck] No Telegram chat ID found for keeper %s", keeperName)
		}

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

func (h *Handler) KeeperHealthCheckIn(w http.ResponseWriter, r *http.Request) {
	var keeperHealth types.UpdateKeeperHealth
	if err := json.NewDecoder(r.Body).Decode(&keeperHealth); err != nil {
		h.logger.Errorf("[KeeperHealthCheckIn] Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	keeperHealth.KeeperAddress = strings.ToLower(keeperHealth.KeeperAddress)

	if len(keeperHealth.KeeperAddress) > 0 && !bytes.HasPrefix([]byte(keeperHealth.KeeperAddress), []byte("0x")) {
		h.logger.Infof("[KeeperHealthCheckIn] Adding 0x prefix to keeper address: %s", keeperHealth.KeeperAddress)
		keeperHealth.KeeperAddress = "0x" + keeperHealth.KeeperAddress
	}

	var keeperID int64
	if err := h.db.Session().Query(`
		SELECT keeper_id FROM triggerx.keeper_data WHERE keeper_address = ? ALLOW FILTERING`,
		keeperHealth.KeeperAddress).Scan(&keeperID); err != nil {
		h.logger.Errorf("[KeeperHealthCheckIn] Error retrieving keeper_id for address %s: %v",
			keeperHealth.KeeperAddress, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if keeperID == 0 {
		h.logger.Errorf("[KeeperHealthCheckIn] No keeper found with address: %s", keeperHealth.KeeperAddress)
		http.Error(w, "Keeper not found", http.StatusNotFound)
		return
	}
	if keeperHealth.PeerID == "" {
		keeperHealth.PeerID = "no-peer-id"
	}

	h.logger.Infof("[KeeperHealthCheckIn] Keeper ID: %s | Online: %t", keeperID, keeperHealth.Active)

	var keeperPoints float64
	var isVerified bool
	var status bool
	if err := h.db.Session().Query(`
		SELECT keeper_points, verified, status FROM triggerx.keeper_data WHERE keeper_id = ?`,
		keeperID).Scan(&keeperPoints, &isVerified, &status); err != nil {
		h.logger.Errorf("[KeeperHealthCheckIn] Error checking keeper points: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if keeperPoints == 0 && isVerified && keeperHealth.Active && status {
		if err := h.db.Session().Query(`
			UPDATE triggerx.keeper_data 
			SET online = ?, peer_id = ?, keeper_points = ? 
			WHERE keeper_id = ?`,
			keeperHealth.Active, keeperHealth.PeerID, 10.0, keeperID).Exec(); err != nil {
			h.logger.Errorf("[KeeperHealthCheckIn] Error updating keeper status and points: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		h.logger.Infof("[KeeperHealthCheckIn] Added initial 10 points to keeper ID %d", keeperID)
	} else {
		if err := h.db.Session().Query(`
			UPDATE triggerx.keeper_data SET online = ?, peer_id = ? WHERE keeper_id = ?`,
			keeperHealth.Active, keeperHealth.PeerID, keeperID).Exec(); err != nil {
			h.logger.Errorf("[KeeperHealthCheckIn] Error updating keeper status: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if !keeperHealth.Active {
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

	var keeperID string
	if err := h.db.Session().Query(`
		SELECT keeper_id FROM triggerx.keeper_data 
		WHERE keeper_name = ?  ALLOW FILTERING`, requestData.KeeperName).Consistency(gocql.One).Scan(&keeperID); err != nil {
		h.logger.Errorf("[UpdateKeeperChatID] Error finding keeper ID for keeper %s: %v", requestData.KeeperName, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[UpdateKeeperChatID] Updating chat ID for keeper ID: %s", keeperID)

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
