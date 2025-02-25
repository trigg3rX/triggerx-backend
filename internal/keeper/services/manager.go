package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/localtunnel/go-localtunnel"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
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