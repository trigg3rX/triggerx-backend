package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-gomail/gomail"
	"github.com/trigg3rX/triggerx-backend/internal/health/config"
	"github.com/trigg3rX/triggerx-backend/internal/health/telegram"

	"github.com/trigg3rX/triggerx-backend/pkg/datastore/interfaces"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// DatabaseManager handles database operations
type DatabaseManager struct {
	logger      logging.Logger
	telegramBot *telegram.Bot
	keeperRepo  interfaces.GenericRepository[types.KeeperDataEntity]
}

var instance *DatabaseManager

// InitDatabaseManager initializes the database manager with a logger
func InitDatabaseManager(logger logging.Logger, keeperRepo interfaces.GenericRepository[types.KeeperDataEntity], telegramBot *telegram.Bot) {
	if logger == nil {
		panic("logger cannot be nil")
	}
	if keeperRepo == nil {
		panic("keeper repository cannot be nil")
	}
	if telegramBot == nil {
		logger.Warn("Telegram bot is nil, notifications will not be sent")
	}

	// Create a new logger with component field and proper level
	dbLogger := logger.With("component", "database")

	instance = &DatabaseManager{
		logger:      dbLogger,
		telegramBot: telegramBot,
		keeperRepo:  keeperRepo,
	}
}

// GetInstance returns the database manager instance
func GetInstance() *DatabaseManager {
	if instance == nil {
		panic("database manager not initialized")
	}
	return instance
}

// KeeperRegistered registers a new keeper or updates an existing one (status = true)
func (dm *DatabaseManager) UpdateKeeperHealth(keeperHealth types.KeeperHealthCheckIn, isActive bool) error {
	dm.logger.Debug("Updating keeper status in database",
		"keeper", keeperHealth.KeeperAddress,
		"active", isActive,
	)

	keeperHealth.KeeperAddress = strings.ToLower(keeperHealth.KeeperAddress)
	keeperHealth.ConsensusAddress = strings.ToLower(keeperHealth.ConsensusAddress)

	if len(keeperHealth.KeeperAddress) > 0 && !bytes.HasPrefix([]byte(keeperHealth.KeeperAddress), []byte("0x")) {
		dm.logger.Debug("Adding 0x prefix to keeper address",
			"keeper", keeperHealth.KeeperAddress,
		)
		keeperHealth.KeeperAddress = "0x" + keeperHealth.KeeperAddress
	}

	ctx := context.Background()

	// Fetch keeper using repository
	keeper, err := dm.keeperRepo.GetByNonID(ctx, "keeper_address", keeperHealth.KeeperAddress)
	if err != nil {
		dm.logger.Error("Failed to retrieve keeper",
			"keeper", keeperHealth.KeeperAddress,
			"error", err,
		)
		return err
	}

	if keeper == nil {
		dm.logger.Errorf("[KeeperHealthCheckIn] No keeper found with address: %s", keeperHealth.KeeperAddress)
		return errors.New("keeper not found")
	}

	keeperID := keeper.KeeperID
	prevOnline := keeper.Online
	prevLastCheckedIn := keeper.LastCheckedIn
	prevUptime := keeper.Uptime

	if keeperHealth.PeerID == "" {
		keeperHealth.PeerID = "no-peer-id"
	}

	dm.logger.Infof("[KeeperHealthCheckIn] Keeper ID: %d | Online: %t", keeperID, isActive)

	// --- UPTIME LOGIC ---
	// If previously online, add to uptime (regardless of new isActive)
	if prevOnline {
		now := time.Now().UTC()
		uptimeToAdd := int64(now.Sub(prevLastCheckedIn).Seconds())
		if uptimeToAdd < 0 {
			uptimeToAdd = 0 // avoid negative values
		}
		keeper.Uptime = prevUptime + uptimeToAdd

		// Update uptime field using repository
		if err := dm.keeperRepo.Update(ctx, keeper); err != nil {
			dm.logger.Error("Failed to update keeper uptime",
				"error", err,
				"keeper_id", keeperID,
				"keeper", keeperHealth.KeeperAddress,
			)
			return err
		}
	}
	// --- END UPTIME LOGIC ---

	if !isActive {
		// If not active, just set online = false
		keeper.Online = false
		if err := dm.keeperRepo.Update(ctx, keeper); err != nil {
			dm.logger.Error("Failed to update keeper inactive status",
				"error", err,
				"keeper_id", keeperID,
				"keeper", keeperHealth.KeeperAddress,
			)
			return err
		}
		return nil
	}

	// If active, update all fields including last_checked_in
	keeper.ConsensusAddress = keeperHealth.ConsensusAddress
	keeper.Online = true
	keeper.PeerID = keeperHealth.PeerID
	keeper.Version = keeperHealth.Version
	keeper.LastCheckedIn = keeperHealth.Timestamp

	if err := dm.keeperRepo.Update(ctx, keeper); err != nil {
		dm.logger.Error("Failed to update keeper status",
			"error", err,
			"keeper_id", keeperID,
		)
		return err
	}

	if !isActive {
		go dm.checkAndNotifyOfflineKeeper(keeperID)
	}

	dm.logger.Info("Successfully updated keeper status",
		"keeper_id", keeperID,
		"active", isActive,
	)
	return nil
}

func (dm *DatabaseManager) checkAndNotifyOfflineKeeper(keeperID int64) {
	time.Sleep(10 * time.Minute)

	dm.logger.Debug("Checking current status for offline keeper",
		"keeper_id", keeperID,
	)

	ctx := context.Background()
	keeper, err := dm.keeperRepo.GetByID(ctx, keeperID)
	if err != nil {
		dm.logger.Error("Failed to check keeper online status",
			"error", err,
			"keeper_id", keeperID,
		)
		return
	}

	if keeper == nil {
		dm.logger.Error("Keeper not found",
			"keeper_id", keeperID,
		)
		return
	}

	online := keeper.Online

	if !online {
		chatID := keeper.ChatID
		keeperName := keeper.KeeperName
		emailID := keeper.EmailID

		if chatID != 0 {
			telegramMsg := fmt.Sprintf("Keeper %s is down for more than 10 minutes. Please check and start it.", keeperName)
			if err := dm.telegramBot.SendMessage(chatID, telegramMsg); err != nil {
				dm.logger.Error("Failed to send Telegram notification",
					"error", err,
					"keeper", keeperName,
					"keeper_id", keeperID,
				)
			}
		} else {
			dm.logger.Warn("No Telegram chat ID found",
				"keeper", keeperName,
				"keeper_id", keeperID,
			)
		}

		if emailID != "" {
			subject := fmt.Sprintf("TriggerX Keeper Down Alert - %s", keeperName)
			emailBody := fmt.Sprintf(`
				<h2>Keeper Update</h2>
				<p>This is a critical information from TriggerX. Your keeper <strong>%s</strong> has been down for more than 10 minutes. Please take action immediately.</p>
				<p>Regards,<br>TriggerX Team</p>
			`, keeperName)

			if err := dm.sendEmailNotification(emailID, subject, emailBody); err != nil {
				dm.logger.Error("Failed to send email notification",
					"error", err,
					"keeper", keeperName,
					"keeper_id", keeperID,
				)
			}
		} else {
			dm.logger.Warn("No email address found",
				"keeper", keeperName,
				"keeper_id", keeperID,
			)
		}

		dm.logger.Info("Completed notification process for offline keeper",
			"keeper", keeperName,
			"keeper_id", keeperID,
		)
	} else {
		dm.logger.Info("Keeper is back online",
			"keeper_id", keeperID,
		)
	}
}

// func (dm *DatabaseManager) sendTelegramNotification(chatID int64, message string) error {
// 	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", config.GetBotToken())
// 	payload := map[string]interface{}{
// 		"chat_id": chatID,
// 		"text":    message,
// 	}

// 	jsonData, err := json.Marshal(payload)
// 	if err != nil {
// 		dm.logger.Errorf("[Notification] Failed to marshal Telegram payload: %v", err)
// 		return err
// 	}

// 	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
// 	if err != nil {
// 		dm.logger.Errorf("[Notification] Failed to send Telegram message: %v", err)
// 		return err
// 	}
// 	defer resp.Body.Close()

// 	dm.logger.Infof("[Notification] Telegram message sent successfully to chat ID: %d (Status: %d)", chatID, resp.StatusCode)
// 	return nil
// }

func (dm *DatabaseManager) sendEmailNotification(to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", config.GetEmailUser())
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer("smtp.zeptomail.in", 587, config.GetEmailUser(), config.GetEmailPassword())
	if err := d.DialAndSend(m); err != nil {
		dm.logger.Errorf("[Notification] Failed to send email to %s: %v", to, err)
		return err
	}

	dm.logger.Infof("[Notification] Email sent successfully to: %s", to)
	return nil
}

// GetVerifiedKeepers retrieves only verified keepers from the database
func (dm *DatabaseManager) GetVerifiedKeepers() ([]types.HealthKeeperInfo, error) {
	ctx := context.Background()

	// Get all keepers and filter for verified ones
	allKeepers, err := dm.keeperRepo.GetByFields(ctx, map[string]interface{}{
		"registered":  true,
		"whitelisted": true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get verified keepers: %w", err)
	}

	var keepers []types.HealthKeeperInfo
	for _, keeper := range allKeepers {
		keepers = append(keepers, types.HealthKeeperInfo{
			KeeperName:       keeper.KeeperName,
			KeeperAddress:    keeper.KeeperAddress,
			ConsensusAddress: keeper.ConsensusAddress,
			OperatorID:       fmt.Sprintf("%d", keeper.OperatorID),
			Version:          keeper.Version,
			PeerID:           keeper.PeerID,
			LastCheckedIn:    keeper.LastCheckedIn,
			IsImua:           keeper.OnImua,
		})
	}

	dm.logger.Debug("Retrieved verified keepers from database",
		"count", len(keepers),
	)
	return keepers, nil
}
