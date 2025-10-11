package mocks

import (
	"github.com/stretchr/testify/mock"
)

// MockNotificationBot is a mock implementation of NotificationBotInterface
type MockNotificationBot struct {
	mock.Mock
}

// Start mocks the Start method
func (m *MockNotificationBot) Start() {
	m.Called()
}

// SendTGMessage mocks the SendTGMessage method
func (m *MockNotificationBot) SendTGMessage(chatID int64, message string) error {
	args := m.Called(chatID, message)
	return args.Error(0)
}

// SendEmailMessage mocks the SendEmailMessage method
func (m *MockNotificationBot) SendEmailMessage(to, subject, body string) error {
	args := m.Called(to, subject, body)
	return args.Error(0)
}

// Stop mocks the Stop method
func (m *MockNotificationBot) Stop() {
	m.Called()
}

// MockTelegramBot is deprecated, use MockNotificationBot instead
// Kept for backward compatibility
type MockTelegramBot = MockNotificationBot
