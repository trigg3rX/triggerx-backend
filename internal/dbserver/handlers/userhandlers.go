package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (h *Handler) GetUserData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]
	h.logger.Infof("[GetUserData] Retrieving user with ID: %s", userID)

	var userData types.UserData

	if err := h.db.Session().Query(`
        SELECT user_id, user_address, job_ids, account_balance
        FROM triggerx.user_data 
        WHERE user_id = ?`, userID).Scan(
		&userData.UserID, &userData.UserAddress, &userData.JobIDs, &userData.AccountBalance); err != nil {
		h.logger.Errorf("[GetUserData] Error retrieving user with ID %s: %v", userID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infof("[GetUserData] Successfully retrieved user with ID: %s", userID)

	response := types.UserData{
		UserID:         userData.UserID,
		UserAddress:    userData.UserAddress,
		JobIDs:         userData.JobIDs,
		AccountBalance: userData.AccountBalance,
	}

	json.NewEncoder(w).Encode(response)
}
