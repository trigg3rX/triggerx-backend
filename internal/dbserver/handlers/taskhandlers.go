package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	ttypes "github.com/trigg3rX/triggerx-backend/pkg/types"
)

/*
	TODO:
		- Add GetTasksByJobId
		- Add GetTasksByPerformerId
*/

func (h *Handler) CreateTaskData(w http.ResponseWriter, r *http.Request) {
	var taskData ttypes.CreateTaskData
	var taskResponse ttypes.CreateTaskResponse
	if err := json.NewDecoder(r.Body).Decode(&taskData); err != nil {
		h.logger.Errorf("[CreateTaskData] Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the next task ID
	var maxTaskID int64
	if err := h.db.Session().Query(`
		SELECT MAX(task_id) FROM triggerx.task_data`).Scan(&maxTaskID); err != nil {
		h.logger.Errorf("[CreateTaskData] Error getting max task ID: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	taskResponse.TaskID = maxTaskID + 1

	h.logger.Infof("[CreateTaskData] Creating task with ID: %d", taskResponse.TaskID)

	if err := h.db.Session().Query(`
        INSERT INTO triggerx.task_data (
            task_id, job_id, task_definition_id, created_at,
            task_performer_id, is_approved
        ) VALUES (?, ?, ?, ?, ?, ?)`,
		taskResponse.TaskID, taskData.JobID, taskData.TaskDefinitionID,
		time.Now(), taskData.TaskPerformerID, false).Exec(); err != nil {
		h.logger.Errorf("[CreateTaskData] Error inserting task with ID %d: %v", taskResponse.TaskID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	taskResponse.JobID = taskData.JobID
	taskResponse.TaskDefinitionID = taskData.TaskDefinitionID
	taskResponse.TaskPerformerID = taskData.TaskPerformerID
	taskResponse.IsApproved = true

	h.logger.Infof("[CreateTaskData] Successfully created task with ID: %d", taskResponse.TaskID)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(taskResponse)
}

func (h *Handler) GetTaskData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]
	h.logger.Infof("[GetTaskData] Fetching task with ID: %s", taskID)

	var taskData ttypes.TaskData
	if err := h.db.Session().Query(`
        SELECT task_id, job_id, task_definition_id, created_at,
               task_fee, execution_timestamp, execution_tx_hash, task_performer_id, 
			   proof_of_task, action_data_cid, task_attester_ids,
			   is_approved, tp_signature, ta_signature, task_submission_tx_hash,
			   is_successful
        FROM triggerx.task_data
        WHERE task_id = ?`, taskID).Scan(
		&taskData.TaskID, &taskData.JobID, &taskData.TaskDefinitionID, &taskData.CreatedAt, 
		&taskData.TaskFee, &taskData.ExecutionTimestamp, &taskData.ExecutionTxHash, &taskData.TaskPerformerID,
		&taskData.ProofOfTask, &taskData.ActionDataCID, &taskData.TaskAttesterIDs,
		&taskData.IsApproved, &taskData.TpSignature, &taskData.TaSignature,
		&taskData.TaskSubmissionTxHash, &taskData.IsSuccessful); err != nil {
		h.logger.Errorf("[GetTaskData] Error retrieving task with ID %s: %v", taskID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[GetTaskData] Successfully retrieved task with ID: %s", taskID)
	json.NewEncoder(w).Encode(taskData)
}
