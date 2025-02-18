package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (h *Handler) CreateKeeperData(w http.ResponseWriter, r *http.Request) {
	var keeperData types.KeeperData
	if err := json.NewDecoder(r.Body).Decode(&keeperData); err != nil {
		h.logger.Errorf("[CreateKeeperData] Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the maximum keeper ID from the database
	var maxKeeperID int64
	if err := h.db.Session().Query(`
		SELECT MAX(keeperID) FROM triggerx.keeper_data`).Scan(&maxKeeperID); err != nil {
		h.logger.Errorf("[CreateKeeperData] Error getting max keeper ID: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	keeperData.KeeperID = maxKeeperID + 1

	h.logger.Infof("[CreateKeeperData] Creating keeper with ID: %d", keeperData.KeeperID)
	if err := h.db.Session().Query(`
        INSERT INTO triggerx.keeper_data (
            keeper_id, keeper_address, keeper_type, registered_tx, 
            rewards_address, stakes, strategies, verified, status, consensus_keys, connection_address
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		keeperData.KeeperID, keeperData.KeeperAddress, keeperData.KeeperType,
		keeperData.RegisteredTx, keeperData.RewardsAddress, keeperData.Stakes,
		keeperData.Strategies, keeperData.Verified, keeperData.Status,
		keeperData.ConsensusKeys, keeperData.ConnectionAddress).Exec(); err != nil {
		h.logger.Errorf("[CreateKeeperData] Error creating keeper with ID %d: %v", keeperData.KeeperID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[CreateKeeperData] Successfully created keeper with ID: %d", keeperData.KeeperID)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(keeperData)
}

func (h *Handler) GetKeeperData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keeperID := vars["id"]
	h.logger.Infof("[GetKeeperData] Retrieving keeper with ID: %s", keeperID)

	var keeperData types.KeeperData
	if err := h.db.Session().Query(`
        SELECT keeper_id, keeper_address, keeper_type, rewards_address, stakes, strategies, 
               verified, registered_tx, status, consensus_keys, connection_address
        FROM triggerx.keeper_data 
        WHERE keeper_id = ? AND keeper_type = 2`, keeperID).Scan(
		&keeperData.KeeperID, &keeperData.KeeperAddress, &keeperData.KeeperType,
		&keeperData.RewardsAddress, &keeperData.Stakes, &keeperData.Strategies,
		&keeperData.Verified, &keeperData.RegisteredTx, &keeperData.Status,
		&keeperData.ConsensusKeys, &keeperData.ConnectionAddress); err != nil {
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
		&keeper.KeeperID, &keeper.KeeperAddress, &keeper.KeeperType,
		&keeper.RewardsAddress, &keeper.Stakes, &keeper.Strategies,
		&keeper.Verified, &keeper.RegisteredTx, &keeper.Status,
		&keeper.ConsensusKeys, &keeper.ConnectionAddress) {
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

func (h *Handler) UpdateKeeperData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keeperID := vars["id"]

	var keeperData types.KeeperData
	if err := json.NewDecoder(r.Body).Decode(&keeperData); err != nil {
		h.logger.Errorf("[UpdateKeeperData] Error decoding request body for keeper ID %s: %v", keeperID, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.logger.Infof("[UpdateKeeperData] Updating keeper with ID: %s", keeperID)
	if err := h.db.Session().Query(`
        UPDATE triggerx.keeper_data 
        SET rewards_address = ?, stakes = ?, strategies = ?, 
            verified = ?, status = ?, 
            consensus_keys = ?, connection_address = ?
        WHERE keeper_id = ?`,
		keeperData.RewardsAddress, keeperData.Stakes, keeperData.Strategies,
		keeperData.Verified, keeperData.Status,
		keeperData.ConsensusKeys, keeperData.ConnectionAddress,
		keeperID).Exec(); err != nil {
		h.logger.Errorf("[UpdateKeeperData] Error updating keeper with ID %s: %v", keeperID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[UpdateKeeperData] Successfully updated keeper with ID: %s", keeperID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(keeperData)
}
