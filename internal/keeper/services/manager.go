package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/localtunnel/go-localtunnel"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

var logger = logging.GetLogger(logging.Development, logging.KeeperProcess)
var activeTunnel *localtunnel.Listener
var tunnelServer *http.Server

func Init() {
	config.Init()
	logger.Info("Config Initialized")
}

func SetupTunnel(port string, keeperAddress string) (string, error) {
	logger.Info("Setting up localtunnel for port")

	// Convert port string to int
	portInt := 0
	_, err := fmt.Sscanf(port, "%d", &portInt)
	if err != nil {
		return "", fmt.Errorf("invalid port format: %w", err)
	}

	// Create a unique subdomain to avoid conflicts
	subdomain := strings.ToLower(keeperAddress)

	// Remove any non-alphanumeric characters that might cause issues
	subdomain = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return '-'
	}, subdomain)

	logger.Info("Using subdomain for tunnel", "subdomain", subdomain)

	// Create a localtunnel listener
	listener, err := localtunnel.Listen(localtunnel.Options{
		Subdomain: subdomain,
	})

	if err != nil {
		logger.Warn("Failed to create tunnel with specific subdomain, trying without subdomain",
			"error", err,
			"subdomain", subdomain)

		// Try again without a specific subdomain
		listener, err = localtunnel.Listen(localtunnel.Options{})
		if err != nil {
			return "", fmt.Errorf("failed to create localtunnel listener: %w", err)
		}
	}

	// Store the listener for later cleanup
	activeTunnel = listener

	// Get the tunnel URL
	tunnelURL := listener.Addr().String()

	// Ensure the URL has the http:// prefix
	if !strings.HasPrefix(tunnelURL, "http://") && !strings.HasPrefix(tunnelURL, "https://") {
		tunnelURL = "https://" + tunnelURL
	}

	// Start a simple HTTP server to handle requests through the tunnel
	mux := http.NewServeMux()

	// Add a health endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// Add response headers to help with tunnel password bypass
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("User-Agent", "TriggerX-Keeper-Service")
		w.Header().Set("bypass-tunnel-reminder", "true")

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":         "healthy",
			"keeper_address": config.KeeperAddress,
			"timestamp":      time.Now().Unix(),
		})
	})

	// Add a root endpoint for testing
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Add response headers to help with tunnel password bypass
		w.Header().Set("User-Agent", "TriggerX-Keeper-Service")
		w.Header().Set("bypass-tunnel-reminder", "true")

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "TriggerX Keeper Service")
	})

	// Create and start the server
	tunnelServer = &http.Server{
		Handler: mux,
	}

	// Start the server in a goroutine
	go func() {
		if err := tunnelServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			logger.Error("Tunnel server error", "error", err)
		}
	}()

	return tunnelURL, nil
}

// CloseTunnel closes the active tunnel if one exists
func CloseTunnel() {
	// First, close the HTTP server if it exists
	if tunnelServer != nil {
		logger.Info("Shutting down tunnel HTTP server")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := tunnelServer.Shutdown(ctx); err != nil {
			logger.Error("Error shutting down tunnel server", "error", err)
		}
		tunnelServer = nil
	}

	// Then close the tunnel listener
	if activeTunnel != nil {
		logger.Info("Closing localtunnel")
		if err := activeTunnel.Close(); err != nil {
			logger.Error("Error closing tunnel", "error", err)
		} else {
			logger.Info("Tunnel closed successfully")
		}
		activeTunnel = nil
	}
}


func ConnectToTaskManager(keeperAddress string, connectionAddress string) (bool, error) {
	taskManagerRPCAddress := fmt.Sprintf("%s/connect", config.ManagerIPAddress)

	var payload types.UpdateKeeperConnectionData
	payload.KeeperAddress = keeperAddress
	payload.ConnectionAddress = connectionAddress

	// Ensure the connection address has the proper format for health checks
	if !strings.HasPrefix(payload.ConnectionAddress, "http://") && !strings.HasPrefix(payload.ConnectionAddress, "https://") {
		payload.ConnectionAddress = "https://" + payload.ConnectionAddress
	}

	logger.Info("Connecting to task manager",
		"keeper_address", keeperAddress,
		"connection_address", payload.ConnectionAddress,
		"task_manager", taskManagerRPCAddress)

	var response types.UpdateKeeperConnectionDataResponse

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := http.Post(taskManagerRPCAddress,
		"application/json",
		bytes.NewBuffer(jsonPayload))
	if err != nil {
		return false, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	logger.Info("Connected to task manager successfully",
		"keeperID", response.KeeperID,
		"keeperAddress", response.KeeperAddress,
		"keeperURL", payload.ConnectionAddress)

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("task manager returned non-200 status code: %d", resp.StatusCode)
	}

	envFile := ".env"
	keeperIDLine := fmt.Sprintf("\nKEEPER_ID=%d", response.KeeperID)

	f, err := os.OpenFile(envFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return false, fmt.Errorf("failed to open .env file: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(keeperIDLine); err != nil {
		return false, fmt.Errorf("failed to write keeper ID to .env: %w", err)
	}

	return true, nil
}
