package keeper

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/trigg3rX/triggerx-backend/internal/health/interfaces"
	"github.com/trigg3rX/triggerx-backend/internal/health/mocks"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func TestLoadVerifiedKeepers(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*mocks.MockDatabaseManager)
		expectedError string
		expectedCount int
	}{
		{
			name: "should load verified keepers successfully",
			setupMocks: func(mockDB *mocks.MockDatabaseManager) {
				keepers := []types.HealthKeeperInfo{
					{
						KeeperName:       "keeper1",
						KeeperAddress:    "0x123",
						ConsensusAddress: "0x456",
						OperatorID:       "op1",
						Version:          "1.0.0",
						PeerID:           "peer1",
						LastCheckedIn:    time.Now(),
						IsImua:           false,
					},
					{
						KeeperName:       "keeper2",
						KeeperAddress:    "0x789",
						ConsensusAddress: "0xabc",
						OperatorID:       "op2",
						Version:          "1.1.0",
						PeerID:           "peer2",
						LastCheckedIn:    time.Now(),
						IsImua:           true,
					},
				}
				mockDB.On("GetVerifiedKeepers").Return(keepers, nil)
			},
			expectedError: "",
			expectedCount: 2,
		},
		{
			name: "should handle database error",
			setupMocks: func(mockDB *mocks.MockDatabaseManager) {
				mockDB.On("GetVerifiedKeepers").Return(nil, errors.New("database connection failed"))
			},
			expectedError: "failed to load verified keepers from database: database connection failed",
			expectedCount: 0,
		},
		{
			name: "should handle empty keeper list",
			setupMocks: func(mockDB *mocks.MockDatabaseManager) {
				mockDB.On("GetVerifiedKeepers").Return([]types.HealthKeeperInfo{}, nil)
			},
			expectedError: "",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			logger := logging.NewNoOpLogger()
			mockDB := &mocks.MockDatabaseManager{}
			tt.setupMocks(mockDB)

			sm := &StateManager{
				keepers: make(map[string]*types.HealthKeeperInfo),
				logger:  logger,
				db:      interfaces.DatabaseManagerInterface(mockDB),
			}

			// Execute
			err := sm.LoadVerifiedKeepers()

			// Verify
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Len(t, sm.keepers, tt.expectedCount)

				// Verify all loaded keepers are initially inactive
				for _, keeper := range sm.keepers {
					assert.False(t, keeper.IsActive)
				}
			}

			// Verify mock expectations
			mockDB.AssertExpectations(t)
		})
	}
}

func TestDumpState(t *testing.T) {
	tests := []struct {
		name             string
		initialKeepers   map[string]*types.HealthKeeperInfo
		setupMocks       func(*mocks.MockDatabaseManager, map[string]*types.HealthKeeperInfo)
		expectedError    string
		expectedLogCalls int
	}{
		{
			name: "should dump active keepers successfully",
			initialKeepers: map[string]*types.HealthKeeperInfo{
				"0x123": {
					KeeperName:       "keeper1",
					KeeperAddress:    "0x123",
					ConsensusAddress: "0x456",
					IsActive:         true,
					LastCheckedIn:    time.Now(),
				},
				"0x789": {
					KeeperName:       "keeper2",
					KeeperAddress:    "0x789",
					ConsensusAddress: "0xabc",
					IsActive:         false, // This should be skipped
					LastCheckedIn:    time.Now(),
				},
			},
			setupMocks: func(mockDB *mocks.MockDatabaseManager, keepers map[string]*types.HealthKeeperInfo) {
				// Only expect call for active keeper
				mockDB.On("UpdateKeeperHealth", mock.MatchedBy(func(health types.KeeperHealthCheckIn) bool {
					return health.KeeperAddress == "0x123"
				}), false).Return(nil)
			},
			expectedError: "",
		},
		{
			name: "should handle database update error",
			initialKeepers: map[string]*types.HealthKeeperInfo{
				"0x123": {
					KeeperName:       "keeper1",
					KeeperAddress:    "0x123",
					ConsensusAddress: "0x456",
					IsActive:         true,
					LastCheckedIn:    time.Now(),
				},
			},
			setupMocks: func(mockDB *mocks.MockDatabaseManager, keepers map[string]*types.HealthKeeperInfo) {
				mockDB.On("UpdateKeeperHealth", mock.Anything, false).Return(errors.New("database error"))
			},
			expectedError: "",
		},
		{
			name:           "should handle empty keeper map",
			initialKeepers: map[string]*types.HealthKeeperInfo{},
			setupMocks: func(mockDB *mocks.MockDatabaseManager, keepers map[string]*types.HealthKeeperInfo) {
				// No database calls expected for empty map
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			logger := logging.NewNoOpLogger()
			mockDB := &mocks.MockDatabaseManager{}
			tt.setupMocks(mockDB, tt.initialKeepers)

			sm := &StateManager{
				keepers: tt.initialKeepers,
				logger:  logger,
				db:      interfaces.DatabaseManagerInterface(mockDB),
			}

			// Execute
			err := sm.DumpState()

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

func TestRetryWithBackoff(t *testing.T) {
	tests := []struct {
		name           string
		operation      func() error
		maxRetries     int
		expectedError  string
		expectedCalls  int
	}{
		{
			name: "should succeed on first try",
			operation: func() error {
				return nil
			},
			maxRetries:    3,
			expectedError: "",
			expectedCalls: 1,
		},
		{
			name: "should succeed after retries",
			operation: func() func() error {
				callCount := 0
				return func() error {
					callCount++
					if callCount < 3 {
						return errors.New("temporary error")
					}
					return nil
				}
			}(),
			maxRetries:    3,
			expectedError: "",
			expectedCalls: 3,
		},
		{
			name: "should fail after max retries",
			operation: func() error {
				return errors.New("persistent error")
			},
			maxRetries:    2,
			expectedError: "operation failed after 2 retries: persistent error",
			expectedCalls: 2,
		},
		{
			name: "should handle zero retries",
			operation: func() error {
				return errors.New("error")
			},
			maxRetries:    0,
			expectedError: "operation failed after 0 retries:",
			expectedCalls: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			logger := logging.NewNoOpLogger()
			sm := &StateManager{
				logger: logger,
			}

			callCount := 0
			wrappedOperation := func() error {
				callCount++
				return tt.operation()
			}

			// Execute
			err := sm.retryWithBackoff(wrappedOperation, tt.maxRetries)

			// Verify
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedCalls, callCount)
		})
	}
}

func TestLoadVerifiedKeepers_Integration(t *testing.T) {
	// This test verifies the integration between LoadVerifiedKeepers and the keeper state
	logger := logging.NewNoOpLogger()
	mockDB := &mocks.MockDatabaseManager{}

	testKeepers := []types.HealthKeeperInfo{
		{
			KeeperName:       "test-keeper-1",
			KeeperAddress:    "0x1234567890abcdef",
			ConsensusAddress: "0xabcdef1234567890",
			OperatorID:       "operator-1",
			Version:          "1.0.0",
			PeerID:           "peer-12345",
			LastCheckedIn:    time.Now().Add(-1 * time.Hour),
			IsImua:           false,
		},
		{
			KeeperName:       "test-keeper-2",
			KeeperAddress:    "0xfedcba0987654321",
			ConsensusAddress: "0x0987654321fedcba",
			OperatorID:       "operator-2",
			Version:          "1.1.0",
			PeerID:           "peer-67890",
			LastCheckedIn:    time.Now().Add(-30 * time.Minute),
			IsImua:           true,
		},
	}

	mockDB.On("GetVerifiedKeepers").Return(testKeepers, nil)

	sm := &StateManager{
		keepers: make(map[string]*types.HealthKeeperInfo),
		logger:  logger,
		db:      interfaces.DatabaseManagerInterface(mockDB),
	}

	// Execute
	err := sm.LoadVerifiedKeepers()

	// Verify
	assert.NoError(t, err)
	assert.Len(t, sm.keepers, 2)

	// Verify keeper 1
	keeper1 := sm.keepers["0x1234567890abcdef"]
	assert.NotNil(t, keeper1)
	assert.Equal(t, "test-keeper-1", keeper1.KeeperName)
	assert.Equal(t, "0x1234567890abcdef", keeper1.KeeperAddress)
	assert.Equal(t, "0xabcdef1234567890", keeper1.ConsensusAddress)
	assert.Equal(t, "operator-1", keeper1.OperatorID)
	assert.Equal(t, "1.0.0", keeper1.Version)
	assert.Equal(t, "peer-12345", keeper1.PeerID)
	assert.False(t, keeper1.IsActive) // Should be initially inactive
	assert.False(t, keeper1.IsImua)

	// Verify keeper 2
	keeper2 := sm.keepers["0xfedcba0987654321"]
	assert.NotNil(t, keeper2)
	assert.Equal(t, "test-keeper-2", keeper2.KeeperName)
	assert.Equal(t, "0xfedcba0987654321", keeper2.KeeperAddress)
	assert.Equal(t, "0x0987654321fedcba", keeper2.ConsensusAddress)
	assert.Equal(t, "operator-2", keeper2.OperatorID)
	assert.Equal(t, "1.1.0", keeper2.Version)
	assert.Equal(t, "peer-67890", keeper2.PeerID)
	assert.False(t, keeper2.IsActive) // Should be initially inactive
	assert.True(t, keeper2.IsImua)

	// Verify mock expectations
	mockDB.AssertExpectations(t)
}

func TestDumpState_Integration(t *testing.T) {
	// This test verifies the integration between DumpState and database updates
	logger := logging.NewNoOpLogger()
	mockDB := &mocks.MockDatabaseManager{}

	// Setup initial state with mixed active/inactive keepers
	initialKeepers := map[string]*types.HealthKeeperInfo{
		"0x123": {
			KeeperName:       "active-keeper-1",
			KeeperAddress:    "0x123",
			ConsensusAddress: "0x456",
			IsActive:         true,
			LastCheckedIn:    time.Now(),
		},
		"0x789": {
			KeeperName:       "active-keeper-2",
			KeeperAddress:    "0x789",
			ConsensusAddress: "0xabc",
			IsActive:         true,
			LastCheckedIn:    time.Now(),
		},
		"0xdef": {
			KeeperName:       "inactive-keeper",
			KeeperAddress:    "0xdef",
			ConsensusAddress: "0x012",
			IsActive:         false,
			LastCheckedIn:    time.Now(),
		},
	}

	// Expect database calls only for active keepers
	mockDB.On("UpdateKeeperHealth", mock.MatchedBy(func(health types.KeeperHealthCheckIn) bool {
		return health.KeeperAddress == "0x123"
	}), false).Return(nil)

	mockDB.On("UpdateKeeperHealth", mock.MatchedBy(func(health types.KeeperHealthCheckIn) bool {
		return health.KeeperAddress == "0x789"
	}), false).Return(nil)

	// No call expected for inactive keeper 0xdef

	sm := &StateManager{
		keepers: initialKeepers,
		logger:  logger,
		db:      interfaces.DatabaseManagerInterface(mockDB),
	}

	// Execute
	err := sm.DumpState()

	// Verify
	assert.NoError(t, err)

	// Verify mock expectations
	mockDB.AssertExpectations(t)

	// Verify that all calls were made exactly twice (for two active keepers)
	mockDB.AssertNumberOfCalls(t, "UpdateKeeperHealth", 2)
}
