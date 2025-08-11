package taskdispatcher

import (
	"fmt"
	"net/http"
	"time"

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
func NewHealthClient(logger logging.Logger, baseURL string) *HealthClient {
	return &HealthClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger:  logger,
		baseURL: baseURL,
	}
}

// GetPerformerData gets a performer using the dynamic selection system
func (hc *HealthClient) GetPerformerData(isImua bool) (types.PerformerData, error) {
	hc.logger.Debug("Getting performer data from health service", "is_imua", isImua)

	// Refresh performers if needed
	// if time.Since(pm.lastRefresh) > PerformerRefreshTTL {
	// 	pm.logger.Debug("Refreshing performers from health service")
	// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// 	defer cancel()
	// 	if err := pm.refreshPerformers(ctx); err != nil {
	// 		pm.logger.Error("Failed to refresh performers", "error", err)
	// 		// Fall back to cached performers
	// 	}
	// }

	var availablePerformers []types.PerformerData
	// availablePerformers := pm.GetAvailablePerformers()
	// pm.logger.Debug("Available performers count", "count", len(availablePerformers))

	if len(availablePerformers) == 0 {
		hc.logger.Warn("No performers available from health service, using fallback")
		fallbackPerformers := []types.PerformerData{
			{
				OperatorID:    1,
				KeeperAddress: "0x011fcbae5f306cd793456ab7d4c0cc86756c693d",
				IsImua:        false,
			},
			{
				OperatorID:    4,
				KeeperAddress: "0x80ce5df21ef5e7c7bdaf52534f0e05e72fcfb9be",
				IsImua:        true,
			},
		}
		availablePerformers = fallbackPerformers
		hc.logger.Info("Using fallback performers", "count", len(availablePerformers))
	}

	// Log available performers for debugging
	// for i, performer := range availablePerformers {
	// 	hc.logger.Debug("Available performer",
	// 		"index", i,
	// 		"operator_id", performer.OperatorID,
	// 		"keeper_address", performer.KeeperAddress,
	// 		"is_imua", performer.IsImua)
	// }

	// Filter by Imua status
	var filteredPerformer types.PerformerData
	for _, performer := range availablePerformers {
		if performer.IsImua == isImua {
			filteredPerformer = performer
		}
	}

	hc.logger.Debug("Filtered performers by Imua status",
		"is_imua", isImua,
		"total_available", len(availablePerformers),
		"filtered_count", 1)

	if filteredPerformer == (types.PerformerData{}) {
		hc.logger.Error("No suitable performers available after Imua filtering",
			"is_imua", isImua,
			"total_available", len(availablePerformers))
		return types.PerformerData{}, fmt.Errorf("no suitable performers available for isImua=%v", isImua)
	}

	hc.logger.Info("Selected performer from health service",
		"performer_id", filteredPerformer.OperatorID,
		"performer_address", filteredPerformer.KeeperAddress,
		"performer_is_imua", filteredPerformer.IsImua)

	return filteredPerformer, nil
}
