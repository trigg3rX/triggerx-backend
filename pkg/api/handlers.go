package api

import (
	"encoding/json"
	"io"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"time"

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
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Creating user with ID: %d, Address: %s", userData.UserID, userData.UserAddress)

	// Convert stake amount to Gwei and store as varint
	stakeAmountGwei := new(big.Int)
	stakeAmountGwei = userData.StakeAmount

	if err := h.db.Session().Query(`
        INSERT INTO triggerx.user_data (user_id, user_address, job_ids, stake_amount)
        VALUES (?, ?, ?, ?)`,
		userData.UserID, userData.UserAddress, userData.JobIDs, stakeAmountGwei).Exec(); err != nil {
		log.Printf("Error inserting user data: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully created user with ID: %d and stake amount: %v Gwei", userData.UserID, stakeAmountGwei)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(userData)
}

func (h *Handler) GetUserData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]
	log.Printf("Handling GetUserData request for ID: %s", userID)

	var (
		userData struct {
			UserID      int64    `json:"user_id"`
			UserAddress string   `json:"user_address"`
			JobIDs      []int64  `json:"job_ids"`
			StakeAmount *big.Int `json:"-"` // Use big.Int for database interaction
		}
	)

	if err := h.db.Session().Query(`
        SELECT user_id, user_address, job_ids, stake_amount 
        FROM triggerx.user_data 
        WHERE user_id = ?`, userID).Scan(
		&userData.UserID, &userData.UserAddress, &userData.JobIDs, &userData.StakeAmount); err != nil {
		log.Printf("Error retrieving user data: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Retrieved user data for ID %s: %+v", userID, userData)

	// Convert big.Int to float64 for JSON response
	stakeAmountFloat := new(big.Float).SetInt(userData.StakeAmount)
	stakeAmountFloat64, _ := stakeAmountFloat.Float64()

	// Create a response struct with float64 stake amount
	response := struct {
		UserID      int64   `json:"user_id"`
		UserAddress string  `json:"user_address"`
		JobIDs      []int64 `json:"job_ids"`
		StakeAmount float64 `json:"stake_amount"`
	}{
		UserID:      userData.UserID,
		UserAddress: userData.UserAddress,
		JobIDs:      userData.JobIDs,
		StakeAmount: stakeAmountFloat64,
	}

	json.NewEncoder(w).Encode(response)
}

func (h *Handler) UpdateUserData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]
	log.Printf("Handling UpdateUserData request for ID: %s", userID)

	var userData models.UserData
	if err := json.NewDecoder(r.Body).Decode(&userData); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Updating user data: %+v", userData)

	if err := h.db.Session().Query(`
        UPDATE triggerx.user_data 
        SET user_address = ?, job_ids = ?, stake_amount = ?
        WHERE user_id = ?`,
		userData.UserAddress, userData.JobIDs, userData.StakeAmount, userID).Exec(); err != nil {
		log.Printf("Error updating user data: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully updated user with ID: %s", userID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(userData)
}

func (h *Handler) DeleteUserData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]
	log.Printf("Handling DeleteUserData request for ID: %s", userID)

	if err := h.db.Session().Query(`
        DELETE FROM triggerx.user_data 
        WHERE user_id = ?`, userID).Exec(); err != nil {
		log.Printf("Error deleting user data: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully deleted user with ID: %s", userID)
	w.WriteHeader(http.StatusNoContent)
}

// Job Handlers
func (h *Handler) CreateJobData(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request method: %s", r.Method)

	// Handle preflight
	if r.Method == http.MethodOptions {
		log.Printf("Handling preflight request")
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

	// Create a temporary struct to handle string chain_id
	type tempJobData struct {
		JobID             int64    `json:"job_id"`
		JobType           int64    `json:"jobType"`
		UserAddress       string   `json:"user_address"`
		ChainID           string   `json:"chain_id"` // Changed to string
		TimeFrame         int64    `json:"time_frame"`
		TimeInterval      int64    `json:"time_interval"`
		ContractAddress   string   `json:"contract_address"`
		TargetFunction    string   `json:"target_function"`
		ArgType           int64    `json:"arg_type"`
		Arguments         []string `json:"arguments"`
		Status            bool     `json:"status"`
		JobCostPrediction int64    `json:"job_cost_prediction"`
		ScriptFunction    string   `json:"script_function"`
		ScriptIpfsUrl     string   `json:"script_ipfs_url"`
		StakeAmount       float64  `json:"stake_amount"`
	}

	var tempJob tempJobData
	if err := json.Unmarshal(body, &tempJob); err != nil {
		log.Printf("Error decoding JSON: %v", err)
		http.Error(w, "Error decoding request: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Convert hex string to int64
	chainID, err := strconv.ParseInt(tempJob.ChainID[2:], 16, 64) // Remove "0x" prefix and parse as hex
	if err != nil {
		log.Printf("Error parsing chain_id: %v", err)
		http.Error(w, "Invalid chain_id format", http.StatusBadRequest)
		return
	}

	// Create the actual JobData struct
	jobData := models.JobData{
		JobID:             tempJob.JobID,
		JobType:           int(tempJob.JobType),
		UserAddress:       tempJob.UserAddress,
		ChainID:           int(chainID),
		TimeFrame:         tempJob.TimeFrame,
		TimeInterval:      int(tempJob.TimeInterval),
		ContractAddress:   tempJob.ContractAddress,
		TargetFunction:    tempJob.TargetFunction,
		ArgType:           int(tempJob.ArgType),
		Arguments:         tempJob.Arguments,
		Status:            tempJob.Status,
		JobCostPrediction: int(tempJob.JobCostPrediction),
		ScriptFunction:    tempJob.ScriptFunction,
		ScriptIpfsUrl:     tempJob.ScriptIpfsUrl,
		TimeCheck:         time.Now().UTC(),
	}

	log.Printf("Created job data: %+v", jobData)

	// Check if user exists by user_address
	var existingUserID int64
	var existingJobIDs []int64
	var existingStakeAmount *big.Int

	err = h.db.Session().Query(`
        SELECT user_id, job_ids, stake_amount 
        FROM triggerx.user_data 
        WHERE user_address = ? ALLOW FILTERING`,
		jobData.UserAddress).Scan(&existingUserID, &existingJobIDs, &existingStakeAmount)

	if err != nil && err != gocql.ErrNotFound {
		log.Printf("Error checking user existence: %v", err)
		http.Error(w, "Error checking user existence: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get new user ID if user doesn't exist
	if err == gocql.ErrNotFound {
		log.Printf("User not found, creating new user")
		var maxUserID int64
		if err := h.db.Session().Query(`
            SELECT MAX(user_id) FROM triggerx.user_data
        `).Scan(&maxUserID); err != nil && err != gocql.ErrNotFound {
			log.Printf("Error getting max user ID: %v", err)
			http.Error(w, "Error getting max user ID: "+err.Error(), http.StatusInternalServerError)
			return
		}
		existingUserID = maxUserID + 1

		// Convert stake amount to Gwei and store as varint
		stakeAmountGwei := new(big.Float).SetFloat64(tempJob.StakeAmount)
		stakeAmountInt, _ := stakeAmountGwei.Int(nil)

		if err := h.db.Session().Query(`
            INSERT INTO triggerx.user_data (
                user_id, user_address, job_ids, stake_amount
            ) VALUES (?, ?, ?, ?)`,
			existingUserID, jobData.UserAddress, []int64{jobData.JobID}, stakeAmountInt).Exec(); err != nil {
			log.Printf("Error creating user data: %v", err)
			http.Error(w, "Error creating user data: "+err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("Created new user with ID: %d and stake amount: %v Gwei", existingUserID, stakeAmountInt)
	} else {
		// Update existing user's job IDs and add to existing stake amount
		updatedJobIDs := append(existingJobIDs, jobData.JobID)

		// Convert new stake amount to big.Int and add to existing
		newStakeFloat := new(big.Float).SetFloat64(tempJob.StakeAmount)
		newStakeInt, _ := newStakeFloat.Int(nil)
		newStakeAmount := new(big.Int).Add(existingStakeAmount, newStakeInt)

		if err := h.db.Session().Query(`
            UPDATE triggerx.user_data 
            SET job_ids = ?, stake_amount = ?
            WHERE user_id = ?`,
			updatedJobIDs, newStakeAmount, existingUserID).Exec(); err != nil {
			log.Printf("Error updating user data: %v", err)
			http.Error(w, "Error updating user data: "+err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("Updated existing user data for user ID: %d, new stake amount: %v", existingUserID, newStakeAmount)
	}

	// Create the job
	if err := h.db.Session().Query(`
        INSERT INTO triggerx.job_data (
            job_id, jobType, user_id, chain_id, 
            time_frame, time_interval, contract_address, target_function, 
            arg_type, arguments, status, job_cost_prediction,
            script_function, script_ipfs_url, time_check, user_address
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		jobData.JobID, jobData.JobType, existingUserID, jobData.ChainID,
		jobData.TimeFrame, jobData.TimeInterval, jobData.ContractAddress,
		jobData.TargetFunction, jobData.ArgType, jobData.Arguments,
		jobData.Status, jobData.JobCostPrediction,
		jobData.ScriptFunction, jobData.ScriptIpfsUrl, jobData.TimeCheck, jobData.UserAddress).Exec(); err != nil {
		log.Printf("Error inserting job data: %v", err)
		http.Error(w, "Error inserting job data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully created job with ID: %d", jobData.JobID)

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
	log.Printf("Handling GetJobData request for ID: %s", jobID)

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
		log.Printf("Error retrieving job data: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Retrieved job data: %+v", jobData)
	json.NewEncoder(w).Encode(jobData)
}

func (h *Handler) GetAllJobs(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling GetAllJobs request")
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
		log.Printf("Error retrieving all jobs: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Retrieved %d jobs", len(jobs))
	json.NewEncoder(w).Encode(jobs)
}

func (h *Handler) UpdateJobData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["id"]
	log.Printf("Handling UpdateJobData request for ID: %s", jobID)

	var jobData models.JobData
	if err := json.NewDecoder(r.Body).Decode(&jobData); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Updating job data: %+v", jobData)

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
		log.Printf("Error updating job data: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully updated job with ID: %s", jobID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(jobData)
}

func (h *Handler) DeleteJobData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["id"]
	log.Printf("Handling DeleteJobData request for ID: %s", jobID)

	if err := h.db.Session().Query(`
        DELETE FROM triggerx.job_data 
        WHERE job_id = ?`, jobID).Exec(); err != nil {
		log.Printf("Error deleting job data: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully deleted job with ID: %s", jobID)
	w.WriteHeader(http.StatusNoContent)
}

// Get Latest JobID
func (h *Handler) GetLatestJobID(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling GetLatestJobID request")
	var latestJobID int64

	// Query to get the maximum job_id
	if err := h.db.Session().Query(`
        SELECT MAX(job_id) FROM triggerx.job_data
    `).Scan(&latestJobID); err != nil {
		if err == gocql.ErrNotFound {
			// If no jobs exist, start with job_id 1
			log.Printf("No jobs found, starting with job_id 0")
			latestJobID = 0
			json.NewEncoder(w).Encode(map[string]int64{"latest_job_id": latestJobID})
			return
		}
		log.Printf("Error fetching latest job ID: %v", err)
		http.Error(w, "Error fetching latest job ID: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Latest job ID: %d", latestJobID)
	// Return the latest job_id
	json.NewEncoder(w).Encode(map[string]int64{"latest_job_id": latestJobID})
}

func (h *Handler) GetJobsByUserID(w http.ResponseWriter, r *http.Request) {
    // Handle CORS preflight
    if r.Method == http.MethodOptions {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
        w.WriteHeader(http.StatusOK)
        return
    }

    // Set CORS headers
    w.Header().Set("Access-Control-Allow-Origin", "*")

    // Extract user ID from the URL path
    vars := mux.Vars(r)
    userIDStr := vars["user_id"]
    
    // Convert user ID to int64
    userID, err := strconv.ParseInt(userIDStr, 10, 64)
    if err != nil {
        log.Printf("Invalid user ID format: %v", err)
        http.Error(w, "Invalid user ID format", http.StatusBadRequest)
        return
    }
    
    log.Printf("Handling GetJobsByUserID request for user ID: %d", userID)

    // Struct to match exactly the fields we want to retrieve
    type JobSummary struct {
        JobID    int64 `json:"job_id"`
        JobType  int   `json:"jobType"`
        Status   bool  `json:"status"`
    }

    // Prepare a slice to store jobs for the user
    var userJobs []JobSummary

    // Query to fetch only job_id, jobType, and status for the specific user ID
    iter := h.db.Session().Query(`
        SELECT job_id, jobType, status 
        FROM triggerx.job_data 
        WHERE user_id = ? ALLOW FILTERING
    `, userID).Iter()

    var job JobSummary
    for iter.Scan(&job.JobID, &job.JobType, &job.Status) {
        userJobs = append(userJobs, job)
    }

    if err := iter.Close(); err != nil {
        log.Printf("Error retrieving jobs for user ID: %v", err)
        http.Error(w, "Error retrieving jobs: "+err.Error(), http.StatusInternalServerError)
        return
    }

    log.Printf("Retrieved %d jobs for user ID %d", len(userJobs), userID)

    // Set response headers
    w.Header().Set("Content-Type", "application/json")
    
    // Encode and send the response
    if err := json.NewEncoder(w).Encode(userJobs); err != nil {
        log.Printf("Error encoding response: %v", err)
        http.Error(w, "Error encoding response", http.StatusInternalServerError)
        return
    }
}

// Task Handlers
func (h *Handler) CreateTaskData(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling CreateTaskData request")
	var taskData models.TaskData
	if err := json.NewDecoder(r.Body).Decode(&taskData); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Creating task data: %+v", taskData)

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
		log.Printf("Error inserting task data: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully created task with ID: %d", taskData.TaskID)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(taskData)
}

func (h *Handler) GetTaskData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]
	log.Printf("Handling GetTaskData request for ID: %s", taskID)

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
		log.Printf("Error retrieving task data: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Retrieved task data: %+v", taskData)
	json.NewEncoder(w).Encode(taskData)
}

func (h *Handler) UpdateTaskData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]
	log.Printf("Handling UpdateTaskData request for ID: %s", taskID)

	var taskData models.TaskData
	if err := json.NewDecoder(r.Body).Decode(&taskData); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Updating task data: %+v", taskData)

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
		log.Printf("Error updating task data: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully updated task with ID: %s", taskID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(taskData)
}

func (h *Handler) DeleteTaskData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]
	log.Printf("Handling DeleteTaskData request for ID: %s", taskID)

	if err := h.db.Session().Query(`
        DELETE FROM triggerx.task_data 
        WHERE task_id = ?`, taskID).Exec(); err != nil {
		log.Printf("Error deleting task data: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully deleted task with ID: %s", taskID)
	w.WriteHeader(http.StatusNoContent)
}

// Quorum Handlers
func (h *Handler) CreateQuorumData(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling CreateQuorumData request")
	var quorumData models.QuorumData
	if err := json.NewDecoder(r.Body).Decode(&quorumData); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Creating quorum data: %+v", quorumData)

	if err := h.db.Session().Query(`
        INSERT INTO triggerx.quorum_data (quorum_id, quorum_no, quorum_creation_block, quorum_tx_hash, keepers, quorum_stake_total, quorum_threshold, task_ids)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		quorumData.QuorumID, quorumData.QuorumNo, quorumData.QuorumCreationBlock, quorumData.QuorumTxHash,
		quorumData.Keepers, quorumData.QuorumStakeTotal, quorumData.QuorumThreshold, quorumData.TaskIDs).Exec(); err != nil {
		log.Printf("Error inserting quorum data: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully created quorum with ID: %d", quorumData.QuorumID)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(quorumData)
}

func (h *Handler) GetQuorumData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	quorumID := vars["id"]
	log.Printf("Handling GetQuorumData request for ID: %s", quorumID)

	var quorumData models.QuorumData
	if err := h.db.Session().Query(`
        SELECT quorum_id, quorum_no, quorum_creation_block, quorum_tx_hash, keepers, quorum_stake_total, quorum_threshold, task_ids 
        FROM triggerx.quorum_data 
        WHERE quorum_id = ?`, quorumID).Scan(
		&quorumData.QuorumID, &quorumData.QuorumNo, &quorumData.QuorumCreationBlock, &quorumData.QuorumTxHash,
		&quorumData.Keepers, &quorumData.QuorumStakeTotal, &quorumData.QuorumThreshold, &quorumData.TaskIDs); err != nil {
		log.Printf("Error retrieving quorum data: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Retrieved quorum data: %+v", quorumData)
	json.NewEncoder(w).Encode(quorumData)
}

func (h *Handler) UpdateQuorumData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	quorumID := vars["id"]
	log.Printf("Handling UpdateQuorumData request for ID: %s", quorumID)

	var quorumData models.QuorumData
	if err := json.NewDecoder(r.Body).Decode(&quorumData); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Updating quorum data: %+v", quorumData)

	if err := h.db.Session().Query(`
        UPDATE triggerx.quorum_data 
        SET quorum_no = ?, quorum_creation_block = ?, quorum_tx_hash = ?, keepers = ?, quorum_stake_total = ?, quorum_threshold = ?, task_ids = ?
        WHERE quorum_id = ?`,
		quorumData.QuorumNo, quorumData.QuorumCreationBlock, quorumData.QuorumTxHash,
		quorumData.Keepers, quorumData.QuorumStakeTotal, quorumData.QuorumThreshold, quorumData.TaskIDs, quorumID).Exec(); err != nil {
		log.Printf("Error updating quorum data: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully updated quorum with ID: %s", quorumID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(quorumData)
}

func (h *Handler) DeleteQuorumData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	quorumID := vars["id"]
	log.Printf("Handling DeleteQuorumData request for ID: %s", quorumID)

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
