package manager

import (
	"fmt"
	"net/http"
	"strings"
	"time"
	"bytes"
	"encoding/json"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
	// "github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// handleKeeperConnection handles incoming keeper connection requests
func HandleKeeperConnection(c *gin.Context) {
	var keeper types.UpdateKeeperConnectionData
	var response types.UpdateKeeperConnectionDataResponse
	if err := c.BindJSON(&keeper); err != nil {
		logger.Error("Failed to parse keeper connection request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	if !strings.HasPrefix(keeper.ConnectionAddress, "http://") && !strings.HasPrefix(keeper.ConnectionAddress, "https://") {
		logger.Error("Invalid keeper URL format", "url", keeper.ConnectionAddress)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid keeper URL format"})
		return
	}

	jsonData, err := json.Marshal(keeper)
    if err != nil {
        logger.Error("Error marshaling data", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error marshaling data"})
		return
    }

	url := "http://data.triggerx.network/api/keepers/" + keeper.KeeperAddress
    resp, err := http.Post(url, 
        "application/json", 
        bytes.NewBuffer(jsonData))
    if err != nil {
        logger.Error("Error sending request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error sending request"})
		return
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusCreated {
        logger.Error("Unexpected status code", "status", resp.StatusCode)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unexpected status code"})
		return
    }

	json.NewDecoder(resp.Body).Decode(&response)	

	logger.Info("Keeper connected successfully",
		"keeperID", response.KeeperID,
		"keeperAddress", response.KeeperAddress,
		"keeperURL", keeper.ConnectionAddress)

	// Verify the keeper's endpoint is accessible
	if err := verifyKeeperEndpoint(keeper.ConnectionAddress); err != nil {
		logger.Warn("Keeper endpoint verification failed",
			"keeper_address", keeper.KeeperAddress,
			"keeper_url", keeper.ConnectionAddress,
			"error", err)
		// We still return success but log the warning
	}

	c.JSON(http.StatusOK, gin.H{"response": response})
}

// verifyKeeperEndpoint attempts to verify that the keeper's endpoint is accessible
func verifyKeeperEndpoint(keeperURL string) error {
	// Add health check endpoint if it doesn't exist in URL
	if !strings.HasSuffix(keeperURL, "/health") {
		keeperURL = strings.TrimSuffix(keeperURL, "/") + "/health"
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(keeperURL)
	if err != nil {
		return fmt.Errorf("failed to connect to keeper endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("keeper endpoint returned non-200 status: %d", resp.StatusCode)
	}

	return nil
}

// SendTaskToPerformer sends a job execution request to the performer service running on port 4003.
// It takes job and task metadata, formats them into the expected payload structure, and makes a POST request.
// Returns true if the task was successfully sent and accepted by the performer, false with error otherwise.
func SendTaskToPerformer(jobData *types.HandleCreateJobData, triggerData *types.TriggerData) (bool, error) {
	client := &http.Client{}

	// Construct payload with task definition ID and job details required for execution
	payload := map[string]interface{}{
		"job": jobData,
		"trigger": triggerData,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", "http://localhost:9002/task/execute", bytes.NewBuffer(jsonData))
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("performer returned non-200 status code: %d, body: %s", resp.StatusCode, string(body))
	}

	return true, nil
}
