package repository

import (
	"sort"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/repository/queries"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
)

type KeeperRepository interface {
	GetKeeperLeaderboard() ([]types.KeeperLeaderboardEntry, error)
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

	for iter.Scan(
		&keeperEntry.KeeperID,
		&keeperEntry.KeeperAddress,
		&keeperEntry.KeeperName,
		&keeperEntry.NoExecutedTasks,
		&keeperEntry.NoAttestedTasks,
		&keeperEntry.KeeperPoints,
		&keeperEntry.OnImua,
	) {
		keeperLeaderboard = append(keeperLeaderboard, keeperEntry)
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	// Sort leaderboard by UserPoints (desc), TotalJobs (desc), TotalTasks (desc), UserID (asc)
	sort.Slice(keeperLeaderboard, func(i, j int) bool {
		// First compare UserPoints
		if keeperLeaderboard[i].KeeperPoints != keeperLeaderboard[j].KeeperPoints {
			return keeperLeaderboard[i].KeeperPoints > keeperLeaderboard[j].KeeperPoints
		}
		// If UserPoints equal, compare TotalJobs
		if keeperLeaderboard[i].NoExecutedTasks != keeperLeaderboard[j].NoExecutedTasks {
			return keeperLeaderboard[i].NoExecutedTasks > keeperLeaderboard[j].NoExecutedTasks
		}
		// If TotalJobs equal, compare TotalTasks
		if keeperLeaderboard[i].NoAttestedTasks != keeperLeaderboard[j].NoAttestedTasks {
			return keeperLeaderboard[i].NoAttestedTasks > keeperLeaderboard[j].NoAttestedTasks
		}
		// If all else equal, sort by UserID ascending
		return keeperLeaderboard[i].KeeperID < keeperLeaderboard[j].KeeperID
	})

	return keeperLeaderboard, nil
}
