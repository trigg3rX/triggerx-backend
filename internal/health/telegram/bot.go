package telegram

import (
	"context"
	"strconv"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gocql/gocql"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

type Bot struct {
	api    *tgbotapi.BotAPI
	logger observability.Logger
	db     *database.Connection
}

func NewBot(token string, logger observability.Logger, db *database.Connection) (*Bot, error) {
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

func (b *Bot) Start(ctx context.Context) {
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
				_, err := b.api.Send(msg)
				if err != nil {
					b.logger.Error(ctx, "Failed to send message", observability.Error(err))
				}
				continue
			}

			chatID := update.Message.Chat.ID
			keeperAddress := update.Message.Text

			err := b.updateKeeperChatID(ctx, keeperAddress, chatID)
			if err != nil {
				b.logger.Error(ctx, "Failed to update keeper chat ID", observability.Error(err))
				msg := tgbotapi.NewMessage(chatID, "Failed to register your keeper name. Please try again.")
				_, err = b.api.Send(msg)
				if err != nil {
					b.logger.Error(ctx, "Failed to send message", observability.Error(err))
				}
				continue
			}

			msg := tgbotapi.NewMessage(chatID, "Thanks! You will get the latest notifications")
			_, err = b.api.Send(msg)
			if err != nil {
				b.logger.Error(ctx, "Failed to send message", observability.Error(err))
			}

			testMsg := tgbotapi.NewMessage(chatID, "This is a test message to confirm your chat ID works!")
			_, err = b.api.Send(testMsg)
			if err != nil {
				b.logger.Error(ctx, "Failed to send message", observability.Error(err))
			}
		}
	}()

	wg.Wait()
}

func (b *Bot) updateKeeperChatID(ctx context.Context, keeperAddress string, chatID int64) error {
	b.logger.Debug(ctx, "Finding keeper ID for keeper", observability.String("keeper", keeperAddress))

	var keeperID string
	if err := b.db.Session().Query(`
		SELECT keeper_id FROM triggerx.keeper_data 
		WHERE keeper_address = ? ALLOW FILTERING`, keeperAddress).Consistency(gocql.One).Scan(&keeperID); err != nil {
		b.logger.Error(ctx, "Error finding keeper ID for keeper", observability.String("keeper", keeperAddress), observability.Error(err))
		return err
	}

	b.logger.Debug(ctx, "Updating chat ID for keeper ID", observability.String("keeper_id", keeperID))

	chatIDStr := strconv.FormatInt(chatID, 10)

	if err := b.db.Session().Query(`
		UPDATE triggerx.keeper_data 
		SET chat_id = ? 
		WHERE keeper_id = ?`,
		chatIDStr, keeperID).Exec(); err != nil {
		b.logger.Error(ctx, "Error updating chat ID for keeper ID", observability.String("keeper_id", keeperID), observability.Error(err))
		return err
	}

	b.logger.Debug(ctx, "Successfully updated chat ID for keeper", observability.String("keeper", keeperAddress))
	return nil
}

func (b *Bot) SendMessage(chatID int64, message string) error {
	msg := tgbotapi.NewMessage(chatID, message)
	_, err := b.api.Send(msg)
	return err
}
