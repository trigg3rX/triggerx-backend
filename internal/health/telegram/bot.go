package telegram

import (
	"strconv"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gocql/gocql"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type Bot struct {
	api    *tgbotapi.BotAPI
	logger logging.Logger
	db     *database.Connection
}

func NewBot(token string, logger logging.Logger, db *database.Connection) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	return &Bot{
		api:    bot,
		logger: logger,
		db:     db,
	}, nil
}

func (b *Bot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for update := range updates {
			if update.Message == nil {
				continue
			}

			if update.Message.IsCommand() && update.Message.Command() == "start" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Enter Your Operator address (Keeper address)")
				b.api.Send(msg)
				continue
			}

			chatID := update.Message.Chat.ID
			keeperAddress := update.Message.Text

			err := b.updateKeeperChatID(keeperAddress, chatID)
			if err != nil {
				b.logger.Errorf("Failed to update keeper chat ID: %v", err)
				msg := tgbotapi.NewMessage(chatID, "Failed to register your keeper name. Please try again.")
				b.api.Send(msg)
				continue
			}

			msg := tgbotapi.NewMessage(chatID, "Thanks! You will get the latest notifications")
			b.api.Send(msg)

			testMsg := tgbotapi.NewMessage(chatID, "This is a test message to confirm your chat ID works!")
			b.api.Send(testMsg)
		}
	}()

	wg.Wait()
}

func (b *Bot) updateKeeperChatID(keeperAddress string, chatID int64) error {
	b.logger.Infof("[UpdateKeeperChatID] Finding keeper ID for keeper: %s", keeperAddress)

	var keeperID string
	if err := b.db.Session().Query(`
		SELECT keeper_id FROM triggerx.keeper_data 
		WHERE keeper_address = ? ALLOW FILTERING`, keeperAddress).Consistency(gocql.One).Scan(&keeperID); err != nil {
		b.logger.Errorf("[UpdateKeeperChatID] Error finding keeper ID for keeper %s: %v", keeperAddress, err)
		return err
	}

	b.logger.Infof("[UpdateKeeperChatID] Updating chat ID for keeper ID: %s", keeperID)

	chatIDStr := strconv.FormatInt(chatID, 10)

	if err := b.db.Session().Query(`
		UPDATE triggerx.keeper_data 
		SET chat_id = ? 
		WHERE keeper_id = ?`,
		chatIDStr, keeperID).Exec(); err != nil {
		b.logger.Errorf("[UpdateKeeperChatID] Error updating chat ID for keeper ID %s: %v", keeperID, err)
		return err
	}

	b.logger.Infof("[UpdateKeeperChatID] Successfully updated chat ID for keeper: %s", keeperAddress)
	return nil
}

func (b *Bot) SendMessage(chatID int64, message string) error {
	msg := tgbotapi.NewMessage(chatID, message)
	_, err := b.api.Send(msg)
	return err
}
