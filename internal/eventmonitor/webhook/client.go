package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/config"
	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/types"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// Client handles webhook delivery
type Client struct {
	httpClient *http.Client
	logger     logging.Logger
}

// NewClient creates a new webhook client
func NewClient(logger logging.Logger) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: config.GetWebhookTimeout(),
		},
		logger: logger,
	}
}

// Send sends an event notification to a webhook URL
func (c *Client) Send(webhookURL string, notification *types.EventNotification) error {
	// Marshal notification to JSON
	body, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	// Retry logic with exponential backoff
	maxRetries := config.GetWebhookMaxRetries()
	retryDelay := config.GetWebhookRetryDelay()

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			delay := retryDelay * time.Duration(1<<uint(attempt-1))
			c.logger.Debug("Retrying webhook delivery",
				"webhook_url", webhookURL,
				"attempt", attempt,
				"delay", delay)
			time.Sleep(delay)
		}

		// Create request
		req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(body))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")

		// Send request
		resp, err := c.httpClient.Do(req)
		if err != nil {
			c.logger.Warn("Webhook delivery failed",
				"webhook_url", webhookURL,
				"attempt", attempt+1,
				"error", err)
			if attempt == maxRetries {
				return fmt.Errorf("failed to deliver webhook after %d attempts: %w", maxRetries+1, err)
			}
			continue
		}

		// Check response status
		resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			c.logger.Debug("Webhook delivered successfully",
				"webhook_url", webhookURL,
				"status_code", resp.StatusCode)
			return nil
		}

		c.logger.Warn("Webhook returned non-2xx status",
			"webhook_url", webhookURL,
			"status_code", resp.StatusCode,
			"attempt", attempt+1)

		if attempt == maxRetries {
			return fmt.Errorf("webhook returned non-2xx status after %d attempts: %d", maxRetries+1, resp.StatusCode)
		}
	}

	return fmt.Errorf("unexpected error in webhook delivery")
}
