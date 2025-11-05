package cache

import (
	"context"
	"time"
)

// RewardsCacheInterface defines the interface for rewards cache operations
type RewardsCacheInterface interface {
	// Daily uptime operations
	IncrementDailyUptime(ctx context.Context, keeperAddress string, seconds int64) error
	GetDailyUptime(ctx context.Context, keeperAddress string) (int64, error)
	GetAllDailyUptimes(ctx context.Context) (map[string]int64, error)
	ResetDailyUptime(ctx context.Context, keeperAddress string) error
	ResetAllDailyUptimes(ctx context.Context) error

	// Rewards metadata operations
	SetLastRewardsDistribution(ctx context.Context, timestamp time.Time) error
	GetLastRewardsDistribution(ctx context.Context) (time.Time, error)
	SetCurrentPeriodStart(ctx context.Context, timestamp time.Time) error
	GetCurrentPeriodStart(ctx context.Context) (time.Time, error)

	// Connection management
	Close() error
}
