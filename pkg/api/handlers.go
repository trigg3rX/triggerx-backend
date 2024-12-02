package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/gocql/gocql"
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
        INSERT INTO triggerx.user_data (user_id, user_address, job_ids, stake_amount)
        VALUES (?, ?, ?, ?)`,
		userData.UserID, userData.UserAddress, userData.JobIDs, userData.StakeAmount).Exec(); err != nil {
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
        SELECT user_id, user_address, job_ids, stake_amount 
        FROM triggerx.user_data 
        WHERE user_id = ?`, userID).Scan(
		&userData.UserID, &userData.UserAddress, &userData.JobIDs, &userData.StakeAmount); err != nil {
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
        SET user_address = ?, job_ids = ?, stake_amount = ?
        WHERE user_id = ?`,
		userData.UserAddress, userData.JobIDs, userData.StakeAmount, userID).Exec(); err != nil {
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

	// Check if user exists by user_address
	var existingUserID int64
	var existingJobIDs []int64
	var existingStakeAmount float64
	
	err = h.db.Session().Query(`
        SELECT user_id, job_ids, stake_amount 
        FROM triggerx.user_data 
        WHERE user_address = ? ALLOW FILTERING`, 
        jobData.UserAddress).Scan(&existingUserID, &existingJobIDs, &existingStakeAmount)

	if err != nil && err != gocql.ErrNotFound {
		http.Error(w, "Error checking user existence: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get new user ID if user doesn't exist
	if err == gocql.ErrNotFound {
		var maxUserID int64
		if err := h.db.Session().Query(`
            SELECT MAX(user_id) FROM triggerx.user_data
        `).Scan(&maxUserID); err != nil && err != gocql.ErrNotFound {
			http.Error(w, "Error getting max user ID: "+err.Error(), http.StatusInternalServerError)
			return
		}
		existingUserID = maxUserID + 1
		existingJobIDs = []int64{}
		existingStakeAmount = 0
	}

	// Create the job
	if err := h.db.Session().Query(`
        INSERT INTO triggerx.job_data (
            job_id, jobType, user_id, chain_id, 
            time_frame, time_interval, contract_address, target_function, 
            arg_type, arguments, status, job_cost_prediction
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
        jobData.JobID, jobData.JobType, existingUserID, jobData.ChainID,
        jobData.TimeFrame, jobData.TimeInterval, jobData.ContractAddress,
        jobData.TargetFunction, jobData.ArgType, jobData.Arguments,
        jobData.Status, jobData.JobCostPrediction).Exec(); err != nil {
		http.Error(w, "Error inserting job data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Update or create user data
	updatedJobIDs := append(existingJobIDs, jobData.JobID)
	if err == gocql.ErrNotFound {
		// Create new user
		if err := h.db.Session().Query(`
            INSERT INTO triggerx.user_data (
                user_id, user_address, job_ids, stake_amount
            ) VALUES (?, ?, ?, ?)`,
            existingUserID, jobData.UserAddress, updatedJobIDs, existingStakeAmount).Exec(); err != nil {
			http.Error(w, "Error creating user data: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// Update existing user
		if err := h.db.Session().Query(`
            UPDATE triggerx.user_data 
            SET job_ids = ?
            WHERE user_id = ?`,
            updatedJobIDs, existingUserID).Exec(); err != nil {
			http.Error(w, "Error updating user data: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

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

// Get Latest JobID
func (h *Handler) GetLatestJobID(w http.ResponseWriter, r *http.Request) {
    var latestJobID int64

    // Query to get the maximum job_id
    if err := h.db.Session().Query(`
        SELECT MAX(job_id) FROM triggerx.job_data
    `).Scan(&latestJobID); err != nil {
        if err == gocql.ErrNotFound {
            // If no jobs exist, start with job_id 1
            latestJobID = 0
            json.NewEncoder(w).Encode(map[string]int64{"latest_job_id": latestJobID})
            return
        }
        http.Error(w, "Error fetching latest job ID: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // Return the latest job_id
    json.NewEncoder(w).Encode(map[string]int64{"latest_job_id": latestJobID})
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
        INSERT INTO triggerx.keeper_data (
            keeper_id, withdrawal_address, stakes, strategies, 
            verified, current_quorum_no, registered_tx, status, 
            bls_signing_keys, connection_address
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
        keeperData.KeeperID, keeperData.WithdrawalAddress, keeperData.Stakes,
        keeperData.Strategies, keeperData.Verified, keeperData.CurrentQuorumNo,
        keeperData.RegisteredTx, keeperData.Status, keeperData.BlsSigningKeys,
        keeperData.ConnectionAddress).Exec(); err != nil {
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
        SELECT keeper_id, withdrawal_address, stakes, strategies, 
               verified, current_quorum_no, registered_tx, status, 
               bls_signing_keys, connection_address
        FROM triggerx.keeper_data 
        WHERE keeper_id = ?`, keeperID).Scan(
        &keeperData.KeeperID, &keeperData.WithdrawalAddress, &keeperData.Stakes,
        &keeperData.Strategies, &keeperData.Verified, &keeperData.CurrentQuorumNo,
        &keeperData.RegisteredTx, &keeperData.Status, &keeperData.BlsSigningKeys,
        &keeperData.ConnectionAddress); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(keeperData)
}

func (h *Handler) GetAllKeepers(w http.ResponseWriter, r *http.Request) {
    var keepers []models.KeeperData
    iter := h.db.Session().Query(`SELECT * FROM triggerx.keeper_data`).Iter()

    var keeper models.KeeperData
    for iter.Scan(
        &keeper.KeeperID, &keeper.WithdrawalAddress, &keeper.Stakes,
        &keeper.Strategies, &keeper.Verified, &keeper.CurrentQuorumNo,
        &keeper.RegisteredTx, &keeper.Status, &keeper.BlsSigningKeys,
        &keeper.ConnectionAddress) {
        keepers = append(keepers, keeper)
    }

    if err := iter.Close(); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(keepers)
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
        SET withdrawal_address = ?, stakes = ?, strategies = ?, 
            verified = ?, current_quorum_no = ?, registered_tx = ?, 
            status = ?, bls_signing_keys = ?, connection_address = ?
        WHERE keeper_id = ?`,
        keeperData.WithdrawalAddress, keeperData.Stakes, keeperData.Strategies,
        keeperData.Verified, keeperData.CurrentQuorumNo, keeperData.RegisteredTx,
        keeperData.Status, keeperData.BlsSigningKeys, keeperData.ConnectionAddress,
        keeperID).Exec(); err != nil {
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

