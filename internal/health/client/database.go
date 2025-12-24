package client

import (
	"bytes"
	"context"

	// "encoding/json"
	"errors"
	"fmt"

	// "net/http"
	"strings"
	"time"

	"github.com/go-gomail/gomail"
	"github.com/trigg3rX/triggerx-backend/internal/health/config"
	"github.com/trigg3rX/triggerx-backend/internal/health/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/health/telegram"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// DatabaseManager handles database operations
type DatabaseManager struct {
	logger      observability.Logger
	tracer      observability.Tracer
	db          *database.Connection
	telegramBot *telegram.Bot
}

var instance *DatabaseManager

// InitDatabaseManager initializes the database manager with a logger and tracer
func InitDatabaseManager(ctx context.Context, logger observability.Logger, tracer observability.Tracer, connection *database.Connection, telegramBot *telegram.Bot) {
	if logger == nil {
		panic("logger cannot be nil")
	}
	if tracer == nil {
		panic("tracer cannot be nil")
	}
	if connection == nil {
		panic("database connection cannot be nil")
	}
	if telegramBot == nil {
		logger.Warn(ctx, "Telegram bot is nil, notifications will not be sent")
	}

	// Create a new logger with component field and proper level
	dbLogger := logger.With(observability.String("component", "database"))

	instance = &DatabaseManager{
		logger:      dbLogger,
		tracer:      tracer,
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
func (dm *DatabaseManager) UpdateKeeperHealth(ctx context.Context, keeperHealth types.KeeperHealthCheckIn, isActive bool) error {
	// Start a span for the database update operation
	ctx, span := dm.tracer.Start(ctx, "db.update_keeper_health",
		observability.WithSpanKind(trace.SpanKindClient),
		observability.WithAttributes(
			attribute.String("db.system", "cassandra"),
			attribute.String("db.operation", "update"),
			attribute.String("db.collection", "keeper_data"),
			attribute.String("keeper.address", keeperHealth.KeeperAddress),
			attribute.Bool("keeper.active", isActive),
		),
	)
	defer span.End()

	// dm.logger.Debug(ctx, "Updating keeper status in database",
	// 	observability.String("keeper", keeperHealth.KeeperAddress),
	// 	observability.Bool("active", isActive),
	// )

	keeperHealth.KeeperAddress = strings.ToLower(keeperHealth.KeeperAddress)
	keeperHealth.ConsensusAddress = strings.ToLower(keeperHealth.ConsensusAddress)

	if len(keeperHealth.KeeperAddress) > 0 && !bytes.HasPrefix([]byte(keeperHealth.KeeperAddress), []byte("0x")) {
		dm.logger.Debug(ctx, "Adding 0x prefix to keeper address",
			observability.String("keeper", keeperHealth.KeeperAddress),
		)
		keeperHealth.KeeperAddress = "0x" + keeperHealth.KeeperAddress
	}

	var keeperID int64
	var prevOnline bool
	var prevLastCheckedIn time.Time
	var prevUptime int64

	// Fetch previous online status, last_checked_in, and uptime
	selectStart := time.Now()
	_, selectSpan := dm.tracer.Start(ctx, "db.select_keeper_data",
		observability.WithSpanKind(trace.SpanKindClient),
		observability.WithAttributes(
			attribute.String("db.system", "cassandra"),
			attribute.String("db.operation", "select"),
			attribute.String("db.collection", "keeper_data"),
		),
	)
	err := dm.db.Session().Query(`
		SELECT keeper_id, online, last_checked_in, uptime FROM triggerx.keeper_data WHERE keeper_address = ? ALLOW FILTERING`,
		keeperHealth.KeeperAddress).Scan(&keeperID, &prevOnline, &prevLastCheckedIn, &prevUptime)
	selectSpan.End()
	metrics.RecordDBOperationDuration(ctx, "select", time.Since(selectStart))
	if err != nil {
		selectSpan.SetStatus(codes.Error, err.Error())
		dm.logger.Error(ctx, "Failed to retrieve keeper_id and previous status",
			observability.String("keeper", keeperHealth.KeeperAddress),
			observability.Error(err),
		)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	selectSpan.SetStatus(codes.Ok, "")

	if keeperID == 0 {
		dm.logger.Error(ctx, "No keeper found with address",
			observability.String("keeper", keeperHealth.KeeperAddress),
		)
		return errors.New("keeper not found")
	}

	if keeperHealth.PeerID == "" {
		keeperHealth.PeerID = "no-peer-id"
	}

	// dm.logger.Info(ctx, "Keeper ID",
	// 	observability.Int64("keeper_id", keeperID),
	// 	observability.Bool("online", isActive),
	// )

	// --- UPTIME LOGIC ---
	// If previously online, add to uptime (regardless of new isActive)
	if prevOnline {
		now := time.Now().UTC()
		uptimeToAdd := int64(now.Sub(prevLastCheckedIn).Seconds())
		if uptimeToAdd < 0 {
			uptimeToAdd = 0 // avoid negative values
		}
		newUptime := prevUptime + uptimeToAdd

		// Update uptime field
		uptimeStart := time.Now()
		uptimeCtx, uptimeSpan := dm.tracer.Start(ctx, "db.update_keeper_uptime",
			observability.WithSpanKind(trace.SpanKindClient),
			observability.WithAttributes(
				attribute.String("db.system", "cassandra"),
				attribute.String("db.operation", "update"),
				attribute.String("db.collection", "keeper_data"),
				attribute.Int64("keeper.id", keeperID),
			),
		)
		err = dm.db.Session().Query(`
			UPDATE triggerx.keeper_data 
			SET uptime = ?
			WHERE keeper_id = ?`,
			newUptime, keeperID).Exec()
		uptimeSpan.End()
		metrics.RecordDBOperationDuration(uptimeCtx, "update", time.Since(uptimeStart))
		if err != nil {
			uptimeSpan.SetStatus(codes.Error, err.Error())
			dm.logger.Error(uptimeCtx, "Failed to update keeper uptime",
				observability.Error(err),
				observability.Int64("keeper_id", keeperID),
				observability.String("keeper", keeperHealth.KeeperAddress),
			)
			span.SetStatus(codes.Error, err.Error())
			return err
		}
		uptimeSpan.SetStatus(codes.Ok, "")
	}
	// --- END UPTIME LOGIC ---

	if !isActive {
		// If not active, just set online = false
		updateStart := time.Now()
		updateCtx, updateSpan := dm.tracer.Start(ctx, "db.update_keeper_inactive",
			observability.WithSpanKind(trace.SpanKindClient),
			observability.WithAttributes(
				attribute.String("db.system", "cassandra"),
				attribute.String("db.operation", "update"),
				attribute.String("db.collection", "keeper_data"),
				attribute.Int64("keeper.id", keeperID),
			),
		)
		err = dm.db.Session().Query(`
			UPDATE triggerx.keeper_data 
			SET online = ?
			WHERE keeper_id = ?`,
			false, keeperID).Exec()
		updateSpan.End()
		metrics.RecordDBOperationDuration(updateCtx, "update", time.Since(updateStart))
		if err != nil {
			updateSpan.SetStatus(codes.Error, err.Error())
			dm.logger.Error(updateCtx, "Failed to update keeper inactive status",
				observability.Error(err),
				observability.Int64("keeper_id", keeperID),
				observability.String("keeper", keeperHealth.KeeperAddress),
			)
			span.SetStatus(codes.Error, err.Error())
			return err
		}
		updateSpan.SetStatus(codes.Ok, "")
		span.SetStatus(codes.Ok, "")
		return nil
	}

	// If active, update all fields including last_checked_in
	updateActiveStart := time.Now()
	updateActiveCtx, updateActiveSpan := dm.tracer.Start(ctx, "db.update_keeper_active",
		observability.WithSpanKind(trace.SpanKindClient),
		observability.WithAttributes(
			attribute.String("db.system", "cassandra"),
			attribute.String("db.operation", "update"),
			attribute.String("db.collection", "keeper_data"),
			attribute.Int64("keeper.id", keeperID),
		),
	)
	err = dm.db.Session().Query(`
		UPDATE triggerx.keeper_data 
		SET consensus_address = ?, online = ?, peer_id = ?, version = ?, last_checked_in = ? 
		WHERE keeper_id = ?`,
		keeperHealth.ConsensusAddress, true, keeperHealth.PeerID, keeperHealth.Version, keeperHealth.Timestamp, keeperID).Exec()
	updateActiveSpan.End()
	metrics.RecordDBOperationDuration(updateActiveCtx, "update", time.Since(updateActiveStart))
	if err != nil {
		updateActiveSpan.SetStatus(codes.Error, err.Error())
		dm.logger.Error(updateActiveCtx, "Failed to update keeper status",
			observability.Error(err),
			observability.Int64("keeper_id", keeperID),
		)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	updateActiveSpan.SetStatus(codes.Ok, "")
	span.SetStatus(codes.Ok, "")

	if !isActive {
		go dm.checkAndNotifyOfflineKeeper(ctx, keeperID)
	}

	// dm.logger.Info(ctx, "Successfully updated keeper status",
	// 	observability.Int64("keeper_id", keeperID),
	// 	observability.Bool("active", isActive),
	// )
	return nil
}

func (dm *DatabaseManager) checkAndNotifyOfflineKeeper(ctx context.Context, keeperID int64) {
	time.Sleep(10 * time.Minute)

	// dm.logger.Debug(ctx, "Checking current status for offline keeper",
	// 	observability.Int64("keeper_id", keeperID),
	// )

	var online bool
	err := dm.db.Session().Query(`
		SELECT online FROM triggerx.keeper_data WHERE keeper_id = ?`,
		keeperID).Scan(&online)

	if err != nil {
		dm.logger.Error(ctx, "Failed to check keeper online status",
			observability.Error(err),
			observability.Int64("keeper_id", keeperID),
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
			dm.logger.Error(ctx, "Failed to fetch keeper communication info",
				observability.Error(err),
				observability.Int64("keeper_id", keeperID),
			)
			return
		}

		if chatID != 0 {
			telegramMsg := fmt.Sprintf("Keeper %s is down for more than 10 minutes. Please check and start it.", keeperName)
			if err := dm.telegramBot.SendMessage(chatID, telegramMsg); err != nil {
				dm.logger.Error(ctx, "Failed to send Telegram notification",
					observability.Error(err),
					observability.String("keeper", keeperName),
					observability.Int64("keeper_id", keeperID),
				)
			} else {
				// Record successful telegram notification
				// We need to get keeper address from keeperName or keeperID
				var keeperAddress string
				if err := dm.db.Session().Query(`
					SELECT keeper_address FROM triggerx.keeper_data WHERE keeper_id = ?`,
					keeperID).Scan(&keeperAddress); err == nil {
					metrics.RecordTelegramNotification(ctx, keeperAddress)
				}
			}
		} else {
			dm.logger.Warn(ctx, "No Telegram chat ID found",
				observability.String("keeper", keeperName),
				observability.Int64("keeper_id", keeperID),
			)
		}

		if emailID != "" {
			subject := fmt.Sprintf("TriggerX Keeper Down Alert - %s", keeperName)
			emailBody := fmt.Sprintf(`
				<h2>Keeper Update</h2>
				<p>This is a critical information from TriggerX. Your keeper <strong>%s</strong> has been down for more than 10 minutes. Please take action immediately.</p>
				<p>Regards,<br>TriggerX Team</p>
			`, keeperName)

			if err := dm.sendEmailNotification(ctx, emailID, subject, emailBody); err != nil {
				dm.logger.Error(ctx, "Failed to send email notification",
					observability.Error(err),
					observability.String("keeper", keeperName),
					observability.Int64("keeper_id", keeperID),
				)
			}
		} else {
			dm.logger.Warn(ctx, "No email address found",
				observability.String("keeper", keeperName),
				observability.Int64("keeper_id", keeperID),
			)
		}

		dm.logger.Debug(ctx, "Completed notification process for offline keeper",
			observability.String("keeper", keeperName),
			observability.Int64("keeper_id", keeperID),
		)
	} else {
		dm.logger.Debug(ctx, "Keeper is back online",
			observability.Int64("keeper_id", keeperID),
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
// 		dm.logger.Errorf("[Notification] Failed to marshal Telegram payload", err)
// 		return err
// 	}

// 	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
// 	if err != nil {
// 		dm.logger.Errorf("[Notification] Failed to send Telegram message", err)
// 		return err
// 	}
// 	defer resp.Body.Close()

// 	dm.logger.Infof("[Notification] Telegram message sent successfully to chat ID", chatID, resp.StatusCode)
// 	return nil
// }

func (dm *DatabaseManager) sendEmailNotification(ctx context.Context, to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", config.GetEmailUser())
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer("smtp.zeptomail.in", 587, config.GetEmailUser(), config.GetEmailPassword())
	if err := d.DialAndSend(m); err != nil {
		dm.logger.Error(ctx, "Failed to send email to",
			observability.String("to", to),
			observability.Error(err),
		)
		return err
	}

	dm.logger.Debug(ctx, "Email sent successfully to",
		observability.String("to", to),
	)
	return nil
}

// GetVerifiedKeepers retrieves only verified keepers from the database
func (dm *DatabaseManager) GetVerifiedKeepers(ctx context.Context) ([]types.KeeperInfo, error) {
	// Start a span for the database query operation
	ctx, span := dm.tracer.Start(ctx, "db.get_verified_keepers",
		observability.WithSpanKind(trace.SpanKindClient),
		observability.WithAttributes(
			attribute.String("db.system", "cassandra"),
			attribute.String("db.operation", "select"),
			attribute.String("db.collection", "keeper_data"),
		),
	)
	defer span.End()

	var keepers []types.KeeperInfo

	iter := dm.db.Session().Query(`
		SELECT keeper_name, keeper_address, consensus_address, operator_id, version, peer_id, last_checked_in, on_imua
		FROM triggerx.keeper_data 
		WHERE registered = true AND whitelisted = true 
		ALLOW FILTERING`).Iter()

	var keeperName, keeperAddress, consensusAddress, operatorID, version, peerID string
	var lastCheckedIn time.Time
	var isImua bool

	for iter.Scan(&keeperName, &keeperAddress, &consensusAddress, &operatorID, &version, &peerID, &lastCheckedIn, &isImua) {
		keepers = append(keepers, types.KeeperInfo{
			KeeperName:       keeperName,
			KeeperAddress:    keeperAddress,
			ConsensusAddress: consensusAddress,
			OperatorID:       operatorID,
			Version:          version,
			PeerID:           peerID,
			LastCheckedIn:    lastCheckedIn,
			IsImua:           isImua,
		})
	}

	if err := iter.Close(); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("error closing iterator: %w", err)
	}

	span.SetAttributes(attribute.Int("db.rows_returned", len(keepers)))
	span.SetStatus(codes.Ok, "")

	dm.logger.Debug(ctx, "Retrieved verified keepers from database",
		observability.Int("count", len(keepers)),
	)
	return keepers, nil
}
