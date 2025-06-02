package repository

import (
	"errors"
	"sort"
	// "time"

	"github.com/gocql/gocql"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/repository/queries"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
)

type KeeperRepository interface {
	CheckKeeperExists(address string) (int64, error)
	CreateKeeper(keeperData types.CreateKeeperData) (int64, error)
	GetKeeperAsPerformer() ([]types.GetPerformerData, error)
	GetKeeperDataByID(id int64) (types.KeeperData, error)
	IncrementKeeperTaskCount(id int64) (int64, error)
	GetKeeperTaskCount(id int64) (int64, error)
	UpdateKeeperPoints(id int64, taskFee int64) (float64, error)
	UpdateKeeperChatID(address string, chatID int64) error
	GetKeeperPointsByIDInDB(id int64) (float64, error)
	GetKeeperCommunicationInfo(id int64) (types.KeeperCommunicationInfo, error)
	GetKeeperLeaderboard() ([]types.KeeperLeaderboardEntry, error)
	GetKeeperLeaderboardByIdentifierInDB(address string, name string) (types.KeeperLeaderboardEntry, error)
}

type keeperRepository struct {
	db *database.Connection
}

func NewKeeperRepository(db *database.Connection) KeeperRepository {
	return &keeperRepository{
		db: db,
	}
}

func (r *keeperRepository) CreateKeeper(keeperData types.CreateKeeperData) (int64, error) {
	var maxKeeperID int64
	err := r.db.Session().Query(queries.GetMaxKeeperIDQuery).Scan(&maxKeeperID)
	if err != nil {
		return -1, err
	}

	err = r.db.Session().Query(queries.CreateNewKeeperQuery, maxKeeperID + 1, keeperData.KeeperName, keeperData.KeeperAddress, 1, 0.0, true, keeperData.EmailID).Exec()
	if err != nil {
		return -1, err
	}

	return maxKeeperID + 1, nil
}

func (r *keeperRepository) GetKeeperAsPerformer() ([]types.GetPerformerData, error) {
	iter := r.db.Session().Query(queries.GetKeeperAsPerformersQuery).Iter()

	var performers []types.GetPerformerData
	var performer types.GetPerformerData
	for iter.Scan(
		&performer.KeeperID, &performer.KeeperAddress) {
		performers = append(performers, performer)
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	return performers, nil
}

func (r *keeperRepository) GetKeeperDataByID(id int64) (types.KeeperData, error) {
	var keeperData types.KeeperData
	err := r.db.Session().Query(queries.GetKeeperDataByIDQuery, id).Scan(&keeperData.KeeperID, &keeperData.KeeperName, &keeperData.KeeperAddress, &keeperData.RegisteredTx, &keeperData.OperatorID, &keeperData.RewardsAddress, &keeperData.RewardsBooster, &keeperData.VotingPower, &keeperData.KeeperPoints, &keeperData.ConnectionAddress, &keeperData.Strategies, &keeperData.Whitelisted, &keeperData.Registered, &keeperData.Online, &keeperData.Version, &keeperData.NoExecutedTasks, &keeperData.NoAttestedTasks, &keeperData.ChatID, &keeperData.EmailID, &keeperData.LastCheckedIn)
	if err != nil {
		return types.KeeperData{}, err
	}
	return keeperData, nil
}

func (r *keeperRepository) CheckKeeperExists(address string) (int64, error) {
	var id int64
	err := r.db.Session().Query(queries.GetKeeperIDByAddressQuery, address).Scan(&id)
	if err == gocql.ErrNotFound {
		return -1, errors.New("keeper not found")
	}
	if err != nil {
		return -1, err
	}
	return id, nil
}

func (r *keeperRepository) UpdateKeeperChatID(address string, chatID int64) error {
	var id int64
	err := r.db.Session().Query(queries.GetKeeperIDByAddressQuery, address).Scan(&id)
	if err != nil {
		return err
	}

	err = r.db.Session().Query(queries.UpdateKeeperChatIDQuery, id, chatID).Exec()
	if err != nil {
		return err
	}

	return nil
}

func (r *keeperRepository) IncrementKeeperTaskCount(id int64) (int64, error) {
	var currentCount int64
	err := r.db.Session().Query(queries.GetKeeperTaskCountByIDQuery, id).Scan(&currentCount)
	if err != nil {
		return 0, err
	}

	newCount := currentCount + 1

	err = r.db.Session().Query(queries.UpdateKeeperTaskCountQuery, newCount, id).Exec()
	if err != nil {
		return 0, err
	}

	return newCount, nil
}

func (r *keeperRepository) GetKeeperTaskCount(id int64) (int64, error) {
	var taskCount int64
	err := r.db.Session().Query(queries.GetKeeperTaskCountByIDQuery, id).Scan(&taskCount)
	if err != nil {
		return 0, err
	}
	return taskCount, nil
}

func (r *keeperRepository) GetKeeperPointsByIDInDB(id int64) (float64, error) {
	var points float64
	err := r.db.Session().Query(queries.GetKeeperPointsByIDQuery, id).Scan(&points)
	if err != nil {
		return 0, err
	}
	return points, nil
}

func (r *keeperRepository) UpdateKeeperPoints(id int64, taskFee int64) (float64, error) {
	var existingPoints float64
	err := r.db.Session().Query(queries.GetKeeperPointsByIDQuery, id).Scan(&existingPoints)
	if err == gocql.ErrNotFound {
		existingPoints = 0
	}
	if err != nil {
		return 0, err
	}

	newPoints := existingPoints + float64(taskFee)

	err = r.db.Session().Query(queries.UpdateKeeperPointsQuery, newPoints, id).Exec()
	if err != nil {
		return 0, err
	}
	return newPoints, nil
}

func (r *keeperRepository) GetKeeperCommunicationInfo(id int64) (types.KeeperCommunicationInfo, error) {
	var keeperCommunicationInfo types.KeeperCommunicationInfo
	err := r.db.Session().Query(queries.GetKeeperCommunicationInfoQuery, id).Scan(&keeperCommunicationInfo.ChatID, &keeperCommunicationInfo.KeeperName, &keeperCommunicationInfo.EmailID)
	if err != nil {
		return types.KeeperCommunicationInfo{}, err
	}
	return keeperCommunicationInfo, nil
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

func (r *keeperRepository) GetKeeperLeaderboardByIdentifierInDB(address string, name string) (types.KeeperLeaderboardEntry, error) {
	var keeperEntry types.KeeperLeaderboardEntry
	var query string
	var args []interface{}

	if address != "" {
		query = queries.GetKeeperLeaderboardByAddressQuery
		args = append(args, address)
	} else {
		query = queries.GetKeeperLeaderboardByNameQuery
		args = append(args, name)
	}

	err := r.db.Session().Query(query, args...).Scan(&keeperEntry.KeeperID, &keeperEntry.KeeperAddress, &keeperEntry.KeeperName, &keeperEntry.NoExecutedTasks, &keeperEntry.NoAttestedTasks, &keeperEntry.KeeperPoints)
	if err != nil {
		return types.KeeperLeaderboardEntry{}, err
	}

	return keeperEntry, nil
}