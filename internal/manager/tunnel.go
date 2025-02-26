package manager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
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

	jsonData, err := json.Marshal(keeper)
	if err != nil {
		logger.Error("Error marshaling data", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error marshaling data"})
		return
	}

	url := fmt.Sprintf("%s/api/keepers/connection", DatabaseIPAddress)
	resp, err := http.Post(url,
		"application/json",
		bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Error("Error sending request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error sending request"})
		return
	}
	defer resp.Body.Close()

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

	c.JSON(http.StatusOK, response)
}

// verifyKeeperEndpoint attempts to verify that the keeper's endpoint is accessible
func verifyKeeperEndpoint(keeperURL string) error {
	// Ensure the URL has http:// or https:// prefix
	if !strings.HasPrefix(keeperURL, "http://") && !strings.HasPrefix(keeperURL, "https://") {
		keeperURL = "https://" + keeperURL
	}

	// Add health check endpoint if it doesn't exist in URL
	if !strings.HasSuffix(keeperURL, "/health") {
		keeperURL = strings.TrimSuffix(keeperURL, "/") + "/health"
	}

	logger.Info("Verifying keeper endpoint", "url", keeperURL)

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	// Create a custom request with headers to bypass localtunnel password page
	req, err := http.NewRequest("GET", keeperURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set custom headers to bypass localtunnel password page
	req.Header.Set("User-Agent", "TriggerX-Manager-Service")
	req.Header.Set("bypass-tunnel-reminder", "true")

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to keeper endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("keeper endpoint returned non-200 status: %d", resp.StatusCode)
	}

	// Read and log the response body for debugging
	var healthResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&healthResponse); err != nil {
		logger.Warn("Failed to decode health response", "error", err)
	} else {
		logger.Info("Keeper health check successful",
			"status", healthResponse["status"],
			"keeper_address", healthResponse["keeper_address"],
			"timestamp", healthResponse["timestamp"])
	}

	return nil
}
