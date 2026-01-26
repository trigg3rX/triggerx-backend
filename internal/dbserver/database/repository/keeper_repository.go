package repository

import (
	"sort"
	"strings"

	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type KeeperRepository interface {
	GetKeeperPoints(keeperAddress string) (string, error)
	GetKeeperLeaderboard() ([]types.KeeperLeaderboardEntry, error)
	GetKeeperLeaderboardOnImua() ([]types.KeeperLeaderboardEntry, error)
}

type keeperRepository struct {
	db *database.Connection
}

func NewKeeperRepository(db *database.Connection) KeeperRepository {
	return &keeperRepository{
		db: db,
	}
}

// Read Queries
const (
	GetKeeperPointsQuery = `
		SELECT keeper_points
		FROM triggerx.keeper_data 
		WHERE keeper_address = ?`

	GetKeeperLeaderboardQuery = `
		SELECT keeper_address, keeper_name, no_executed_tasks, no_attested_tasks, keeper_points
		FROM triggerx.keeper_data 
		WHERE registered = true AND network = ? ALLOW FILTERING`
)

func (r *keeperRepository) GetKeeperPoints(keeperAddress string) (string, error) {
	var keeperPoints string
	err := r.db.Session().Query(GetKeeperPointsQuery, strings.ToLower(keeperAddress)).Scan(&keeperPoints)
	if err != nil {
		return "", err
	}
	return keeperPoints, nil
}

func (r *keeperRepository) GetKeeperLeaderboard() ([]types.KeeperLeaderboardEntry, error) {
	// Get keeper leaderboard for Mainnet and Holesky
	iterMainnet := r.db.Session().Query(GetKeeperLeaderboardQuery, types.NetworkMainnet).Iter()
	iterHolesky := r.db.Session().Query(GetKeeperLeaderboardQuery, types.NetworkHolesky).Iter()

	var keeperLeaderboard []types.KeeperLeaderboardEntry
	var keeperEntry types.KeeperLeaderboardEntry

	for iterMainnet.Scan(
		&keeperEntry.KeeperAddress,
		&keeperEntry.KeeperName,
		&keeperEntry.NoExecutedTasks,
		&keeperEntry.NoAttestedTasks,
		&keeperEntry.KeeperPoints,
	) {
		keeperLeaderboard = append(keeperLeaderboard, keeperEntry)
	}

	for iterHolesky.Scan(
		&keeperEntry.KeeperAddress,
		&keeperEntry.KeeperName,
		&keeperEntry.NoExecutedTasks,
		&keeperEntry.NoAttestedTasks,
		&keeperEntry.KeeperPoints,
	) {
		keeperLeaderboard = append(keeperLeaderboard, keeperEntry)
	}

	if err := iterMainnet.Close(); err != nil {
		return nil, err
	}
	if err := iterHolesky.Close(); err != nil {
		return nil, err
	}

	// Sort leaderboard by KeeperPoints (desc), NoExecutedTasks (desc), NoAttestedTasks (desc), KeeperAddress (asc)
	// KeeperPoints is a Wei-based string, so we use big integer comparison
	sort.Slice(keeperLeaderboard, func(i, j int) bool {
		// First compare KeeperPoints using big integer comparison (descending)
		if !types.IsEqual(keeperLeaderboard[i].KeeperPoints, keeperLeaderboard[j].KeeperPoints) {
			return types.IsGreater(keeperLeaderboard[i].KeeperPoints, keeperLeaderboard[j].KeeperPoints)
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

func (r *keeperRepository) GetKeeperLeaderboardOnImua() ([]types.KeeperLeaderboardEntry, error) {
	// Get keeper leaderboard for Imua
	iter := r.db.Session().Query(GetKeeperLeaderboardQuery, types.NetworkMainnet).Iter()

	var keeperLeaderboard []types.KeeperLeaderboardEntry
	var keeperEntry types.KeeperLeaderboardEntry

	for iter.Scan(
		&keeperEntry.KeeperAddress,
		&keeperEntry.KeeperName,
		&keeperEntry.NoExecutedTasks,
		&keeperEntry.NoAttestedTasks,
		&keeperEntry.KeeperPoints,
	) {
		keeperLeaderboard = append(keeperLeaderboard, keeperEntry)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}

	// Sort leaderboard by KeeperPoints (desc), NoExecutedTasks (desc), NoAttestedTasks (desc), KeeperAddress (asc)
	// KeeperPoints is a Wei-based string, so we use big integer comparison
	sort.Slice(keeperLeaderboard, func(i, j int) bool {
		// First compare KeeperPoints using big integer comparison (descending)
		if !types.IsEqual(keeperLeaderboard[i].KeeperPoints, keeperLeaderboard[j].KeeperPoints) {
			return types.IsGreater(keeperLeaderboard[i].KeeperPoints, keeperLeaderboard[j].KeeperPoints)
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
