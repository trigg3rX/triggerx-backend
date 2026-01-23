package repository

import (
	"sort"
	"strconv"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/repository/queries"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type KeeperRepository interface {
	GetKeeperLeaderboard() ([]types.KeeperLeaderboardEntry, error)
	GetKeeperLeaderboardByOnImua(onImua bool) ([]types.KeeperLeaderboardEntry, error) // Returns all since schema doesn't have on_imua
}

type keeperRepository struct {
	db *database.Connection
}

func NewKeeperRepository(db *database.Connection) KeeperRepository {
	return &keeperRepository{
		db: db,
	}
}

func (r *keeperRepository) GetKeeperLeaderboard() ([]types.KeeperLeaderboardEntry, error) {
	iter := r.db.Session().Query(queries.GetKeeperLeaderboardQuery).Iter()

	var keeperLeaderboard []types.KeeperLeaderboardEntry
	var keeperEntry types.KeeperLeaderboardEntry
	var keeperPointsStr string

	for iter.Scan(
		&keeperEntry.KeeperAddress,
		&keeperEntry.KeeperName,
		&keeperEntry.NoExecutedTasks,
		&keeperEntry.NoAttestedTasks,
		&keeperPointsStr,
	) {
		// Convert keeperPoints from string to float64
		if keeperPointsFloat, err := strconv.ParseFloat(keeperPointsStr, 64); err == nil {
			keeperEntry.KeeperPoints = keeperPointsFloat
		} else {
			keeperEntry.KeeperPoints = 0
		}
		keeperEntry.OnImua = false // Schema doesn't have on_imua field
		keeperLeaderboard = append(keeperLeaderboard, keeperEntry)
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	// Sort leaderboard by KeeperPoints (desc), NoExecutedTasks (desc), NoAttestedTasks (desc), KeeperAddress (asc)
	sort.Slice(keeperLeaderboard, func(i, j int) bool {
		// First compare KeeperPoints
		if keeperLeaderboard[i].KeeperPoints != keeperLeaderboard[j].KeeperPoints {
			return keeperLeaderboard[i].KeeperPoints > keeperLeaderboard[j].KeeperPoints
		}
		// If KeeperPoints equal, compare NoExecutedTasks
		if keeperLeaderboard[i].NoExecutedTasks != keeperLeaderboard[j].NoExecutedTasks {
			return keeperLeaderboard[i].NoExecutedTasks > keeperLeaderboard[j].NoExecutedTasks
		}
		// If NoExecutedTasks equal, compare NoAttestedTasks
		if keeperLeaderboard[i].NoAttestedTasks != keeperLeaderboard[j].NoAttestedTasks {
			return keeperLeaderboard[i].NoAttestedTasks > keeperLeaderboard[j].NoAttestedTasks
		}
		// If all else equal, sort by KeeperAddress ascending
		return keeperLeaderboard[i].KeeperAddress < keeperLeaderboard[j].KeeperAddress
	})

	return keeperLeaderboard, nil
}

func (r *keeperRepository) GetKeeperLeaderboardByOnImua(onImua bool) ([]types.KeeperLeaderboardEntry, error) {
	// Schema doesn't have on_imua field, so return all keepers
	// This method is kept for backward compatibility with handlers
	return r.GetKeeperLeaderboard()
}
