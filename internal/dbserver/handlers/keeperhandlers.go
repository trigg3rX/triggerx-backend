package handlers

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/gorilla/mux"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)



func (h *Handler)HandleUpdateKeeperStatus(w http.ResponseWriter, r *http.Request) {
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

	// Update keeper status in database
	err = h.db.Session().Query(`
		UPDATE triggerx.keeper_data 
		SET status = ? 
		WHERE keeper_address = ? `,
		statusUpdate.Status, keeperAddress).Exec()
	
	if err != nil {
		h.logger.Error("Failed to update keeper status: " + err.Error())
		http.Error(w, "Failed to update keeper status", http.StatusInternalServerError)
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
            registered_tx = ?, consensus_keys = ?, status = ?
        WHERE keeper_id = ?`,
		keeperData.RegisteredTx, keeperData.ConsensusKeys, true, currentKeeperID).Exec(); err != nil {
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
            rewards_address, no_exctask, keeper_points, verified
        ) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		currentKeeperID, keeperData.KeeperName, keeperData.KeeperAddress,
		keeperData.RewardsAddress, 0, 0, true).Exec(); err != nil {
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
               no_exctask, keeper_points, keeper_name
        FROM triggerx.keeper_data 
        WHERE keeper_id = ?`, keeperID).Scan(
		&keeperData.KeeperID, &keeperData.KeeperAddress,
		&keeperData.RewardsAddress, &keeperData.Stakes, &keeperData.Strategies,
		&keeperData.Verified, &keeperData.RegisteredTx, &keeperData.Status,
		&keeperData.ConsensusKeys, &keeperData.ConnectionAddress,
		&keeperData.NoExcTask, &keeperData.KeeperPoints, &keeperData.KeeperName); err != nil {
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
		       connection_address, no_exctask, keeper_points 
		FROM triggerx.keeper_data`).Iter()

	var keeper types.KeeperData
	var tmpConsensusKeys, tmpStrategies []string
	var tmpStakes []float64

	for iter.Scan(
		&keeper.KeeperID, &keeper.KeeperAddress, &keeper.RegisteredTx,
		&keeper.RewardsAddress, &tmpStakes, &tmpStrategies,
		&keeper.Verified, &keeper.Status, &tmpConsensusKeys,
		&keeper.ConnectionAddress, &keeper.NoExcTask, &keeper.KeeperPoints) {

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

func (h *Handler) KeeperHealthCheckIn(w http.ResponseWriter, r *http.Request) {
	var keeperHealth types.UpdateKeeperHealth
	if err := json.NewDecoder(r.Body).Decode(&keeperHealth); err != nil {
		h.logger.Errorf("[KeeperHealthCheckIn] Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
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

	h.logger.Infof("[KeeperHealthCheckIn] Keeper ID: %s | Online: %t", keeperID, keeperHealth.Active)

	if keeperHealth.Version == "" {
		if err := h.db.Session().Query(`
			UPDATE triggerx.keeper_data SET online = ? WHERE keeper_id = ?`,
			keeperHealth.Active, keeperID).Exec(); err != nil {
			h.logger.Errorf("[KeeperHealthCheckIn] Error updating keeper status: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		if err := h.db.Session().Query(`
			UPDATE triggerx.keeper_data SET online = ?, version = ? WHERE keeper_id = ?`,
			keeperHealth.Active, keeperHealth.Version, keeperID).Exec(); err != nil {
			h.logger.Errorf("[KeeperHealthCheckIn] Error updating keeper status: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	
	h.logger.Infof("[KeeperHealthCheckIn] Updated Keeper status for ID: %s", keeperID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(keeperHealth)
}
