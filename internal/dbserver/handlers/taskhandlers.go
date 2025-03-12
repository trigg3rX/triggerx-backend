package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/client"
	"github.com/gorilla/mux"

	"github.com/docker/docker/api/types"
	"github.com/trigg3rX/triggerx-backend/pkg/resources"
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

func (h *Handler) GetTaskFees(w http.ResponseWriter, r *http.Request) {
	// Get the IPFS URLs from the query parameter
	ipfsURLs := r.URL.Query().Get("ipfs_url") // Get the single string of URLs
	if ipfsURLs == "" {
		http.Error(w, "Missing ipfs_url query parameter", http.StatusBadRequest)
		return
	}

	h.logger.Infof("[GetTaskFees] IPFS URLs: %s", ipfsURLs)

	// Split the IPFS URLs by comma
	urlList := strings.Split(ipfsURLs, ",")
	totalFee := 0.0
	var mu sync.Mutex // Mutex to protect totalFee
	var wg sync.WaitGroup // WaitGroup to wait for all goroutines to finish

	ctx := context.Background()

	// Create Docker client
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		h.logger.Errorf("[GetTaskFees] Error creating Docker client: %v", err)
		http.Error(w, "Failed to create Docker client", http.StatusInternalServerError)
		return
	}
	defer cli.Close()

	// Process each IPFS URL concurrently
	for _, ipfsURL := range urlList {
		ipfsURL = strings.TrimSpace(ipfsURL) // Trim any whitespace
		wg.Add(1) // Increment the WaitGroup counter

		go func(url string) {
			defer wg.Done() // Decrement the counter when the goroutine completes
			h.logger.Infof("[GetTaskFees] Starting processing for URL: %s", url) // Log when goroutine starts

			// Download and process the IPFS file
			codePath, err := resources.DownloadIPFSFile(url)
			if err != nil {
				h.logger.Errorf("[GetTaskFees] Error downloading IPFS file for URL %s: %v", url, err)
				return // Exit the goroutine on error
			}
			defer os.RemoveAll(filepath.Dir(codePath))

			// Create container
			containerID, err := resources.CreateDockerContainer(ctx, cli, codePath)
			if err != nil {
				h.logger.Errorf("[GetTaskFees] Error creating container for URL %s: %v", url, err)
				return // Exit the goroutine on error
			}
			defer cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{Force: true})

			// Monitor resources and get stats
			stats, err := resources.MonitorResources(ctx, cli, containerID)
			if err != nil {
				h.logger.Errorf("[GetTaskFees] Error monitoring resources for URL %s: %v", url, err)
				return // Exit the goroutine on error
			}

			// Add the fee for this URL to the total fee
			mu.Lock() // Lock the mutex before updating totalFee
			totalFee += stats.TotalFee
			mu.Unlock() // Unlock the mutex

			h.logger.Infof("[GetTaskFees] Finished processing for URL: %s, Fee: %f", url, stats.TotalFee) // Log when goroutine finishes
		}(ipfsURL) // Pass the current URL to the goroutine
	}

	wg.Wait() // Wait for all goroutines to finish

	// Return the total fee calculation
	response := struct {
		TotalFee float64 `json:"total_fee"`
	}{
		TotalFee: totalFee,
	}

	h.logger.Infof("[GetTaskFees] TotalFeeeeee: %v", response)
	h.logger.Infof("[GetTaskFees] Response Type: %T", response)

	// w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "https://www.triggerx.network")
    w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(response)
}
