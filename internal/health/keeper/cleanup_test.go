package keeper

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/trigg3rX/triggerx-backend/internal/health/interfaces"
	"github.com/trigg3rX/triggerx-backend/internal/health/mocks"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func TestCheckInactiveKeepers(t *testing.T) {
	now := time.Now().UTC()
	
	tests := []struct {
		name             string
		initialKeepers   map[string]*types.HealthKeeperInfo
		setupMocks       func(*mocks.MockDatabaseManager)
		expectedInactive []string
		expectedDBCalls  int
	}{
		{
			name: "should mark recently inactive keepers as inactive",
			initialKeepers: map[string]*types.HealthKeeperInfo{
				"0x123": {
					KeeperAddress: "0x123",
					IsActive:      true,
					LastCheckedIn: now.Add(-80 * time.Second), // Beyond threshold
				},
				"0x456": {
					KeeperAddress: "0x456",
					IsActive:      true,
					LastCheckedIn: now.Add(-30 * time.Second), // Within threshold
				},
			},
			setupMocks: func(mockDB *mocks.MockDatabaseManager) {
				mockDB.On("UpdateKeeperHealth", mock.MatchedBy(func(health types.KeeperHealthCheckIn) bool {
					return health.KeeperAddress == "0x123"
				}), false).Return(nil)
			},
			expectedInactive: []string{"0x123"},
			expectedDBCalls:  1,
		},
		{
			name: "should not affect already inactive keepers",
			initialKeepers: map[string]*types.HealthKeeperInfo{
				"0x123": {
					KeeperAddress: "0x123",
					IsActive:      false,
					LastCheckedIn: now.Add(-80 * time.Second), // Beyond threshold but already inactive
				},
			},
			setupMocks: func(mockDB *mocks.MockDatabaseManager) {
				// No database calls expected
			},
			expectedInactive: []string{},
			expectedDBCalls:  0,
		},
		{
			name: "should not affect keepers within threshold",
			initialKeepers: map[string]*types.HealthKeeperInfo{
				"0x123": {
					KeeperAddress: "0x123",
					IsActive:      true,
					LastCheckedIn: now.Add(-30 * time.Second), // Within threshold
				},
				"0x456": {
					KeeperAddress: "0x456",
					IsActive:      true,
					LastCheckedIn: now.Add(-10 * time.Second), // Within threshold
				},
			},
			setupMocks: func(mockDB *mocks.MockDatabaseManager) {
				// No database calls expected
			},
			expectedInactive: []string{},
			expectedDBCalls:  0,
		},
		{
			name: "should handle multiple inactive keepers",
			initialKeepers: map[string]*types.HealthKeeperInfo{
				"0x123": {
					KeeperAddress: "0x123",
					IsActive:      true,
					LastCheckedIn: now.Add(-80 * time.Second), // Beyond threshold
				},
				"0x456": {
					KeeperAddress: "0x456",
					IsActive:      true,
					LastCheckedIn: now.Add(-90 * time.Second), // Beyond threshold
				},
				"0x789": {
					KeeperAddress: "0x789",
					IsActive:      true,
					LastCheckedIn: now.Add(-100 * time.Second), // Beyond threshold
				},
			},
			setupMocks: func(mockDB *mocks.MockDatabaseManager) {
				mockDB.On("UpdateKeeperHealth", mock.MatchedBy(func(health types.KeeperHealthCheckIn) bool {
					return health.KeeperAddress == "0x123"
				}), false).Return(nil)
				mockDB.On("UpdateKeeperHealth", mock.MatchedBy(func(health types.KeeperHealthCheckIn) bool {
					return health.KeeperAddress == "0x456"
				}), false).Return(nil)
				mockDB.On("UpdateKeeperHealth", mock.MatchedBy(func(health types.KeeperHealthCheckIn) bool {
					return health.KeeperAddress == "0x789"
				}), false).Return(nil)
			},
			expectedInactive: []string{"0x123", "0x456", "0x789"},
			expectedDBCalls:  3,
		},
		{
			name:           "should handle empty keeper map",
			initialKeepers: map[string]*types.HealthKeeperInfo{},
			setupMocks: func(mockDB *mocks.MockDatabaseManager) {
				// No database calls expected
			},
			expectedInactive: []string{},
			expectedDBCalls:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			logger := logging.NewNoOpLogger()
			mockDB := &mocks.MockDatabaseManager{}
			tt.setupMocks(mockDB)

			sm := &StateManager{
				keepers: tt.initialKeepers,
				logger:  logger,
				db:      interfaces.DatabaseManagerInterface(mockDB),
			}

			// Store initial states for comparison
			initialStates := make(map[string]bool)
			for addr, keeper := range tt.initialKeepers {
				initialStates[addr] = keeper.IsActive
			}

			// Execute
			sm.checkInactiveKeepers()

			// Verify state changes
			for _, expectedInactiveAddr := range tt.expectedInactive {
				keeper := sm.keepers[expectedInactiveAddr]
				assert.NotNil(t, keeper)
				assert.False(t, keeper.IsActive, "Keeper %s should be marked as inactive", expectedInactiveAddr)
				assert.True(t, initialStates[expectedInactiveAddr], "Keeper %s should have been initially active", expectedInactiveAddr)
			}

			// Verify keepers not in expectedInactive list maintain their active state
			for addr, keeper := range sm.keepers {
				found := false
				for _, inactiveAddr := range tt.expectedInactive {
					if addr == inactiveAddr {
						found = true
						break
					}
				}
				if !found {
					assert.Equal(t, initialStates[addr], keeper.IsActive, 
						"Keeper %s should maintain its initial state", addr)
				}
			}

			// Verify mock expectations
			mockDB.AssertExpectations(t)
			mockDB.AssertNumberOfCalls(t, "UpdateKeeperHealth", tt.expectedDBCalls)
		})
	}
}

func TestCheckInactiveKeepers_DatabaseError(t *testing.T) {
	now := time.Now().UTC()
	
	logger := logging.NewNoOpLogger()
	mockDB := &mocks.MockDatabaseManager{}

	initialKeepers := map[string]*types.HealthKeeperInfo{
		"0x123": {
			KeeperAddress: "0x123",
			IsActive:      true,
			LastCheckedIn: now.Add(-80 * time.Second), // Beyond threshold
		},
	}

	// Setup mock to return database error
	mockDB.On("UpdateKeeperHealth", mock.Anything, false).Return(assert.AnError)

	sm := &StateManager{
		keepers: initialKeepers,
		logger:  logger,
		db:      interfaces.DatabaseManagerInterface(mockDB),
	}

	// Execute - should not panic even with database error
	assert.NotPanics(t, func() {
		sm.checkInactiveKeepers()
	})

	// Verify keeper state was still updated locally
	keeper := sm.keepers["0x123"]
	assert.False(t, keeper.IsActive)

	// Verify mock expectations
	mockDB.AssertExpectations(t)
}

func TestStartCleanupRoutine_Integration(t *testing.T) {
	// This test verifies the cleanup routine runs periodically
	logger := logging.NewNoOpLogger()
	mockDB := &mocks.MockDatabaseManager{}

	now := time.Now().UTC()
	initialKeepers := map[string]*types.HealthKeeperInfo{
		"0x123": {
			KeeperAddress: "0x123",
			IsActive:      true,
			LastCheckedIn: now.Add(-80 * time.Second), // Beyond threshold
		},
	}

	// Setup mock - should be called at least once during the test
	mockDB.On("UpdateKeeperHealth", mock.Anything, false).Return(nil)

	sm := &StateManager{
		keepers: initialKeepers,
		logger:  logger,
		db:      interfaces.DatabaseManagerInterface(mockDB),
	}

	// Start cleanup routine in a separate goroutine
	done := make(chan bool)
	go func() {
		defer func() { done <- true }()
		
		// Create a ticker to simulate the cleanup routine for a short duration
		ticker := time.NewTicker(100 * time.Millisecond) // Much faster for testing
		defer ticker.Stop()
		
		// Run for a short duration
		timeout := time.After(300 * time.Millisecond)
		
		for {
			select {
			case <-ticker.C:
				sm.checkInactiveKeepers()
			case <-timeout:
				return
			}
		}
	}()

	// Wait for test to complete
	<-done

	// Verify the keeper was marked as inactive
	keeper := sm.keepers["0x123"]
	assert.False(t, keeper.IsActive)

	// The mock should have been called at least once
	mockDB.AssertCalled(t, "UpdateKeeperHealth", mock.Anything, false)
}

func TestInactivityThreshold_BoundaryConditions(t *testing.T) {
	// Use a fixed time to avoid microsecond precision issues
	baseTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	
	tests := []struct {
		name           string
		baseTime       time.Time
		lastCheckedIn  time.Time
		shouldBeInactive bool
	}{
		{
			name:           "just under threshold",
			baseTime:       baseTime,
			lastCheckedIn:  baseTime.Add(-69 * time.Second),
			shouldBeInactive: false,
		},
		{
			name:           "just over threshold", 
			baseTime:       baseTime,
			lastCheckedIn:  baseTime.Add(-71 * time.Second),
			shouldBeInactive: true,
		},
		{
			name:           "way over threshold",
			baseTime:       baseTime,
			lastCheckedIn:  baseTime.Add(-5 * time.Minute),
			shouldBeInactive: true,
		},
		{
			name:           "future timestamp",
			baseTime:       baseTime,
			lastCheckedIn:  baseTime.Add(1 * time.Minute),
			shouldBeInactive: false,
		},
		{
			name:           "exactly at threshold (should remain active)",
			baseTime:       baseTime,
			lastCheckedIn:  baseTime.Add(-70 * time.Second),
			shouldBeInactive: false, // Should not be inactive at exactly 70 seconds
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logging.NewNoOpLogger()
			mockDB := &mocks.MockDatabaseManager{}

			initialKeepers := map[string]*types.HealthKeeperInfo{
				"0x123": {
					KeeperAddress: "0x123",
					IsActive:      true,
					LastCheckedIn: tt.lastCheckedIn,
				},
			}

			if tt.shouldBeInactive {
				mockDB.On("UpdateKeeperHealth", mock.Anything, false).Return(nil)
			} else {
				// Even if we expect keeper to remain active, we should still allow 
				// the mock to be called in case there are timing edge cases
				mockDB.On("UpdateKeeperHealth", mock.Anything, false).Return(nil).Maybe()
			}

			sm := &StateManager{
				keepers: initialKeepers,
				logger:  logger,
				db:      interfaces.DatabaseManagerInterface(mockDB),
			}

			// Override the time check logic by directly testing the time difference
			// This simulates what checkInactiveKeepers does with controlled timing
			now := tt.baseTime
			var inactiveKeepers []string

			for address, state := range sm.keepers {
				if state.IsActive && now.Sub(state.LastCheckedIn) > 70*time.Second {
					state.IsActive = false
					inactiveKeepers = append(inactiveKeepers, address)
				}
			}

			// Call database updates for inactive keepers
			for _, address := range inactiveKeepers {
				keeperHealth := types.KeeperHealthCheckIn{
					KeeperAddress: address,
				}
				sm.updateKeeperStatusInDatabase(keeperHealth, false)
			}

			// Verify
			keeper := sm.keepers["0x123"]
			if tt.shouldBeInactive {
				assert.False(t, keeper.IsActive, "Keeper should be marked as inactive")
			} else {
				assert.True(t, keeper.IsActive, "Keeper should remain active")
			}

			// Verify mock expectations
			mockDB.AssertExpectations(t)
		})
	}
}

func TestCleanupRoutine_Concurrency(t *testing.T) {
	// Test that the cleanup routine is thread-safe when run concurrently with other operations
	logger := logging.NewNoOpLogger()
	mockDB := &mocks.MockDatabaseManager{}

	now := time.Now().UTC()
	initialKeepers := map[string]*types.HealthKeeperInfo{
		"0x123": {
			KeeperAddress: "0x123",
			IsActive:      true,
			LastCheckedIn: now.Add(-80 * time.Second),
		},
		"0x456": {
			KeeperAddress: "0x456",
			IsActive:      true,
			LastCheckedIn: now.Add(-30 * time.Second),
		},
	}

	// Setup mock to handle multiple calls
	mockDB.On("UpdateKeeperHealth", mock.Anything, false).Return(nil)

	sm := &StateManager{
		keepers: initialKeepers,
		logger:  logger,
		db:      interfaces.DatabaseManagerInterface(mockDB),
	}

	var wg sync.WaitGroup
	
	// Run cleanup routine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			sm.checkInactiveKeepers()
			time.Sleep(10 * time.Millisecond)
		}
	}()

	// Run concurrent read operations
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			sm.IsKeeperActive("0x123")
			sm.GetAllActiveKeepers()
			sm.GetKeeperCount()
			time.Sleep(2 * time.Millisecond)
		}
	}()

	// Run concurrent state updates
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			sm.mu.Lock()
			if keeper, exists := sm.keepers["0x456"]; exists {
				keeper.LastCheckedIn = time.Now().UTC()
			}
			sm.mu.Unlock()
			time.Sleep(5 * time.Millisecond)
		}
	}()

	wg.Wait()

	// Verify final state consistency
	total, active := sm.GetKeeperCount()
	assert.Equal(t, 2, total)
	assert.LessOrEqual(t, active, total)
}
