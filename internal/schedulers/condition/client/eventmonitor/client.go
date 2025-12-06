package eventmonitor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/types"
	httppkg "github.com/trigg3rX/triggerx-backend/pkg/http"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// Client handles communication with the Event Monitor Service
type Client struct {
	baseURL    string
	httpClient *httppkg.HTTPClient
	logger     logging.Logger
}

// NewClient creates a new Event Monitor Service client
func NewClient(baseURL string, logger logging.Logger) (*Client, error) {
	httpConfig := httppkg.DefaultHTTPRetryConfig()
	httpConfig.Timeout = 10 * time.Second

	httpClient, err := httppkg.NewHTTPClient(httpConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	return &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
		logger:     logger,
	}, nil
}

// Register registers a monitoring request with the Event Monitor Service
func (c *Client) Register(req *types.MonitoringRequest) error {
	url := fmt.Sprintf("%s/api/v1/monitor/register", c.baseURL)

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.DoWithRetry(context.Background(), httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func () {
		if err := resp.Body.Close(); err != nil {
			c.logger.Errorf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			if errorMsg, ok := errorResp["error"].(string); ok {
				return fmt.Errorf("registration failed: %s", errorMsg)
			}
		}
		return fmt.Errorf("registration failed with status: %d", resp.StatusCode)
	}

	var registerResp types.RegisterResponse
	if err := json.NewDecoder(resp.Body).Decode(&registerResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !registerResp.Success {
		return fmt.Errorf("registration failed: %s", registerResp.Message)
	}

	c.logger.Info("Registered monitoring request with Event Monitor Service",
		"request_id", req.RequestID,
		"chain_id", req.ChainID,
		"contract_address", req.ContractAddr)

	return nil
}

// Unregister unregisters a monitoring request from the Event Monitor Service
func (c *Client) Unregister(requestID string) error {
	url := fmt.Sprintf("%s/api/v1/monitor/unregister", c.baseURL)

	reqBody := types.UnregisterRequest{
		RequestID: requestID,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.DoWithRetry(context.Background(), httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func () {
		if err := resp.Body.Close(); err != nil {
			c.logger.Errorf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			if errorMsg, ok := errorResp["error"].(string); ok {
				return fmt.Errorf("unregistration failed: %s", errorMsg)
			}
		}
		return fmt.Errorf("unregistration failed with status: %d", resp.StatusCode)
	}

	var unregisterResp types.UnregisterResponse
	if err := json.NewDecoder(resp.Body).Decode(&unregisterResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !unregisterResp.Success {
		return fmt.Errorf("unregistration failed: %s", unregisterResp.Message)
	}

	c.logger.Info("Unregistered monitoring request from Event Monitor Service",
		"request_id", requestID)

	return nil
}

// Close closes the HTTP client
func (c *Client) Close() {
	if c.httpClient != nil {
		c.httpClient.Close()
	}
}
