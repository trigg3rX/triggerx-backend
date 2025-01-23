package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

/* Status:
true: quorum is active
false: quorum is inactive
*/

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (h *Handler) CreateQuorumData(w http.ResponseWriter, r *http.Request) {
	var quorumData types.QuorumData
	if err := json.NewDecoder(r.Body).Decode(&quorumData); err != nil {
		h.logger.Error("[CreateQuorumData] Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.logger.Info("[CreateQuorumData] Creating quorum with ID: %d", quorumData.QuorumID)

	if err := h.db.Session().Query(`
        INSERT INTO triggerx.quorum_data (quorum_id, quorum_no, quorum_creation_block, quorum_tx_hash, keepers, quorum_stake_total, task_ids, quorum_status)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		quorumData.QuorumID, quorumData.QuorumNo, quorumData.QuorumCreationBlock, quorumData.QuorumTxHash,
		quorumData.Keepers, quorumData.QuorumStakeTotal, quorumData.TaskIDs, quorumData.QuorumStatus).Exec(); err != nil {
		h.logger.Error("[CreateQuorumData] Error creating quorum with ID %d: %v", quorumData.QuorumID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("[CreateQuorumData] Successfully created quorum with ID: %d", quorumData.QuorumID)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(quorumData)
}

func (h *Handler) GetQuorumData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	quorumID := vars["id"]
	h.logger.Info("[GetQuorumData] Retrieving quorum with ID: %s", quorumID)

	var quorumData types.QuorumData
	if err := h.db.Session().Query(`
        SELECT quorum_id, quorum_no, quorum_creation_block, quorum_tx_hash, keepers, quorum_stake_total, task_ids, quorum_status 
        FROM triggerx.quorum_data 
        WHERE quorum_id = ?`, quorumID).Scan(
		&quorumData.QuorumID, &quorumData.QuorumNo, &quorumData.QuorumCreationBlock, &quorumData.QuorumTxHash,
		&quorumData.Keepers, &quorumData.QuorumStakeTotal, &quorumData.TaskIDs, &quorumData.QuorumStatus); err != nil {
		h.logger.Error("[GetQuorumData] Error retrieving quorum with ID %s: %v", quorumID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("[GetQuorumData] Successfully retrieved quorum with ID: %s", quorumID)
	json.NewEncoder(w).Encode(quorumData)
}

func (h *Handler) GetQuorumNoForRegistration(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("[GetQuorumNoForRegistration] Finding optimal quorum for registration")

	var quorums []types.QuorumDataResponse
	iter := h.db.Session().Query(`
		SELECT quorum_id, quorum_no, quorum_status, quorum_stake_total, keepers
		FROM triggerx.quorum_data 
		ALLOW FILTERING`).Iter()

	var quorum types.QuorumDataResponse
	var keepers []string
	quorumMap := make(map[int]types.QuorumDataResponse)

	// Collect active quorums
	for iter.Scan(&quorum.QuorumID, &quorum.QuorumNo, &quorum.QuorumStatus, &quorum.QuorumStakeTotal, &keepers) {
		quorum.QuorumStrength = len(keepers)
		quorums = append(quorums, quorum)
		quorumMap[quorum.QuorumNo] = quorum
	}

	h.logger.Info("[GetQuorumNoForRegistration] Quorums: %v", quorums)

	if err := iter.Close(); err != nil {
		h.logger.Error("[GetQuorumNoForRegistration] Error retrieving quorums: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// First check if quorum strengths are balanced (within +-1)
	balanced := true
	baseStrength := quorums[0].QuorumStrength
	for _, q := range quorums {
		if abs(q.QuorumStrength-baseStrength) > 1 {
			balanced = false
			break
		}
	}

	var selectedQuorumNo int
	if !balanced {
		// Find quorum with minimum strength
		minStrength := quorums[0].QuorumStrength
		selectedQuorumNo = quorums[0].QuorumNo
		for _, q := range quorums {
			if q.QuorumStrength < minStrength {
				minStrength = q.QuorumStrength
				selectedQuorumNo = q.QuorumNo
			}
		}
	} else {
		// Quorum strengths are balanced, so balance by stake
		minStake := quorums[0].QuorumStakeTotal
		selectedQuorumNo = quorums[0].QuorumNo
		for _, q := range quorums {
			if q.QuorumStakeTotal < minStake {
				minStake = q.QuorumStakeTotal
				selectedQuorumNo = q.QuorumNo
			}
		}
	}

	h.logger.Info("[GetQuorumNoForRegistration] Selected quorum number: %d", selectedQuorumNo)
	json.NewEncoder(w).Encode(selectedQuorumNo)
}

func (h *Handler) GetAllQuorums(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("[GetAllQuorums] Retrieving all quorums")

	var quorums []types.QuorumDataResponse
	iter := h.db.Session().Query(`
		SELECT quorum_id, quorum_no, quorum_status, quorum_stake_total, keepers
		FROM triggerx.quorum_data 
		ALLOW FILTERING`).Iter()

	var quorum types.QuorumDataResponse
	var keepers []string
	for iter.Scan(&quorum.QuorumID, &quorum.QuorumNo, &quorum.QuorumStatus, &quorum.QuorumStakeTotal, &keepers) {
		quorum.QuorumStrength = len(keepers)
		quorums = append(quorums, quorum)
	}

	if err := iter.Close(); err != nil {
		h.logger.Error("[GetAllQuorums] Error retrieving quorums: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("[GetAllQuorums] Successfully retrieved %d quorums", len(quorums))
	json.NewEncoder(w).Encode(quorums)
}

func (h *Handler) UpdateQuorumData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	quorumID := vars["id"]

	var quorumData types.QuorumData
	if err := json.NewDecoder(r.Body).Decode(&quorumData); err != nil {
		h.logger.Error("[UpdateQuorumData] Error decoding request body for quorum ID %s: %v", quorumID, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.logger.Info("[UpdateQuorumData] Updating quorum with ID: %s", quorumID)

	if err := h.db.Session().Query(`
        UPDATE triggerx.quorum_data 
        SET quorum_no = ?, quorum_creation_block = ?, quorum_tx_hash = ?, keepers = ?, quorum_stake_total = ?, quorum_threshold = ?, task_ids = ?, quorum_status = ?
        WHERE quorum_id = ?`,
		quorumData.QuorumNo, quorumData.QuorumCreationBlock, quorumData.QuorumTxHash,
		quorumData.Keepers, quorumData.QuorumStakeTotal, quorumData.TaskIDs, quorumData.QuorumStatus, quorumID).Exec(); err != nil {
		h.logger.Error("[UpdateQuorumData] Error updating quorum with ID %s: %v", quorumID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("[UpdateQuorumData] Successfully updated quorum with ID: %s", quorumID)
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
