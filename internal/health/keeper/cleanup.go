package keeper

import (
	"context"
	"fmt"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

const (
	inactivityThreshold  = 70 * time.Second
	stateCleanupInterval = 5 * time.Second
	periodicDumpInterval = 5 * time.Minute  // Persist uptime every 5 minutes
	notificationDelay    = 10 * time.Minute // Send notification after 10 mins of inactivity
	notificationCooldown = 1 * time.Hour    // Wait 1 hour between repeated alerts
	warningThreshold     = 10 * time.Minute // First warning
	criticalThreshold    = 30 * time.Minute // Escalated alert
	urgentThreshold      = 1 * time.Hour    // Final warning
)

func (sm *StateManager) startCleanupRoutine() {
	ticker := time.NewTicker(stateCleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		sm.checkInactiveKeepers()
	}
}

func (sm *StateManager) checkInactiveKeepers() {
	now := time.Now().UTC()
	var inactiveKeepers []*types.HealthKeeperInfo
	var notifyKeepers []*types.HealthKeeperInfo

	sm.mu.Lock()
	for address, state := range sm.keepers {
		inactiveDuration := now.Sub(state.LastCheckedIn)

		// Mark keeper as inactive after 70 seconds
		if state.IsActive && inactiveDuration > inactivityThreshold {
			sm.logger.Info("Keeper became inactive",
				"keeper", address,
				"lastSeen", state.LastCheckedIn.Format(time.RFC3339),
			)
			state.IsActive = false
			// Store a copy of the complete state
			stateCopy := *state
			inactiveKeepers = append(inactiveKeepers, &stateCopy)
		}

		// Send notification after 10 minutes of inactivity
		if !state.IsActive && inactiveDuration > notificationDelay {
			lastNotification, exists := sm.notificationsSent[address]
			// Send if never notified OR cooldown period has passed
			if !exists || now.Sub(lastNotification) > notificationCooldown {
				stateCopy := *state
				notifyKeepers = append(notifyKeepers, &stateCopy)
				sm.notificationsSent[address] = now
			}
		}

		// Clear notification tracking when keeper comes back online
		if state.IsActive {
			delete(sm.notificationsSent, address)
		}
	}
	sm.mu.Unlock()

	// Update database for all inactive keepers
	for _, keeperState := range inactiveKeepers {
		if err := sm.updateKeeperStatusInDatabase(context.Background(), keeperState, false); err != nil {
			sm.logger.Error("Failed to update inactive status",
				"error", err,
				"keeper", keeperState.KeeperAddress,
			)
		}
	}

	// Send notifications for prolonged inactivity
	for _, keeperState := range notifyKeepers {
		sm.sendInactivityNotification(keeperState)
	}
}

// sendInactivityNotification sends Telegram and email notifications for inactive keepers
func (sm *StateManager) sendInactivityNotification(keeper *types.HealthKeeperInfo) {
	ctx := context.Background()
	inactiveDuration := time.Since(keeper.LastCheckedIn)

	// Determine severity level for escalation
	severity := "⚠️ WARNING"
	if inactiveDuration >= urgentThreshold {
		severity = "🚨 URGENT"
	} else if inactiveDuration >= criticalThreshold {
		severity = "⛔ CRITICAL"
	}

	// Get keeper contact information from database
	chatID, email, err := sm.db.GetKeeperChatInfo(ctx, keeper.KeeperAddress)
	if err != nil {
		sm.logger.Error("Failed to get keeper contact info",
			"error", err,
			"keeper", keeper.KeeperAddress,
		)
		return
	}

	// Build notification message
	message := fmt.Sprintf(
		"%s *Keeper Inactivity Alert*\n\n"+
			"Your keeper has been offline for *%s*\n\n"+
			"📍 Keeper: `%s`\n"+
			"🕒 Last seen: %s\n"+
			"📊 Uptime before disconnect: %s\n"+
			"🔢 Version: %s\n\n"+
			"⚡ *Action Required:* Please check your keeper immediately!",
		severity,
		formatDuration(inactiveDuration),
		keeper.KeeperAddress,
		keeper.LastCheckedIn.Format("2006-01-02 15:04:05 UTC"),
		formatDuration(time.Duration(keeper.Uptime)*time.Second),
		keeper.Version,
	)

	// Send Telegram notification if chat ID exists
	if chatID != 0 && sm.notifier != nil {
		if err := sm.notifier.SendTGMessage(chatID, message); err != nil {
			sm.logger.Error("Failed to send Telegram notification",
				"error", err,
				"keeper", keeper.KeeperAddress,
				"chatID", chatID,
			)
		} else {
			sm.logger.Info("Sent inactivity notification via Telegram",
				"keeper", keeper.KeeperAddress,
				"severity", severity,
				"inactive_duration", inactiveDuration.String(),
			)
		}
	} else {
		sm.logger.Warn("No Telegram chat ID found for keeper",
			"keeper", keeper.KeeperAddress,
		)
	}

	// Send email notification as backup if email exists
	if email != "" && sm.notifier != nil {
		subject := fmt.Sprintf("%s Keeper Inactivity Alert - %s", severity, keeper.KeeperAddress[:10])
		htmlBody := fmt.Sprintf(`
			<html>
			<body style="font-family: Arial, sans-serif;">
				<h2 style="color: #e74c3c;">%s Keeper Inactivity Alert</h2>
				<p>Your keeper has been offline for <strong>%s</strong></p>
				<table style="border-collapse: collapse; margin: 20px 0;">
					<tr><td style="padding: 8px;"><strong>Keeper Address:</strong></td><td style="padding: 8px;">%s</td></tr>
					<tr><td style="padding: 8px;"><strong>Last Seen:</strong></td><td style="padding: 8px;">%s</td></tr>
					<tr><td style="padding: 8px;"><strong>Uptime Before Disconnect:</strong></td><td style="padding: 8px;">%s</td></tr>
					<tr><td style="padding: 8px;"><strong>Version:</strong></td><td style="padding: 8px;">%s</td></tr>
				</table>
				<p style="color: #e74c3c; font-weight: bold;">⚡ Action Required: Please check your keeper immediately!</p>
			</body>
			</html>
		`,
			severity,
			formatDuration(inactiveDuration),
			keeper.KeeperAddress,
			keeper.LastCheckedIn.Format("2006-01-02 15:04:05 UTC"),
			formatDuration(time.Duration(keeper.Uptime)*time.Second),
			keeper.Version,
		)

		if err := sm.notifier.SendEmailMessage(email, subject, htmlBody); err != nil {
			sm.logger.Error("Failed to send email notification",
				"error", err,
				"keeper", keeper.KeeperAddress,
				"email", email,
			)
		} else {
			sm.logger.Info("Sent inactivity notification via email",
				"keeper", keeper.KeeperAddress,
			)
		}
	}
}

// formatDuration converts duration to human-readable format
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d seconds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	}
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	if hours < 24 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	days := hours / 24
	remainingHours := hours % 24
	return fmt.Sprintf("%dd %dh", days, remainingHours)
}

// startPeriodicDumpRoutine periodically persists keeper state to database
func (sm *StateManager) startPeriodicDumpRoutine() {
	ticker := time.NewTicker(periodicDumpInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := sm.PeriodicDump(context.Background()); err != nil {
			sm.logger.Error("Periodic dump failed", "error", err)
		}
	}
}
