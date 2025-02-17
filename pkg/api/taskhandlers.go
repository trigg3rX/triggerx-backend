package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	// "time"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/gorilla/mux"
	"github.com/trigg3rX/triggerx-backend/pkg/resources"
	ttypes "github.com/trigg3rX/triggerx-backend/pkg/types"
)

/*
	TODO:
		- Add GetTasksByJobId
		- Add GetTasksByQuorumId
*/

func (h *Handler) CreateTaskData(w http.ResponseWriter, r *http.Request) {
	var taskData ttypes.TaskData
	if err := json.NewDecoder(r.Body).Decode(&taskData); err != nil {
		h.logger.Errorf("[CreateTaskData] Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the next task ID
	var maxTaskID int64
	if err := h.db.Session().Query(`
		SELECT MAX(taskID) FROM triggerx.task_data`).Scan(&maxTaskID); err != nil {
		h.logger.Errorf("[CreateTaskData] Error getting max task ID: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	taskData.TaskID = maxTaskID + 1

	var jobData ttypes.JobData
	if err := h.db.Session().Query(`
        SELECT scriptIPFSUrl FROM triggerx.job_data WHERE jobID = ?`,
		taskData.JobID).Scan(&jobData.ScriptIPFSUrl); err != nil {
		h.logger.Errorf("[CreateTaskData] Error fetching job data: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		h.logger.Errorf("[CreateTaskData] Error creating Docker client: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cli.Close()

	codePath, err := resources.DownloadIPFSFile(jobData.ScriptIPFSUrl)
	if err != nil {
		h.logger.Errorf("[CreateTaskData] Error downloading IPFS file: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(filepath.Dir(codePath))

	containerID, err := resources.CreateDockerContainer(context.Background(), cli, codePath)
	if err != nil {
		h.logger.Errorf("[CreateTaskData] Error creating container: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cli.ContainerRemove(context.Background(), containerID, types.ContainerRemoveOptions{Force: true})

	stats, err := resources.MonitorResources(context.Background(), cli, containerID)
	if err != nil {
		h.logger.Errorf("[CreateTaskData] Error monitoring resources: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	taskFeeStr := fmt.Sprintf("%.6f", stats.TotalFee)

	h.logger.Infof("[CreateTaskData] Creating task with ID: %d", taskData.TaskID)

	if err := h.db.Session().Query(`
        INSERT INTO triggerx.task_data (
            taskID, jobID, taskDefinitionID,
            taskFee, taskPerformer, isApproved
        ) VALUES (?, ?, ?, ?, ?, ?)`,
		taskData.TaskID, taskData.JobID, taskData.TaskDefinitionID,
		taskFeeStr, taskData.TaskPerformer, taskData.IsApproved).Exec(); err != nil {
		h.logger.Errorf("[CreateTaskData] Error inserting task with ID %d: %v", taskData.TaskID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[CreateTaskData] Successfully created task with ID: %d", taskData.TaskID)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(taskData)
}

func (h *Handler) GetTaskData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]
	h.logger.Infof("[GetTaskData] Fetching task with ID: %s", taskID)

	var taskData ttypes.TaskData
	if err := h.db.Session().Query(`
        SELECT taskID, jobID, taskDefinitionID,
               taskRespondedTxHash, taskResponseHash, taskFee,
               proofOfTask, data, taskPerformer,
               isApproved, tpSignature, taSignature,
               operatorIds, executedAt
        FROM triggerx.task_data 
        WHERE taskID = ?`, taskID).Scan(
		&taskData.TaskID, &taskData.JobID, &taskData.TaskDefinitionID,
		&taskData.TaskRespondedTxHash, &taskData.TaskResponseHash, &taskData.TaskFee,
		&taskData.ProofOfTask, &taskData.Data, &taskData.TaskPerformer,
		&taskData.IsApproved, &taskData.TpSignature, &taskData.TaSignature,
		&taskData.OperatorIds, &taskData.ExecutedAt); err != nil {
		h.logger.Errorf("[GetTaskData] Error retrieving task with ID %s: %v", taskID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[GetTaskData] Successfully retrieved task with ID: %s", taskID)
	json.NewEncoder(w).Encode(taskData)
}

func (h *Handler) GetTaskFees(w http.ResponseWriter, r *http.Request) {
	ipfsURL := r.URL.Query().Get("ipfs_url")
	if ipfsURL == "" {
		http.Error(w, "Missing ipfs_url query parameter", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

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

	codePath, err := resources.DownloadIPFSFile(ipfsURL)
	if err != nil {
		h.logger.Errorf("[GetTaskFees] Error downloading IPFS file: %v", err)
		http.Error(w, "Failed to download IPFS file", http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(filepath.Dir(codePath))

	containerID, err := resources.CreateDockerContainer(ctx, cli, codePath)
	if err != nil {
		h.logger.Errorf("[GetTaskFees] Error creating container: %v", err)
		http.Error(w, "Failed to create container", http.StatusInternalServerError)
		return
	}
	defer cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{Force: true})

	stats, err := resources.MonitorResources(ctx, cli, containerID)
	if err != nil {
		h.logger.Errorf("[GetTaskFees] Error monitoring resources: %v", err)
		http.Error(w, "Failed to monitor resources", http.StatusInternalServerError)
		return
	}

	response := struct {
		TotalFee float64 `json:"total_fee"`
	}{
		TotalFee: stats.TotalFee,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
