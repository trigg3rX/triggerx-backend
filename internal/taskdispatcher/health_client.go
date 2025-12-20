package taskdispatcher

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// HealthClient handles communication with the health service
type HealthClient struct {
	client  *http.Client
	logger  observability.Logger
	baseURL string
}

// PerformerResponse represents the response from health service
type PerformerResponse struct {
	Performers []types.PerformerData `json:"performers"`
	Count      int                   `json:"count"`
	Timestamp  string                `json:"timestamp"`
}

// NewHealthClient creates a new health client
func NewHealthClient(logger observability.Logger, baseURL string) *HealthClient {
	return &HealthClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger:  logger,
		baseURL: baseURL,
	}
}

// GetPerformerData gets a performer using the dynamic selection system
func (hc *HealthClient) GetPerformerData(ctx context.Context, isImua bool, isMainnet bool) (types.PerformerData, error) {
	hc.logger.Debug(ctx, "Getting performer data from health service", observability.Bool("is_imua", isImua))

	if isMainnet {
		return types.PerformerData{
			OperatorID:    1002,
			KeeperAddress: "0x235813b36eea7e48b7069821a78c0bc8384a3c79",
			IsImua:        false,
		}, nil
	}

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
		hc.logger.Warn(ctx, "No performers available from health service, using fallback")
		fallbackPerformers := []types.PerformerData{
			{
				OperatorID:    2,
				KeeperAddress: "0x0a067a261c5F5e8C4c0b9137430b4FE1255EB62e",
				IsImua:        false,
			},
			{
				OperatorID:    1,
				KeeperAddress: "0xcacce39134e3b9d5d9220d87fc546c6f0fb9cc37",
				IsImua:        true,
			},
		}
		availablePerformers = fallbackPerformers
		hc.logger.Info(ctx, "Using fallback performers", observability.Int("count", len(availablePerformers)))
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

	hc.logger.Debug(ctx, "Filtered performers by Imua status",
		observability.Bool("is_imua", isImua),
		observability.Int("total_available", len(availablePerformers)),
		observability.Int("filtered_count", 1))

	if filteredPerformer == (types.PerformerData{}) {
		hc.logger.Error(ctx, "No suitable performers available after Imua filtering",
			observability.Bool("is_imua", isImua),
			observability.Int("total_available", len(availablePerformers)))
		return types.PerformerData{}, fmt.Errorf("no suitable performers available for isImua=%v", isImua)
	}

	hc.logger.Info(ctx, "Selected performer from health service",
		observability.Int64("performer_id", filteredPerformer.OperatorID),
		observability.String("performer_address", filteredPerformer.KeeperAddress),
		observability.Bool("performer_is_imua", filteredPerformer.IsImua))

	return filteredPerformer, nil
}
