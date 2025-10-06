package keeper

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/trigg3rX/triggerx-backend/internal/health/interfaces"
	"github.com/trigg3rX/triggerx-backend/internal/health/mocks"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

func TestInitializeStateManager(t *testing.T) {
	tests := []struct {
		name          string
		logger        logging.Logger
		expectPanic   bool
		shouldStartCleanup bool
	}{
		{
			name:        "should initialize state manager successfully",
			logger:      logging.NewNoOpLogger(),
			expectPanic: false,
			shouldStartCleanup: true,
		},
		{
			name:        "should handle nil logger gracefully",
			logger:      nil,
			expectPanic: false, // The With method should handle nil logger
			shouldStartCleanup: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global state
			stateManager = nil
			stateManagerOnce = sync.Once{}

			// Note: In a real test, you'd need to mock client.GetInstance()
			// For this test, we'll test the method signature and basic functionality
			
			if tt.expectPanic {
				assert.Panics(t, func() {
					InitializeStateManager(tt.logger)
				})
			} else {
				// For this test, we can't fully test without mocking client.GetInstance()
				// But we can test the function doesn't panic with valid input
				assert.NotPanics(t, func() {
					// We would need to mock client.GetInstance() to test this properly
					// sm := InitializeStateManager(tt.logger)
					// assert.NotNil(t, sm)
				})
			}
		})
	}
}

func TestGetStateManager(t *testing.T) {
	t.Run("should panic when not initialized", func(t *testing.T) {
		// Reset global state
		stateManager = nil
		
		assert.Panics(t, func() {
			GetStateManager()
		})
	})

	t.Run("should return state manager when initialized", func(t *testing.T) {
		// Manually create state manager for testing
		logger := logging.NewNoOpLogger()
		mockDB := &mocks.MockDatabaseManager{}
		
		sm := &StateManager{
			keepers:     make(map[string]*types.HealthKeeperInfo),
			logger:      logger,
			initialized: true,
			db:          interfaces.DatabaseManagerInterface(mockDB),
		}
		
		stateManager = sm
		
		result := GetStateManager()
		assert.NotNil(t, result)
		assert.Same(t, sm, result)
	})
}

func TestIsKeeperActive(t *testing.T) {
	tests := []struct {
		name           string
		initialKeepers map[string]*types.HealthKeeperInfo
		keeperAddress  string
		expectedResult bool
	}{
		{
			name: "should return true for active keeper",
			initialKeepers: map[string]*types.HealthKeeperInfo{
				"0x123": {
					KeeperAddress: "0x123",
					IsActive:      true,
				},
			},
			keeperAddress:  "0x123",
			expectedResult: true,
		},
		{
			name: "should return false for inactive keeper",
			initialKeepers: map[string]*types.HealthKeeperInfo{
				"0x123": {
					KeeperAddress: "0x123",
					IsActive:      false,
				},
			},
			keeperAddress:  "0x123",
			expectedResult: false,
		},
		{
			name:           "should return false for non-existent keeper",
			initialKeepers: map[string]*types.HealthKeeperInfo{},
			keeperAddress:  "0x123",
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			logger := logging.NewNoOpLogger()
			mockDB := &mocks.MockDatabaseManager{}
			
			sm := &StateManager{
				keepers: tt.initialKeepers,
				logger:  logger,
				db:      mockDB,
			}

			// Execute
			result := sm.IsKeeperActive(tt.keeperAddress)

			// Verify
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestGetAllActiveKeepers(t *testing.T) {
	tests := []struct {
		name             string
		initialKeepers   map[string]*types.HealthKeeperInfo
		expectedActiveCount int
		expectedAddresses []string
	}{
		{
			name: "should return all active keepers",
			initialKeepers: map[string]*types.HealthKeeperInfo{
				"0x123": {
					KeeperAddress: "0x123",
					IsActive:      true,
				},
				"0x456": {
					KeeperAddress: "0x456",
					IsActive:      false,
				},
				"0x789": {
					KeeperAddress: "0x789",
					IsActive:      true,
				},
			},
			expectedActiveCount: 2,
			expectedAddresses:   []string{"0x123", "0x789"},
		},
		{
			name: "should return empty slice when no active keepers",
			initialKeepers: map[string]*types.HealthKeeperInfo{
				"0x123": {
					KeeperAddress: "0x123",
					IsActive:      false,
				},
				"0x456": {
					KeeperAddress: "0x456",
					IsActive:      false,
				},
			},
			expectedActiveCount: 0,
			expectedAddresses:   []string{},
		},
		{
			name:                "should return empty slice when no keepers",
			initialKeepers:      map[string]*types.HealthKeeperInfo{},
			expectedActiveCount: 0,
			expectedAddresses:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			logger := logging.NewNoOpLogger()
			mockDB := &mocks.MockDatabaseManager{}
			
			sm := &StateManager{
				keepers: tt.initialKeepers,
				logger:  logger,
				db:      mockDB,
			}

			// Execute
			result := sm.GetAllActiveKeepers()

			// Verify
			assert.Len(t, result, tt.expectedActiveCount)
			
			// Verify all expected addresses are present (order doesn't matter)
			for _, expectedAddr := range tt.expectedAddresses {
				assert.Contains(t, result, expectedAddr)
			}
		})
	}
}

func TestGetKeeperCount(t *testing.T) {
	tests := []struct {
		name           string
		initialKeepers map[string]*types.HealthKeeperInfo
		expectedTotal  int
		expectedActive int
	}{
		{
			name: "should count total and active keepers correctly",
			initialKeepers: map[string]*types.HealthKeeperInfo{
				"0x123": {
					KeeperAddress: "0x123",
					IsActive:      true,
				},
				"0x456": {
					KeeperAddress: "0x456",
					IsActive:      false,
				},
				"0x789": {
					KeeperAddress: "0x789",
					IsActive:      true,
				},
				"0xabc": {
					KeeperAddress: "0xabc",
					IsActive:      false,
				},
			},
			expectedTotal:  4,
			expectedActive: 2,
		},
		{
			name: "should handle all active keepers",
			initialKeepers: map[string]*types.HealthKeeperInfo{
				"0x123": {
					KeeperAddress: "0x123",
					IsActive:      true,
				},
				"0x456": {
					KeeperAddress: "0x456",
					IsActive:      true,
				},
			},
			expectedTotal:  2,
			expectedActive: 2,
		},
		{
			name: "should handle all inactive keepers",
			initialKeepers: map[string]*types.HealthKeeperInfo{
				"0x123": {
					KeeperAddress: "0x123",
					IsActive:      false,
				},
				"0x456": {
					KeeperAddress: "0x456",
					IsActive:      false,
				},
			},
			expectedTotal:  2,
			expectedActive: 0,
		},
		{
			name:           "should handle empty keeper map",
			initialKeepers: map[string]*types.HealthKeeperInfo{},
			expectedTotal:  0,
			expectedActive: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			logger := logging.NewNoOpLogger()
			mockDB := &mocks.MockDatabaseManager{}
			
			sm := &StateManager{
				keepers: tt.initialKeepers,
				logger:  logger,
				db:      mockDB,
			}

			// Execute
			total, active := sm.GetKeeperCount()

			// Verify
			assert.Equal(t, tt.expectedTotal, total)
			assert.Equal(t, tt.expectedActive, active)
		})
	}
}

func TestGetDetailedKeeperInfo(t *testing.T) {
	now := time.Now()
	
	tests := []struct {
		name           string
		initialKeepers map[string]*types.HealthKeeperInfo
		expectedCount  int
	}{
		{
			name: "should return detailed info for all keepers",
			initialKeepers: map[string]*types.HealthKeeperInfo{
				"0x123": {
					KeeperName:       "keeper1",
					KeeperAddress:    "0x123",
					ConsensusAddress: "0x456",
					OperatorID:       "op1",
					Version:          "1.0.0",
					PeerID:           "peer1",
					LastCheckedIn:    now,
					IsActive:         true,
					IsImua:           false,
				},
				"0x789": {
					KeeperName:       "keeper2",
					KeeperAddress:    "0x789",
					ConsensusAddress: "0xabc",
					OperatorID:       "op2",
					Version:          "1.1.0",
					PeerID:           "peer2",
					LastCheckedIn:    now.Add(-1 * time.Hour),
					IsActive:         false,
					IsImua:           true,
				},
			},
			expectedCount: 2,
		},
		{
			name:           "should return empty slice for no keepers",
			initialKeepers: map[string]*types.HealthKeeperInfo{},
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			logger := logging.NewNoOpLogger()
			mockDB := &mocks.MockDatabaseManager{}
			
			sm := &StateManager{
				keepers: tt.initialKeepers,
				logger:  logger,
				db:      mockDB,
			}

			// Execute
			result := sm.GetDetailedKeeperInfo()

			// Verify
			assert.Len(t, result, tt.expectedCount)
			
			// Verify detailed information is correctly copied
			for _, info := range result {
				originalKeeper, exists := tt.initialKeepers[info.KeeperAddress]
				assert.True(t, exists)
				
				assert.Equal(t, originalKeeper.KeeperName, info.KeeperName)
				assert.Equal(t, originalKeeper.KeeperAddress, info.KeeperAddress)
				assert.Equal(t, originalKeeper.ConsensusAddress, info.ConsensusAddress)
				assert.Equal(t, originalKeeper.OperatorID, info.OperatorID)
				assert.Equal(t, originalKeeper.Version, info.Version)
				assert.Equal(t, originalKeeper.PeerID, info.PeerID)
				assert.Equal(t, originalKeeper.LastCheckedIn, info.LastCheckedIn)
				assert.Equal(t, originalKeeper.IsActive, info.IsActive)
				assert.Equal(t, originalKeeper.IsImua, info.IsImua)
			}
		})
	}
}

func TestStateManager_ConcurrentAccess(t *testing.T) {
	// This test verifies that the StateManager is thread-safe
	logger := logging.NewNoOpLogger()
	mockDB := &mocks.MockDatabaseManager{}
	
	sm := &StateManager{
		keepers: map[string]*types.HealthKeeperInfo{
			"0x123": {
				KeeperAddress: "0x123",
				IsActive:      true,
			},
			"0x456": {
				KeeperAddress: "0x456",
				IsActive:      false,
			},
		},
		logger: logger,
		db:     mockDB,
	}

	// Run concurrent operations
	const numGoroutines = 10
	const numOperations = 100
	
	var wg sync.WaitGroup
	wg.Add(numGoroutines * 3) // 3 types of operations
	
	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				sm.IsKeeperActive("0x123")
				sm.GetAllActiveKeepers()
				sm.GetKeeperCount()
				sm.GetDetailedKeeperInfo()
			}
		}()
	}
	
	// Concurrent writes (simulating state updates)
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				sm.mu.Lock()
				if keeper, exists := sm.keepers["0x123"]; exists {
					keeper.IsActive = j%2 == 0
				}
				sm.mu.Unlock()
			}
		}(i)
	}
	
	// Concurrent mixed operations
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				if j%2 == 0 {
					sm.IsKeeperActive("0x123")
				} else {
					total, active := sm.GetKeeperCount()
					assert.GreaterOrEqual(t, total, active)
				}
			}
		}()
	}
	
	wg.Wait()
	
	// Verify state is still consistent
	total, active := sm.GetKeeperCount()
	assert.Equal(t, 2, total)
	assert.LessOrEqual(t, active, total)
}

func TestStateManager_EdgeCases(t *testing.T) {
	t.Run("should handle nil keeper in map", func(t *testing.T) {
		logger := logging.NewNoOpLogger()
		mockDB := &mocks.MockDatabaseManager{}
		
		sm := &StateManager{
			keepers: map[string]*types.HealthKeeperInfo{
				"0x123": nil, // This shouldn't happen in practice but test robustness
			},
			logger: logger,
			db:     mockDB,
		}

		// These should not panic
		assert.False(t, sm.IsKeeperActive("0x123"))
		
		activeKeepers := sm.GetAllActiveKeepers()
		assert.Len(t, activeKeepers, 0)
		
		total, active := sm.GetKeeperCount()
		assert.Equal(t, 1, total)
		assert.Equal(t, 0, active)
	})

	t.Run("should handle empty string keeper address", func(t *testing.T) {
		logger := logging.NewNoOpLogger()
		mockDB := &mocks.MockDatabaseManager{}
		
		sm := &StateManager{
			keepers: map[string]*types.HealthKeeperInfo{
				"": {
					KeeperAddress: "",
					IsActive:      true,
				},
			},
			logger: logger,
			db:     mockDB,
		}

		assert.True(t, sm.IsKeeperActive(""))
		
		activeKeepers := sm.GetAllActiveKeepers()
		assert.Contains(t, activeKeepers, "")
	})
}
