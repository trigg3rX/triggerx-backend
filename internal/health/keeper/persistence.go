package keeper

import (
	"context"
	"fmt"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/health/config"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// LoadVerifiedKeepers loads only verified keepers from the database
func (sm *StateManager) LoadVerifiedKeepers(ctx context.Context) error {
	// Start a span for loading keepers
	ctx, span := sm.tracer.Start(ctx, "state_manager.load_verified_keepers",
		observability.WithSpanKind(trace.SpanKindInternal),
	)
	defer span.End()

	sm.logger.Debug(ctx, "Loading verified keepers from database...")

	// Get only verified keepers from database
	keepers, err := sm.db.GetVerifiedKeepers(ctx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to load verified keepers from database: %w", err)
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Clear existing state
	sm.keepers = make(map[string]*types.KeeperInfo)

	// Load each keeper's state (initially marked as inactive)
	for _, keeper := range keepers {
		state := &types.KeeperInfo{
			KeeperName:       keeper.KeeperName,
			KeeperAddress:    keeper.KeeperAddress,
			ConsensusAddress: keeper.ConsensusAddress,
			OperatorID:       keeper.OperatorID,
			Version:          keeper.Version,
			PeerID:           keeper.PeerID,
			IsActive:         false,
			LastCheckedIn:    keeper.LastCheckedIn,
			IsImua:           keeper.IsImua,
		}
		sm.keepers[keeper.KeeperAddress] = state
	}

	span.SetAttributes(attribute.Int("keepers.loaded", len(sm.keepers)))
	span.SetStatus(codes.Ok, "")
	sm.logger.Info(ctx, "Successfully loaded verified keepers",
		observability.Int("count", len(sm.keepers)),
	)
	return nil
}

// DumpState updates all keepers to inactive in the database
func (sm *StateManager) DumpState(ctx context.Context) error {
	// Start a span for the state dump operation
	ctx, span := sm.tracer.Start(ctx, "state_manager.dump_state",
		observability.WithSpanKind(trace.SpanKindInternal),
	)
	defer span.End()

	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.logger.Info(ctx, "Dumping keeper state to database...")

	activeCount := 0
	for address, state := range sm.keepers {
		if state.IsActive {
			activeCount++
			// Create a minimal health check-in with just the address
			health := types.KeeperHealthCheckIn{
				KeeperAddress: address,
			}

			if err := sm.retryWithBackoff(ctx, func() error {
				return sm.updateKeeperStatusInDatabase(ctx, health, false)
			}, config.GetHealthCheckMaxRetries()); err != nil {
				sm.logger.Error(ctx, "Failed to update keeper status during state dump",
					observability.Error(err),
					observability.String("keeper", address),
				)
				continue
			}
		}
	}

	span.SetAttributes(attribute.Int("keepers.dumped", activeCount))
	span.SetStatus(codes.Ok, "")
	// sm.logger.Info(ctx, "Successfully dumped keeper state")
	return nil
}

// RetryWithBackoff retries a database operation with exponential backoff
func (sm *StateManager) retryWithBackoff(ctx context.Context, operation func() error, maxRetries int) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		err = operation()
		if err == nil {
			return nil
		}

		// Calculate backoff duration (exponential backoff with jitter)
		backoff := config.GetHealthCheckRetryBackoff() * time.Duration(i+1)
		// sm.logger.Warn(ctx, "Database operation failed, retrying...",
		// 	observability.Error(err),
		// 	observability.Int("attempt", i+1),
		// 	observability.Int("max_retries", maxRetries),
		// 	observability.Duration("backoff", backoff),
		// )

		time.Sleep(backoff)
	}

	return fmt.Errorf("operation failed after %d retries: %w", maxRetries, err)
}
