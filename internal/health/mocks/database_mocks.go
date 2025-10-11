package mocks

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// MockDatabaseManager is a mock implementation of DatabaseManagerInterface
type MockDatabaseManager struct {
	mock.Mock
}

// UpdateKeeperStatus mocks the UpdateKeeperStatus method
func (m *MockDatabaseManager) UpdateKeeperStatus(
	ctx context.Context,
	keeperAddress string,
	consensusAddress string,
	version string,
	uptime int64,
	timestamp time.Time,
	publicIP string,
	isActive bool,
) error {
	args := m.Called(ctx, keeperAddress, consensusAddress, version, uptime, timestamp, publicIP, isActive)
	return args.Error(0)
}

// GetVerifiedKeepers mocks the GetVerifiedKeepers method
func (m *MockDatabaseManager) GetVerifiedKeepers(ctx context.Context) ([]types.HealthKeeperInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]types.HealthKeeperInfo), args.Error(1)
}

// UpdateAllKeepersStatus mocks the UpdateAllKeepersStatus method
func (m *MockDatabaseManager) UpdateAllKeepersStatus(ctx context.Context, onlineKeepers []types.HealthKeeperInfo) error {
	args := m.Called(ctx, onlineKeepers)
	return args.Error(0)
}

// UpdateKeeperChatID mocks the UpdateKeeperChatID method
func (m *MockDatabaseManager) UpdateKeeperChatID(ctx context.Context, keeperAddress string, chatID int64) error {
	args := m.Called(ctx, keeperAddress, chatID)
	return args.Error(0)
}

// GetKeeperChatInfo mocks the GetKeeperChatInfo method
func (m *MockDatabaseManager) GetKeeperChatInfo(ctx context.Context, keeperAddress string) (int64, string, error) {
	args := m.Called(ctx, keeperAddress)
	return args.Get(0).(int64), args.Get(1).(string), args.Error(2)
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
			Uptime:           0,
			IsActive:         i%2 == 0,
			LastCheckedIn:    time.Now().Add(-time.Duration(i) * time.Hour),
			IsImua:           i%3 == 0,
		}
	}
	return keepers
}
