package checkin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"

	// "github.com/trigg3rX/triggerx-backend/pkg/crypto"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

var logger = logging.GetLogger(logging.Development, logging.KeeperProcess)

// CheckInWithHealthService sends a health check-in request to the health service
func CheckInWithHealthService() error {
	healthServiceURL := fmt.Sprintf("%s/health", config.HealthIPAddress)

	// // Sign the message
	// signature, err := crypto.SignMessage(config.KeeperAddress, config.PrivateKeyController)
	// if err != nil {
	// 	logger.Error("Failed to sign check-in message", "error", err)
	// 	return fmt.Errorf("failed to sign check-in message: %w", err)
	// }
	signature := "0x"

	payload := types.KeeperHealth{
		KeeperAddress: config.KeeperAddress,
		Version:       "0.0.7",
		Timestamp:     time.Now().UTC(),
		Signature:     signature,
		PeerID:        config.PeerID,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		logger.Error("Failed to marshal check-in payload", "error", err)
		return fmt.Errorf("failed to marshal check-in payload: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", healthServiceURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		logger.Error("Failed to create check-in request", "error", err)
		return fmt.Errorf("failed to create check-in request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Failed to send check-in request", "error", err)
		return fmt.Errorf("failed to send check-in request: %w", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		logger.Error("Health service returned non-OK status", "status", resp.StatusCode)
		return fmt.Errorf("health service returned status: %d", resp.StatusCode)
	}

	return nil
}
