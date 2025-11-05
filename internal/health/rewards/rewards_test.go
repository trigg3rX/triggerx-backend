package rewards

import (
	"context"
	"testing"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/health/cache"
	"github.com/trigg3rX/triggerx-backend/internal/health/interfaces"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// Mock implementations
type mockRewardsCache struct {
	uptimes              map[string]int64
	lastDistribution     time.Time
	currentPeriodStart   time.Time
	incrementError       error
	getError             error
	getAllError          error
	resetAllError        error
	setDistributionError error
	setPeriodError       error
}

func (m *mockRewardsCache) IncrementDailyUptime(ctx context.Context, keeperAddress string, seconds int64) error {
	if m.incrementError != nil {
		return m.incrementError
	}
	if m.uptimes == nil {
		m.uptimes = make(map[string]int64)
	}
	m.uptimes[keeperAddress] += seconds
	return nil
}

func (m *mockRewardsCache) GetDailyUptime(ctx context.Context, keeperAddress string) (int64, error) {
	if m.getError != nil {
		return 0, m.getError
	}
	return m.uptimes[keeperAddress], nil
}

func (m *mockRewardsCache) GetAllDailyUptimes(ctx context.Context) (map[string]int64, error) {
	if m.getAllError != nil {
		return nil, m.getAllError
	}
	if m.uptimes == nil {
		return make(map[string]int64), nil
	}
	return m.uptimes, nil
}

func (m *mockRewardsCache) ResetDailyUptime(ctx context.Context, keeperAddress string) error {
	delete(m.uptimes, keeperAddress)
	return nil
}

func (m *mockRewardsCache) ResetAllDailyUptimes(ctx context.Context) error {
	if m.resetAllError != nil {
		return m.resetAllError
	}
	m.uptimes = make(map[string]int64)
	return nil
}

func (m *mockRewardsCache) SetLastRewardsDistribution(ctx context.Context, timestamp time.Time) error {
	if m.setDistributionError != nil {
		return m.setDistributionError
	}
	m.lastDistribution = timestamp
	return nil
}

func (m *mockRewardsCache) GetLastRewardsDistribution(ctx context.Context) (time.Time, error) {
	return m.lastDistribution, nil
}

func (m *mockRewardsCache) SetCurrentPeriodStart(ctx context.Context, timestamp time.Time) error {
	if m.setPeriodError != nil {
		return m.setPeriodError
	}
	m.currentPeriodStart = timestamp
	return nil
}

func (m *mockRewardsCache) GetCurrentPeriodStart(ctx context.Context) (time.Time, error) {
	return m.currentPeriodStart, nil
}

func (m *mockRewardsCache) Close() error {
	return nil
}

// Verify interface compliance
var _ cache.RewardsCacheInterface = (*mockRewardsCache)(nil)

type mockDatabase struct {
	pointsAdded map[string]int64
	addError    error
}

func (m *mockDatabase) AddKeeperPoints(ctx context.Context, keeperAddress string, points int64) error {
	if m.addError != nil {
		return m.addError
	}
	if m.pointsAdded == nil {
		m.pointsAdded = make(map[string]int64)
	}
	m.pointsAdded[keeperAddress] += points
	return nil
}

func (m *mockDatabase) UpdateKeeperStatus(ctx context.Context, keeperAddress string, consensusAddress string, version string, uptime int64, timestamp time.Time, publicIP string, isActive bool) error {
	return nil
}

func (m *mockDatabase) GetVerifiedKeepers(ctx context.Context) ([]types.HealthKeeperInfo, error) {
	return nil, nil
}

func (m *mockDatabase) UpdateAllKeepersStatus(ctx context.Context, onlineKeepers []types.HealthKeeperInfo) error {
	return nil
}

func (m *mockDatabase) UpdateKeeperChatID(ctx context.Context, keeperAddress string, chatID int64) error {
	return nil
}

func (m *mockDatabase) GetKeeperChatInfo(ctx context.Context, keeperAddress string) (int64, string, error) {
	return 0, "", nil
}

// Verify interface compliance
var _ interfaces.DatabaseManagerInterface = (*mockDatabase)(nil)

func testLogger() logging.Logger {
	logger, _ := logging.NewZapLogger(logging.LoggerConfig{
		ProcessName:   "test",
		IsDevelopment: true,
	})
	return logger
}

func TestCalculateRewardPoints(t *testing.T) {
	tests := []struct {
		name             string
		dailyUptimeHours float64
		expectedPoints   int64
	}{
		{"below minimum (5 hours)", 5, 0},
		{"minimum threshold (6 hours)", 6, 0},
		{"between 6-10 hours (7 hours)", 7, 83},  // ~25% of 333
		{"between 6-10 hours (8 hours)", 8, 166}, // ~50% of 333
		{"between 6-10 hours (9 hours)", 9, 249}, // ~75% of 333
		{"tier 3 minimum (10 hours)", 10, 333},   // 1/3 reward
		{"mid tier 3 (12 hours)", 12, 333},       // 1/3 reward
		{"tier 2 minimum (15 hours)", 15, 667},   // 2/3 reward
		{"mid tier 2 (17 hours)", 17, 667},       // 2/3 reward
		{"tier 1 minimum (20 hours)", 20, 1000},  // full reward
		{"above tier 1 (24 hours)", 24, 1000},    // full reward
	}

	logger := testLogger()
	mockCache := &mockRewardsCache{}
	mockDB := &mockDatabase{}
	service := NewService(logger, mockCache, mockDB)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uptimeSeconds := int64(tt.dailyUptimeHours * 3600)
			got := service.calculateRewardPoints(uptimeSeconds)

			// Allow small tolerance for rounding in fractional calculations
			tolerance := int64(5)
			diff := got - tt.expectedPoints
			if diff < 0 {
				diff = -diff
			}

			if diff > tolerance {
				t.Errorf("calculateRewardPoints(%v hours = %v seconds) = %v points, want %v points (diff: %v)",
					tt.dailyUptimeHours, uptimeSeconds, got, tt.expectedPoints, diff)
			}
		})
	}
}

func TestCalculateNextDistributionTime(t *testing.T) {
	logger := testLogger()
	mockCache := &mockRewardsCache{}
	mockDB := &mockDatabase{}
	service := NewService(logger, mockCache, mockDB)

	tests := []struct {
		name     string
		now      time.Time
		expected time.Time
	}{
		{
			name:     "before distribution time today",
			now:      time.Date(2025, 10, 16, 5, 0, 0, 0, time.UTC),
			expected: time.Date(2025, 10, 16, 6, 30, 0, 0, time.UTC),
		},
		{
			name:     "after distribution time today",
			now:      time.Date(2025, 10, 16, 8, 0, 0, 0, time.UTC),
			expected: time.Date(2025, 10, 17, 6, 30, 0, 0, time.UTC),
		},
		{
			name:     "exactly at distribution time",
			now:      time.Date(2025, 10, 16, 6, 30, 0, 0, time.UTC),
			expected: time.Date(2025, 10, 17, 6, 30, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.calculateNextDistributionTime(tt.now)
			if !got.Equal(tt.expected) {
				t.Errorf("calculateNextDistributionTime() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDistributeRewards(t *testing.T) {
	tests := []struct {
		name           string
		uptimes        map[string]int64
		expectedPoints map[string]int64
		cacheError     error
		dbError        error
		wantErr        bool
	}{
		{
			name: "distribute to multiple keepers",
			uptimes: map[string]int64{
				"0x123": 20 * 3600, // 20 hours = 1000 points
				"0x456": 15 * 3600, // 15 hours = 667 points
				"0x789": 10 * 3600, // 10 hours = 333 points
				"0xabc": 5 * 3600,  // 5 hours = 0 points
			},
			expectedPoints: map[string]int64{
				"0x123": 1000,
				"0x456": 667,
				"0x789": 333,
			},
			wantErr: false,
		},
		{
			name:           "no keepers to reward",
			uptimes:        map[string]int64{},
			expectedPoints: map[string]int64{},
			wantErr:        false,
		},
		{
			name: "cache error getting uptimes",
			uptimes: map[string]int64{
				"0x123": 20 * 3600,
			},
			cacheError: context.DeadlineExceeded,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := testLogger()
			mockCache := &mockRewardsCache{
				uptimes:     tt.uptimes,
				getAllError: tt.cacheError,
			}
			mockDB := &mockDatabase{
				addError: tt.dbError,
			}
			service := NewService(logger, mockCache, mockDB)

			err := service.distributeRewards()

			if (err != nil) != tt.wantErr {
				t.Errorf("distributeRewards() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify correct points were awarded
				for keeper, expectedPoints := range tt.expectedPoints {
					if mockDB.pointsAdded[keeper] != expectedPoints {
						t.Errorf("Keeper %s received %v points, want %v",
							keeper, mockDB.pointsAdded[keeper], expectedPoints)
					}
				}

				// Verify keepers below threshold didn't receive points
				for keeper := range tt.uptimes {
					if _, expected := tt.expectedPoints[keeper]; !expected {
						if points, got := mockDB.pointsAdded[keeper]; got && points > 0 {
							t.Errorf("Keeper %s should not have received points but got %v",
								keeper, points)
						}
					}
				}

				// Verify uptimes were reset
				if len(mockCache.uptimes) != 0 {
					t.Errorf("Expected uptimes to be reset, but %d remain", len(mockCache.uptimes))
				}
			}
		})
	}
}

func TestGetRewardsHealth(t *testing.T) {
	logger := testLogger()

	t.Run("healthy status", func(t *testing.T) {
		lastDist := time.Now().UTC().Add(-12 * time.Hour)
		mockCache := &mockRewardsCache{
			lastDistribution:   lastDist,
			currentPeriodStart: lastDist,
			uptimes: map[string]int64{
				"0x123": 3600,
				"0x456": 7200,
			},
		}
		mockDB := &mockDatabase{}
		service := NewService(logger, mockCache, mockDB)

		health := service.GetRewardsHealth()

		if health["service"] != "rewards" {
			t.Errorf("Expected service = rewards, got %v", health["service"])
		}
		if health["status"] != "running" {
			t.Errorf("Expected status = running, got %v", health["status"])
		}
		if health["keepers_tracked"] != 2 {
			t.Errorf("Expected keepers_tracked = 2, got %v", health["keepers_tracked"])
		}
	})

	t.Run("overdue status", func(t *testing.T) {
		lastDist := time.Now().UTC().Add(-26 * time.Hour)
		mockCache := &mockRewardsCache{
			lastDistribution:   lastDist,
			currentPeriodStart: lastDist,
		}
		mockDB := &mockDatabase{}
		service := NewService(logger, mockCache, mockDB)

		health := service.GetRewardsHealth()

		if health["status"] != "overdue" {
			t.Errorf("Expected status = overdue, got %v", health["status"])
		}
		if health["warning"] != "rewards distribution is overdue" {
			t.Errorf("Expected overdue warning, got %v", health["warning"])
		}
	})
}

func TestGetKeeperDailyUptime(t *testing.T) {
	logger := testLogger()
	mockCache := &mockRewardsCache{
		uptimes: map[string]int64{
			"0x123": 7200, // 2 hours
		},
	}
	mockDB := &mockDatabase{}
	service := NewService(logger, mockCache, mockDB)

	uptime, err := service.GetKeeperDailyUptime("0x123")
	if err != nil {
		t.Errorf("GetKeeperDailyUptime() error = %v", err)
	}

	expectedDuration := 2 * time.Hour
	if uptime != expectedDuration {
		t.Errorf("GetKeeperDailyUptime() = %v, want %v", uptime, expectedDuration)
	}
}

func TestFormatPoints(t *testing.T) {
	tests := []struct {
		points   int64
		expected string
	}{
		{0, "0"},
		{100, "100"},
		{1000, "1000"},
		{999999, "999999"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := FormatPoints(tt.points)
			if got != tt.expected {
				t.Errorf("FormatPoints(%d) = %s, want %s", tt.points, got, tt.expected)
			}
		})
	}
}

func TestRewardTiers(t *testing.T) {
	// Verify the reward tier constants
	if Tier1Threshold != 20*3600 {
		t.Errorf("Tier1Threshold = %d, want %d", Tier1Threshold, 20*3600)
	}
	if Tier2Threshold != 15*3600 {
		t.Errorf("Tier2Threshold = %d, want %d", Tier2Threshold, 15*3600)
	}
	if Tier3Threshold != 10*3600 {
		t.Errorf("Tier3Threshold = %d, want %d", Tier3Threshold, 10*3600)
	}
	if MinimumUptime != 6*3600 {
		t.Errorf("MinimumUptime = %d, want %d", MinimumUptime, 6*3600)
	}

	// Verify reward points
	if FullRewardPoints != 1000 {
		t.Errorf("FullRewardPoints = %d, want 1000", FullRewardPoints)
	}
	if TwoThirdsRewardPoints != 667 {
		t.Errorf("TwoThirdsRewardPoints = %d, want 667", TwoThirdsRewardPoints)
	}
	if OneThirdRewardPoints != 333 {
		t.Errorf("OneThirdRewardPoints = %d, want 333", OneThirdRewardPoints)
	}
}
