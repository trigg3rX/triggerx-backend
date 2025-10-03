package telegram

import (
	"context"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/trigg3rX/triggerx-backend/pkg/datastore/interfaces"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type Bot struct {
	api        *tgbotapi.BotAPI
	logger     logging.Logger
	keeperRepo interfaces.GenericRepository[types.KeeperDataEntity]
}

func NewBot(token string, logger logging.Logger, keeperRepo interfaces.GenericRepository[types.KeeperDataEntity]) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	return &Bot{
		api:        bot,
		logger:     logger,
		keeperRepo: keeperRepo,
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
				_, err := b.api.Send(msg)
				if err != nil {
					b.logger.Errorf("Failed to send message: %v", err)
				}
				continue
			}

			chatID := update.Message.Chat.ID
			keeperAddress := update.Message.Text

			err := b.updateKeeperChatID(keeperAddress, chatID)
			if err != nil {
				b.logger.Errorf("Failed to update keeper chat ID: %v", err)
				msg := tgbotapi.NewMessage(chatID, "Failed to register your keeper name. Please try again.")
				_, err = b.api.Send(msg)
				if err != nil {
					b.logger.Errorf("Failed to send message: %v", err)
				}
				continue
			}

			msg := tgbotapi.NewMessage(chatID, "Thanks! You will get the latest notifications")
			_, err = b.api.Send(msg)
			if err != nil {
				b.logger.Errorf("Failed to send message: %v", err)
			}

			testMsg := tgbotapi.NewMessage(chatID, "This is a test message to confirm your chat ID works!")
			_, err = b.api.Send(testMsg)
			if err != nil {
				b.logger.Errorf("Failed to send message: %v", err)
			}
		}
	}()

	wg.Wait()
}

func (b *Bot) updateKeeperChatID(keeperAddress string, chatID int64) error {
	b.logger.Infof("[UpdateKeeperChatID] Finding keeper ID for keeper: %s", keeperAddress)

	ctx := context.Background()
	keeper, err := b.keeperRepo.GetByNonID(ctx, "keeper_address", keeperAddress)
	if err != nil {
		b.logger.Errorf("[UpdateKeeperChatID] Error finding keeper ID for keeper %s: %v", keeperAddress, err)
		return err
	}

	if keeper == nil {
		b.logger.Errorf("[UpdateKeeperChatID] Keeper not found: %s", keeperAddress)
		return err
	}

	b.logger.Infof("[UpdateKeeperChatID] Updating chat ID for keeper ID: %d", keeper.KeeperID)

	keeper.ChatID = chatID
	if err := b.keeperRepo.Update(ctx, keeper); err != nil {
		b.logger.Errorf("[UpdateKeeperChatID] Error updating chat ID for keeper ID %d: %v", keeper.KeeperID, err)
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
