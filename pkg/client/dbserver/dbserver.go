package dbserver

import (
	"context"
	"fmt"
	"net/http"
	"time"

	httppkg "github.com/trigg3rX/triggerx-backend/pkg/http"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// DBServerClient handles communication with the DBServer service
type DBServerClient struct {
	logger      logging.Logger
	dbserverUrl string
	httpClient  *httppkg.HTTPClient
}

// NewDBServerClient creates a new instance of DBServerClient
func NewDBServerClient(logger logging.Logger, dbserverUrl string) (*DBServerClient, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}
	if dbserverUrl == "" {
		return nil, fmt.Errorf("RPC address cannot be empty")
	}

	httpClient, err := httppkg.NewHTTPClient(httppkg.DefaultHTTPRetryConfig(), logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	return &DBServerClient{
		logger:      logger,
		dbserverUrl: dbserverUrl,
		httpClient:  httpClient,
	}, nil
}

// HealthCheck checks if the database server is healthy
func (c *DBServerClient) HealthCheck() error {
	url := fmt.Sprintf("%s/api/health", c.dbserverUrl)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %v", err)
	}

	resp, err := c.httpClient.DoWithRetry(ctx, req)
	if err != nil {
		return fmt.Errorf("health check request failed: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: status code %d", resp.StatusCode)
	}

	return nil
}

// Close closes the HTTP client
func (c *DBServerClient) Close() {
	c.httpClient.Close()
}
