package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

/*
	TODO:
		- Add GetKeepersByQuorumId
		- Add GetQuorumsByKeeperId
		- Add GetTasksByKeeperId
*/

func (h *Handler) CreateKeeperData(w http.ResponseWriter, r *http.Request) {
	var keeperData types.KeeperData
	if err := json.NewDecoder(r.Body).Decode(&keeperData); err != nil {
		log.Printf("[CreateKeeperData] Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[CreateKeeperData] Creating keeper with ID: %s", keeperData.KeeperID)
	if err := h.db.Session().Query(`
        INSERT INTO triggerx.keeper_data (
            keeper_id, withdrawal_address, stakes, strategies, 
            verified, current_quorum_no, registered_tx, status, 
            bls_signing_keys, connection_address
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		keeperData.KeeperID, keeperData.WithdrawalAddress, keeperData.Stakes,
		keeperData.Strategies, keeperData.Verified, keeperData.CurrentQuorumNo,
		keeperData.RegisteredTx, keeperData.Status, keeperData.BlsSigningKeys,
		keeperData.ConnectionAddress).Exec(); err != nil {
		log.Printf("[CreateKeeperData] Error creating keeper with ID %s: %v", keeperData.KeeperID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[CreateKeeperData] Successfully created keeper with ID: %s", keeperData.KeeperID)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(keeperData)
}

func (h *Handler) GetKeeperData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keeperID := vars["id"]
	log.Printf("[GetKeeperData] Retrieving keeper with ID: %s", keeperID)

	var keeperData types.KeeperData
	if err := h.db.Session().Query(`
        SELECT keeper_id, withdrawal_address, stakes, strategies, 
               verified, current_quorum_no, registered_tx, status, 
               bls_signing_keys, connection_address
        FROM triggerx.keeper_data 
        WHERE keeper_id = ?`, keeperID).Scan(
		&keeperData.KeeperID, &keeperData.WithdrawalAddress, &keeperData.Stakes,
		&keeperData.Strategies, &keeperData.Verified, &keeperData.CurrentQuorumNo,
		&keeperData.RegisteredTx, &keeperData.Status, &keeperData.BlsSigningKeys,
		&keeperData.ConnectionAddress); err != nil {
		log.Printf("[GetKeeperData] Error retrieving keeper with ID %s: %v", keeperID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[GetKeeperData] Successfully retrieved keeper with ID: %s", keeperID)
	json.NewEncoder(w).Encode(keeperData)
}

func (h *Handler) GetAllKeepers(w http.ResponseWriter, r *http.Request) {
	log.Printf("[GetAllKeepers] Retrieving all keepers")
	var keepers []types.KeeperData
	iter := h.db.Session().Query(`SELECT * FROM triggerx.keeper_data`).Iter()

	var keeper types.KeeperData
	for iter.Scan(
		&keeper.KeeperID, &keeper.WithdrawalAddress, &keeper.Stakes,
		&keeper.Strategies, &keeper.Verified, &keeper.CurrentQuorumNo,
		&keeper.RegisteredTx, &keeper.Status, &keeper.BlsSigningKeys,
		&keeper.ConnectionAddress) {
		keepers = append(keepers, keeper)
	}

	if err := iter.Close(); err != nil {
		log.Printf("[GetAllKeepers] Error retrieving keepers: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[GetAllKeepers] Successfully retrieved %d keepers", len(keepers))
	json.NewEncoder(w).Encode(keepers)
}

func (h *Handler) UpdateKeeperData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keeperID := vars["id"]

	var keeperData types.KeeperData
	if err := json.NewDecoder(r.Body).Decode(&keeperData); err != nil {
		log.Printf("[UpdateKeeperData] Error decoding request body for keeper ID %s: %v", keeperID, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[UpdateKeeperData] Updating keeper with ID: %s", keeperID)
	if err := h.db.Session().Query(`
        UPDATE triggerx.keeper_data 
        SET withdrawal_address = ?, stakes = ?, strategies = ?, 
            verified = ?, current_quorum_no = ?, registered_tx = ?, 
            status = ?, bls_signing_keys = ?, connection_address = ?
        WHERE keeper_id = ?`,
		keeperData.WithdrawalAddress, keeperData.Stakes, keeperData.Strategies,
		keeperData.Verified, keeperData.CurrentQuorumNo, keeperData.RegisteredTx,
		keeperData.Status, keeperData.BlsSigningKeys, keeperData.ConnectionAddress,
		keeperID).Exec(); err != nil {
		log.Printf("[UpdateKeeperData] Error updating keeper with ID %s: %v", keeperID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[UpdateKeeperData] Successfully updated keeper with ID: %s", keeperID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(keeperData)
}

// func (h *Handler) DeleteKeeperData(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	keeperID := vars["id"]
// 	log.Printf("[DeleteKeeperData] Deleting keeper with ID: %s", keeperID)

// 	if err := h.db.Session().Query(`
//         DELETE FROM triggerx.keeper_data
//         WHERE keeper_id = ?`, keeperID).Exec(); err != nil {
// 		log.Printf("[DeleteKeeperData] Error deleting keeper with ID %s: %v", keeperID, err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	log.Printf("[DeleteKeeperData] Successfully deleted keeper with ID: %s", keeperID)
// 	w.WriteHeader(http.StatusNoContent)
// }

func (h *Handler) CreateTaskHistory(w http.ResponseWriter, r *http.Request) {
	var taskHistory types.TaskHistory
	if err := json.NewDecoder(r.Body).Decode(&taskHistory); err != nil {
		log.Printf("[CreateTaskHistory] Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[CreateTaskHistory] Creating task history for task ID: %s", taskHistory.TaskID)
	if err := h.db.Session().Query(`
        INSERT INTO triggerx.task_history (task_id, quorum_id, keepers, responses, consensus_method, validation_status, tx_hash)
        VALUES (?, ?, ?, ?, ?, ?, ?)`,
		taskHistory.TaskID, taskHistory.QuorumID, taskHistory.Keepers, taskHistory.Responses,
		taskHistory.ConsensusMethod, taskHistory.ValidationStatus, taskHistory.TxHash).Exec(); err != nil {
		log.Printf("[CreateTaskHistory] Error creating task history for task ID %s: %v", taskHistory.TaskID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[CreateTaskHistory] Successfully created task history for task ID: %s", taskHistory.TaskID)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(taskHistory)
}

func (h *Handler) GetTaskHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]
	log.Printf("[GetTaskHistory] Retrieving task history for task ID: %s", taskID)

	var taskHistory types.TaskHistory
	if err := h.db.Session().Query(`
        SELECT task_id, quorum_id, keepers, responses, consensus_method, validation_status, tx_hash 
        FROM triggerx.task_history 
        WHERE task_id = ?`, taskID).Scan(
		&taskHistory.TaskID, &taskHistory.QuorumID, &taskHistory.Keepers, &taskHistory.Responses,
		&taskHistory.ConsensusMethod, &taskHistory.ValidationStatus, &taskHistory.TxHash); err != nil {
		log.Printf("[GetTaskHistory] Error retrieving task history for task ID %s: %v", taskID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[GetTaskHistory] Successfully retrieved task history for task ID: %s", taskID)
	json.NewEncoder(w).Encode(taskHistory)
}

// func (h *Handler) UpdateTaskHistory(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	taskID := vars["id"]

// 	var taskHistory models.TaskHistory
// 	if err := json.NewDecoder(r.Body).Decode(&taskHistory); err != nil {
// 		log.Printf("[UpdateTaskHistory] Error decoding request body for task ID %s: %v", taskID, err)
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	log.Printf("[UpdateTaskHistory] Updating task history for task ID: %s", taskID)
// 	if err := h.db.Session().Query(`
//         UPDATE triggerx.task_history
//         SET quorum_id = ?, keepers = ?, responses = ?, consensus_method = ?, validation_status = ?, tx_hash = ?
//         WHERE task_id = ?`,
// 		taskHistory.QuorumID, taskHistory.Keepers, taskHistory.Responses,
// 		taskHistory.ConsensusMethod, taskHistory.ValidationStatus, taskHistory.TxHash, taskID).Exec(); err != nil {
// 		log.Printf("[UpdateTaskHistory] Error updating task history for task ID %s: %v", taskID, err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	log.Printf("[UpdateTaskHistory] Successfully updated task history for task ID: %s", taskID)
// 	w.WriteHeader(http.StatusOK)
// 	json.NewEncoder(w).Encode(taskHistory)
// }

// func (h *Handler) DeleteTaskHistory(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	taskID := vars["id"]
// 	log.Printf("[DeleteTaskHistory] Deleting task history for task ID: %s", taskID)

// 	if err := h.db.Session().Query(`
//         DELETE FROM triggerx.task_history
//         WHERE task_id = ?`, taskID).Exec(); err != nil {
// 		log.Printf("[DeleteTaskHistory] Error deleting task history for task ID %s: %v", taskID, err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	log.Printf("[DeleteTaskHistory] Successfully deleted task history for task ID: %s", taskID)
// 	w.WriteHeader(http.StatusNoContent)
// }
