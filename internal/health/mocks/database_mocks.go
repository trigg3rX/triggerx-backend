package mocks

import (
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// MockDatabaseManager is a mock implementation of DatabaseManager
type MockDatabaseManager struct {
	mock.Mock
}

// UpdateKeeperHealth mocks the UpdateKeeperHealth method
func (m *MockDatabaseManager) UpdateKeeperHealth(keeperHealth types.KeeperHealthCheckIn, isActive bool) error {
	args := m.Called(keeperHealth, isActive)
	return args.Error(0)
}

// GetVerifiedKeepers mocks the GetVerifiedKeepers method
func (m *MockDatabaseManager) GetVerifiedKeepers() ([]types.HealthKeeperInfo, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]types.HealthKeeperInfo), args.Error(1)
}

// Simplified test helpers for creating test data
func CreateTestKeepers(count int) []types.HealthKeeperInfo {
	keepers := make([]types.HealthKeeperInfo, count)
	for i := 0; i < count; i++ {
		keepers[i] = types.HealthKeeperInfo{
			KeeperName:       "test-keeper-" + string(rune(i)),
			KeeperAddress:    "0x" + string(rune(65+i)) + "23",
			ConsensusAddress: "0x" + string(rune(65+i)) + "45",
			OperatorID:       "op-" + string(rune(i)),
			Version:          "1.0.0",
			PeerID:           "peer-" + string(rune(i)),
			IsActive:         i%2 == 0,
			LastCheckedIn:    time.Now().Add(-time.Duration(i) * time.Hour),
			IsImua:           i%3 == 0,
		}
	}
	return keepers
}
