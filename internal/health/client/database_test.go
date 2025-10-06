package client

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/trigg3rX/triggerx-backend/internal/health/mocks"
	"github.com/trigg3rX/triggerx-backend/internal/health/telegram"
	"github.com/trigg3rX/triggerx-backend/internal/health/types"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	commonTypes "github.com/trigg3rX/triggerx-backend/pkg/types"
)

// TestDatabaseManagerInterface tests the interface without relying on concrete implementations
func TestDatabaseManagerInterface(t *testing.T) {
	t.Run("should create database manager with mock dependencies", func(t *testing.T) {
		// Reset global instance
		instance = nil
		
		// Create mock dependencies
		logger := logging.NewNoOpLogger()
		
		// Create minimal concrete instances for the test
		conn := &database.Connection{}
		bot := &telegram.Bot{}
		
		// This test verifies the initialization doesn't panic with valid inputs
		assert.NotPanics(t, func() {
			InitDatabaseManager(logger, conn, bot)
		})
		
		// Verify instance was created
		dbManager := GetInstance()
		assert.NotNil(t, dbManager)
	})
	
	t.Run("should panic with nil logger", func(t *testing.T) {
		// Reset global instance
		instance = nil
		
		conn := &database.Connection{}
		bot := &telegram.Bot{}
		
		assert.Panics(t, func() {
			InitDatabaseManager(nil, conn, bot)
		})
	})
	
	t.Run("should panic with nil connection", func(t *testing.T) {
		// Reset global instance
		instance = nil
		
		logger := logging.NewNoOpLogger()
		bot := &telegram.Bot{}
		
		assert.Panics(t, func() {
			InitDatabaseManager(logger, nil, bot)
		})
	})
}

// TestDatabaseManagerMocked tests using our mock for the business logic
func TestDatabaseManagerMocked(t *testing.T) {
	t.Run("should use mock database manager for business logic tests", func(t *testing.T) {
		mockDB := &mocks.MockDatabaseManager{}
		
		// Test GetVerifiedKeepers
		expectedKeepers := []types.KeeperInfo{
			{
				KeeperName:    "test-keeper",
				KeeperAddress: "0x123",
				IsActive:      true,
				LastCheckedIn: time.Now(),
			},
		}
		
		mockDB.On("GetVerifiedKeepers").Return(expectedKeepers, nil)
		
		result, err := mockDB.GetVerifiedKeepers()
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "test-keeper", result[0].KeeperName)
		
		mockDB.AssertExpectations(t)
	})

	t.Run("should handle UpdateKeeperHealth", func(t *testing.T) {
		mockDB := &mocks.MockDatabaseManager{}
		
		health := commonTypes.KeeperHealthCheckIn{
			KeeperAddress: "0x123",
			Version:       "1.0.0",
			Timestamp:     time.Now(),
		}
		
		mockDB.On("UpdateKeeperHealth", health, true).Return(nil)
		
		err := mockDB.UpdateKeeperHealth(health, true)
		assert.NoError(t, err)
		
		mockDB.AssertExpectations(t)
	})

	t.Run("should handle database errors", func(t *testing.T) {
		mockDB := &mocks.MockDatabaseManager{}
		
		mockDB.On("GetVerifiedKeepers").Return(nil, errors.New("database error"))
		
		result, err := mockDB.GetVerifiedKeepers()
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "database error")
		
		mockDB.AssertExpectations(t)
	})
}
