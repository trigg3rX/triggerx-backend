package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (h *Handler) CreateKeeperData(w http.ResponseWriter, r *http.Request) {
	var keeperData types.CreateKeeperData
	if err := json.NewDecoder(r.Body).Decode(&keeperData); err != nil {
		h.logger.Errorf("[CreateKeeperData] Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the maximum keeper ID from the database
	var maxKeeperID int64
	if err := h.db.Session().Query(`
		SELECT MAX(keeper_id) FROM triggerx.keeper_data`).Scan(&maxKeeperID); err != nil {
		h.logger.Errorf("[CreateKeeperData] Error getting max keeper ID: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	currentKeeperID := maxKeeperID + 1

	h.logger.Infof("[CreateKeeperData] Creating keeper with ID: %d", currentKeeperID)
	if err := h.db.Session().Query(`
        INSERT INTO triggerx.keeper_data (
            keeper_id, keeper_address, registered_tx, 
            rewards_address, 
            consensus_keys, no_exctask, keeper_points
        ) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		currentKeeperID, keeperData.KeeperAddress, keeperData.RegisteredTx,
		keeperData.RewardsAddress,
		keeperData.ConsensusKeys, 0, 0).Exec(); err != nil {
		h.logger.Errorf("[CreateKeeperData] Error creating keeper with ID %d: %v", currentKeeperID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[CreateKeeperData] Successfully created keeper with ID: %d", currentKeeperID)
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
               no_exctask, keeper_points
        FROM triggerx.keeper_data 
        WHERE keeper_id = ?`, keeperID).Scan(
		&keeperData.KeeperID, &keeperData.KeeperAddress,
		&keeperData.RewardsAddress, &keeperData.Stakes, &keeperData.Strategies,
		&keeperData.Verified, &keeperData.RegisteredTx, &keeperData.Status,
		&keeperData.ConsensusKeys, &keeperData.ConnectionAddress,
		&keeperData.NoExcTask, &keeperData.KeeperPoints); err != nil {
		h.logger.Errorf("[GetKeeperData] Error retrieving keeper with ID %s: %v", keeperID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[GetKeeperData] Successfully retrieved keeper with ID: %s", keeperID)
	json.NewEncoder(w).Encode(keeperData)
}

func (h *Handler) GetAllKeepers(w http.ResponseWriter, r *http.Request) {
	h.logger.Infof("[GetAllKeepers] Retrieving all keepers")
	var keepers []types.KeeperData
	iter := h.db.Session().Query(`SELECT * FROM triggerx.keeper_data`).Iter()

	var keeper types.KeeperData
	for iter.Scan(
		&keeper.KeeperID, &keeper.KeeperAddress,
		&keeper.RewardsAddress, &keeper.Stakes, &keeper.Strategies,
		&keeper.Verified, &keeper.RegisteredTx, &keeper.Status,
		&keeper.ConsensusKeys, &keeper.ConnectionAddress,
		&keeper.NoExcTask, &keeper.KeeperPoints) {
		keepers = append(keepers, keeper)
	}

	if err := iter.Close(); err != nil {
		h.logger.Errorf("[GetAllKeepers] Error retrieving keepers: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[GetAllKeepers] Successfully retrieved %d keepers", len(keepers))
	json.NewEncoder(w).Encode(keepers)
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
