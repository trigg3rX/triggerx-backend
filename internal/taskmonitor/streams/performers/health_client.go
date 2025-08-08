package performers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/config"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// HealthClient handles communication with the health service
type HealthClient struct {
	client  *http.Client
	logger  logging.Logger
	baseURL string
}

// PerformerResponse represents the response from health service
type PerformerResponse struct {
	Performers []types.PerformerData `json:"performers"`
	Count      int                   `json:"count"`
	Timestamp  string                `json:"timestamp"`
}

// NewHealthClient creates a new health client
func NewHealthClient(logger logging.Logger) *HealthClient {
	return &HealthClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger:  logger,
		baseURL: config.GetHealthRPCUrl(),
	}
}

// GetActivePerformers fetches active performers from health service
func (hc *HealthClient) GetActivePerformers(ctx context.Context) ([]types.PerformerData, error) {
	url := fmt.Sprintf("%s/performers", hc.baseURL)

	hc.logger.Debug("Fetching performers from health service", "url", url)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		hc.logger.Error("Failed to create request", "error", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := hc.client.Do(req)
	if err != nil {
		hc.logger.Error("Failed to fetch performers", "url", url, "error", err)
		return nil, fmt.Errorf("failed to fetch performers: %w", err)
	}
	defer resp.Body.Close()

	hc.logger.Debug("Health service response", "status_code", resp.StatusCode, "url", url)

	if resp.StatusCode != http.StatusOK {
		hc.logger.Error("Health service returned error status", "status_code", resp.StatusCode, "url", url)
		return nil, fmt.Errorf("health service returned status %d", resp.StatusCode)
	}

	var performerResp PerformerResponse
	if err := json.NewDecoder(resp.Body).Decode(&performerResp); err != nil {
		hc.logger.Error("Failed to decode response", "error", err)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	hc.logger.Debug("Decoded performer response", "count", performerResp.Count, "timestamp", performerResp.Timestamp)

	// Convert to PerformerData format
	performers := make([]types.PerformerData, 0, len(performerResp.Performers))
	for _, p := range performerResp.Performers {
		performer := types.PerformerData{
			OperatorID:    p.OperatorID,
			KeeperAddress: p.KeeperAddress,
			IsImua:        p.IsImua,
		}
		performers = append(performers, performer)
		hc.logger.Debug("Converted performer", "operator_id", p.OperatorID, "keeper_address", p.KeeperAddress, "is_imua", p.IsImua)
	}

	hc.logger.Info("Fetched active performers from health service",
		"count", len(performers),
		"url", url)

	return performers, nil
}
