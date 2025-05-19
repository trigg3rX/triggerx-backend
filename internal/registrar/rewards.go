package registrar

import (
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/registrar/config"
	"github.com/trigg3rX/triggerx-backend/internal/registrar/database"
)

func StartDailyRewardsPoints() {
	logger.Info("Starting daily rewards service...")

	lastRewardsUpdate, err := time.Parse(time.RFC3339, config.LastRewardsUpdate)
	if err != nil {
		logger.Errorf("Failed to parse last rewards update: %v", err)
		lastRewardsUpdate = time.Now().AddDate(0, 0, -1)
	}

	logger.Infof("Last rewards update: %v", lastRewardsUpdate)

	now := time.Now().UTC()
	startDate := lastRewardsUpdate.AddDate(0, 0, 1).Truncate(24 * time.Hour)
	today := now.Truncate(24 * time.Hour)

	for d := startDate; !d.After(today); d = d.AddDate(0, 0, 1) {
		rewardTime := time.Date(d.Year(), d.Month(), d.Day(), 6, 30, 0, 0, time.UTC)

		if now.After(rewardTime) {
			logger.Infof("Distributing missed rewards for %v", d.Format("2006-01-02"))
			err := database.DailyRewardsPoints() // If possible, make this accept a date
			if err != nil {
				logger.Errorf("Failed to distribute rewards for %v: %v", d.Format("2006-01-02"), err)
				continue
			}
			config.UpdateLastRewardsTimestamp(rewardTime.Format(time.RFC3339))
			logger.Infof("Rewards distributed for %v", d.Format("2006-01-02"))
		}
	}

	go scheduleNextReward()
}

func scheduleNextReward() {
	for {
		now := time.Now()
		nextReward := time.Date(now.Year(), now.Month(), now.Day(), 06, 30, 0, 0, time.UTC)

		// If we've already passed 15:30 today, schedule for tomorrow
		if now.After(nextReward) {
			nextReward = nextReward.AddDate(0, 0, 1)
		}

		// Calculate duration until next reward time
		waitDuration := nextReward.Sub(now)
		logger.Infof("Next reward scheduled for: %v (in %v)", nextReward, waitDuration)

		// Wait until the scheduled time
		time.Sleep(waitDuration)

		// It's time to distribute rewards
		logger.Info("It's 07:30, distributing daily rewards now...")
		err := database.DailyRewardsPoints()
		if err != nil {
			logger.Errorf("Failed to distribute daily rewards: %v", err)
		} else {
			newTimestamp := time.Now().Format(time.RFC3339)
			config.UpdateLastRewardsTimestamp(newTimestamp)
			logger.Info("Daily rewards distributed successfully")
		}

		// Short sleep to avoid potential double execution
		time.Sleep(time.Minute)
	}
}
