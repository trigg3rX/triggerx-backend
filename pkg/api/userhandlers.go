package api

import (
	"encoding/json"
	"math/big"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func (h *Handler) GetUserData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]
	h.logger.Infof("[GetUserData] Retrieving user with ID: %s", userID)

	var (
		userData struct {
			UserID         int64     `json:"user_id"`
			UserAddress    string    `json:"user_address"`
			JobIDs         []int64   `json:"job_ids"`
			StakeAmount    *big.Int  `json:"stake_amount"`
			AccountBalance *big.Int  `json:"account_balance"`
			CreatedAt      time.Time `json:"created_at"`
			LastUpdatedAt  time.Time `json:"last_updated_at"`
		}
	)

	if err := h.db.Session().Query(`
        SELECT userID, userAddress, jobIDs, stakeAmount, accountBalance, createdAt, lastUpdatedAt 
        FROM triggerx.user_data 
        WHERE userID = ?`, userID).Scan(
		&userData.UserID, &userData.UserAddress, &userData.JobIDs,
		&userData.StakeAmount, &userData.AccountBalance, &userData.CreatedAt, &userData.LastUpdatedAt); err != nil {
		h.logger.Errorf("[GetUserData] Error retrieving user with ID %s: %v", userID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[GetUserData] Successfully retrieved user with ID: %s", userID)

	stakeAmountFloat := new(big.Float).SetInt(userData.StakeAmount)
	stakeAmountFloat64, _ := stakeAmountFloat.Float64()

	accountBalanceFloat := new(big.Float).SetInt(userData.AccountBalance)
	accountBalanceFloat64, _ := accountBalanceFloat.Float64()

	response := struct {
		UserID         int64   `json:"user_id"`
		UserAddress    string  `json:"user_address"`
		JobIDs         []int64 `json:"job_ids"`
		StakeAmount    float64 `json:"stake_amount"`
		AccountBalance float64 `json:"account_balance"`
	}{
		UserID:         userData.UserID,
		UserAddress:    userData.UserAddress,
		JobIDs:         userData.JobIDs,
		StakeAmount:    stakeAmountFloat64,
		AccountBalance: accountBalanceFloat64,
	}

	json.NewEncoder(w).Encode(response)
}
