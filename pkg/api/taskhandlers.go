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
		- Add GetTasksByJobId
		- Add GetTasksByQuorumId
*/

func (h *Handler) CreateTaskData(w http.ResponseWriter, r *http.Request) {
	var taskData types.TaskData
	if err := json.NewDecoder(r.Body).Decode(&taskData); err != nil {
		log.Printf("[CreateTaskData] Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[CreateTaskData] Creating task with ID: %d", taskData.TaskID)

	if err := h.db.Session().Query(`
        INSERT INTO triggerx.task_data (task_id, job_id, task_no, quorum_id,
            quorum_number, quorum_threshold, task_created_block, task_created_tx_hash,
            task_responded_block, task_responded_tx_hash, task_hash, 
            task_response_hash, quorum_keeper_hash)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		taskData.TaskID, taskData.JobID, taskData.TaskNo, taskData.QuorumID,
		taskData.QuorumNumber, taskData.QuorumThreshold, taskData.TaskCreatedBlock,
		taskData.TaskCreatedTxHash, taskData.TaskRespondedBlock, taskData.TaskRespondedTxHash,
		taskData.TaskHash, taskData.TaskResponseHash, taskData.QuorumKeeperHash).Exec(); err != nil {
		log.Printf("[CreateTaskData] Error inserting task with ID %d: %v", taskData.TaskID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[CreateTaskData] Successfully created task with ID: %d", taskData.TaskID)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(taskData)
}

func (h *Handler) GetTaskData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]
	log.Printf("[GetTaskData] Fetching task with ID: %s", taskID)

	var taskData types.TaskData
	if err := h.db.Session().Query(`
        SELECT task_id, job_id, task_no, quorum_id, quorum_number, 
               quorum_threshold, task_created_block, task_created_tx_hash,
               task_responded_block, task_responded_tx_hash, task_hash, 
               task_response_hash, quorum_keeper_hash
        FROM triggerx.task_data 
        WHERE task_id = ?`, taskID).Scan(
		&taskData.TaskID, &taskData.JobID, &taskData.TaskNo, &taskData.QuorumID,
		&taskData.QuorumNumber, &taskData.QuorumThreshold, &taskData.TaskCreatedBlock,
		&taskData.TaskCreatedTxHash, &taskData.TaskRespondedBlock, &taskData.TaskRespondedTxHash,
		&taskData.TaskHash, &taskData.TaskResponseHash, &taskData.QuorumKeeperHash); err != nil {
		log.Printf("[GetTaskData] Error retrieving task with ID %s: %v", taskID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[GetTaskData] Successfully retrieved task with ID: %s", taskID)
	json.NewEncoder(w).Encode(taskData)
}

// func (h *Handler) UpdateTaskData(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	taskID := vars["id"]

// 	var taskData models.TaskData
// 	if err := json.NewDecoder(r.Body).Decode(&taskData); err != nil {
// 		log.Printf("[UpdateTaskData] Error decoding request body for task ID %s: %v", taskID, err)
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	log.Printf("[UpdateTaskData] Updating task with ID: %s", taskID)

// 	if err := h.db.Session().Query(`
//         UPDATE triggerx.task_data
//         SET job_id = ?, task_no = ?, quorum_id = ?,
//             quorum_number = ?, quorum_threshold = ?, task_created_block = ?,
//             task_created_tx_hash = ?, task_responded_block = ?, task_responded_tx_hash = ?,
//             task_hash = ?, task_response_hash = ?, quorum_keeper_hash = ?
//         WHERE task_id = ?`,
// 		taskData.JobID, taskData.TaskNo, taskData.QuorumID,
// 		taskData.QuorumNumber, taskData.QuorumThreshold, taskData.TaskCreatedBlock,
// 		taskData.TaskCreatedTxHash, taskData.TaskRespondedBlock, taskData.TaskRespondedTxHash,
// 		taskData.TaskHash, taskData.TaskResponseHash, taskData.QuorumKeeperHash,
// 		taskID).Exec(); err != nil {
// 		log.Printf("[UpdateTaskData] Error updating task with ID %s: %v", taskID, err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	log.Printf("[UpdateTaskData] Successfully updated task with ID: %s", taskID)
// 	w.WriteHeader(http.StatusOK)
// 	json.NewEncoder(w).Encode(taskData)
// }

// func (h *Handler) DeleteTaskData(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	taskID := vars["id"]
// 	log.Printf("[DeleteTaskData] Deleting task with ID: %s", taskID)

// 	if err := h.db.Session().Query(`
//         DELETE FROM triggerx.task_data
//         WHERE task_id = ?`, taskID).Exec(); err != nil {
// 		log.Printf("[DeleteTaskData] Error deleting task with ID %s: %v", taskID, err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	log.Printf("[DeleteTaskData] Successfully deleted task with ID: %s", taskID)
// 	w.WriteHeader(http.StatusNoContent)
// }
