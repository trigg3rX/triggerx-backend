package notification

import (
	"context"

	"github.com/go-gomail/gomail"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/trigg3rX/triggerx-backend/internal/health/config"
	"github.com/trigg3rX/triggerx-backend/internal/health/interfaces"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type NotificationBot struct {
	tgBotAPI  *tgbotapi.BotAPI
	dbManager interfaces.DatabaseManagerInterface
	logger    logging.Logger
}

func NewBot(
	token string,
	logger logging.Logger,
	dbManager interfaces.DatabaseManagerInterface,
) (*NotificationBot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	return &NotificationBot{
		tgBotAPI:  bot,
		dbManager: dbManager,
		logger:    logger,
	}, nil
}

func (b *NotificationBot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.tgBotAPI.GetUpdatesChan(u)

	go func() {
		for update := range updates {
			if update.Message == nil {
				continue
			}

			if update.Message.IsCommand() && update.Message.Command() == "start" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Enter Your Operator address (Keeper address)")
				_, err := b.tgBotAPI.Send(msg)
				if err != nil {
					b.logger.Errorf("Failed to send message: %v", err)
				}
				continue
			}

			chatID := update.Message.Chat.ID
			keeperAddress := update.Message.Text

			err := b.dbManager.UpdateKeeperChatID(context.Background(), keeperAddress, chatID)
			if err != nil {
				b.logger.Errorf("Failed to update keeper chat ID: %v", err)
			}

			msg := tgbotapi.NewMessage(chatID, "Thanks! You will get the latest notifications")
			_, err = b.tgBotAPI.Send(msg)
			if err != nil {
				b.logger.Errorf("Failed to send message: %v", err)
			}

			err = b.SendTGMessage(chatID, "This is a test message to confirm your chat ID works!")
			if err != nil {
				b.logger.Errorf("Failed to send message: %v", err)
			}
		}
	}()
}

func (b *NotificationBot) SendTGMessage(chatID int64, message string) error {
	msg := tgbotapi.NewMessage(chatID, message)
	_, err := b.tgBotAPI.Send(msg)
	return err
}

func (b *NotificationBot) SendEmailMessage(to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", config.GetEmailUser())
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer("smtp.zeptomail.in", 587, config.GetEmailUser(), config.GetEmailPassword())
	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}

func (b *NotificationBot) Stop() {
	b.tgBotAPI.StopReceivingUpdates()
}