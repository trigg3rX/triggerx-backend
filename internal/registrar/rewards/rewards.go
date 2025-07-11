package rewards

import (
	"context"
	"time"

	"github.com/trigg3rX/triggerx-backend-imua/internal/registrar/clients/database"
	"github.com/trigg3rX/triggerx-backend-imua/internal/registrar/sync"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/logging"
)

type RewardsService struct {
	logger         logging.Logger
	stateManager   *sync.StateManager
	databaseClient *database.DatabaseClient
	ctx            context.Context
}

func NewRewardsService(logger logging.Logger, stateManager *sync.StateManager, databaseClient *database.DatabaseClient) *RewardsService {
	return &RewardsService{
		logger:         logger,
		stateManager:   stateManager,
		databaseClient: databaseClient,
		ctx:            context.Background(),
	}
}

func (s *RewardsService) StartDailyRewardsPoints() {
	s.logger.Info("Starting daily rewards service...")

	ctx, cancel := context.WithTimeout(s.ctx, 10*time.Second)
	defer cancel()

	lastRewardsUpdate, err := s.stateManager.GetLastRewardsUpdate(ctx)
	if err != nil {
		s.logger.Errorf("Failed to get last rewards update from Redis: %v", err)
		lastRewardsUpdate = time.Now().AddDate(0, 0, -1)
	} else if lastRewardsUpdate.IsZero() {
		s.logger.Info("No previous rewards update found, starting from yesterday")
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
			err := s.databaseClient.DailyRewardsPoints()
			if err != nil {
				s.logger.Errorf("Failed to distribute rewards for %v: %v", d.Format("2006-01-02"), err)
				continue
			}
			s.updateLastRewardsTimestamp(rewardTime)
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
		err := s.databaseClient.DailyRewardsPoints()
		if err != nil {
			s.logger.Errorf("Failed to distribute daily rewards: %v", err)
		} else {
			s.updateLastRewardsTimestamp(time.Now().UTC())
			s.logger.Info("Daily rewards distributed successfully")
		}

		time.Sleep(time.Minute)
	}
}

func (s *RewardsService) updateLastRewardsTimestamp(timestamp time.Time) {
	ctx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
	defer cancel()

	if err := s.stateManager.SetLastRewardsUpdate(ctx, timestamp); err != nil {
		s.logger.Errorf("Failed to update last rewards timestamp in Redis: %v", err)
	} else {
		s.logger.Debugf("Updated last rewards timestamp to %s", timestamp.Format(time.RFC3339))
	}
}

// GetRewardsHealth returns health information about the rewards service
func (s *RewardsService) GetRewardsHealth() map[string]interface{} {
	ctx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
	defer cancel()

	health := map[string]interface{}{
		"service": "rewards",
		"status":  "running",
	}

	// Get last rewards update from Redis
	lastUpdate, err := s.stateManager.GetLastRewardsUpdate(ctx)
	if err != nil {
		health["last_rewards_update_error"] = err.Error()
		health["status"] = "degraded"
	} else {
		health["last_rewards_update"] = lastUpdate.Format(time.RFC3339)
		health["last_update_age"] = time.Since(lastUpdate).String()

		// Check if rewards are overdue (more than 25 hours since last update)
		if time.Since(lastUpdate) > 25*time.Hour {
			health["status"] = "overdue"
			health["warning"] = "rewards distribution is overdue"
		}
	}

	// Calculate next reward time
	now := time.Now()
	nextReward := time.Date(now.Year(), now.Month(), now.Day(), 6, 30, 0, 0, time.UTC)
	if now.After(nextReward) {
		nextReward = nextReward.AddDate(0, 0, 1)
	}
	health["next_scheduled_reward"] = nextReward.Format(time.RFC3339)
	health["time_until_next_reward"] = nextReward.Sub(now).String()

	return health
}

func (s *RewardsService) Close() {
	s.databaseClient.Close()
}