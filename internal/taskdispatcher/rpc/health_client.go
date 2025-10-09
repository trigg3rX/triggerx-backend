package rpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	pb "github.com/trigg3rX/triggerx-backend/pkg/rpc/proto"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// HealthClient provides a gRPC client for the health service
type HealthClient struct {
	conn   *grpc.ClientConn
	client pb.HealthServiceClient
	logger logging.Logger
	target string
}

// NewHealthClient creates a new health service gRPC client
func NewHealthClient(target string, logger logging.Logger) (*HealthClient, error) {
	// Create gRPC connection with appropriate options (using new API)
	conn, err := grpc.NewClient(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create health service client: %w", err)
	}

	client := pb.NewHealthServiceClient(conn)

	return &HealthClient{
		conn:   conn,
		client: client,
		logger: logger.With("component", "health_grpc_client"),
		target: target,
	}, nil
}

// Close closes the gRPC connection
func (hc *HealthClient) Close() error {
	if hc.conn != nil {
		return hc.conn.Close()
	}
	return nil
}

// GetPerformerData gets the next performer for task assignment using gRPC
func (hc *HealthClient) GetPerformerData(ctx context.Context, isImua bool, isMainnet bool) (types.PerformerData, error) {
	hc.logger.Debug("Getting performer data via gRPC", "is_imua", isImua)

	// Mainnet fallback
	if isMainnet {
		return types.PerformerData{
			OperatorID:    1002,
			KeeperAddress: "0x235813b36eea7e48b7069821a78c0bc8384a3c79",
			IsImua:        false,
		}, nil
	}

	// Call gRPC GetNextPerformer method
	req := &pb.GetNextPerformerRequest{
		IsImuaTask: isImua,
		TaskType:   "", // Can be extended for task-specific routing
	}

	resp, err := hc.client.GetNextPerformer(ctx, req)
	if err != nil {
		hc.logger.Error("Failed to get next performer via gRPC", "error", err)
		return hc.getFallbackPerformer(isImua), fmt.Errorf("gRPC call failed: %w", err)
	}

	if resp.Performer == nil {
		hc.logger.Warn("No performer returned from health service")
		return hc.getFallbackPerformer(isImua), fmt.Errorf("no performer available")
	}

	// Convert proto Performer to types.PerformerData
	performer := types.PerformerData{
		OperatorID:    resp.Performer.OperatorId,
		KeeperAddress: resp.Performer.KeeperAddress,
		IsImua:        resp.Performer.IsImua,
	}

	hc.logger.Info("Selected performer from health service",
		"performer_id", performer.OperatorID,
		"performer_address", performer.KeeperAddress,
		"performer_is_imua", performer.IsImua,
		"strategy", resp.Strategy,
	)

	return performer, nil
}

// GetActivePerformers gets all active performers from the health service
func (hc *HealthClient) GetActivePerformers(ctx context.Context, includeImua bool, limit int32) ([]types.PerformerData, error) {
	hc.logger.Debug("Getting active performers via gRPC",
		"include_imua", includeImua,
		"limit", limit,
	)

	req := &pb.GetActivePerformersRequest{
		IncludeImua: includeImua,
		Limit:       limit,
	}

	resp, err := hc.client.GetActivePerformers(ctx, req)
	if err != nil {
		hc.logger.Error("Failed to get active performers via gRPC", "error", err)
		return nil, fmt.Errorf("gRPC call failed: %w", err)
	}

	// Convert proto Performers to types.PerformerData
	performers := make([]types.PerformerData, len(resp.Performers))
	for i, p := range resp.Performers {
		performers[i] = types.PerformerData{
			OperatorID:    p.OperatorId,
			KeeperAddress: p.KeeperAddress,
			IsImua:        p.IsImua,
		}
	}

	hc.logger.Info("Retrieved active performers",
		"count", len(performers),
	)

	return performers, nil
}

// getFallbackPerformer returns a hardcoded fallback performer
func (hc *HealthClient) getFallbackPerformer(isImua bool) types.PerformerData {
	fallbackPerformers := []types.PerformerData{
		{
			OperatorID:    4,
			KeeperAddress: "0x0a067a261c5f5e8c4c0b9137430b4fe1255eb62e",
			IsImua:        false,
		},
		{
			OperatorID:    1,
			KeeperAddress: "0xcacce39134e3b9d5d9220d87fc546c6f0fb9cc37",
			IsImua:        true,
		},
	}

	for _, performer := range fallbackPerformers {
		if performer.IsImua == isImua {
			hc.logger.Warn("Using fallback performer",
				"operator_id", performer.OperatorID,
				"is_imua", performer.IsImua,
			)
			return performer
		}
	}

	// Return first fallback if no match
	return fallbackPerformers[0]
}
