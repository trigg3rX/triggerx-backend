package dispatcher

import (
	"context"
	"fmt"

	"github.com/trigg3rX/triggerx-backend/internal/taskdispatcher/rpc/clients/health"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// PerformerFetcher handles performer selection logic
type PerformerFetcher struct {
	healthClient *health.Client
	logger       observability.Logger
}

// NewPerformerFetcher creates a new PerformerFetcher instance
func NewPerformerFetcher(healthClient *health.Client, logger observability.Logger) *PerformerFetcher {
	return &PerformerFetcher{
		healthClient: healthClient,
		logger:       logger,
	}
}

// FetchPerformer retrieves a performer from the health service based on network
func (pf *PerformerFetcher) FetchPerformer(ctx context.Context, network types.KeeperNetwork) (types.PerformerData, error) {
	pf.logger.Debug(ctx, "Fetching performer data",
		observability.String("network", string(network)))

	performer, err := pf.healthClient.GetPerformerData(ctx, network)
	if err != nil {
		pf.logger.Error(ctx, "Failed to get performer data dynamically",
			observability.String("network", string(network)),
			observability.Error(err))
		return types.PerformerData{}, fmt.Errorf("failed to get performer: %w", err)
	}

	pf.logger.Info(ctx, "Successfully fetched performer",
		observability.Int64("operator_id", performer.OperatorID),
		observability.String("keeper_address", performer.KeeperAddress),
		observability.String("network", string(network)))

	return performer, nil
}

// Close closes the health client
func (pf *PerformerFetcher) Close(ctx context.Context) error {
	if pf.healthClient != nil {
		return pf.healthClient.Close(ctx)
	}
	return nil
}
