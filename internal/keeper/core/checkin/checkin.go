package checkin

import (
	"context"
	"errors"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

// Manager handles periodic health check-ins with the health service
type Manager struct {
	client     *Client
	logger     observability.Logger
	interval   time.Duration
	onShutdown func()
}

// ManagerConfig holds the configuration for the check-in manager
type ManagerConfig struct {
	Interval   time.Duration
	OnShutdown func() // Callback to trigger graceful shutdown when keeper is not verified
}

// NewManager creates a new check-in manager
func NewManager(client *Client, logger observability.Logger, cfg ManagerConfig) *Manager {
	if cfg.Interval == 0 {
		cfg.Interval = config.GetHealthCheckInterval()
	}

	return &Manager{
		client:     client,
		logger:     logger,
		interval:   cfg.Interval,
		onShutdown: cfg.OnShutdown,
	}
}

// PerformInitialCheckIn performs the initial health check-in during startup
// Returns error if the keeper is not verified or check-in fails
func (m *Manager) PerformInitialCheckIn(ctx context.Context) error {
	response, err := m.client.CheckIn(ctx)
	if err != nil {
		if errors.Is(err, ErrKeeperNotVerified) {
			return err
		}
		m.logger.Error(ctx, "Failed initial health check-in", observability.Any("error", response.Data))
		return err
	}

	if !response.Status {
		m.logger.Error(ctx, "Initial health check-in returned failure", observability.Any("data", response.Data))
		return errors.New("initial health check-in failed")
	}

	return nil
}

// Start begins the periodic health check-in routine
// This should be called in a goroutine
func (m *Manager) Start(ctx context.Context) {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.performPeriodicCheckIn(ctx)
		case <-ctx.Done():
			return
		}
	}
}

// performPeriodicCheckIn performs a single health check-in
func (m *Manager) performPeriodicCheckIn(ctx context.Context) {
	response, err := m.client.CheckIn(ctx)
	if err != nil {
		if errors.Is(err, ErrKeeperNotVerified) {
			m.logger.Error(ctx, "Keeper is not verified. Shutting down...", observability.Error(err))
			if m.onShutdown != nil {
				m.onShutdown()
			}
			return
		}
		m.logger.Error(ctx, "Failed health check-in", observability.Any("error", response.Data))
	}
}
