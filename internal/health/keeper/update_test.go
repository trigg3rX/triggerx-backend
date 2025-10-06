package keeper

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/trigg3rX/triggerx-backend/internal/health/interfaces"
	"github.com/trigg3rX/triggerx-backend/internal/health/mocks"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

func TestUpdateKeeperHealth(t *testing.T) {
	now := time.Now().UTC()
	
	tests := []struct {
		name           string
		initialKeepers map[string]*types.HealthKeeperInfo
		keeperHealth   types.KeeperHealthCheckIn
		setupMocks     func(*mocks.MockDatabaseManager)
		expectedError  string
		verifyResult   func(*testing.T, *StateManager, types.KeeperHealthCheckIn)
	}{
		{
			name: "should update existing verified keeper successfully",
			initialKeepers: map[string]*types.HealthKeeperInfo{
				"0x123": {
					KeeperName:       "test-keeper",
					KeeperAddress:    "0x123",
					ConsensusAddress: "0x456",
					OperatorID:       "op1",
					Version:          "1.0.0",
					PeerID:           "old-peer",
					IsActive:         false,
					LastCheckedIn:    now.Add(-1 * time.Hour),
					IsImua:           false,
				},
			},
			keeperHealth: types.KeeperHealthCheckIn{
				KeeperAddress:    "0x123",
				ConsensusAddress: "0x789",
				Version:          "1.1.0",
				PeerID:           "new-peer",
				Timestamp:        now,
				IsImua:           true,
			},
			setupMocks: func(mockDB *mocks.MockDatabaseManager) {
				mockDB.On("UpdateKeeperHealth", mock.Anything, true).Return(nil)
			},
			expectedError: "",
			verifyResult: func(t *testing.T, sm *StateManager, health types.KeeperHealthCheckIn) {
				keeper := sm.keepers[health.KeeperAddress]
				assert.NotNil(t, keeper)
				assert.Equal(t, health.Version, keeper.Version)
				assert.Equal(t, health.PeerID, keeper.PeerID)
				assert.True(t, keeper.IsActive)
				assert.True(t, keeper.IsImua)
				assert.WithinDuration(t, time.Now().UTC(), keeper.LastCheckedIn, 5*time.Second)
			},
		},
		{
			name:           "should return error for unverified keeper",
			initialKeepers: map[string]*types.HealthKeeperInfo{},
			keeperHealth: types.KeeperHealthCheckIn{
				KeeperAddress:    "0x123",
				ConsensusAddress: "0x456",
				Version:          "1.0.0",
				PeerID:           "peer123",
				Timestamp:        now,
			},
			setupMocks: func(mockDB *mocks.MockDatabaseManager) {
				// No database calls expected for unverified keeper
			},
			expectedError: "keeper not verified",
			verifyResult: func(t *testing.T, sm *StateManager, health types.KeeperHealthCheckIn) {
				// Keeper should not exist in the map
				_, exists := sm.keepers[health.KeeperAddress]
				assert.False(t, exists)
			},
		},
		{
			name: "should handle database update failure with retries",
			initialKeepers: map[string]*types.HealthKeeperInfo{
				"0x123": {
					KeeperAddress: "0x123",
					IsActive:      false,
					LastCheckedIn: now.Add(-1 * time.Hour),
				},
			},
			keeperHealth: types.KeeperHealthCheckIn{
				KeeperAddress: "0x123",
				Version:       "1.0.0",
				PeerID:        "peer123",
				Timestamp:     now,
			},
			setupMocks: func(mockDB *mocks.MockDatabaseManager) {
				mockDB.On("UpdateKeeperHealth", mock.Anything, true).Return(errors.New("database error"))
			},
			expectedError: "failed to update keeper status in database",
			verifyResult: func(t *testing.T, sm *StateManager, health types.KeeperHealthCheckIn) {
				// State should still be updated locally even if database fails
				keeper := sm.keepers[health.KeeperAddress]
				assert.NotNil(t, keeper)
				assert.Equal(t, health.Version, keeper.Version)
				assert.Equal(t, health.PeerID, keeper.PeerID)
				assert.True(t, keeper.IsActive)
			},
		},
		{
			name: "should update keeper with empty peer ID",
			initialKeepers: map[string]*types.HealthKeeperInfo{
				"0x123": {
					KeeperAddress: "0x123",
					IsActive:      false,
					LastCheckedIn: now.Add(-1 * time.Hour),
				},
			},
			keeperHealth: types.KeeperHealthCheckIn{
				KeeperAddress: "0x123",
				Version:       "1.0.0",
				PeerID:        "", // Empty peer ID
				Timestamp:     now,
			},
			setupMocks: func(mockDB *mocks.MockDatabaseManager) {
				mockDB.On("UpdateKeeperHealth", mock.Anything, true).Return(nil)
			},
			expectedError: "",
			verifyResult: func(t *testing.T, sm *StateManager, health types.KeeperHealthCheckIn) {
				keeper := sm.keepers[health.KeeperAddress]
				assert.NotNil(t, keeper)
				assert.Equal(t, "", keeper.PeerID) // Should preserve empty peer ID
				assert.True(t, keeper.IsActive)
			},
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

			// Execute
			err := sm.UpdateKeeperHealth(tt.keeperHealth)

			// Verify error expectation
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			// Verify result
			if tt.verifyResult != nil {
				tt.verifyResult(t, sm, tt.keeperHealth)
			}

			// Verify mock expectations
			mockDB.AssertExpectations(t)
		})
	}
}

func TestUpdateKeeperStatusInDatabase(t *testing.T) {
	tests := []struct {
		name          string
		keeperHealth  types.KeeperHealthCheckIn
		isActive      bool
		setupMocks    func(*mocks.MockDatabaseManager)
		expectedError string
	}{
		{
			name: "should update database successfully",
			keeperHealth: types.KeeperHealthCheckIn{
				KeeperAddress:    "0x123",
				ConsensusAddress: "0x456",
				Version:          "1.0.0",
				PeerID:           "peer123",
				Timestamp:        time.Now(),
			},
			isActive: true,
			setupMocks: func(mockDB *mocks.MockDatabaseManager) {
				mockDB.On("UpdateKeeperHealth", mock.Anything, true).Return(nil)
			},
			expectedError: "",
		},
		{
			name: "should handle database error",
			keeperHealth: types.KeeperHealthCheckIn{
				KeeperAddress: "0x123",
				Version:       "1.0.0",
				Timestamp:     time.Now(),
			},
			isActive: false,
			setupMocks: func(mockDB *mocks.MockDatabaseManager) {
				mockDB.On("UpdateKeeperHealth", mock.Anything, false).Return(errors.New("database connection failed"))
			},
			expectedError: "failed to update keeper status in database: database connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			logger := logging.NewNoOpLogger()
			mockDB := &mocks.MockDatabaseManager{}
			tt.setupMocks(mockDB)

			sm := &StateManager{
				logger: logger,
				db:     mockDB,
			}

			// Execute
			err := sm.updateKeeperStatusInDatabase(tt.keeperHealth, tt.isActive)

			// Verify
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			// Verify mock expectations
			mockDB.AssertExpectations(t)
		})
	}
}

func TestErrKeeperNotVerified(t *testing.T) {
	// Test that the custom error is properly defined
	assert.Equal(t, "keeper not verified", ErrKeeperNotVerified.Error())
	assert.Equal(t, ErrKeeperNotVerified, errors.New("keeper not verified"))
}

func TestMaxRetries(t *testing.T) {
	// Test that the maxRetries constant is properly defined
	assert.Equal(t, 3, maxRetries)
}

func TestUpdateKeeperHealth_ConcurrentAccess(t *testing.T) {
	// Test concurrent updates to the same keeper
	logger := logging.NewNoOpLogger()
	mockDB := &mocks.MockDatabaseManager{}

	initialKeepers := map[string]*types.HealthKeeperInfo{
		"0x123": {
			KeeperAddress: "0x123",
			IsActive:      false,
			LastCheckedIn: time.Now().Add(-1 * time.Hour),
		},
	}

	// Setup mock to handle multiple concurrent calls
	mockDB.On("UpdateKeeperHealth", mock.Anything, true).Return(nil)

	sm := &StateManager{
		keepers: initialKeepers,
		logger:  logger,
		db:      interfaces.DatabaseManagerInterface(mockDB),
	}

	// Run concurrent updates
	const numGoroutines = 10
	const numUpdates = 5

	doneChan := make(chan error, numGoroutines*numUpdates)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < numUpdates; j++ {
				health := types.KeeperHealthCheckIn{
					KeeperAddress:    "0x123",
					ConsensusAddress: "0x456",
					Version:          "1.0.0",
					PeerID:           "peer123",
					Timestamp:        time.Now(),
					IsImua:           j%2 == 0, // Alternate IsImua value
				}
				
				err := sm.UpdateKeeperHealth(health)
				doneChan <- err
			}
		}(i)
	}

	// Collect results
	var errors []error
	for i := 0; i < numGoroutines*numUpdates; i++ {
		if err := <-doneChan; err != nil {
			errors = append(errors, err)
		}
	}

	// Verify no errors occurred
	assert.Empty(t, errors)

	// Verify final state
	keeper := sm.keepers["0x123"]
	assert.NotNil(t, keeper)
	assert.True(t, keeper.IsActive)
	assert.Equal(t, "1.0.0", keeper.Version)
	assert.Equal(t, "peer123", keeper.PeerID)

	// Verify mock was called the expected number of times
	mockDB.AssertNumberOfCalls(t, "UpdateKeeperHealth", numGoroutines*numUpdates)
}

func TestUpdateKeeperHealth_EdgeCases(t *testing.T) {
	logger := logging.NewNoOpLogger()
	mockDB := &mocks.MockDatabaseManager{}

	t.Run("should handle keeper with all fields", func(t *testing.T) {
		now := time.Now().UTC()
		
		initialKeepers := map[string]*types.HealthKeeperInfo{
			"0x123": {
				KeeperName:       "test-keeper",
				KeeperAddress:    "0x123",
				ConsensusAddress: "0x456",
				OperatorID:       "op1",
				Version:          "1.0.0",
				PeerID:           "old-peer",
				IsActive:         false,
				LastCheckedIn:    now.Add(-1 * time.Hour),
				IsImua:           false,
			},
		}

		mockDB.On("UpdateKeeperHealth", mock.Anything, true).Return(nil)

		sm := &StateManager{
			keepers: initialKeepers,
			logger:  logger,
			db:      interfaces.DatabaseManagerInterface(mockDB),
		}

		health := types.KeeperHealthCheckIn{
			KeeperAddress:    "0x123",
			ConsensusPubKey:  "pubkey123",
			ConsensusAddress: "0x789",
			Version:          "2.0.0",
			Timestamp:        now,
			Signature:        "sig123",
			PeerID:           "new-peer",
			IsImua:           true,
		}

		err := sm.UpdateKeeperHealth(health)
		assert.NoError(t, err)

		// Verify all fields were updated appropriately
		keeper := sm.keepers["0x123"]
		assert.Equal(t, "2.0.0", keeper.Version)
		assert.Equal(t, "new-peer", keeper.PeerID)
		assert.True(t, keeper.IsActive)
		assert.True(t, keeper.IsImua)
		assert.WithinDuration(t, now, keeper.LastCheckedIn, 5*time.Second)

		// Original fields should remain unchanged
		assert.Equal(t, "test-keeper", keeper.KeeperName)
		assert.Equal(t, "0x123", keeper.KeeperAddress)
		assert.Equal(t, "op1", keeper.OperatorID)

		mockDB.AssertExpectations(t)
	})

	t.Run("should handle zero timestamp", func(t *testing.T) {
		initialKeepers := map[string]*types.HealthKeeperInfo{
			"0x123": {
				KeeperAddress: "0x123",
				IsActive:      false,
			},
		}

		mockDB.On("UpdateKeeperHealth", mock.Anything, true).Return(nil)

		sm := &StateManager{
			keepers: initialKeepers,
			logger:  logger,
			db:      interfaces.DatabaseManagerInterface(mockDB),
		}

		health := types.KeeperHealthCheckIn{
			KeeperAddress: "0x123",
			Version:       "1.0.0",
			Timestamp:     time.Time{}, // Zero timestamp
		}

		err := sm.UpdateKeeperHealth(health)
		assert.NoError(t, err)

		// Should use current time instead of zero timestamp
		keeper := sm.keepers["0x123"]
		assert.WithinDuration(t, time.Now().UTC(), keeper.LastCheckedIn, 5*time.Second)

		mockDB.AssertExpectations(t)
	})
}

func TestUpdateKeeperHealth_DatabaseRetry(t *testing.T) {
	// Test the retry mechanism with database failures
	logger := logging.NewNoOpLogger()
	mockDB := &mocks.MockDatabaseManager{}

	initialKeepers := map[string]*types.HealthKeeperInfo{
		"0x123": {
			KeeperAddress: "0x123",
			IsActive:      false,
		},
	}

	// Setup mock to fail first two times, then succeed
	mockDB.On("UpdateKeeperHealth", mock.Anything, true).Return(errors.New("temporary database error")).Times(2)
	mockDB.On("UpdateKeeperHealth", mock.Anything, true).Return(nil).Once()

	sm := &StateManager{
		keepers: initialKeepers,
		logger:  logger,
		db:      interfaces.DatabaseManagerInterface(mockDB),
	}

	health := types.KeeperHealthCheckIn{
		KeeperAddress: "0x123",
		Version:       "1.0.0",
		Timestamp:     time.Now(),
	}

	err := sm.UpdateKeeperHealth(health)
	assert.NoError(t, err)

	// State should still be updated
	keeper := sm.keepers["0x123"]
	assert.True(t, keeper.IsActive)

	// Verify mock was called the expected number of times (2 failures + 1 success)
	mockDB.AssertExpectations(t)
}
