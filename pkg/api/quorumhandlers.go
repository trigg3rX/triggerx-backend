package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

/* Status:
true: quorum is active
false: quorum is inactive
*/

func (h *Handler) CreateQuorumData(w http.ResponseWriter, r *http.Request) {
	var quorumData types.QuorumData
	if err := json.NewDecoder(r.Body).Decode(&quorumData); err != nil {
		log.Printf("[CreateQuorumData] Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[CreateQuorumData] Creating quorum with ID: %d", quorumData.QuorumID)

	if err := h.db.Session().Query(`
        INSERT INTO triggerx.quorum_data (quorum_id, quorum_no, quorum_creation_block, quorum_tx_hash, keepers, quorum_stake_total, quorum_threshold, task_ids, quorum_status)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		quorumData.QuorumID, quorumData.QuorumNo, quorumData.QuorumCreationBlock, quorumData.QuorumTxHash,
		quorumData.Keepers, quorumData.QuorumStakeTotal, quorumData.TaskIDs, quorumData.QuorumStatus).Exec(); err != nil {
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
        SELECT quorum_id, quorum_no, quorum_creation_block, quorum_tx_hash, keepers, quorum_stake_total, quorum_threshold, task_ids, quorum_status 
        FROM triggerx.quorum_data 
        WHERE quorum_id = ?`, quorumID).Scan(
		&quorumData.QuorumID, &quorumData.QuorumNo, &quorumData.QuorumCreationBlock, &quorumData.QuorumTxHash,
		&quorumData.Keepers, &quorumData.QuorumStakeTotal, &quorumData.TaskIDs, &quorumData.QuorumStatus); err != nil {
		log.Printf("[GetQuorumData] Error retrieving quorum with ID %s: %v", quorumID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[GetQuorumData] Successfully retrieved quorum with ID: %s", quorumID)
	json.NewEncoder(w).Encode(quorumData)
}

func (h *Handler) GetFreeQuorum(w http.ResponseWriter, r *http.Request) {
	log.Printf("[GetFreeQuorum] Retrieving quorums with status false")

	var quorumIDs []int64
	iter := h.db.Session().Query(`
		SELECT quorum_id 
		FROM triggerx.quorum_data 
		WHERE quorum_status = ? 
		ALLOW FILTERING`, false).Iter()

	var quorumID int64
	for iter.Scan(&quorumID) {
		quorumIDs = append(quorumIDs, quorumID)
	}

	if err := iter.Close(); err != nil {
		log.Printf("[GetFreeQuorum] Error retrieving free quorums: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[GetFreeQuorum] Successfully retrieved %d free quorums", len(quorumIDs))
	json.NewEncoder(w).Encode(map[string]interface{}{
		"free_quorum_ids": quorumIDs,
	})
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
        SET quorum_no = ?, quorum_creation_block = ?, quorum_tx_hash = ?, keepers = ?, quorum_stake_total = ?, quorum_threshold = ?, task_ids = ?, quorum_status = ?
        WHERE quorum_id = ?`,
		quorumData.QuorumNo, quorumData.QuorumCreationBlock, quorumData.QuorumTxHash,
		quorumData.Keepers, quorumData.QuorumStakeTotal, quorumData.TaskIDs, quorumData.QuorumStatus, quorumID).Exec(); err != nil {
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
