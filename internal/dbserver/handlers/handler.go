package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/config"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type NotificationConfig struct {
	EmailFrom     string
	EmailPassword string
	BotToken      string
}

type Handler struct {
	db     *database.Connection
	logger logging.Logger
	config NotificationConfig
}

func NewHandler(db *database.Connection, logger logging.Logger, config NotificationConfig) *Handler {
	return &Handler{
		db:     db,
		logger: logger,
		config: config,
	}
}

func (h *Handler) SendDataToManager(route string, data interface{}) (bool, error) {
	apiURL := fmt.Sprintf("%s%s", config.GetManagerRPCAddress(), route)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return false, fmt.Errorf("error marshaling data: %v", err)
	}

	// Create a client with aggressive timeouts and connection pooling
	client := &http.Client{
		Timeout: 3 * time.Second, // Reduced timeout
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     30 * time.Second,
			DisableKeepAlives:   false,
			DialContext: (&net.Dialer{
				Timeout:   2 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   2 * time.Second,
			ResponseHeaderTimeout: 2 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	// Retry logic with shorter backoff
	maxRetries := 3
	backoff := 200 * time.Millisecond

	for i := 0; i < maxRetries; i++ {
		req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return false, fmt.Errorf("error creating request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Close = true // Ensure connection is closed after request

		resp, err := client.Do(req)
		if err != nil {
			if i == maxRetries-1 {
				return false, fmt.Errorf("error sending data to manager after %d retries: %v", maxRetries, err)
			}
			h.logger.Warnf("Attempt %d failed, retrying in %v: %v", i+1, backoff, err)
			time.Sleep(backoff)
			backoff *= 2 // Exponential backoff
			continue
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			if i == maxRetries-1 {
				return false, fmt.Errorf("manager service error (status=%d): %s", resp.StatusCode, string(body))
			}
			h.logger.Warnf("Attempt %d failed with status %d, retrying in %v", i+1, resp.StatusCode, backoff)
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		return true, nil
	}

	return false, fmt.Errorf("failed to send data to manager after %d retries", maxRetries)
}
