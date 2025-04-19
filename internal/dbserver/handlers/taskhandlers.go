package handlers

import (
	"context"
	"encoding/json"
	"fmt"
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
		time.Now().UTC(), taskData.TaskPerformerID, false).Exec(); err != nil {
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

// Move the core logic to a service function
func (h *Handler) CalculateTaskFees(ipfsURLs string) (float64, error) {
	if ipfsURLs == "" {
		return 0, fmt.Errorf("missing IPFS URLs")
	}

	urlList := strings.Split(ipfsURLs, ",")
	totalFee := 0.0
	var mu sync.Mutex
	var wg sync.WaitGroup

	ctx := context.Background()

	// Create Docker client
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create Docker client: %v", err)
	}
	defer cli.Close()

	// Process each IPFS URL concurrently
	for _, ipfsURL := range urlList {
		ipfsURL = strings.TrimSpace(ipfsURL)
		wg.Add(1)

		go func(url string) {
			defer wg.Done()

			codePath, err := resources.DownloadIPFSFile(url)
			if err != nil {
				h.logger.Errorf("Error downloading IPFS file: %v", err)
				return
			}
			defer os.RemoveAll(filepath.Dir(codePath))

			containerID, err := resources.CreateDockerContainer(ctx, cli, codePath)
			if err != nil {
				h.logger.Errorf("Error creating container: %v", err)
				return
			}
			defer cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{Force: true})

			stats, err := resources.MonitorResources(ctx, cli, containerID)
			if err != nil {
				h.logger.Errorf("Error monitoring resources: %v", err)
				return
			}

			mu.Lock()
			totalFee += stats.TotalFee
			mu.Unlock()
		}(ipfsURL)
	}

	wg.Wait()
	return totalFee, nil
}

// Update the HTTP handler to use the service function
func (h *Handler) GetTaskFees(w http.ResponseWriter, r *http.Request) {
	ipfsURLs := r.URL.Query().Get("ipfs_url")

	totalFee, err := h.CalculateTaskFees(ipfsURLs)
	if err != nil {
		h.logger.Errorf("[GetTaskFees] Error calculating fees: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		TotalFee float64 `json:"total_fee"`
	}{
		TotalFee: totalFee,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) UpdateTaskFee(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]
	h.logger.Infof("[UpdateTaskFee] Updating task fee for task with ID: %s", taskID)

	var taskFee struct {
		Fee float64 `json:"fee"`
	}
	if err := json.NewDecoder(r.Body).Decode(&taskFee); err != nil {
		h.logger.Errorf("[UpdateTaskFee] Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.db.Session().Query(`
		UPDATE triggerx.task_data
		SET task_fee = ?
		WHERE task_id = ?`, taskFee.Fee, taskID).Exec(); err != nil {
		h.logger.Errorf("[UpdateTaskFee] Error updating task fee: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[UpdateTaskFee] Successfully updated task fee for task with ID: %s", taskID)
	json.NewEncoder(w).Encode(taskFee)
}
