package services

// import (
// 	"bytes"
// 	"encoding/json"
// 	"fmt"
// 	"net/http"
// 	"time"

// 	"github.com/trigg3rX/triggerx-backend/pkg/types"
// 	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
// 	"github.com/trigg3rX/triggerx-backend/pkg/logging"
// )

// var logger = logging.GetLogger(logging.Development, logging.KeeperProcess)

// func Init() {
// 	config.Init()
// 	logger.Info("Config Initialized")
// }

// func ConnectToManager() (bool, error) {
// 	// Prepare the keeper data to send to the manager
// 	keeperData := types.UpdateKeeperConnectionData{
// 		KeeperAddress: config.KeeperAddress,
// 		ConnectionAddress: config.ConnectionAddress,
// 	}

// 	// Convert the data to JSON
// 	jsonData, err := json.Marshal(keeperData)
// 	if err != nil {
// 		logger.Error("Failed to marshal keeper data", "error", err)
// 		return false, err
// 	}

// 	// Construct the URL for the manager's keeper connect endpoint
// 	url := fmt.Sprintf("%s/keeper/connect", config.ManagerIPAddress)
	
// 	// Create a new HTTP request
// 	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
// 	if err != nil {
// 		logger.Error("Failed to create HTTP request", "error", err)
// 		return false, err
// 	}
	
// 	// Set headers
// 	req.Header.Set("Content-Type", "application/json")
	
// 	// Create HTTP client with timeout
// 	client := &http.Client{
// 		Timeout: 10 * time.Second,
// 	}
	
// 	// Send the request
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		logger.Error("Failed to connect to manager", "error", err)
// 		return false, err
// 	}
// 	defer resp.Body.Close()
	
// 	// Check response status
// 	if resp.StatusCode != http.StatusOK {
// 		logger.Error("Manager returned non-OK status", "status", resp.StatusCode)
// 		return false, fmt.Errorf("manager returned status: %d", resp.StatusCode)
// 	}
	
// 	// Decode the response
// 	var response map[string]interface{}
// 	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
// 		logger.Error("Failed to decode response", "error", err)
// 		return false, err
// 	}
	
// 	logger.Info("Successfully connected to manager", "response", response)
// 	return true, nil
// }