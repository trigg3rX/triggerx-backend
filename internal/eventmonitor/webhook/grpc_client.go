package webhook

import (
	"context"
	"fmt"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/config"
	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/types"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	rpcclient "github.com/trigg3rX/triggerx-backend/pkg/rpc/client"
)

// NotificationClient is an interface for sending event notifications
type NotificationClient interface {
	Send(ctx context.Context, serviceURL string, notification *types.EventNotification) error
}

// GRPCClient handles gRPC notification delivery
type GRPCClient struct {
	logger observability.Logger
	tracer observability.Tracer
}

// NewGRPCClient creates a new gRPC notification client
func NewGRPCClient(logger observability.Logger, tracer observability.Tracer) *GRPCClient {
	return &GRPCClient{
		logger: logger,
		tracer: tracer,
	}
}

// Send sends an event notification via gRPC to the condition scheduler
func (c *GRPCClient) Send(ctx context.Context, serviceURL string, notification *types.EventNotification) error {
	// Create gRPC client for the service URL
	client := rpcclient.NewClient(rpcclient.Config{
		ServiceName: serviceURL,
		Timeout:     30 * time.Second,
		MaxRetries:  config.GetWebhookMaxRetries(),
		RetryDelay:  config.GetWebhookRetryDelay(),
		PoolSize:    10,
		PoolTimeout: 5 * time.Second,
	}, c.logger, c.tracer)
	defer func() {
		_ = client.Close(ctx)
	}()

	// Retry logic with exponential backoff
	maxRetries := config.GetWebhookMaxRetries()
	retryDelay := config.GetWebhookRetryDelay()

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			delay := retryDelay * time.Duration(1<<uint(attempt-1))
			c.logger.Debug(ctx, "Retrying gRPC notification delivery",
				observability.String("service_url", serviceURL),
				observability.Int("attempt", attempt),
				observability.Duration("delay", delay))
			time.Sleep(delay)
		}

		// Make gRPC call to event-notification method
		var response map[string]interface{}
		err := client.Call(ctx, "event-notification", notification, &response)
		if err != nil {
			lastErr = err
			c.logger.Warn(ctx, "gRPC notification delivery failed",
				observability.String("service_url", serviceURL),
				observability.Int("attempt", attempt+1),
				observability.Error(err))
			if attempt == maxRetries {
				return fmt.Errorf("failed to deliver notification after %d attempts: %w", maxRetries+1, err)
			}
			continue
		}

		// Check response
		if success, ok := response["success"].(bool); ok && success {
			c.logger.Debug(ctx, "gRPC notification delivered successfully",
				observability.String("service_url", serviceURL),
				observability.String("request_id", notification.RequestID))
			return nil
		}

		// Extract error message if available
		errorMsg := "unknown error"
		if errStr, ok := response["error"].(string); ok {
			errorMsg = errStr
		}
		lastErr = fmt.Errorf("notification failed: %s", errorMsg)

		c.logger.Warn(ctx, "gRPC notification returned error",
			observability.String("service_url", serviceURL),
			observability.Int("attempt", attempt+1),
			observability.String("error", errorMsg))

		if attempt == maxRetries {
			return fmt.Errorf("notification failed after %d attempts: %s", maxRetries+1, errorMsg)
		}
	}

	return fmt.Errorf("unexpected error in notification delivery: %w", lastErr)
}
