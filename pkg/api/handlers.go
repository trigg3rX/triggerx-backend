package api

import (
    "encoding/json"
    "net/http"
    "github.com/gorilla/mux"
    "github.com/trigg3rX/go-backend/pkg/database"
    "github.com/trigg3rX/go-backend/pkg/models"
    "time"
)

type Handler struct {
    db *database.Connection
}

func NewHandler(db *database.Connection) *Handler {
    return &Handler{db: db}
}

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

// Job Handlers
func (h *Handler) CreateJobData(w http.ResponseWriter, r *http.Request) {
    var jobData models.JobData
    if err := json.NewDecoder(r.Body).Decode(&jobData); err != nil {
        http.Error(w, "Error decoding request: "+err.Error(), http.StatusBadRequest)
        return
    }

    if jobData.TimeFrame == 0 {
        jobData.TimeFrame = time.Now().Unix()
    }

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