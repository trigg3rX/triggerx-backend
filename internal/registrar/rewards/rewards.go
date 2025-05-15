package rewards

import (
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/registrar/config"
	"github.com/trigg3rX/triggerx-backend/internal/registrar/client"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type RewardsService struct {
	logger logging.Logger
}

func NewRewardsService(logger logging.Logger) *RewardsService {
	return &RewardsService{
		logger: logger,
	}
}

func (s *RewardsService) StartDailyRewardsPoints() {
	s.logger.Info("Starting daily rewards service...")

	lastRewardsUpdate, err := time.Parse(time.RFC3339, config.GetLastRewardsUpdate())
	if err != nil {
		s.logger.Errorf("Failed to parse last rewards update: %v", err)
		lastRewardsUpdate = time.Now().AddDate(0, 0, -1)
	}

	s.logger.Infof("Last rewards update: %v", lastRewardsUpdate)

	now := time.Now()
	rewardTime := time.Date(now.Year(), now.Month(), now.Day(), 06, 30, 0, 0, time.UTC)

	if now.After(rewardTime) && lastRewardsUpdate.Day() != now.Day() {
		s.logger.Info("06:30 has already passed for today and rewards haven't been distributed yet, distributing rewards now...")
		err := client.DailyRewardsPoints()
		if err != nil {
			s.logger.Errorf("Failed to distribute daily rewards: %v", err)
		} else {
			newTimestamp := time.Now().Format(time.RFC3339)
			s.updateLastRewardsTimestamp(newTimestamp)
			s.logger.Info("Daily rewards distributed successfully")
		}
	}

	go s.scheduleNextReward()
}

func (s *RewardsService) scheduleNextReward() {
	for {
		now := time.Now()
		nextReward := time.Date(now.Year(), now.Month(), now.Day(), 06, 30, 0, 0, time.UTC)

		if now.After(nextReward) {
			nextReward = nextReward.AddDate(0, 0, 1)
		}

		waitDuration := nextReward.Sub(now)
		s.logger.Infof("Next reward scheduled for: %v (in %v)", nextReward, waitDuration)

		time.Sleep(waitDuration)

		s.logger.Info("It's 06:30, distributing daily rewards now...")
		err := client.DailyRewardsPoints()
		if err != nil {
			s.logger.Errorf("Failed to distribute daily rewards: %v", err)
		} else {
			newTimestamp := time.Now().Format(time.RFC3339)
			s.updateLastRewardsTimestamp(newTimestamp)
			s.logger.Info("Daily rewards distributed successfully")
		}

		time.Sleep(time.Minute)
	}
}

func (s *RewardsService) updateLastRewardsTimestamp(timestamp string) {
	// TODO: Implement timestamp update logic
}
