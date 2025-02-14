package attester

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ProofResponse represents the structure of the proof data received from the API
type ProofResponse struct {
	ProofHash string `json:"proofHash"`
	CID       string `json:"cid"`
}

// IPFSResponse represents the structure of the JSON stored in IPFS
type IPFSResponse struct {
	JobID            string `json:"job_id"`
	JobType          string `json:"job_type"`
	TaskID           string `json:"task_id"`
	TaskDefinitionID string `json:"task_definition_id"`
	Trigger          struct {
		Timestamp               string `json:"timestamp"`
		Value                   string `json:"value"`
		TxHash                  string `json:"txHash"`
		EventName               string `json:"eventName"`
		ConditionEndpoint       string `json:"conditionEndpoint"`
		ConditionValue          string `json:"conditionValue"`
		CustomTriggerDefinition struct {
			Type   string                 `json:"type"`
			Params map[string]interface{} `json:"params"`
		} `json:"customTriggerDefinition"`
	} `json:"trigger"`
	Action struct {
		Timestamp string `json:"timestamp"`
		TxHash    string `json:"txHash"`
		GasUsed   string `json:"gasUsed"`
		Status    string `json:"status"`
	} `json:"action"`
	Proof struct {
		CertificateHash string `json:"certificateHash"`
		ResponseHash    string `json:"responseHash"`
		Timestamp       string `json:"timestamp"`
	} `json:"proof"`
}

// ValidationResult represents the response we'll send back
type ValidationResult struct {
	IsValid bool   `json:"isValid"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// ValidateTask handles the validation request
func ValidateTask(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		sendResponse(w, ValidationResult{
			IsValid: false,
			Message: "Method not allowed",
			Error:   "Only POST requests are accepted",
		}, http.StatusMethodNotAllowed)
		return
	}

	// Parse the request body
	var proofResp ProofResponse
	if err := json.NewDecoder(r.Body).Decode(&proofResp); err != nil {
		sendResponse(w, ValidationResult{
			IsValid: false,
			Message: "Failed to parse request body",
			Error:   err.Error(),
		}, http.StatusBadRequest)
		return
	}

	// Fetch IPFS content
	ipfsData, err := fetchIPFSContent(proofResp.CID)
	if err != nil {
		sendResponse(w, ValidationResult{
			IsValid: false,
			Message: "Failed to fetch IPFS content",
			Error:   err.Error(),
		}, http.StatusInternalServerError)
		return
	}

	// Parse IPFS JSON content
	var ipfsResp IPFSResponse
	if err := json.Unmarshal([]byte(ipfsData), &ipfsResp); err != nil {
		sendResponse(w, ValidationResult{
			IsValid: false,
			Message: "Failed to parse IPFS content",
			Error:   err.Error(),
		}, http.StatusInternalServerError)
		return
	}

	// Compare hashes
	if proofResp.ProofHash == ipfsResp.Proof.ResponseHash {
		sendResponse(w, ValidationResult{
			IsValid: true,
			Message: "Proof hash matches response hash",
		}, http.StatusOK)
	} else {
		sendResponse(w, ValidationResult{
			IsValid: false,
			Message: "Proof hash does not match response hash",
			Error:   "Hash mismatch",
		}, http.StatusOK)
	}
}

// fetchIPFSContent retrieves content from IPFS gateway
func fetchIPFSContent(cid string) (string, error) {
	// You can use any public IPFS gateway
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

// sendResponse is a helper function to send JSON responses
func sendResponse(w http.ResponseWriter, result ValidationResult, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(result)
}
