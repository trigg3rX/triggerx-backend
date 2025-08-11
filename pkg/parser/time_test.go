package parser_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trigg3rX/triggerx-backend/pkg/parser"
)

func TestCalculateNextExecutionTime_IntervalSchedule_ValidInterval(t *testing.T) {
	tests := []struct {
		name                 string
		currentExecutionTime time.Time
		timeInterval         int64
		expectedNextTime     time.Time
		expectedError        bool
	}{
		{
			name:                 "positive interval 60 seconds",
			currentExecutionTime: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			timeInterval:         60,
			expectedNextTime:     time.Date(2024, 1, 1, 12, 1, 0, 0, time.UTC),
			expectedError:        false,
		},
		{
			name:                 "positive interval 3600 seconds (1 hour)",
			currentExecutionTime: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			timeInterval:         3600,
			expectedNextTime:     time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
			expectedError:        false,
		},
		{
			name:                 "positive interval 86400 seconds (1 day)",
			currentExecutionTime: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			timeInterval:         86400,
			expectedNextTime:     time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC),
			expectedError:        false,
		},
		{
			name:                 "zero interval should error",
			currentExecutionTime: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			timeInterval:         0,
			expectedNextTime:     time.Time{},
			expectedError:        true,
		},
		{
			name:                 "negative interval should error",
			currentExecutionTime: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			timeInterval:         -60,
			expectedNextTime:     time.Time{},
			expectedError:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextTime, err := parser.CalculateNextExecutionTime(
				tt.currentExecutionTime,
				"interval",
				tt.timeInterval,
				"",
				"",
			)

			if tt.expectedError {
				assert.Error(t, err)
				assert.True(t, nextTime.IsZero())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedNextTime, nextTime)
			}
		})
	}
}

func TestCalculateNextExecutionTime_CronSchedule_NotImplemented(t *testing.T) {
	tests := []struct {
		name                 string
		currentExecutionTime time.Time
		cronExpression       string
		expectedError        string
	}{
		{
			name:                 "empty cron expression",
			currentExecutionTime: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			cronExpression:       "",
			expectedError:        "cron expression is required for cron schedule type",
		},
		{
			name:                 "valid cron expression but not implemented",
			currentExecutionTime: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			cronExpression:       "0 0 * * *",
			expectedError:        "cron expression parsing not implemented yet",
		},
		{
			name:                 "invalid cron expression",
			currentExecutionTime: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			cronExpression:       "invalid cron",
			expectedError:        "cron expression parsing not implemented yet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextTime, err := parser.CalculateNextExecutionTime(
				tt.currentExecutionTime,
				"cron",
				0,
				tt.cronExpression,
				"",
			)

			assert.Error(t, err)
			assert.Equal(t, tt.expectedError, err.Error())
			assert.True(t, nextTime.IsZero())
		})
	}
}

func TestCalculateNextExecutionTime_SpecificSchedule_NotImplemented(t *testing.T) {
	tests := []struct {
		name                 string
		currentExecutionTime time.Time
		specificSchedule     string
		expectedError        string
	}{
		{
			name:                 "empty specific schedule",
			currentExecutionTime: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			specificSchedule:     "",
			expectedError:        "specific schedule is required for specific schedule type",
		},
		{
			name:                 "valid specific schedule but not implemented",
			currentExecutionTime: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			specificSchedule:     "1st day of month",
			expectedError:        "specific schedule parsing not implemented yet",
		},
		{
			name:                 "every Sunday 2 PM schedule",
			currentExecutionTime: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			specificSchedule:     "every Sunday 2 PM",
			expectedError:        "specific schedule parsing not implemented yet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextTime, err := parser.CalculateNextExecutionTime(
				tt.currentExecutionTime,
				"specific",
				0,
				"",
				tt.specificSchedule,
			)

			assert.Error(t, err)
			assert.Equal(t, tt.expectedError, err.Error())
			assert.True(t, nextTime.IsZero())
		})
	}
}

func TestCalculateNextExecutionTime_UnknownScheduleType_ReturnsError(t *testing.T) {
	tests := []struct {
		name          string
		scheduleType  string
		expectedError string
	}{
		{
			name:          "unknown schedule type",
			scheduleType:  "unknown",
			expectedError: "unknown schedule type: unknown",
		},
		{
			name:          "empty schedule type",
			scheduleType:  "",
			expectedError: "unknown schedule type: ",
		},
		{
			name:          "daily schedule type",
			scheduleType:  "daily",
			expectedError: "unknown schedule type: daily",
		},
		{
			name:          "weekly schedule type",
			scheduleType:  "weekly",
			expectedError: "unknown schedule type: weekly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextTime, err := parser.CalculateNextExecutionTime(
				time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
				tt.scheduleType,
				0,
				"",
				"",
			)

			assert.Error(t, err)
			assert.Equal(t, tt.expectedError, err.Error())
			assert.True(t, nextTime.IsZero())
		})
	}
}

func TestCalculateNextExecutionTime_EdgeCases(t *testing.T) {
	t.Run("very large interval", func(t *testing.T) {
		currentTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		largeInterval := int64(365 * 24 * 60 * 60) // 1 year in seconds

		nextTime, err := parser.CalculateNextExecutionTime(
			currentTime,
			"interval",
			largeInterval,
			"",
			"",
		)

		require.NoError(t, err)
		expectedTime := time.Date(2024, 12, 31, 12, 0, 0, 0, time.UTC) // 365 days later (2024 is leap year)
		assert.Equal(t, expectedTime, nextTime)
	})

	t.Run("very small interval", func(t *testing.T) {
		currentTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		smallInterval := int64(1) // 1 second

		nextTime, err := parser.CalculateNextExecutionTime(
			currentTime,
			"interval",
			smallInterval,
			"",
			"",
		)

		require.NoError(t, err)
		expectedTime := time.Date(2024, 1, 1, 12, 0, 1, 0, time.UTC)
		assert.Equal(t, expectedTime, nextTime)
	})

	t.Run("leap year handling", func(t *testing.T) {
		currentTime := time.Date(2024, 2, 28, 12, 0, 0, 0, time.UTC)
		interval := int64(24 * 60 * 60) // 1 day

		nextTime, err := parser.CalculateNextExecutionTime(
			currentTime,
			"interval",
			interval,
			"",
			"",
		)

		require.NoError(t, err)
		expectedTime := time.Date(2024, 2, 29, 12, 0, 0, 0, time.UTC) // Leap day
		assert.Equal(t, expectedTime, nextTime)
	})

	t.Run("timezone preservation", func(t *testing.T) {
		// Test with different timezone
		loc := time.FixedZone("EST", -5*60*60) // EST timezone
		currentTime := time.Date(2024, 1, 1, 12, 0, 0, 0, loc)
		interval := int64(60) // 1 minute

		nextTime, err := parser.CalculateNextExecutionTime(
			currentTime,
			"interval",
			interval,
			"",
			"",
		)

		require.NoError(t, err)
		expectedTime := time.Date(2024, 1, 1, 12, 1, 0, 0, loc)
		assert.Equal(t, expectedTime, nextTime)
		assert.Equal(t, currentTime.Location(), nextTime.Location())
	})
}

func TestCalculateNextExecutionTime_IntegrationScenarios(t *testing.T) {
	t.Run("multiple interval calculations", func(t *testing.T) {
		baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		interval := int64(60) // 1 minute

		// Calculate next execution time multiple times
		nextTime1, err := parser.CalculateNextExecutionTime(
			baseTime,
			"interval",
			interval,
			"",
			"",
		)
		require.NoError(t, err)

		nextTime2, err := parser.CalculateNextExecutionTime(
			nextTime1,
			"interval",
			interval,
			"",
			"",
		)
		require.NoError(t, err)

		nextTime3, err := parser.CalculateNextExecutionTime(
			nextTime2,
			"interval",
			interval,
			"",
			"",
		)
		require.NoError(t, err)

		// Verify the progression
		assert.Equal(t, time.Date(2024, 1, 1, 12, 1, 0, 0, time.UTC), nextTime1)
		assert.Equal(t, time.Date(2024, 1, 1, 12, 2, 0, 0, time.UTC), nextTime2)
		assert.Equal(t, time.Date(2024, 1, 1, 12, 3, 0, 0, time.UTC), nextTime3)
	})

	t.Run("mixed schedule type errors", func(t *testing.T) {
		currentTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

		// Test interval (should work)
		nextTime, err := parser.CalculateNextExecutionTime(
			currentTime,
			"interval",
			60,
			"",
			"",
		)
		require.NoError(t, err)
		assert.False(t, nextTime.IsZero())

		// Test cron (should fail)
		nextTime, err = parser.CalculateNextExecutionTime(
			currentTime,
			"cron",
			0,
			"0 0 * * *",
			"",
		)
		assert.Error(t, err)
		assert.True(t, nextTime.IsZero())

		// Test specific (should fail)
		nextTime, err = parser.CalculateNextExecutionTime(
			currentTime,
			"specific",
			0,
			"",
			"1st day of month",
		)
		assert.Error(t, err)
		assert.True(t, nextTime.IsZero())

		// Test unknown (should fail)
		nextTime, err = parser.CalculateNextExecutionTime(
			currentTime,
			"unknown",
			0,
			"",
			"",
		)
		assert.Error(t, err)
		assert.True(t, nextTime.IsZero())
	})
}
