package api

import (
	// "context"
	"encoding/json"
	"net/http"
	// "os"
	// "strings"

	// "github.com/joho/godotenv"

	// "github.com/ethereum/go-ethereum/common"
	// sdktypes "github.com/ethereum/go-ethereum/core/types"
	// "github.com/ethereum/go-ethereum/ethclient"
	"github.com/gorilla/mux"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

/*
	TODO:
		- Add GetKeepersByQuorumId
		- Add GetQuorumsByKeeperId
		- Add GetTasksByKeeperId
*/

func (h *Handler) CreateKeeperData(w http.ResponseWriter, r *http.Request) {
	var keeperData types.KeeperData
	if err := json.NewDecoder(r.Body).Decode(&keeperData); err != nil {
		h.logger.Errorf("[CreateKeeperData] Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// validationStatus := checkRegistration(h, keeperData.RegisteredTx, keeperData.KeeperAddress)
	// if !validationStatus {
	// 	h.logger.Errorf("[CreateKeeperData] Registration validation failed for keeper ID: %d", keeperData.KeeperID)
	// 	http.Error(w, "Registration validation failed", http.StatusBadRequest)
	// 	return
	// }

	h.logger.Infof("[CreateKeeperData] Creating keeper with ID: %d", keeperData.KeeperID)
	if err := h.db.Session().Query(`
        INSERT INTO triggerx.keeper_data (
            keeperID, keeperAddress, rewardsAddress, stakes, strategies, 
            verified, registeredTx, status, blsSigningKeys, connectionAddress, keeperType
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		keeperData.KeeperID, keeperData.KeeperAddress, keeperData.RewardsAddress,
		keeperData.Stakes, keeperData.Strategies, keeperData.Verified,
		keeperData.RegisteredTx, keeperData.Status, keeperData.BlsSigningKeys,
		keeperData.ConnectionAddress, 2).Exec(); err != nil {
		h.logger.Errorf("[CreateKeeperData] Error creating keeper with ID %d: %v", keeperData.KeeperID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[CreateKeeperData] Successfully created keeper with ID: %d", keeperData.KeeperID)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(keeperData)
}

func (h *Handler) GetKeeperData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keeperID := vars["id"]
	h.logger.Infof("[GetKeeperData] Retrieving keeper with ID: %s", keeperID)

	var keeperData types.KeeperData
	if err := h.db.Session().Query(`
        SELECT keeperID, keeperAddress, rewardsAddress, stakes, strategies, 
               verified, registeredTx, status, blsSigningKeys, connectionAddress
        FROM triggerx.keeper_data 
        WHERE keeperID = ? AND keeperType = 2`, keeperID).Scan(
		&keeperData.KeeperID, &keeperData.KeeperAddress, &keeperData.RewardsAddress,
		&keeperData.Stakes, &keeperData.Strategies, &keeperData.Verified,
		&keeperData.RegisteredTx, &keeperData.Status, &keeperData.BlsSigningKeys,
		&keeperData.ConnectionAddress); err != nil {
		h.logger.Errorf("[GetKeeperData] Error retrieving keeper with ID %s: %v", keeperID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[GetKeeperData] Successfully retrieved keeper with ID: %s", keeperID)
	json.NewEncoder(w).Encode(keeperData)
}

func (h *Handler) GetAllKeepers(w http.ResponseWriter, r *http.Request) {
	h.logger.Infof("[GetAllKeepers] Retrieving all keepers")
	var keepers []types.KeeperData
	iter := h.db.Session().Query(`SELECT keeperID, keeperAddress, rewardsAddress, stakes, strategies, verified, registeredTx, status, blsSigningKeys, connectionAddress FROM triggerx.keeper_data`).Iter()

	var keeper types.KeeperData
	for iter.Scan(
		&keeper.KeeperID, &keeper.KeeperAddress, &keeper.RewardsAddress,
		&keeper.Stakes, &keeper.Strategies, &keeper.Verified,
		&keeper.RegisteredTx, &keeper.Status, &keeper.BlsSigningKeys,
		&keeper.ConnectionAddress) {
		keepers = append(keepers, keeper)
	}

	if err := iter.Close(); err != nil {
		h.logger.Errorf("[GetAllKeepers] Error retrieving keepers: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[GetAllKeepers] Successfully retrieved %d keepers", len(keepers))
	json.NewEncoder(w).Encode(keepers)
}

func (h *Handler) UpdateKeeperData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keeperID := vars["id"]

	var keeperData types.KeeperData
	if err := json.NewDecoder(r.Body).Decode(&keeperData); err != nil {
		h.logger.Errorf("[UpdateKeeperData] Error decoding request body for keeper ID %s: %v", keeperID, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.logger.Infof("[UpdateKeeperData] Updating keeper with ID: %s", keeperID)
	if err := h.db.Session().Query(`
        UPDATE triggerx.keeper_data 
        SET keeperAddress = ?, rewardsAddress = ?, stakes = ?, strategies = ?, 
            verified = ?, registeredTx = ?, status = ?, 
            blsSigningKeys = ?, connectionAddress = ?
        WHERE keeperID = ?`,
		keeperData.KeeperAddress, keeperData.RewardsAddress, keeperData.Stakes,
		keeperData.Strategies, keeperData.Verified, keeperData.RegisteredTx,
		keeperData.Status, keeperData.BlsSigningKeys, keeperData.ConnectionAddress,
		keeperID).Exec(); err != nil {
		h.logger.Errorf("[UpdateKeeperData] Error updating keeper with ID %s: %v", keeperID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[UpdateKeeperData] Successfully updated keeper with ID: %s", keeperID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(keeperData)
}

// func (h *Handler) CreateTaskHistory(w http.ResponseWriter, r *http.Request) {
// 	var taskHistory types.TaskHistory
// 	if err := json.NewDecoder(r.Body).Decode(&taskHistory); err != nil {
// 		h.logger.Errorf("[CreateTaskHistory] Error decoding request body: %v", err)
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	h.logger.Infof("[CreateTaskHistory] Creating task history for task ID: %d", taskHistory.TaskID)
// 	if err := h.db.Session().Query(`
//         INSERT INTO triggerx.task_history (task_id, quorum_id, keepers, responses, consensus_method, validation_status, tx_hash)
//         VALUES (?, ?, ?, ?, ?, ?, ?)`,
// 		taskHistory.TaskID, taskHistory.QuorumID, taskHistory.Keepers, taskHistory.Responses,
// 		taskHistory.ConsensusMethod, taskHistory.ValidationStatus, taskHistory.TxHash).Exec(); err != nil {
// 		h.logger.Errorf("[CreateTaskHistory] Error creating task history for task ID %d: %v", taskHistory.TaskID, err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	h.logger.Infof("[CreateTaskHistory] Successfully created task history for task ID: %d", taskHistory.TaskID)
// 	w.WriteHeader(http.StatusCreated)
// 	json.NewEncoder(w).Encode(taskHistory)
// }

// func (h *Handler) GetTaskHistory(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	taskID := vars["id"]
// 	h.logger.Infof("[GetTaskHistory] Retrieving task history for task ID: %s", taskID)

// 	var taskHistory types.TaskHistory
// 	if err := h.db.Session().Query(`
//         SELECT task_id, quorum_id, keepers, responses, consensus_method, validation_status, tx_hash 
//         FROM triggerx.task_history 
//         WHERE task_id = ?`, taskID).Scan(
// 		&taskHistory.TaskID, &taskHistory.QuorumID, &taskHistory.Keepers, &taskHistory.Responses,
// 		&taskHistory.ConsensusMethod, &taskHistory.ValidationStatus, &taskHistory.TxHash); err != nil {
// 		h.logger.Errorf("[GetTaskHistory] Error retrieving task history for task ID %s: %v", taskID, err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	h.logger.Infof("[GetTaskHistory] Successfully retrieved task history for task ID: %s", taskID)
// 	json.NewEncoder(w).Encode(taskHistory)
// }

// func checkRegistration(h *Handler, registeredTx string, withdrawalAddress string) bool {
// 	// Load environment variables from .env file
// 	if err := godotenv.Load(); err != nil {
// 		h.logger.Errorf("[checkRegistration] Error loading .env file: %v", err)
// 	}

// 	// Get ETH RPC URL from environment
// 	ethRPCURL := os.Getenv("ETH_RPC_URL")
// 	if ethRPCURL == "" {
// 		h.logger.Errorf("[checkRegistration] ETH_RPC_URL environment variable not set")
// 		return false
// 	}

// 	// Create Ethereum client
// 	client, err := ethclient.Dial(ethRPCURL)
// 	if err != nil {
// 		h.logger.Errorf("[checkRegistration] Failed to connect to Ethereum client: %v", err)
// 		return false
// 	}
// 	defer client.Close()

// 	// Convert tx hash string to hash
// 	txHash := common.HexToHash(registeredTx)

// 	// Get transaction
// 	tx, isPending, err := client.TransactionByHash(context.Background(), txHash)
// 	if err != nil {
// 		h.logger.Errorf("[checkRegistration] Failed to get transaction: %v", err)
// 		return false
// 	}

// 	if isPending {
// 		h.logger.Errorf("[checkRegistration] Transaction is still pending")
// 		return false
// 	}

// 	// Get transaction sender
// 	from, err := sdktypes.Sender(sdktypes.LatestSignerForChainID(tx.ChainId()), tx)
// 	if err != nil {
// 		h.logger.Errorf("[checkRegistration] Failed to get transaction sender: %v", err)
// 		return false
// 	}

// 	// Compare addresses (case-insensitive)
// 	if !strings.EqualFold(from.Hex(), withdrawalAddress) {
// 		h.logger.Errorf("[checkRegistration] Address mismatch - From: %s, Expected: %s", from.Hex(), withdrawalAddress)
// 		return false
// 	}

// 	return true
// }
