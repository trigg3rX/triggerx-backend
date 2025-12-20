package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/config"
	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/types"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

// Client handles webhook delivery
type Client struct {
	httpClient *http.Client
	logger     observability.Logger
}

// NewClient creates a new webhook client
func NewClient(logger observability.Logger) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: config.GetWebhookTimeout(),
		},
		logger: logger,
	}
}

// Send sends an event notification to a webhook URL
func (c *Client) Send(ctx context.Context, webhookURL string, notification *types.EventNotification) error {
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
			c.logger.Debug(ctx, "Retrying webhook delivery",
				observability.String("webhook_url", webhookURL),
				observability.Int("attempt", attempt),
				observability.Duration("delay", delay))
			time.Sleep(delay)
		}

		// Create request
		req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(body))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")

		// Inject trace context into HTTP headers
		propagator := otel.GetTextMapPropagator()
		propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))

		// Send request
		resp, err := c.httpClient.Do(req)
		if err != nil {
			c.logger.Warn(ctx, "Webhook delivery failed",
				observability.String("webhook_url", webhookURL),
				observability.Int("attempt", attempt+1),
				observability.Error(err))
			if attempt == maxRetries {
				return fmt.Errorf("failed to deliver webhook after %d attempts: %w", maxRetries+1, err)
			}
			continue
		}

		// Check response status
		defer func() {
			if err := resp.Body.Close(); err != nil {
				c.logger.Error(ctx, "Error closing response body: %v", observability.Error(err))
			}
		}()
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			c.logger.Debug(ctx, "Webhook delivered successfully",
				observability.String("webhook_url", webhookURL),
				observability.Int("status_code", resp.StatusCode))
			return nil
		}

		c.logger.Warn(ctx, "Webhook returned non-2xx status",
			observability.String("webhook_url", webhookURL),
			observability.Int("status_code", resp.StatusCode),
			observability.Int("attempt", attempt+1))

		if attempt == maxRetries {
			return fmt.Errorf("webhook returned non-2xx status after %d attempts: %d", maxRetries+1, resp.StatusCode)
		}
	}

	return fmt.Errorf("unexpected error in webhook delivery")
}
