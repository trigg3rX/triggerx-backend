package mocks

import (
	"github.com/stretchr/testify/mock"
)

// MockTelegramBot is a mock implementation of the Telegram bot
type MockTelegramBot struct {
	mock.Mock
}

// SendMessage mocks the SendMessage method
func (m *MockTelegramBot) SendMessage(chatID int64, message string) error {
	args := m.Called(chatID, message)
	return args.Error(0)
}

// Start mocks the Start method
func (m *MockTelegramBot) Start() {
	m.Called()
}
