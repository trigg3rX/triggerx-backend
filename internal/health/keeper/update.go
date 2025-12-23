package keeper

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	maxRetries = 3
)

// Custom error types
var (
	ErrKeeperNotVerified = errors.New("keeper not verified")
)

// UpdateKeeperHealth updates the health status of a keeper
func (sm *StateManager) UpdateKeeperHealth(ctx context.Context, keeperHealth types.KeeperHealthCheckIn) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	address := keeperHealth.KeeperAddress
	now := time.Now().UTC()

	existingState, exists := sm.keepers[address]
	if !exists {
		sm.logger.Warn(ctx, "Received health check-in from unverified keeper",
			observability.String("keeper", address),
		)
		return ErrKeeperNotVerified
	}

	// Update the state with new health check-in data
	existingState.Version = keeperHealth.Version
	existingState.PeerID = keeperHealth.PeerID
	existingState.LastCheckedIn = now
	existingState.IsActive = true
	existingState.IsImua = keeperHealth.IsImua

	// Update database
	if err := sm.retryWithBackoff(ctx, func() error {
		return sm.updateKeeperStatusInDatabase(ctx, keeperHealth, true)
	}, maxRetries); err != nil {
		return fmt.Errorf("failed to update keeper status in database: %w", err)
	}

	sm.logger.Info(ctx, "Updated keeper health status",
		observability.String("keeper", address),
		observability.String("version", keeperHealth.Version),
		observability.Bool("is_imua", keeperHealth.IsImua),
	)
	return nil
}

func (sm *StateManager) updateKeeperStatusInDatabase(ctx context.Context, keeperHealth types.KeeperHealthCheckIn, isActive bool) error {
	// Start a span for the database update operation
	ctx, span := sm.tracer.Start(ctx, "state_manager.update_keeper_status",
		observability.WithSpanKind(trace.SpanKindInternal),
		observability.WithAttributes(
			attribute.String("keeper.address", keeperHealth.KeeperAddress),
			attribute.Bool("keeper.active", isActive),
			attribute.String("keeper.version", keeperHealth.Version),
		),
	)
	defer span.End()

	if err := sm.db.UpdateKeeperHealth(ctx, keeperHealth, isActive); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to update keeper status in database: %w", err)
	}

	span.SetStatus(codes.Ok, "")
	// sm.logger.Debug(ctx, "Updated keeper status in database",
	// 	observability.String("keeper", keeperHealth.KeeperAddress),
	// 	observability.Bool("active", isActive),
	// 	observability.String("version", keeperHealth.Version),
	// )
	return nil
}
