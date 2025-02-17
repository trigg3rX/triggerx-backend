package attester

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	// "github.com/trigg3rX/triggerx-backend/pkg/types"
)

type ProofResponse struct {
	ProofHash string `json:"proofHash"`
	CID       string `json:"cid"`
}

type ValidationResult struct {
	IsValid bool   `json:"isValid"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

type TaskValidationRequest struct {
	ProofOfTask      string `json:"proofOfTask"`
	Data             string `json:"data"`
	TaskDefinitionID uint16 `json:"taskDefinitionId"`
	Performer        string `json:"performer"`
}

type ValidationResponse struct {
	Data    bool   `json:"data"`
	Error   bool   `json:"error"`
	Message string `json:"message,omitempty"`
}

func ValidateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendResponse(w, ValidationResponse{
			Data:    false,
			Error:   true,
			Message: "Only POST requests are accepted",
		}, http.StatusMethodNotAllowed)
		return
	}

	var taskRequest TaskValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&taskRequest); err != nil {
		sendResponse(w, ValidationResponse{
			Data:    false,
			Error:   true,
			Message: fmt.Sprintf("Failed to parse request body: %v", err),
		}, http.StatusBadRequest)
		return
	}

	log.Printf("Received Task Validation Request:")
	log.Printf("Proof of Task: %s", taskRequest.ProofOfTask)
	log.Printf("Data: %s", taskRequest.Data)
	log.Printf("Task Definition ID: %d", taskRequest.TaskDefinitionID)
	log.Printf("Performer Address: %s", taskRequest.Performer)

	ipfsData, err := fetchIPFSContent(taskRequest.ProofOfTask)
	if err != nil {
		sendResponse(w, ValidationResponse{
			Data:    false,
			Error:   true,
			Message: fmt.Sprintf("Failed to fetch IPFS content: %v", err),
		}, http.StatusInternalServerError)
		return
	}

	log.Printf("IPFS Data Fetched")
	log.Printf("IPFS Data: %s", ipfsData)

	// var ipfsResp types.IPFSResponse
	// if err := json.Unmarshal([]byte(ipfsData), &ipfsResp); err != nil {
	// 	sendResponse(w, ValidationResponse{
	// 		Data:    false,
	// 		Error:   true,
	// 		Message: fmt.Sprintf("Failed to parse IPFS content: %v", err),
	// 	}, http.StatusInternalServerError)
	// 	return
	// }

	sendResponse(w, ValidationResponse{
		Data:    true,
		Error:   false,
		Message: "",
	}, http.StatusOK)
}

func fetchIPFSContent(cid string) (string, error) {
	ipfsGateway := "https://aquamarine-urgent-limpet-846.mypinata.cloud/ipfs/"
	resp, err := http.Get(ipfsGateway + cid)
	if err != nil {
		return "", fmt.Errorf("failed to fetch IPFS content: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch IPFS content: status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	return string(body), nil
}

func sendResponse(w http.ResponseWriter, response ValidationResponse, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
