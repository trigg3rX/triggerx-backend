package rewards

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/health/cache"
	"github.com/trigg3rX/triggerx-backend/internal/health/interfaces"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

const (
	// Reward tiers (in seconds)
	Tier1Threshold = 20 * 3600 // 20 hours
	Tier2Threshold = 15 * 3600 // 15 hours
	Tier3Threshold = 10 * 3600 // 10 hours
	MinimumUptime  = 6 * 3600  // 6 hours (minimum to get any rewards)

	// Daily base points per tier
	FullRewardPoints      = 1000 // Full reward for 20+ hours
	TwoThirdsRewardPoints = 667  // 2/3 reward for 15-20 hours
	OneThirdRewardPoints  = 333  // 1/3 reward for 10-15 hours

	// Distribution time (UTC)
	RewardDistributionHour   = 6
	RewardDistributionMinute = 30
)

// Service manages the daily rewards distribution
type Service struct {
	logger   logging.Logger
	cache    cache.RewardsCacheInterface
	database interfaces.DatabaseManagerInterface
	ctx      context.Context
	stopChan chan struct{}
}

// NewService creates a new rewards service
func NewService(
	logger logging.Logger,
	cache cache.RewardsCacheInterface,
	database interfaces.DatabaseManagerInterface,
) *Service {
	return &Service{
		logger:   logger.With("component", "rewards_service"),
		cache:    cache,
		database: database,
		ctx:      context.Background(),
		stopChan: make(chan struct{}),
	}
}

// Start begins the rewards distribution service
func (s *Service) Start() error {
	s.logger.Info("Starting daily rewards service...")

	// Initialize current period if not set
	ctx, cancel := context.WithTimeout(s.ctx, 10*time.Second)
	defer cancel()

	periodStart, err := s.cache.GetCurrentPeriodStart(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current period start: %w", err)
	}

	now := time.Now().UTC()
	if periodStart.IsZero() {
		// No period set, initialize to start of current day at distribution time
		periodStart = time.Date(now.Year(), now.Month(), now.Day(),
			RewardDistributionHour, RewardDistributionMinute, 0, 0, time.UTC)

		// If we're past today's distribution time, the period started today
		// Otherwise it started yesterday
		if now.Before(periodStart) {
			periodStart = periodStart.AddDate(0, 0, -1)
		}

		if err := s.cache.SetCurrentPeriodStart(ctx, periodStart); err != nil {
			s.logger.Error("Failed to set current period start", "error", err)
		} else {
			s.logger.Info("Initialized current period start", "start", periodStart)
		}
	}

	// Check for any missed distributions
	if err := s.catchUpMissedDistributions(); err != nil {
		s.logger.Error("Failed to catch up missed distributions", "error", err)
		// Continue anyway, we'll catch up on next cycle
	}

	// Start the scheduler
	go s.scheduleNextReward()

	s.logger.Info("Daily rewards service started successfully")
	return nil
}

// Stop stops the rewards service
func (s *Service) Stop() {
	s.logger.Info("Stopping rewards service...")
	close(s.stopChan)
}

// catchUpMissedDistributions handles any missed reward distributions
func (s *Service) catchUpMissedDistributions() error {
	ctx, cancel := context.WithTimeout(s.ctx, 30*time.Second)
	defer cancel()

	lastDistribution, err := s.cache.GetLastRewardsDistribution(ctx)
	if err != nil {
		return fmt.Errorf("failed to get last distribution time: %w", err)
	}

	now := time.Now().UTC()

	// If never distributed or last distribution was more than 24 hours ago
	if lastDistribution.IsZero() {
		s.logger.Info("No previous rewards distribution found, will start fresh")
		return nil
	}

	// Calculate how many distributions we missed
	nextDistribution := s.getNextDistributionTime(lastDistribution)
	missedCount := 0

	for nextDistribution.Before(now) {
		missedCount++
		s.logger.Warn("Detected missed rewards distribution",
			"scheduled_time", nextDistribution,
			"missed_count", missedCount)

		// For missed distributions, we can't accurately distribute rewards
		// because we don't have the historical daily uptime data
		// So we just log and move to next
		nextDistribution = nextDistribution.AddDate(0, 0, 1)
	}

	if missedCount > 0 {
		s.logger.Warn("Skipping missed distributions due to lack of historical data",
			"missed_count", missedCount)
	}

	return nil
}

// scheduleNextReward schedules and executes reward distributions
func (s *Service) scheduleNextReward() {
	for {
		now := time.Now().UTC()
		nextDistribution := s.calculateNextDistributionTime(now)
		waitDuration := nextDistribution.Sub(now)

		s.logger.Info("Next reward distribution scheduled",
			"time", nextDistribution.Format(time.RFC3339),
			"wait_duration", waitDuration.String())

		// Wait until distribution time
		select {
		case <-time.After(waitDuration):
			// Execute distribution
			s.logger.Info("Executing daily rewards distribution...")
			if err := s.distributeRewards(); err != nil {
				s.logger.Error("Failed to distribute daily rewards", "error", err)
			} else {
				s.logger.Info("Daily rewards distributed successfully")
			}

			// Sleep for a minute to avoid double execution
			time.Sleep(time.Minute)

		case <-s.stopChan:
			s.logger.Info("Rewards scheduler stopped")
			return
		}
	}
}

// calculateNextDistributionTime calculates the next distribution time
func (s *Service) calculateNextDistributionTime(now time.Time) time.Time {
	next := time.Date(now.Year(), now.Month(), now.Day(),
		RewardDistributionHour, RewardDistributionMinute, 0, 0, time.UTC)

	// If we're past today's distribution time, schedule for tomorrow
	if now.After(next) || now.Equal(next) {
		next = next.AddDate(0, 0, 1)
	}

	return next
}

// getNextDistributionTime gets the next distribution time after a given time
func (s *Service) getNextDistributionTime(after time.Time) time.Time {
	next := time.Date(after.Year(), after.Month(), after.Day(),
		RewardDistributionHour, RewardDistributionMinute, 0, 0, time.UTC)

	// Move to next day's distribution time
	next = next.AddDate(0, 0, 1)

	return next
}

// distributeRewards executes the daily reward distribution
func (s *Service) distributeRewards() error {
	ctx, cancel := context.WithTimeout(s.ctx, 60*time.Second)
	defer cancel()

	// Get all daily uptimes from Redis
	uptimes, err := s.cache.GetAllDailyUptimes(ctx)
	if err != nil {
		return fmt.Errorf("failed to get daily uptimes: %w", err)
	}

	s.logger.Info("Processing rewards for keepers", "keeper_count", len(uptimes))

	// Calculate and distribute rewards for each keeper
	rewardsSummary := make(map[string]int64)
	for keeperAddress, dailyUptime := range uptimes {
		points := s.calculateRewardPoints(dailyUptime)

		if points > 0 {
			// Add points to keeper's account
			if err := s.database.AddKeeperPoints(ctx, keeperAddress, points); err != nil {
				s.logger.Error("Failed to add keeper points",
					"keeper", keeperAddress,
					"points", points,
					"error", err)
				continue
			}

			rewardsSummary[keeperAddress] = points
			s.logger.Debug("Distributed rewards",
				"keeper", keeperAddress,
				"daily_uptime_hours", float64(dailyUptime)/3600,
				"points", points)
		}
	}

	// Update last distribution timestamp
	now := time.Now().UTC()
	if err := s.cache.SetLastRewardsDistribution(ctx, now); err != nil {
		s.logger.Error("Failed to update last distribution timestamp", "error", err)
	}

	// Update period start to now
	periodStart := time.Date(now.Year(), now.Month(), now.Day(),
		RewardDistributionHour, RewardDistributionMinute, 0, 0, time.UTC)
	if err := s.cache.SetCurrentPeriodStart(ctx, periodStart); err != nil {
		s.logger.Error("Failed to update period start", "error", err)
	}

	// Reset all daily uptime counters for new period
	if err := s.cache.ResetAllDailyUptimes(ctx); err != nil {
		s.logger.Error("Failed to reset daily uptime counters", "error", err)
	}

	s.logger.Info("Rewards distribution complete",
		"total_keepers_rewarded", len(rewardsSummary),
		"distribution_time", now.Format(time.RFC3339))

	return nil
}

// calculateRewardPoints calculates reward points based on daily uptime
func (s *Service) calculateRewardPoints(dailyUptimeSeconds int64) int64 {
	// Check minimum threshold
	if dailyUptimeSeconds < MinimumUptime {
		return 0
	}

	// Determine tier and assign points
	if dailyUptimeSeconds >= Tier1Threshold {
		return FullRewardPoints // 20+ hours: full reward
	} else if dailyUptimeSeconds >= Tier2Threshold {
		return TwoThirdsRewardPoints // 15-20 hours: 2/3 reward
	} else if dailyUptimeSeconds >= Tier3Threshold {
		return OneThirdRewardPoints // 10-15 hours: 1/3 reward
	}

	// Between 6-10 hours: some fractional reward
	// Linear interpolation between 0 and 1/3
	fraction := float64(dailyUptimeSeconds-MinimumUptime) / float64(Tier3Threshold-MinimumUptime)
	return int64(fraction * float64(OneThirdRewardPoints))
}

// GetRewardsHealth returns health status of the rewards service
func (s *Service) GetRewardsHealth() map[string]interface{} {
	ctx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
	defer cancel()

	health := map[string]interface{}{
		"service": "rewards",
		"status":  "running",
	}

	// Get last distribution time
	lastDistribution, err := s.cache.GetLastRewardsDistribution(ctx)
	if err != nil {
		health["last_distribution_error"] = err.Error()
		health["status"] = "degraded"
	} else if !lastDistribution.IsZero() {
		health["last_distribution"] = lastDistribution.Format(time.RFC3339)
		health["time_since_last_distribution"] = time.Since(lastDistribution).String()

		// Check if distribution is overdue (more than 25 hours)
		if time.Since(lastDistribution) > 25*time.Hour {
			health["status"] = "overdue"
			health["warning"] = "rewards distribution is overdue"
		}
	}

	// Get current period start
	periodStart, err := s.cache.GetCurrentPeriodStart(ctx)
	if err != nil {
		health["period_start_error"] = err.Error()
	} else if !periodStart.IsZero() {
		health["current_period_start"] = periodStart.Format(time.RFC3339)
		health["current_period_duration"] = time.Since(periodStart).String()
	}

	// Calculate next distribution time
	now := time.Now().UTC()
	nextDistribution := s.calculateNextDistributionTime(now)
	health["next_distribution"] = nextDistribution.Format(time.RFC3339)
	health["time_until_next_distribution"] = nextDistribution.Sub(now).String()

	// Get active uptime tracking count
	uptimes, err := s.cache.GetAllDailyUptimes(ctx)
	if err != nil {
		health["uptime_tracking_error"] = err.Error()
	} else {
		health["keepers_tracked"] = len(uptimes)
	}

	return health
}

// GetKeeperDailyUptime returns the daily uptime for a specific keeper
func (s *Service) GetKeeperDailyUptime(keeperAddress string) (time.Duration, error) {
	ctx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
	defer cancel()

	uptimeSeconds, err := s.cache.GetDailyUptime(ctx, keeperAddress)
	if err != nil {
		return 0, fmt.Errorf("failed to get daily uptime: %w", err)
	}

	return time.Duration(uptimeSeconds) * time.Second, nil
}

// FormatPoints formats points as a string (compatible with keeper_points field)
func FormatPoints(points int64) string {
	return new(big.Int).SetInt64(points).String()
}
