package api

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"

	"github.com/ethereum/go-ethereum/common"
	sdktypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gorilla/mux"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

/*
	TODO:
		- Add GetKeepersByQuorumId
		- Add GetQuorumsByKeeperId
		- Add GetTasksByKeeperId
*/

func (h *Handler) GetKeeperPeerInfo(w http.ResponseWriter, r *http.Request) {
	var peerInfo types.PeerInfo
	if err := json.NewDecoder(r.Body).Decode(&peerInfo); err != nil {
		logger.Error("[GetKeeperPeerInfo] Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logger.Info("[GetKeeperPeerInfo] Retrieving peer info for keeper ID: %s", peerInfo.ID)
	if err := h.db.Session().Query(`
		SELECT id, addresses FROM triggerx.keeper_data WHERE keeper_id = ?`, peerInfo.ID).Scan(
		&peerInfo.ID, &peerInfo.Addresses); err != nil {
		logger.Error("[GetKeeperPeerInfo] Error retrieving peer info for keeper ID %s: %v", peerInfo.ID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Info("[GetKeeperPeerInfo] Successfully retrieved peer info for keeper ID: %s", peerInfo.ID)
	json.NewEncoder(w).Encode(peerInfo)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) CheckKeeperRegistration(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keeperAddress := vars["address"]
	logger.Info("[CheckKeeperRegistration] Retrieving keeper by address: %s", keeperAddress)

	var keeper types.KeeperData
	if err := h.db.Session().Query(`
		SELECT keeper_id, registered_tx, connection_address 
		FROM triggerx.keeper_data 
		WHERE withdrawal_address = ? ALLOW FILTERING`, keeperAddress).Scan(
		&keeper.KeeperID, &keeper.RegisteredTx, &keeper.ConnectionAddress); err != nil {
		logger.Error("[CheckKeeperRegistration] Error retrieving keeper by address %s: %v", keeperAddress, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if keeper.KeeperID == 0 {
		logger.Error("[CheckKeeperRegistration] No keeper found for address: %s", keeperAddress)
		http.Error(w, "Keeper not found", http.StatusNotFound)
		return
	}

	response := types.RegisterKeeperResponse{
		KeeperID:     keeper.KeeperID,
		RegisteredTx: keeper.RegisteredTx,
		PeerID:       keeper.ConnectionAddress,
	}

	logger.Info("[CheckKeeperRegistration] Successfully retrieved keeper by address: %s", keeperAddress)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) CreateKeeperData(w http.ResponseWriter, r *http.Request) {
	var keeperData types.KeeperData
	if err := json.NewDecoder(r.Body).Decode(&keeperData); err != nil {
		logger.Error("[CreateKeeperData] Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	validationStatus := checkRegistration(keeperData.RegisteredTx, keeperData.WithdrawalAddress)
	if !validationStatus {
		logger.Error("[CreateKeeperData] Registration validation failed for keeper ID: %d", keeperData.KeeperID)
		http.Error(w, "Registration validation failed", http.StatusBadRequest)
		return
	}

	logger.Info("[CreateKeeperData] Creating keeper with ID: %d", keeperData.KeeperID)
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
		logger.Error("[CreateKeeperData] Error creating keeper with ID %d: %v", keeperData.KeeperID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Info("[CreateKeeperData] Successfully created keeper with ID: %d", keeperData.KeeperID)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(keeperData)
}

func (h *Handler) GetKeeperData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keeperID := vars["id"]
	logger.Info("[GetKeeperData] Retrieving keeper with ID: %s", keeperID)

	var keeperData types.KeeperData
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
		logger.Error("[GetKeeperData] Error retrieving keeper with ID %s: %v", keeperID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Info("[GetKeeperData] Successfully retrieved keeper with ID: %s", keeperID)
	json.NewEncoder(w).Encode(keeperData)
}

func (h *Handler) GetAllKeepers(w http.ResponseWriter, r *http.Request) {
	logger.Info("[GetAllKeepers] Retrieving all keepers")
	var keepers []types.KeeperData
	iter := h.db.Session().Query(`SELECT * FROM triggerx.keeper_data`).Iter()

	var keeper types.KeeperData
	for iter.Scan(
		&keeper.KeeperID, &keeper.WithdrawalAddress, &keeper.Stakes,
		&keeper.Strategies, &keeper.Verified, &keeper.CurrentQuorumNo,
		&keeper.RegisteredTx, &keeper.Status, &keeper.BlsSigningKeys,
		&keeper.ConnectionAddress) {
		keepers = append(keepers, keeper)
	}

	if err := iter.Close(); err != nil {
		logger.Error("[GetAllKeepers] Error retrieving keepers: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Info("[GetAllKeepers] Successfully retrieved %d keepers", len(keepers))
	json.NewEncoder(w).Encode(keepers)
}

func (h *Handler) UpdateKeeperData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keeperID := vars["id"]

	var keeperData types.KeeperData
	if err := json.NewDecoder(r.Body).Decode(&keeperData); err != nil {
		logger.Error("[UpdateKeeperData] Error decoding request body for keeper ID %s: %v", keeperID, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logger.Info("[UpdateKeeperData] Updating keeper with ID: %s", keeperID)
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
		logger.Error("[UpdateKeeperData] Error updating keeper with ID %s: %v", keeperID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Info("[UpdateKeeperData] Successfully updated keeper with ID: %s", keeperID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(keeperData)
}

// func (h *Handler) DeleteKeeperData(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	keeperID := vars["id"]
// 	log.Printf("[DeleteKeeperData] Deleting keeper with ID: %s", keeperID)

// 	if err := h.db.Session().Query(`
//         DELETE FROM triggerx.keeper_data
//         WHERE keeper_id = ?`, keeperID).Exec(); err != nil {
// 		log.Printf("[DeleteKeeperData] Error deleting keeper with ID %s: %v", keeperID, err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	log.Printf("[DeleteKeeperData] Successfully deleted keeper with ID: %s", keeperID)
// 	w.WriteHeader(http.StatusNoContent)
// }

func (h *Handler) CreateTaskHistory(w http.ResponseWriter, r *http.Request) {
	var taskHistory types.TaskHistory
	if err := json.NewDecoder(r.Body).Decode(&taskHistory); err != nil {
		logger.Error("[CreateTaskHistory] Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logger.Info("[CreateTaskHistory] Creating task history for task ID: %d", taskHistory.TaskID)
	if err := h.db.Session().Query(`
        INSERT INTO triggerx.task_history (task_id, quorum_id, keepers, responses, consensus_method, validation_status, tx_hash)
        VALUES (?, ?, ?, ?, ?, ?, ?)`,
		taskHistory.TaskID, taskHistory.QuorumID, taskHistory.Keepers, taskHistory.Responses,
		taskHistory.ConsensusMethod, taskHistory.ValidationStatus, taskHistory.TxHash).Exec(); err != nil {
		logger.Error("[CreateTaskHistory] Error creating task history for task ID %d: %v", taskHistory.TaskID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Info("[CreateTaskHistory] Successfully created task history for task ID: %d", taskHistory.TaskID)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(taskHistory)
}

func (h *Handler) GetTaskHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]
	logger.Info("[GetTaskHistory] Retrieving task history for task ID: %s", taskID)

	var taskHistory types.TaskHistory
	if err := h.db.Session().Query(`
        SELECT task_id, quorum_id, keepers, responses, consensus_method, validation_status, tx_hash 
        FROM triggerx.task_history 
        WHERE task_id = ?`, taskID).Scan(
		&taskHistory.TaskID, &taskHistory.QuorumID, &taskHistory.Keepers, &taskHistory.Responses,
		&taskHistory.ConsensusMethod, &taskHistory.ValidationStatus, &taskHistory.TxHash); err != nil {
		logger.Error("[GetTaskHistory] Error retrieving task history for task ID %s: %v", taskID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Info("[GetTaskHistory] Successfully retrieved task history for task ID: %s", taskID)
	json.NewEncoder(w).Encode(taskHistory)
}

// func (h *Handler) UpdateTaskHistory(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	taskID := vars["id"]

// 	var taskHistory models.TaskHistory
// 	if err := json.NewDecoder(r.Body).Decode(&taskHistory); err != nil {
// 		log.Printf("[UpdateTaskHistory] Error decoding request body for task ID %s: %v", taskID, err)
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	log.Printf("[UpdateTaskHistory] Updating task history for task ID: %s", taskID)
// 	if err := h.db.Session().Query(`
//         UPDATE triggerx.task_history
//         SET quorum_id = ?, keepers = ?, responses = ?, consensus_method = ?, validation_status = ?, tx_hash = ?
//         WHERE task_id = ?`,
// 		taskHistory.QuorumID, taskHistory.Keepers, taskHistory.Responses,
// 		taskHistory.ConsensusMethod, taskHistory.ValidationStatus, taskHistory.TxHash, taskID).Exec(); err != nil {
// 		log.Printf("[UpdateTaskHistory] Error updating task history for task ID %s: %v", taskID, err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	log.Printf("[UpdateTaskHistory] Successfully updated task history for task ID: %s", taskID)
// 	w.WriteHeader(http.StatusOK)
// 	json.NewEncoder(w).Encode(taskHistory)
// }

// func (h *Handler) DeleteTaskHistory(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	taskID := vars["id"]
// 	log.Printf("[DeleteTaskHistory] Deleting task history for task ID: %s", taskID)

// 	if err := h.db.Session().Query(`
//         DELETE FROM triggerx.task_history
//         WHERE task_id = ?`, taskID).Exec(); err != nil {
// 		log.Printf("[DeleteTaskHistory] Error deleting task history for task ID %s: %v", taskID, err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	log.Printf("[DeleteTaskHistory] Successfully deleted task history for task ID: %s", taskID)
// 	w.WriteHeader(http.StatusNoContent)
// }

func checkRegistration(registeredTx string, withdrawalAddress string) bool {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		logger.Error("[checkRegistration] Error loading .env file: %v", err)
	}

	// Get ETH RPC URL from environment
	ethRPCURL := os.Getenv("ETH_RPC_URL")
	if ethRPCURL == "" {
		logger.Error("[checkRegistration] ETH_RPC_URL environment variable not set")
		return false
	}

	// Create Ethereum client
	client, err := ethclient.Dial(ethRPCURL)
	if err != nil {
		logger.Error("[checkRegistration] Failed to connect to Ethereum client: %v", err)
		return false
	}
	defer client.Close()

	// Convert tx hash string to hash
	txHash := common.HexToHash(registeredTx)

	// Get transaction
	tx, isPending, err := client.TransactionByHash(context.Background(), txHash)
	if err != nil {
		logger.Error("[checkRegistration] Failed to get transaction: %v", err)
		return false
	}

	if isPending {
		logger.Error("[checkRegistration] Transaction is still pending")
		return false
	}

	// Get transaction sender
	from, err := sdktypes.Sender(sdktypes.LatestSignerForChainID(tx.ChainId()), tx)
	if err != nil {
		logger.Error("[checkRegistration] Failed to get transaction sender: %v", err)
		return false
	}

	// Compare addresses (case-insensitive)
	if !strings.EqualFold(from.Hex(), withdrawalAddress) {
		logger.Error("[checkRegistration] Address mismatch - From: %s, Expected: %s", from.Hex(), withdrawalAddress)
		return false
	}

	return true
}
