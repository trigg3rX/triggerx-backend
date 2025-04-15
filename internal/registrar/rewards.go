package registrar

import (
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/registrar/config"
	"github.com/trigg3rX/triggerx-backend/internal/registrar/database"
)

func StartDailyRewardsPoints() {
	logger.Info("Starting daily rewards service...")

	// Check if we already rewarded today
	lastRewardsUpdate, err := time.Parse(time.RFC3339, config.LastRewardsUpdate)
	if err != nil {
		logger.Errorf("Failed to parse last rewards update: %v", err)
		// Continue with default time if parsing fails
		lastRewardsUpdate = time.Now().AddDate(0, 0, -1) // Default to yesterday
	}

	logger.Infof("Last rewards update: %v", lastRewardsUpdate)

	// Check if we need to reward immediately upon service start
	now := time.Now()
	rewardTime := time.Date(now.Year(), now.Month(), now.Day(), 15, 30, 0, 0, now.Location())

	// If the scheduled time for today has already passed AND we haven't rewarded today yet
	if now.After(rewardTime) && lastRewardsUpdate.Day() != now.Day() {
		logger.Info("15:30 has already passed for today and rewards haven't been distributed yet, distributing rewards now...")
		err := database.DailyRewardsPoints()
		if err != nil {
			logger.Errorf("Failed to distribute daily rewards: %v", err)
		} else {
			newTimestamp := time.Now().Format(time.RFC3339)
			config.UpdateLastRewardsTimestamp(newTimestamp)
			logger.Info("Daily rewards distributed successfully")
		}
	}

	// Schedule the next reward distribution
	go scheduleNextReward()
}

func scheduleNextReward() {
	for {
		now := time.Now()
		nextReward := time.Date(now.Year(), now.Month(), now.Day(), 05, 30, 0, 0, time.UTC)

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
		logger.Info("It's 15:30, distributing daily rewards now...")
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
