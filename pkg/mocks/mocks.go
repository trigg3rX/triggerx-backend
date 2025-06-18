package mocks

import (
	"github.com/gocql/gocql"
	"github.com/trigg3rX/triggerx-backend/internal/health/client"
	"github.com/trigg3rX/triggerx-backend/internal/health/telegram"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// MockDatabaseManager is a mock implementation of the database manager
type MockDatabaseManager struct{}

// NewMockDatabaseManager creates a new mock database manager
func NewMockDatabaseManager() *MockDatabaseManager {
	return &MockDatabaseManager{}
}

// MockTelegramBot is a mock implementation of the telegram bot
type MockTelegramBot struct{}

// NewMockTelegramBot creates a new mock telegram bot
func NewMockTelegramBot() *telegram.Bot {
	return &telegram.Bot{}
}

// MockDatabaseConnection is a mock implementation of the database connection
type MockDatabaseConnection struct{}

// NewMockDatabaseConnection creates a new mock database connection
func NewMockDatabaseConnection() *database.Connection {
	return &database.Connection{}
}

// InitializeTestDependencies initializes all required dependencies for testing
func InitializeTestDependencies(logger logging.Logger) {
	// Initialize mock database connection
	mockDBConn := NewMockDatabaseConnection()

	// Initialize mock telegram bot
	mockTelegramBot := NewMockTelegramBot()

	// Initialize database manager with mocks
	client.InitDatabaseManager(logger, mockDBConn, mockTelegramBot)
}

// Mock methods for database manager
func (m *MockDatabaseManager) GetInstance() *client.DatabaseManager {
	return &client.DatabaseManager{}
}

// Mock methods for telegram bot
func (m *MockTelegramBot) SendMessage(chatID int64, message string) error {
	return nil
}

// Mock methods for database connection
func (m *MockDatabaseConnection) Close() {
	// Mock implementation
}

func (m *MockDatabaseConnection) Session() database.Sessioner {
	return &MockSession{}
}

// MockSession is a mock implementation of database.Sessioner
type MockSession struct{}

func (m *MockSession) Query(query string, args ...interface{}) *gocql.Query {
	return &gocql.Query{}
}

func (m *MockSession) ExecuteBatch(batch *gocql.Batch) error {
	return nil
}

func (m *MockSession) Close() {
	// Mock implementation
}
