package api

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

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

	// Get the associated job data to access the script content
	var jobData ttypes.JobData
	if err := h.db.Session().Query(`
        SELECT script_ipfs_url FROM triggerx.job_data WHERE job_id = ?`,
		taskData.JobID).Scan(&jobData.ScriptIpfsUrl); err != nil {
		h.logger.Errorf("[CreateTaskData] Error fetching job data: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		h.logger.Errorf("[CreateTaskData] Error creating Docker client: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cli.Close()

	// Download and process the script
	codePath, err := resources.DownloadIPFSFile(jobData.ScriptIpfsUrl)
	if err != nil {
		h.logger.Errorf("[CreateTaskData] Error downloading IPFS file: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(filepath.Dir(codePath))

	// Create and monitor container
	containerID, err := resources.CreateDockerContainer(context.Background(), cli, codePath)
	if err != nil {
		h.logger.Errorf("[CreateTaskData] Error creating container: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cli.ContainerRemove(context.Background(), containerID, types.ContainerRemoveOptions{Force: true})

	// Monitor resources and get stats
	stats, err := resources.MonitorResources(context.Background(), cli, containerID)
	if err != nil {
		h.logger.Errorf("[CreateTaskData] Error monitoring resources: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set the task fee from the calculated value
	taskData.TaskFee = stats.TotalFee

	h.logger.Infof("[CreateTaskData] Creating task with ID: %d", taskData.TaskID)

	// Insert task data with the calculated fee
	if err := h.db.Session().Query(`
        INSERT INTO triggerx.task_data (
            task_id, job_id, task_no,
            quorum_number, quorum_threshold, task_created_tx_hash,
            task_responded_tx_hash, task_hash,
            task_response_hash, task_fee,
            job_type, block_expiry, base_reward_fee_for_attesters,
            base_reward_fee_for_performer, base_reward_fee_for_aggregator,
            dispute_period_blocks, minimum_voting_power, restricted_operator_indexes,
            proof_of_task, data, task_performer,
            is_approved, tp_signature, ta_signature,
            operator_ids
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		taskData.TaskID, taskData.JobID, taskData.TaskNo,
		taskData.QuorumNumber, taskData.QuorumThreshold, taskData.TaskCreatedTxHash,
		taskData.TaskRespondedTxHash, taskData.TaskHash,
		taskData.TaskResponseHash, taskData.TaskFee,
		taskData.JobType, taskData.BlockExpiry, taskData.BaseRewardFeeForAttesters,
		taskData.BaseRewardFeeForPerformer, taskData.BaseRewardFeeForAggregator,
		taskData.DisputePeriodBlocks, taskData.MinimumVotingPower, taskData.RestrictedOperatorIndexes,
		taskData.ProofOfTask, taskData.Data, taskData.TaskPerformer,
		taskData.IsApproved, taskData.TpSignature, taskData.TaSignature,
		taskData.OperatorIds).Exec(); err != nil {
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
        SELECT task_id, job_id, task_no, quorum_number, quorum_threshold, 
               task_created_tx_hash, task_responded_tx_hash, task_hash, 
               task_response_hash, task_fee,
               job_type, block_expiry, base_reward_fee_for_attesters,
               base_reward_fee_for_performer, base_reward_fee_for_aggregator,
               dispute_period_blocks, minimum_voting_power, restricted_operator_indexes,
               proof_of_task, data, task_performer,
               is_approved, tp_signature, ta_signature,
               operator_ids
        FROM triggerx.task_data 
        WHERE task_id = ?`, taskID).Scan(
		&taskData.TaskID, &taskData.JobID, &taskData.TaskNo,
		&taskData.QuorumNumber, &taskData.QuorumThreshold,
		&taskData.TaskCreatedTxHash, &taskData.TaskRespondedTxHash, &taskData.TaskHash,
		&taskData.TaskResponseHash, &taskData.TaskFee,
		&taskData.JobType, &taskData.BlockExpiry, &taskData.BaseRewardFeeForAttesters,
		&taskData.BaseRewardFeeForPerformer, &taskData.BaseRewardFeeForAggregator,
		&taskData.DisputePeriodBlocks, &taskData.MinimumVotingPower, &taskData.RestrictedOperatorIndexes,
		&taskData.ProofOfTask, &taskData.Data, &taskData.TaskPerformer,
		&taskData.IsApproved, &taskData.TpSignature, &taskData.TaSignature,
		&taskData.OperatorIds); err != nil {
		h.logger.Errorf("[GetTaskData] Error retrieving task with ID %s: %v", taskID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[GetTaskData] Successfully retrieved task with ID: %s", taskID)
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

func (h *Handler) GetTaskFees(w http.ResponseWriter, r *http.Request) {
	// Get IPFS URL from query parameter
	ipfsURL := r.URL.Query().Get("ipfs_url")
	if ipfsURL == "" {
		http.Error(w, "Missing ipfs_url query parameter", http.StatusBadRequest)
		return
	}

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

	// Download and process the IPFS file
	codePath, err := resources.DownloadIPFSFile(ipfsURL)
	if err != nil {
		h.logger.Errorf("[GetTaskFees] Error downloading IPFS file: %v", err)
		http.Error(w, "Failed to download IPFS file", http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(filepath.Dir(codePath))

	// Create container
	containerID, err := resources.CreateDockerContainer(ctx, cli, codePath)
	if err != nil {
		h.logger.Errorf("[GetTaskFees] Error creating container: %v", err)
		http.Error(w, "Failed to create container", http.StatusInternalServerError)
		return
	}
	defer cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{Force: true})

	// Monitor resources and get stats
	stats, err := resources.MonitorResources(ctx, cli, containerID)
	if err != nil {
		h.logger.Errorf("[GetTaskFees] Error monitoring resources: %v", err)
		http.Error(w, "Failed to monitor resources", http.StatusInternalServerError)
		return
	}

	// Return the fee calculation
	response := struct {
		TotalFee float64 `json:"total_fee"`
	}{
		TotalFee: stats.TotalFee,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
