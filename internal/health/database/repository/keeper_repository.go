package repository

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-gomail/gomail"
	"github.com/trigg3rX/triggerx-backend/internal/health/config"
	"github.com/trigg3rX/triggerx-backend/internal/health/core/telegram"
	"github.com/trigg3rX/triggerx-backend/internal/health/metrics"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// KeeperRepository defines the interface for keeper data operations
type KeeperRepository interface {
	UpdateKeeperHealth(ctx context.Context, keeperHealth types.KeeperHealthCheckInRequest, isActive bool) error
	GetVerifiedKeepers(ctx context.Context) ([]types.KeeperInfo, error)
	GetKeeperUptimes(ctx context.Context) (map[string]int64, error)
}

type keeperRepository struct {
	db          *database.Connection
	logger      observability.Logger
	tracer      observability.Tracer
	telegramBot *telegram.Bot
}

// NewKeeperRepository creates a new keeper repository instance
func NewKeeperRepository(db *database.Connection, logger observability.Logger, tracer observability.Tracer, telegramBot *telegram.Bot) KeeperRepository {
	// Create a new logger with component field
	dbLogger := logger.With(observability.String("component", "keeper_repository"))

	return &keeperRepository{
		db:          db,
		logger:      dbLogger,
		tracer:      tracer,
		telegramBot: telegramBot,
	}
}

// UpdateKeeperHealth registers a new keeper or updates an existing one (status = true)
func (r *keeperRepository) UpdateKeeperHealth(ctx context.Context, keeperHealth types.KeeperHealthCheckInRequest, isActive bool) error {
	// Start a span for the database update operation
	ctx, span := r.tracer.Start(ctx, "db.update_keeper_health",
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

	keeperHealth.KeeperAddress = strings.ToLower(keeperHealth.KeeperAddress)
	keeperHealth.ConsensusAddress = strings.ToLower(keeperHealth.ConsensusAddress)

	if len(keeperHealth.KeeperAddress) > 0 && !bytes.HasPrefix([]byte(keeperHealth.KeeperAddress), []byte("0x")) {
		r.logger.Debug(ctx, "Adding 0x prefix to keeper address",
			observability.String("keeper", keeperHealth.KeeperAddress),
		)
		keeperHealth.KeeperAddress = "0x" + keeperHealth.KeeperAddress
	}

	var prevOnline bool
	var prevLastCheckedIn time.Time
	var prevUptime int64

	// Fetch previous online status, last_checked_in, and uptime
	selectStart := time.Now()
	_, selectSpan := r.tracer.Start(ctx, "db.select_keeper_data",
		observability.WithSpanKind(trace.SpanKindClient),
		observability.WithAttributes(
			attribute.String("db.system", "cassandra"),
			attribute.String("db.operation", "select"),
			attribute.String("db.collection", "keeper_data"),
		),
	)
	err := r.db.Session().Query(`
		SELECT online, last_checked_in, uptime FROM triggerx.keeper_data WHERE keeper_address = ?`,
		keeperHealth.KeeperAddress).Scan(&prevOnline, &prevLastCheckedIn, &prevUptime)
	selectSpan.End()
	metrics.RecordDBOperationDuration(ctx, "select", time.Since(selectStart))
	if err != nil {
		selectSpan.SetStatus(codes.Error, err.Error())
		r.logger.Error(ctx, "Failed to retrieve keeper previous status",
			observability.String("keeper", keeperHealth.KeeperAddress),
			observability.Error(err),
		)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	selectSpan.SetStatus(codes.Ok, "")

	if keeperHealth.PeerID == "" {
		keeperHealth.PeerID = "no-peer-id"
	}

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
		uptimeCtx, uptimeSpan := r.tracer.Start(ctx, "db.update_keeper_uptime",
			observability.WithSpanKind(trace.SpanKindClient),
			observability.WithAttributes(
				attribute.String("db.system", "cassandra"),
				attribute.String("db.operation", "update"),
				attribute.String("db.collection", "keeper_data"),
				attribute.String("keeper.address", keeperHealth.KeeperAddress),
			),
		)
		err = r.db.Session().Query(`
			UPDATE triggerx.keeper_data 
			SET uptime = ?
			WHERE keeper_address = ?`,
			newUptime, keeperHealth.KeeperAddress).Exec()
		uptimeSpan.End()
		metrics.RecordDBOperationDuration(uptimeCtx, "update", time.Since(uptimeStart))
		if err != nil {
			uptimeSpan.SetStatus(codes.Error, err.Error())
			r.logger.Error(uptimeCtx, "Failed to update keeper uptime",
				observability.Error(err),
				observability.String("keeper", keeperHealth.KeeperAddress),
			)
			span.SetStatus(codes.Error, err.Error())
			return err
		}
		uptimeSpan.SetStatus(codes.Ok, "")
		// Update metric with the new uptime
		metrics.UpdateKeeperUptime(ctx, keeperHealth.KeeperAddress, float64(newUptime))
	} else {
		// If keeper wasn't previously online, still update metric with current uptime
		// This ensures the metric is always up-to-date
		metrics.UpdateKeeperUptime(ctx, keeperHealth.KeeperAddress, float64(prevUptime))
	}
	// --- END UPTIME LOGIC ---

	if !isActive {
		// If not active, just set online = false
		updateStart := time.Now()
		updateCtx, updateSpan := r.tracer.Start(ctx, "db.update_keeper_inactive",
			observability.WithSpanKind(trace.SpanKindClient),
			observability.WithAttributes(
				attribute.String("db.system", "cassandra"),
				attribute.String("db.operation", "update"),
				attribute.String("db.collection", "keeper_data"),
				attribute.String("keeper.address", keeperHealth.KeeperAddress),
			),
		)
		err = r.db.Session().Query(`
			UPDATE triggerx.keeper_data 
			SET online = ?
			WHERE keeper_address = ?`,
			false, keeperHealth.KeeperAddress).Exec()
		updateSpan.End()
		metrics.RecordDBOperationDuration(updateCtx, "update", time.Since(updateStart))
		if err != nil {
			updateSpan.SetStatus(codes.Error, err.Error())
			r.logger.Error(updateCtx, "Failed to update keeper inactive status",
				observability.Error(err),
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
	updateActiveCtx, updateActiveSpan := r.tracer.Start(ctx, "db.update_keeper_active",
		observability.WithSpanKind(trace.SpanKindClient),
		observability.WithAttributes(
			attribute.String("db.system", "cassandra"),
			attribute.String("db.operation", "update"),
			attribute.String("db.collection", "keeper_data"),
			attribute.String("keeper.address", keeperHealth.KeeperAddress),
		),
	)
	// Default network to "mainnet" if not provided (backward compatibility)
	network := keeperHealth.Network
	if network == "" {
		network = "mainnet"
	}

	err = r.db.Session().Query(`
		UPDATE triggerx.keeper_data 
		SET consensus_address = ?, online = ?, peer_id = ?, version = ?, last_checked_in = ?, network = ? 
		WHERE keeper_address = ?`,
		keeperHealth.ConsensusAddress, true, keeperHealth.PeerID, keeperHealth.Version, keeperHealth.Timestamp, network, keeperHealth.KeeperAddress).Exec()
	updateActiveSpan.End()
	metrics.RecordDBOperationDuration(updateActiveCtx, "update", time.Since(updateActiveStart))
	if err != nil {
		updateActiveSpan.SetStatus(codes.Error, err.Error())
		r.logger.Error(updateActiveCtx, "Failed to update keeper status",
			observability.Error(err),
			observability.String("keeper", keeperHealth.KeeperAddress),
		)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	updateActiveSpan.SetStatus(codes.Ok, "")
	span.SetStatus(codes.Ok, "")

	if !isActive {
		go r.checkAndNotifyOfflineKeeper(ctx, keeperHealth.KeeperAddress)
	}

	return nil
}

func (r *keeperRepository) checkAndNotifyOfflineKeeper(ctx context.Context, keeperAddress string) {
	time.Sleep(config.GetNotificationOfflineDelay())

	var online bool
	err := r.db.Session().Query(`
		SELECT online FROM triggerx.keeper_data WHERE keeper_address = ?`,
		keeperAddress).Scan(&online)

	if err != nil {
		r.logger.Error(ctx, "Failed to check keeper online status",
			observability.Error(err),
			observability.String("keeper", keeperAddress),
		)
		return
	}

	if !online {
		var chatID int64
		var keeperName, emailID string
		err := r.db.Session().Query(`
			SELECT chat_id, keeper_name, email_id 
			FROM triggerx.keeper_data 
			WHERE keeper_address = ?`,
			keeperAddress).Scan(&chatID, &keeperName, &emailID)

		if err != nil {
			r.logger.Error(ctx, "Failed to fetch keeper communication info",
				observability.Error(err),
				observability.String("keeper", keeperAddress),
			)
			return
		}

		if chatID != 0 {
			telegramMsg := fmt.Sprintf("Keeper %s is down for more than 10 minutes. Please check and start it.", keeperName)
			if err := r.telegramBot.SendMessage(chatID, telegramMsg); err != nil {
				r.logger.Error(ctx, "Failed to send Telegram notification",
					observability.Error(err),
					observability.String("keeper", keeperName),
					observability.String("keeper", keeperAddress),
				)
			} else {
				metrics.RecordTelegramNotification(ctx, keeperAddress)
			}
		} else {
			r.logger.Warn(ctx, "No Telegram chat ID found",
				observability.String("keeper", keeperName),
				observability.String("keeper", keeperAddress),
			)
		}

		if emailID != "" {
			subject := fmt.Sprintf("TriggerX Keeper Down Alert - %s", keeperName)
			emailBody := fmt.Sprintf(`
				<h2>Keeper Update</h2>
				<p>This is a critical information from TriggerX. Your keeper <strong>%s</strong> has been down for more than 10 minutes. Please take action immediately.</p>
				<p>Regards,<br>TriggerX Team</p>
			`, keeperName)

			if err := r.sendEmailNotification(ctx, emailID, subject, emailBody); err != nil {
				r.logger.Error(ctx, "Failed to send email notification",
					observability.Error(err),
					observability.String("keeper", keeperName),
					observability.String("keeper", keeperAddress),
				)
			}
		} else {
			r.logger.Warn(ctx, "No email address found",
				observability.String("keeper", keeperName),
				observability.String("keeper", keeperAddress),
			)
		}

		r.logger.Debug(ctx, "Completed notification process for offline keeper",
			observability.String("keeper", keeperName),
			observability.String("keeper", keeperAddress),
		)
	} else {
		r.logger.Debug(ctx, "Keeper is back online",
			observability.String("keeper", keeperAddress),
		)
	}
}

func (r *keeperRepository) sendEmailNotification(ctx context.Context, to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", config.GetEmailUser())
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer("smtp.zeptomail.in", 587, config.GetEmailUser(), config.GetEmailPassword())
	if err := d.DialAndSend(m); err != nil {
		r.logger.Error(ctx, "Failed to send email to",
			observability.String("to", to),
			observability.Error(err),
		)
		return err
	}

	r.logger.Debug(ctx, "Email sent successfully to",
		observability.String("to", to),
	)
	return nil
}

// GetVerifiedKeepers retrieves only verified keepers from the database
func (r *keeperRepository) GetVerifiedKeepers(ctx context.Context) ([]types.KeeperInfo, error) {
	// Start a span for the database query operation
	ctx, span := r.tracer.Start(ctx, "db.get_verified_keepers",
		observability.WithSpanKind(trace.SpanKindClient),
		observability.WithAttributes(
			attribute.String("db.system", "cassandra"),
			attribute.String("db.operation", "select"),
			attribute.String("db.collection", "keeper_data"),
		),
	)
	defer span.End()

	var keepers []types.KeeperInfo

	iter := r.db.Session().Query(`
		SELECT keeper_name, keeper_address, consensus_address, operator_id, version, peer_id, last_checked_in, network
		FROM triggerx.keeper_data 
		WHERE registered = true 
		ALLOW FILTERING`).Iter()

	var keeperName, keeperAddress, consensusAddress, operatorID, version, peerID, network string
	var lastCheckedIn time.Time

	for iter.Scan(&keeperName, &keeperAddress, &consensusAddress, &operatorID, &version, &peerID, &lastCheckedIn, &network) {
		keepers = append(keepers, types.KeeperInfo{
			KeeperName:       keeperName,
			KeeperAddress:    keeperAddress,
			ConsensusAddress: consensusAddress,
			OperatorID:       operatorID,
			Version:          version,
			PeerID:           peerID,
			LastCheckedIn:    lastCheckedIn,
			Network:          network,
		})
	}

	if err := iter.Close(); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("error closing iterator: %w", err)
	}

	span.SetAttributes(attribute.Int("db.rows_returned", len(keepers)))
	span.SetStatus(codes.Ok, "")

	r.logger.Debug(ctx, "Retrieved verified keepers from database",
		observability.Int("count", len(keepers)),
	)
	return keepers, nil
}

// GetKeeperUptimes retrieves uptime for all keepers from the database
func (r *keeperRepository) GetKeeperUptimes(ctx context.Context) (map[string]int64, error) {
	// Start a span for the database query operation
	_, span := r.tracer.Start(ctx, "db.get_keeper_uptimes",
		observability.WithSpanKind(trace.SpanKindClient),
		observability.WithAttributes(
			attribute.String("db.system", "cassandra"),
			attribute.String("db.operation", "select"),
			attribute.String("db.collection", "keeper_data"),
		),
	)
	defer span.End()

	uptimes := make(map[string]int64)

	iter := r.db.Session().Query(`
		SELECT keeper_address, uptime
		FROM triggerx.keeper_data 
		WHERE registered = true 
		ALLOW FILTERING`).Iter()

	var keeperAddress string
	var uptime int64

	for iter.Scan(&keeperAddress, &uptime) {
		uptimes[strings.ToLower(keeperAddress)] = uptime
	}

	if err := iter.Close(); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("error closing iterator: %w", err)
	}

	span.SetAttributes(attribute.Int("db.rows_returned", len(uptimes)))
	span.SetStatus(codes.Ok, "")

	return uptimes, nil
}
