package interfaces

type NotificationBotInterface interface {
	Start()
	SendTGMessage(chatID int64, message string) error
	SendEmailMessage(to, subject, body string) error
	Stop()
}