package api

import (
	"encoding/json"
	"log"
	"math/big"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (h *Handler) CreateUserData(w http.ResponseWriter, r *http.Request) {
	var userData types.UserData
	if err := json.NewDecoder(r.Body).Decode(&userData); err != nil {
		log.Printf("[CreateUserData] Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[CreateUserData] Creating user with ID: %d", userData.UserID)

	// Convert stake amount to Gwei and store as varint
	stakeAmountGwei := new(big.Int)
	stakeAmountGwei = userData.StakeAmount

	if err := h.db.Session().Query(`
        INSERT INTO triggerx.user_data (
            user_id, user_address, job_ids, stake_amount, created_at, last_updated_at
        ) VALUES (?, ?, ?, ?, ?, ?)`,
		userData.UserID, userData.UserAddress, userData.JobIDs, stakeAmountGwei,
		time.Now().UTC(), time.Now().UTC()).Exec(); err != nil {
		log.Printf("[CreateUserData] Error creating user with ID %d: %v", userData.UserID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[CreateUserData] Successfully created user with ID: %d", userData.UserID)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(userData)
}

func (h *Handler) GetUserData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]
	log.Printf("[GetUserData] Retrieving user with ID: %s", userID)

	var (
		userData struct {
			UserID        int64     `json:"user_id"`
			UserAddress   string    `json:"user_address"`
			JobIDs        []int64   `json:"job_ids"`
			StakeAmount   *big.Int  `json:"stake_amount"`
			CreatedAt     time.Time `json:"created_at"`
			LastUpdatedAt time.Time `json:"last_updated_at"`
		}
	)

	if err := h.db.Session().Query(`
        SELECT user_id, user_address, job_ids, stake_amount, created_at, last_updated_at 
        FROM triggerx.user_data 
        WHERE user_id = ?`, userID).Scan(
		&userData.UserID, &userData.UserAddress, &userData.JobIDs,
		&userData.StakeAmount, &userData.CreatedAt, &userData.LastUpdatedAt); err != nil {
		log.Printf("[GetUserData] Error retrieving user with ID %s: %v", userID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[GetUserData] Successfully retrieved user with ID: %s", userID)

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
	log.Printf("[UpdateUserData] Updating user with ID: %s", userID)

	var userData types.UserData
	if err := json.NewDecoder(r.Body).Decode(&userData); err != nil {
		log.Printf("[UpdateUserData] Error decoding request body for user ID %s: %v", userID, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.db.Session().Query(`
        UPDATE triggerx.user_data 
        SET user_address = ?, job_ids = ?, stake_amount = ?, last_updated_at = ?
        WHERE user_id = ?`,
		userData.UserAddress, userData.JobIDs, userData.StakeAmount,
		time.Now().UTC(), userID).Exec(); err != nil {
		log.Printf("[UpdateUserData] Error updating user with ID %s: %v", userID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[UpdateUserData] Successfully updated user with ID: %s", userID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(userData)
}

// func (h *Handler) DeleteUserData(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	userID := vars["id"]
// 	log.Printf("[DeleteUserData] Deleting user with ID: %s", userID)

// 	if err := h.db.Session().Query(`
//         DELETE FROM triggerx.user_data
//         WHERE user_id = ?`, userID).Exec(); err != nil {
// 		log.Printf("[DeleteUserData] Error deleting user with ID %s: %v", userID, err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	log.Printf("[DeleteUserData] Successfully deleted user with ID: %s", userID)
// 	w.WriteHeader(http.StatusNoContent)
// }
