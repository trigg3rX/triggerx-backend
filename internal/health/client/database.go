package client

import (
	"bytes"
	// "encoding/json"
	"errors"
	"fmt"
	// "net/http"
	"strings"
	"time"

	"github.com/go-gomail/gomail"
	"github.com/trigg3rX/triggerx-backend/internal/health/config"
	"github.com/trigg3rX/triggerx-backend/internal/health/telegram"

	"github.com/trigg3rX/triggerx-backend/internal/health/types"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	commonTypes "github.com/trigg3rX/triggerx-backend/pkg/types"
)

// DatabaseManager handles database operations
type DatabaseManager struct {
	logger      logging.Logger
	db          *database.Connection
	telegramBot *telegram.Bot
}

var instance *DatabaseManager

// InitDatabaseManager initializes the database manager with a logger
func InitDatabaseManager(logger logging.Logger, connection *database.Connection, telegramBot *telegram.Bot) {
	if logger == nil {
		panic("logger cannot be nil")
	}
	if connection == nil {
		panic("database connection cannot be nil")
	}
	if telegramBot == nil {
		logger.Warn("Telegram bot is nil, notifications will not be sent")
	}

	// Create a new logger with component field and proper level
	dbLogger := logger.With("component", "database")

	instance = &DatabaseManager{
		logger:      dbLogger,
		db:          connection,
		telegramBot: telegramBot,
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
func (dm *DatabaseManager) UpdateKeeperHealth(keeperHealth commonTypes.KeeperHealthCheckIn, isActive bool) error {
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

	var keeperID int64
	if err := dm.db.Session().Query(`
		SELECT keeper_id FROM triggerx.keeper_data WHERE keeper_address = ? ALLOW FILTERING`,
		keeperHealth.KeeperAddress).Scan(&keeperID); err != nil {
		dm.logger.Error("Failed to retrieve keeper_id",
			"keeper", keeperHealth.KeeperAddress,
			"error", err,
		)
		return err
	}

	if keeperID == 0 {
		dm.logger.Errorf("[KeeperHealthCheckIn] No keeper found with address: %s", keeperHealth.KeeperAddress)
		return errors.New("keeper not found")
	}

	if keeperHealth.PeerID == "" {
		keeperHealth.PeerID = "no-peer-id"
	}

	dm.logger.Infof("[KeeperHealthCheckIn] Keeper ID: %d | Online: %t", keeperID, isActive)

	if !isActive {
		if err := dm.db.Session().Query(`
			UPDATE triggerx.keeper_data 
			SET online = ?
			WHERE keeper_id = ?`,
			false, keeperID).Exec(); err != nil {
			dm.logger.Error("Failed to update keeper inactive status",
				"error", err,
				"keeper_id", keeperID,
				"keeper", keeperHealth.KeeperAddress,
			)
			return err
		}
		return nil
	}

	if err := dm.db.Session().Query(`
		UPDATE triggerx.keeper_data 
		SET consensus_address = ?, online = ?, peer_id = ?, version = ?, last_checked_in = ? 
		WHERE keeper_id = ?`,
		keeperHealth.ConsensusAddress, true, keeperHealth.PeerID, keeperHealth.Version, keeperHealth.Timestamp, keeperID).Exec(); err != nil {
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

	var online bool
	err := dm.db.Session().Query(`
		SELECT online FROM triggerx.keeper_data WHERE keeper_id = ?`,
		keeperID).Scan(&online)

	if err != nil {
		dm.logger.Error("Failed to check keeper online status",
			"error", err,
			"keeper_id", keeperID,
		)
		return
	}

	if !online {
		var chatID int64
		var keeperName, emailID string
		err := dm.db.Session().Query(`
			SELECT chat_id, keeper_name, email_id 
			FROM triggerx.keeper_data 
			WHERE keeper_id = ?`,
			keeperID).Scan(&chatID, &keeperName, &emailID)

		if err != nil {
			dm.logger.Error("Failed to fetch keeper communication info",
				"error", err,
				"keeper_id", keeperID,
			)
			return
		}

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

// Public wrapper functions
func UpdateKeeperHealth(keeperHealth commonTypes.KeeperHealthCheckIn, isActive bool) error {
	return GetInstance().UpdateKeeperHealth(keeperHealth, isActive)
}

// GetVerifiedKeepers retrieves only verified keepers from the database
func (dm *DatabaseManager) GetVerifiedKeepers() ([]types.KeeperInfo, error) {
	var keepers []types.KeeperInfo

	iter := dm.db.Session().Query(`
		SELECT keeper_name, keeper_address, consensus_address, operator_id, version, peer_id, last_checked_in 
		FROM triggerx.keeper_data 
		WHERE registered = true AND whitelisted = true 
		ALLOW FILTERING`).Iter()

	var keeperName, keeperAddress, consensusAddress, operatorID, version, peerID string
	var lastCheckedIn time.Time

	for iter.Scan(&keeperName, &keeperAddress, &consensusAddress, &operatorID, &version, &peerID, &lastCheckedIn) {
		keepers = append(keepers, types.KeeperInfo{
			KeeperName:       keeperName,
			KeeperAddress:    keeperAddress,
			ConsensusAddress: consensusAddress,
			OperatorID:       operatorID,
			Version:          version,
			PeerID:           peerID,
			LastCheckedIn:    lastCheckedIn,
		})
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("error closing iterator: %w", err)
	}

	dm.logger.Debug("Retrieved verified keepers from database",
		"count", len(keepers),
	)
	return keepers, nil
}
