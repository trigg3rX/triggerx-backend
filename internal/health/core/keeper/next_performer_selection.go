package keeper

import (
	"context"
	"fmt"
	"strconv"

	"github.com/trigg3rX/triggerx-backend/internal/health/redis"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// PerformerSelector handles the selection of performers using Redis-based round-robin
type PerformerSelector struct {
	stateManager *StateManager
	redisClient  *redis.Client
	logger       observability.Logger
}

// NewPerformerSelector creates a new PerformerSelector instance
func NewPerformerSelector(stateManager *StateManager, redisClient *redis.Client, logger observability.Logger) *PerformerSelector {
	return &PerformerSelector{
		stateManager: stateManager,
		redisClient:  redisClient,
		logger:       logger.With(observability.String("component", "performer_selector")),
	}
}

// GetNextPerformer selects the next performer using round-robin selection
// It stores only the last operator ID in Redis to maintain selection state across restarts
func (ps *PerformerSelector) GetNextPerformer(ctx context.Context, network types.KeeperNetwork) (*types.PerformerData, error) {
	// Get all active keepers for the specified network
	activeKeepers := ps.stateManager.GetActiveKeepersByNetwork(ctx, network)

	if len(activeKeepers) == 0 {
		ps.logger.Warn(ctx, "No active keepers available for network",
			observability.String("network", string(network)),
		)
		return nil, fmt.Errorf("no active keepers available for network %s", network)
	}

	// Get the last operator ID from Redis
	lastOperatorID, err := ps.redisClient.GetLastOperatorID(ctx, string(network))
	if err != nil {
		ps.logger.Warn(ctx, "Failed to get last operator ID from Redis, starting from first keeper",
			observability.Error(err),
		)
		lastOperatorID = 0
	}

	// Find the next operator ID available from active keepers
	selectedKeeper := ps.findNextKeeper(activeKeepers, lastOperatorID)

	// Parse operator ID from string to int64
	operatorID, err := strconv.ParseInt(selectedKeeper.OperatorID, 10, 64)
	if err != nil {
		ps.logger.Error(ctx, "Failed to parse operator ID",
			observability.Error(err),
			observability.String("operator_id", selectedKeeper.OperatorID),
		)
		operatorID = 0
	}

	// Update Redis with the new operator ID
	if err := ps.redisClient.SetLastOperatorID(ctx, string(network), operatorID); err != nil {
		ps.logger.Warn(ctx, "Failed to update last operator ID in Redis",
			observability.Error(err),
		)
		// Continue anyway - selection still works, just won't persist across restarts
	}

	performer := &types.PerformerData{
		OperatorID:    operatorID,
		KeeperAddress: selectedKeeper.KeeperAddress,
		Network:       network,
	}

	ps.logger.Info(ctx, "Selected performer via round-robin",
		observability.Int64("operator_id", performer.OperatorID),
		observability.String("keeper_address", performer.KeeperAddress),
		observability.String("network", string(performer.Network)),
		observability.Int64("last_operator_id", lastOperatorID),
		observability.Int("total_active", len(activeKeepers)),
	)

	return performer, nil
}

// findNextKeeper finds the next keeper after the last selected operator ID
// If the last operator ID is not found or is 0, returns the first keeper
func (ps *PerformerSelector) findNextKeeper(activeKeepers []types.KeeperInfo, lastOperatorID int64) types.KeeperInfo {
	if lastOperatorID == 0 || len(activeKeepers) == 1 {
		return activeKeepers[0]
	}

	// Find the index of the last selected operator
	lastIndex := -1
	for i, keeper := range activeKeepers {
		opID, err := strconv.ParseInt(keeper.OperatorID, 10, 64)
		if err != nil {
			continue
		}
		if opID == lastOperatorID {
			lastIndex = i
			break
		}
	}

	// Select the next keeper (wrap around to first if at end or not found)
	nextIndex := (lastIndex + 1) % len(activeKeepers)
	return activeKeepers[nextIndex]
}

// GetNextPerformerWithFallback selects the next performer with fallback options
// It first tries to use the round-robin selection, then falls back to hardcoded values
func (ps *PerformerSelector) GetNextPerformerWithFallback(ctx context.Context, network types.KeeperNetwork) (*types.PerformerData, error) {
	// Try round-robin selection first
	performer, err := ps.GetNextPerformer(ctx, network)
	if err == nil {
		return performer, nil
	}

	ps.logger.Warn(ctx, "Round-robin selection failed, using fallback performers",
		observability.Error(err),
		observability.String("network", string(network)),
	)

	// Fallback to hardcoded performers based on network
	return ps.getFallbackPerformer(network)
}

// getFallbackPerformer returns a hardcoded fallback performer for the given network
func (ps *PerformerSelector) getFallbackPerformer(network types.KeeperNetwork) (*types.PerformerData, error) {
	switch network {
	case types.NetworkMainnet:
		return &types.PerformerData{
			OperatorID:    1002,
			KeeperAddress: "0x235813b36eea7e48b7069821a78c0bc8384a3c79",
			Network:       types.NetworkMainnet,
		}, nil
	case types.NetworkSepolia:
		return &types.PerformerData{
			OperatorID:    2,
			KeeperAddress: "0x0a067a261c5F5e8C4c0b9137430b4FE1255EB62e",
			Network:       types.NetworkSepolia,
		}, nil
	case types.NetworkImua:
		return &types.PerformerData{
			OperatorID:    1,
			KeeperAddress: "0xcacce39134e3b9d5d9220d87fc546c6f0fb9cc37",
			Network:       types.NetworkImua,
		}, nil
	default:
		return nil, fmt.Errorf("no fallback performer available for network %s", network)
	}
}
