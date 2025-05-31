package rewards

import (
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/registrar/client"
	"github.com/trigg3rX/triggerx-backend/internal/registrar/config"
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
	startDate := lastRewardsUpdate.AddDate(0, 0, 1).Truncate(24 * time.Hour)
	today := now.Truncate(24 * time.Hour)

	for d := startDate; !d.After(today); d = d.AddDate(0, 0, 1) {
		rewardTime := time.Date(d.Year(), d.Month(), d.Day(), 6, 30, 0, 0, time.UTC)

		if now.After(rewardTime) {
			s.logger.Infof("Distributing missed rewards for %v", d.Format("2006-01-02"))
			err := client.DailyRewardsPoints()
			if err != nil {
				s.logger.Errorf("Failed to distribute rewards for %v: %v", d.Format("2006-01-02"), err)
				continue
			}
			s.updateLastRewardsTimestamp(rewardTime.Format(time.RFC3339))
			s.logger.Infof("Rewards distributed for %v", d.Format("2006-01-02"))
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
