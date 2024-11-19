package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/trigg3rX/go-backend/pkg/database"
	"github.com/trigg3rX/go-backend/pkg/models"
)

type Handler struct {
	db *database.Connection
}

func NewHandler(db *database.Connection) *Handler {
	return &Handler{db: db}
}

// User Handlers
// User Handlers
func (h *Handler) CreateUserData(w http.ResponseWriter, r *http.Request) {
	var userData models.UserData
	if err := json.NewDecoder(r.Body).Decode(&userData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.db.Session().Query(`
        INSERT INTO triggerx.user_data (user_id, user_address, job_ids)
        VALUES (?, ?, ?)`,
		userData.UserID, userData.UserAddress, userData.JobIDs).Exec(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(userData)
}

func (h *Handler) GetUserData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	var userData models.UserData
	if err := h.db.Session().Query(`
        SELECT user_id, user_address, job_ids 
        FROM triggerx.user_data 
        WHERE user_id = ?`, userID).Scan(
		&userData.UserID, &userData.UserAddress, &userData.JobIDs); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(userData)
}

func (h *Handler) UpdateUserData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	var userData models.UserData
	if err := json.NewDecoder(r.Body).Decode(&userData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.db.Session().Query(`
        UPDATE triggerx.user_data 
        SET user_address = ?, job_ids = ?
        WHERE user_id = ?`,
		userData.UserAddress, userData.JobIDs, userID).Exec(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(userData)
}

func (h *Handler) DeleteUserData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	if err := h.db.Session().Query(`
        DELETE FROM triggerx.user_data 
        WHERE user_id = ?`, userID).Exec(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Job Handlers
func (h *Handler) CreateJobData(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request method: %s", r.Method)

	// Handle preflight
	if r.Method == http.MethodOptions {
		return
	}

	log.Printf("Processing job creation request")

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	log.Printf("Received body: %s", string(body))

	var jobData models.JobData
	if err := json.Unmarshal(body, &jobData); err != nil {
		log.Printf("Error decoding JSON: %v", err)
		http.Error(w, "Error decoding request: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Rest of your handler code...
	log.Printf("api is calling........")
	if err := h.db.Session().Query(`
        INSERT INTO triggerx.job_data (
            job_id, jobType, user_id, chain_id, 
            time_frame, time_interval, contract_address, target_function, 
            arg_type, arguments, status, job_cost_prediction
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		jobData.JobID, jobData.JobType, jobData.UserID, jobData.ChainID,
		jobData.TimeFrame, jobData.TimeInterval, jobData.ContractAddress,
		jobData.TargetFunction, jobData.ArgType, jobData.Arguments,
		jobData.Status, jobData.JobCostPrediction).Exec(); err != nil {
		http.Error(w, "Error inserting data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(jobData)

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Return response
	response := map[string]interface{}{
		"message": "Job created successfully",
		"job":     jobData,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}

	log.Printf("Job created successfully")
}

func (h *Handler) GetJobData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["id"]

	var jobData models.JobData
	if err := h.db.Session().Query(`
        SELECT job_id, jobType, user_id, chain_id, time_frame, 
               time_interval, contract_address, target_function, 
               arg_type, arguments, status, job_cost_prediction
        FROM triggerx.job_data 
        WHERE job_id = ?`, jobID).Scan(
		&jobData.JobID, &jobData.JobType, &jobData.UserID, &jobData.ChainID,
		&jobData.TimeFrame, &jobData.TimeInterval, &jobData.ContractAddress,
		&jobData.TargetFunction, &jobData.ArgType, &jobData.Arguments,
		&jobData.Status, &jobData.JobCostPrediction); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(jobData)
}

func (h *Handler) GetAllJobs(w http.ResponseWriter, r *http.Request) {
	var jobs []models.JobData
	iter := h.db.Session().Query(`SELECT * FROM triggerx.job_data`).Iter()

	var job models.JobData
	for iter.Scan(
		&job.JobID, &job.JobType, &job.UserID, &job.ChainID,
		&job.TimeFrame, &job.TimeInterval, &job.ContractAddress,
		&job.TargetFunction, &job.ArgType, &job.Arguments,
		&job.Status, &job.JobCostPrediction) {
		jobs = append(jobs, job)
	}

	if err := iter.Close(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(jobs)
}

func (h *Handler) UpdateJobData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["id"]

	var jobData models.JobData
	if err := json.NewDecoder(r.Body).Decode(&jobData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.db.Session().Query(`
        UPDATE triggerx.job_data 
        SET jobType = ?, user_id = ?, chain_id = ?, 
            time_frame = ?, time_interval = ?, contract_address = ?,
            target_function = ?, arg_type = ?, arguments = ?,
            status = ?, job_cost_prediction = ?
        WHERE job_id = ?`,
		jobData.JobType, jobData.UserID, jobData.ChainID,
		jobData.TimeFrame, jobData.TimeInterval, jobData.ContractAddress,
		jobData.TargetFunction, jobData.ArgType, jobData.Arguments,
		jobData.Status, jobData.JobCostPrediction, jobID).Exec(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(jobData)
}

func (h *Handler) DeleteJobData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["id"]

	if err := h.db.Session().Query(`
        DELETE FROM triggerx.job_data 
        WHERE job_id = ?`, jobID).Exec(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Task Handlers
func (h *Handler) CreateTaskData(w http.ResponseWriter, r *http.Request) {
	var taskData models.TaskData
	if err := json.NewDecoder(r.Body).Decode(&taskData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(taskData)
}

func (h *Handler) GetTaskData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	var taskData models.TaskData
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(taskData)
}

func (h *Handler) GetAllTasks(w http.ResponseWriter, r *http.Request) {
	var tasks []models.TaskData
	iter := h.db.Session().Query(`SELECT * FROM triggerx.task_data`).Iter()

	var task models.TaskData
	for iter.Scan(
		&task.TaskID, &task.JobID, &task.TaskNo, &task.QuorumID,
		&task.QuorumNumber, &task.QuorumThreshold, &task.TaskCreatedBlock,
		&task.TaskCreatedTxHash, &task.TaskRespondedBlock, &task.TaskRespondedTxHash,
		&task.TaskHash, &task.TaskResponseHash, &task.QuorumKeeperHash) {
		tasks = append(tasks, task)
	}

	if err := iter.Close(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(tasks)
}

func (h *Handler) UpdateTaskData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	var taskData models.TaskData
	if err := json.NewDecoder(r.Body).Decode(&taskData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.db.Session().Query(`
        UPDATE triggerx.task_data 
        SET job_id = ?, task_no = ?, quorum_id = ?,
            quorum_number = ?, quorum_threshold = ?, task_created_block = ?,
            task_created_tx_hash = ?, task_responded_block = ?, task_responded_tx_hash = ?,
            task_hash = ?, task_response_hash = ?, quorum_keeper_hash = ?
        WHERE task_id = ?`,
		taskData.JobID, taskData.TaskNo, taskData.QuorumID,
		taskData.QuorumNumber, taskData.QuorumThreshold, taskData.TaskCreatedBlock,
		taskData.TaskCreatedTxHash, taskData.TaskRespondedBlock, taskData.TaskRespondedTxHash,
		taskData.TaskHash, taskData.TaskResponseHash, taskData.QuorumKeeperHash,
		taskID).Exec(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(taskData)
}

func (h *Handler) DeleteTaskData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	if err := h.db.Session().Query(`
        DELETE FROM triggerx.task_data 
        WHERE task_id = ?`, taskID).Exec(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Quorum Handlers
func (h *Handler) CreateQuorumData(w http.ResponseWriter, r *http.Request) {
	var quorumData models.QuorumData
	if err := json.NewDecoder(r.Body).Decode(&quorumData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.db.Session().Query(`
        INSERT INTO triggerx.quorum_data (quorum_id, quorum_no, quorum_creation_block, quorum_tx_hash, keepers, quorum_stake_total, quorum_threshold, task_ids)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		quorumData.QuorumID, quorumData.QuorumNo, quorumData.QuorumCreationBlock, quorumData.QuorumTxHash,
		quorumData.Keepers, quorumData.QuorumStakeTotal, quorumData.QuorumThreshold, quorumData.TaskIDs).Exec(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(quorumData)
}

func (h *Handler) GetQuorumData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	quorumID := vars["id"]

	var quorumData models.QuorumData
	if err := h.db.Session().Query(`
        SELECT quorum_id, quorum_no, quorum_creation_block, quorum_tx_hash, keepers, quorum_stake_total, quorum_threshold, task_ids 
        FROM triggerx.quorum_data 
        WHERE quorum_id = ?`, quorumID).Scan(
		&quorumData.QuorumID, &quorumData.QuorumNo, &quorumData.QuorumCreationBlock, &quorumData.QuorumTxHash,
		&quorumData.Keepers, &quorumData.QuorumStakeTotal, &quorumData.QuorumThreshold, &quorumData.TaskIDs); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(quorumData)
}

func (h *Handler) UpdateQuorumData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	quorumID := vars["id"]

	var quorumData models.QuorumData
	if err := json.NewDecoder(r.Body).Decode(&quorumData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.db.Session().Query(`
        UPDATE triggerx.quorum_data 
        SET quorum_no = ?, quorum_creation_block = ?, quorum_tx_hash = ?, keepers = ?, quorum_stake_total = ?, quorum_threshold = ?, task_ids = ?
        WHERE quorum_id = ?`,
		quorumData.QuorumNo, quorumData.QuorumCreationBlock, quorumData.QuorumTxHash,
		quorumData.Keepers, quorumData.QuorumStakeTotal, quorumData.QuorumThreshold, quorumData.TaskIDs, quorumID).Exec(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(quorumData)
}

func (h *Handler) DeleteQuorumData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	quorumID := vars["id"]

	if err := h.db.Session().Query(`
        DELETE FROM triggerx.quorum_data 
        WHERE quorum_id = ?`, quorumID).Exec(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Keeper Handlers
func (h *Handler) CreateKeeperData(w http.ResponseWriter, r *http.Request) {
	var keeperData models.KeeperData
	if err := json.NewDecoder(r.Body).Decode(&keeperData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.db.Session().Query(`
        INSERT INTO triggerx.keeper_data (keeper_id, withdrawal_address, stakes, strategies, verified, status, current_quorum_no, registered_block_no, register_tx_hash, connection_address, keystore_data)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		keeperData.KeeperID, keeperData.WithdrawalAddress, keeperData.Stakes, keeperData.Strategies,
		keeperData.Verified, keeperData.Status, keeperData.CurrentQuorumNo, keeperData.RegisteredBlockNo,
		keeperData.RegisterTxHash, keeperData.ConnectionAddress, keeperData.KeystoreData).Exec(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(keeperData)
}

func (h *Handler) GetKeeperData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keeperID := vars["id"]

	var keeperData models.KeeperData
	if err := h.db.Session().Query(`
        SELECT keeper_id, withdrawal_address, stakes, strategies, verified, status, current_quorum_no, registered_block_no, register_tx_hash, connection_address, keystore_data 
        FROM triggerx.keeper_data 
        WHERE keeper_id = ?`, keeperID).Scan(
		&keeperData.KeeperID, &keeperData.WithdrawalAddress, &keeperData.Stakes, &keeperData.Strategies,
		&keeperData.Verified, &keeperData.Status, &keeperData.CurrentQuorumNo, &keeperData.RegisteredBlockNo,
		&keeperData.RegisterTxHash, &keeperData.ConnectionAddress, &keeperData.KeystoreData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(keeperData)
}

func (h *Handler) UpdateKeeperData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keeperID := vars["id"]

	var keeperData models.KeeperData
	if err := json.NewDecoder(r.Body).Decode(&keeperData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.db.Session().Query(`
        UPDATE triggerx.keeper_data 
        SET withdrawal_address = ?, stakes = ?, strategies = ?, verified = ?, status = ?, current_quorum_no = ?, registered_block_no = ?, register_tx_hash = ?, connection_address = ?, keystore_data = ?
        WHERE keeper_id = ?`,
		keeperData.WithdrawalAddress, keeperData.Stakes, keeperData.Strategies, keeperData.Verified,
		keeperData.Status, keeperData.CurrentQuorumNo, keeperData.RegisteredBlockNo, keeperData.RegisterTxHash,
		keeperData.ConnectionAddress, keeperData.KeystoreData, keeperID).Exec(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(keeperData)
}

func (h *Handler) DeleteKeeperData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keeperID := vars["id"]

	if err := h.db.Session().Query(`
        DELETE FROM triggerx.keeper_data 
        WHERE keeper_id = ?`, keeperID).Exec(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Task History Handlers
func (h *Handler) CreateTaskHistory(w http.ResponseWriter, r *http.Request) {
	var taskHistory models.TaskHistory
	if err := json.NewDecoder(r.Body).Decode(&taskHistory); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.db.Session().Query(`
        INSERT INTO triggerx.task_history (task_id, quorum_id, keepers, responses, consensus_method, validation_status, tx_hash)
        VALUES (?, ?, ?, ?, ?, ?, ?)`,
		taskHistory.TaskID, taskHistory.QuorumID, taskHistory.Keepers, taskHistory.Responses,
		taskHistory.ConsensusMethod, taskHistory.ValidationStatus, taskHistory.TxHash).Exec(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(taskHistory)
}

func (h *Handler) GetTaskHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	var taskHistory models.TaskHistory
	if err := h.db.Session().Query(`
        SELECT task_id, quorum_id, keepers, responses, consensus_method, validation_status, tx_hash 
        FROM triggerx.task_history 
        WHERE task_id = ?`, taskID).Scan(
		&taskHistory.TaskID, &taskHistory.QuorumID, &taskHistory.Keepers, &taskHistory.Responses,
		&taskHistory.ConsensusMethod, &taskHistory.ValidationStatus, &taskHistory.TxHash); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(taskHistory)
}

func (h *Handler) UpdateTaskHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	var taskHistory models.TaskHistory
	if err := json.NewDecoder(r.Body).Decode(&taskHistory); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.db.Session().Query(`
        UPDATE triggerx.task_history 
        SET quorum_id = ?, keepers = ?, responses = ?, consensus_method = ?, validation_status = ?, tx_hash = ?
        WHERE task_id = ?`,
		taskHistory.QuorumID, taskHistory.Keepers, taskHistory.Responses,
		taskHistory.ConsensusMethod, taskHistory.ValidationStatus, taskHistory.TxHash, taskID).Exec(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(taskHistory)
}

func (h *Handler) DeleteTaskHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	if err := h.db.Session().Query(`
        DELETE FROM triggerx.task_history 
        WHERE task_id = ?`, taskID).Exec(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
