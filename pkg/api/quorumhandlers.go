package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (h *Handler) CreateQuorumData(w http.ResponseWriter, r *http.Request) {
	var quorumData types.QuorumData
	if err := json.NewDecoder(r.Body).Decode(&quorumData); err != nil {
		log.Printf("[CreateQuorumData] Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[CreateQuorumData] Creating quorum with ID: %d", quorumData.QuorumID)

	if err := h.db.Session().Query(`
        INSERT INTO triggerx.quorum_data (quorum_id, quorum_no, quorum_creation_block, quorum_tx_hash, keepers, quorum_stake_total, quorum_threshold, task_ids)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		quorumData.QuorumID, quorumData.QuorumNo, quorumData.QuorumCreationBlock, quorumData.QuorumTxHash,
		quorumData.Keepers, quorumData.QuorumStakeTotal, quorumData.QuorumThreshold, quorumData.TaskIDs).Exec(); err != nil {
		log.Printf("[CreateQuorumData] Error creating quorum with ID %d: %v", quorumData.QuorumID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[CreateQuorumData] Successfully created quorum with ID: %d", quorumData.QuorumID)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(quorumData)
}

func (h *Handler) GetQuorumData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	quorumID := vars["id"]
	log.Printf("[GetQuorumData] Retrieving quorum with ID: %s", quorumID)

	var quorumData types.QuorumData
	if err := h.db.Session().Query(`
        SELECT quorum_id, quorum_no, quorum_creation_block, quorum_tx_hash, keepers, quorum_stake_total, quorum_threshold, task_ids 
        FROM triggerx.quorum_data 
        WHERE quorum_id = ?`, quorumID).Scan(
		&quorumData.QuorumID, &quorumData.QuorumNo, &quorumData.QuorumCreationBlock, &quorumData.QuorumTxHash,
		&quorumData.Keepers, &quorumData.QuorumStakeTotal, &quorumData.QuorumThreshold, &quorumData.TaskIDs); err != nil {
		log.Printf("[GetQuorumData] Error retrieving quorum with ID %s: %v", quorumID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[GetQuorumData] Successfully retrieved quorum with ID: %s", quorumID)
	json.NewEncoder(w).Encode(quorumData)
}

func (h *Handler) UpdateQuorumData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	quorumID := vars["id"]

	var quorumData types.QuorumData
	if err := json.NewDecoder(r.Body).Decode(&quorumData); err != nil {
		log.Printf("[UpdateQuorumData] Error decoding request body for quorum ID %s: %v", quorumID, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[UpdateQuorumData] Updating quorum with ID: %s", quorumID)

	if err := h.db.Session().Query(`
        UPDATE triggerx.quorum_data 
        SET quorum_no = ?, quorum_creation_block = ?, quorum_tx_hash = ?, keepers = ?, quorum_stake_total = ?, quorum_threshold = ?, task_ids = ?
        WHERE quorum_id = ?`,
		quorumData.QuorumNo, quorumData.QuorumCreationBlock, quorumData.QuorumTxHash,
		quorumData.Keepers, quorumData.QuorumStakeTotal, quorumData.QuorumThreshold, quorumData.TaskIDs, quorumID).Exec(); err != nil {
		log.Printf("[UpdateQuorumData] Error updating quorum with ID %s: %v", quorumID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[UpdateQuorumData] Successfully updated quorum with ID: %s", quorumID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(quorumData)
}

// func (h *Handler) DeleteQuorumData(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	quorumID := vars["id"]
// 	log.Printf("[DeleteQuorumData] Deleting quorum with ID: %s", quorumID)

// 	if err := h.db.Session().Query(`
//         DELETE FROM triggerx.quorum_data
//         WHERE quorum_id = ?`, quorumID).Exec(); err != nil {
// 		log.Printf("[DeleteQuorumData] Error deleting quorum with ID %s: %v", quorumID, err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	log.Printf("[DeleteQuorumData] Successfully deleted quorum with ID: %s", quorumID)
// 	w.WriteHeader(http.StatusNoContent)
// }
