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

	now := time.Now()
	rewardTime := time.Date(now.Year(), now.Month(), now.Day(), 06, 30, 0, 0, time.UTC)

	if now.After(rewardTime) && lastRewardsUpdate.Day() != now.Day() {
		logger.Info("06:30 has already passed for today and rewards haven't been distributed yet, distributing rewards now...")
		err := database.DailyRewardsPoints()
		if err != nil {
			logger.Errorf("Failed to distribute daily rewards: %v", err)
		} else {
			newTimestamp := time.Now().Format(time.RFC3339)
			config.UpdateLastRewardsTimestamp(newTimestamp)
			logger.Info("Daily rewards distributed successfully")
		}
	}

	go scheduleNextReward()
}

func scheduleNextReward() {
	for {
		now := time.Now()
		nextReward := time.Date(now.Year(), now.Month(), now.Day(), 06, 30, 0, 0, time.UTC)

		if now.After(nextReward) {
			nextReward = nextReward.AddDate(0, 0, 1)
		}

		waitDuration := nextReward.Sub(now)
		logger.Infof("Next reward scheduled for: %v (in %v)", nextReward, waitDuration)

		time.Sleep(waitDuration)

		logger.Info("It's 06:30, distributing daily rewards now...")
		err := database.DailyRewardsPoints()
		if err != nil {
			logger.Errorf("Failed to distribute daily rewards: %v", err)
		} else {
			newTimestamp := time.Now().Format(time.RFC3339)
			config.UpdateLastRewardsTimestamp(newTimestamp)
			logger.Info("Daily rewards distributed successfully")
		}

		time.Sleep(time.Minute)
	}
}
